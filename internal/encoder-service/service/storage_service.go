package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"go.uber.org/zap"

	"streamkit/internal/encoder-service/models"
)

// StorageService handles MinIO/S3 operations
type StorageService struct {
	logger     *zap.Logger
	s3Client   *s3.S3
	bucketName string
	cdnBaseURL string
}

// NewStorageService creates a new storage service
func NewStorageService(
	logger *zap.Logger,
	config *models.StorageConfig,
	cdnBaseURL string,
) (*StorageService, error) {
	// Create AWS session for MinIO
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			config.AccessKeyID,
			config.SecretAccessKey,
			"",
		),
		Endpoint:         aws.String(config.Endpoint),
		Region:           aws.String(config.Region),
		DisableSSL:       aws.Bool(!config.UseSSL),
		S3ForcePathStyle: aws.Bool(true), // Required for MinIO
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	s3Client := s3.New(sess)

	// Ensure bucket exists
	_, err = s3Client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(config.BucketName),
	})
	if err != nil {
		// Create bucket if it doesn't exist
		_, err = s3Client.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(config.BucketName),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		logger.Info("Created bucket", zap.String("bucket", config.BucketName))
	}

	return &StorageService{
		logger:     logger,
		s3Client:   s3Client,
		bucketName: config.BucketName,
		cdnBaseURL: cdnBaseURL,
	}, nil
}

// UploadFile uploads a file to MinIO/S3
func (s *StorageService) UploadFile(localPath, s3Key string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info for content type
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	contentType := s.getContentType(filepath.Ext(localPath))

	_, err = s.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(s.bucketName),
		Key:           aws.String(s3Key),
		Body:          file,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(fileInfo.Size()),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	s.logger.Info("Uploaded file to S3",
		zap.String("local_path", localPath),
		zap.String("s3_key", s3Key),
		zap.String("content_type", contentType),
	)

	return nil
}

// UploadHLSFiles uploads HLS files for a stream
func (s *StorageService) UploadHLSFiles(streamKey, localDir string) error {
	// Upload playlist file
	playlistPath := filepath.Join(localDir, "playlist.m3u8")
	playlistKey := fmt.Sprintf("hls/%s/playlist.m3u8", streamKey)

	if err := s.UploadFile(playlistPath, playlistKey); err != nil {
		return fmt.Errorf("failed to upload playlist: %w", err)
	}

	// Upload segment files
	segmentFiles, err := filepath.Glob(filepath.Join(localDir, "segment_*.ts"))
	if err != nil {
		return fmt.Errorf("failed to glob segment files: %w", err)
	}

	for _, segmentPath := range segmentFiles {
		segmentName := filepath.Base(segmentPath)
		segmentKey := fmt.Sprintf("hls/%s/%s", streamKey, segmentName)

		if err := s.UploadFile(segmentPath, segmentKey); err != nil {
			s.logger.Error("Failed to upload segment",
				zap.String("segment_path", segmentPath),
				zap.Error(err),
			)
			// Continue with other segments
		}
	}

	s.logger.Info("Uploaded HLS files for stream",
		zap.String("stream_key", streamKey),
		zap.Int("segment_count", len(segmentFiles)),
	)

	return nil
}

// GetFileContent retrieves file content directly from storage
func (s *StorageService) GetFileContent(key string) ([]byte, error) {
	result, err := s.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer result.Body.Close()

	content, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}

	return content, nil
}

// GetSignedURL generates a signed URL for file access
func (s *StorageService) GetSignedURL(key string, expires time.Duration) (string, error) {
	req, _ := s.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	url, err := req.Presign(expires)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return url, nil
}

// GetPublicURL generates a public URL for file access
func (s *StorageService) GetPublicURL(key string) string {
	if s.cdnBaseURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.cdnBaseURL, "/"), key)
	}
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucketName, key)
}

// ListStreamFiles lists all files for a stream
func (s *StorageService) ListStreamFiles(streamKey string) ([]*models.StorageFile, error) {
	prefix := fmt.Sprintf("hls/%s/", streamKey)

	result, err := s.s3Client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	var files []*models.StorageFile
	for _, obj := range result.Contents {
		files = append(files, &models.StorageFile{
			Key:          *obj.Key,
			Size:         *obj.Size,
			LastModified: *obj.LastModified,
			ETag:         strings.Trim(*obj.ETag, `"`),
			ContentType:  s.getContentType(filepath.Ext(*obj.Key)),
		})
	}

	return files, nil
}

// DeleteStreamFiles deletes all files for a stream
func (s *StorageService) DeleteStreamFiles(streamKey string) error {
	prefix := fmt.Sprintf("hls/%s/", streamKey)

	// List all objects with the prefix
	result, err := s.s3Client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return fmt.Errorf("failed to list objects for deletion: %w", err)
	}

	// Delete all objects
	for _, obj := range result.Contents {
		_, err := s.s3Client.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(s.bucketName),
			Key:    obj.Key,
		})
		if err != nil {
			s.logger.Error("Failed to delete object",
				zap.String("key", *obj.Key),
				zap.Error(err),
			)
		}
	}

	s.logger.Info("Deleted stream files",
		zap.String("stream_key", streamKey),
		zap.Int("file_count", len(result.Contents)),
	)

	return nil
}

// getContentType returns the content type based on file extension
func (s *StorageService) getContentType(ext string) string {
	switch ext {
	case ".m3u8":
		return "application/vnd.apple.mpegurl"
	case ".ts":
		return "video/mp2t"
	case ".mp4":
		return "video/mp4"
	default:
		return "application/octet-stream"
	}
}
