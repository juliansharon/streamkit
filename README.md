# StreamKit

A complete live streaming platform with RTMP server, Go API, and PostgreSQL database.

## 🚀 Features

- **RTMP Server** - NGINX-RTMP for live streaming
- **HLS Playback** - HTTP Live Streaming for web playback
- **Go API** - RESTful API for stream management
- **PostgreSQL** - Database for stream metadata
- **Web Player** - HTML5 player for viewing streams
- **CRUD Operations** - Complete stream management

## 📋 Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)

## 🏃‍♂️ Quick Start

### 1. Clone and Setup

```bash
git clone <repository-url>
cd streamkit
```

### 2. Start All Services

```bash
docker-compose up -d
```

This will start:
- **PostgreSQL** on port 5432
- **Go API** on port 8080
- **RTMP Server** on port 1935
- **RTMP HTTP** on port 8081

### 3. Check Services

```bash
# Check if all containers are running
docker-compose ps

# Check API health
curl http://localhost:8080/health

# Check RTMP stats
curl http://localhost:8081/stat
```

## 📚 API Usage

### Create a Stream

```bash
curl -X POST http://localhost:8080/api/streams \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My Gaming Stream",
    "stream_name": "gaming",
    "stream_created_by": "gamer123",
    "description": "Live gaming session"
  }'
```

### Get All Streams

```bash
curl http://localhost:8080/api/streams
```

### Get Stream by ID

```bash
curl http://localhost:8080/api/streams/1
```

## 🎥 Streaming with OBS

1. **Create a stream** using the API
2. **In OBS Studio:**
   - Server: `rtmp://localhost:1935/live`
   - Stream Key: `{stream_key_from_api}`
3. **Start streaming**

## 📺 Viewing Streams

- **Web Player**: `http://localhost:8081/player.html`
- **Direct HLS**: `http://localhost:8081/hls/{stream_key}.m3u8`
- **Stats Page**: `http://localhost:8081/stat`

## 🛠️ Development

### Local Development

1. **Install Go dependencies:**
```bash
go mod tidy
```

2. **Set up environment variables:**
```bash
# Create .env file with your configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=streamkit
PORT=8080
RTMP_HOST=localhost
```

3. **Run the API locally:**
```bash
go run cmd/server/main.go
```

### Database Migrations

The database schema is automatically created when PostgreSQL starts up. The migration file is located at:
`internal/api/migrations/001_create_live_streams_table.sql`

## 📁 Project Structure

```
streamkit/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/        # HTTP handlers
│   │   ├── models/          # Data models
│   │   ├── repos/           # Database repositories
│   │   ├── routes/          # Route definitions
│   │   ├── service/         # Business logic
│   │   └── migrations/      # Database migrations
│   └── RTMP-server/
│       ├── Dockerfile       # Custom RTMP server Dockerfile
│       ├── nginx.conf       # RTMP server configuration
│       ├── player.html      # Web player
│       └── hls/             # HLS files directory
├── docker-compose.yml       # All services configuration
├── Dockerfile               # API server Dockerfile
├── go.mod                   # Go module definition
└── README.md
```

## 🔧 Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `password` | Database password |
| `DB_NAME` | `streamkit` | Database name |
| `PORT` | `8080` | API server port |
| `RTMP_HOST` | `localhost` | RTMP server host |

### Ports

| Service | Port | Description |
|---------|------|-------------|
| API Server | 8080 | Go API endpoints |
| RTMP Server | 1935 | RTMP ingest |
| RTMP HTTP | 8081 | HLS playback & stats |
| PostgreSQL | 5432 | Database |

## 🐛 Troubleshooting

### Common Issues

1. **Port conflicts**: Make sure ports 8080, 8081, 1935, and 5432 are available
2. **Database connection**: Wait for PostgreSQL to fully start before the API
3. **RTMP not working**: Check if the nginx-rtmp container is running

### Logs

```bash
# View all logs
docker-compose logs

# View specific service logs
docker-compose logs api
docker-compose logs rtmp
docker-compose logs postgres
```

### Useful Commands

```bash
# Start all services
docker-compose up -d

# Stop all services
docker-compose down

# Rebuild and start
docker-compose up --build -d

# View running containers
docker-compose ps

# Access PostgreSQL
docker-compose exec postgres psql -U postgres -d streamkit
```

## 📖 API Documentation

Complete API documentation is available at:
`internal/api/README.md`

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License. 