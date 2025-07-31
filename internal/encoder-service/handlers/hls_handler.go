package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"streamkit/internal/encoder-service/models"
	"streamkit/internal/encoder-service/service"
)

// HLSHandler handles HLS file serving from S3
type HLSHandler struct {
	logger         *zap.Logger
	storageService *service.StorageService
}

// NewHLSHandler creates a new HLS handler
func NewHLSHandler(logger *zap.Logger, storageService *service.StorageService) *HLSHandler {
	return &HLSHandler{
		logger:         logger,
		storageService: storageService,
	}
}

// ServeHLSPlaylist serves the HLS playlist for a stream
func (h *HLSHandler) ServeHLSPlaylist(w http.ResponseWriter, r *http.Request) {
	// Extract stream key from URL path
	// Expected format: /hls/{stream_key}/playlist.m3u8
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	streamKey := pathParts[2]
	fileName := pathParts[len(pathParts)-1]

	h.logger.Info("Serving HLS file",
		zap.String("stream_key", streamKey),
		zap.String("file_name", fileName),
		zap.String("user_agent", r.UserAgent()),
	)

	// Set CORS headers
	h.setCORSHeaders(w)

	// Generate signed URL for the file
	s3Key := fmt.Sprintf("hls/%s/%s", streamKey, fileName)
	signedURL, err := h.storageService.GetSignedURL(s3Key, 1*time.Hour)
	if err != nil {
		h.logger.Error("Failed to generate signed URL",
			zap.String("stream_key", streamKey),
			zap.String("file_name", fileName),
			zap.Error(err),
		)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Redirect to signed URL
	http.Redirect(w, r, signedURL, http.StatusTemporaryRedirect)
}

// ServeHLSSegment serves an HLS segment file
func (h *HLSHandler) ServeHLSSegment(w http.ResponseWriter, r *http.Request) {
	// Extract stream key and segment name from URL path
	// Expected format: /hls/{stream_key}/segment_001.ts
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	streamKey := pathParts[2]
	segmentName := pathParts[len(pathParts)-1]

	h.logger.Info("Serving HLS segment",
		zap.String("stream_key", streamKey),
		zap.String("segment_name", segmentName),
	)

	// Set CORS headers
	h.setCORSHeaders(w)

	// Generate signed URL for the segment
	s3Key := fmt.Sprintf("hls/%s/%s", streamKey, segmentName)
	signedURL, err := h.storageService.GetSignedURL(s3Key, 1*time.Hour)
	if err != nil {
		h.logger.Error("Failed to generate signed URL for segment",
			zap.String("stream_key", streamKey),
			zap.String("segment_name", segmentName),
			zap.Error(err),
		)
		http.Error(w, "Segment not found", http.StatusNotFound)
		return
	}

	// Redirect to signed URL
	http.Redirect(w, r, signedURL, http.StatusTemporaryRedirect)
}

// GetStreamManifest returns stream manifest information
func (h *HLSHandler) GetStreamManifest(w http.ResponseWriter, r *http.Request) {
	// Extract stream key from query parameter
	streamKey := r.URL.Query().Get("stream_key")
	if streamKey == "" {
		http.Error(w, "Missing stream_key parameter", http.StatusBadRequest)
		return
	}

	h.logger.Info("Getting stream manifest", zap.String("stream_key", streamKey))

	// Set CORS headers
	h.setCORSHeaders(w)

	// List files for the stream
	files, err := h.storageService.ListStreamFiles(streamKey)
	if err != nil {
		h.logger.Error("Failed to list stream files",
			zap.String("stream_key", streamKey),
			zap.Error(err),
		)
		http.Error(w, "Failed to get stream files", http.StatusInternalServerError)
		return
	}

	// Build manifest
	manifest := &models.HLSManifest{
		StreamKey:   streamKey,
		PlaylistURL: h.storageService.GetPublicURL(fmt.Sprintf("hls/%s/playlist.m3u8", streamKey)),
		Segments:    make([]models.HLSSegment, 0),
	}

	// Add segments
	for _, file := range files {
		if strings.HasSuffix(file.Key, ".ts") {
			segment := models.HLSSegment{
				URL:  h.storageService.GetPublicURL(file.Key),
				Size: file.Size,
			}
			manifest.Segments = append(manifest.Segments, segment)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(manifest)
}

// setCORSHeaders sets CORS headers for HLS serving
func (h *HLSHandler) setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Range, Accept-Ranges, Content-Range")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Range")
	w.Header().Set("Cache-Control", "no-cache")
}
