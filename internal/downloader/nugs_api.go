package downloader

import (
	"bytes"         // Added for re-creating io.ReadCloser from logged body
	"nugs-dl/internal/logger"
	"encoding/json"
	"errors"
	"fmt"
	"io"            // Added for io.ReadAll
	"net/http"
	"net/url"
	"strconv"
	// We might need StreamParams type here from types.go
)

// Constants related to Nugs API endpoints and user agents
const (
	streamApiBase = "https://streamapi.nugs.net/" // Base URL for most metadata APIs
	userAgentTwo  = "nugsnetAndroid"              // A different user agent used for some API calls
	devKey        = "x7f54tgbdyc64y656thy47er4"   // Developer key used for some secure calls
)

// --- Metadata Fetching Functions ---

// getAlbumMeta retrieves metadata for a specific album/release/video container ID.
func (d *Downloader) getAlbumMeta(containerId string) (*AlbumMeta, error) {
	logger.Debug("[getAlbumMeta] Entry", "containerId", containerId)
	req, err := http.NewRequest(http.MethodGet, streamApiBase+"api.aspx", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create album meta request: %w", err)
	}
	query := url.Values{}
	query.Set("method", "catalog.container")
	query.Set("containerID", containerId)
	query.Set("vdisp", "1")
	req.URL.RawQuery = query.Encode()
	// Use the standard userAgent for this one
	req.Header.Add("User-Agent", userAgent)
	do, err := d.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform album meta request: %w", err)
	}
	defer do.Body.Close()

	// Read the body for logging, then replace it for decoding
	bodyBytes, errRead := io.ReadAll(do.Body)
	if errRead != nil {
		logger.Warn("[getAlbumMeta] Failed to read response body for logging", "containerId", containerId, "error", errRead)
		// If reading fails, we can't log the body.
		// The original defer do.Body.Close() will still run.
		// We must return an error as we can't proceed with decoding.
		return nil, fmt.Errorf("failed to read album meta response body: %w", errRead)
	} else {
		logger.Debug("[getAlbumMeta] Raw API Response", "containerId", containerId, "responseBody", string(bodyBytes))
	}
	// Restore the body for the JSON decoder
	do.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if do.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get album meta failed: %s", do.Status)
	}
	var obj AlbumMeta
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return nil, fmt.Errorf("failed to decode album meta response: %w", err)
	}
	return &obj, nil
}

// getPlistMeta retrieves metadata for a playlist (catalog or user).
func (d *Downloader) getPlistMeta(plistId, email, legacyToken string, isCatalogPlist bool) (*PlistMeta, error) {
	var path string
	if isCatalogPlist {
		path = "api.aspx"
	} else {
		// User playlists require the secure API path and additional auth parameters
		path = "secureApi.aspx"
	}
	req, err := http.NewRequest(http.MethodGet, streamApiBase+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create plist meta request: %w", err)
	}
	query := url.Values{}
	if isCatalogPlist {
		query.Set("method", "catalog.playlist")
		query.Set("plGUID", plistId)
	} else {
		query.Set("method", "user.playlist")
		query.Set("playlistID", plistId)
		query.Set("developerKey", devKey)
		query.Set("user", email)        // Requires user email
		query.Set("token", legacyToken) // Requires legacy token
	}
	req.URL.RawQuery = query.Encode()
	// User Agent Two is used for playlist calls
	req.Header.Add("User-Agent", userAgentTwo)
	do, err := d.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform plist meta request: %w", err)
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get plist meta failed: %s", do.Status)
	}
	var obj PlistMeta
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return nil, fmt.Errorf("failed to decode plist meta response: %w", err)
	}
	return &obj, nil
}

// getArtistMeta retrieves all containers (albums/videos) for a given artist ID.
// Note: This performs multiple requests if the artist has more than 100 items.
func (d *Downloader) getArtistMeta(artistId string) ([]*AlbArtResp, error) {
	var allContainers []*AlbArtResp
	offset := 1
	limit := 100 // API limit per page
	query := url.Values{}
	query.Set("method", "catalog.containersAll")
	query.Set("limit", strconv.Itoa(limit))
	query.Set("artistList", artistId)
	query.Set("availType", "1") // Filter for available items?
	query.Set("vdisp", "1")

	for {
		req, err := http.NewRequest(http.MethodGet, streamApiBase+"api.aspx", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create artist meta request: %w", err)
		}
		query.Set("startOffset", strconv.Itoa(offset))
		req.URL.RawQuery = query.Encode()
		// Standard user agent for artist meta
		req.Header.Add("User-Agent", userAgent)
		do, err := d.HTTPClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to perform artist meta request (offset %d): %w", offset, err)
		}

		if do.StatusCode != http.StatusOK {
			status := do.Status
			do.Body.Close()
			return nil, fmt.Errorf("get artist meta failed (offset %d): %s", offset, status)
		}

		var page ArtistMeta
		err = json.NewDecoder(do.Body).Decode(&page)
		do.Body.Close() // Close body after decoding
		if err != nil {
			return nil, fmt.Errorf("failed to decode artist meta response (offset %d): %w", offset, err)
		}

		if page.Response == nil || page.Response.Containers == nil {
			// Handle cases where Response might be nil
			if offset == 1 {
				return nil, errors.New("artist API returned no response data")
			} else {
				break // Assume end if response is empty on later pages
			}
		}

		retLen := len(page.Response.Containers)
		if retLen == 0 {
			break // No more containers found
		}

		allContainers = append(allContainers, page.Response.Containers...)
		offset += retLen
	}

	if len(allContainers) == 0 {
		return nil, errors.New("artist API returned no containers")
	}
	return allContainers, nil
}

// getPurchasedManUrl retrieves the manifest URL for a purchased video.
func (d *Downloader) getPurchasedManUrl(skuID int, showID, userID, uguID string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, streamApiBase+"bigriver/vidPlayer.aspx", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create purchased manifest request: %w", err)
	}
	query := url.Values{}
	query.Set("skuId", strconv.Itoa(skuID))
	query.Set("showId", showID)
	query.Set("uguid", uguID)
	query.Set("nn_userID", userID)
	query.Set("app", "1")
	req.URL.RawQuery = query.Encode()
	// User agent two for purchased videos
	req.Header.Add("User-Agent", userAgentTwo)
	do, err := d.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform purchased manifest request: %w", err)
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get purchased manifest failed: %s", do.Status)
	}
	var obj PurchasedManResp
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return "", fmt.Errorf("failed to decode purchased manifest response: %w", err)
	}
	if obj.ResponseCode != 0 || obj.FileURL == "" { // Check response code too
		return "", fmt.Errorf("purchased manifest API reported an error (code %d)", obj.ResponseCode)
	}
	return obj.FileURL, nil
}

// getStreamMeta retrieves the streaming URL for a track or video chapter using subscription parameters.
// format = 0 for video/chapters, specific format ID (1-5) for audio tracks.
func (d *Downloader) getStreamMeta(trackIdOrContainerId, skuId, format int, streamParams *StreamParams) (string, error) {
	if streamParams == nil {
		return "", errors.New("stream parameters are required to get stream metadata")
	}
	req, err := http.NewRequest(http.MethodGet, streamApiBase+"bigriver/subPlayer.aspx", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create stream meta request: %w", err)
	}
	query := url.Values{}
	if format == 0 { // Indicates video or chapter
		query.Set("skuId", strconv.Itoa(skuId))
		query.Set("containerID", strconv.Itoa(trackIdOrContainerId))
		query.Set("chap", "1")
	} else { // Indicates audio track
		query.Set("platformID", strconv.Itoa(format)) // Format ID corresponds to platformID
		query.Set("trackID", strconv.Itoa(trackIdOrContainerId))
	}
	query.Set("app", "1")
	// Add required stream parameters from subscription info
	query.Set("subscriptionID", streamParams.SubscriptionID)
	query.Set("subCostplanIDAccessList", streamParams.SubCostplanIDAccessList)
	query.Set("nn_userID", streamParams.UserID)
	query.Set("startDateStamp", streamParams.StartStamp)
	query.Set("endDateStamp", streamParams.EndStamp)
	req.URL.RawQuery = query.Encode()
	// User agent two for stream metadata
	req.Header.Add("User-Agent", userAgentTwo)
	do, err := d.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform stream meta request: %w", err)
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get stream meta failed: %s", do.Status)
	}
	var obj StreamMeta
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return "", fmt.Errorf("failed to decode stream meta response: %w", err)
	}
	if obj.StreamLink == "" { // Explicitly check for empty stream link
		return "", errors.New("stream metadata API returned an empty stream link")
	}
	return obj.StreamLink, nil
}
