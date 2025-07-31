# StreamKit - RTMP Streaming Server with Go API

A complete streaming solution with Go-based HLS encoding, built with Go API and PostgreSQL.

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   OBS Studio   │───▶│  RTMP Server    │───▶│  Go API        │
│   (Client)     │    │  (Ingest Only)  │    │  (HLS Encoder)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │                        │
                              ▼                        ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │  PostgreSQL     │    │  HLS Files      │
                       │  (Database)     │    │  (Shared Volume)│
                       └─────────────────┘    └─────────────────┘
```

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- Git

### 1. Clone and Setup
```bash
git clone <your-repo-url>
cd streamkit
```

### 2. Start Services
```bash
docker-compose up -d
```

### 3. Verify Services
```bash
# Check all services are running
docker-compose ps

# Check API health
curl http://localhost:8080/health

# Check RTMP server stats
curl http://localhost:8081/stat
```

## 📋 Services

### 1. **RTMP Server** (`streamkit-rtmp`)
- **Port**: 1935 (RTMP), 8081 (HTTP)
- **Purpose**: RTMP ingest only
- **URL**: `rtmp://localhost:1935/live/{stream_key}`

### 2. **Go API** (`streamkit-api`)
- **Port**: 8080
- **Purpose**: Stream management, metadata, and HLS encoding
- **Features**: 
  - CRUD operations
  - UUID stream keys
  - Go-based FFmpeg encoding
  - Automatic stream encoding

### 3. **PostgreSQL** (`streamkit-postgres`)
- **Port**: 5432
- **Purpose**: Stream metadata storage

## 🎯 Usage

### 1. Create a Stream
```bash
curl -X POST http://localhost:8080/api/streams \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My Gaming Stream",
    "stream_name": "gaming-stream",
    "stream_created_by": "gamer123",
    "description": "Live gaming session"
  }'
```

### 2. Configure OBS Studio
- **Server**: `rtmp://localhost:1935/live`
- **Stream Key**: Use the `stream_key` from the API response

### 3. View Stream
- **HLS URL**: `http://localhost:8081/hls/{stream_key}/playlist.m3u8`
- **Player**: `http://localhost:8081/player`

## 🔧 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | API health check |
| `POST` | `/api/streams` | Create new stream |
| `GET` | `/api/streams` | Get all streams |
| `GET` | `/api/streams/{id}` | Get stream by ID |
| `GET` | `/api/streams/key/{stream_key}` | Get stream by key |
| `PUT` | `/api/streams/{id}` | Update stream |
| `DELETE` | `/api/streams/{id}` | Delete stream |
| `PATCH` | `/api/streams/{id}/status` | Update stream status |
| `POST` | `/api/streams/{stream_key}/encode/start` | Start encoding |
| `POST` | `/api/streams/{stream_key}/encode/stop` | Stop encoding |
| `GET` | `/api/streams/{stream_key}/encode/status` | Get encoding status |

## 📁 Project Structure

```
streamkit/
├── cmd/server/main.go          # Go API entry point
├── internal/
│   ├── api/                    # Go API layers
│   │   ├── handlers/           # HTTP handlers
│   │   ├── models/             # Data models
│   │   ├── repos/              # Database layer
│   │   ├── routes/             # Route definitions
│   │   └── service/            # Business logic
│   ├── RTMP-server/            # RTMP server
│   │   ├── Dockerfile          # RTMP container
│   │   ├── nginx.conf          # NGINX config
│   │   └── player.html         # Web player
│   └── encoder/                # Go HLS encoder
│       └── encoder.go          # Encoding logic
├── docker-compose.yml          # Multi-service setup
├── Dockerfile                  # Go API container (with FFmpeg)
├── go.mod                      # Go dependencies
└── README.md                   # This file
```

## 🔍 Monitoring

### RTMP Server Stats
```bash
curl http://localhost:8081/stat
```

### API Logs
```bash
docker-compose logs api
```

### RTMP Server Logs
```bash
docker-compose logs rtmp
```

## 🛠️ Development

### Rebuild Services
```bash
# Rebuild all services
docker-compose build

# Rebuild specific service
docker-compose build api
docker-compose build rtmp
```

### Database Migrations
```bash
# Access PostgreSQL
docker-compose exec postgres psql -U postgres -d streamkit

# Run migration manually
docker-compose exec postgres psql -U postgres -d streamkit -f /tmp/migration.sql
```

### Testing
```bash
# Test API
curl http://localhost:8080/health

# Test RTMP server
curl http://localhost:8081/health

# Test stream creation
curl -X POST http://localhost:8080/api/streams \
  -H "Content-Type: application/json" \
  -d '{"title":"Test","stream_name":"test","stream_created_by":"user"}'

# Test encoding control
curl -X POST http://localhost:8080/api/streams/{stream_key}/encode/start
curl -X GET http://localhost:8080/api/streams/{stream_key}/encode/status
```

## 📦 Features

- ✅ **Go-based Encoding**: HLS encoding handled by Go service with FFmpeg
- ✅ **Auto-scaling**: Each stream gets its own encoding process
- ✅ **UUID Stream Keys**: Secure, unique stream identifiers
- ✅ **RESTful API**: Complete CRUD operations
- ✅ **Structured Logging**: Zap logger throughout
- ✅ **Docker Compose**: Easy deployment
- ✅ **PostgreSQL**: Reliable data storage
- ✅ **CORS Support**: Web-friendly API
- ✅ **Health Checks**: Service monitoring
- ✅ **Encoding Control**: Manual start/stop/status endpoints

## 🚀 Production Considerations

1. **Environment Variables**: Set proper `RTMP_HOST` for production
2. **SSL/TLS**: Add HTTPS for production
3. **Load Balancing**: Consider multiple RTMP servers
4. **CDN**: Use CDN for HLS delivery
5. **Monitoring**: Add Prometheus/Grafana
6. **Backup**: Database backup strategy
7. **Resource Limits**: Set CPU/memory limits for encoding processes

## 📄 License

MIT License - see LICENSE file for details. 