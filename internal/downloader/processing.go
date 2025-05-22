package downloader

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	// Corrected import if previously missed
	"nugs-dl/pkg/api"
	// TODO: Move utils like sanitise here or to utils.go
)

// Constants related to downloading
const (
	playerUrl = "https://play.nugs.net/" // Used as Referer for downloads
)

// --- Track Processing ---

// queryQuality finds the first matching quality definition for a stream URL.
// (Moved from main.go)
func queryQuality(streamUrl string) *Quality {
	for k, v := range qualityMap { // Assumes qualityMap is available (defined in types.go)
		if strings.Contains(streamUrl, k) {
			// Create a new Quality instance to avoid modifying the map's template
			foundQuality := v
			foundQuality.URL = streamUrl
			return &foundQuality
		}
	}
	return nil
}

// getTrackQual selects the desired quality from available ones, handling fallbacks.
// (Moved from main.go)
func getTrackQual(quals []*Quality, wantFmt int) *Quality {
	// Try to find the exact wanted format
	for _, quality := range quals {
		if quality.Format == wantFmt {
			return quality
		}
	}

	// If exact not found, try fallback
	fbFmt, hasFallback := trackFallback[wantFmt]
	if hasFallback {
		fmt.Printf("Format %d unavailable, falling back to %d...\n", wantFmt, fbFmt)
		for _, quality := range quals {
			if quality.Format == fbFmt {
				return quality
			}
		}
	}

	// If fallback also not found, maybe return the first available or highest quality?
	if len(quals) > 0 {
		fmt.Println("Fallback format also unavailable, selecting first available.")
		return quals[0]
	}

	return nil // No qualities available
}

// checkIfHlsOnly checks if all available quality URLs are HLS manifests.
// (Moved from main.go)
func checkIfHlsOnly(quals []*Quality) bool {
	if len(quals) == 0 {
		return false // Or true? Undefined case.
	}
	for _, quality := range quals {
		if !strings.Contains(quality.URL, ".m3u8?") {
			return false // Found a non-HLS URL
		}
	}
	return true // All URLs were HLS
}

// downloadFile performs the actual HTTP GET request and saves the file,
// reporting progress via WriteCounter.
// Now requires jobID to associate progress.
func (d *Downloader) downloadFile(jobID, filePath, downloadUrl string) error {
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create/open file %s: %w", filePath, err)
	}
	defer f.Close()

	// Send initial "Starting download" progress update
	d.sendProgress(api.ProgressUpdate{
		JobID:       jobID,
		Message:     "Starting download...",
		CurrentFile: filepath.Base(filePath),
	})

	req, err := http.NewRequest(http.MethodGet, downloadUrl, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request for %s: %w", downloadUrl, err)
	}
	req.Header.Add("Referer", playerUrl)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Range", "bytes=0-")

	do, err := d.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to start download for %s: %w", downloadUrl, err)
	}
	defer do.Body.Close()

	if do.StatusCode != http.StatusOK && do.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("bad status code %s downloading %s", do.Status, downloadUrl)
	}

	totalBytes := do.ContentLength
	totalStr := "Unknown Size"
	if totalBytes > 0 {
		totalStr = humanize.Bytes(uint64(totalBytes))
		fmt.Printf("Content-Length: %d bytes (%s)\n", totalBytes, totalStr)
	} else {
		fmt.Printf("Warning: No Content-Length header received, progress will be indeterminate\n")
	}

	// Initialize progress counter with JobID and channel
	counter := &WriteCounter{
		JobID:          jobID,
		Total:          totalBytes,
		TotalStr:       totalStr,
		StartTime:      time.Now().UnixMilli(),
		ProgressChan:   d.ProgressChan, // Pass the channel
		lastUpdateTime: 0,              // Initialize last update time
	}

	_, err = io.Copy(f, io.TeeReader(do.Body, counter))
	// fmt.Println("") // No longer needed as WriteCounter doesn't print newline
	if err != nil {
		return fmt.Errorf("failed during copy for %s: %w", downloadUrl, err)
	}

	// Send final 100% update
	d.sendProgress(api.ProgressUpdate{
		JobID:           jobID,
		Percentage:      100.0,
		BytesDownloaded: counter.Downloaded,
		TotalBytes:      counter.Total,
		SpeedBPS:        0, // Speed is irrelevant on completion
	})

	return nil
}

// sendProgress is a helper to safely send updates on the progress channel.
func (d *Downloader) sendProgress(update api.ProgressUpdate) {
	if d.ProgressChan == nil {
		fmt.Printf("[Downloader] Progress channel is nil for Job %s\n", update.JobID)
		return // No channel configured
	}
	
	fmt.Printf("[Downloader] Sending progress update: Job %s, Percentage: %.1f%%, Speed: %d B/s, Message: %s\n", 
		update.JobID, update.Percentage, update.SpeedBPS, update.Message)
	
	// Use non-blocking send
	select {
	case d.ProgressChan <- update:
		fmt.Printf("[Downloader] Progress update sent successfully for Job %s\n", update.JobID)
	default:
		fmt.Printf("[Downloader Warning] Progress channel full for Job %s, discarding update: %s\n", update.JobID, update.Message)
	}
}

// processTrack handles fetching metadata, selecting quality, and downloading a single track.
// (Refactored from processTrack in main.go)
func (d *Downloader) processTrack(jobID string, folPath string, trackNum, trackTotal int, track *Track, streamParams *StreamParams) error {
	// Calculate track-based progress percentage (completed tracks / total tracks * 100)
	trackProgressPercentage := float64(trackNum-1) / float64(trackTotal) * 100.0
	wantFmt := d.Config.Format // Get desired format from downloader config
	var (
		quals      []*Quality
		chosenQual *Quality
		err        error
	)

	// --- Get Stream URLs for different formats ---
	// Try formats 1, 4, 7, 10 to cover different possibilities
	for _, apiFmtId := range [4]int{1, 4, 7, 10} {
		streamUrl, err := d.getStreamMeta(track.TrackID, 0, apiFmtId, streamParams)
		if err != nil {
			// Log non-fatal error, maybe only one format failed
			fmt.Printf("Warning: failed to get stream metadata for track %d, format %d: %v\n", track.TrackID, apiFmtId, err)
			continue // Try next format
		}
		quality := queryQuality(streamUrl)
		if quality != nil {
			// Check if this format (by platformID/format code) is already in our list
			found := false
			for _, q := range quals {
				if q.Format == quality.Format {
					found = true
					break
				}
			}
			if !found {
				quals = append(quals, quality)
			}
		} else {
			fmt.Printf("Warning: Unsupported quality format from URL: %s\n", streamUrl)
		}
	}

	if len(quals) == 0 {
		return errors.New("no valid stream URLs found for track")
	}

	// --- Select Quality / Handle HLS ---
	isHlsOnly := checkIfHlsOnly(quals)

	// --- Prepare Filename (determine path first, extension might change for HLS) ---
	// Use a placeholder extension for HLS initially, downloadHls will create final path.
	initialExtension := ".tmp_download"
	if !isHlsOnly {
		// For non-HLS, we need chosenQual now to get the correct extension
		chosenQual = getTrackQual(quals, wantFmt)
		if chosenQual == nil {
			return errors.New("could not determine a suitable download quality/format")
		}
		initialExtension = chosenQual.Extension
	}
	trackFname := fmt.Sprintf(
		"%02d. %s%s", trackNum, SanitizeFilename(track.SongTitle), initialExtension,
	)
	trackPath := filepath.Join(folPath, trackFname)

	if isHlsOnly {
		fmt.Println("HLS-only track. Only AAC is available.")
		// Need the *master* playlist URL to start the HLS process
		masterPlaylistUrl := ""
		for _, q := range quals {
			if q.Format == 6 { // Find the original master playlist URL
				masterPlaylistUrl = q.URL
				break
			}
		}
		if masterPlaylistUrl == "" {
			return errors.New("could not find master playlist URL for HLS track")
		}
		// Before HLS download call:
		d.sendProgress(api.ProgressUpdate{
			JobID: jobID, 
			Message: fmt.Sprintf("Downloading HLS track %d/%d", trackNum, trackTotal), 
			CurrentFile: trackFname,
			Percentage: trackProgressPercentage,
			CurrentTrack: trackNum,
			TotalTracks: trackTotal,
		})
		// Call the HLS download function
		err = d.downloadHls(jobID, trackPath, masterPlaylistUrl)
		// Whether HLS succeeds or fails, we return the result here.
		return err
	} else {
		// Non-HLS path
		// chosenQual is already set from above
		if chosenQual == nil { // Safety check
			return errors.New("internal error: chosenQual is nil in non-HLS path")
		}

		// --- Check Existence (for non-HLS) ---
		exists, err := FileExists(trackPath) // Use utility function
		if err != nil {
			return fmt.Errorf("failed to check if track exists %s: %w", trackPath, err)
		}
		if exists {
			fmt.Printf("Track %d of %d already exists: %s\n", trackNum, trackTotal, trackFname)
			return nil // Skip download
		}

		// --- Download (for non-HLS) ---
		fmt.Printf(
			"Downloading track %d of %d: %s - %s\n", trackNum, trackTotal, track.SongTitle,
			chosenQual.Specs,
		)
		// Before download call:
		d.sendProgress(api.ProgressUpdate{
			JobID: jobID, 
			Message: fmt.Sprintf("Downloading track %d/%d", trackNum, trackTotal), 
			CurrentFile: trackFname,
			Percentage: trackProgressPercentage,
			CurrentTrack: trackNum,
			TotalTracks: trackTotal,
		})
		// Make download call pass jobID
		err = d.downloadFile(jobID, trackPath, chosenQual.URL)

		if err != nil {
			os.Remove(trackPath)
			return fmt.Errorf("download failed for track %s: %w", trackFname, err)
		}

		fmt.Printf("Successfully downloaded track %d: %s\n", trackNum, trackFname)
		// After download call: calculate progress based on completed tracks
		completedTrackProgress := float64(trackNum) / float64(trackTotal) * 100.0
		d.sendProgress(api.ProgressUpdate{
			JobID: jobID, 
			Message: fmt.Sprintf("Finished track %d/%d", trackNum, trackTotal), 
			CurrentFile: trackFname, 
			Percentage: completedTrackProgress,
			CurrentTrack: trackNum,
			TotalTracks: trackTotal,
		})
		return nil // Explicitly return nil on non-HLS success
	}
	// Code below the if/else is now truly unreachable
}

// --- Album / Artist / Playlist Processing ---

// getVideoSku finds the SKU ID for video products.
// (Moved from main.go)
func getVideoSku(products []Product) int {
	for _, product := range products {
		formatStr := product.FormatStr
		if formatStr == "VIDEO ON DEMAND" || formatStr == "LIVE HD VIDEO" {
			return product.SkuID
		}
	}
	return 0
}

// getLstreamSku finds the SKU ID for livestream video products.
// (Moved from main.go)
func getLstreamSku(products []*ProductFormatList) int {
	for _, product := range products {
		if product.FormatStr == "LIVE HD VIDEO" {
			return product.SkuID
		}
	}
	return 0
}

// processAlbum downloads all tracks or the video for an album/show.
// (Refactored from album in main.go)
func (d *Downloader) processAlbum(jobID string, albumID string, opts DownloadOptions, streamParams *StreamParams, preloadedMeta *AlbArtResp) error {
	var (
		meta   *AlbArtResp
		tracks []Track
		err    error
	)

	if preloadedMeta != nil {
		// Use metadata preloaded by artist call
		meta = preloadedMeta
	} else {
		// Fetch metadata directly if album ID is provided
		albumMeta, err := d.getAlbumMeta(albumID)
		if err != nil {
			return fmt.Errorf("failed to get metadata for album %s: %w", albumID, err)
		}
		if albumMeta.Response == nil {
			return fmt.Errorf("API returned empty response for album %s", albumID)
		}
		meta = albumMeta.Response
	}

	// Extract and update artwork URL
	artworkURL := extractArtworkUrl(meta) // Use a helper function
	if artworkURL != "" {
		d.QueueMgr.UpdateJobArtwork(jobID, artworkURL)
	}

	// Determine track list (API uses 'tracks' or 'songs' field inconsistently)
	if len(meta.Tracks) > 0 {
		tracks = meta.Tracks
	} else if len(meta.Songs) > 0 {
		tracks = meta.Songs
	}
	trackTotal := len(tracks)

	// Check for video
	skuID := getVideoSku(meta.Products)

	if skuID == 0 && trackTotal < 1 {
		return fmt.Errorf("release %s has no tracks or videos", albumID)
	}

	// --- Decide whether to download video or tracks ---
	if skuID != 0 { // Video exists
		if opts.SkipVideos {
			fmt.Printf("Skipping video for album/show ID %s\n", albumID)
			// If it's ONLY video, we skip entirely. If tracks also exist, continue to tracks.
			if trackTotal < 1 {
				return nil
			}
		} else if opts.ForceVideo || trackTotal < 1 {
			fmt.Printf("Processing video for album/show ID %s\n", albumID)
			// TODO: Call refactored video processing function
			err = d.processVideo(jobID, albumID, "", opts, streamParams, meta, false)
			return errors.New("video download not yet implemented in refactor") // Temp error
		}
		// If video exists but not forced and tracks exist, fall through to download tracks
	}

	// --- Download Tracks ---
	albumFolder := meta.ArtistName + " - " + strings.TrimRight(meta.ContainerInfo, " ")
	fmt.Println("Album:", albumFolder)

	// Update job title with the proper album information
	d.QueueMgr.UpdateJobTitle(jobID, albumFolder)

	// Sanitize and potentially shorten folder name (keep original logic?)
	albumPath := filepath.Join(d.Config.OutPath, SanitizeFilename(albumFolder))
	err = MakeDirs(albumPath) // TODO: Move MakeDirs to utils
	if err != nil {
		return fmt.Errorf("failed to create album folder %s: %w", albumPath, err)
	}

	for i, track := range tracks {
		trackNum := i + 1
		err := d.processTrack(jobID, albumPath, trackNum, trackTotal, &track, streamParams)
		if err != nil {
			// Log error but continue with other tracks?
			fmt.Printf("Error processing track %d (%s): %v\n", trackNum, track.SongTitle, err)
			// Optionally collect errors and return them at the end
		}
	}
	return nil // Or return collected errors
}

// processArtist downloads all albums/shows for an artist.
// (Refactored from artist in main.go)
func (d *Downloader) processArtist(jobID string, artistId string, opts DownloadOptions, streamParams *StreamParams) error {
	containers, err := d.getArtistMeta(artistId)
	if err != nil {
		return fmt.Errorf("failed to get metadata for artist %s: %w", artistId, err)
	}
	if len(containers) == 0 {
		return fmt.Errorf("no containers found for artist %s", artistId)
	}

	fmt.Println("Artist:", containers[0].ArtistName) // Assuming first container has artist name
	itemTotal := len(containers)
	fmt.Printf("Found %d items for artist.\n", itemTotal)

	// Update job title with the artist name
	d.QueueMgr.UpdateJobTitle(jobID, containers[0].ArtistName)

	var firstErr error // Variable to store the first error encountered

	for i, containerMeta := range containers {
		fmt.Printf("\nProcessing item %d of %d: %s\n", i+1, itemTotal, containerMeta.ContainerInfo)
		containerIDStr := strconv.Itoa(containerMeta.ContainerID)

		// Call processAlbum and store the potential error
		processErr := d.processAlbum(jobID, containerIDStr, opts, streamParams, nil)

		if processErr != nil { // Check the error
			fmt.Printf("Error processing item %d (%s): %v\n", i+1, containerMeta.ContainerInfo, processErr)
			// Store the first error encountered
			if firstErr == nil {
				firstErr = processErr
			}
			// Continue processing other items even if one fails
		}
	}
	// Return the first error encountered, or nil if all succeeded
	return firstErr
}

// processPlaylist downloads all tracks for a playlist.
// (Refactored from playlist in main.go)
func (d *Downloader) processPlaylist(jobID string, plistId, legacyToken string, isCatalogPlist bool, streamParams *StreamParams) error {
	// Playlist requires user email from config
	email := d.Config.Email
	meta, err := d.getPlistMeta(plistId, email, legacyToken, isCatalogPlist)
	if err != nil {
		return fmt.Errorf("failed to get metadata for playlist %s: %w", plistId, err)
	}
	if meta.Response == nil || len(meta.Response.Items) == 0 {
		return fmt.Errorf("playlist %s is empty or returned no data", plistId)
	}

	plistName := meta.Response.PlayListName
	fmt.Println("Playlist:", plistName)

	// Update job title with the playlist name
	d.QueueMgr.UpdateJobTitle(jobID, plistName)

	plistPath := filepath.Join(d.Config.OutPath, SanitizeFilename(plistName))
	err = MakeDirs(plistPath) // TODO: Move MakeDirs to utils
	if err != nil {
		return fmt.Errorf("failed to create playlist folder %s: %w", plistPath, err)
	}

	trackTotal := len(meta.Response.Items)
	for i, item := range meta.Response.Items {
		trackNum := i + 1
		err := d.processTrack(jobID, plistPath, trackNum, trackTotal, &item.Track, streamParams)
		if err != nil {
			fmt.Printf("Error processing track %d (%s) in playlist: %v\n", trackNum, item.Track.SongTitle, err)
			// Optionally collect errors
		}
	}
	return nil // Or return collected errors
}

// --- Helper function to extract artwork ---
func extractArtworkUrl(meta *AlbArtResp) string {
	if meta == nil {
		return ""
	}
	// Prioritize specific image fields if they exist
	if meta.VodPlayerImage != "" { // Seems relevant for videos
		return meta.VodPlayerImage
	}
	if meta.CoverImage != nil {
		if coverUrl, ok := meta.CoverImage.(string); ok && coverUrl != "" {
			return coverUrl
		}
	}
	// Fallback to Img structure
	if meta.Img.URL != "" {
		return meta.Img.URL
	}
	// Fallback to Pics array
	if len(meta.Pics) > 0 && meta.Pics[0].URL != "" {
		return meta.Pics[0].URL
	}
	return "" // No artwork found
}

// Utility functions moved to utils.go
// func SanitizeFilename(name string) string { ... }
// func FileExists(path string) (bool, error) { ... }
// func MakeDirs(path string) error { ... }
