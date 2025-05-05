# Tech Context

## Technologies Used

-   **Backend:**
    -   Language: Go (version specified in `go.mod`)
    -   Web Framework: Gin (`github.com/gin-gonic/gin`)
    -   Real-time Comm: Server-Sent Events (`github.com/gin-contrib/sse`)
    -   UUIDs: `github.com/google/uuid`
    -   External Dependencies: FFmpeg (required for video conversion and some audio tracks)
-   **Frontend:**
    -   Framework/Library: React 19 (via Vite)
    -   Language: TypeScript
    -   Build Tool: Vite
    -   Styling: Tailwind CSS v4 (`@tailwindcss/vite` plugin)
    -   UI Components: shadcn/ui
    -   Real-time Comm: `EventSource` API (built-in browser API for SSE)
    -   Package Manager: pnpm (used during setup)

## Development Setup

-   **Go Backend:**
    -   Go environment needs to be set up.
    -   Run the server: `go run ./cmd/server/main.go` (from project root).
-   **Frontend:**
    -   Node.js and pnpm (or npm/yarn) required.
    -   Install dependencies: `cd webui && pnpm install`.
    -   Run the dev server: `cd webui && pnpm dev`.
    -   Access UI at the localhost URL provided by Vite (e.g., `http://localhost:5173`).
-   **Shared:**
    -   FFmpeg needs to be installed and accessible (either in PATH or `./ffmpeg` based on `config.json`).
    -   A `config.json` file in the project root is used for credentials and default settings.

## Technical Constraints

-   Requires user credentials or token for nugs.net authentication.
-   Relies on FFmpeg installation for full functionality.
-   Requires Go backend server and Frontend dev server to be running concurrently.
-   Web UI relies on backend API and SSE stream availability.

-   Internet connection is required to access nugs.net. 