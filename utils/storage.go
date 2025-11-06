package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

// StorageClient defines the interface for storage operations
// This allows easy swapping of storage implementations (S3, GCS, Azure Blob, etc.)
type StorageClient interface {
	// UploadFile uploads a file to storage and returns the public URL
	UploadFile(ctx context.Context, fileReader io.Reader, filename, contentType string) (string, error)

	// GetBucket returns the bucket name
	GetBucket() string

	// GetEndpoint returns the storage endpoint
	GetEndpoint() string
}

// S3StorageClient implements StorageClient interface for S3-compatible storage (Supabase Storage, AWS S3, etc.)
type S3StorageClient struct {
	client   *s3.Client
	bucket   string
	endpoint string
}

// NewS3StorageClient creates a new S3 storage client
func NewS3StorageClient(client *s3.Client, bucket, endpoint string) StorageClient {
	return &S3StorageClient{
		client:   client,
		bucket:   bucket,
		endpoint: endpoint,
	}
}

// UploadFile uploads a file to storage and returns the public URL
func (s *S3StorageClient) UploadFile(ctx context.Context, fileReader io.Reader, filename, contentType string) (string, error) {
	// Generate unique filename
	ext := filepath.Ext(filename)
	newFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	objectKey := fmt.Sprintf("images/%s", newFilename)

	// Read file content into buffer
	fileContent, err := io.ReadAll(fileReader)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Detect content type if not provided
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload to storage
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(fileContent),
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to storage: %w", err)
	}

	// Generate public URL
	return s.generatePublicURL(objectKey), nil
}

// GetBucket returns the bucket name
func (s *S3StorageClient) GetBucket() string {
	return s.bucket
}

// GetEndpoint returns the storage endpoint
func (s *S3StorageClient) GetEndpoint() string {
	return s.endpoint
}

// generatePublicURL generates the public URL for the uploaded file
func (s *S3StorageClient) generatePublicURL(objectKey string) string {
	// Supabase Storage public URL format: https://<project-ref>.supabase.co/storage/v1/object/public/<bucket>/<path>
	if strings.HasPrefix(s.endpoint, "https://") {
		// Extract project ref: https://mhheblvgktovcrdcsjdo.storage.supabase.co/storage/v1/s3
		parts := strings.Split(s.endpoint, ".")
		if len(parts) > 0 {
			projectRef := strings.TrimPrefix(parts[0], "https://")
			return fmt.Sprintf("https://%s.supabase.co/storage/v1/object/public/%s/%s", projectRef, s.bucket, objectKey)
		}
	}
	// Fallback to S3 endpoint format
	return fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, objectKey)
}

// NewStorageClient creates a new storage client based on the provided config
// This factory function returns the StorageClient interface, allowing easy swapping of implementations
func NewStorageClient(config *Config) (StorageClient, error) {
	// Validate required storage config
	if config.StorageEndpoint == "" || config.StorageRegion == "" || config.StorageAccessKey == "" {
		return nil, fmt.Errorf("storage configuration is incomplete")
	}

	// Configure AWS SDK for S3-compatible storage (Supabase Storage)
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(config.StorageRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.StorageAccessKey,
			config.StorageAccessKey, // Supabase uses access key as both access key and secret
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with custom endpoint
	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(config.StorageEndpoint)
		o.UsePathStyle = true // Required for Supabase Storage
	})

	// Use default bucket if not specified
	bucket := config.StorageBucket
	if bucket == "" {
		bucket = "images"
	}

	return NewS3StorageClient(s3Client, bucket, config.StorageEndpoint), nil
}
