package usecase

import (
	"context"
	"fmt"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// ConfirmUploadUseCase handles upload confirmation after a client uploads to S3.
type ConfirmUploadUseCase struct {
	artifactRepo repository.ArtifactRepository
	fileStorage  repository.FileStorageRepository
	auditRepo    repository.AuditRepository
}

// NewConfirmUploadUseCase creates a new ConfirmUploadUseCase.
func NewConfirmUploadUseCase(
	artifactRepo repository.ArtifactRepository,
	fileStorage repository.FileStorageRepository,
	auditRepo repository.AuditRepository,
) *ConfirmUploadUseCase {
	return &ConfirmUploadUseCase{
		artifactRepo: artifactRepo,
		fileStorage:  fileStorage,
		auditRepo:    auditRepo,
	}
}

// ConfirmUploadInput holds the input for confirming an upload.
type ConfirmUploadInput struct {
	ArtifactID string
	CoachID    string
}

// ConfirmUploadOutput holds the result of the confirmation.
type ConfirmUploadOutput struct {
	Artifact *entities.Artifact `json:"artifact"`
}

// Execute verifies the file exists in storage and updates the artifact status.
func (uc *ConfirmUploadUseCase) Execute(ctx context.Context, input ConfirmUploadInput) (*ConfirmUploadOutput, error) {
	if input.ArtifactID == "" {
		return nil, &ValidationError{Message: "artifact_id is required"}
	}
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}

	artifact, err := uc.artifactRepo.FindByID(ctx, input.ArtifactID)
	if err != nil {
		return nil, fmt.Errorf("find artifact: %w", err)
	}
	if artifact == nil {
		return nil, &ValidationError{Message: "artifact not found"}
	}

	// Verify the requesting coach owns this artifact
	if artifact.CoachID != input.CoachID {
		return nil, &AuthenticationError{Message: "not authorized to confirm this upload"}
	}

	if artifact.Status != entities.ArtifactStatusPending {
		return nil, &ValidationError{Message: fmt.Sprintf("artifact status is %s, expected pending", artifact.Status)}
	}

	// Verify the file actually exists in storage
	exists, err := uc.fileStorage.ObjectExists(ctx, artifact.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("check object existence: %w", err)
	}
	if !exists {
		// Mark as failed since the file wasn't uploaded
		_, _ = uc.artifactRepo.UpdateStatus(ctx, input.ArtifactID, entities.ArtifactStatusFailed)
		return nil, &ValidationError{Message: "file not found in storage"}
	}

	// Update status to uploaded
	updated, err := uc.artifactRepo.UpdateStatus(ctx, input.ArtifactID, entities.ArtifactStatusUploaded)
	if err != nil {
		return nil, fmt.Errorf("update artifact status: %w", err)
	}

	_ = uc.auditRepo.LogEvent(ctx, input.CoachID, "artifact.upload_confirmed", "artifact", input.ArtifactID, map[string]interface{}{
		"storage_key": artifact.StorageKey,
		"file_name":   artifact.FileName,
		"client_id":   artifact.ClientID,
	})

	return &ConfirmUploadOutput{
		Artifact: updated,
	}, nil
}
