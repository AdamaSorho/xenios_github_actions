package repository

import (
	"context"
	"fmt"
	"sync"
)

// InMemoryFileContentReader is an in-memory implementation of FileContentReader for testing.
type InMemoryFileContentReader struct {
	mu    sync.RWMutex
	files map[string][]byte
}

// NewInMemoryFileContentReader creates a new InMemoryFileContentReader.
func NewInMemoryFileContentReader() *InMemoryFileContentReader {
	return &InMemoryFileContentReader{
		files: make(map[string][]byte),
	}
}

// ReadContent returns file content by storage key.
func (r *InMemoryFileContentReader) ReadContent(_ context.Context, key string) ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	content, ok := r.files[key]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", key)
	}
	cp := make([]byte, len(content))
	copy(cp, content)
	return cp, nil
}

// PutContent stores file content for testing.
func (r *InMemoryFileContentReader) PutContent(key string, content []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := make([]byte, len(content))
	copy(cp, content)
	r.files[key] = cp
}
