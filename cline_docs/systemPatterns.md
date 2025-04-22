# System Patterns

## How the system is built

Currently, the system is built as a standalone Command Line Interface (CLI) application written in Go. It takes user input via command-line arguments or a configuration file (`config.json`).

## Key technical decisions

-   **Go Language:** Chosen for its performance and suitability for CLI tools.
-   **CLI Interface:** Provides a simple way for users to interact with the downloader.
-   **Configuration File:** Allows users to store credentials and preferences persistently.
-   **FFmpeg Integration:** Leveraged for media format conversions where necessary.

## Architecture patterns

-   The primary pattern is a simple CLI application architecture.
-   It likely involves modules for argument parsing, configuration loading, API interaction with nugs.net, download management, and invoking external processes (FFmpeg). 