package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings" // Added for strings.ToLower
)

var appLogger *slog.Logger

// Init initializes the global application logger based on the provided log level and directory.
func Init(logLevelStr string, logDir string) error {
	var logLevel slog.Level
	logLevelStr = strings.ToLower(logLevelStr) // Ensure lowercase
	switch logLevelStr {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo // Default to info for unknown levels
	}

	var output io.Writer = os.Stdout

	if logDir != "" {
		// Ensure log directory exists
		// Ensure log directory exists
		if err := os.MkdirAll(logDir, 0755); err != nil {
			// If we can't even create the directory, log warning and fallback to stdout
			slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})).Warn("Could not create log directory, falling back to console-only logging", "directory", logDir, "error", err)
		} else {
			logFilePath := filepath.Join(logDir, "nugs-dl.log")
			logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				// If we can't open the file, log warning and fallback to stdout
				slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})).Warn("Could not open log file, falling back to console-only logging", "path", logFilePath, "error", err)
			} else {
				// If file logging is successful, write to both stdout and the file.
				output = io.MultiWriter(os.Stdout, logFile)
			}
		} 
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
		// AddSource: true, // Uncomment to include source file and line number
	}

	var handler slog.Handler
	// For now, let's use TextHandler for readability, JSONHandler is better for machine processing.
	handler = slog.NewTextHandler(output, opts)
	// handler = slog.NewJSONHandler(output, opts)

	appLogger = slog.New(handler)
	slog.SetDefault(appLogger) // Optionally set as default for global slog calls

	appLogger.Info("Logger initialized", "level", logLevel.String(), "logDir", logDir)
	return nil
}

// Get returns the global application logger.
// It's recommended to call Init first, but this provides a basic default logger if not.
func Get() *slog.Logger {
	if appLogger == nil {
		// Fallback to a default logger if Init hasn't been called
		// This shouldn't happen in normal operation if Init is called early in main()
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return appLogger
}

// Debug logs a message at DebugLevel.
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

// Info logs a message at InfoLevel.
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Warn logs a message at WarnLevel.
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

// Error logs a message at ErrorLevel.
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}
