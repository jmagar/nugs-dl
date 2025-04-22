# Tech Context

## Technologies Used

-   **Primary Language:** Go
-   **External Dependencies:** FFmpeg (required for video conversion and some audio tracks)

## Development Setup

-   Go environment needs to be set up.
-   FFmpeg needs to be installed and accessible either in the system's PATH or in the application's directory (configurable via `useFfmpegEnvVar` in the config file).
-   A `config.json` file is used for credentials (email/password or token) and default settings (format, videoFormat, outPath, useFfmpegEnvVar).

## Technical Constraints

-   Requires user credentials or a token for authentication with nugs.net.
-   Relies on the availability and correct installation of FFmpeg for full functionality.
-   Internet connection is required to access nugs.net. 