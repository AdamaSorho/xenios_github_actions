package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	domainrepo "github.com/xenios/backend/internal/domain/repository"
)

// InMemoryFileStorage is an in-memory implementation of FileStorageRepository for testing.
type InMemoryFileStorage struct {
	mu       sync.RWMutex
	objects  map[string]bool
	contents map[string][]byte
	baseURL  string
}

// NewInMemoryFileStorage creates a new InMemoryFileStorage.
func NewInMemoryFileStorage() *InMemoryFileStorage {
	return &InMemoryFileStorage{
		objects:  make(map[string]bool),
		contents: make(map[string][]byte),
		baseURL:  "https://test-bucket.s3.amazonaws.com",
	}
}

// GenerateUploadURL generates a mock presigned upload URL.
func (s *InMemoryFileStorage) GenerateUploadURL(_ context.Context, key string, contentType string, expiry time.Duration) (*domainrepo.PresignedURL, error) {
	return &domainrepo.PresignedURL{
		URL:       fmt.Sprintf("%s/%s?upload=true&contentType=%s", s.baseURL, key, contentType),
		ExpiresAt: time.Now().Add(expiry),
	}, nil
}

// GenerateDownloadURL generates a mock presigned download URL.
func (s *InMemoryFileStorage) GenerateDownloadURL(_ context.Context, key string, expiry time.Duration) (*domainrepo.PresignedURL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.objects[key] {
		return nil, fmt.Errorf("object not found: %s", key)
	}

	return &domainrepo.PresignedURL{
		URL:       fmt.Sprintf("%s/%s?download=true", s.baseURL, key),
		ExpiresAt: time.Now().Add(expiry),
	}, nil
}

// ObjectExists checks if an object exists in the mock storage.
func (s *InMemoryFileStorage) ObjectExists(_ context.Context, key string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.objects[key], nil
}

// GetObjectContent returns the content stored at the given key.
func (s *InMemoryFileStorage) GetObjectContent(_ context.Context, key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	content, ok := s.contents[key]
	if !ok {
		return nil, fmt.Errorf("object not found: %s", key)
	}
	result := make([]byte, len(content))
	copy(result, content)
	return result, nil
}

// PutObject simulates adding an object to storage (for testing purposes).
func (s *InMemoryFileStorage) PutObject(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[key] = true
}

// PutObjectWithContent simulates adding an object with content to storage (for testing purposes).
func (s *InMemoryFileStorage) PutObjectWithContent(key string, content []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[key] = true
	stored := make([]byte, len(content))
	copy(stored, content)
	s.contents[key] = stored
}
