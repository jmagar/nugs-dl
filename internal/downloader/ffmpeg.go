package downloader

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"nugs-dl/internal/logger" // Import the logger package
)

// Constants related to FFmpeg interaction
const (
	durRegex       = `Duration: ([\d:.]+)`      // Regex to extract duration from ffmpeg output
	chapsFileFname = "chapters_nugs_dl_tmp.txt" // Temp filename for chapter metadata
)

// --- FFmpeg Interaction Functions ---

// getFfmpegCmd determines the correct path/command for ffmpeg based on config.
// TODO: Needs a reliable way to get script/binary dir if not using PATH.
func (d *Downloader) getFfmpegCmd() string {
	if d.Config.UseFfmpegEnvVar {
		return "ffmpeg"
	}
	// Placeholder - assumes ffmpeg is in PATH even if config says otherwise
	// Needs fix when refactoring how script dir is found.
	logger.Warn("Cannot determine relative ffmpeg path, assuming ffmpeg is in system PATH.", "useFfmpegEnvVar", d.Config.UseFfmpegEnvVar)
	return "ffmpeg"
	// return "./ffmpeg" // Original alternative
}

// extractDuration parses ffmpeg's stderr output to find the duration.
// (Moved from main.go)
func extractDuration(errStr string) string {
	regex := regexp.MustCompile(durRegex)
	match := regex.FindStringSubmatch(errStr)
	if match != nil && len(match) > 1 {
		return match[1]
	}
	return ""
}

// parseDuration converts duration string (HH:MM:SS.ms) to total seconds (int).
// (Moved from main.go)
func parseDuration(dur string) (int, error) {
	// Handle potential format variations if needed
	dur = strings.Replace(dur, ":", "h", 1)
	dur = strings.Replace(dur, ":", "m", 1)
	dur = strings.Replace(dur, ".", "s", 1)
	dur += "ms"

	parsedDur, err := time.ParseDuration(dur)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration string '%s': %w", dur, err)
	}
	// Round to nearest second
	roundedSeconds := math.Round(parsedDur.Seconds())
	return int(roundedSeconds), nil
}

// getDuration runs ffmpeg on a TS file to extract its duration in seconds.
// (Moved from main.go)
func (d *Downloader) getDuration(tsPath string) (int, error) {
	ffmpegCmd := d.getFfmpegCmd()
	var errBuffer bytes.Buffer
	// Run ffmpeg with -i input, no output file needed, just stderr parsing
	args := []string{"-hide_banner", "-i", tsPath}
	cmd := exec.Command(ffmpegCmd, args...)
	cmd.Stderr = &errBuffer

	// FFmpeg usually exits with status 1 when only input is specified
	err := cmd.Run()
	errStr := errBuffer.String()
	if err != nil && err.Error() != "exit status 1" {
		// Unexpected error other than the expected exit code 1
		return 0, fmt.Errorf("ffmpeg duration check failed: %w\nOutput:\n%s", err, errStr)
	}

	// Check if stderr contains the expected message when no output is given
	if !strings.Contains(errStr, "At least one output file must be specified") {
		// If the expected message isn't there, something else might be wrong
		logger.Warn("Unexpected ffmpeg output during duration check. Will attempt to parse duration anyway.", "ffmpegOutput", errStr, "tsPath", tsPath)
		// Continue trying to parse duration anyway, it might be there
	}

	durString := extractDuration(errStr)
	if durString == "" {
		return 0, fmt.Errorf("could not extract duration from ffmpeg output:\n%s", errStr)
	}

	durSecs, err := parseDuration(durString)
	if err != nil {
		return 0, fmt.Errorf("failed to parse extracted duration string '%s': %w", durString, err)
	}
	return durSecs, nil
}

// getNextChapStart finds the start time of the next chapter.
// (Helper for writeChapsFile, moved from main.go)
func getNextChapStart(chapters []interface{}, currentIndex int) float64 {
	nextIndex := currentIndex + 1
	if nextIndex < len(chapters) {
		if chapterMap, ok := chapters[nextIndex].(map[string]interface{}); ok {
			if startTime, ok := chapterMap["chapterSeconds"].(float64); ok {
				return startTime
			}
		}
	}
	return -1 // Indicate no next chapter or error
}

// writeChapsFile creates the metadata file used by ffmpeg to embed chapters.
// (Moved from main.go)
func writeChapsFile(chapters []interface{}, durationSeconds int) error {
	f, err := os.OpenFile(chapsFileFname, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create chapter file %s: %w", chapsFileFname, err)
	}
	defer f.Close()

	// Write FFMETADATA header
	_, err = f.WriteString(";FFMETADATA1\n")
	if err != nil {
		return fmt.Errorf("failed to write chapter file header: %w", err)
	}

	for i, chapter := range chapters {
		chapterMap, ok := chapter.(map[string]interface{})
		if !ok {
			logger.Warn("Skipping invalid chapter data, expected map[string]interface{}.", "chapterIndex", i, "chapterData", chapter)
			continue
		}

		startSeconds, startOk := chapterMap["chapterSeconds"].(float64)
		chapterName, nameOk := chapterMap["chaptername"].(string)
		if !startOk || !nameOk {
			logger.Warn("Skipping chapter due to missing 'chapterSeconds' or 'chaptername'.", "chapterIndex", i, "startOk", startOk, "nameOk", nameOk, "chapterData", chapterMap)
			continue
		}

		startRounded := int(math.Round(startSeconds))
		endRounded := durationSeconds // Default end is video duration

		// Find end time (start of next chapter - 1, or video duration for last chapter)
		nextChapStart := getNextChapStart(chapters, i)
		if nextChapStart >= 0 && nextChapStart > startSeconds { // Ensure next chapter is after current
			endRounded = int(math.Round(nextChapStart)) - 1
		}
		// Ensure end time is not before start time
		if endRounded < startRounded {
			endRounded = startRounded // Set end=start if calculation is weird
		}
		// Ensure end time doesn't exceed video duration
		if endRounded > durationSeconds {
			endRounded = durationSeconds
		}

		// Write chapter block
		_, err = fmt.Fprintf(f, "\n[CHAPTER]\nTIMEBASE=1/1\nSTART=%d\nEND=%d\nTITLE=%s\n",
			startRounded,
			endRounded,
			chapterName, // TODO: Escape title if needed?
		)
		if err != nil {
			return fmt.Errorf("failed to write chapter %d data: %w", i, err)
		}
	}
	logger.Info("FFmpeg chapter metadata file created successfully.", "filename", chapsFileFname)
	return nil
}

// tsToMp4 remuxes a TS file (downloaded video segments) to an MP4 container,
// optionally embedding chapter metadata.
// (Moved from main.go)
func (d *Downloader) tsToMp4(tsInputPath, mp4OutputPath string, chaptersAvailable bool) error {
	ffmpegCmd := d.getFfmpegCmd()
	var errBuffer bytes.Buffer
	args := []string{"-hide_banner", "-i", tsInputPath}

	if chaptersAvailable {
		// Check if chapter file exists first
		if _, err := os.Stat(chapsFileFname); err == nil {
			args = append(args, "-f", "ffmetadata", "-i", chapsFileFname, "-map_metadata", "1")
		} else {
			logger.Warn("Chapter metadata file not found, skipping chapter embedding.", "expectedFile", chapsFileFname, "inputTS", tsInputPath)
			// Reset flag so we don't try to delete it later
			chaptersAvailable = false
		}
	}
	args = append(args, "-c", "copy", "-y", mp4OutputPath) // Copy streams, overwrite output

	cmd := exec.Command(ffmpegCmd, args...)
	cmd.Stderr = &errBuffer

	logger.Info("Executing FFmpeg remux command", "command", ffmpegCmd, "arguments", args)
	err := cmd.Run()
	if err != nil {
		// Clean up potentially incomplete MP4 file on error
		os.Remove(mp4OutputPath)
		return fmt.Errorf("ffmpeg remux failed: %w\nOutput:\n%s", err, errBuffer.String())
	}

	// --- Cleanup ---
	// Delete the raw TS file after successful remux
	err = os.Remove(tsInputPath)
	if err != nil {
		logger.Warn("Failed to delete temporary TS file after remux.", "file", tsInputPath, "error", err)
	}
	// Delete the chapter file if it was used
	if chaptersAvailable {
		err = os.Remove(chapsFileFname)
		if err != nil {
			logger.Warn("Failed to delete temporary chapter metadata file.", "file", chapsFileFname, "error", err)
		}
	}

	logger.Info("FFmpeg remux completed successfully.", "outputFile", mp4OutputPath)
	return nil
}
