package queue

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid" // Using UUID for job IDs

	// Import the shared API types
	"nugs-dl/pkg/api"
)

// QueueManager manages the download jobs.
type QueueManager struct {
	jobs  []*api.DownloadJob // Simple slice for the queue for now
	mutex sync.RWMutex       // Mutex to protect concurrent access to the jobs slice
}

// NewQueueManager creates a new queue manager instance.
func NewQueueManager() *QueueManager {
	return &QueueManager{
		jobs: make([]*api.DownloadJob, 0),
	}
}

// AddJob creates a new DownloadJob for a single URL, adds it to the queue, and returns it.
func (qm *QueueManager) AddJob(url string, opts api.DownloadOptions) (*api.DownloadJob, error) {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	// Check for existing active/queued job with the same URL
	for _, existingJob := range qm.jobs {
		if existingJob.OriginalUrl == url &&
			(existingJob.Status == api.StatusQueued || existingJob.Status == api.StatusProcessing) {
			return nil, fmt.Errorf("job for URL %s already exists in queue (ID: %s, Status: %s)", url, existingJob.ID, existingJob.Status)
		}
	}

	job := &api.DownloadJob{
		ID:          uuid.NewString(),
		OriginalUrl: url, // Store single URL
		Options:     opts,
		Status:      api.StatusQueued,
		CreatedAt:   time.Now().UTC(),
		Progress:    0,
	}

	qm.jobs = append(qm.jobs, job)
	fmt.Printf("[QueueManager] Job added: ID=%s, URL=%s\n", job.ID, url)
	return job, nil
}

// GetJob retrieves a specific job by its ID.
func (qm *QueueManager) GetJob(jobID string) (*api.DownloadJob, bool) {
	qm.mutex.RLock()
	defer qm.mutex.RUnlock()

	for _, job := range qm.jobs {
		if job.ID == jobID {
			// Return a copy to prevent modification of the internal slice item?
			// For now, return pointer directly, assuming read-only use or subsequent updates via UpdateJobStatus.
			return job, true
		}
	}
	return nil, false
}

// GetAllJobs returns a slice containing all current jobs.
// Returns a copy of the slice to ensure thread safety.
func (qm *QueueManager) GetAllJobs() []*api.DownloadJob {
	qm.mutex.RLock()
	defer qm.mutex.RUnlock()

	// Return a copy of the slice to prevent race conditions if the caller modifies it
	jobsCopy := make([]*api.DownloadJob, len(qm.jobs))
	copy(jobsCopy, qm.jobs)
	return jobsCopy
}

// GetCompletedJobs returns a slice containing only completed jobs.
// Returns a copy of the slice to ensure thread safety.
func (qm *QueueManager) GetCompletedJobs() []*api.DownloadJob {
	qm.mutex.RLock()
	defer qm.mutex.RUnlock()

	// Initialize with empty slice to ensure JSON serializes as [] not null
	completedJobs := make([]*api.DownloadJob, 0)
	for _, job := range qm.jobs {
		if job.Status == api.StatusComplete {
			completedJobs = append(completedJobs, job)
		}
	}
	return completedJobs
}

// UpdateJobStatus updates the status and optionally the error message of a job.
// It also sets StartedAt and CompletedAt timestamps appropriately.
func (qm *QueueManager) UpdateJobStatus(jobID string, status api.JobStatus, errMsg string) bool {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	for _, job := range qm.jobs {
		if job.ID == jobID {
			now := time.Now().UTC()
			job.Status = status
			job.ErrorMessage = errMsg

			if status == api.StatusProcessing && job.StartedAt == nil {
				job.StartedAt = &now
			}
			if (status == api.StatusComplete || status == api.StatusFailed) && job.CompletedAt == nil {
				job.CompletedAt = &now
				if status == api.StatusComplete {
					job.Progress = 100 // Ensure progress is 100 on completion
				}
			}
			fmt.Printf("[QueueManager] Job status updated: ID=%s, Status=%s\n", job.ID, status)
			return true
		}
	}
	fmt.Printf("[QueueManager] Failed to update status for unknown job ID: %s\n", jobID)
	return false
}

// GetNextJob finds the first job with 'queued' status, marks it as 'processing',
// and returns it. Returns nil, false if no queued jobs are found.
// This is a simple FIFO implementation.
func (qm *QueueManager) GetNextJob() (*api.DownloadJob, bool) {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	for _, job := range qm.jobs {
		if job.Status == api.StatusQueued {
			now := time.Now().UTC()
			job.Status = api.StatusProcessing
			job.StartedAt = &now
			fmt.Printf("[QueueManager] Picking next job: ID=%s\n", job.ID)
			// Return pointer to the job in the slice - worker needs to update it
			return job, true
		}
	}

	return nil, false // No queued jobs found
}

// UpdateJobArtwork updates the artwork URL for a specific job.
func (qm *QueueManager) UpdateJobArtwork(jobID string, artworkURL string) bool {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	for _, job := range qm.jobs {
		if job.ID == jobID {
			job.ArtworkURL = artworkURL
			// Optional: Log this update?
			// fmt.Printf("[QueueManager] Artwork updated for Job ID: %s\n", jobID)
			return true
		}
	}
	fmt.Printf("[QueueManager] Failed to update artwork for unknown job ID: %s\n", jobID)
	return false
}

// UpdateJobTitle updates the title for a specific job.
func (qm *QueueManager) UpdateJobTitle(jobID string, title string) bool {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	for _, job := range qm.jobs {
		if job.ID == jobID {
			job.Title = title
			fmt.Printf("[QueueManager] Title updated for Job ID: %s -> %s\n", jobID, title)
			return true
		}
	}
	fmt.Printf("[QueueManager] Failed to update title for unknown job ID: %s\n", jobID)
	return false
}

// UpdateJobProgress updates the progress, speed, current file, and track information for a specific job.
func (qm *QueueManager) UpdateJobProgress(jobID string, progress float64, speedBps int64, currentFile string, currentTrack, totalTracks int) bool {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	for _, job := range qm.jobs {
		if job.ID == jobID {
			job.Progress = progress
			job.SpeedBPS = speedBps
			if currentFile != "" {
				job.CurrentFile = currentFile
			}
			if currentTrack > 0 {
				job.CurrentTrack = currentTrack
			}
			if totalTracks > 0 {
				job.TotalTracks = totalTracks
			}
			fmt.Printf("[QueueManager] Progress updated for Job ID: %s -> %.1f%% (Track %d/%d, Speed: %d B/s)\n", 
				jobID, progress, currentTrack, totalTracks, speedBps)
			return true
		}
	}
	fmt.Printf("[QueueManager] Failed to update progress for unknown job ID: %s\n", jobID)
	return false
}

// RemoveJob removes a job from the queue by its ID.
// Returns true if the job was found and removed, false otherwise.
// Only allows removal if job is in Queued, Failed, or Complete status.
func (qm *QueueManager) RemoveJob(jobID string) bool {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	removeIndex := -1
	for i, job := range qm.jobs {
		if job.ID == jobID {
			// Check if job is in a removable state
			if job.Status == api.StatusQueued || job.Status == api.StatusFailed || job.Status == api.StatusComplete {
				removeIndex = i
				break
			} else {
				fmt.Printf("[QueueManager] Cannot remove job %s in state %s\n", jobID, job.Status)
				return false // Cannot remove job in this state
			}
		}
	}

	if removeIndex == -1 {
		fmt.Printf("[QueueManager] Failed to remove job: ID %s not found\n", jobID)
		return false // Job not found
	}

	// Remove the element at removeIndex
	// https://github.com/golang/go/wiki/SliceTricks#delete
	qm.jobs = append(qm.jobs[:removeIndex], qm.jobs[removeIndex+1:]...)

	fmt.Printf("[QueueManager] Job removed: ID=%s\n", jobID)
	return true
}

// TODO: Add functions for updating progress, current file, speed, artwork URL.
// func (qm *QueueManager) UpdateJobProgress(...) bool
// func (qm *QueueManager) RemoveJob(...) bool // Optional: for cleanup
