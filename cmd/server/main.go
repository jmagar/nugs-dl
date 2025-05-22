package main

import (
	// Needed for marshalling SSE data

	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io" // Needed for SSE io.EOF check
	"net/http"
	"net/http/cookiejar" // Import cookiejar
	"os"                 // For file path operations
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-contrib/sse" // Import SSE
	"github.com/gin-gonic/gin"

	// Use standard import paths now with new module name
	appConfig "nugs-dl/internal/config"
	// Import queue and api packages
	"nugs-dl/internal/broadcast"
	"nugs-dl/internal/downloader"
	"nugs-dl/internal/queue"
	"nugs-dl/internal/worker"

	// Remove alias, use full path
	"nugs-dl/pkg/api"
)

// Global variable to hold the loaded configuration
// Using a mutex for safe concurrent access if handlers modify it
var (
	currentConfig     *appConfig.AppConfig
	configMutex       sync.RWMutex
	queueManager      *queue.QueueManager     // Global queue manager instance
	downloaderService *downloader.Downloader  // Global downloader instance
	sharedHttpClient  *http.Client            // Global HTTP client
	progressUpdates   chan api.ProgressUpdate // Keep using package name
	messageHub        *broadcast.Hub          // Global broadcaster hub instance
)

func main() {
	// Load initial configuration
	var loadErr error
	currentConfig, loadErr = appConfig.LoadConfig()
	if loadErr != nil {
		fmt.Printf("Error loading config.json: %v\n", loadErr)
		// Decide if server should run with default/empty config or exit
		// For now, let's exit if config is required and fails to load
		os.Exit(1)
	}

	fmt.Printf("Configuration loaded: %+v\n", currentConfig)

	// Initialize shared HTTP client with cookie jar
	jar, _ := cookiejar.New(nil)
	sharedHttpClient = &http.Client{Jar: jar}
	fmt.Println("HTTP Client initialized.")

	// Create the progress channel with larger buffer for high-frequency downloads
	progressUpdates = make(chan api.ProgressUpdate, 1000) // Keep using package name

	// Initialize the Queue Manager FIRST
	queueManager = queue.NewQueueManager()
	fmt.Println("Queue Manager initialized.")

	// Initialize the Downloader Service, passing channel AND queue manager
	downloaderService = downloader.NewDownloader(currentConfig, sharedHttpClient, progressUpdates, queueManager)
	fmt.Println("Downloader Service initialized.")

	// Initialize and run the Broadcaster Hub
	messageHub = broadcast.NewHub()
	go messageHub.Run()

	// Start the Progress Consumer goroutine to forward updates to the Hub
	go func() {
		fmt.Println("[Progress Consumer] Started...")
		for update := range progressUpdates {
			fmt.Printf("[Progress Consumer] Received update: Job %s, Percentage: %.1f%%, Speed: %d B/s\n", 
				update.JobID, update.Percentage, update.SpeedBPS)
			
			// Update the job progress in the queue manager
			queueManager.UpdateJobProgress(update.JobID, update.Percentage, update.SpeedBPS, update.CurrentFile, update.CurrentTrack, update.TotalTracks)
			
			// Forward specific ProgressUpdate to the hub for SSE broadcasting
			messageHub.BroadcastProgressUpdate(update) // Use specific method
		}
		fmt.Println("[Progress Consumer] Channel closed, exiting.")
	}()

	// Start the Background Worker
	worker.StartWorker(queueManager, downloaderService, messageHub)

	router := gin.Default()

	// API group
	apiGroup := router.Group("/api") // Renamed variable for clarity
	{
		apiGroup.GET("/config", getConfigHandler)
		apiGroup.POST("/config", updateConfigHandler)
		// Download queue endpoints
		apiGroup.POST("/downloads", addDownloadHandler)
		apiGroup.GET("/downloads", getDownloadsHandler)             // New endpoint for list
		apiGroup.GET("/downloads/:jobId", getDownloadJobHandler)    // New endpoint for specific job
		apiGroup.DELETE("/downloads/:jobId", removeDownloadHandler) // Added DELETE route
		apiGroup.GET("/status-stream", sseStatusHandler)            // SSE endpoint
		// History endpoint
		apiGroup.GET("/history", getHistoryHandler)                 // New endpoint for completed downloads
		// File download endpoint
		apiGroup.GET("/download/:jobId", downloadFileHandler)       // New endpoint to download completed files
	}

	// Simple health check endpoint
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Run the server
	port := "8080" // Consider making the port configurable later
	fmt.Println("Starting server on http://localhost:" + port)
	err := router.Run(":" + port)
	if err != nil {
		panic("Failed to start server: " + err.Error())
	}
}

// getConfigHandler handles GET /api/config requests
func getConfigHandler(c *gin.Context) {
	fmt.Println("[getConfigHandler] Received request") // Log entry
	// Reload config from file on each request to ensure freshness
	cfg, err := appConfig.LoadConfig()
	if err != nil {
		fmt.Printf("[getConfigHandler] Error reloading config: %v\n", err)
		// Determine appropriate error response - e.g., internal server error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load server configuration"})
		return
	}
	fmt.Printf("[getConfigHandler] Successfully loaded config: %+v\n", cfg) // Log loaded config

	// No need for mutex here as we are loading fresh from file
	// configMutex.RLock()
	// defer configMutex.RUnlock()
	// if currentConfig == nil { // No longer using global currentConfig directly here
	// 	fmt.Println("Error: currentConfig is nil in getConfigHandler")
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Server configuration is not loaded"})
	// 	return
	// }

	// Return the freshly loaded config
	fmt.Println("[getConfigHandler] Sending response") // Log before sending
	c.JSON(http.StatusOK, cfg)
}

// updateConfigHandler handles POST /api/config requests
func updateConfigHandler(c *gin.Context) {
	var updatedConfig appConfig.AppConfig

	// Bind the incoming JSON to the AppConfig struct
	if err := c.ShouldBindJSON(&updatedConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// --- Basic Validation (similar to SaveConfig, could be refactored) ---
	if !(updatedConfig.Format >= 1 && updatedConfig.Format <= 5) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track Format (must be 1-5)"})
		return
	}
	if !(updatedConfig.VideoFormat >= 1 && updatedConfig.VideoFormat <= 5) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video Format (must be 1-5)"})
		return
	}
	// Add more validation as needed (e.g., for OutPath)
	//----------------------------------------------------------------------

	// We still need the mutex here to prevent concurrent *writes* to the file
	// and to update the global state if other parts of the app need it immediately.
	configMutex.Lock()
	defer configMutex.Unlock()

	// Save the validated config to file
	if err := appConfig.SaveConfig(&updatedConfig); err != nil {
		fmt.Printf("Error saving config: %v\n", err) // Log error server-side
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save configuration"})
		return
	}

	// Update the in-memory config *after* successful save
	// Other parts of the application might rely on this in-memory version
	currentConfig = &updatedConfig

	fmt.Printf("Configuration updated and saved: %+v\n", currentConfig)

	// Return the newly saved configuration (which should match what GET returns now)
	c.JSON(http.StatusOK, *currentConfig) // Keep returning current state here
}

// addDownloadHandler handles POST /api/downloads requests
func addDownloadHandler(c *gin.Context) {
	var req api.AddDownloadRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	var results []api.AddDownloadResponseItem
	var addedJobs []*api.DownloadJob // Collect successfully added jobs
	for _, url := range req.Urls {
		job, err := queueManager.AddJob(url, req.Options)
		if err != nil {
			fmt.Printf("Error adding job for URL %s: %v\n", url, err)
			results = append(results, api.AddDownloadResponseItem{
				Url:   url,
				Error: fmt.Sprintf("Failed to add job to queue: %v", err),
			})
		} else {
			results = append(results, api.AddDownloadResponseItem{
				Url:   url,
				JobID: job.ID,
			})
			addedJobs = append(addedJobs, job) // Collect successful job
		}
	}

	// Broadcast newly added jobs via SSE AFTER adding them all
	for _, job := range addedJobs {
		messageHub.BroadcastJobAdded(job) // Use specific method
	}

	c.JSON(http.StatusAccepted, results)
}

// getDownloadsHandler handles GET /api/downloads requests (list all jobs)
func getDownloadsHandler(c *gin.Context) {
	// Retrieve all jobs from the manager (returns a safe copy)
	jobs := queueManager.GetAllJobs()

	// Return the list (might be empty)
	c.JSON(http.StatusOK, jobs)
}

// getDownloadJobHandler handles GET /api/downloads/:jobId requests (specific job)
func getDownloadJobHandler(c *gin.Context) {
	jobID := c.Param("jobId") // Get job ID from URL path parameter

	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	job, found := queueManager.GetJob(jobID)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Job with ID %s not found", jobID)})
		return
	}

	// Return the found job details
	c.JSON(http.StatusOK, job)
}

// sseStatusHandler handles Server-Sent Event connections for real-time updates.
func sseStatusHandler(c *gin.Context) {
	// Create a channel for this specific client with large buffer for progress updates
	clientChan := make(chan []byte, 500)

	// Register the client with the hub
	messageHub.RegisterClient(clientChan)

	// Ensure client is unregistered when the connection closes
	defer func() {
		messageHub.UnregisterClient(clientChan)
	}()

	// Set headers for SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins for simplicity (adjust in production)

	// Stream updates to the client
	c.Stream(func(w io.Writer) bool {
		select {
		case msgBytes, ok := <-clientChan:
			if !ok {
				return false // Stop streaming
			}
			// Determine event type by attempting to unmarshal into SSEEvent
			var sseEvent api.SSEEvent
			if json.Unmarshal(msgBytes, &sseEvent) == nil {
				// Successfully unmarshalled, use the Type field
				fmt.Printf("[SSE Handler] Sending event: %s with data length: %d\n", sseEvent.Type, len(msgBytes))
				// Still send the original msgBytes containing the wrapped data
				sse.Encode(w, sse.Event{
					Event: string(sseEvent.Type), // Use event type from struct
					Data:  string(msgBytes),
				})
			} else {
				fmt.Println("[SSE Handler] Failed to unmarshal SSEEvent from message bytes")
				// Optionally send a generic event if unmarshal fails?
				sse.Encode(w, sse.Event{Event: "message", Data: string(msgBytes)})
			}
			return true // Continue streaming
		case <-c.Request.Context().Done():
			// Client disconnected
			fmt.Println("[SSE Handler] Client disconnected.")
			return false // Stop streaming
		}
	})
}

// removeDownloadHandler handles DELETE /api/downloads/:jobId requests
func removeDownloadHandler(c *gin.Context) {
	jobID := c.Param("jobId")

	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	success := queueManager.RemoveJob(jobID)

	if !success {
		// Attempt to distinguish not found vs. wrong state? QueueManager logs it.
		// For the API, maybe just return 404 if not removed for any reason other than server error.
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Job %s not found or could not be removed (may be processing)", jobID)})
		return
	}

	// Return 204 No Content on successful removal
	c.Status(http.StatusNoContent)
}

// getHistoryHandler handles GET /api/history requests (list completed downloads)
func getHistoryHandler(c *gin.Context) {
	// Retrieve completed jobs from the manager
	completedJobs := queueManager.GetCompletedJobs()

	// Return the list (might be empty)
	c.JSON(http.StatusOK, completedJobs)
}

// downloadFileHandler handles GET /api/download/:jobId requests (download completed files)
func downloadFileHandler(c *gin.Context) {
	jobID := c.Param("jobId")

	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	configMutex.RLock()
	downloadPath := currentConfig.OutPath
	configMutex.RUnlock()

	var albumPath string
	var sanitizedTitle string

	// Try to get the job details from queue first
	job, found := queueManager.GetJob(jobID)
	if found && job.Status == api.StatusComplete {
		// Use job title if available
		sanitizedTitle = sanitizeForFilename(job.Title)
		albumPath = filepath.Join(downloadPath, sanitizedTitle)
	} else {
		// Job not in queue (server restart), scan downloads directory for available albums
		availableAlbums, err := findAvailableAlbums(downloadPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error scanning downloads directory: %v", err)})
			return
		}

		if len(availableAlbums) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "No completed downloads found"})
			return
		}

		// For now, use the first available album (in production, you might want to store job ID -> folder mapping)
		// Or better yet, use a different endpoint that lists available downloads
		sanitizedTitle = availableAlbums[0]
		albumPath = filepath.Join(downloadPath, sanitizedTitle)
		
		// Log for debugging
		fmt.Printf("[Download Handler] Job %s not in queue, using first available album: %s\n", jobID, sanitizedTitle)
	}

	// Check if the album directory exists
	if _, err := os.Stat(albumPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Downloaded files not found",
			"detail": fmt.Sprintf("Expected folder: %s", albumPath),
		})
		return
	}

	// Find all audio files in the directory
	audioFiles, err := findAudioFiles(albumPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error scanning download directory: %v", err)})
		return
	}

	if len(audioFiles) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No audio files found in download directory"})
		return
	}

	// Create a zip file in memory
	zipBuffer, err := createZipArchive(audioFiles, albumPath, sanitizedTitle)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error creating zip archive: %v", err)})
		return
	}

	// Set headers for file download
	filename := fmt.Sprintf("%s.zip", sanitizedTitle)
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Length", fmt.Sprintf("%d", len(zipBuffer)))

	// Serve the zip file
	c.Data(http.StatusOK, "application/zip", zipBuffer)
}

// sanitizeForFilename removes or replaces characters that are invalid in filenames
func sanitizeForFilename(filename string) string {
	// Replace invalid characters with safe alternatives
	invalid := []string{"<", ">", ":", "\"", "|", "?", "*", "/", "\\"}
	result := filename
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	// Trim spaces and dots from the end
	result = strings.TrimRight(result, " .")
	return result
}

// findAvailableAlbums scans the downloads directory and returns album folder names
func findAvailableAlbums(downloadsDir string) ([]string, error) {
	var albums []string
	
	entries, err := os.ReadDir(downloadsDir)
	if err != nil {
		return nil, err
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if this directory contains audio files
			albumPath := filepath.Join(downloadsDir, entry.Name())
			audioFiles, err := findAudioFiles(albumPath)
			if err == nil && len(audioFiles) > 0 {
				albums = append(albums, entry.Name())
			}
		}
	}
	
	return albums, nil
}

// findAudioFiles recursively finds all audio files in a directory
func findAudioFiles(dir string) ([]string, error) {
	var audioFiles []string
	audioExtensions := []string{".flac", ".mp3", ".aac", ".m4a", ".wav", ".alac"}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			for _, validExt := range audioExtensions {
				if ext == validExt {
					audioFiles = append(audioFiles, path)
					break
				}
			}
		}
		return nil
	})

	return audioFiles, err
}

// createZipArchive creates a zip archive containing all the specified files
func createZipArchive(files []string, basePath, albumName string) ([]byte, error) {
	// Create a buffer to write our archive to
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	for _, file := range files {
		// Get relative path from the base album directory
		relPath, err := filepath.Rel(basePath, file)
		if err != nil {
			relPath = filepath.Base(file) // Fallback to just the filename
		}

		// Create a zip file entry with album name as root folder
		zipPath := filepath.Join(albumName, relPath)
		zipFile, err := zipWriter.Create(zipPath)
		if err != nil {
			zipWriter.Close()
			return nil, fmt.Errorf("failed to create zip entry for %s: %v", relPath, err)
		}

		// Open the source file
		sourceFile, err := os.Open(file)
		if err != nil {
			zipWriter.Close()
			return nil, fmt.Errorf("failed to open source file %s: %v", file, err)
		}

		// Copy file contents to zip
		_, err = io.Copy(zipFile, sourceFile)
		sourceFile.Close()
		if err != nil {
			zipWriter.Close()
			return nil, fmt.Errorf("failed to copy file %s to zip: %v", file, err)
		}
	}

	// Close the zip writer to finalize the archive
	err := zipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %v", err)
	}

	return buf.Bytes(), nil
}
