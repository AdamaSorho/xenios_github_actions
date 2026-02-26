package repository

import "context"

// FileContentReader provides read access to file content stored in object storage.
type FileContentReader interface {
	ReadContent(ctx context.Context, key string) ([]byte, error)
}
