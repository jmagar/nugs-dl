package api

import (
	"time"
	// Import downloader options if needed directly, or redefine relevant fields
	// downloader "main/internal/downloader"
)

// JobStatus represents the status of a download job.
type JobStatus string

const (
	StatusQueued     JobStatus = "queued"
	StatusProcessing JobStatus = "processing"
	StatusComplete   JobStatus = "complete"
	StatusFailed     JobStatus = "failed"
)

// DownloadOptions mirrors the options needed by the downloader.
// Redefined here to avoid direct dependency cycles if downloader imports api.
type DownloadOptions struct {
	ForceVideo   bool `json:"forceVideo"`
	SkipVideos   bool `json:"skipVideos"`
	SkipChapters bool `json:"skipChapters"`
	// Add format overrides if needed
}

// DownloadJob represents a single download task in the queue.
type DownloadJob struct {
	ID           string          `json:"id"`                     // Unique identifier (e.g., UUID)
	OriginalUrl  string          `json:"originalUrl"`            // Single URL for this job
	Options      DownloadOptions `json:"options"`                // Options for this specific job
	Status       JobStatus       `json:"status"`                 // Current status of the job
	ErrorMessage string          `json:"errorMessage,omitempty"` // Error message if status is Failed
	CreatedAt    time.Time       `json:"createdAt"`              // Timestamp when the job was added
	StartedAt    *time.Time      `json:"startedAt,omitempty"`    // Timestamp when processing started
	CompletedAt  *time.Time      `json:"completedAt,omitempty"`  // Timestamp when completed or failed
	Progress     float64         `json:"progress"`               // Overall progress percentage (0-100)
	CurrentFile  string          `json:"currentFile,omitempty"`  // Name of the file currently being downloaded/processed
	SpeedBPS     int64           `json:"speedBps"`               // Current download speed in Bytes per second
	ArtworkURL   string          `json:"artworkUrl,omitempty"`   // URL for album/video artwork
}

// AddDownloadRequest is the expected request body for adding new download jobs.
// Accepts potentially multiple URLs; handler creates one job per URL.
type AddDownloadRequest struct {
	Urls    []string        `json:"urls" binding:"required,dive,url"`
	Options DownloadOptions `json:"options"`
}

// AddDownloadResponseItem represents the result of adding a single URL.
type AddDownloadResponseItem struct {
	Url   string `json:"url"`
	JobID string `json:"jobId,omitempty"` // Included on success
	Error string `json:"error,omitempty"` // Included on failure to add
}

// ProgressUpdate represents a real-time update on a download job's progress.
type ProgressUpdate struct {
	JobID           string    `json:"jobId"`                 // ID of the job being updated
	Status          JobStatus `json:"status,omitempty"`      // Optional: Can signal status change during progress
	Message         string    `json:"message,omitempty"`     // e.g., "Downloading file X", "Remuxing video"
	CurrentFile     string    `json:"currentFile,omitempty"` // File being processed
	Percentage      float64   `json:"percentage"`            // Overall job percentage (0-100), might be estimated
	SpeedBPS        int64     `json:"speedBps"`              // Current speed in Bytes per second
	BytesDownloaded int64     `json:"bytesDownloaded"`       // Bytes downloaded for the current file/segment
	TotalBytes      int64     `json:"totalBytes"`            // Total bytes for the current file/segment (-1 if unknown)
}

// --- SSE Event Structure ---

// SSEEventType defines the type of event being sent over SSE.
type SSEEventType string

const (
	SSEJobAdded       SSEEventType = "jobAdded"
	SSEProgressUpdate SSEEventType = "progressUpdate"
	// Add other event types later if needed (e.g., jobRemoved)
)

// SSEEvent is a wrapper for data sent over the Server-Sent Events stream.
type SSEEvent struct {
	Type SSEEventType `json:"type"`
	Data interface{}  `json:"data"` // Can hold DownloadJob or ProgressUpdate or other data
}
