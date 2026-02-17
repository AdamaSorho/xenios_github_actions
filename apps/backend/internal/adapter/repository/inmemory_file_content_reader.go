package repository

import (
	"context"
	"fmt"
	"sync"
)

// InMemoryFileContentReader is an in-memory implementation of FileContentReader for testing.
type InMemoryFileContentReader struct {
	mu      sync.RWMutex
	objects map[string][]byte
}

// NewInMemoryFileContentReader creates a new InMemoryFileContentReader.
func NewInMemoryFileContentReader() *InMemoryFileContentReader {
	return &InMemoryFileContentReader{
		objects: make(map[string][]byte),
	}
}

// GetObject returns the content of a file stored in memory.
func (r *InMemoryFileContentReader) GetObject(_ context.Context, key string) ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, ok := r.objects[key]
	if !ok {
		return nil, fmt.Errorf("object not found: %s", key)
	}
	return data, nil
}

// PutObject stores file content in memory for testing.
func (r *InMemoryFileContentReader) PutObject(key string, data []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.objects[key] = data
}
