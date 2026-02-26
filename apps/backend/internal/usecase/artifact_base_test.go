package usecase

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func newArtifactBase() (artifactBase, *repository.InMemoryArtifactRepository, *repository.InMemoryAuditRepository) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	auditRepo := repository.NewInMemoryAuditRepository()
	base := artifactBase{
		artifactRepo: artifactRepo,
		fileStorage:  fileStorage,
		auditRepo:    auditRepo,
	}
	return base, artifactRepo, auditRepo
}

func TestArtifactBase_LogAudit_LogsEvent(t *testing.T) {
	base, _, auditRepo := newArtifactBase()

	base.logAudit(context.Background(), "actor-1", "test.action", "entity-1", map[string]interface{}{
		"key": "value",
	})

	events := auditRepo.GetEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Action != "test.action" {
		t.Errorf("expected action 'test.action', got %q", events[0].Action)
	}
	if events[0].EntityType != "artifact" {
		t.Errorf("expected entity_type 'artifact', got %q", events[0].EntityType)
	}
}

func TestArtifactBase_FindAndVerifyOwnership_Success(t *testing.T) {
	base, artifactRepo, _ := newArtifactBase()

	art, _ := artifactRepo.Create(context.Background(), &entities.Artifact{
		CoachID: "coach-1",
		Status:  entities.ArtifactStatusPending,
	})

	found, err := base.findAndVerifyOwnership(context.Background(), art.ID, "coach-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.ID != art.ID {
		t.Errorf("expected artifact ID %q, got %q", art.ID, found.ID)
	}
}

func TestArtifactBase_FindAndVerifyOwnership_NotFound(t *testing.T) {
	base, _, _ := newArtifactBase()

	_, err := base.findAndVerifyOwnership(context.Background(), "nonexistent", "coach-1")
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestArtifactBase_FindAndVerifyOwnership_WrongOwner(t *testing.T) {
	base, artifactRepo, _ := newArtifactBase()

	art, _ := artifactRepo.Create(context.Background(), &entities.Artifact{
		CoachID: "coach-1",
		Status:  entities.ArtifactStatusPending,
	})

	_, err := base.findAndVerifyOwnership(context.Background(), art.ID, "different-coach")
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsAuthenticationError(err) {
		t.Errorf("expected AuthenticationError, got %T", err)
	}
}
