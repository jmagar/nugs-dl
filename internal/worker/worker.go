package worker

import (
	"errors" // Added for errors.Is
	"time"

	"nugs-dl/internal/broadcast"
	"nugs-dl/internal/downloader"
	"nugs-dl/internal/logger" // Import the logger package
	"nugs-dl/internal/queue"
	"nugs-dl/pkg/api"
)

// StartWorker launches a background goroutine to process jobs from the queue.
func StartWorker(qm *queue.QueueManager, dl *downloader.Downloader, hub *broadcast.Hub) {
	logger.Info("[Worker] Starting background queue processor...")

	go func() {
		for {
			// Check for the next available job
			job, found := qm.GetNextJob()

			if !found {
				// No job found, wait a bit before checking again
				time.Sleep(5 * time.Second) // Adjust sleep duration as needed
				continue
			}

			logger.Info("[Worker] Processing job", "jobID", job.ID, "url", job.OriginalUrl)

			// Execute the download logic
			// Pass the whole job object to the Download method
			err := dl.Download(job)

			// Update job status based on the result
			if err != nil {
				if errors.Is(err, downloader.ErrDuplicateCompleted) {
					logger.Info("[Worker] Job is a duplicate of already completed content, skipping.", "jobID", job.ID, "originalError", err.Error())
					// Update job status to Failed with the specific duplicate error message
					qm.UpdateJobStatus(job.ID, api.StatusFailed, err.Error()) // err.Error() will contain the formatted message

					job.Status = api.StatusFailed
					job.ErrorMessage = err.Error() // Use the detailed error message
					now := time.Now().UTC()
					if job.CompletedAt == nil { // Mark as "completed" in terms of processing attempt
						job.CompletedAt = &now
					}
					logger.Debug("[Worker] Broadcasting skipped (duplicate) job status", "jobID", job.ID)
					hub.BroadcastJobStatusUpdate(job) // Broadcast the update
				} else {
					// Handle other general errors
					logger.Error("[Worker] Job failed with a general error", "jobID", job.ID, "error", err)
					qm.UpdateJobStatus(job.ID, api.StatusFailed, err.Error())
					// Update the job object directly and broadcast
					job.Status = api.StatusFailed
					job.ErrorMessage = err.Error()
					now := time.Now().UTC()
					if job.CompletedAt == nil {
						job.CompletedAt = &now
					}
					logger.Debug("[Worker] Broadcasting failed job status", "jobID", job.ID)
					hub.BroadcastJobStatusUpdate(job)
				}
			} else {
				// Handle successful completion
				logger.Info("[Worker] Job completed successfully", "jobID", job.ID)
				qm.UpdateJobStatus(job.ID, api.StatusComplete, "")
				// Update the job object directly and broadcast
				job.Status = api.StatusComplete
				job.Progress = 100 // Ensure progress is 100%
				job.ErrorMessage = ""
				now := time.Now().UTC()
				if job.CompletedAt == nil {
					job.CompletedAt = &now
				}
				logger.Debug("[Worker] Broadcasting completed job status", "jobID", job.ID)
				hub.BroadcastJobStatusUpdate(job)
			}

			// Optional short delay between processing jobs?
			// time.Sleep(1 * time.Second)
		}
	}() // Launch the goroutine
}
