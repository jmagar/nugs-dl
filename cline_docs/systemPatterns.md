# System Patterns

## How the system is built

The system now comprises two main parts:

1.  **Core Downloader Logic (Go):** An internal Go package (`internal/downloader`) containing the refactored logic for authentication, nugs.net API interaction, file downloading, HLS processing, and FFmpeg interaction.
2.  **Interfaces:**
    *   **CLI:** The original command-line executable (`main.go`) which directly uses the core logic (or will, once fully refactored).
    *   **Web Server (Go/Gin):** A backend API server (`cmd/server/main.go`) built with the Gin framework. It exposes REST endpoints (`/api/config`, `/api/downloads`) and a Server-Sent Events (SSE) endpoint (`/api/status-stream`) for real-time updates.
    *   **Web UI (React/Vite):** A single-page application (`webui/`) built with React, TypeScript, Vite, Tailwind CSS, and shadcn/ui. It interacts with the Go backend API.

## Key technical decisions

-   **Go Language:** For backend performance, CLI tooling, and core download logic.
-   **React/Vite/TypeScript:** For a modern, type-safe frontend development experience.
-   **Tailwind CSS / shadcn/ui:** For efficient UI styling and component building.
-   **Gin Framework:** For the Go web API backend.
-   **Server-Sent Events (SSE):** For unidirectional real-time progress updates from backend to frontend.
-   **API Layer:** Decoupling the frontend from the core download logic.
-   **Modular Backend:** Refactoring core logic into internal packages (`internal/downloader`, `internal/queue`, `internal/config`, `pkg/api`).
-   **Configuration File (`config.json`):** Persisting user settings.
-   **FFmpeg Integration:** Leveraging external tool for media conversions.

## Architecture patterns

-   **CLI Application:** For the original interface.
-   **Client-Server Architecture:** For the Web UI (React frontend client, Go backend server).
-   **REST API:** For frontend-backend communication (config, adding downloads, getting bulk status).
-   **Server-Sent Events (SSE):** For pushing real-time status/progress updates.
-   **Background Worker (Goroutine):** For processing download jobs asynchronously from the queue.
-   **Modular Design:** Separating concerns into packages (API types, config, downloader, queue, worker, server command, web UI). 