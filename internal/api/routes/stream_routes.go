package routes

import (
	"streamkit/internal/api/handlers"

	"github.com/gorilla/mux"
)

// SetupStreamRoutes configures all stream-related routes
func SetupStreamRoutes(router *mux.Router, handler *handlers.StreamHandler) {
	// Stream CRUD operations
	router.HandleFunc("/api/streams", handler.CreateStream).Methods("POST")
	router.HandleFunc("/api/streams", handler.GetAllStreams).Methods("GET")
	router.HandleFunc("/api/streams/{id:[0-9]+}", handler.GetStream).Methods("GET")
	router.HandleFunc("/api/streams/key/{streamKey}", handler.GetStreamByKey).Methods("GET")
	router.HandleFunc("/api/streams/{id:[0-9]+}", handler.UpdateStream).Methods("PUT")
	router.HandleFunc("/api/streams/{id:[0-9]+}", handler.DeleteStream).Methods("DELETE")
	router.HandleFunc("/api/streams/{id:[0-9]+}/status", handler.UpdateStreamStatus).
		Methods("PATCH")
}
