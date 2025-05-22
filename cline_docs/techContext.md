# Tech Context

## Technologies Used

-   **Backend:**
    -   Language: Go (version specified in `go.mod`)
    -   Web Framework: Gin (`github.com/gin-gonic/gin`)
    -   Real-time Comm: Server-Sent Events (`github.com/gin-contrib/sse`)
    -   UUIDs: `github.com/google/uuid`
    -   Archive Creation: Standard library `archive/zip`
    -   External Dependencies: FFmpeg (required for video conversion and some audio tracks)
-   **Frontend:**
    -   Framework/Library: React 19 (via Vite)
    -   Language: TypeScript
    -   Build Tool: Vite
    -   Styling: Tailwind CSS v4 (`@tailwindcss/vite` plugin)
    -   UI Components: shadcn/ui
    -   Real-time Comm: `EventSource` API (built-in browser API for SSE)
    -   Package Manager: pnpm (used during setup)
-   **Containerization:**
    -   Runtime: Docker with multi-stage builds
    -   Orchestration: Docker Compose
    -   Base Images: Node.js (build), Alpine Linux (runtime)
    -   Process Management: Docker health checks

## Development Setup

-   **Go Backend:**
    -   Go environment needs to be set up.
    -   Configuration: Create `config/config.json` in project root.
    -   Run the server: `go run ./cmd/server/main.go` (from project root).
-   **Frontend:**
    -   Node.js and pnpm (or npm/yarn) required.
    -   Install dependencies: `cd webui && pnpm install`.
    -   Build for production: `cd webui && pnpm build`.
    -   Run the dev server: `cd webui && pnpm dev`.
    -   Access UI at the localhost URL provided by Vite (e.g., `http://localhost:5173`).
-   **Docker Deployment:**
    -   Create directories: `mkdir -p downloads config`
    -   Create config file: `config/config.json` with credentials and settings.
    -   Build and run: `docker compose up -d --build`
    -   Access UI at `http://localhost:8080`
-   **Shared:**
    -   FFmpeg needs to be installed and accessible (either in PATH or `./ffmpeg` based on `config.json`).
    -   Configuration file supports both `config/config.json` (preferred) and legacy `config.json` locations.

## Technical Constraints

-   Requires user credentials or token for nugs.net authentication.
-   Relies on FFmpeg installation for full functionality (included in Docker containers).
-   For development: Requires Go backend server and Frontend dev server to be running concurrently.
-   For production: Single Docker container serves both backend API and frontend assets.
-   Web UI relies on backend API and SSE stream availability.
-   Large file downloads require sufficient disk space and memory for zip creation.
-   Internet connection is required to access nugs.net.

## Production Considerations

-   **Performance:** Zip creation happens in memory, requiring adequate RAM for large downloads (600MB+ tested).
-   **Storage:** Downloads persist in volume-mounted directories for container restarts.
-   **Security:** Container runs as non-root user, config mounted read-only in production.
-   **Monitoring:** Health checks available on `/ping` endpoint.
-   **Scaling:** Single-instance application, scaling requires external load balancer and shared storage. 