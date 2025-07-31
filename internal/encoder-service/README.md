# StreamKit Encoder Service

The encoder service handles RTMP stream encoding and HLS file generation with MinIO/S3 storage integration.

## Features

- **RTMP to HLS Encoding**: Converts RTMP streams to HLS format using FFmpeg
- **Database Tracking**: Persistent stream status tracking in PostgreSQL
- **MinIO/S3 Storage**: Automatic upload of HLS files to object storage
- **S3-Compatible Serving**: Serve HLS files via signed URLs or CDN
- **Event-Driven**: Responds to publish/unpublish events from RTMP server

## Architecture

```
RTMP Server → Encoder Service → MinIO/S3 → Web Players
     ↓              ↓              ↓
  Events      HLS Files     Signed URLs
```

## API Endpoints

- `POST /events/published` - Handle stream publish/unpublish events
- `GET /health` - Health check
- `GET /stats` - Stream statistics
- `GET /streams/active` - List active streams
- `GET /hls/{stream_key}/playlist.m3u8` - Serve HLS playlist
- `GET /hls/{stream_key}/segment_*.ts` - Serve HLS segments
- `GET /manifest?stream_key={key}` - Get stream manifest

## Environment Variables

### Database
- `DB_HOST` - PostgreSQL host (default: localhost)
- `DB_PORT` - PostgreSQL port (default: 5432)
- `DB_USER` - Database user (default: postgres)
- `DB_PASSWORD` - Database password (default: password)
- `DB_NAME` - Database name (default: streamkit)

### RTMP
- `RTMP_SERVER` - RTMP server host (default: rtmp)
- `RTMP_PORT` - RTMP server port (default: 1935)
- `HLS_OUTPUT_DIR` - Local HLS output directory (default: /tmp/hls)

### MinIO/S3
- `MINIO_ENDPOINT` - MinIO endpoint (default: localhost:9000)
- `MINIO_ACCESS_KEY` - Access key (default: minioadmin)
- `MINIO_SECRET_KEY` - Secret key (default: minioadmin)
- `MINIO_BUCKET` - Bucket name (default: hls-streams)
- `MINIO_REGION` - Region (default: us-east-1)
- `MINIO_USE_SSL` - Use SSL (default: false)

### Service
- `SERVER_PORT` - HTTP server port (default: 8080)
- `CDN_BASE_URL` - CDN base URL for public serving (optional)

## Usage

### Send Publish Event
```bash
curl -X POST http://localhost:8082/events/published \
  -H "Content-Type: application/json" \
  -d '{
    "stream_key": "my-stream",
    "action": "publish",
    "timestamp": "2024-01-15T10:30:00Z"
  }'
```

### Send Unpublish Event
```bash
curl -X POST http://localhost:8082/events/published \
  -H "Content-Type: application/json" \
  -d '{
    "stream_key": "my-stream",
    "action": "unpublish",
    "timestamp": "2024-01-15T10:35:00Z"
  }'
```

### Get Stream Statistics
```bash
curl http://localhost:8082/stats
```

### Play HLS Stream
```html
<video controls>
  <source src="http://localhost:8082/hls/my-stream/playlist.m3u8" type="application/x-mpegURL">
</video>
```

## Docker

The service is designed to run in Docker with the provided docker-compose.yml:

```bash
docker-compose up -d
```

This will start:
- PostgreSQL database
- MinIO object storage
- RTMP server
- Encoder service

## File Structure

```
internal/encoder-service/
├── main.go                    # Application entry point
├── Dockerfile                 # Container configuration
├── README.md                  # This file
├── models/
│   ├── event.go              # Event structures
│   ├── encoder.go            # Encoder structures
│   ├── stream.go             # Database stream model
│   └── storage.go            # Storage configuration
├── repos/
│   └── stream_repo.go        # Database operations
├── service/
│   ├── encoder_service.go    # Encoding business logic
│   └── storage_service.go    # MinIO/S3 operations
├── handlers/
│   ├── event_handler.go      # Event webhook handler
│   └── hls_handler.go        # HLS serving handler
└── migrations/
    ├── 001_create_streams_table.sql
    └── run_migrations.sh
``` 