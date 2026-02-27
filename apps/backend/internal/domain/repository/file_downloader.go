package repository

import "context"

// FileDownloader defines the interface for downloading file content from object storage.
type FileDownloader interface {
	DownloadFile(ctx context.Context, key string) ([]byte, error)
}
