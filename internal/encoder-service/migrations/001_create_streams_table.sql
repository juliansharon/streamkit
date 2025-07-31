-- Create streams table
CREATE TABLE IF NOT EXISTS streams (
    id BIGSERIAL PRIMARY KEY,
    stream_key VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'inactive',
    started_at TIMESTAMP,
    stopped_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on stream_key for faster lookups
CREATE INDEX IF NOT EXISTS idx_streams_stream_key ON streams(stream_key);

-- Create index on status for filtering
CREATE INDEX IF NOT EXISTS idx_streams_status ON streams(status);

-- Create index on updated_at for sorting
CREATE INDEX IF NOT EXISTS idx_streams_updated_at ON streams(updated_at DESC); 