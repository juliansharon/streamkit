package service

import (
	"fmt"
	"os"

	"streamkit/internal/api/models"
	"streamkit/internal/api/repos"

	"go.uber.org/zap"
)

type StreamService struct {
	repo   *repos.StreamRepository
	logger *zap.Logger
}

func NewStreamService(repo *repos.StreamRepository, logger *zap.Logger) *StreamService {
	logger.Info("Initializing StreamService")
	return &StreamService{repo: repo, logger: logger}
}

// CreateStream creates a new stream with auto-generated URLs
func (s *StreamService) CreateStream(stream *models.LiveStream) error {
	s.logger.Info("Creating stream",
		zap.String("title", stream.Title),
		zap.String("stream_name", stream.StreamName),
		zap.String("created_by", stream.StreamCreatedBy),
	)

	// Get host from environment variable
	host := os.Getenv("RTMP_HOST")
	if host == "" {
		host = "localhost"
	}

	streamHttpHost := os.Getenv("STREAM_HTTP_HOST")
	if streamHttpHost == "" {
		streamHttpHost = "localhost"
	}

	s.logger.Info("Using RTMP_HOST", zap.String("host", host))

	// Generate URLs based on stream key (will be set by repository)
	stream.IngestURL = fmt.Sprintf("rtmp://%s/live", host)
	stream.PlaybackURL = fmt.Sprintf("http://%s:8080/hls/{stream_key}.m3u8", host)

	s.logger.Info("Generated URLs",
		zap.String("ingest_url", stream.IngestURL),
		zap.String("playback_url", stream.PlaybackURL),
	)

	err := s.repo.Create(stream)
	if err != nil {
		s.logger.Error("Error creating stream", zap.Error(err))
		return err
	}

	s.logger.Info("Successfully created stream",
		zap.Int("id", stream.ID),
		zap.String("stream_key", stream.StreamKey),
	)
	return nil
}

// GetStreamByID retrieves a stream by ID
func (s *StreamService) GetStreamByID(id int) (*models.LiveStream, error) {
	s.logger.Info("Getting stream by ID", zap.Int("id", id))

	stream, err := s.repo.GetByID(id)
	if err != nil {
		s.logger.Error("Error getting stream by ID",
			zap.Int("id", id),
			zap.Error(err),
		)
		return nil, err
	}

	s.logger.Info("Successfully retrieved stream by ID", zap.Int("id", id))
	return stream, nil
}

// GetStreamByStreamKey retrieves a stream by stream key
func (s *StreamService) GetStreamByStreamKey(streamKey string) (*models.LiveStream, error) {
	s.logger.Info("Getting stream by stream key", zap.String("stream_key", streamKey))

	stream, err := s.repo.GetByStreamKey(streamKey)
	if err != nil {
		s.logger.Error("Error getting stream by key",
			zap.String("stream_key", streamKey),
			zap.Error(err),
		)
		return nil, err
	}

	s.logger.Info("Successfully retrieved stream by key", zap.String("stream_key", streamKey))
	return stream, nil
}

// GetAllStreams retrieves all streams
func (s *StreamService) GetAllStreams() ([]*models.LiveStream, error) {
	s.logger.Info("Getting all streams")

	streams, err := s.repo.GetAll()
	if err != nil {
		s.logger.Error("Error getting all streams", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Successfully retrieved streams", zap.Int("count", len(streams)))
	return streams, nil
}

// UpdateStream updates a stream
func (s *StreamService) UpdateStream(stream *models.LiveStream) error {
	s.logger.Info("Updating stream",
		zap.Int("id", stream.ID),
		zap.String("title", stream.Title),
	)

	err := s.repo.Update(stream)
	if err != nil {
		s.logger.Error("Error updating stream",
			zap.Int("id", stream.ID),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("Successfully updated stream", zap.Int("id", stream.ID))
	return nil
}

// DeleteStream deletes a stream by ID
func (s *StreamService) DeleteStream(id int) error {
	s.logger.Info("Deleting stream", zap.Int("id", id))

	err := s.repo.Delete(id)
	if err != nil {
		s.logger.Error("Error deleting stream",
			zap.Int("id", id),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("Successfully deleted stream", zap.Int("id", id))
	return nil
}

// UpdateStreamStatus updates the status of a stream
func (s *StreamService) UpdateStreamStatus(id int, status string) error {
	s.logger.Info("Updating stream status",
		zap.Int("id", id),
		zap.String("status", status),
	)

	err := s.repo.UpdateStatus(id, status)
	if err != nil {
		s.logger.Error("Error updating stream status",
			zap.Int("id", id),
			zap.String("status", status),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("Successfully updated stream status",
		zap.Int("id", id),
		zap.String("status", status),
	)
	return nil
}

// GetStreamWithFullURLs returns a stream with complete URLs (replacing placeholders)
func (s *StreamService) GetStreamWithFullURLs(stream *models.LiveStream) *models.LiveStream {
	s.logger.Debug("Getting stream with full URLs", zap.Int("id", stream.ID))

	host := os.Getenv("RTMP_HOST")
	if host == "" {
		host = "localhost"
	}

	// Create a copy to avoid modifying the original
	fullStream := *stream
	fullStream.IngestURL = fmt.Sprintf("rtmp://%s/live", host)
	fullStream.PlaybackURL = fmt.Sprintf("http://%s:8080/hls/%s.m3u8", host, stream.StreamKey)

	s.logger.Debug("Generated full URLs",
		zap.String("ingest_url", fullStream.IngestURL),
		zap.String("playback_url", fullStream.PlaybackURL),
	)

	return &fullStream
}

// GetAllStreamsWithFullURLs returns all streams with complete URLs
func (s *StreamService) GetAllStreamsWithFullURLs() ([]*models.LiveStream, error) {
	s.logger.Info("Getting all streams with full URLs")

	streams, err := s.repo.GetAll()
	if err != nil {
		s.logger.Error("Error getting all streams", zap.Error(err))
		return nil, err
	}

	var fullStreams []*models.LiveStream
	for _, stream := range streams {
		fullStreams = append(fullStreams, s.GetStreamWithFullURLs(stream))
	}

	s.logger.Info(
		"Successfully generated full URLs for streams",
		zap.Int("count", len(fullStreams)),
	)
	return fullStreams, nil
}
