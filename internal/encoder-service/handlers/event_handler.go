package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"go.uber.org/zap"

	"streamkit/internal/encoder-service/models"
	"streamkit/internal/encoder-service/service"
)

// EventHandler handles HTTP requests for stream events
type EventHandler struct {
	logger         *zap.Logger
	encoderService *service.EncoderService
}

// NewEventHandler creates a new event handler
func NewEventHandler(logger *zap.Logger, encoderService *service.EncoderService) *EventHandler {
	return &EventHandler{
		logger:         logger,
		encoderService: encoderService,
	}
}

// HandlePublishedEvent handles incoming published events
func (h *EventHandler) HandlePublishedEvent(w http.ResponseWriter, r *http.Request) {
	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			h.logger.Error("Panic recovered in HandlePublishedEvent",
				zap.Any("panic", r),
				zap.String("stack", string(debug.Stack())),
			)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}()

	// Accept both GET and POST requests
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event models.PublishedEvent

	// Handle different request types
	if r.Method == http.MethodPost {
		h.logger.Info("Received POST request")

		// Check if it's form-encoded (nginx-rtmp format)
		contentType := r.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/x-www-form-urlencoded") {
			// Parse form-encoded data (nginx-rtmp format)
			if err := r.ParseForm(); err != nil {
				h.logger.Error("Failed to parse form", zap.Error(err))
				http.Error(w, "Failed to parse form", http.StatusBadRequest)
				return
			}

			streamKey := r.Form.Get("name")
			if streamKey == "" {
				h.logger.Error("Missing stream key in form data")
				http.Error(w, "Missing stream key", http.StatusBadRequest)
				return
			}

			// Check if this is a publish_done event
			callType := r.Form.Get("call")
			action := "publish"
			if callType == "publish_done" {
				action = "unpublish"
			}

			event = models.PublishedEvent{
				StreamKey: streamKey,
				Action:    action,
				Timestamp: time.Now(),
			}

			h.logger.Info("Parsed form data",
				zap.String("stream_key", streamKey),
				zap.String("call_type", callType),
				zap.String("action", action),
				zap.Any("all_form_data", r.Form),
			)
		} else {
			// Try JSON format
			body, err := io.ReadAll(r.Body)
			if err != nil {
				h.logger.Error("Failed to read request body", zap.Error(err))
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			bodyStr := strings.TrimSpace(string(body))
			h.logger.Info("Received POST request",
				zap.String("body", bodyStr),
				zap.Int("body_length", len(bodyStr)),
			)

			// Try to parse as JSON
			if strings.HasPrefix(bodyStr, "{") {
				// JSON format
				if err := json.Unmarshal(body, &event); err != nil {
					h.logger.Error("Failed to decode JSON event", zap.Error(err))
					http.Error(w, "Invalid JSON", http.StatusBadRequest)
					return
				}
			} else {
				// Simple string format
				event = models.PublishedEvent{
					StreamKey: bodyStr,
					Action:    "publish",
					Timestamp: time.Now(),
				}
			}
		}
	} else {
		// Handle GET request from nginx-rtmp
		// nginx-rtmp sends GET requests with query parameters
		streamKey := r.URL.Query().Get("name") // nginx-rtmp sends stream key as 'name' parameter
		if streamKey == "" {
			h.logger.Error("Missing stream key in GET request")
			http.Error(w, "Missing stream key", http.StatusBadRequest)
			return
		}

		event = models.PublishedEvent{
			StreamKey: streamKey,
			Action:    "publish", // Default to publish for GET requests
			Timestamp: time.Now(),
		}

		// Log all query parameters for debugging
		h.logger.Info("Received GET request with parameters",
			zap.String("stream_key", streamKey),
			zap.Any("all_params", r.URL.Query()),
		)
	}

	h.logger.Info("Received published event",
		zap.String("stream_key", event.StreamKey),
		zap.String("action", event.Action),
		zap.Time("timestamp", event.Timestamp),
		zap.String("method", r.Method),
	)

	// Handle different actions
	switch event.Action {
	case "publish":
		if err := h.encoderService.StartEncoding(event.StreamKey); err != nil {
			h.logger.Error("Failed to start encoding",
				zap.String("stream_key", event.StreamKey),
				zap.Error(err),
			)
			http.Error(w, "Failed to start encoding", http.StatusInternalServerError)
			return
		}
	case "unpublish":
		h.encoderService.StopEncoding(event.StreamKey)
	default:
		h.logger.Warn("Unknown action", zap.String("action", event.Action))
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "processed"}`))
}
