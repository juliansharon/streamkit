package repos

import (
	"database/sql"
	"time"

	"go.uber.org/zap"

	"streamkit/internal/encoder-service/models"
)

// StreamRepo handles database operations for streams
type StreamRepo struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewStreamRepo creates a new stream repository
func NewStreamRepo(db *sql.DB, logger *zap.Logger) *StreamRepo {
	return &StreamRepo{
		db:     db,
		logger: logger,
	}
}

// CreateStream creates a new stream record
func (r *StreamRepo) CreateStream(streamKey string) (*models.Stream, error) {
	now := time.Now()

	query := `
		INSERT INTO streams (stream_key, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, stream_key, status, started_at, stopped_at, created_at, updated_at
	`

	stream := &models.Stream{}
	err := r.db.QueryRow(
		query,
		streamKey,
		models.StreamStatusInactive,
		now,
		now,
	).Scan(
		&stream.ID,
		&stream.StreamKey,
		&stream.Status,
		&stream.StartedAt,
		&stream.StoppedAt,
		&stream.CreatedAt,
		&stream.UpdatedAt,
	)
	if err != nil {
		r.logger.Error("Failed to create stream",
			zap.String("stream_key", streamKey),
			zap.Error(err),
		)
		return nil, err
	}

	r.logger.Info("Created stream record",
		zap.String("stream_key", streamKey),
		zap.Int64("id", stream.ID),
	)

	return stream, nil
}

// StartStream marks a stream as active
func (r *StreamRepo) StartStream(streamKey string) error {
	now := time.Now()

	query := `
		UPDATE streams 
		SET status = $1, started_at = $2, updated_at = $3
		WHERE stream_key = $4
	`

	result, err := r.db.Exec(query, models.StreamStatusActive, now, now, streamKey)
	if err != nil {
		r.logger.Error("Failed to start stream",
			zap.String("stream_key", streamKey),
			zap.Error(err),
		)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Stream doesn't exist, create it
		_, err = r.CreateStream(streamKey)
		if err != nil {
			return err
		}
		// Try to start it again
		_, err = r.db.Exec(query, models.StreamStatusActive, now, now, streamKey)
		if err != nil {
			return err
		}
	}

	r.logger.Info("Started stream", zap.String("stream_key", streamKey))
	return nil
}

// StopStream marks a stream as inactive
func (r *StreamRepo) StopStream(streamKey string) error {
	now := time.Now()

	query := `
		UPDATE streams 
		SET status = $1, stopped_at = $2, updated_at = $3
		WHERE stream_key = $4
	`

	_, err := r.db.Exec(query, models.StreamStatusInactive, now, now, streamKey)
	if err != nil {
		r.logger.Error("Failed to stop stream",
			zap.String("stream_key", streamKey),
			zap.Error(err),
		)
		return err
	}

	r.logger.Info("Stopped stream", zap.String("stream_key", streamKey))
	return nil
}

// GetActiveStreams returns all active streams
func (r *StreamRepo) GetActiveStreams() ([]*models.Stream, error) {
	query := `
		SELECT id, stream_key, status, started_at, stopped_at, created_at, updated_at
		FROM streams 
		WHERE status = $1
		ORDER BY updated_at DESC
	`

	rows, err := r.db.Query(query, models.StreamStatusActive)
	if err != nil {
		r.logger.Error("Failed to get active streams", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var streams []*models.Stream
	for rows.Next() {
		stream := &models.Stream{}
		err := rows.Scan(
			&stream.ID,
			&stream.StreamKey,
			&stream.Status,
			&stream.StartedAt,
			&stream.StoppedAt,
			&stream.CreatedAt,
			&stream.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan stream", zap.Error(err))
			continue
		}
		streams = append(streams, stream)
	}

	return streams, nil
}

// GetStreamStats returns stream statistics
func (r *StreamRepo) GetStreamStats() (*models.StreamStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_streams,
			COUNT(CASE WHEN status = $1 THEN 1 END) as active_streams,
			COUNT(CASE WHEN status = $2 THEN 1 END) as inactive_streams,
			COUNT(CASE WHEN status = $3 THEN 1 END) as error_streams
		FROM streams
	`

	stats := &models.StreamStats{}
	err := r.db.QueryRow(
		query,
		models.StreamStatusActive,
		models.StreamStatusInactive,
		models.StreamStatusError,
	).Scan(
		&stats.TotalStreams,
		&stats.ActiveStreams,
		&stats.InactiveStreams,
		&stats.ErrorStreams,
	)
	if err != nil {
		r.logger.Error("Failed to get stream stats", zap.Error(err))
		return nil, err
	}

	return stats, nil
}

// GetStreamByKey returns a stream by stream key
func (r *StreamRepo) GetStreamByKey(streamKey string) (*models.Stream, error) {
	query := `
		SELECT id, stream_key, status, started_at, stopped_at, created_at, updated_at
		FROM streams 
		WHERE stream_key = $1
	`

	stream := &models.Stream{}
	err := r.db.QueryRow(query, streamKey).Scan(
		&stream.ID,
		&stream.StreamKey,
		&stream.Status,
		&stream.StartedAt,
		&stream.StoppedAt,
		&stream.CreatedAt,
		&stream.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Failed to get stream by key",
			zap.String("stream_key", streamKey),
			zap.Error(err),
		)
		return nil, err
	}

	return stream, nil
}
