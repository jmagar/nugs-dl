package downloader

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"nugs-dl/pkg/api"

	"github.com/grafov/m3u8"
)

// Constants related to HLS processing
const (
	// Regex to extract bitrate from HLS variant filenames
	bitrateRegex = `[\w]+(?:_(\d+)k_v\d+)`
	tempEncFile  = "temp_enc.ts" // Temporary file for encrypted segment
)

// --- HLS Specific Functions ---

// getManifestBase extracts the base URL path and query from a full manifest URL.
// (Moved from main.go)
func getManifestBase(manifestUrl string) (string, string, error) {
	u, err := url.Parse(manifestUrl)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse manifest URL %s: %w", manifestUrl, err)
	}
	path := u.Path
	lastPathIdx := strings.LastIndex(path, "/")
	if lastPathIdx == -1 {
		return "", "", fmt.Errorf("could not find path separator in manifest URL %s", manifestUrl)
	}
	base := u.Scheme + "://" + u.Host + path[:lastPathIdx+1]
	return base, "?" + u.RawQuery, nil
}

// extractBitrate extracts the bitrate number from an HLS variant URI.
// (Moved from main.go)
func extractBitrate(manUrl string) string {
	regex := regexp.MustCompile(bitrateRegex)
	match := regex.FindStringSubmatch(manUrl)
	if match != nil && len(match) > 1 {
		return match[1]
	}
	return ""
}

// parseHlsMaster parses the master HLS playlist, finds the best quality variant,
// updates the Quality struct with the variant URL and bitrate specs.
// (Moved from main.go)
func (d *Downloader) parseHlsMaster(qual *Quality) error {
	req, err := d.HTTPClient.Get(qual.URL)
	if err != nil {
		return fmt.Errorf("failed to GET HLS master playlist %s: %w", qual.URL, err)
	}
	defer req.Body.Close()
	if req.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status for HLS master playlist %s: %s", qual.URL, req.Status)
	}

	playlist, listType, err := m3u8.DecodeFrom(req.Body, true)
	if err != nil {
		return fmt.Errorf("failed to decode HLS master playlist %s: %w", qual.URL, err)
	}
	if listType != m3u8.MASTER {
		return fmt.Errorf("expected HLS master playlist but got media playlist for %s", qual.URL)
	}

	master := playlist.(*m3u8.MasterPlaylist)
	if len(master.Variants) == 0 {
		return fmt.Errorf("HLS master playlist %s contains no variants", qual.URL)
	}
	// Sort variants by bandwidth (highest first)
	sort.Slice(master.Variants, func(x, y int) bool {
		return master.Variants[x].Bandwidth > master.Variants[y].Bandwidth
	})

	bestVariant := master.Variants[0]
	variantUri := bestVariant.URI
	bitrate := extractBitrate(variantUri)
	if bitrate == "" {
		// Attempt to get bitrate from Bandwidth field if regex fails
		if bestVariant.Bandwidth > 0 {
			bitrate = fmt.Sprintf("%d", bestVariant.Bandwidth/1000)
		}
	}

	if bitrate == "" {
		fmt.Printf("Warning: could not determine bitrate for HLS variant %s\n", variantUri)
		qual.Specs = "AAC (Unknown Bitrate)" // Fallback spec
	} else {
		qual.Specs = bitrate + " Kbps AAC"
	}

	manBase, query, err := getManifestBase(qual.URL)
	if err != nil {
		return err // Error already formatted by getManifestBase
	}

	// Construct the full URL for the chosen variant's media playlist
	qual.URL = manBase + variantUri + query
	return nil
}

// getKey fetches the decryption key from the key URL specified in the HLS manifest.
// (Moved from main.go)
func (d *Downloader) getKey(keyUrl string) ([]byte, error) {
	req, err := d.HTTPClient.Get(keyUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to GET HLS key %s: %w", keyUrl, err)
	}
	defer req.Body.Close()
	if req.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status for HLS key %s: %s", keyUrl, req.Status)
	}
	// Key should be 16 bytes for AES-128
	buf := make([]byte, 16)
	_, err = io.ReadFull(req.Body, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read HLS key from %s: %w", keyUrl, err)
	}
	return buf, nil
}

// decryptTrack decrypts the AES-128 encrypted data (typically a single TS segment for HLS audio).
// Uses CBC mode as indicated by the manifest standard (IV is provided).
// (Moved from main.go - Reads from temp file, returns decrypted bytes)
func decryptTrack(key, iv []byte) ([]byte, error) {
	encData, err := os.ReadFile(tempEncFile) // Read from temporary file
	if err != nil {
		return nil, fmt.Errorf("failed to read encrypted temp file %s: %w", tempEncFile, err)
	}
	os.Remove(tempEncFile) // Clean up temp file immediately after reading

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	if len(encData)%aes.BlockSize != 0 {
		return nil, errors.New("encrypted data is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(encData))

	fmt.Println("Decrypting HLS segment...")
	mode.CryptBlocks(decrypted, encData)

	// Remove PKCS#7 padding (common for AES CBC)
	decrypted, err = pkcs7Unpad(decrypted, aes.BlockSize)
	if err != nil {
		return nil, fmt.Errorf("failed to unpad decrypted data: %w", err)
	}

	return decrypted, nil
}

// pkcs7Unpad removes PKCS#7 padding.
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("pkcs7: data is empty")
	}
	if length%blockSize != 0 {
		return nil, errors.New("pkcs7: data is not block-aligned")
	}
	padLen := int(data[length-1])
	if padLen == 0 || padLen > blockSize {
		// Treat as not padded, or invalid padding
		return data, nil // Or return error? Depending on expected input.
	}
	// Check padding bytes
	for i := length - padLen; i < length; i++ {
		if data[i] != byte(padLen) {
			return data, nil // Treat as not padded if bytes don't match
		}
	}
	return data[:length-padLen], nil
}

// tsToAac uses FFmpeg to losslessly extract the AAC audio from a decrypted TS container.
// (Moved from main.go)
func tsToAac(decData []byte, outPath, ffmpegNameStr string) error {
	// TODO: Move ffmpeg execution to ffmpeg.go
	fmt.Println("Remuxing TS to AAC container...")
	var errBuffer bytes.Buffer
	cmd := exec.Command(ffmpegNameStr, "-i", "pipe:0", "-c:a", "copy", "-vn", "-y", outPath) // pipe:0 specifies stdin
	cmd.Stdin = bytes.NewReader(decData)
	cmd.Stderr = &errBuffer
	err := cmd.Run()
	if err != nil {
		errString := fmt.Sprintf("ffmpeg remux failed: %s\nOutput:\n%s", err, errBuffer.String())
		return errors.New(errString)
	}
	return nil
}

// downloadHls handles the full process for HLS-only audio tracks.
// (Refactored from hlsOnly in main.go)
func (d *Downloader) downloadHls(jobID, trackPath, masterPlaylistUrl string) error {
	// 1. Parse the master playlist to get the media playlist URL for the best variant
	fmt.Println("Parsing HLS master playlist...")
	qual := &Quality{URL: masterPlaylistUrl} // Create a temporary quality struct
	err := d.parseHlsMaster(qual)
	if err != nil {
		return fmt.Errorf("failed to parse HLS master playlist: %w", err)
	}
	mediaPlaylistUrl := qual.URL
	fmt.Printf("Selected HLS variant: %s (URL: %s)\n", qual.Specs, mediaPlaylistUrl)

	// Send initial "Starting download" progress update
	d.sendProgress(api.ProgressUpdate{
		JobID:       jobID,
		Message:     "Starting HLS download...",
		CurrentFile: filepath.Base(trackPath),
	})

	// 2. Fetch and parse the media playlist
	req, err := d.HTTPClient.Get(mediaPlaylistUrl)
	if err != nil {
		return fmt.Errorf("failed to GET HLS media playlist %s: %w", mediaPlaylistUrl, err)
	}
	defer req.Body.Close()
	if req.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status for HLS media playlist %s: %s", mediaPlaylistUrl, req.Status)
	}

	playlist, listType, err := m3u8.DecodeFrom(req.Body, true)
	if err != nil {
		return fmt.Errorf("failed to decode HLS media playlist %s: %w", mediaPlaylistUrl, err)
	}
	if listType != m3u8.MEDIA {
		return fmt.Errorf("expected HLS media playlist but got master for %s", mediaPlaylistUrl)
	}

	media := playlist.(*m3u8.MediaPlaylist)
	if media.Key == nil || media.Key.URI == "" {
		return errors.New("HLS media playlist does not contain encryption key info")
	}
	if len(media.Segments) == 0 || media.Segments[0] == nil {
		return errors.New("HLS media playlist contains no segments")
	}

	manBase, query, err := getManifestBase(mediaPlaylistUrl)
	if err != nil {
		return err
	}

	// 3. Fetch Key
	keyUri := media.Key.URI
	if !strings.HasPrefix(keyUri, "http") {
		keyUri = manBase + keyUri
	}
	fmt.Println("Fetching HLS key...")
	keyBytes, err := d.getKey(keyUri)
	if err != nil {
		return fmt.Errorf("failed to get HLS key from %s: %w", keyUri, err)
	}

	// 4. Decode IV
	ivString := media.Key.IV
	if ivString == "" || !strings.HasPrefix(ivString, "0x") {
		return errors.New("HLS key IV is missing or invalid format")
	}
	iv, err := hex.DecodeString(ivString[2:])
	if err != nil {
		return fmt.Errorf("failed to decode IV hex string %s: %w", ivString, err)
	}
	if len(iv) != 16 {
		return fmt.Errorf("decoded IV is not 16 bytes: %d bytes", len(iv))
	}

	// 5. Download First Segment (assuming audio only needs one)
	segmentUri := media.Segments[0].URI
	segmentUrl := manBase + segmentUri + query
	fmt.Println("Downloading encrypted HLS segment...")
	// Before download segment:
	d.sendProgress(api.ProgressUpdate{JobID: jobID, Message: "Downloading HLS segment..."})
	err = d.downloadFile(jobID, tempEncFile, segmentUrl) // Pass jobID here
	if err != nil {
		return fmt.Errorf("failed to download HLS segment %s: %w", segmentUrl, err)
	}

	// 6. Decrypt
	decData, err := decryptTrack(keyBytes, iv)
	if err != nil {
		os.Remove(tempEncFile) // Clean up temp file on decrypt error
		return fmt.Errorf("failed to decrypt HLS segment: %w", err)
	}

	// 7. Remux using FFmpeg
	ffmpegCmd := d.getFfmpegCmd()
	// Before remux:
	d.sendProgress(api.ProgressUpdate{JobID: jobID, Message: "Remuxing HLS segment..."})
	err = tsToAac(decData, trackPath, ffmpegCmd)
	if err != nil {
		return fmt.Errorf("failed to remux HLS segment to AAC: %w", err)
	}

	// Final Success:
	d.sendProgress(api.ProgressUpdate{
		JobID:      jobID,
		Message:    "HLS track processed successfully.",
		Percentage: 100.0, // Mark 100%
	})
	return nil
}

// --- HLS Video specific functions (To be moved/merged with ffmpeg/video logic) ---

// getSegUrls parses a media playlist to get all segment URIs.
// (Moved from main.go, used by video download)
func (d *Downloader) getSegUrls(mediaPlaylistUrl, query string) ([]string, error) {
	var segUrls []string
	req, err := d.HTTPClient.Get(mediaPlaylistUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to GET HLS media playlist %s: %w", mediaPlaylistUrl, err)
	}
	defer req.Body.Close()
	if req.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status for HLS media playlist %s: %s", mediaPlaylistUrl, req.Status)
	}

	playlist, listType, err := m3u8.DecodeFrom(req.Body, true)
	if err != nil {
		return nil, fmt.Errorf("failed to decode HLS media playlist %s: %w", mediaPlaylistUrl, err)
	}
	if listType != m3u8.MEDIA {
		return nil, fmt.Errorf("expected HLS media playlist but got master for %s", mediaPlaylistUrl)
	}

	media := playlist.(*m3u8.MediaPlaylist)
	for _, seg := range media.Segments {
		if seg == nil {
			break
		}
		segUrls = append(segUrls, seg.URI+query)
	}
	if len(segUrls) == 0 {
		return nil, errors.New("HLS media playlist contained no segments")
	}
	return segUrls, nil
}
