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

// Download returns the file content for the given key.
func (d *InMemoryFileDownloader) Download(_ context.Context, key string) ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	content, ok := d.files[key]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", key)
	}
	result := make([]byte, len(content))
	copy(result, content)
	return result, nil
}

// PutFile stores file content for testing purposes.
func (d *InMemoryFileDownloader) PutFile(key string, content []byte) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.files[key] = content
}
