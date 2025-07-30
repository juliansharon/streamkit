package models

import "time"

type LiveStream struct {
	ID              int       `json:"id"`
	StreamKey       string    `json:"stream_key"`
	IngestURL       string    `json:"ingest_url"`
	PlaybackURL     string    `json:"playback_url"`
	Title           string    `json:"title"`
	StreamName      string    `json:"stream_name"`
	StreamCreatedBy string    `json:"stream_created_by"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
	Status          string    `json:"status"`
}
