package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

type RTMPStats struct {
	Server struct {
		Applications []struct {
			Name  string `json:"name"`
			Lives []struct {
				Stream string `json:"stream"`
			} `json:"lives"`
		} `json:"applications"`
	} `json:"server"`
}

type EncoderService struct {
	logger        *zap.Logger
	rtmpServer    string
	rtmpPort      string
	outputDir     string
	activeStreams map[string]*exec.Cmd
}

func NewEncoderService(logger *zap.Logger, rtmpServer, rtmpPort, outputDir string) *EncoderService {
	return &EncoderService{
		logger:        logger,
		rtmpServer:    rtmpServer,
		rtmpPort:      rtmpPort,
		outputDir:     outputDir,
		activeStreams: make(map[string]*exec.Cmd),
	}
}

// Monitor streams and start encoding for new ones
func (e *EncoderService) MonitorStreams() {
	e.logger.Info("Starting stream monitoring")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.checkAndStartEncoding()
		}
	}
}

// Check RTMP stats and start encoding for new streams
func (e *EncoderService) checkAndStartEncoding() {
	statsURL := fmt.Sprintf("http://%s:%s/stat", e.rtmpServer, "8081")

	resp, err := http.Get(statsURL)
	if err != nil {
		e.logger.Error("Failed to get RTMP stats", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	var stats RTMPStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		e.logger.Error("Failed to decode RTMP stats", zap.Error(err))
		return
	}

	// Check for active streams
	for _, app := range stats.Server.Applications {
		if app.Name == "live" {
			for _, live := range app.Lives {
				streamKey := live.Stream
				if streamKey != "" {
					e.startEncodingIfNeeded(streamKey)
				}
			}
		}
	}

	// Clean up finished processes
	e.cleanupFinishedProcesses()
}

// Start encoding if not already running
func (e *EncoderService) startEncodingIfNeeded(streamKey string) {
	if _, exists := e.activeStreams[streamKey]; exists {
		return // Already encoding
	}

	e.logger.Info("Starting encoding for new stream", zap.String("stream_key", streamKey))

	// Create output directory
	streamOutputDir := filepath.Join(e.outputDir, streamKey)
	if err := os.MkdirAll(streamOutputDir, 0o755); err != nil {
		e.logger.Error("Failed to create output directory", zap.Error(err))
		return
	}

	// RTMP input URL
	rtmpURL := fmt.Sprintf("rtmp://%s:%s/live/%s", e.rtmpServer, e.rtmpPort, streamKey)

	// HLS output files
	playlistFile := filepath.Join(streamOutputDir, "playlist.m3u8")
	segmentPattern := filepath.Join(streamOutputDir, "segment_%03d.ts")

	// FFmpeg command
	cmd := exec.Command("ffmpeg",
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

	// Start the process
	if err := cmd.Start(); err != nil {
		e.logger.Error("Failed to start FFmpeg", zap.Error(err))
		return
	}

	e.activeStreams[streamKey] = cmd
	e.logger.Info("Started encoding process",
		zap.String("stream_key", streamKey),
		zap.Int("pid", cmd.Process.Pid),
	)

	// Monitor process completion
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
		delete(e.activeStreams, streamKey)
	}()
}

// Clean up finished processes
func (e *EncoderService) cleanupFinishedProcesses() {
	for streamKey, cmd := range e.activeStreams {
		if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
			e.logger.Info("Removing finished encoding process", zap.String("stream_key", streamKey))
			delete(e.activeStreams, streamKey)
		}
	}
}

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	logger.Info("Starting StreamKit Encoder Service")

	// Get configuration from environment
	rtmpServer := os.Getenv("RTMP_SERVER")
	if rtmpServer == "" {
		rtmpServer = "rtmp"
	}

	rtmpPort := os.Getenv("RTMP_PORT")
	if rtmpPort == "" {
		rtmpPort = "1935"
	}

	outputDir := os.Getenv("HLS_OUTPUT_DIR")
	if outputDir == "" {
		outputDir = "/tmp/hls"
	}

	logger.Info("Encoder service configuration",
		zap.String("rtmp_server", rtmpServer),
		zap.String("rtmp_port", rtmpPort),
		zap.String("output_dir", outputDir),
	)

	// Create encoder service
	encoder := NewEncoderService(logger, rtmpServer, rtmpPort, outputDir)

	// Start monitoring
	encoder.MonitorStreams()
}
