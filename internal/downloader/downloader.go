package downloader

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	appConfig "nugs-dl/internal/config"
	"nugs-dl/internal/logger" // Import the logger package
	"nugs-dl/internal/queue"
	"nugs-dl/pkg/api"
)

// ErrDuplicateCompleted is returned when a download is attempted for content
// that has already been successfully downloaded in a previous job.
var ErrDuplicateCompleted = errors.New("content already downloaded in a completed job")

// Downloader handles the core logic for fetching and downloading.
type Downloader struct {
	Config       *appConfig.AppConfig      // Holds the application configuration
	HTTPClient   *http.Client              // Shared HTTP client (with cookie jar)
	ProgressChan chan<- api.ProgressUpdate // Channel to send progress updates
	QueueMgr     *queue.QueueManager       // Added QueueManager reference
	// TODO: Add fields for progress reporting callbacks/channels
}

// DownloadOptions specifies options for a specific download operation.
// This will replace direct reliance on the global Config struct from main.go
type DownloadOptions struct {
	ForceVideo   bool
	SkipVideos   bool
	SkipChapters bool
	// We might need specific format overrides here too if the API allows
}

// NewDownloader creates a new Downloader instance.
func NewDownloader(cfg *appConfig.AppConfig, client *http.Client, progressChan chan<- api.ProgressUpdate, qm *queue.QueueManager) *Downloader {
	return &Downloader{
		Config:       cfg,
		HTTPClient:   client,
		ProgressChan: progressChan, // Store the channel
		QueueMgr:     qm,           // Store queue manager
	}
}

// Download processes a single job based on its URL and options.
// Changed signature to accept *api.DownloadJob
func (d *Downloader) Download(job *api.DownloadJob) error {
	logger.Info("[Downloader] Starting job", "jobID", job.ID, "url", job.OriginalUrl, "options", job.Options)

	var (
		token        string
		userID       string
		legacyToken  string
		legacyUguid  string // Needed for purchased items
		streamParams *StreamParams
		err          error
	)

	// --- Authentication (Uses d.Config) ---
	if d.Config.Token != "" {
		token = d.Config.Token // Use provided token
		logger.Info("Using provided auth token from config.")
	} else if d.Config.Email != "" && d.Config.Password != "" {
		// Authenticate with email/password
		token, err = d.Authenticate(d.Config.Email, d.Config.Password)
		if err != nil {
			logger.Error("Authentication failed using email/password", "error", err)
			return fmt.Errorf("authentication failed: %w", err)
		}
		logger.Info("Successfully authenticated using email/password.")
	} else {
		logger.Error("Authentication required: email/password or token missing in config")
		return errors.New("authentication required: provide email/password or token in config")
	}

	// --- Get User Info & Subscription Details (Needed for StreamParams) ---
	// Use token (either provided or obtained via login)
	logger.Info("Fetching user and subscription info...")
	userID, err = d.GetUserInfo(token)
	if err != nil {
		logger.Error("Failed to get user info", "error", err)
		return fmt.Errorf("failed to get user info: %w", err)
	}
	subInfo, err := d.GetSubInfo(token)
	if err != nil {
		logger.Error("Failed to get subscription info", "error", err)
		return fmt.Errorf("failed to get subscription info: %w", err)
	}

	planDesc, _ := getPlan(subInfo) // Use internal getPlan
	logger.Info("User subscription plan determined", "plan", planDesc)

	// --- Extract Legacy Tokens (Needed for some playlist types AND purchased URLs) ---
	legacyToken, legacyUguid, err = ExtractLegacyTokens(token)
	if err != nil {
		logger.Warn("Could not extract legacy tokens from main auth token", "error", err)
	}

	// --- Parse Stream Params (Needed for most downloads) ---
	streamParams, err = ParseStreamParams(userID, subInfo)
	if err != nil {
		logger.Error("Failed to parse stream parameters", "error", err)
		// This is likely fatal for most downloads
		return fmt.Errorf("failed to parse stream parameters: %w", err)
	}

	// --- Process the Single URL ---
	rawUrl := job.OriginalUrl
	logger.Info("Processing URL", "url", rawUrl, "jobID", job.ID)

	// Convert job options from api.DownloadOptions to downloader.DownloadOptions
	// (They are currently identical, but this makes dependencies clearer)
	dlOpts := DownloadOptions{
		ForceVideo:   job.Options.ForceVideo,
		SkipVideos:   job.Options.SkipVideos,
		SkipChapters: job.Options.SkipChapters,
	}

	id, urlType := CheckUrl(rawUrl) // Use CheckUrl from utils.go

	switch urlType {
	case ReleaseUrl:
		logger.Info("URL Type: Release/Album/Show", "id", id, "jobID", job.ID)
		err = d.processAlbum(job.ID, id, dlOpts, streamParams, nil) // Pass job.ID
	case UserPlaylistHashUrl, UserPlaylistLibUrl:
		logger.Info("URL Type: User Playlist", "id", id, "jobID", job.ID)
		err = d.processPlaylist(job.ID, id, legacyToken, false, streamParams) // Pass job.ID
	case CatalogPlaylistUrl:
		logger.Info("URL Type: Catalog Playlist (Short URL)", "originalUrl", rawUrl, "jobID", job.ID)
		resolvedUrl, resolveErr := d.resolveRedirectURL(rawUrl)
		if resolveErr != nil {
			logger.Error("Failed to resolve catalog playlist short URL", "url", rawUrl, "error", resolveErr, "jobID", job.ID)
			err = fmt.Errorf("failed to resolve catalog playlist short URL: %w", resolveErr)
		} else {
			resolvedId, resolvedType := CheckUrl(resolvedUrl)
			if resolvedType == UserPlaylistHashUrl || resolvedType == UserPlaylistLibUrl {
				logger.Info("Processing resolved playlist ID", "resolvedID", resolvedId, "jobID", job.ID)
				err = d.processPlaylist(job.ID, resolvedId, legacyToken, true, streamParams) // Pass job.ID
			} else {
				logger.Error("Resolved URL is not a recognized playlist type", "resolvedUrl", resolvedUrl, "type", resolvedType, "jobID", job.ID)
				err = fmt.Errorf("resolved URL %s is not a recognized playlist type (type %d)", resolvedUrl, resolvedType)
			}
		}
	case VideoUrlHash:
		logger.Info("URL Type: Video", "id", id, "jobID", job.ID)
		err = d.processVideo(job.ID, id, "", dlOpts, streamParams, nil, false) // Pass job.ID
	case ArtistUrl:
		logger.Info("URL Type: Artist", "id", id, "jobID", job.ID)
		err = d.processArtist(job.ID, id, dlOpts, streamParams) // Pass job.ID
	case ExclusiveLivestreamUrl, WatchExclusiveLivestreamUrl, MyWebcastLibUrl, WatchReleaseUrl:
		logger.Info("URL Type: Livestream/Webcast/WatchRelease (Container ID)", "id", id, "urlType", urlType, "jobID", job.ID)
		err = d.processAlbum(job.ID, id, dlOpts, streamParams, nil) // Pass job.ID
	case MyWebcastHashUrl:
		logger.Info("URL Type: Livestream/Webcast (Show ID)", "id", id, "jobID", job.ID)
		err = d.processVideo(job.ID, id, "", dlOpts, streamParams, nil, true) // Pass job.ID
	case PurchasedUrl:
		logger.Info("URL Type: Purchased Item", "url", rawUrl, "jobID", job.ID)
		if legacyUguid == "" {
			logger.Error("Cannot download purchased item: failed to extract legacy Uguid from token", "jobID", job.ID)
			err = errors.New("cannot download purchased item: failed to extract legacy Uguid from token")
		} else {
			parsedUrl, _ := url.Parse(rawUrl)
			queryParams, _ := url.ParseQuery(parsedUrl.RawQuery)
			showID := queryParams.Get("showID")
			if showID == "" {
				logger.Error("Could not extract showID from purchased URL query parameters", "url", rawUrl, "jobID", job.ID)
				err = errors.New("could not extract showID from purchased URL query parameters")
			} else {
				logger.Info("Processing purchased video item", "showID", showID, "jobID", job.ID)
				err = d.processVideo(job.ID, showID, legacyUguid, dlOpts, streamParams, nil, true) // Pass job.ID
			}
		}
	default:
		logger.Error("Unsupported URL type", "url", rawUrl, "urlType", urlType, "jobID", job.ID)
		err = fmt.Errorf("unsupported URL type: %s (type %d)", rawUrl, urlType)
	}

	if err != nil {
		logger.Error("Error processing job in Download method", "jobID", job.ID, "url", rawUrl, "error", err)
		// Worker will handle setting job status to failed based on this error
		return err
	}

	logger.Info("Downloader finished job successfully", "jobID", job.ID, "url", rawUrl)
	return nil // Success
}

// Need helper functions for processAlbum, processPlaylist etc. to accept jobID now
// e.g., func (d *Downloader) processAlbum(jobID string, albumID string, opts DownloadOptions, streamParams *StreamParams, preloadedMeta *AlbArtResp) error
// Need to update processTrack calls within those to pass jobID

// Need to update processArtist to pass jobID down to processAlbum
