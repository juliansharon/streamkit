package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"

	"streamkit/internal/encoder-service/models"
	"streamkit/internal/encoder-service/repos"
)

// EncoderService handles stream encoding operations
type EncoderService struct {
	logger          *zap.Logger
	rtmpServer      string
	rtmpPort        string
	outputDir       string
	streamRepo      *repos.StreamRepo
	storageService  *StorageService
	activeProcesses map[string]*models.StreamEncoder
	mu              sync.RWMutex
}

// NewEncoderService creates a new encoder service
func NewEncoderService(
	logger *zap.Logger,
	rtmpServer, rtmpPort, outputDir string,
	streamRepo *repos.StreamRepo,
	storageService *StorageService,
) *EncoderService {
	return &EncoderService{
		logger:          logger,
		rtmpServer:      rtmpServer,
		rtmpPort:        rtmpPort,
		outputDir:       outputDir,
		streamRepo:      streamRepo,
		storageService:  storageService,
		activeProcesses: make(map[string]*models.StreamEncoder),
	}
}

// StartEncoding starts encoding for a specific stream
func (e *EncoderService) StartEncoding(streamKey string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Check if already encoding
	if _, exists := e.activeProcesses[streamKey]; exists {
		e.logger.Info("Already encoding stream", zap.String("stream_key", streamKey))
		return nil
	}

	e.logger.Info("Starting encoding for stream", zap.String("stream_key", streamKey))

	// Update database status
	if err := e.streamRepo.StartStream(streamKey); err != nil {
		e.logger.Error("Failed to update stream status in database",
			zap.String("stream_key", streamKey),
			zap.Error(err),
		)
		return err
	}

	// Create output directory
	streamOutputDir := filepath.Join(e.outputDir, streamKey)
	if err := os.MkdirAll(streamOutputDir, 0o755); err != nil {
		e.logger.Error("Failed to create output directory",
			zap.String("stream_key", streamKey),
			zap.Error(err),
		)
		return err
	}

	// RTMP input URL
	rtmpURL := fmt.Sprintf("rtmp://%s:%s/live/%s", e.rtmpServer, e.rtmpPort, streamKey)

	// HLS output files
	playlistFile := filepath.Join(streamOutputDir, "playlist.m3u8")
	segmentPattern := filepath.Join(streamOutputDir, "segment_%03d.ts")

	// Create context for this stream
	streamCtx, cancel := context.WithCancel(context.Background())

	// FFmpeg command for HLS encoding
	cmd := exec.CommandContext(streamCtx, "ffmpeg",
		"-i", rtmpURL,
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-tune", "zerolatency",
		"-c:a", "aac",
		"-b:a", "128k",
		"-f", "hls",
		"-hls_time", "3",
		"-hls_list_size", "60",
		"-hls_flags", "delete_segments",
		"-hls_segment_filename", segmentPattern,
		playlistFile,
	)

	// Set up logging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Create stream encoder
	streamEncoder := &models.StreamEncoder{
		StreamKey: streamKey,
		Cmd:       cmd,
		Ctx:       streamCtx,
		Cancel:    cancel,
		Logger:    e.logger,
	}

	// Add to active processes
	e.activeProcesses[streamKey] = streamEncoder

	e.logger.Info("Started encoding process",
		zap.String("stream_key", streamKey),
		zap.String("rtmp_url", rtmpURL),
		zap.String("output_dir", streamOutputDir),
	)

	// Start the process
	if err := cmd.Start(); err != nil {
		e.logger.Error("Failed to start FFmpeg",
			zap.String("stream_key", streamKey),
			zap.Error(err),
		)
		delete(e.activeProcesses, streamKey)
		return err
	}

	// Start file upload monitoring in separate goroutine
	go e.monitorAndUploadFiles(streamKey, streamOutputDir)

	// Monitor process completion in separate goroutine
	go func() {
		if err := cmd.Wait(); err != nil {
			e.logger.Error("FFmpeg process failed",
				zap.String("stream_key", streamKey),
				zap.Error(err),
			)
		} else {
			e.logger.Info("FFmpeg process completed",
				zap.String("stream_key", streamKey),
			)
		}

		// Remove from active processes and update database
		e.mu.Lock()
		delete(e.activeProcesses, streamKey)
		e.mu.Unlock()

		// Update database status
		if err := e.streamRepo.StopStream(streamKey); err != nil {
			e.logger.Error("Failed to update stream status in database",
				zap.String("stream_key", streamKey),
				zap.Error(err),
			)
		}

		// Clean up storage files
		if e.storageService != nil {
			if err := e.storageService.DeleteStreamFiles(streamKey); err != nil {
				e.logger.Error("Failed to delete stream files from storage",
					zap.String("stream_key", streamKey),
					zap.Error(err),
				)
			}
		}
	}()

	return nil
}

// monitorAndUploadFiles monitors HLS files and uploads them to storage
func (e *EncoderService) monitorAndUploadFiles(streamKey, outputDir string) {
	// Wait a bit for FFmpeg to create the first files
	time.Sleep(3 * time.Second)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Create a context for this monitoring goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		select {
		case <-ticker.C:
			// Check if process is still active
			e.mu.RLock()
			_, exists := e.activeProcesses[streamKey]
			e.mu.RUnlock()

			if !exists {
				e.logger.Info("Stream process stopped, ending file upload monitoring",
					zap.String("stream_key", streamKey))
				return // Process stopped
			}

			// Upload files to storage
			if e.storageService != nil {
				if err := e.storageService.UploadHLSFiles(streamKey, outputDir); err != nil {
					e.logger.Error("Failed to upload HLS files",
						zap.String("stream_key", streamKey),
						zap.Error(err),
					)
				}
			}
		case <-ctx.Done():
			e.logger.Info("File upload monitoring cancelled",
				zap.String("stream_key", streamKey))
			return
		}
	}
}

// StopEncoding stops encoding for a specific stream
func (e *EncoderService) StopEncoding(streamKey string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if streamEncoder, exists := e.activeProcesses[streamKey]; exists {
		e.logger.Info("Stopping encoding for stream", zap.String("stream_key", streamKey))
		streamEncoder.Cancel() // This will kill the FFmpeg process
		delete(e.activeProcesses, streamKey)
	} else {
		e.logger.Info("No active encoding found for stream", zap.String("stream_key", streamKey))
	}

	// Update database status
	if err := e.streamRepo.StopStream(streamKey); err != nil {
		e.logger.Error("Failed to update stream status in database",
			zap.String("stream_key", streamKey),
			zap.Error(err),
		)
	}
}

// GetActiveStreamsCount returns the number of active encoding streams
func (e *EncoderService) GetActiveStreamsCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.activeProcesses)
}

// GetActiveStreams returns a list of active stream keys from database
func (e *EncoderService) GetActiveStreams() ([]*models.Stream, error) {
	return e.streamRepo.GetActiveStreams()
}

// GetStreamStats returns stream statistics from database
func (e *EncoderService) GetStreamStats() (*models.StreamStats, error) {
	return e.streamRepo.GetStreamStats()
}
