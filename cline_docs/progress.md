# Progress

## What works

-   **CLI Downloader:** Core functionality for downloading various nugs.net content types via CLI.
-   **Web UI Backend:**
    -   Go/Gin API server (`cmd/server`).
    -   Configuration loading and API endpoints (`GET/POST /api/config`).
    -   Download job queueing (`POST /api/downloads`, creates separate job per URL).
    -   Background worker processing queued jobs.
    -   Job status API (`GET /api/downloads`, `GET /api/downloads/:jobId`).
    -   Real-time progress update broadcasting via Server-Sent Events (`GET /api/status-stream`).
    -   Artwork URL fetching and association with jobs.
    -   Job removal API (`DELETE /api/downloads/:jobId`).
-   **Core Downloader Logic (`internal/downloader`):**
    -   Refactored logic for authentication, metadata fetching, track processing, video processing (HLS), HLS audio processing, FFmpeg interaction (remux, chapters), and utilities.
    -   Reports progress updates via a channel.
-   **Web UI Frontend (`webui/`):**
    -   Built with React/Vite/TS/Tailwind/shadcn.
    -   Configuration form (in slide-out Sheet) that loads/saves settings via API.
    -   Download submission form.
    -   Queue display table showing job details (ID, URL, Status, Progress, Speed, Current File, Artwork, Error).
    -   Real-time updates via SSE connection.
    -   Ability to remove jobs from the queue (with confirmation).
    -   Ability to copy original URL for completed jobs.
    -   Basic light/dark theme toggle.

## What's left to build

-   **Core Downloader Logic:**
    -   Implement `CatalogPlaylistUrl` resolution (follow shortlink redirect).
    -   Implement `PurchasedUrl` parameter parsing and handling.
    -   Implement resumable downloads for direct video (optional).
    -   Refine error handling/collection.
    -   Refine FFmpeg path detection (`getFfmpegCmd` placeholder).
    -   Improve filename sanitization (`SanitizeFilename` TODO).
-   **Backend:**
    -   Implement API/Queue logic for reordering jobs (optional).
    -   More robust error handling and reporting.
-   **Frontend:**
    -   Implement UI for reordering jobs (optional).
    -   Improve error display (e.g., full error in tooltip was added, but verify/enhance).
    -   Improve responsiveness further (e.g., table column handling).
    -   Add explicit UI feedback for ongoing FFmpeg processes (remuxing, chapters).
    -   Consider adding ability to clear *all* completed/failed jobs.
    -   End-to-End Testing.

## Progress status

-   Go CLI downloader remains functional.
-   Web UI feature (Climb wEb1) is complete, providing core functionality for configuration, adding downloads, and viewing real-time queue status.
-   Core backend logic is largely implemented but needs final URL type handlers and potentially more robust error/progress details. 