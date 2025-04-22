# Product Context

## Why this project exists

This project, Nugs-Downloader, exists to allow users to download audio and video content from the nugs.net streaming service for offline access.

## What problems it solves

It solves the problem of users wanting to keep local copies of purchased or streamed content from nugs.net, which might otherwise only be available for online streaming. It supports various audio and video formats and different types of nugs.net content URLs (albums, artists, playlists, livestreams, videos).

## How it should work

The application currently works as a command-line interface (CLI) tool written in Go. Users provide their nugs.net credentials via a configuration file or command-line arguments. They then provide URLs to the nugs.net content they wish to download. The tool handles authentication, fetches the media streams in the desired format/quality, and saves the files to a specified output directory. It utilizes FFmpeg for certain video conversions and HLS-only tracks. 