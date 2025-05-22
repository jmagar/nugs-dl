# Docker Setup for nugs-dl

This document explains how to run nugs-dl using Docker and Docker Compose.

## Quick Start

### 1. Prerequisites
- Docker and Docker Compose installed
- Your nugs.net credentials

### 2. Create directories
```bash
mkdir -p downloads config
```

### 3. Create config.json
Create a `config.json` file in the `config/` directory:
```bash
# Create the config file
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

### 4. Build and run
```bash
# Build and start the container
docker compose up -d

# View logs
docker compose logs -f nugs-dl

# Stop the container
docker compose down
```

### 5. Access the application
Open http://localhost:8080 in your browser.

## Configuration

### Environment Variables
- `TZ`: Set timezone (default: UTC)
- `NUGS_OUTPUT_PATH`: Custom download path (default: /app/downloads)

### Volume Mounts
- `./downloads:/app/downloads` - Downloaded files persist here
- `./config:/app/config` - Configuration directory (includes config.json)

## Building from Source

### Build the Docker image manually:
```bash
docker build -t nugs-dl .
```

### Run without docker compose:
```bash
docker run -d \
  --name nugs-dl \
  -p 8080:8080 \
  -v $(pwd)/downloads:/app/downloads \
  -v $(pwd)/config:/app/config \
  --restart unless-stopped \
  nugs-dl
```

## Production Deployment

### With Reverse Proxy
For production use, consider adding a reverse proxy for HTTPS:

1. Uncomment the nginx service in `docker compose.yml`
2. Create `nginx.conf`:
```nginx
events {
    worker_connections 1024;
}

http {
    upstream nugs-dl {
        server nugs-dl:8080;
    }
    
    server {
        listen 80;
        server_name your-domain.com;
        
        location / {
            proxy_pass http://nugs-dl;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```

### Security Considerations
- The container runs as a non-root user (`nugsuser`)
- Config file is mounted read-only
- Consider using Docker secrets for credentials in production
- Enable firewall rules to restrict access to port 8080

## Troubleshooting

### Check container health
```bash
docker compose ps
docker inspect nugs-dl-container --format='{{.State.Health.Status}}'
```

### View container logs
```bash
docker compose logs nugs-dl
```

### Debug inside container
```bash
docker compose exec nugs-dl sh
```

### Common Issues

**Build fails on frontend:**
- Ensure Node.js 18+ is available
- Check for missing package files

**FFmpeg not found:**
- FFmpeg is included in the Alpine base image
- Check if container started properly

**Config not loading:**
- Verify config.json syntax
- Check file permissions
- Ensure volume mount is correct

**Downloads not persisting:**
- Check volume mount for downloads directory
- Verify directory permissions

## Updates

To update nugs-dl:
```bash
# Pull latest code
git pull

# Rebuild and restart
docker compose down
docker compose up -d --build
```

## Monitoring

### Health Check
The container includes a health check that pings `/ping` endpoint every 30 seconds.

### Resource Usage
Monitor with:
```bash
docker stats nugs-dl
``` 