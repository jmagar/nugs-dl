package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	// Use alias for local config package to avoid potential future conflicts
	appConfig "main/internal/config"
)

// Global variable to hold the loaded configuration
// Using a mutex for safe concurrent access if handlers modify it
var (
	currentConfig *appConfig.AppConfig
	configMutex   sync.RWMutex
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

	router := gin.Default()

	// API group
	api := router.Group("/api")
	{
		api.GET("/config", getConfigHandler)
		api.POST("/config", updateConfigHandler)
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
	configMutex.RLock()         // Lock for reading
	defer configMutex.RUnlock() // Ensure unlock

	// Return a copy to avoid race conditions if the caller modifies the map/slice later
	// Although AppConfig is simple now, this is good practice.
	respConfig := *currentConfig

	c.JSON(http.StatusOK, respConfig)
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

	// Lock for writing
	configMutex.Lock()
	defer configMutex.Unlock()

	// Save the validated config to file
	if err := appConfig.SaveConfig(&updatedConfig); err != nil {
		fmt.Printf("Error saving config: %v\n", err) // Log error server-side
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save configuration"})
		return
	}

	// Update the in-memory config *after* successful save
	currentConfig = &updatedConfig

	fmt.Printf("Configuration updated and saved: %+v\n", currentConfig)

	// Return the newly saved configuration
	c.JSON(http.StatusOK, *currentConfig)
}
