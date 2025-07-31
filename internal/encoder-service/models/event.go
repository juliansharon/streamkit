package models

import "time"

// PublishedEvent represents the event structure from RTMP server
type PublishedEvent struct {
	StreamKey string    `json:"stream_key"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}
