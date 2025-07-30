package repos

import (
	"database/sql"
	"errors"
	"time"

	"streamkit/internal/api/models"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type StreamRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewStreamRepository(db *sql.DB, logger *zap.Logger) *StreamRepository {
	return &StreamRepository{db: db, logger: logger}
}

// Create creates a new stream
func (r *StreamRepository) Create(stream *models.LiveStream) error {
	r.logger.Info("Creating stream",
		zap.String("title", stream.Title),
		zap.String("stream_name", stream.StreamName),
		zap.String("created_by", stream.StreamCreatedBy),
	)

	stream.StreamKey = uuid.New().String()
	stream.CreatedAt = time.Now()
	stream.Status = "inactive"

	r.logger.Info("Generated stream key", zap.String("stream_key", stream.StreamKey))

	query := `
		INSERT INTO live_streams (stream_key, ingest_url, playback_url, title, stream_name, stream_created_by, description, created_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	r.logger.Debug("Executing query",
		zap.String("query", query),
		zap.String("stream_key", stream.StreamKey),
		zap.String("title", stream.Title),
		zap.String("stream_name", stream.StreamName),
	)

	var id int
	err := r.db.QueryRow(query,
		stream.StreamKey,
		stream.IngestURL,
		stream.PlaybackURL,
		stream.Title,
		stream.StreamName,
		stream.StreamCreatedBy,
		stream.Description,
		stream.CreatedAt,
		stream.Status,
	).Scan(&id)
	if err != nil {
		r.logger.Error("Error creating stream",
			zap.Error(err),
			zap.String("stream_key", stream.StreamKey),
			zap.String("title", stream.Title),
		)
		return err
	}

	stream.ID = id
	r.logger.Info("Successfully created stream",
		zap.Int("id", id),
		zap.String("stream_key", stream.StreamKey),
	)
	return nil
}

// GetByID retrieves a stream by ID
func (r *StreamRepository) GetByID(id int) (*models.LiveStream, error) {
	r.logger.Info("Getting stream by ID", zap.Int("id", id))

	stream := &models.LiveStream{}
	query := `
		SELECT id, stream_key, ingest_url, playback_url, title, stream_name, stream_created_by, description, created_at, status
		FROM live_streams WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&stream.ID,
		&stream.StreamKey,
		&stream.IngestURL,
		&stream.PlaybackURL,
		&stream.Title,
		&stream.StreamName,
		&stream.StreamCreatedBy,
		&stream.Description,
		&stream.CreatedAt,
		&stream.Status,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("Stream not found", zap.Int("id", id))
			return nil, errors.New("stream not found")
		}
		r.logger.Error("Error getting stream by ID",
			zap.Int("id", id),
			zap.Error(err),
		)
		return nil, err
	}

	r.logger.Info("Successfully retrieved stream", zap.Int("id", id))
	return stream, nil
}

// GetByStreamKey retrieves a stream by stream key
func (r *StreamRepository) GetByStreamKey(streamKey string) (*models.LiveStream, error) {
	r.logger.Info("Getting stream by stream key", zap.String("stream_key", streamKey))

	stream := &models.LiveStream{}
	query := `
		SELECT id, stream_key, ingest_url, playback_url, title, stream_name, stream_created_by, description, created_at, status
		FROM live_streams WHERE stream_key = $1
	`

	err := r.db.QueryRow(query, streamKey).Scan(
		&stream.ID,
		&stream.StreamKey,
		&stream.IngestURL,
		&stream.PlaybackURL,
		&stream.Title,
		&stream.StreamName,
		&stream.StreamCreatedBy,
		&stream.Description,
		&stream.CreatedAt,
		&stream.Status,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("Stream not found", zap.String("stream_key", streamKey))
			return nil, errors.New("stream not found")
		}
		r.logger.Error("Error getting stream by key",
			zap.String("stream_key", streamKey),
			zap.Error(err),
		)
		return nil, err
	}

	r.logger.Info("Successfully retrieved stream", zap.String("stream_key", streamKey))
	return stream, nil
}

// GetAll retrieves all streams
func (r *StreamRepository) GetAll() ([]*models.LiveStream, error) {
	r.logger.Info("Getting all streams")

	query := `
		SELECT id, stream_key, ingest_url, playback_url, title, stream_name, stream_created_by, description, created_at, status
		FROM live_streams ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		r.logger.Error("Error getting all streams", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var streams []*models.LiveStream
	for rows.Next() {
		stream := &models.LiveStream{}
		err := rows.Scan(
			&stream.ID,
			&stream.StreamKey,
			&stream.IngestURL,
			&stream.PlaybackURL,
			&stream.Title,
			&stream.StreamName,
			&stream.StreamCreatedBy,
			&stream.Description,
			&stream.CreatedAt,
			&stream.Status,
		)
		if err != nil {
			r.logger.Error("Error scanning stream row", zap.Error(err))
			return nil, err
		}
		streams = append(streams, stream)
	}

	r.logger.Info("Successfully retrieved streams", zap.Int("count", len(streams)))
	return streams, nil
}

// Update updates a stream
func (r *StreamRepository) Update(stream *models.LiveStream) error {
	r.logger.Info("Updating stream",
		zap.Int("id", stream.ID),
		zap.String("title", stream.Title),
	)

	query := `
		UPDATE live_streams 
		SET title = $1, stream_name = $2, stream_created_by = $3, description = $4, status = $5
		WHERE id = $6
	`

	result, err := r.db.Exec(query,
		stream.Title,
		stream.StreamName,
		stream.StreamCreatedBy,
		stream.Description,
		stream.Status,
		stream.ID,
	)
	if err != nil {
		r.logger.Error("Error updating stream",
			zap.Int("id", stream.ID),
			zap.Error(err),
		)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Error getting rows affected", zap.Error(err))
		return err
	}

	if rowsAffected == 0 {
		r.logger.Warn("No stream found to update", zap.Int("id", stream.ID))
		return errors.New("stream not found")
	}

	r.logger.Info("Successfully updated stream", zap.Int("id", stream.ID))
	return nil
}

// Delete deletes a stream by ID
func (r *StreamRepository) Delete(id int) error {
	r.logger.Info("Deleting stream", zap.Int("id", id))

	query := `DELETE FROM live_streams WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		r.logger.Error("Error deleting stream",
			zap.Int("id", id),
			zap.Error(err),
		)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Error getting rows affected", zap.Error(err))
		return err
	}

	if rowsAffected == 0 {
		r.logger.Warn("No stream found to delete", zap.Int("id", id))
		return errors.New("stream not found")
	}

	r.logger.Info("Successfully deleted stream", zap.Int("id", id))
	return nil
}

// UpdateStatus updates the status of a stream
func (r *StreamRepository) UpdateStatus(id int, status string) error {
	r.logger.Info("Updating stream status",
		zap.Int("id", id),
		zap.String("status", status),
	)

	query := `UPDATE live_streams SET status = $1 WHERE id = $2`

	result, err := r.db.Exec(query, status, id)
	if err != nil {
		r.logger.Error("Error updating stream status",
			zap.Int("id", id),
			zap.String("status", status),
			zap.Error(err),
		)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Error getting rows affected", zap.Error(err))
		return err
	}

	if rowsAffected == 0 {
		r.logger.Warn("No stream found to update status", zap.Int("id", id))
		return errors.New("stream not found")
	}

	r.logger.Info("Successfully updated stream status",
		zap.Int("id", id),
		zap.String("status", status),
	)
	return nil
}
