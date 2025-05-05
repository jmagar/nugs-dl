package main

import (
	// Needed for marshalling SSE data
	"encoding/json"
	"fmt"
	"io" // Needed for SSE io.EOF check
	"net/http"
	"net/http/cookiejar" // Import cookiejar
	"os"
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

	// Create the progress channel
	progressUpdates = make(chan api.ProgressUpdate, 100) // Keep using package name

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
			// Forward specific ProgressUpdate to the hub
			messageHub.BroadcastProgressUpdate(update) // Use specific method
		}
		fmt.Println("[Progress Consumer] Channel closed, exiting.")
	}()

	// Start the Background Worker
	worker.StartWorker(queueManager, downloaderService)

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
	// Create a channel for this specific client
	clientChan := make(chan []byte)

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
