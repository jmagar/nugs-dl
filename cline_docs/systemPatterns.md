# System Patterns

## How the system is built

The system now comprises two main parts:

1.  **Core Downloader Logic (Go):** An internal Go package (`internal/downloader`) containing the refactored logic for authentication, nugs.net API interaction, file downloading, HLS processing, and FFmpeg interaction.
2.  **Interfaces:**
    *   **CLI:** The original command-line executable (`main.go`) which directly uses the core logic.
    *   **Web Server (Go/Gin):** A backend API server (`cmd/server/main.go`) built with the Gin framework. It exposes REST endpoints (`/api/config`, `/api/downloads`, `/api/download/:jobId`, `/api/history`) and a Server-Sent Events (SSE) endpoint (`/api/status-stream`) for real-time updates.
    *   **Web UI (React/Vite):** A single-page application (`webui/`) built with React, TypeScript, Vite, Tailwind CSS, and shadcn/ui. It interacts with the Go backend API.
    *   **Container Deployment:** Docker-based deployment with multi-stage builds and production-ready configuration.

## Key technical decisions

-   **Go Language:** For backend performance, CLI tooling, and core download logic.
-   **React/Vite/TypeScript:** For a modern, type-safe frontend development experience.
-   **Tailwind CSS / shadcn/ui:** For efficient UI styling and component building.
-   **Gin Framework:** For the Go web API backend.
-   **Server-Sent Events (SSE):** For unidirectional real-time progress updates from backend to frontend.
-   **API Layer:** Decoupling the frontend from the core download logic.
-   **Modular Backend:** Refactoring core logic into internal packages (`internal/downloader`, `internal/queue`, `internal/config`, `internal/broadcast`, `internal/worker`, `pkg/api`).
-   **Configuration Management:** File-based configuration with Docker-friendly structure (`config/config.json`).
-   **FFmpeg Integration:** Leveraging external tool for media conversions.
-   **Zip Archive Delivery:** In-memory zip creation for completed download delivery.
-   **Docker Containerization:** Multi-stage builds for production deployment.

## Architecture patterns

-   **CLI Application:** For the original interface.
-   **Client-Server Architecture:** For the Web UI (React frontend client, Go backend server).
-   **REST API:** For frontend-backend communication (config, adding downloads, getting bulk status, downloading files).
-   **Server-Sent Events (SSE):** For pushing real-time status/progress updates with specific event types:
    *   `jobAdded` - New jobs added to queue
    *   `progressUpdate` - Download progress updates
    *   `jobStatusUpdate` - Job completion/failure status changes
-   **Background Worker (Goroutine):** For processing download jobs asynchronously from the queue.
-   **Broadcast Hub Pattern:** Central message broadcasting to multiple SSE clients.
-   **File Archive Pattern:** Zip archive creation for download delivery with album folder structure.
-   **Resilient File Handling:** Directory scanning for post-restart file availability.
-   **Modular Design:** Separating concerns into packages (API types, config, downloader, queue, worker, broadcast, server command, web UI).
-   **Container Orchestration:** Docker Compose for service management with health checks and volume persistence.

## Data Flow Patterns

1.  **Download Submission:** Frontend → REST API → Queue Manager → Background Worker
2.  **Progress Updates:** Downloader → Progress Channel → Hub → SSE → Frontend
3.  **Status Changes:** Worker → Queue Manager → Hub → SSE → Frontend  
4.  **File Delivery:** Frontend → REST API → File Scanner → Zip Creator → HTTP Response
5.  **Configuration:** Frontend ↔ REST API ↔ File System (`config/config.json`)

## Error Handling Patterns

-   **Graceful Degradation:** System continues operating when individual components fail
-   **Client Reconnection:** Automatic SSE reconnection on connection loss
-   **File System Resilience:** Download availability maintained across server restarts
-   **API Error Responses:** Structured error messages with appropriate HTTP status codes 