package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings" // Added for string validation

	"gopkg.in/yaml.v3"
	"nugs-dl/internal/logger" // Import the logger package
)

// ArtistConfig holds settings specific to an artist, overriding defaults.
type ArtistConfig struct {
	ID                   int    `yaml:"id"`
	Name                 string `yaml:"name"`
	Enabled              bool   `yaml:"enabled"`
	MonitorIntervalHours int    `yaml:"monitorIntervalHours,omitempty"`
	Notifications        *bool  `yaml:"notifications,omitempty"` // Pointer to distinguish between false and not set
	Format               int    `yaml:"format,omitempty"`        // Overrides global format
	VideoFormat          int    `yaml:"videoFormat,omitempty"`   // Overrides global videoFormat
	OutPath              string `yaml:"outPath,omitempty"`       // Specific output path
	// Add other overridable fields as needed
}

// AppConfig holds the entire application configuration, loaded from config.yaml.
type AppConfig struct {
	Email                  string `yaml:"email"`
	Password               string `yaml:"password"`
	Format                 int    `yaml:"format"`      // Global default: 1: ALAC, 2: FLAC, 3: MQA, 4: 360RA/Best, 5: AAC
	VideoFormat            int    `yaml:"videoFormat"` // Global default: 1: 480p, 2: 720p, 3: 1080p, 4: 1440p, 5: 4K/Best
	OutPath                string `yaml:"outPath"`
	LiveVideoPath          string `yaml:"liveVideoPath,omitempty"`
	Token                  string `yaml:"token,omitempty"`
	UseFfmpegEnvVar        bool   `yaml:"useFfmpegEnvVar"`

	DryRun                 bool   `yaml:"dryRun"`
	LogDir                 string `yaml:"logDir"`
	LogLevel               string `yaml:"logLevel"` // debug, info, warn, error
	NugsBinaryPath         string `yaml:"nugsBinaryPath,omitempty"`
	FfmpegPath             string `yaml:"ffmpegPath,omitempty"`

	ForceVideo             bool   `yaml:"forceVideo"`
	SkipVideos             bool   `yaml:"skipVideos"`
	SkipChapters           bool   `yaml:"skipChapters"`

	MaxConcurrentDownloads int    `yaml:"maxConcurrentDownloads"`
	MaxRetries             int    `yaml:"maxRetries"`
	RetryDelaySeconds      int    `yaml:"retryDelaySeconds"`

	Monitor                bool   `yaml:"monitor"`
	MonitorIntervalHours   int    `yaml:"monitorIntervalHours"` // Global interval

	Notifications          bool   `yaml:"notifications"` // Global notification toggle
	GotifyURL              string `yaml:"gotifyUrl,omitempty"`
	GotifyToken            string `yaml:"gotifyToken,omitempty"`
	GotifyPriority         int    `yaml:"gotifyPriority,omitempty"` // Added from previous GotifyConfig

	Artists []ArtistConfig `yaml:"artists,omitempty"`

	// Disk Space Check
	CheckDiskSpace         bool `yaml:"checkDiskSpace,omitempty"`
	DiskSpaceLowWarningGB  int  `yaml:"diskSpaceLowWarningGB,omitempty"`
	HaltOnLowSpace         bool `yaml:"haltOnLowSpace,omitempty"`

	// Server settings (if they are ever needed, currently not in user's YAML)
	// ServerHost string `yaml:"serverHost,omitempty"`
	// ServerPort int    `yaml:"serverPort,omitempty"`
}

// Default configuration values
const (
	defaultOutPath            = "Nugs downloads"
	defaultLogDir             = "logs"
	configFileName            = "config.yaml"
	defaultFormat             = 2 // FLAC
	defaultVideoFormat        = 3 // 1080p
	defaultMaxConcurrent      = 2
	defaultMaxRetries         = 3
	defaultRetryDelay         = 10
	defaultLogLevel           = "info"
	defaultMonitorInterval    = 6
	defaultGotifyPriority        = 5
	defaultCheckDiskSpace        = false
	defaultDiskSpaceLowWarningGB = 5
	defaultHaltOnLowSpace        = false
)

// getConfigPath returns the path to the config file (config.yaml).
// (Implementation remains the same)
func getConfigPath() string {
	if _, err := os.Stat("config"); err == nil {
		dockerConfigPath := filepath.Join("config", configFileName)
		if _, err := os.Stat(dockerConfigPath); err == nil {
			return dockerConfigPath
		}
	}
	return configFileName
}


// LoadConfig reads the configuration file (config.yaml) and returns the AppConfig struct.
// It applies default values for missing or invalid fields.
func LoadConfig() (*AppConfig, error) {
	cfg := &AppConfig{}

	configPath := getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Error("Configuration file not found", "path", configPath)
			return nil, fmt.Errorf("configuration file '%s' not found", configPath)
		}
		logger.Error("Error reading configuration file", "path", configPath, "error", err)
		return nil, fmt.Errorf("error reading configuration file '%s': %w", configPath, err)
	}

	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		logger.Error("Error unmarshalling YAML from configuration file", "path", configPath, "error", err)
		return nil, fmt.Errorf("error unmarshalling YAML from '%s': %w", configPath, err)
	}

	// Apply defaults and validate
	if cfg.OutPath == "" {
		cfg.OutPath = defaultOutPath
	}
	cfg.OutPath, err = filepath.Abs(cfg.OutPath)
	if err != nil {
		logger.Error("Failed to get absolute path for outPath", "path", cfg.OutPath, "error", err)
		return nil, fmt.Errorf("failed to get absolute path for outPath: %w", err)
	}

	if cfg.Format == 0 {
		logger.Info("Global track Format not set, defaulting", "defaultFormat", defaultFormat)
		cfg.Format = defaultFormat
	}
	if !(cfg.Format >= 1 && cfg.Format <= 5) {
		logger.Error("Invalid global track Format", "format", cfg.Format)
		return nil, fmt.Errorf("config error: global track Format must be between 1 and 5, got %d", cfg.Format)
	}

	if cfg.VideoFormat == 0 {
		logger.Info("Global video Format not set, defaulting", "defaultVideoFormat", defaultVideoFormat)
		cfg.VideoFormat = defaultVideoFormat
	}
	if !(cfg.VideoFormat >= 1 && cfg.VideoFormat <= 5) {
		logger.Error("Invalid global video Format", "videoFormat", cfg.VideoFormat)
		return nil, fmt.Errorf("config error: global video Format must be between 1 and 5, got %d", cfg.VideoFormat)
	}
	
	if cfg.MaxConcurrentDownloads <= 0 {
		cfg.MaxConcurrentDownloads = defaultMaxConcurrent
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = defaultMaxRetries
	}
	if cfg.RetryDelaySeconds <= 0 {
		cfg.RetryDelaySeconds = defaultRetryDelay
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = defaultLogLevel
	}
	validLogLevels := []string{"debug", "info", "warn", "error"}
	isValidLevel := false
	for _, level := range validLogLevels {
		if strings.ToLower(cfg.LogLevel) == level {
			isValidLevel = true
			cfg.LogLevel = strings.ToLower(cfg.LogLevel) // Ensure lowercase
			break
		}
	}
	if !isValidLevel {
		logger.Error("Invalid logLevel", "logLevel", cfg.LogLevel, "validLevels", validLogLevels)
		return nil, fmt.Errorf("invalid logLevel '%s', must be one of: %v", cfg.LogLevel, validLogLevels)
	}


	if cfg.LogDir == "" {
		cfg.LogDir = defaultLogDir
	}
	cfg.LogDir, err = filepath.Abs(cfg.LogDir)
	if err != nil {
		logger.Error("Failed to get absolute path for logDir", "path", cfg.LogDir, "error", err)
		return nil, fmt.Errorf("failed to get absolute path for logDir: %w", err)
	}
	
	if cfg.MonitorIntervalHours <= 0 && cfg.Monitor { // Only default if monitoring is enabled
		cfg.MonitorIntervalHours = defaultMonitorInterval
	}

	if cfg.GotifyURL != "" && cfg.GotifyToken != "" && cfg.GotifyPriority == 0 {
		cfg.GotifyPriority = defaultGotifyPriority
		logger.Info("Defaulting gotifyPriority to", "gotifyPriority", cfg.GotifyPriority)
	}

	// Apply defaults for Disk Space Check
	// Note: CheckDiskSpace defaults to false (zero value for bool) if not in YAML, which is defaultCheckDiskSpace.
	// HaltOnLowSpace also defaults to false (zero value for bool), which is defaultHaltOnLowSpace.
	if cfg.CheckDiskSpace { // Only apply other defaults if the feature is enabled
		if cfg.DiskSpaceLowWarningGB <= 0 {
			cfg.DiskSpaceLowWarningGB = defaultDiskSpaceLowWarningGB
			logger.Info("Defaulting diskSpaceLowWarningGB", "value", defaultDiskSpaceLowWarningGB)
		}
	} else {
		// If CheckDiskSpace is false (either explicitly or by omission in YAML),
		// ensure dependent values are at their "disabled" or non-interfering defaults.
	}
	
	// Validate artist-specific formats if provided
	for i, artist := range cfg.Artists {
		if artist.Format != 0 && !(artist.Format >= 1 && artist.Format <= 5) {
			logger.Error("Invalid track Format for artist", "artist", artist.Name, "format", artist.Format)
			return nil, errors.New("config error for artist")
		}
		if artist.VideoFormat != 0 && !(artist.VideoFormat >= 1 && artist.VideoFormat <= 5) {
			logger.Error("Invalid video Format for artist", "artist", artist.Name, "videoFormat", artist.VideoFormat)
			return nil, errors.New("config error for artist")
		}
		if artist.OutPath != "" {
			cfg.Artists[i].OutPath, err = filepath.Abs(artist.OutPath)
			if err != nil {
				logger.Error("Failed to get absolute path for artist outPath", "artist", artist.Name, "path", artist.OutPath, "error", err)
				return nil, errors.New("failed to get absolute path for artist outPath")
			}
		}
	}

	return cfg, nil
}

// SaveConfig saves the provided AppConfig struct to the configuration file (config.yaml).
func SaveConfig(cfg *AppConfig) error {
	if cfg.OutPath == "" {
		logger.Error("SaveConfig validation failed: download output path cannot be empty")
		return errors.New("download output path cannot be empty")
	}
	if !(cfg.Format >= 1 && cfg.Format <= 5) {
		logger.Error("SaveConfig validation failed: global track Format invalid", "format", cfg.Format)
		return fmt.Errorf("track Format must be between 1 and 5, got %d", cfg.Format)
	}
	if !(cfg.VideoFormat >= 1 && cfg.VideoFormat <= 5) {
		logger.Error("SaveConfig validation failed: global video Format invalid", "videoFormat", cfg.VideoFormat)
		return fmt.Errorf("video Format must be between 1 and 5, got %d", cfg.VideoFormat)
	}
	// Validate artist-specific formats
	for _, artist := range cfg.Artists {
		if artist.Format != 0 && !(artist.Format >= 1 && artist.Format <= 5) {
			logger.Error("SaveConfig validation failed: artist track Format invalid", "artist", artist.Name, "format", artist.Format)
			return fmt.Errorf("config error for artist '%s': track Format must be between 1 and 5, got %d", artist.Name, artist.Format)
		}
		if artist.VideoFormat != 0 && !(artist.VideoFormat >= 1 && artist.VideoFormat <= 5) {
			logger.Error("SaveConfig validation failed: artist video Format invalid", "artist", artist.Name, "videoFormat", artist.VideoFormat)
			return fmt.Errorf("config error for artist '%s': video Format must be between 1 and 5, got %d", artist.Name, artist.VideoFormat)
		}
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		logger.Error("Error marshalling YAML for SaveConfig", "error", err)
		return fmt.Errorf("error marshalling YAML: %w", err)
	}

	configPath := getConfigPath()
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		logger.Error("Error writing configuration file for SaveConfig", "path", configPath, "error", err)
		return fmt.Errorf("error writing configuration file '%s': %w", configPath, err)
	}
	logger.Info("Configuration successfully saved", "path", configPath)
	return nil
}

// GetEffectiveArtistConfig is a placeholder for logic to merge global and artist-specific settings.
// func (c *AppConfig) GetEffectiveArtistConfig(artistNameOrID string) (*ArtistConfig, error) {
// 	 // TODO: Implement logic to find artist by name/ID and merge with c.Download defaults.
// 	 return nil, errors.New("not implemented")
// }
