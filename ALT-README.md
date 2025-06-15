# nugs-dl - purps edition

This project is proudly based on the outstanding work of [@Sorrow446/Nugs-Downloader](https://github.com/Sorrow446/Nugs-Downloader).  
Much appreciation to Sorrow446 and contributors for building the foundation and API understanding that makes this automation possible!

---

> #### Why This Exists  
> I've amassed a massive collection from nugs.net—over **22TB**—and these features were sorely needed to help automate keeping everything organized and up to date.

---

## Progress Tracker

Below are the key features and enhancements being implemented.  
**Check the boxes as features are developed and tested!**

- [ ] Batch & selective downloading (all artists or by name)
- [ ] Per-artist automatic output paths
- [ ] **YAML-based configuration** (switch from JSON)
- [ ] Dry run mode (simulate without downloading)
- [ ] Gap detection (collection audit: compare API and local)
- [ ] Intelligent queue management (only download missing)
- [ ] Latest vs. complete catalog downloads
- [ ] Pre-download disk space check (if file sizes available)
- [ ] Monitoring mode (poll for new shows/releases)
- [ ] Gotify notification support (global & per-artist)
- [ ] Robust error handling and summary reports
- [ ] Logging (configurable, with verbosity)
- [ ] Automatic directory creation
- [ ] Parallel downloads
- [ ] Retry policy
- [ ] Video/audio handling options (force video, skip video, skip chapters)
- [ ] Artist enable/disable
- [ ] Config validation (runs automatically at startup)
- [ ] Sample `config.yaml.example`
- [ ] Containerization (Dockerfile + docker-compose.yaml)
- [ ] Detailed README documentation and usage
- [ ] Contribution guidelines and MIT license

---

## What's Different: YAML Configuration

This project switches from the original **JSON** configuration style to **YAML** for all configuration.  
YAML is more flexible and easier to manage for complex options like:
- Per-artist settings
- Nested options (e.g., monitoring intervals, notifications)
- Multi-line and structured values

All user configuration is now handled in a single, commented YAML file (`config.yaml`).  
**See below for a full example and documentation.**

---

## Containerization

You can easily run the downloader using Docker.  
A `Dockerfile` and `docker-compose.yaml` are included for convenience.

### Usage

```bash
# 1. Copy and edit the example config to your own config
cp config.yaml.example config.yaml

# 2. Build and run with Docker Compose -- IMPLEMENT LAST -- ONLY WHEN EVERYTHING ELSE IS WORKING
docker-compose up --build
```

- The container will automatically mount your `config.yaml` and media directory (see comments in `docker-compose.yaml`).
- Logs and downloaded media will be accessible via your mapped host directories.

---

## Config Validation

**Config validation is always performed at startup.**  
If your `config.yaml` is missing required fields or contains invalid values, the app will exit with a clear error message describing what needs to be fixed.

---

## Sample Configuration

A sample configuration file is provided as `config.yaml.example`.  
**To get started:**

```bash
cp config.yaml.example config.yaml
# Edit config.yaml with your credentials and desired options before first run.
```

See the Configuration section below for detailed documentation of all fields.

---

## Features

- **Batch & Selective Downloading**  
  Download all configured artists, or just a specific one by name.
- **Per-Artist Output Paths (Automatic)**  
  Specify only the artist's name and ID—output paths are built automatically.
- **External YAML Configuration**  
  Credentials, download options, artist list, paths, and advanced settings in one easy-to-edit file.
- **Dry Run Mode**  
  Simulate downloads and see what would happen, without saving any files.
- **Gap Detection (Collection Audit)**  
  For each artist, the tool fetches a list of all available shows and videos from nugs.net, then compares with your local collection.  
  Any missing shows—available online but not present locally—are reported as "gaps," so you can keep your collection complete and up-to-date.
- **Intelligent Queue Management**  
  Before starting any download, the tool pre-fetches the list of all available shows/releases for the artist from the nugs.net API, checks your collection, and builds a queue containing only what you are missing.  
  This means downloads are only attempted for new or missing shows, saving time and bandwidth.
- **Latest vs. Complete Downloads**  
  - Use the `latest` subcommand to download only the most recent releases for an artist (from `/artist/<id>/latest`).
  - Use the `complete` subcommand (or default behavior) to download the full catalog for an artist (from `/artist/<id>`).
- **Pre-Download Disk Space Check**  
  Before beginning downloads, the tool can (optionally) check available disk space in the artist's output path.  
  If the nugs.net API provides file size metadata for each show/release, the tool will sum the sizes of the missing items and warn you if you may not have enough space.
- **Monitoring Mode (Automated Polling & Downloads)**  
  Enable `monitor: true` to have the downloader run continuously, polling nugs.net for each enabled artist every X hours (set globally or per artist), automatically downloading new content as soon as it's posted.
- **Gotify Notifications**  
  Supports push notifications via [Gotify](https://gotify.net/). Receive alerts about downloads, errors, and status updates directly on your devices.  
  Notifications can be enabled/disabled globally or overridden per artist.
- **Verbose, Colorful Console Output & Detailed Logging**  
  Logs are timestamped, stored in a configurable directory, and support multiple verbosity levels.
- **Automatic Directory Creation**  
  Output folders are created as needed.
- **Parallel Downloads**  
  Configure how many downloads run concurrently.
- **Retry Policy**  
  Automatically retry failed downloads.
- **Advanced Audio/Video Handling**  
  Support for forced video, skipping videos, and skipping video chapters.
- **Artist Enable/Disable**  
  Easily exclude artists temporarily without removing them from your config.
- **Robust Error Handling and Summary Reports**

---

## Gotify Notification Support

You can configure the downloader to send push notifications using [Gotify](https://gotify.net/).  
Notifications can be enabled or disabled globally or overridden per artist.  
Use notifications for events such as new downloads, errors, completion, or monitoring alerts.

**Example configuration:**

```yaml
notifications: true
gotifyUrl: "https://gotify.example.com"
gotifyToken: "YOUR_GOTIFY_TOKEN"

artists:
  - id: 1125
    name: "Billy Strings"
    enabled: true
    notifications: false     # Disable notifications for this artist only
  - id: 1205
    name: "Goose"
    enabled: true           # Uses global notifications setting (true)
```

> If `notifications` is not set for an artist, the global value is used.
> If `notifications` is `false` globally, no notifications are sent at all (even if set to `true` for an artist).

---

## Monitoring Mode (Automated Polling & Downloads)

Enable `monitor: true` in your config to run the downloader in monitoring mode.  
The tool will automatically poll nugs.net for each enabled artist every X hours (settable globally or per artist), checking for new releases and downloading them as soon as they're available.  
All notification settings apply when running in monitor mode.

Example:

```yaml
monitor: true
monitorIntervalHours: 6

artists:
  - id: 1125
    name: "Billy Strings"
    enabled: true
    monitorIntervalHours: 2    # Poll Billy Strings every 2 hours
  - id: 1205
    name: "Goose"
    enabled: true              # Uses global interval (6 hours)
```

**How it works:**
- The tool stays running and checks the `/artist/<id>/latest` endpoint for each artist on schedule.
- If new shows are found (not in your collection), they are downloaded automatically.
- All other downloader features (gap detection, logging, queue management) apply.

---

## How Queue Management & Gap Detection Work

Instead of blindly downloading every show again, the downloader:

1. **Fetches the complete list of available shows/releases for the selected artist from the nugs.net API**  
   (using either the `/artist/<id>/latest` or `/artist/<id>` endpoint depending on your command).
2. **Scans your local output directory for that artist**  
   to see which shows you already own.
3. **Builds a queue of only missing shows/releases**  
   so only those are downloaded.
4. **(Optional) Checks available disk space**  
   If the nugs.net API provides file size info for each release, the downloader will sum the total size of the shows/releases in your download queue and compare with available disk space in the target path.  
   - If the API does not provide this info, the downloader will warn you that a space check could not be performed.
   - If the API does provide file sizes, the downloader will warn you if space is insufficient before starting downloads.
5. **Downloads each missing show/release in the queue**  
   and stops when the queue is empty (or on error, if not in retry mode).

**Example Usage:**

Suppose you have the first 100 Billy Strings shows in `/mnt/user/data/media/music/Billy Strings`, and Billy has released 25 more.  
When you run:

```bash
./nugs latest "Billy Strings"
# Uses: https://play.nugs.net/#/artist/461/latest
```
or
```bash
./nugs complete "Billy Strings"
# Uses: https://play.nugs.net/#/artist/461
```
the downloader will:

- Fetch the list of all (125) available shows from nugs.net
- Scan your output directory and see that you only have 100
- Queue up the 25 missing shows
- Optionally check if you have enough disk space for those shows
- Download only the new/missing 25 shows

**This avoids re-downloading shows you already have, and saves time, bandwidth, and disk wear.**

> **Note:**  
> Whether file size metadata is available for each show/release depends on the nugs.net API.  
> If available, disk space checks are performed; if not, the downloader will warn you and proceed.

---

## Prerequisites

1. **Go Environment**  
   The downloader is written in Go. Build or install as per standard Go tools.
2. **`ffmpeg`**  
   Required for some downloads (e.g., video, HLS).  
   Make sure the path to ffmpeg is set correctly in your config, or available in your environment.
3. **nugs.net Account**  
   You must provide valid login credentials or token in your config.
4. **nugs Binary**  
   If your workflow uses the original `nugs`/`nugs-dl` binary, ensure its path is set in the config.

---

## Configuration

All options are set in a single YAML file (default: `config.yaml`).  
**Example:**

```yaml
email: "your@email.com"
password: "yourpassword"
format: 4
videoFormat: 5
outPath: "/mnt/user/data/media/music"  # Global base output directory for all artists
token: ""
useFfmpegEnvVar: false

dryRun: true                        # Enable dry run mode (simulate actions only)
logDir: "./logs"                    # Directory for log files
logLevel: "info"                    # Logging verbosity: debug, info, warn, error
nugsBinaryPath: "./nugs"            # Path to nugs executable
ffmpegPath: "/usr/bin/ffmpeg"       # Path to ffmpeg executable

forceVideo: false                   # Force video when it co-exists with audio
skipVideos: false                   # Skip videos in artist URLs
skipChapters: false                 # Skip chapters for videos

maxConcurrentDownloads: 2           # Number of concurrent downloads allowed
maxRetries: 3                       # How many times to retry a failed download
retryDelaySeconds: 10               # Seconds to wait between retries

monitor: true                       # Enable monitoring mode (auto-poll for new shows)
monitorIntervalHours: 6             # Global interval (in hours) for polling artists

notifications: true                 # Enable/disable Gotify notifications globally
gotifyUrl: "https://gotify.example.com"   # Gotify server URL
gotifyToken: "YOUR_GOTIFY_TOKEN"          # Gotify app token

artists:
  - id: 1125
    name: "Billy Strings"
    enabled: true
    monitorIntervalHours: 2         # Poll Billy Strings every 2 hours (overrides global)
    notifications: false            # Disable notifications for this artist (overrides global)
  - id: 1205
    name: "Goose"
    enabled: true                   # Uses global monitorIntervalHours (6 hours) and notifications (true)
  # ... add more artists as desired ...
```

- **Add/Remove Artists:**  
  List additional artists by ID and name.  
  Output paths are automatically set to `outPath/Artist Name`.
- **Temporarily Exclude an Artist:**  
  Set `enabled: false` for any artist you want to skip, without removing them from the config.
- **Set polling interval per artist:**  
  Set `monitorIntervalHours` for any artist to override the global interval.
- **Enable/disable notifications per artist:**  
  Set `notifications` for any artist to override the global setting.

---

- **Monitoring mode:**  
  If `monitor: true`, the downloader runs continuously, polling and downloading as configured.
- **Gap detection and queue management happen automatically**  
  The tool will only attempt to download new/missing shows.
- **Dry run (simulate, no files downloaded):**  
  Set `dryRun: true` in your config or use the `--dry-run` CLI flag, if supported.
- **Change config file location:**  
  Use the `--config /path/to/config.yaml` flag if the program supports it.

---

## Gap Detection & Disk Space Check (Technical Details)

- The tool compares the list of available shows/releases for each artist (from the nugs.net API) with the list of directories/files in your local output path.
- Only missing shows are queued for download.
- If file size metadata is available from the API, the tool will sum the sizes of the missing shows and compare with your available disk space.  
  If you do not have enough space, it will warn you before starting downloads.
- If file sizes are not available from the API, it will warn you that space checks could not be performed.

---

## Logging

- Log files are stored in the directory specified by `logDir`.
- Each run creates a new log file named with the date and time.
- Console output uses color for readability; log files are plain text.
- Logging verbosity is controlled by the `logLevel` setting.

---

## Advanced Options

- **Parallelism:**  
  `maxConcurrentDownloads` controls how many artists/releases can be downloaded at the same time.
- **Retry Policy:**  
  `maxRetries` and `retryDelaySeconds` control download retry behavior.
- **Video/Audio Handling:**  
  Control forced video, skipping videos, or video chapters with the corresponding flags.
- **ffmpeg Path:**  
  If not using the default, set it directly in the config.

---

## Troubleshooting

- **"nugs binary not found or not executable":**  
  Check `nugsBinaryPath` and permissions.
- **"ffmpeg not found":**  
  Check `ffmpegPath` and your environment.
- **"Invalid credentials":**  
  Double-check email/password/token in your config.
- **"Artist not downloaded":**  
  Ensure `enabled: true` and that the artist name/ID is correct.
- **Permission Issues:**  
  Ensure the output and log directories are writable.

---

## Contributions & License

Contributions are welcome!  
Please open an issue or pull request for bugs, improvements, or feature requests.

License: [MIT](LICENSE)