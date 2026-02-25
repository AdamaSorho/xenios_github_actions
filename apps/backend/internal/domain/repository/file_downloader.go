package repository

import "context"

// FileDownloader defines the interface for downloading file content from storage.
type FileDownloader interface {
	Download(ctx context.Context, key string) ([]byte, error)
}
