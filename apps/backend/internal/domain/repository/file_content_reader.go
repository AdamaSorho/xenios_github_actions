package repository

import "context"

// FileContentReader defines the interface for reading file content from storage.
// This is separate from FileStorageRepository to follow interface segregation.
type FileContentReader interface {
	GetObject(ctx context.Context, key string) ([]byte, error)
}
