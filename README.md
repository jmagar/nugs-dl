# Nugs-Downloader Web Application

A modern web application for downloading music and videos from nugs.net with real-time progress tracking and queue management.

> **Built on the foundation of [Sorrow446/Nugs-Downloader](https://github.com/Sorrow446/Nugs-Downloader)**  
> This web application extends the excellent CLI downloader created by [@Sorrow446](https://github.com/Sorrow446) with a modern React-based web interface, real-time progress tracking, and Docker deployment capabilities. All core download functionality is based on their original Go implementation.

![Web UI Dashboard](assets/Screenshot%202025-05-22%20182133.png)
![Download Queue](assets/Screenshot%202025-05-22%20182239.png)
![Settings Panel](assets/Screenshot%202025-05-22%20182116.png)

## Features

✨ **Modern Web Interface**
- Clean, responsive React-based UI with real-time progress tracking
- Track-based progress display ("Track x of x")
- Live download queue management with status updates
- Album artwork display and download history
- One-click zip archive downloads

🚀 **Production Ready**
- Docker containerization with multi-stage builds
- Health monitoring and graceful restarts
- Persistent storage across container restarts
- High-performance zip creation (600MB+ archives at 37MB/s)

🎵 **Comprehensive Media Support**
- Multiple audio formats (FLAC, ALAC, AAC, MQA, 360 Audio)
- Video downloads (480p to 4K) with FFmpeg integration
- Albums, artists, playlists, livestreams, and webcasts
- Batch processing with queue management

⚡ **Real-Time Updates**
- Server-Sent Events (SSE) for instant progress updates
- Live status notifications for job completion
- No page refresh needed - everything updates automatically

## Quick Start

### Option 1: Docker (Recommended)

1. **Create directories and config:**
```bash
mkdir -p downloads config
```

2. **Create configuration file:**
```bash
cat > config/config.json << EOF
{
  "email": "your-email@example.com",
  "password": "your-password",
  "format": 3,
  "videoFormat": 3,
  "outPath": "/app/downloads",
  "useFfmpegEnvVar": true,
  "token": ""
}
EOF
```

3. **Start the application:**
```bash
docker compose up -d
```

4. **Access the web interface:**
Open http://localhost:8080 in your browser

### Option 2: Development Setup

1. **Prerequisites:**
   - Go 1.21+ installed
   - Node.js 18+ and pnpm
   - FFmpeg installed and in PATH

2. **Setup backend:**
```bash
# Create config file
mkdir -p config
cp config.json config/config.json  # Edit with your credentials

# Start the Go server
go run ./cmd/server/main.go
```

3. **Setup frontend:**
```bash
cd webui
pnpm install
pnpm dev
```

4. **Access the application:**
   - Backend API: http://localhost:8080
   - Frontend Dev: http://localhost:5173

## Configuration

### Format Options

**Audio Formats:**
- `1` - 16-bit / 44.1 kHz ALAC
- `2` - 16-bit / 44.1 kHz FLAC  
- `3` - 24-bit / 48 kHz MQA *(recommended)*
- `4` - 360 Reality Audio / best available
- `5` - 150 Kbps AAC

**Video Formats:**
- `1` - 480p
- `2` - 720p
- `3` - 1080p *(recommended)*
- `4` - 1440p
- `5` - 4K / best available

### Configuration File

Create `config/config.json`:
```json
{
  "email": "your-email@example.com",
  "password": "your-password",
  "format": 3,
  "videoFormat": 3,
  "outPath": "/app/downloads",
  "useFfmpegEnvVar": true,
  "token": ""
}
```

**Options:**
- `email` - Your nugs.net email address
- `password` - Your nugs.net password  
- `format` - Audio download quality (see above)
- `videoFormat` - Video download quality (see above)
- `outPath` - Download directory path
- `useFfmpegEnvVar` - Use FFmpeg from PATH (true) or local directory (false)
- `token` - Optional token for Apple/Google accounts ([how to get token](token.md))

## Supported Media Types

The application supports downloading from various nugs.net URLs:

| Type | Example URL |
|------|-------------|
| **Album** | `https://play.nugs.net/release/23329` |
| **Artist** | `https://play.nugs.net/#/artist/461/latest` |
| **Playlist** | `https://play.nugs.net/#/playlists/playlist/1215400` |
| **Video** | `https://play.nugs.net/#/videos/artist/1045/Dead%20and%20Company/container/27323` |
| **Livestream** | `https://play.nugs.net/watch/livestreams/exclusive/30119` |
| **Webcast** | `https://play.nugs.net/#/my-webcasts/5826189-30369-0-624602` |
| **Catalog** | `https://2nu.gs/3PmqXLW` |

## Usage

### Web Interface

1. **Configure Settings:** Use the settings panel to input your credentials and preferences
2. **Add Downloads:** Paste nugs.net URLs into the download form
3. **Monitor Progress:** Watch real-time progress with track-based updates
4. **Download Files:** Click download buttons to get zip archives of completed jobs
5. **Manage Queue:** Remove completed jobs or retry failed ones

### Command Line (Legacy)

The original CLI interface is still available:

```bash
# Download albums
./nugs-dl https://play.nugs.net/release/23329 https://play.nugs.net/release/23790

# Download with custom settings
./nugs-dl --format 2 --outpath ./downloads https://play.nugs.net/release/23329

# Download from file list
./nugs-dl urls.txt
```

## Docker Deployment

### Production Setup

1. **Create production docker-compose.yml:**
```yaml
version: '3.8'
services:
  nugs-dl:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./downloads:/app/downloads
      - ./config:/app/config:ro
    environment:
      - TZ=America/New_York
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/ping"]
      interval: 30s
      timeout: 10s
      retries: 3
```

2. **Deploy:**
```bash
docker compose up -d --build
```

### Environment Variables

- `TZ` - Set timezone (default: UTC)
- `NUGS_OUTPUT_PATH` - Custom download path (default: /app/downloads)

### Volumes

- `./downloads:/app/downloads` - Persistent download storage
- `./config:/app/config:ro` - Configuration (read-only in production)

## Architecture

### Technology Stack

**Backend:**
- **Language:** Go with Gin web framework
- **Real-time:** Server-Sent Events (SSE) for live updates
- **Storage:** File system with ZIP archive creation
- **Dependencies:** FFmpeg for media processing

**Frontend:**
- **Framework:** React 19 with TypeScript
- **Build Tool:** Vite for fast development and builds
- **Styling:** Tailwind CSS v4 with modern design system
- **Components:** shadcn/ui for consistent UI patterns
- **Real-time:** EventSource API for SSE connection

**Infrastructure:**
- **Containerization:** Docker multi-stage builds
- **Orchestration:** Docker Compose with health checks
- **Runtime:** Alpine Linux for minimal footprint

### API Endpoints

- `GET /api/config` - Get current configuration
- `POST /api/config` - Update configuration  
- `POST /api/download-url` - Submit download URL
- `GET /api/queue` - Get download queue status
- `DELETE /api/queue/:id` - Remove job from queue
- `GET /api/download/:id` - Download completed archive
- `GET /api/status-stream` - SSE endpoint for real-time updates
- `GET /ping` - Health check endpoint

## Troubleshooting

### Common Issues

**FFmpeg not found:**
```bash
# Install FFmpeg
sudo apt install ffmpeg  # Linux
brew install ffmpeg      # macOS
```

**Permission denied:**
```bash
# Fix download directory permissions
sudo chown -R $USER:$USER downloads/
chmod 755 downloads/
```

**Container won't start:**
```bash
# Check logs
docker compose logs nugs-dl

# Verify config file
cat config/config.json
```

**Downloads not appearing:**
- Check your nugs.net credentials
- Verify URL format is supported
- Check browser console for errors

### Performance

- **Memory Usage:** Large downloads (500MB+) require adequate RAM for zip creation
- **Disk Space:** Ensure sufficient space in download directory
- **Network:** Stable internet connection required for streaming service access

## Attribution

This project is built upon the excellent work of [@Sorrow446](https://github.com/Sorrow446) and their [Nugs-Downloader](https://github.com/Sorrow446/Nugs-Downloader) CLI application. The core download logic, nugs.net API integration, and media processing functionality are all derived from their original implementation.

**What we've added:**
- Modern React-based web interface
- Real-time progress tracking with Server-Sent Events
- Docker containerization and production deployment
- Queue management system
- ZIP archive creation and download functionality
- Track-based progress display

Special thanks to the original contributors: [@Sorrow446](https://github.com/Sorrow446), [@twalker1998](https://github.com/twalker1998), [@marksibert](https://github.com/marksibert), and [@khord](https://github.com/khord).

## License & Disclaimer

- This project is for educational and personal use only
- Users are responsible for complying with nugs.net terms of service
- Nugs brand and name are registered trademarks of their respective owners
- This project has no partnership, sponsorship, or endorsement with Nugs

## Contributing

Contributions are welcome! Please ensure:
- Go code follows standard formatting (`go fmt`)
- Frontend code passes TypeScript checks
- Docker builds complete successfully
- All new features include appropriate documentation
