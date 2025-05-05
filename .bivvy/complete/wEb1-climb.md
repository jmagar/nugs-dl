**STARTFILE wEb1-climb.md**
<Climb>
  <header>
    <id>wEb1</id>
    <type>feature</type>
    <description>A web-based UI for the nugs-dl Go application, replicating all CLI functionalities.</description>
  </header>
  <newDependencies>
    - Frontend: React 19, Vite, Tailwind CSS 4, shadcn/ui
    - Backend: A Go web framework (e.g., Gin, Echo) to create an API layer.
  </newDependencies>
  <prerequisitChanges>
    - The core Go application logic might need refactoring to be callable from web handlers/API endpoints, separating it from direct CLI argument parsing and execution.
    - An API layer needs to be designed and implemented in Go to connect the frontend and the core downloader logic.
    - Need to decide how configuration (credentials, paths, formats) is handled in a web context (vs. the current config.json/CLI args).
  </prerequisitChanges>
  <relevantFiles>
    - Existing Go source code (`main.go` and any other `.go` files containing downloader logic).
    - `README.md` (for understanding CLI options).
    - `config.json` (for current configuration approach).
  </relevantFiles>
  <everythingElse>
    ## Feature Overview
    - **Purpose:** Provide a user-friendly graphical interface for interacting with the nugs-dl downloader, eliminating the need for command-line usage.
    - **Problem Solved:** Makes the downloader accessible to users less comfortable with CLIs. Provides a visual way to manage downloads.
    - **Success Metrics:** Users can successfully configure, initiate, and monitor downloads via the web UI for all supported URL types and options.

    ## Requirements
    ### Functional Requirements
    1.  **Configuration:**
        -   Allow users to input and save nugs.net credentials (email/password or token). Consider security implications for web storage.
        -   Allow users to set/modify the default audio format (`format`).
        -   Allow users to set/modify the default video format (`videoFormat`).
        -   Allow users to set/modify the default output path (`outPath`).
        -   Allow users to configure FFmpeg usage (`useFfmpegEnvVar`).
        -   Configuration should likely persist between sessions.
    2.  **Download Initiation:**
        -   Provide an input area for users to paste one or more nugs.net URLs or paths to local text files containing URLs.
        -   Allow users to override default format settings for a specific download job.
        -   Provide options equivalent to CLI flags: `--force-video`, `--skip-videos`, `--skip-chapters`.
        -   Button to add the download request(s) to a queue.
    3.  **Download Feedback & Management:**
        -   Display the status of downloads (e.g., "queued", "downloading", "converting", "completed", "error") **in real-time**.
        -   Show **real-time** progress indicators (e.g., percentage complete, file size downloaded, download speed).
        -   Display logs or messages from the underlying Go downloader for each job.
        -   Display **album/video artwork** if available from nugs.net.
        -   List queued, active, completed, and failed downloads.
        -   Allow users to view the download queue.
        -   Allow users to potentially reorder or remove items from the queue before they start downloading.
        -   Allow users to clear completed/failed downloads from the list.
    4.  **Queue Management:**
        -   Implement a queue to handle multiple download requests sequentially or with controlled concurrency.
        -   Downloads should be processed based on their order in the queue.
    5.  **Sharing:**
        -   Provide a button/option for completed downloads to easily **copy the original nugs.net URL** to the clipboard (useful for sharing what was downloaded).
    ### Technical Requirements
    1.  **Frontend:** Build with React 19, Vite, Tailwind CSS 4, shadcn/ui.
    2.  **Backend:** Go-based API server using a web framework (e.g., Gin, Echo).
    3.  **API:** Define clear API endpoints for configuration, adding downloads to the queue, fetching **real-time** status/progress/logs (potentially via WebSockets or Server-Sent Events), managing the queue, and fetching artwork URLs.
    4.  **Concurrency & Queue:** The backend must manage a download queue and process jobs from it, potentially allowing a configurable number of concurrent downloads.
    ### User Requirements
    -   The interface should be intuitive and easy to navigate.
    -   Users should receive clear feedback on the status of their downloads.
    -   Configuration should be straightforward.
    ### Constraints
    -   Requires the Go backend and the underlying downloader logic to be running.
    -   Subject to the same technical constraints as the CLI (FFmpeg dependency, nugs.net authentication).

    ## Design and Implementation
    - **User Flow:**
        1. User opens web UI.
        2. User navigates to settings (if needed) to input credentials and configure defaults.
        3. User navigates to the main download page.
        4. User pastes URL(s) into the input field.
        5. User selects any non-default options (formats, flags).
        6. User clicks "Download".
        7. UI displays progress and status updates.
        8. User can view completed downloads.
    - **Architecture Overview:** A standard client-server architecture. React frontend communicates via HTTP requests to the Go backend API. The Go backend manages download jobs, potentially using goroutines, and interacts with the refactored downloader logic.
    - **API Specifications:** To be defined (e.g., `POST /api/downloads`, `GET /api/status`, `POST /api/config`).
    - **Data Models:** Need models for DownloadJob (URL, status, progress, options), Configuration (credentials, formats, path).

    ## Development Details
    - **Implementation Considerations:** How to manage the lifecycle of the Go backend process? How to handle long-running download processes initiated via API requests? **Efficient mechanism for streaming real-time progress** (WebSockets/SSE preferred over polling). Secure handling of credentials. Fetching artwork URLs. Managing the download queue state.
    - **Security Considerations:** Primarily around storing/handling nugs.net credentials entered via the web UI. Avoid storing plain text passwords.

    ## Testing Approach
    - **Test Cases:** Test each CLI option's equivalent in the UI. Test different URL types. Test concurrent downloads. Test error handling (invalid URLs, auth failures, download failures).
    - **Acceptance Criteria:** All CLI functionalities are usable via the UI. Downloads complete successfully. Status updates are accurate.

    ## Future Considerations
    - History of downloaded items (persistence beyond session).
    - More advanced configuration options.
    - Ability to pause/resume downloads (if feasible).
    - Ability to retry failed downloads.
  </everythingElse>
</Climb>
**ENDFILE** 