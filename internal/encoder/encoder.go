package encoder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"go.uber.org/zap"
)

type HLSEncoder struct {
	logger     *zap.Logger
	outputDir  string
	rtmpServer string
	rtmpPort   string
}

func NewHLSEncoder(logger *zap.Logger, outputDir, rtmpServer, rtmpPort string) *HLSEncoder {
	return &HLSEncoder{
		logger:     logger,
		outputDir:  outputDir,
		rtmpServer: rtmpServer,
		rtmpPort:   rtmpPort,
	}
}

// EncodeStream starts HLS encoding for a given stream key
func (e *HLSEncoder) EncodeStream(streamKey string) error {
	e.logger.Info("Starting HLS encoding",
		zap.String("stream_key", streamKey),
		zap.String("rtmp_server", e.rtmpServer),
		zap.String("output_dir", e.outputDir),
	)

	// Create output directory for this stream
	streamOutputDir := filepath.Join(e.outputDir, streamKey)
	if err := os.MkdirAll(streamOutputDir, 0o755); err != nil {
		e.logger.Error("Failed to create output directory",
			zap.String("stream_key", streamKey),
			zap.String("output_dir", streamOutputDir),
			zap.Error(err),
		)
		return err
	}

	// RTMP input URL
	rtmpURL := fmt.Sprintf("rtmp://%s:%s/live/%s", e.rtmpServer, e.rtmpPort, streamKey)

	// HLS output files
	playlistFile := filepath.Join(streamOutputDir, "playlist.m3u8")
	segmentPattern := filepath.Join(streamOutputDir, "segment_%03d.ts")

	e.logger.Info("Starting FFmpeg encoding",
		zap.String("input_url", rtmpURL),
		zap.String("playlist_file", playlistFile),
		zap.String("segment_pattern", segmentPattern),
	)

	// FFmpeg command for HLS encoding
	cmd := exec.Command("ffmpeg",
		"-i", rtmpURL, // Input RTMP stream
		"-c:v", "libx264", // Video codec
		"-preset", "ultrafast", // Encoding preset
		"-tune", "zerolatency", // Low latency tuning
		"-c:a", "aac", // Audio codec
		"-b:a", "128k", // Audio bitrate
		"-f", "hls", // HLS format
		"-hls_time", "3", // Segment duration
		"-hls_list_size", "60", // Number of segments in playlist
		"-hls_flags", "delete_segments", // Delete old segments
		"-hls_segment_filename", segmentPattern,
		playlistFile,
	)

	// Set up logging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	e.logger.Info("Executing FFmpeg command", zap.String("command", cmd.String()))

	// Start the encoding process
	if err := cmd.Start(); err != nil {
		e.logger.Error("Failed to start FFmpeg",
			zap.String("stream_key", streamKey),
			zap.Error(err),
		)
		return err
	}

	e.logger.Info("FFmpeg encoding started",
		zap.String("stream_key", streamKey),
		zap.Int("pid", cmd.Process.Pid),
	)

	// Wait for the process to complete (in a goroutine)
	go func() {
		if err := cmd.Wait(); err != nil {
			e.logger.Error("FFmpeg encoding failed",
				zap.String("stream_key", streamKey),
				zap.Error(err),
			)
		} else {
			e.logger.Info("FFmpeg encoding completed",
				zap.String("stream_key", streamKey),
			)
		}
	}()

	return nil
}

// StopEncoding stops encoding for a given stream key
func (e *HLSEncoder) StopEncoding(streamKey string) error {
	e.logger.Info("Stopping HLS encoding", zap.String("stream_key", streamKey))

	// Find and kill FFmpeg processes for this stream
	cmd := exec.Command("pkill", "-f", fmt.Sprintf("ffmpeg.*%s", streamKey))
	if err := cmd.Run(); err != nil {
		e.logger.Warn("No FFmpeg process found to stop",
			zap.String("stream_key", streamKey),
			zap.Error(err),
		)
	}

	e.logger.Info("HLS encoding stopped", zap.String("stream_key", streamKey))
	return nil
}

// GetStreamStatus checks if a stream is currently being encoded
func (e *HLSEncoder) GetStreamStatus(streamKey string) (bool, error) {
	cmd := exec.Command("pgrep", "-f", fmt.Sprintf("ffmpeg.*%s", streamKey))
	if err := cmd.Run(); err != nil {
		return false, nil // No process found
	}
	return true, nil
}

// ListActiveStreams returns a list of currently encoding streams
func (e *HLSEncoder) ListActiveStreams() ([]string, error) {
	cmd := exec.Command("pgrep", "-f", "ffmpeg.*live")
	output, err := cmd.Output()
	if err != nil {
		return []string{}, nil // No active streams
	}

	// Parse the output to extract stream keys
	// This is a simplified implementation
	// In production, you might want to maintain a registry of active streams
	e.logger.Info("Found active FFmpeg processes", zap.String("output", string(output)))

	// For now, return empty list - you can implement proper parsing
	return []string{}, nil
}
