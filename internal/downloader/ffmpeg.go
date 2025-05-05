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
	fmt.Println("Warning: Cannot determine relative ffmpeg path yet, assuming it's in PATH.")
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
		fmt.Printf("Warning: Unexpected ffmpeg output during duration check:\n%s\n", errStr)
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
			fmt.Printf("Warning: Skipping invalid chapter data at index %d\n", i)
			continue
		}

		startSeconds, startOk := chapterMap["chapterSeconds"].(float64)
		chapterName, nameOk := chapterMap["chaptername"].(string)
		if !startOk || !nameOk {
			fmt.Printf("Warning: Skipping chapter with missing data at index %d\n", i)
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
	fmt.Println("Chapter file created successfully.")
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
			fmt.Printf("Warning: Chapter metadata file %s not found, skipping chapter embedding.\n", chapsFileFname)
			// Reset flag so we don't try to delete it later
			chaptersAvailable = false
		}
	}
	args = append(args, "-c", "copy", "-y", mp4OutputPath) // Copy streams, overwrite output

	cmd := exec.Command(ffmpegCmd, args...)
	cmd.Stderr = &errBuffer

	fmt.Println("Executing FFmpeg remux command:", strings.Join(cmd.Args, " "))
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
		fmt.Printf("Warning: Failed to delete temporary TS file %s: %v\n", tsInputPath, err)
	}
	// Delete the chapter file if it was used
	if chaptersAvailable {
		err = os.Remove(chapsFileFname)
		if err != nil {
			fmt.Printf("Warning: Failed to delete temporary chapter file %s: %v\n", chapsFileFname, err)
		}
	}

	fmt.Println("FFmpeg remux completed successfully.")
	return nil
}
