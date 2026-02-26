package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryArtifactRepository is an in-memory implementation of ArtifactRepository.
type InMemoryArtifactRepository struct {
	mu        sync.RWMutex
	artifacts map[string]*entities.Artifact
}

// NewInMemoryArtifactRepository creates a new InMemoryArtifactRepository.
func NewInMemoryArtifactRepository() *InMemoryArtifactRepository {
	return &InMemoryArtifactRepository{
		artifacts: make(map[string]*entities.Artifact),
	}
}

// Create stores a new artifact.
func (r *InMemoryArtifactRepository) Create(_ context.Context, artifact *entities.Artifact) (*entities.Artifact, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if artifact.ID == "" {
		id, err := generateID()
		if err != nil {
			return nil, err
		}
		artifact.ID = id
	}

	now := time.Now()
	artifact.CreatedAt = now
	artifact.UpdatedAt = now

	stored := *artifact
	r.artifacts[stored.ID] = &stored
	result := stored
	return &result, nil
}

// FindByID retrieves an artifact by its ID.
func (r *InMemoryArtifactRepository) FindByID(_ context.Context, id string) (*entities.Artifact, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	art, ok := r.artifacts[id]
	if !ok {
		return nil, nil
	}
	result := *art
	return &result, nil
}

// updateArtifact is a helper that finds an artifact by ID, applies the
// given mutation, and returns a copy.
func (r *InMemoryArtifactRepository) updateArtifact(id string, mutate func(*entities.Artifact)) (*entities.Artifact, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	art, ok := r.artifacts[id]
	if !ok {
		return nil, fmt.Errorf("artifact not found: %s", id)
	}

	mutate(art)
	art.UpdatedAt = time.Now()

	result := *art
	return &result, nil
}

// UpdateStatus updates the status of an artifact.
func (r *InMemoryArtifactRepository) UpdateStatus(_ context.Context, id string, status entities.ArtifactStatus) (*entities.Artifact, error) {
	return r.updateArtifact(id, func(a *entities.Artifact) { a.Status = status })
}

// UpdateDocumentSubtype updates the document subtype of an artifact.
func (r *InMemoryArtifactRepository) UpdateDocumentSubtype(_ context.Context, id string, subtype entities.DocumentSubtype) (*entities.Artifact, error) {
	return r.updateArtifact(id, func(a *entities.Artifact) { a.DocumentSubtype = subtype })
}
