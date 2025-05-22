package worker

import (
	"fmt"
	"time"

	"nugs-dl/internal/broadcast"
	"nugs-dl/internal/downloader"
	"nugs-dl/internal/queue"
	"nugs-dl/pkg/api"
)

// StartWorker launches a background goroutine to process jobs from the queue.
func StartWorker(qm *queue.QueueManager, dl *downloader.Downloader, hub *broadcast.Hub) {
	fmt.Println("[Worker] Starting background queue processor...")

	go func() {
		for {
			// Check for the next available job
			job, found := qm.GetNextJob()

			if !found {
				// No job found, wait a bit before checking again
				time.Sleep(5 * time.Second) // Adjust sleep duration as needed
				continue
			}

			fmt.Printf("[Worker] Processing job ID: %s, URL: %s\n", job.ID, job.OriginalUrl)

			// Execute the download logic
			// Pass the whole job object to the Download method
			err := dl.Download(job)

			// Update job status based on the result
			if err != nil {
				fmt.Printf("[Worker] Job ID %s failed: %v\n", job.ID, err)
				qm.UpdateJobStatus(job.ID, api.StatusFailed, err.Error())
				// Update the job object directly and broadcast
				job.Status = api.StatusFailed
				job.ErrorMessage = err.Error()
				now := time.Now().UTC()
				if job.CompletedAt == nil {
					job.CompletedAt = &now
				}
				fmt.Printf("[Worker] Broadcasting failed job status: %s\n", job.ID)
				hub.BroadcastJobStatusUpdate(job)
			} else {
				fmt.Printf("[Worker] Job ID %s completed successfully.\n", job.ID)
				qm.UpdateJobStatus(job.ID, api.StatusComplete, "")
				// Update the job object directly and broadcast
				job.Status = api.StatusComplete
				job.Progress = 100 // Ensure progress is 100%
				job.ErrorMessage = ""
				now := time.Now().UTC()
				if job.CompletedAt == nil {
					job.CompletedAt = &now
				}
				fmt.Printf("[Worker] Broadcasting completed job status: %s\n", job.ID)
				hub.BroadcastJobStatusUpdate(job)
			}

			// Optional short delay between processing jobs?
			// time.Sleep(1 * time.Second)
		}
	}() // Launch the goroutine
}
