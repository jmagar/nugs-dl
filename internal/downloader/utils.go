package downloader

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"nugs-dl/internal/logger" // Import the logger package
	// "runtime" // No longer needed here
	// "path/filepath" // No longer needed here
)

// Constants related to utilities
const (
	sanRegexStr = `[\/:*?"<>|]` // Filename sanitization regex
)

// regexStrings defines the URL patterns to identify content types.
// (Moved from main.go)
var regexStrings = [12]string{
	`^https://play.nugs.net/release/(\d+)$`,                     // 0: Album/Show/Video Release
	`^https://play.nugs.net/#/playlists/playlist/(\d+)$`,        // 1: User Playlist (hash)
	`^https://play.nugs.net/library/playlist/(\d+)$`,            // 2: User Playlist (library)
	`(^https://2nu.gs/[a-zA-Z\d]+$)`,                            // 3: Catalog Playlist (shortened URL) - Needs resolution
	`^https://play.nugs.net/#/videos/artist/\d+/.+/(\d+)$`,      // 4: Video Container (hash)
	`^https://play.nugs.net/artist/(\d+)(?:/albums|/latest|)$`,  // 5: Artist Page
	`^https://play.nugs.net/livestream/(\d+)/exclusive$`,        // 6: Exclusive Livestream (needs meta)
	`^https://play.nugs.net/watch/livestreams/exclusive/(\d+)$`, // 7: Exclusive Livestream (watch link)
	`^https://play.nugs.net/#/my-webcasts/\d+-(\d+)-\d+-\d+$`,   // 8: My Webcast (hash, showID)
	`^https://www.nugs.net/on/demandware.store/Sites-NugsNet-Site/default/(?:Stash-QueueVideo|NugsVideo-GetStashVideo)\?([a-zA-Z0-9=%&.-]+$)`, // 9: Purchased Livestream/Video (demandware)
	`^https://play.nugs.net/library/webcast/(\d+)$`,             // 10: My Webcast (library, containerID)
	`^https://play.nugs.net/watch/release/(\d+)$`,               // 11: Watch Release URL (NEW)
}

// UrlType represents the type of nugs.net URL identified.
// Using iota for cleaner enum definition.
type UrlType int

const (
	ReleaseUrl UrlType = iota // 0
	UserPlaylistHashUrl
	UserPlaylistLibUrl
	CatalogPlaylistUrl // Needs resolution
	VideoUrlHash
	ArtistUrl
	ExclusiveLivestreamUrl      // Need to fetch meta to determine if audio/video
	WatchExclusiveLivestreamUrl // Same as above
	MyWebcastHashUrl
	PurchasedUrl
	MyWebcastLibUrl             // This is 10
	WatchReleaseUrl             // This is 11 (NEW)
	UnknownUrl // Sentinel for unknown types // This becomes 12
)

// checkUrl identifies the type of nugs.net URL and extracts the relevant ID.
// (Moved from main.go)
func CheckUrl(url string) (id string, urlType UrlType) {
	// Assume len() handles nil check if needed, range over regexStrings directly
	for i, regexStr := range regexStrings {
		regex := regexp.MustCompile(regexStr)
		match := regex.FindStringSubmatch(url)
		if match != nil && len(match) > 1 {
			// Return the first capture group as the ID and the index as the type
			return match[1], UrlType(i)
		}
	}
	return "", UnknownUrl // Return sentinel value if no match
}

// --- File/Directory Utilities ---

// SanitizeFilename removes characters forbidden in filenames.
// (Moved from main.go)
func SanitizeFilename(filename string) string {
	// Use the pre-compiled regex if possible, or compile here
	sanRegex := regexp.MustCompile(sanRegexStr) // Ensure sanRegexStr constant is defined
	san := sanRegex.ReplaceAllString(filename, "_")
	// Trim trailing tabs as per original code
	// Also trim leading/trailing whitespace which is often problematic
	return strings.TrimSpace(strings.TrimRight(san, "\t"))
}

// FileExists checks if a file exists at the given path.
// (Moved from main.go)
func FileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		return !info.IsDir(), nil // Return true only if it exists and is a file
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil // File does not exist
	}
	// Other error occurred (e.g., permission issue)
	return false, fmt.Errorf("error checking file %s: %w", path, err)
}

// MakeDirs creates all necessary parent directories for a given path.
// (Moved from main.go)
func MakeDirs(path string) error {
	err := os.MkdirAll(path, 0755) // Use standard directory permissions
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

// resolveRedirectURL follows redirects for a given URL (like a shortlink)
// and returns the final destination URL string.
// Adapted from resolveCatPlistId in main.go
func (d *Downloader) resolveRedirectURL(shortUrl string) (string, error) {
	// Create a client that *doesn't* follow redirects automatically
	noRedirectClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Stop after the first response (which should be a redirect)
		},
		Jar:     d.HTTPClient.Jar, // Reuse the main cookie jar if necessary
		Timeout: 10 * time.Second, // Add a timeout
	}

	req, err := http.NewRequest(http.MethodGet, shortUrl, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request for short URL %s: %w", shortUrl, err)
	}
	req.Header.Add("User-Agent", userAgent) // Use standard user agent

	resp, err := noRedirectClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform request for short URL %s: %w", shortUrl, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 300 || resp.StatusCode >= 400 {
		return "", fmt.Errorf("expected redirect for short URL %s, but got status %s", shortUrl, resp.Status)
	}

	finalUrl := resp.Header.Get("Location")
	if finalUrl == "" {
		return "", fmt.Errorf("no Location header found in redirect response for %s", shortUrl)
	}

	// Potentially resolve relative redirects, although Location should be absolute here
	// finalUrlParsed, err := resp.Request.URL.Parse(finalUrl)
	// if err != nil { ... }
	// return finalUrlParsed.String(), nil

	logger.Info("Resolved short URL", "from", shortUrl, "to", finalUrl)
	return finalUrl, nil
}

// // Potential place for getScriptDir, but it relies on runtime.Caller(0)
// // which works based on the *caller's* location, making it hard to use
// // reliably from a library package. Better to determine binary/ffmpeg location
// // during startup in main or cmd/server and pass it via config.
// func getScriptDir() (string, error) { ... }
