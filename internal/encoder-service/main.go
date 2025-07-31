package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"streamkit/internal/encoder-service/handlers"
	"streamkit/internal/encoder-service/models"
	"streamkit/internal/encoder-service/repos"
	"streamkit/internal/encoder-service/service"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	logger.Info("Starting StreamKit Encoder Service")

	// Get configuration from environment
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8082"
	}

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

	// Database configuration
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "password"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "streamkit"
	}

	// MinIO/S3 configuration
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "localhost:9000"
	}

	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	if minioAccessKey == "" {
		minioAccessKey = "minioadmin"
	}

	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	if minioSecretKey == "" {
		minioSecretKey = "minioadmin"
	}

	minioBucket := os.Getenv("MINIO_BUCKET")
	if minioBucket == "" {
		minioBucket = "hls-streams"
	}

	minioRegion := os.Getenv("MINIO_REGION")
	if minioRegion == "" {
		minioRegion = "us-east-1"
	}

	minioUseSSL := os.Getenv("MINIO_USE_SSL") == "true"

	cdnBaseURL := os.Getenv("CDN_BASE_URL")

	// Connect to database
	dbURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}

	logger.Info("Connected to database successfully")

	// Create stream repository
	streamRepo := repos.NewStreamRepo(db, logger)

	// Create storage service
	storageConfig := &models.StorageConfig{
		Endpoint:        minioEndpoint,
		AccessKeyID:     minioAccessKey,
		SecretAccessKey: minioSecretKey,
		UseSSL:          minioUseSSL,
		BucketName:      minioBucket,
		Region:          minioRegion,
	}

	storageService, err := service.NewStorageService(logger, storageConfig, cdnBaseURL)
	if err != nil {
		logger.Fatal("Failed to create storage service", zap.Error(err))
	}

	logger.Info("Connected to MinIO/S3 successfully")

	logger.Info("Encoder service configuration",
		zap.String("port", port),
		zap.String("rtmp_server", rtmpServer),
		zap.String("rtmp_port", rtmpPort),
		zap.String("output_dir", outputDir),
		zap.String("db_host", dbHost),
		zap.String("db_name", dbName),
		zap.String("minio_endpoint", minioEndpoint),
		zap.String("minio_bucket", minioBucket),
		zap.String("cdn_base_url", cdnBaseURL),
	)

	// Create encoder service
	encoderService := service.NewEncoderService(
		logger,
		rtmpServer,
		rtmpPort,
		outputDir,
		streamRepo,
		storageService,
	)

	// Create handlers
	eventHandler := handlers.NewEventHandler(logger, encoderService)
	hlsHandler := handlers.NewHLSHandler(logger, storageService)

	// Setup routes
	http.HandleFunc("/events/published", eventHandler.HandlePublishedEvent)

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy"}`))
	})

	// Stats endpoint
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		stats, err := encoderService.GetStreamStats()
		if err != nil {
			logger.Error("Failed to get stream stats", zap.Error(err))
			http.Error(w, "Failed to get stats", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	// Active streams endpoint
	http.HandleFunc("/streams/active", func(w http.ResponseWriter, r *http.Request) {
		streams, err := encoderService.GetActiveStreams()
		if err != nil {
			logger.Error("Failed to get active streams", zap.Error(err))
			http.Error(w, "Failed to get active streams", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(streams)
	})

	// HLS serving endpoints
	http.HandleFunc("/hls/", func(w http.ResponseWriter, r *http.Request) {
		// Route to appropriate handler based on file type
		if strings.HasSuffix(r.URL.Path, ".m3u8") {
			hlsHandler.ServeHLSPlaylist(w, r)
		} else if strings.HasSuffix(r.URL.Path, ".ts") {
			hlsHandler.ServeHLSSegment(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	// Stream manifest endpoint
	http.HandleFunc("/manifest", hlsHandler.GetStreamManifest)

	// Start server
	logger.Info("Starting encoder service server", zap.String("port", port))
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
