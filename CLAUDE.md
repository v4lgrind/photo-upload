# Photo Sharing

Simple photo upload web app deployed via Docker behind Traefik.

## Stack
- **Backend**: Go (standard library only) — single binary, no dependencies
- **Frontend**: Single HTML file with inline CSS/JS

## Project structure
```
go.mod              # Go module
main.go             # HTTP server: serves static files + handles uploads
static/index.html   # Frontend (upload UI, progress bars, drag-and-drop)
Dockerfile          # Multi-stage build (golang:1.22-alpine → alpine:3.19)
docker-compose.yml  # Service config, Traefik labels, volume mapping
```

## Local development
```bash
UPLOAD_DIR=./photos go run main.go
# Open http://localhost:8080
```

## Environment variables
- `UPLOAD_DIR` — where photos are saved (default: `/uploads`)
- `PORT` — server port (default: `8080`)
- `MAX_FILE_SIZE` — max file size in MB (default: `300`)

## Deployment
```bash
docker compose up -d --build
```
Photos are stored in `./photos/` on the host via Docker volume.
