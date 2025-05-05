# Product Context

## Why this project exists

This project, Nugs-Downloader, exists to allow users to download audio and video content from the nugs.net streaming service for offline access.

## What problems it solves

It solves the problem of users wanting to keep local copies of purchased or streamed content from nugs.net, which might otherwise only be available for online streaming. It supports various audio and video formats and different types of nugs.net content URLs (albums, artists, playlists, livestreams, videos).

## How it should work

The application provides two primary interfaces:

1.  **Command Line Interface (CLI):** The original interface. Users provide credentials and URLs via config/arguments. The tool handles authentication, downloads, and FFmpeg integration.
2.  **Web User Interface (Web UI):** A modern web application providing a graphical interface for the downloader. It allows users to:
    *   Configure settings (credentials, formats, paths) via a UI panel.
    *   Submit download URLs.
    *   View a queue of downloads with real-time status, progress, speed, and artwork.
    *   Receive notifications for job status changes.
    *   Remove jobs from the queue.
    *   Copy the original nugs.net URL for completed items.
    It communicates with a Go backend API server which manages the queue and orchestrates the download process using the refactored core downloader logic. 