package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"streamkit/internal/api/handlers"
	"streamkit/internal/api/repos"
	"streamkit/internal/api/routes"
	"streamkit/internal/api/service"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	logger.Info("Starting StreamKit API server")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logger.Info("No .env file found, using system environment variables")
	}

	// Database connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "streamkit")

	// Create database connection string
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	logger.Info("Connecting to database",
		zap.String("host", dbHost),
		zap.String("port", dbPort),
		zap.String("database", dbName),
	)

	// Connect to database
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}
	logger.Info("Successfully connected to database")

	// Initialize layers with logger
	streamRepo := repos.NewStreamRepository(db, logger)
	streamService := service.NewStreamService(streamRepo, logger)
	streamHandler := handlers.NewStreamHandler(streamService, logger)

	// Setup router
	router := mux.NewRouter()

	// Setup routes
	routes.SetupStreamRoutes(router, streamHandler)

	// Add middleware for CORS
	router.Use(corsMiddleware)

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Health check requested", zap.String("remote_addr", r.RemoteAddr))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "message": "StreamKit API is running"}`))
	}).Methods("GET")

	// Get port from environment
	port := getEnv("PORT", "8080")
	serverAddr := ":" + port

	logger.Info("Starting server",
		zap.String("port", port),
		zap.String("health_url", fmt.Sprintf("http://localhost%s/health", serverAddr)),
		zap.String("api_url", fmt.Sprintf("http://localhost%s/api/streams", serverAddr)),
	)

	// Start server
	if err := http.ListenAndServe(serverAddr, router); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
