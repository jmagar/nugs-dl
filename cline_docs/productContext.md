# Product Context

## Why this project exists

This project, Nugs-Downloader, exists to allow users to download audio and video content from the nugs.net streaming service for offline access. It provides both a command-line interface and a modern web application for managing downloads.

## What problems it solves

It solves the problem of users wanting to keep local copies of purchased or streamed content from nugs.net, which might otherwise only be available for online streaming. The system addresses several key user needs:

-   **Offline Access:** Download content for listening/viewing without internet connection
-   **Quality Control:** Choose specific audio/video quality formats
-   **Batch Processing:** Queue multiple downloads and monitor progress
-   **Organization:** Automatic file organization with proper metadata
-   **Convenience:** User-friendly web interface with real-time progress tracking
-   **Portability:** Downloadable zip archives for easy file management and sharing

## How it should work

The application provides two primary interfaces:

1.  **Command Line Interface (CLI):** The original interface. Users provide credentials and URLs via config/arguments. The tool handles authentication, downloads, and FFmpeg integration.

2.  **Web User Interface (Web UI):** A modern web application providing a comprehensive graphical interface for the downloader. It allows users to:
    *   **Configure settings** (credentials, formats, paths) via a user-friendly UI panel
    *   **Submit download URLs** with real-time validation and feedback
    *   **View and manage a queue** of downloads with real-time status, track-based progress ("Track x of x"), download speed, and album artwork
    *   **Monitor real-time progress** with immediate status updates via Server-Sent Events
    *   **Download completed content** as organized zip archives with proper folder structure
    *   **View download history** with persistent records of completed downloads
    *   **Remove jobs** from the queue with confirmation dialogs
    *   **Access original URLs** for completed items
    *   **Receive notifications** for job status changes and completion

3.  **Production Deployment:** The system supports containerized deployment with:
    *   **Docker containerization** with multi-stage builds for efficient production deployment
    *   **Volume persistence** for downloads and configuration across container restarts
    *   **Health monitoring** with built-in health checks
    *   **Scalable architecture** ready for production workloads

The web interface communicates with a Go backend API server which manages the download queue, orchestrates the download process using the refactored core downloader logic, and provides real-time updates. The system ensures downloads remain available even after server restarts and provides high-performance zip archive delivery for completed content. 