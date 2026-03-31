# Photo Sharing

A minimal web app for uploading photos via a shared link (QR code). No database, no authentication, no gallery — just a simple upload button with progress bars.

## Features

- Big centered upload button, mobile-friendly
- Drag-and-drop support on desktop
- Per-file upload progress bars
- MIME type validation (images only)
- Supports files up to 300MB (configurable)
- Lightweight Docker image (~15MB)

## Deployment

### Prerequisites

- Docker and Docker Compose
- Traefik reverse proxy with a network named `traefik`

### Setup

1. Clone the repository
2. Edit `docker-compose.yml` and replace `photos.example.com` with your domain
3. Verify the Traefik network name and certresolver match your setup
4. Start the service:

```bash
docker compose up -d --build
```

Uploaded photos are saved to `./photos/` on the host.

### Configuration

| Variable | Default | Description |
|---|---|---|
| `UPLOAD_DIR` | `/uploads` | Directory where photos are saved |
| `PORT` | `8080` | Server listen port |
| `MAX_FILE_SIZE` | `300` | Max file size in MB |

## Local development

```bash
UPLOAD_DIR=./photos go run main.go
```

Then open [http://localhost:8080](http://localhost:8080).

## License

MIT
