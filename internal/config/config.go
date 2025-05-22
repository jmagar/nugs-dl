package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// AppConfig holds the application configuration.
// We will load this from config.json and potentially allow overrides via API later.
type AppConfig struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	Format          int    `json:"format"`      // 1: ALAC, 2: FLAC, 3: MQA, 4: 360RA/Best, 5: AAC
	VideoFormat     int    `json:"videoFormat"` // 1: 480p, 2: 720p, 3: 1080p, 4: 1440p, 5: 4K/Best
	OutPath         string `json:"outPath"`
	Token           string `json:"token"`           // Optional auth token
	UseFfmpegEnvVar bool   `json:"useFfmpegEnvVar"` // true: use ffmpeg from PATH, false: use from script dir

	// --- Fields derived or set internally, not directly from config.json ---
	// WantRes         string // Derived from VideoFormat (e.g., "1080") - Will handle later
	// FfmpegNameStr   string // Derived from UseFfmpegEnvVar - Will handle later
}

// Default configuration values
const (
	defaultOutPath = "Nugs downloads"
	configFileName = "config.json"
)

// getConfigPath returns the path to the config file, checking multiple locations
func getConfigPath() string {
	// Check if running in Docker (config directory exists)
	if _, err := os.Stat("config"); err == nil {
		dockerConfigPath := filepath.Join("config", configFileName)
		if _, err := os.Stat(dockerConfigPath); err == nil {
			return dockerConfigPath
		}
	}
	
	// Fallback to current directory
	return configFileName
}

// LoadConfig reads the configuration file (config.json) and returns the AppConfig struct.
// It applies default values for missing fields.
func LoadConfig() (*AppConfig, error) {
	cfg := &AppConfig{}

	configPath := getConfigPath()
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		// If the file doesn't exist, we might proceed with defaults or require it.
		// For now, let's return an error if it's not found, similar to original behavior.
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%s not found", configPath)
		}
		return nil, fmt.Errorf("error reading %s: %w", configPath, err)
	}

	err = json.Unmarshal(data, cfg)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling %s: %w", configPath, err)
	}

	// Apply defaults
	if cfg.OutPath == "" {
		cfg.OutPath = defaultOutPath
	}
	// Ensure OutPath is absolute
	cfg.OutPath, err = filepath.Abs(cfg.OutPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for outPath: %w", err)
	}

	// Validation (basic range checks like in parseCfg)
	if !(cfg.Format >= 1 && cfg.Format <= 5) {
		// Defaulting format if invalid or not set? Or return error?
		// Let's return error for now to match original strictness.
		return nil, fmt.Errorf("config error: track Format must be between 1 and 5")
	}
	if !(cfg.VideoFormat >= 1 && cfg.VideoFormat <= 5) {
		// Defaulting video format if invalid or not set?
		return nil, fmt.Errorf("config error: video Format must be between 1 and 5")
	}

	// Token processing (like TrimPrefix) can be done here or when used.
	// For now, just load it as is.

	// Deriving WantRes and FfmpegNameStr will be handled closer to where they are needed,
	// potentially by functions within this package or the downloader package.

	return cfg, nil
}

// TODO: Add a function SaveConfig(*AppConfig) error for the API to use later.

// SaveConfig saves the provided AppConfig struct to the configuration file (config.json).
func SaveConfig(cfg *AppConfig) error {
	// Basic validation before saving
	if !(cfg.Format >= 1 && cfg.Format <= 5) {
		return errors.New("invalid track Format (must be 1-5)")
	}
	if !(cfg.VideoFormat >= 1 && cfg.VideoFormat <= 5) {
		return errors.New("invalid video Format (must be 1-5)")
	}

	data, err := json.MarshalIndent(cfg, "", "  ") // Use MarshalIndent for readability
	if err != nil {
		return fmt.Errorf("error marshalling config to JSON: %w", err)
	}

	configPath := getConfigPath()
	err = ioutil.WriteFile(configPath, data, 0644) // Use standard file permissions
	if err != nil {
		return fmt.Errorf("error writing %s: %w", configPath, err)
	}

	return nil
}
