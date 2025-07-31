package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"streamkit/internal/api/models"
	"streamkit/internal/api/service"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type StreamHandler struct {
	service *service.StreamService
	logger  *zap.Logger
}

func NewStreamHandler(service *service.StreamService, logger *zap.Logger) *StreamHandler {
	logger.Info("Initializing StreamHandler")
	return &StreamHandler{service: service, logger: logger}
}

// CreateStream handles POST /api/streams
func (h *StreamHandler) CreateStream(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Creating stream",
		zap.String("method", r.Method),
		zap.String("url", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr),
	)

	var stream models.LiveStream
	if err := json.NewDecoder(r.Body).Decode(&stream); err != nil {
		h.logger.Error("Error decoding request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("Received stream data",
		zap.String("title", stream.Title),
		zap.String("stream_name", stream.StreamName),
		zap.String("created_by", stream.StreamCreatedBy),
	)

	// Validate required fields
	if stream.Title == "" || stream.StreamName == "" || stream.StreamCreatedBy == "" {
		h.logger.Warn("Validation failed - missing required fields")
		http.Error(w, "Title, StreamName, and StreamCreatedBy are required", http.StatusBadRequest)
		return
	}

	if err := h.service.CreateStream(&stream); err != nil {
		h.logger.Error("Error creating stream", zap.Error(err))
		http.Error(w, "Failed to create stream: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the created stream with full URLs
	fullStream := h.service.GetStreamWithFullURLs(&stream)

	h.logger.Info("Successfully created stream",
		zap.Int("id", fullStream.ID),
		zap.String("stream_key", fullStream.StreamKey),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(fullStream)
}

// GetStream handles GET /api/streams/{id}
func (h *StreamHandler) GetStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.logger.Warn("Invalid stream ID", zap.String("id", vars["id"]))
		http.Error(w, "Invalid stream ID", http.StatusBadRequest)
		return
	}

	h.logger.Info("Getting stream by ID", zap.Int("id", id))

	stream, err := h.service.GetStreamByID(id)
	if err != nil {
		if err.Error() == "stream not found" {
			h.logger.Warn("Stream not found", zap.Int("id", id))
			http.Error(w, "Stream not found", http.StatusNotFound)
		} else {
			h.logger.Error("Error getting stream by ID",
				zap.Int("id", id),
				zap.Error(err),
			)
			http.Error(w, "Failed to get stream: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	fullStream := h.service.GetStreamWithFullURLs(stream)

	h.logger.Info("Successfully retrieved stream", zap.Int("id", id))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fullStream)
}

// GetStreamByKey handles GET /api/streams/key/{streamKey}
func (h *StreamHandler) GetStreamByKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamKey := vars["streamKey"]

	h.logger.Info("Getting stream by key", zap.String("stream_key", streamKey))

	stream, err := h.service.GetStreamByStreamKey(streamKey)
	if err != nil {
		if err.Error() == "stream not found" {
			h.logger.Warn("Stream not found", zap.String("stream_key", streamKey))
			http.Error(w, "Stream not found", http.StatusNotFound)
		} else {
			h.logger.Error("Error getting stream by key",
				zap.String("stream_key", streamKey),
				zap.Error(err),
			)
			http.Error(w, "Failed to get stream: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	fullStream := h.service.GetStreamWithFullURLs(stream)

	h.logger.Info("Successfully retrieved stream", zap.String("stream_key", streamKey))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fullStream)
}

// GetAllStreams handles GET /api/streams
func (h *StreamHandler) GetAllStreams(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting all streams")

	streams, err := h.service.GetAllStreamsWithFullURLs()
	if err != nil {
		h.logger.Error("Error getting all streams", zap.Error(err))
		http.Error(w, "Failed to get streams: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("Successfully retrieved streams", zap.Int("count", len(streams)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(streams)
}

// UpdateStream handles PUT /api/streams/{id}
func (h *StreamHandler) UpdateStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.logger.Warn("Invalid stream ID", zap.String("id", vars["id"]))
		http.Error(w, "Invalid stream ID", http.StatusBadRequest)
		return
	}

	h.logger.Info("Updating stream", zap.Int("id", id))

	var stream models.LiveStream
	if err := json.NewDecoder(r.Body).Decode(&stream); err != nil {
		h.logger.Error("Error decoding request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	stream.ID = id

	if err := h.service.UpdateStream(&stream); err != nil {
		if err.Error() == "stream not found" {
			h.logger.Warn("Stream not found", zap.Int("id", id))
			http.Error(w, "Stream not found", http.StatusNotFound)
		} else {
			h.logger.Error("Error updating stream",
				zap.Int("id", id),
				zap.Error(err),
			)
			http.Error(w, "Failed to update stream: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Get the updated stream with full URLs
	updatedStream, err := h.service.GetStreamByID(id)
	if err != nil {
		h.logger.Error("Error getting updated stream",
			zap.Int("id", id),
			zap.Error(err),
		)
		http.Error(w, "Failed to get updated stream", http.StatusInternalServerError)
		return
	}

	fullStream := h.service.GetStreamWithFullURLs(updatedStream)

	h.logger.Info("Successfully updated stream", zap.Int("id", id))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fullStream)
}

// DeleteStream handles DELETE /api/streams/{id}
func (h *StreamHandler) DeleteStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.logger.Warn("Invalid stream ID", zap.String("id", vars["id"]))
		http.Error(w, "Invalid stream ID", http.StatusBadRequest)
		return
	}

	h.logger.Info("Deleting stream", zap.Int("id", id))

	if err := h.service.DeleteStream(id); err != nil {
		if err.Error() == "stream not found" {
			h.logger.Warn("Stream not found", zap.Int("id", id))
			http.Error(w, "Stream not found", http.StatusNotFound)
		} else {
			h.logger.Error("Error deleting stream",
				zap.Int("id", id),
				zap.Error(err),
			)
			http.Error(w, "Failed to delete stream: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	h.logger.Info("Successfully deleted stream", zap.Int("id", id))
	w.WriteHeader(http.StatusNoContent)
}

// UpdateStreamStatus handles PATCH /api/streams/{id}/status
func (h *StreamHandler) UpdateStreamStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.logger.Warn("Invalid stream ID", zap.String("id", vars["id"]))
		http.Error(w, "Invalid stream ID", http.StatusBadRequest)
		return
	}

	h.logger.Info("Updating status for stream", zap.Int("id", id))

	var statusUpdate struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&statusUpdate); err != nil {
		h.logger.Error("Error decoding status update", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if statusUpdate.Status == "" {
		h.logger.Warn("Status is required")
		http.Error(w, "Status is required", http.StatusBadRequest)
		return
	}

	h.logger.Info("Updating status", zap.String("status", statusUpdate.Status))

	if err := h.service.UpdateStreamStatus(id, statusUpdate.Status); err != nil {
		if err.Error() == "stream not found" {
			h.logger.Warn("Stream not found", zap.Int("id", id))
			http.Error(w, "Stream not found", http.StatusNotFound)
		} else {
			h.logger.Error("Error updating stream status",
				zap.Int("id", id),
				zap.Error(err),
			)
			http.Error(w, "Failed to update stream status: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Get the updated stream
	stream, err := h.service.GetStreamByID(id)
	if err != nil {
		h.logger.Error("Error getting updated stream",
			zap.Int("id", id),
			zap.Error(err),
		)
		http.Error(w, "Failed to get updated stream", http.StatusInternalServerError)
		return
	}

	fullStream := h.service.GetStreamWithFullURLs(stream)

	h.logger.Info("Successfully updated stream status", zap.Int("id", id))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fullStream)
}

// StartStreamEncoding handles POST /api/streams/{streamKey}/encode/start
func (h *StreamHandler) StartStreamEncoding(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamKey := vars["streamKey"]

	h.logger.Info("Starting stream encoding", zap.String("stream_key", streamKey))

	if err := h.service.StartStreamEncoding(streamKey); err != nil {
		h.logger.Error("Error starting stream encoding",
			zap.String("stream_key", streamKey),
			zap.Error(err),
		)
		http.Error(
			w,
			"Failed to start stream encoding: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	h.logger.Info("Successfully started stream encoding", zap.String("stream_key", streamKey))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Stream encoding started",
		"stream_key": streamKey,
	})
}

// StopStreamEncoding handles POST /api/streams/{streamKey}/encode/stop
func (h *StreamHandler) StopStreamEncoding(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamKey := vars["streamKey"]

	h.logger.Info("Stopping stream encoding", zap.String("stream_key", streamKey))

	if err := h.service.StopStreamEncoding(streamKey); err != nil {
		h.logger.Error("Error stopping stream encoding",
			zap.String("stream_key", streamKey),
			zap.Error(err),
		)
		http.Error(
			w,
			"Failed to stop stream encoding: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	h.logger.Info("Successfully stopped stream encoding", zap.String("stream_key", streamKey))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Stream encoding stopped",
		"stream_key": streamKey,
	})
}

// GetStreamEncodingStatus handles GET /api/streams/{streamKey}/encode/status
func (h *StreamHandler) GetStreamEncodingStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamKey := vars["streamKey"]

	h.logger.Info("Getting stream encoding status", zap.String("stream_key", streamKey))

	isEncoding, err := h.service.GetStreamEncodingStatus(streamKey)
	if err != nil {
		h.logger.Error("Error getting stream encoding status",
			zap.String("stream_key", streamKey),
			zap.Error(err),
		)
		http.Error(
			w,
			"Failed to get stream encoding status: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	h.logger.Info("Successfully retrieved stream encoding status",
		zap.String("stream_key", streamKey),
		zap.Bool("is_encoding", isEncoding),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stream_key":  streamKey,
		"is_encoding": isEncoding,
	})
}
