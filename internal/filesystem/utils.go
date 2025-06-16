package filesystem

import (
	"fmt"
	"syscall"

	"nugs-dl/internal/logger"
)

// GetAvailableDiskSpace returns the available disk space in bytes for the filesystem
// hosting the given path.
func GetAvailableDiskSpace(path string) (uint64, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		logger.Error("Failed to get filesystem stats", "path", path, "error", err)
		return 0, fmt.Errorf("failed to get filesystem stats for path %s: %w", path, err)
	}

	// Available space in bytes = available blocks * block size
	availableBytes := stat.Bavail * uint64(stat.Bsize)
	logger.Debug("Disk space check", "path", path, "availableBytes", availableBytes, "availableGB", float64(availableBytes)/ (1024*1024*1024))
	return availableBytes, nil
}
