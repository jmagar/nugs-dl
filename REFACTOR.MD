# Nugs-dl Backend Refactor Plan (YAML Config & Advanced Features)

## 1. Overview

This document outlines the plan to refactor the nugs-dl Go backend. The primary goals are:
- Migrate from the current simple JSON configuration (`config.json`) to a more comprehensive YAML-based configuration (`config.yaml`).
- Implement a range of advanced features as defined by the new configuration schema, including per-artist overrides, enhanced downloading, dry run mode, improved logging, Gotify notifications, and concurrency/retry mechanisms.

## 2. Core Configuration Changes (`internal/config/config.go`)

- **Objective**: Replace `config.json` with `config.yaml` and update the `AppConfig` struct and related logic.
- **Current State**: Simple `AppConfig` struct, JSON parsing from `config.json`.
- **New `AppConfig` Structure (YAML)**: The new structure is based on the actual `config.yaml` file, which uses a flat structure for global settings and a list for per-artist overrides.
  ```yaml
  # Illustrative structure based on actual config.yaml:
  email: "your_email@example.com"
  password: "your_nugs_password"
  format: 4 # Global default: 1:ALAC, 2:FLAC, 3:MQA, 4:360RA/Best, 5:AAC
  videoFormat: 5 # Global default: 1:480p, 2:720p, 3:1080p, 4:1440p, 5:4K/Best
  outPath: "/music" # Global base output directory
  token: "" # Optional, for session reuse
  useFfmpegEnvVar: false

  dryRun: false
  logDir: "/logs"
  logLevel: "info" # debug, info, warn, error
  nugsBinaryPath: "./nugs" # Or path to your nugs-cli binary
  ffmpegPath: "/usr/bin/ffmpeg"

  forceVideo: false
  skipVideos: false
  skipChapters: false

  maxConcurrentDownloads: 2
  maxRetries: 3
  retryDelaySeconds: 10

  monitor: true
  monitorIntervalHours: 6 # Global interval for polling artists

  notifications: true # Global Gotify notification toggle
  gotifyUrl: "http://gotify_server_url"
  gotifyToken: "YOUR_GOTIFY_APP_TOKEN"
  gotifyPriority: 5

  artists:
    - id: 1125
      name: "Billy Strings"
      enabled: true
      monitorIntervalHours: 2 # Override global monitor interval
      notifications: false    # Override global notification setting
      format: 2               # Override global format (e.g., to FLAC)
      videoFormat: 3          # Override global video format (e.g., to 1080p)
      outPath: "/music/Billy Strings" # Specific output path for this artist
    - id: 1205
      name: "Goose"
      enabled: true
      # ... other artists with their specific overrides or using global defaults
  ```
- **Tasks**:
  - Define new Go structs in `internal/config/config.go` to precisely match the `config.yaml` schema (e.g., `AppConfig`, `ServerConfig`, `DownloadConfig`, `ArtistConfig`, `GotifyConfig`, `LoggingConfig`).
  - Replace `json.Unmarshal` with `yaml.Unmarshal` (e.g., using `gopkg.in/yaml.v3`).
  - Update `LoadConfig()` to read from `config.yaml`.
  - Review `SaveConfig()`: determine if saving the full YAML config is needed or if it's primarily read-only. If parts are dynamic (e.g. token), that specific part might need saving.
  - Implement robust default value handling for all new configuration fields.
  - Add comprehensive validation for the new configuration structure upon loading.
- **Affected Files**:
  - `internal/config/config.go` (primary)
  - `cmd/server/main.go` (config loading and usage)
  - `main.go` (CLI, if its configuration handling is updated to use the new YAML)
  - All other modules and functions that currently consume `AppConfig`.

## 3. Feature Implementation Details

### 3.1. Per-Artist Configuration Overrides

- **Objective**: Allow specific download settings (format, path, enabled status, etc.) per artist, which override global defaults.
- **Current State**: No per-artist settings; configuration is global.
- **Tasks**:
  - Ensure `Artists []ArtistConfig` (with an `Enabled bool` field) is part of the main `AppConfig` struct.
  - Implement helper functions within the `internal/config` package. These functions will take an artist identifier (e.g., name or ID) and return an "effective" configuration for that artist by merging global defaults with any artist-specific overrides. Only process artists where `Enabled` is true.
  - Modify `internal/downloader/downloader.go` and related processing logic to retrieve and use this effective configuration for each download job.
- **Affected Files**:
  - `internal/config/config.go`
  - `internal/downloader/downloader.go`

### 3.2. Dry Run Mode

- **Objective**: Add a mode to simulate download operations without actually downloading files, for testing and verification.
- **Current State**: Not implemented.
- **Tasks**:
  - Add a `DryRun bool` flag to the `DownloadConfig` section of `AppConfig`.
  - In `internal/downloader/downloader.go` and its sub-functions, introduce conditional logic:
    - If `DryRun` is true, the downloader should perform metadata fetches, check for file existence, log intended actions (e.g., "DRY RUN: Would download [track_name] to [target_path] with format [format]"), but skip actual file download, conversion, and tagging.
- **Affected Files**:
  - `internal/config/config.go`
  - `internal/downloader/downloader.go`
  - `internal/downloader/processing.go`

### 3.3. Enhanced Logging and Console Output

- **Objective**: Implement a structured logging system with configurable levels, output to files, and potentially log rotation. Enhance console output for better readability.
- **Current State**: Logging is basic, primarily using `fmt.Printf` and `log.Printf`. Console output is not specifically enhanced.
- **Tasks**:
  - Define `LoggingConfig` within `AppConfig` (e.g., `level: "info"`, `logDir: "/logs"`, `maxSizeMB: 100`, `maxBackups: 3`, `maxAgeDays: 7`).
  - Select and integrate a structured logging library (e.g., `logrus`, `zap`, or the standard library's `slog` available from Go 1.21+).
  - Create a centralized logging utility or wrapper (e.g., in `internal/logger/logger.go`) to initialize and provide access to the logger instance.
  - Systematically replace existing `fmt.Printf` and `log.Printf` calls throughout the codebase with calls to the new structured logger.
  - Implement log file output and rotation if `logDir` is specified and the chosen library supports it or by adding a library for it.
  - Consider using a library or simple ANSI escape codes for more verbose and colorful console output, especially for progress indication and important messages, controlled by the logging level or a separate verbosity flag if needed.
- **Affected Files**:
  - `internal/config/config.go`
  - New `internal/logger/logger.go` (or similar)
  - Widespread changes across most `.go` files to update logging calls.
  - Console output points in `cmd/server/main.go`, `internal/worker/worker.go`, `internal/downloader/downloader.go`.

### 3.4. Notifications (Gotify)

- **Objective**: Send notifications via a Gotify server for significant download events.
- **Current State**: Not implemented.
- **Tasks**:
  - Define `GotifyConfig` within `AppConfig` (e.g., `url`, `token`, `priority`).
  - Implement a Gotify client in a new module (e.g., `internal/notifications/gotify.go`). This client will handle HTTP requests to the Gotify server.
  - Integrate calls to this Gotify client at key event points:
    - Download job completion (both success and failure) in `internal/worker/worker.go` or `internal/downloader/downloader.go`.
    - Potentially other events like queue processing start/stop, or critical errors.
- **Affected Files**:
  - `internal/config/config.go`
  - New `internal/notifications/gotify.go`
  - `internal/worker/worker.go` or `internal/downloader/downloader.go`

### 3.5. Concurrency Control and Retries

- **Objective**: Implement configurable limits for concurrent downloads and automatic retries for failed downloads.
- **Current State**: These features are not fully implemented or are missing.
- **Tasks**:
  - Add `MaxConcurrentDownloads int`, `MaxRetries int`, `RetryDelaySeconds int` to `DownloadConfig`.
  - **Concurrency Control**:
    - In `internal/worker/worker.go`, use a mechanism like a semaphore (e.g., `golang.org/x/sync/semaphore`) or a buffered channel to limit the number of concurrently running download goroutines to `MaxConcurrentDownloads`.
  - **Retry Mechanism**:
    - In `internal/downloader/downloader.go`, wrap the core download attempt logic (for individual files/segments) in a retry loop.
    - This loop should respect `MaxRetries` and wait for `RetryDelaySeconds` between attempts.
    - Log retry attempts and the reasons for failure.
- **Affected Files**:
  - `internal/config/config.go`
  - `internal/worker/worker.go` (for concurrency management)
  - `internal/downloader/downloader.go` (for retry logic)

### 3.6. Intelligent Queue Management & Collection Audit (Missing Show Detection)

- **Objective**: Before starting any download, pre-fetch the list of all available shows/releases for an artist from the nugs.net API, compare against the user's local collection, and build a download queue containing only missing items. This also serves as a collection audit.
- **Current State**: Not implemented.
- **Tasks**:
  - This is a core feature for efficient downloading.
  - Implement logic to:
    - Fetch a list of all available shows/media for an artist from nugs.net (potentially using existing or new nugs.net API interaction logic).
    - Scan the local `outPath` (and per-artist `outPath`s) for already downloaded content for that artist. This includes checking for existing files and potentially parsing metadata if available to confirm completeness or version.
    - Compare the two lists to identify shows/media present on nugs.net but missing or incomplete locally.
    - Automatically build the download queue (`internal/queue/manager.go`) with only these identified missing items.
    - Optionally, present a summary of missing items to the user (via logs or API) before queueing.
- **Affected Files**:
  - `internal/auditor/collection_audit.go` (new or expanded module for this logic).
  - `internal/nugs_client` (or similar for nugs.net API interaction).
  - `internal/queue/manager.go` (for populating the queue).
  - `internal/worker/worker.go` (to trigger this audit before processing artists).
  - `cmd/server/main.go` (if an API endpoint is added to trigger/display audit results or manage the queue manually).

### 3.7. Pre-Download Disk Space Check

- **Objective**: Optionally check available disk space in the artist's output path before initiating downloads and warn the user if space might be insufficient.
- **Current State**: Not implemented.
- **Tasks**:
  - Add a configuration option (e.g., `checkDiskSpace: true` in `AppConfig` or `DownloadConfig`).
  - If enabled, before processing the download queue for an artist (or globally):
    - Attempt to get total size of items in the queue (requires nugs.net API to provide file size metadata, or an estimation heuristic).
    - Check available disk space on the filesystem for the target `outPath`.
    - If estimated required space exceeds available space by a certain margin, log a warning and/or send a Gotify notification. Potentially halt downloads for that artist/batch if configured.
- **Affected Files**:
  - `internal/config/config.go`
  - `internal/worker/worker.go` (before starting downloads)
  - `internal/filesystem/utils.go` (new, for disk space checking utility)
  - `internal/nugs_client` (if relying on API for file sizes).

### 3.8. Monitoring Mode (Automated Polling & Downloads)

- **Objective**: Enable the downloader to run continuously, polling nugs.net for each enabled artist at configurable intervals, automatically downloading new content.
- **Current State**: `monitor: bool` and `monitorIntervalHours: int` fields exist in config, but polling logic is not implemented.
- **Tasks**:
  - If `monitor: true` in `AppConfig`:
    - The main application loop (likely in `cmd/server/main.go` or a dedicated monitoring service) should periodically trigger the download process for all `Enabled` artists.
    - The polling interval should be determined by `AppConfig.MonitorIntervalHours` globally, or `ArtistConfig.MonitorIntervalHours` if set for an artist.
    - This mode implies the application runs as a long-lived service.
    - Ensure robust error handling and logging during continuous operation.
- **Affected Files**:
  - `cmd/server/main.go` (or a new `internal/monitor/service.go`)
  - `internal/worker/worker.go` (to be callable by the monitor)
  - `internal/config/config.go`

## 4. Queue and Worker Updates (`internal/queue/manager.go`, `internal/worker/worker.go`)

- **Objective**: Adapt the existing queue and worker system to support the new features and configuration.
- **Current State**: Core queue (`Job` struct, manager) and worker logic exists.
- **Tasks**:
  - Update the `Job` struct in `internal/queue/manager.go` to carry any new necessary information, such as the effective artist-specific configuration.
  - Ensure `internal/worker/worker.go` correctly retrieves and passes these new parameters to the downloader functions.
  - Integrate concurrency control as described in section 3.5.
- **Affected Files**:
  - `internal/queue/manager.go`
  - `internal/worker/worker.go`

## 5. API Changes (`cmd/server/main.go`)

- **Objective**: Update existing API endpoints and add new ones to support frontend interactions with the new features.
- **Tasks**:
  - Modify existing configuration endpoints (`GET /api/config`, `POST /api/config`) to handle the new YAML structure. Consider security implications if sensitive parts of the config are exposed or modifiable.
  - Review and add new API endpoints as necessary to support new features (e.g., triggering a collection audit, if implemented via API).
  - Review all other API endpoints to ensure compatibility with the new configuration and any changed backend logic.
- **Affected Files**:
  - `cmd/server/main.go`

## 6. Dependencies

- **Required**:
  - `gopkg.in/yaml.v3` (or a similar library) for YAML parsing.
  - A structured logging library (e.g., `sirupsen/logrus`, `go.uber.org/zap`, or standard library `slog`).
- **Potentially**:
  - `golang.org/x/sync/semaphore` for concurrency control if not using other channel-based patterns.
  - Libraries for specific tasks if `ffmpeg` CLI interaction becomes too complex (though direct `exec` is often sufficient).

## 7. Testing Strategy

- **Unit Tests**:
  - For `internal/config`: new YAML parsing, validation, default value handling, and artist config override logic.
  - For new utility modules like `internal/logger` and `internal/notifications/gotify`.
  - For retry and concurrency logic in isolation if possible.
- **Integration Tests**:
  - Test the download pipeline end-to-end with various configurations: dry run, different artist overrides, retry scenarios.
  - Test API endpoints with mock data and validate responses and backend actions.
- **Manual Testing**:
  - Ensure automatic directory creation for output paths works as expected.

  - Thorough testing via the web UI once frontend components are adapted for new features.

## 8. Potential Challenges & Conflicts

- **Scope of Change**: Refactoring `AppConfig` and its usage is a pervasive change across the codebase. Careful planning and staged implementation are crucial.
- **CLI (`main.go`) Future**: A decision is needed on whether to update the CLI to use the new YAML config, simplify it, or deprecate it in favor of the web UI and API. This plan primarily focuses on the server backend.
- **Complexity of Gap Detection**: As noted, this is a significant undertaking and best treated as a stretch goal or future enhancement.
- **State Management**: The system is largely in-memory. While this refactor doesn't aim to change that, future needs for persistent download history or more complex job states might necessitate a database, which is out of the current scope.
- **Error Handling & Reporting**: With more complex logic (retries, overrides, monitoring), robust and clear error handling and reporting (via logs, API responses, and potentially Gotify notifications for critical errors) will be critical. Implement mechanisms for graceful shutdown and state saving if possible during monitoring mode. Consider generating summary reports after batch operations or monitoring cycles.

## 9. Quick Wins & Best Entrypoints

- **1. Foundational Config (`internal/config/config.go`)**:
  - Define all new Go structs for the YAML configuration.
  - Implement YAML loading and basic validation. This is the bedrock for all other changes.
- **2. Structured Logging (`internal/logger/logger.go`)**:
  - Integrate a logging library early. This will aid in debugging all subsequent feature implementations.
- **3. Dry Run Mode**:
  - Implementing this after basic config loading allows testing downloader modifications without actual network/file operations.
- **4. Gotify Notifications**:
  - Can be developed somewhat independently once the config and logging are in place.
- **Iterative Approach**: Tackle features one by one, starting with those that have fewer dependencies on others. For example, per-artist overrides can be built once the core config loading is done. Concurrency and retries can be layered onto the downloader once it's using the new config.

