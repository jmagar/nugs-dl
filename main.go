package main

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alexflint/go-arg"

	// Use correct module path
	appConfig "nugs-dl/internal/config"
	// downloader "nugs-dl/internal/downloader" // Keep commented for now
)

// Keep constants only used by main CLI logic
const (
	// devKey, clientId moved to downloader/auth.go
	// layout moved to downloader/auth.go
	// userAgent, userAgentTwo moved to downloader/auth.go or downloader/nugs_api.go
	// authUrl moved to downloader/auth.go
	// streamApiBase - Keep here or move to downloader/nugs_api.go?
	streamApiBase = "https://streamapi.nugs.net/"
	// subInfoUrl moved to downloader/auth.go
	// userInfoUrl moved to downloader/auth.go
	playerUrl      = "https://play.nugs.net/"   // Used by downloadTrack Referer - move?
	sanRegexStr    = `[\/:*?"<>|]`              // Used by sanitise - move to utils?
	chapsFileFname = "chapters_nugs_dl_tmp.txt" // Used by writeChapsFile - move to ffmpeg?
	durRegex       = `Duration: ([\d:.]+)`      // Used by extractDuration - move to ffmpeg?
	bitrateRegex   = `[\w]+(?:_(\d+)k_v\d+)`    // Used by extractBitrate - move to hls?
)

var (
	jar, _ = cookiejar.New(nil) // Keep for http.Client
	// client = &http.Client{Jar: jar} // Let downloader manage its own client?
	// For now, main can still create a client, but downloader might need one passed in.
	client = &http.Client{Jar: jar}
)

// regexStrings moved to internal/downloader/utils.go
// var regexStrings = [11]string{
// 	`^https://play.nugs.net/release/(\d+)$`,
// 	`^https://play.nugs.net/#/playlists/playlist/(\d+)$`,
// 	`^https://play.nugs.net/library/playlist/(\d+)$`,
// 	`(^https://2nu.gs/[a-zA-Z\d]+$)`,
// 	`^https://play.nugs.net/#/videos/artist/\d+/.+/(\d+)$`,
// 	`^https://play.nugs.net/artist/(\d+)(?:/albums|/latest|)$`,
// 	`^https://play.nugs.net/livestream/(\d+)/exclusive$`,
// 	`^https://play.nugs.net/watch/livestreams/exclusive/(\d+)$`,
// 	`^https://play.nugs.net/#/my-webcasts/\d+-(\d+)-\d+-\d+$`,
// 	`^https://www.nugs.net/on/demandware.store/Sites-NugsNet-Site/d` +
// 		`efault/(?:Stash-QueueVideo|NugsVideo-GetStashVideo)\?([a-zA-Z0-9=%&-]+$)`,
// 	`^https://play.nugs.net/library/webcast/(\d+)$`,
// }

// Quality maps moved to downloader/types.go
// var qualityMap = ...
// var resolveRes = ...
// var trackFallback = ...
// var resFallback = ...

// WriteCounter moved to internal/downloader/types.go
// func (wc *WriteCounter) Write(p []byte) (int, error) {
// 	var speed int64 = 0
// 	n := len(p)
// 	wc.Downloaded += int64(n)
// 	percentage := float64(wc.Downloaded) / float64(wc.Total) * float64(100)
// 	wc.Percentage = int(percentage)
// 	toDivideBy := time.Now().UnixMilli() - wc.StartTime
// 	if toDivideBy != 0 {
// 		speed = int64(wc.Downloaded) / toDivideBy * 1000
// 	}
// 	fmt.Printf("\r%d%% @ %s/s, %s/%s ", wc.Percentage, humanize.Bytes(uint64(speed)),
// 		humanize.Bytes(uint64(wc.Downloaded)), wc.TotalStr)
// 	return n, nil
// }

// handleErr can potentially be moved to a shared utils package
func handleErr(errText string, err error, _panic bool) {
	errString := errText + "\n" + err.Error()
	if _panic {
		panic(errString)
	}
	fmt.Println(errString)
}

// wasRunFromSrc, getScriptDir are CLI specific, keep here.
func wasRunFromSrc() bool {
	buildPath := filepath.Join(os.TempDir(), "go-build")
	return strings.HasPrefix(os.Args[0], buildPath)
}
func getScriptDir() (string, error) {
	var (
		ok    bool
		err   error
		fname string
	)
	runFromSrc := wasRunFromSrc()
	if runFromSrc {
		_, fname, _, ok = runtime.Caller(0)
		if !ok {
			return "", errors.New("failed to get script filename")
		}
	} else {
		fname, err = os.Executable()
		if err != nil {
			return "", err
		}
	}
	return filepath.Dir(fname), nil
}

// readTxtFile, contains, processUrls are related to URL/input processing, keep for CLI?
// Or move `processUrls` if the downloader needs to handle txt files itself.
func readTxtFile(path string) ([]string, error) {
	var lines []string
	f, err := os.OpenFile(path, os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return lines, nil
}
func contains(lines []string, value string) bool {
	for _, line := range lines {
		if strings.EqualFold(line, value) {
			return true
		}
	}
	return false
}
func processUrls(urls []string) ([]string, error) {
	var (
		processed []string
		txtPaths  []string
	)
	for _, _url := range urls {
		if strings.HasSuffix(_url, ".txt") && !contains(txtPaths, _url) {
			txtLines, err := readTxtFile(_url)
			if err != nil {
				return nil, err
			}
			for _, txtLine := range txtLines {
				if !contains(processed, txtLine) {
					txtLine = strings.TrimSuffix(txtLine, "/")
					processed = append(processed, txtLine)
				}
			}
			txtPaths = append(txtPaths, _url)
		} else {
			if !contains(processed, _url) {
				_url = strings.TrimSuffix(_url, "/")
				processed = append(processed, _url)
			}
		}
	}
	return processed, nil
}

// --- Config/Arg parsing remains specific to CLI ---
// Config struct definition moved
// Args struct definition remains

// parseCfg remains here for CLI logic
// It will need adjustments if Config struct changes significantly
// func parseCfg() (*Config, error) { ... }
// Note: This will break because Config type is gone. Need replacement.
type CliConfig struct { // Temporary replacement for original Config used by CLI
	Email           string
	Password        string
	Urls            []string
	Format          int
	OutPath         string
	VideoFormat     int
	WantRes         string // Should be derived later
	Token           string
	UseFfmpegEnvVar bool
	FfmpegNameStr   string // Should be derived later
	ForceVideo      bool
	SkipVideos      bool
	SkipChapters    bool
}

// Temporary adapter function to bridge old parseCfg with new internal config
func loadCliConfig() (*CliConfig, error) {
	// Use internal config loader first
	baseCfg, err := appConfig.LoadConfig() // Assuming appConfig is imported
	if err != nil && !os.IsNotExist(err) { // Allow not exist for CLI defaults
		return nil, fmt.Errorf("failed to read base config.json: %w", err)
	}
	if baseCfg == nil {
		// Create a default if file didn't exist
		baseCfg = &appConfig.AppConfig{
			Format: 2, VideoFormat: 3, OutPath: "Nugs downloads", UseFfmpegEnvVar: false,
		}
	}

	args := parseArgs()

	// Apply args overrides to baseCfg fields
	if args.Format != -1 {
		baseCfg.Format = args.Format
	}
	if args.VideoFormat != -1 {
		baseCfg.VideoFormat = args.VideoFormat
	}
	if args.OutPath != "" {
		baseCfg.OutPath = args.OutPath
	}

	// Validation (keep basic validation here for CLI immediate feedback)
	if !(baseCfg.Format >= 1 && baseCfg.Format <= 5) {
		return nil, errors.New("track Format must be between 1 and 5")
	}
	if !(baseCfg.VideoFormat >= 1 && baseCfg.VideoFormat <= 5) {
		return nil, errors.New("video format must be between 1 and 5")
	}

	// Process URLs specific to CLI input
	processedUrls, err := processUrls(args.Urls)
	if err != nil {
		fmt.Println("Failed to process URLs.")
		return nil, err
	}

	// Create the CliConfig struct needed by the rest of the main function
	cliCfg := &CliConfig{
		Email:           baseCfg.Email,
		Password:        baseCfg.Password,
		Urls:            processedUrls,
		Format:          baseCfg.Format,
		OutPath:         baseCfg.OutPath,
		VideoFormat:     baseCfg.VideoFormat,
		Token:           strings.TrimPrefix(baseCfg.Token, "Bearer "), // Apply trim here
		UseFfmpegEnvVar: baseCfg.UseFfmpegEnvVar,
		ForceVideo:      args.ForceVideo,
		SkipVideos:      args.SkipVideos,
		SkipChapters:    args.SkipChapters,
	}

	// Derive helper fields (can be moved later)
	// cliCfg.WantRes = downloader.resolveRes[cliCfg.VideoFormat] // Removed - derivation should happen in downloader
	if cliCfg.UseFfmpegEnvVar {
		cliCfg.FfmpegNameStr = "ffmpeg"
	} else {
		cliCfg.FfmpegNameStr = "./ffmpeg"
	}

	return cliCfg, nil
}

// func readConfig() (*Config, error) { ... } // Replaced by internal/config

func parseArgs() *Args {
	var args Args
	arg.MustParse(&args)
	return &args
}

// --- Utility functions moved to internal/downloader/utils.go ---
// func makeDirs(path string) error { ... }
// func fileExists(path string) (bool, error) { ... }
// func sanitise(filename string) string { ... }

// --- Auth functions moved to internal/downloader/auth.go ---
// func auth(email, pwd string) (string, error) { ... }
// func getUserInfo(token string) (string, error) { ... }
// func getSubInfo(token string) (*SubInfo, error) { ... }
// func getPlan(subInfo *SubInfo) (string, bool) { ... }
// func parseTimestamps(start, end string) (string, string) { ... }
// func parseStreamParams(userId string, subInfo *SubInfo, isPromo bool) *StreamParams { ... }

// --- FFmpeg related functions (moved to internal/downloader/ffmpeg.go) ---
// func extractDuration(errStr string) string { ... }
// func parseDuration(dur string) (int, error) { ... }
// func getDuration(tsPath, ffmpegNameStr string) (int, error) { ... }
// func writeChapsFile(chapters []interface{}, dur int) error { ... }
// func tsToMp4(VidPathTs, vidPath, ffmpegNameStr string, chapAvail bool) error { ... }

// --- URL resolution / Main logic (moved/integrated into downloader) ---
// func resolveCatPlistId(plistUrl string) (string, error) { ... }
// func catalogPlist(_plistId, legacyToken string, cfg *Config, streamParams *StreamParams) error { ... }
// func paidLstream(query, uguID string, cfg *Config, streamParams *StreamParams) error { ... }

// init() remains here for CLI specific init if needed
func init() {
	fmt.Println(`
 _____                ____                _           _         
|   | |_ _ ___ ___   |    \ ___ _ _ _ ___| |___ ___ _| |___ ___ 
| | | | | | . |_ -|  |  |  | . | | | |   | | . | .'| . | -_|  _|
|_|___|___|_  |___|  |____/|___|_____|_|_|_|___|__,|___|___|_|  
	  |___|
`)
}

// main() is the CLI entrypoint, it needs significant changes
func main() {
	fmt.Println("\nNugs Downloader v1.0\n")
	// Replace parseCfg with new loader
	cliCfg, err := loadCliConfig() // Renamed var to cliCfg to avoid shadowing
	if err != nil {
		handleErr("Config/args error:", err, true)
	}

	fmt.Printf("CLI Config loaded: %+v\n", cliCfg)

	// --- This section needs complete replacement ---
	// It should instantiate the downloader and call its methods
	// instead of performing the logic directly here.

	// Example (Conceptual - Downloader API not fully defined yet):

	// 1. Create HTTP Client (reuse existing global client for now)
	// sharedClient := client

	// 2. Load base AppConfig (needed for Downloader)
	appCfg, err := appConfig.LoadConfig()
	if err != nil && !os.IsNotExist(err) { // Handle error, but allow not exist for CLI?
		handleErr("Failed to load base config.json for downloader:", err, true)
	}
	if appCfg == nil {
		// If config doesn't exist, create a default one for the downloader
		appCfg = &appConfig.AppConfig{
			Format: 2, VideoFormat: 3, OutPath: cliCfg.OutPath, UseFfmpegEnvVar: cliCfg.UseFfmpegEnvVar, // Use paths/ffmpeg from CLI args if file missing
		}
	} else {
		// Ensure OutPath uses the potentially overridden one from CLI args
		appCfg.OutPath = cliCfg.OutPath
		appCfg.UseFfmpegEnvVar = cliCfg.UseFfmpegEnvVar
	}

	// 3. Instantiate the downloader service
	// downloaderService := downloader.NewDownloader(appCfg, sharedClient)

	// 4. Create DownloadOptions from cliCfg flags
	// downloadOpts := downloader.DownloadOptions{
	// 	 ForceVideo:   cliCfg.ForceVideo,
	// 	 SkipVideos:   cliCfg.SkipVideos,
	// 	 SkipChapters: cliCfg.SkipChapters,
	// }

	// 5. Call the main download method (when implemented)
	// err = downloaderService.Download(cliCfg.Urls, downloadOpts)
	// if err != nil {
	// 	 handleErr("Download failed:", err, true)
	// }

	// --- Original Main Loop Removed ---
	// fmt.Println("Logging in...")
	// ... original authentication logic ...
	// fmt.Printf("Subscription plan: %s\n\n", planDesc)
	// ... original URL processing loop ...

	fmt.Println("\nFinished (Placeholder - Download logic not yet implemented).")
}

// Remaining utility functions that need moving:

// resolveRes map moved to internal/downloader/types.go
// var resolveRes = map[int]string{
// 	1: "480",
// 	2: "720",
// 	3: "1080",
// 	4: "1440",
// 	5: "2160",
// }

// resFallback map moved to internal/downloader/types.go
// var resFallback = map[string]string{
// 	"720":  "480",
// 	"1080": "720",
// 	"1440": "1080",
// }

// checkUrl moved to internal/downloader/utils.go
// func checkUrl(_url string) (string, int) {
// 	for i, regexStr := range regexStrings {
// 		regex := regexp.MustCompile(regexStr)
// 		match := regex.FindStringSubmatch(_url)
// 		if match != nil {
// 			return match[1], i
// 		}
// 	}
// 	return "", 0
// }

// --- Metadata functions moved ...
// ...

// --- HLS functions moved to internal/downloader/hls.go ---
// ...

// --- Video specific functions (moved to internal/downloader/video.go) ---
// ...
