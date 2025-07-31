package models

import "time"

// StorageConfig represents MinIO/S3 storage configuration
type StorageConfig struct {
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	UseSSL          bool   `json:"use_ssl"`
	BucketName      string `json:"bucket_name"`
	Region          string `json:"region"`
}

// StorageFile represents a file stored in MinIO/S3
type StorageFile struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	ETag         string    `json:"etag"`
	ContentType  string    `json:"content_type"`
}

// HLSManifest represents HLS playlist structure
type HLSManifest struct {
	StreamKey     string       `json:"stream_key"`
	PlaylistURL   string       `json:"playlist_url"`
	Segments      []HLSSegment `json:"segments"`
	TotalDuration float64      `json:"total_duration"`
}

// HLSSegment represents an HLS segment
type HLSSegment struct {
	URL      string  `json:"url"`
	Duration float64 `json:"duration"`
	Size     int64   `json:"size"`
}
