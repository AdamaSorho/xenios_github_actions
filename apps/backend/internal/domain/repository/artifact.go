package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// ArtifactRepository defines the interface for artifact persistence.
type ArtifactRepository interface {
	Create(ctx context.Context, artifact *entities.Artifact) (*entities.Artifact, error)
	FindByID(ctx context.Context, id string) (*entities.Artifact, error)
	UpdateStatus(ctx context.Context, id string, status entities.ArtifactStatus) (*entities.Artifact, error)
	UpdateDocumentSubtype(ctx context.Context, id string, subtype entities.DocumentSubtype) (*entities.Artifact, error)
}
