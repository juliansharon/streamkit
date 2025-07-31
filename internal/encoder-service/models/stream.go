package models

import (
	"time"
)

// StreamStatus represents the status of a stream
type StreamStatus string

const (
	StreamStatusActive   StreamStatus = "active"
	StreamStatusInactive StreamStatus = "inactive"
	StreamStatusError    StreamStatus = "error"
)

// Stream represents a stream in the database
type Stream struct {
	ID        int64        `json:"id"         db:"id"`
	StreamKey string       `json:"stream_key" db:"stream_key"`
	Status    StreamStatus `json:"status"     db:"status"`
	StartedAt *time.Time   `json:"started_at" db:"started_at"`
	StoppedAt *time.Time   `json:"stopped_at" db:"stopped_at"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
}

// StreamStats represents stream statistics
type StreamStats struct {
	TotalStreams    int64 `json:"total_streams"`
	ActiveStreams   int64 `json:"active_streams"`
	InactiveStreams int64 `json:"inactive_streams"`
	ErrorStreams    int64 `json:"error_streams"`
}
