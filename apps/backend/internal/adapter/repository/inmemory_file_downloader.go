package repository

import (
	"context"
	"fmt"
	"sync"
)

// InMemoryFileDownloader is an in-memory implementation of FileDownloader for testing.
type InMemoryFileDownloader struct {
	mu    sync.RWMutex
	files map[string][]byte
}

// NewInMemoryFileDownloader creates a new InMemoryFileDownloader.
func NewInMemoryFileDownloader() *InMemoryFileDownloader {
	return &InMemoryFileDownloader{
		files: make(map[string][]byte),
	}
}

// DownloadFile returns the file content stored under the given key.
func (d *InMemoryFileDownloader) DownloadFile(_ context.Context, key string) ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	data, ok := d.files[key]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", key)
	}
	result := make([]byte, len(data))
	copy(result, data)
	return result, nil
}

// PutFile stores file content for testing.
func (d *InMemoryFileDownloader) PutFile(key string, data []byte) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.files[key] = data
}
