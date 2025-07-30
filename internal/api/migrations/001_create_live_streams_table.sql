-- Migration: Create live_streams table
-- Created: 2025-07-30

CREATE TABLE IF NOT EXISTS live_streams (
    id SERIAL PRIMARY KEY,
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

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_stream_key ON live_streams(stream_key);
CREATE INDEX IF NOT EXISTS idx_created_at ON live_streams(created_at);
CREATE INDEX IF NOT EXISTS idx_status ON live_streams(status); 