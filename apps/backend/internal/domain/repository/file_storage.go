package repository

import (
	"context"
	"time"
)

// PresignedURL holds a presigned URL and its expiry time.
type PresignedURL struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// FileStorageRepository defines the interface for object storage operations.
type FileStorageRepository interface {
	GenerateUploadURL(ctx context.Context, key string, contentType string, expiry time.Duration) (*PresignedURL, error)
	GenerateDownloadURL(ctx context.Context, key string, expiry time.Duration) (*PresignedURL, error)
	ObjectExists(ctx context.Context, key string) (bool, error)
	// DownloadFile retrieves the raw content of a stored file.
	DownloadFile(ctx context.Context, key string) ([]byte, error)
}
