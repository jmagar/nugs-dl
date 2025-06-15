# Multi-stage build for nugs-dl

# Stage 1: Build React frontend
FROM node:18-alpine AS frontend-builder

WORKDIR /app/webui

# Copy package files
COPY webui/package*.json ./
COPY webui/pnpm-lock.yaml* ./

# Install dependencies
RUN npm install -g pnpm && pnpm install

# Copy frontend source
COPY webui/ ./

# Build frontend
RUN pnpm build

# Stage 2: Build Go backend  
FROM golang:1.23-alpine AS backend-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o nugs-server ./cmd/server

# Stage 3: Final runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ffmpeg \
    ca-certificates \
    tzdata

# Create non-root user
RUN addgroup -S nugsuser && adduser -S nugsuser -G nugsuser

WORKDIR /app

# Copy built backend
COPY --from=backend-builder /app/nugs-server .

# Copy built frontend
COPY --from=frontend-builder /app/webui/dist ./webui/dist

# Create directories for downloads and config
RUN mkdir -p /app/downloads /app/config && \
    chown -R nugsuser:nugsuser /app

# Switch to non-root user
USER nugsuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ping || exit 1

# Run the application
CMD ["./nugs-server"] 