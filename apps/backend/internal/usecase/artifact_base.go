package usecase

import (
	"context"
	"fmt"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// artifactBase holds shared dependencies for artifact-related use cases.
type artifactBase struct {
	artifactRepo repository.ArtifactRepository
	fileStorage  repository.FileStorageRepository
	auditRepo    repository.AuditRepository
}

// logAudit fires an audit event, ignoring any error from the audit repository.
func (b *artifactBase) logAudit(ctx context.Context, actorID, action, entityID string, metadata map[string]interface{}) {
	_ = b.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    actorID,
		Action:     action,
		EntityType: "artifact",
		EntityID:   entityID,
		Metadata:   metadata,
	})
}

// findAndVerifyOwnership looks up an artifact by ID and checks that the
// requesting coach owns it. Returns the artifact or an appropriate error.
func (b *artifactBase) findAndVerifyOwnership(ctx context.Context, artifactID, coachID string) (*entities.Artifact, error) {
	artifact, err := b.artifactRepo.FindByID(ctx, artifactID)
	if err != nil {
		return nil, fmt.Errorf("find artifact: %w", err)
	}
	if artifact == nil {
		return nil, &ValidationError{Message: "artifact not found"}
	}
	if artifact.CoachID != coachID {
		return nil, &AuthenticationError{Message: "not authorized to access this artifact"}
	}
	return artifact, nil
}
