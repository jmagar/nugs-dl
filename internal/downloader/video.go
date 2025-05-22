package downloader

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"nugs-dl/pkg/api"

	"github.com/grafov/m3u8"
)

// --- Video Processing Functions ---

// getVidVariant selects the video variant matching the desired resolution.
// (Moved from main.go)
func getVidVariant(variants []*m3u8.Variant, wantRes string) *m3u8.Variant {
	for _, variant := range variants {
		if strings.HasSuffix(variant.Resolution, "x"+wantRes) {
			return variant
		}
	}
	return nil
}

// formatRes converts resolution string (e.g., "1080") to display format (e.g., "1080p", "4K").
// (Moved from main.go)
func formatRes(res string) string {
	if res == "2160" {
		return "4K"
	} else {
		return res + "p"
	}
}

// chooseVariant parses the master video manifest, selects the best matching variant based
// on desired resolution and fallbacks, and returns the chosen variant and its resolution.
// (Moved from main.go)
func (d *Downloader) chooseVariant(manifestUrl, wantRes string) (*m3u8.Variant, string, error) {
	origWantRes := wantRes
	var wantVariant *m3u8.Variant

	req, err := d.HTTPClient.Get(manifestUrl)
	if err != nil {
		return nil, "", fmt.Errorf("failed to GET video master manifest %s: %w", manifestUrl, err)
	}
	defer req.Body.Close()
	if req.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("bad status for video master manifest %s: %s", manifestUrl, req.Status)
	}

	playlist, listType, err := m3u8.DecodeFrom(req.Body, true)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode video master manifest %s: %w", manifestUrl, err)
	}
	if listType != m3u8.MASTER {
		return nil, "", fmt.Errorf("expected video master playlist but got media playlist for %s", manifestUrl)
	}

	master := playlist.(*m3u8.MasterPlaylist)
	if len(master.Variants) == 0 {
		return nil, "", fmt.Errorf("video master playlist %s contains no variants", manifestUrl)
	}

	// Sort by bandwidth DESC (highest first) to prefer higher quality if resolutions match
	sort.Slice(master.Variants, func(x, y int) bool {
		return master.Variants[x].Bandwidth > master.Variants[y].Bandwidth
	})

	// Handle 4K/Best Available request (originally used `resolveRes[5]` which is "2160")
	if wantRes == "2160" {
		bestVariant := master.Variants[0] // Highest bandwidth variant
		// Extract resolution from the best variant found
		if bestVariant.Resolution != "" {
			parts := strings.SplitN(bestVariant.Resolution, "x", 2)
			if len(parts) == 2 {
				actualRes := parts[1]
				fmt.Printf("Highest available resolution is %sp\n", actualRes)
				return bestVariant, formatRes(actualRes), nil
			}
		}
		// Fallback if resolution couldn't be parsed from best variant
		fmt.Println("Could not determine resolution from highest bandwidth variant, falling back to default selection.")
		// Continue to normal selection logic below
	}

	// Find the desired resolution or fallback
	currentTryRes := wantRes
	for {
		wantVariant = getVidVariant(master.Variants, currentTryRes)
		if wantVariant != nil {
			break // Found it
		}
		// Try fallback resolution
		fbRes, hasFallback := resFallback[currentTryRes]
		if !hasFallback {
			break // No more fallbacks to try
		}
		fmt.Printf("Resolution %s unavailable, falling back to %s...\n", formatRes(currentTryRes), formatRes(fbRes))
		currentTryRes = fbRes
	}

	if wantVariant == nil {
		// If no desired or fallback resolution found, maybe pick the best available?
		fmt.Println("Desired/fallback resolutions not found, selecting highest available variant.")
		wantVariant = master.Variants[0]
		parts := strings.SplitN(wantVariant.Resolution, "x", 2)
		if len(parts) == 2 {
			currentTryRes = parts[1]
		} else {
			return nil, "", fmt.Errorf("failed to find any suitable video variant and could not determine resolution of best variant")
		}
	}

	finalResStr := formatRes(currentTryRes)
	if currentTryRes != origWantRes && origWantRes != "2160" { // Don't print if 'best' was requested
		fmt.Printf("Downloading video in %s instead.\n", finalResStr)
	}

	return wantVariant, finalResStr, nil
}

// downloadVideoSegments downloads HLS video segments sequentially to a single file.
// (Refactored from downloadLstream in main.go)
func (d *Downloader) downloadVideoSegments(jobID, videoPath, baseUrl string, segUrls []string) error {
	f, err := os.OpenFile(videoPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create/open video segment file %s: %w", videoPath, err)
	}
	defer f.Close()

	segTotal := len(segUrls)
	if segTotal == 0 {
		return errors.New("no video segments found to download")
	}
	fmt.Printf("Downloading %d video segments...\n", segTotal)
	// Send initial progress update
	d.sendProgress(api.ProgressUpdate{
		JobID:       jobID,
		Message:     fmt.Sprintf("Downloading %d video segments...", segTotal),
		CurrentFile: filepath.Base(videoPath),
	})

	// Use WriteCounter for overall progress (though individual segment sizes aren't known upfront)
	// Instead of using WriteCounter directly on file write (which doesn't track segments),
	// we'll manually send updates in the loop.
	startTime := time.Now().UnixMilli()
	var totalBytesDownloaded int64 = 0
	var lastUpdateTime int64 = 0

	for i, segRelUrl := range segUrls {
		segNum := i + 1

		segUrl := baseUrl + segRelUrl // Construct full segment URL
		req, err := http.NewRequest(http.MethodGet, segUrl, nil)
		if err != nil {
			fmt.Printf("\nError creating request for segment %d (%s): %v\n", segNum, segUrl, err)
			continue // Skip segment on error?
		}
		req.Header.Add("User-Agent", userAgent)

		do, err := d.HTTPClient.Do(req)
		if err != nil {
			fmt.Printf("\nError downloading segment %d (%s): %v\n", segNum, segUrl, err)
			continue // Skip segment on error?
		}

		if do.StatusCode != http.StatusOK {
			status := do.Status
			do.Body.Close()
			fmt.Printf("\nBad status for segment %d (%s): %s\n", segNum, segUrl, status)
			continue // Skip segment on error?
		}

		// Write segment to file and count bytes
		n, err := io.Copy(f, do.Body)
		do.Body.Close()
		if err != nil {
			fmt.Printf("\nError writing segment %d (%s) to file: %v\n", segNum, segUrl, err)
			// Potentially stop entire download here? For now, continue.
		} else {
			totalBytesDownloaded += n
		}

		// Send throttled progress update based on segment number
		now := time.Now().UnixMilli()
		if now-lastUpdateTime >= progressUpdateInterval { // Use same const as WriteCounter
			lastUpdateTime = now
			percentage := float64(segNum) / float64(segTotal) * 100.0
			elapsed := now - startTime
			var speedBps int64 = 0
			if elapsed > 500 && totalBytesDownloaded > 0 {
				speedBps = (totalBytesDownloaded * 1000) / elapsed
			}
			d.sendProgress(api.ProgressUpdate{
				JobID:           jobID,
				Percentage:      percentage,
				BytesDownloaded: totalBytesDownloaded, // Report total bytes so far
				TotalBytes:      -1,                   // Total size is unknown until all segments download
				SpeedBPS:        speedBps,
				Message:         fmt.Sprintf("Segment %d/%d", segNum, segTotal),
				CurrentFile:     filepath.Base(videoPath),
			})
		}
		// Print simple progress to console as fallback
		fmt.Printf("\rSegment %d of %d...", segNum, segTotal)
	}
	fmt.Println("\nFinished downloading segments.")
	// Send final 100% update
	d.sendProgress(api.ProgressUpdate{
		JobID:           jobID,
		Percentage:      100.0,
		Message:         "Segment download complete",
		CurrentFile:     filepath.Base(videoPath),
		BytesDownloaded: totalBytesDownloaded,
		TotalBytes:      totalBytesDownloaded, // Now we know the total
	})
	return nil
}

// processVideo handles the entire video download and processing workflow.
// (Refactored from video in main.go)
func (d *Downloader) processVideo(jobID, videoID, uguID string, opts DownloadOptions, streamParams *StreamParams, preloadedMeta *AlbArtResp, isLstream bool) error {
	var (
		meta *AlbArtResp
		err  error
	)

	// --- Get Metadata ---
	if preloadedMeta != nil {
		meta = preloadedMeta // Use meta passed from album/artist processing
	} else {
		// Fetch fresh metadata if called directly or for livestreams (original behavior)
		albumMeta, err := d.getAlbumMeta(videoID)
		if err != nil {
			return fmt.Errorf("failed to get metadata for video %s: %w", videoID, err)
		}
		if albumMeta.Response == nil {
			return fmt.Errorf("API returned empty response for video %s", videoID)
		}
		meta = albumMeta.Response
	}

	// Extract and update artwork URL
	artworkURL := extractArtworkUrl(meta) // Use helper defined in processing.go
	if artworkURL != "" {
		d.QueueMgr.UpdateJobArtwork(jobID, artworkURL)
	}

	// Determine if chapters are available and wanted
	chapsAvail := false
	if !opts.SkipChapters {
		chapsAvail = !reflect.ValueOf(meta.VideoChapters).IsZero() && len(meta.VideoChapters) > 0
	}

	// --- Get Manifest URL ---
	var manifestUrl string
	skuID := 0
	if isLstream {
		skuID = getLstreamSku(meta.ProductFormatList)
	} else {
		skuID = getVideoSku(meta.Products)
	}
	if skuID == 0 {
		return errors.New("no suitable video product SKU found in metadata")
	}

	if uguID != "" { // Purchased video
		manifestUrl, err = d.getPurchasedManUrl(skuID, videoID, streamParams.UserID, uguID)
	} else { // Streamed video (requires subscription params)
		manifestUrl, err = d.getStreamMeta(meta.ContainerID, skuID, 0, streamParams)
	}
	if err != nil {
		return fmt.Errorf("failed to get video manifest URL: %w", err)
	}
	if manifestUrl == "" {
		return errors.New("API returned an empty video manifest URL")
	}

	// --- Choose Variant (Resolution) ---
	wantRes := resolveRes[d.Config.VideoFormat]
	variant, chosenResStr, err := d.chooseVariant(manifestUrl, wantRes)
	if err != nil {
		return fmt.Errorf("failed to choose video variant: %w", err)
	}

	// --- Prepare Filename and Check Existence ---
	videoFnameBase := meta.ArtistName + " - " + strings.TrimRight(meta.ContainerInfo, " ")
	fmt.Println("Video:", videoFnameBase)

	// Update job title with the video information
	d.QueueMgr.UpdateJobTitle(jobID, videoFnameBase)

	vidPathNoExt := filepath.Join(d.Config.OutPath, SanitizeFilename(videoFnameBase+"_"+chosenResStr))
	vidPathTs := vidPathNoExt + ".ts"   // Path for raw downloaded segments
	vidPathMp4 := vidPathNoExt + ".mp4" // Final output path

	exists, err := FileExists(vidPathMp4) // Check if final MP4 exists
	if err != nil {
		return fmt.Errorf("failed to check if video exists %s: %w", vidPathMp4, err)
	}
	if exists {
		fmt.Println("Video already exists locally.")
		return nil
	}

	// --- Download Video Segments (HLS) ---
	manBaseUrl, query, err := getManifestBase(manifestUrl) // Needs getManifestBase (in hls.go)
	if err != nil {
		return fmt.Errorf("failed to get manifest base URL: %w", err)
	}

	// Construct full URL to the chosen variant's media playlist
	variantMediaPlaylistUrl := variant.URI // Use the variant URI obtained earlier
	fullVariantUrl := manBaseUrl + variantMediaPlaylistUrl + query

	// Get individual segment URLs from the media playlist
	segUrls, err := d.getSegUrls(fullVariantUrl, query) // Use getSegUrls (in hls.go)
	if err != nil {
		return fmt.Errorf("failed to get video segment URLs: %w", err)
	}
	// Call HLS segment download with jobID
	err = d.downloadVideoSegments(jobID, vidPathTs, manBaseUrl, segUrls)

	if err != nil {
		os.Remove(vidPathTs) // Clean up partial TS file on download error
		return fmt.Errorf("failed to download video segments: %w", err)
	}

	// --- Process Chapters (if available) ---
	if chapsAvail {
		fmt.Println("Processing chapters...")
		durationSecs, err := d.getDuration(vidPathTs) // Use method on d
		if err != nil {
			fmt.Printf("Warning: Failed to get video duration for chapters: %v\n", err)
			chapsAvail = false
		} else {
			// writeChapsFile doesn't need the Downloader receiver
			err = writeChapsFile(meta.VideoChapters, durationSecs)
			if err != nil {
				fmt.Printf("Warning: Failed to write chapter file: %v\n", err)
				chapsAvail = false
			}
		}
	}

	// --- Remux TS to MP4 using FFmpeg ---
	fmt.Println("Remuxing video to MP4...")
	// Call tsToMp4 method on d
	err = d.tsToMp4(vidPathTs, vidPathMp4, chapsAvail)
	if err != nil {
		// Error handling is already inside the edit; tsToMp4 cleans up on error
		return fmt.Errorf("failed to remux video to MP4: %w", err)
	}

	// tsToMp4 handles cleanup on success
	fmt.Println("Video processed successfully:", vidPathMp4)
	return nil // Return nil on success
}
