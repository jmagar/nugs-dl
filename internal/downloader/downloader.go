package downloader

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	appConfig "nugs-dl/internal/config"
	"nugs-dl/internal/queue"
	"nugs-dl/pkg/api"
)

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
	fmt.Printf("[Downloader] Starting job ID: %s, URL: %s, Options: %+v\n", job.ID, job.OriginalUrl, job.Options)

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
		fmt.Println("Using provided auth token.")
	} else if d.Config.Email != "" && d.Config.Password != "" {
		// Authenticate with email/password
		token, err = d.Authenticate(d.Config.Email, d.Config.Password)
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
	} else {
		return errors.New("authentication required: provide email/password or token in config")
	}

	// --- Get User Info & Subscription Details (Needed for StreamParams) ---
	// Use token (either provided or obtained via login)
	fmt.Println("Fetching user and subscription info...")
	userID, err = d.GetUserInfo(token)
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}
	subInfo, err := d.GetSubInfo(token)
	if err != nil {
		return fmt.Errorf("failed to get subscription info: %w", err)
	}

	planDesc, _ := getPlan(subInfo) // Use internal getPlan
	fmt.Printf("Subscription plan: %s\n", planDesc)

	// --- Extract Legacy Tokens (Needed for some playlist types AND purchased URLs) ---
	legacyToken, legacyUguid, err = ExtractLegacyTokens(token)
	if err != nil {
		fmt.Printf("Warning: could not extract legacy tokens: %v\n", err)
	}

	// --- Parse Stream Params (Needed for most downloads) ---
	streamParams, err = ParseStreamParams(userID, subInfo)
	if err != nil {
		// This is likely fatal for most downloads
		return fmt.Errorf("failed to parse stream parameters: %w", err)
	}

	// --- Process the Single URL ---
	rawUrl := job.OriginalUrl
	fmt.Printf("Processing URL: %s\n", rawUrl)

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
		fmt.Println("Type: Release/Album/Show")
		err = d.processAlbum(job.ID, id, dlOpts, streamParams, nil) // Pass job.ID
	case UserPlaylistHashUrl, UserPlaylistLibUrl:
		fmt.Println("Type: User Playlist")
		err = d.processPlaylist(job.ID, id, legacyToken, false, streamParams) // Pass job.ID
	case CatalogPlaylistUrl:
		fmt.Println("Type: Catalog Playlist (Short URL)")
		resolvedUrl, resolveErr := d.resolveRedirectURL(rawUrl)
		if resolveErr != nil {
			err = fmt.Errorf("failed to resolve catalog playlist short URL: %w", resolveErr)
		} else {
			resolvedId, resolvedType := CheckUrl(resolvedUrl)
			if resolvedType == UserPlaylistHashUrl || resolvedType == UserPlaylistLibUrl {
				fmt.Printf("Processing resolved playlist ID: %s\n", resolvedId)
				err = d.processPlaylist(job.ID, resolvedId, legacyToken, true, streamParams) // Pass job.ID
			} else {
				err = fmt.Errorf("resolved URL %s is not a recognized playlist type (type %d)", resolvedUrl, resolvedType)
			}
		}
	case VideoUrlHash:
		fmt.Println("Type: Video")
		err = d.processVideo(job.ID, id, "", dlOpts, streamParams, nil, false) // Pass job.ID
	case ArtistUrl:
		fmt.Println("Type: Artist")
		err = d.processArtist(job.ID, id, dlOpts, streamParams) // Pass job.ID
	case ExclusiveLivestreamUrl, WatchExclusiveLivestreamUrl, MyWebcastLibUrl:
		fmt.Println("Type: Livestream/Webcast (Container ID)")
		err = d.processAlbum(job.ID, id, dlOpts, streamParams, nil) // Pass job.ID
	case MyWebcastHashUrl:
		fmt.Println("Type: Livestream/Webcast (Show ID)")
		err = d.processVideo(job.ID, id, "", dlOpts, streamParams, nil, true) // Pass job.ID
	case PurchasedUrl:
		fmt.Println("Type: Purchased Item")
		if legacyUguid == "" {
			err = errors.New("cannot download purchased item: failed to extract legacy Uguid from token")
		} else {
			parsedUrl, _ := url.Parse(rawUrl)
			queryParams, _ := url.ParseQuery(parsedUrl.RawQuery)
			showID := queryParams.Get("showID")
			if showID == "" {
				err = errors.New("could not extract showID from purchased URL query parameters")
			} else {
				err = d.processVideo(job.ID, showID, legacyUguid, dlOpts, streamParams, nil, true) // Pass job.ID
			}
		}
	default:
		err = fmt.Errorf("unsupported URL type: %s", rawUrl)
	}

	if err != nil {
		fmt.Printf("Error processing job %s for URL %s: %v\n", job.ID, rawUrl, err)
		// Worker will handle setting job status to failed based on this error
		return err
	}

	fmt.Printf("\nDownloader finished job %s for URL %s.\n", job.ID, rawUrl)
	return nil // Success
}

// Need helper functions for processAlbum, processPlaylist etc. to accept jobID now
// e.g., func (d *Downloader) processAlbum(jobID string, albumID string, opts DownloadOptions, streamParams *StreamParams, preloadedMeta *AlbArtResp) error
// Need to update processTrack calls within those to pass jobID

// Need to update processArtist to pass jobID down to processAlbum
