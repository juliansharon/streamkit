# Stream API Documentation

This API provides CRUD operations for managing live streams.

## Environment Variables

Set the following environment variable to configure the RTMP server host:

```bash
export RTMP_HOST=localhost  # or your server IP
```

## API Endpoints

### Create Stream
**POST** `/api/streams`

Creates a new stream with auto-generated stream key and URLs.

**Request Body:**
```json
{
  "title": "My Live Stream",
  "stream_name": "my-stream",
  "stream_created_by": "user123",
  "description": "Optional description"
}
```

**Response:**
```json
{
  "id": 1,
  "stream_key": "550e8400-e29b-41d4-a716-446655440000",
  "ingest_url": "rtmp://localhost/live",
  "playback_url": "http://localhost:8080/hls/550e8400-e29b-41d4-a716-446655440000.m3u8",
  "title": "My Live Stream",
  "stream_name": "my-stream",
  "stream_created_by": "user123",
  "description": "Optional description",
  "created_at": "2025-07-30T22:00:00Z",
  "status": "inactive"
}
```

### Get All Streams
**GET** `/api/streams`

Returns all streams with full URLs.

**Response:**
```json
[
  {
    "id": 1,
    "stream_key": "550e8400-e29b-41d4-a716-446655440000",
    "ingest_url": "rtmp://localhost/live",
    "playback_url": "http://localhost:8080/hls/550e8400-e29b-41d4-a716-446655440000.m3u8",
    "title": "My Live Stream",
    "stream_name": "my-stream",
    "stream_created_by": "user123",
    "description": "Optional description",
    "created_at": "2025-07-30T22:00:00Z",
    "status": "active"
  }
]
```

### Get Stream by ID
**GET** `/api/streams/{id}`

Returns a specific stream by ID.

**Response:** Same as Create Stream response.

### Get Stream by Stream Key
**GET** `/api/streams/key/{streamKey}`

Returns a specific stream by stream key.

**Response:** Same as Create Stream response.

### Update Stream
**PUT** `/api/streams/{id}`

Updates a stream's information.

**Request Body:**
```json
{
  "title": "Updated Stream Title",
  "stream_name": "updated-stream-name",
  "stream_created_by": "user123",
  "description": "Updated description",
  "status": "active"
}
```

**Response:** Same as Create Stream response.

### Delete Stream
**DELETE** `/api/streams/{id}`

Deletes a stream.

**Response:** 204 No Content

### Update Stream Status
**PATCH** `/api/streams/{id}/status`

Updates only the status of a stream.

**Request Body:**
```json
{
  "status": "active"
}
```

**Response:** Same as Create Stream response.

## Usage Examples

### Creating a Stream for OBS

1. **Create a stream:**
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

2. **Use the returned stream key in OBS:**
   - Server: `rtmp://localhost/live`
   - Stream Key: `550e8400-e29b-41d4-a716-446655440000`

3. **Play the stream:**
   - HLS URL: `http://localhost:8080/hls/550e8400-e29b-41d4-a716-446655440000.m3u8`
   - Or use the player: `http://localhost:8080/player.html`

## Database Schema

The `live_streams` table has the following structure:

```sql
CREATE TABLE live_streams (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    stream_key VARCHAR(255) NOT NULL UNIQUE,
    ingest_url VARCHAR(500) NOT NULL,
    playback_url VARCHAR(500) NOT NULL,
    title VARCHAR(255) NOT NULL,
    stream_name VARCHAR(255) NOT NULL,
    stream_created_by VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) DEFAULT 'inactive'
);
```

## Error Responses

All endpoints return appropriate HTTP status codes:

- `200 OK` - Success
- `201 Created` - Resource created
- `204 No Content` - Success (no body)
- `400 Bad Request` - Invalid request
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error 