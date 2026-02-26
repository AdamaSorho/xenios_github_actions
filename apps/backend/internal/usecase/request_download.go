package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// RequestDownloadUseCase handles presigned URL generation for file downloads.
type RequestDownloadUseCase struct {
	artifactBase
}

// NewRequestDownloadUseCase creates a new RequestDownloadUseCase.
func NewRequestDownloadUseCase(
	artifactRepo repository.ArtifactRepository,
	fileStorage repository.FileStorageRepository,
	auditRepo repository.AuditRepository,
) *RequestDownloadUseCase {
	return &RequestDownloadUseCase{
		artifactBase: artifactBase{
			artifactRepo: artifactRepo,
			fileStorage:  fileStorage,
			auditRepo:    auditRepo,
		},
	}
}

// RequestDownloadInput holds the input for requesting a download URL.
type RequestDownloadInput struct {
	ArtifactID string
	CoachID    string
}

// RequestDownloadOutput holds the result of the download URL request.
type RequestDownloadOutput struct {
	PresignedURL string             `json:"presigned_url"`
	ExpiresAt    time.Time          `json:"expires_at"`
	Artifact     *entities.Artifact `json:"artifact"`
}

// Execute validates the download request and generates a presigned download URL.
func (uc *RequestDownloadUseCase) Execute(ctx context.Context, input RequestDownloadInput) (*RequestDownloadOutput, error) {
	if input.ArtifactID == "" {
		return nil, &ValidationError{Message: "artifact_id is required"}
	}
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}

	artifact, err := uc.findAndVerifyOwnership(ctx, input.ArtifactID, input.CoachID)
	if err != nil {
		return nil, err
	}

	if artifact.Status != entities.ArtifactStatusUploaded {
		return nil, &ValidationError{Message: fmt.Sprintf("artifact is not available for download (status: %s)", artifact.Status)}
	}

	expiry := time.Duration(entities.PresignedURLExpiryMinutes) * time.Minute
	presigned, err := uc.fileStorage.GenerateDownloadURL(ctx, artifact.StorageKey, expiry)
	if err != nil {
		return nil, fmt.Errorf("generate download url: %w", err)
	}

	uc.logAudit(ctx, input.CoachID, "artifact.download_requested", input.ArtifactID, map[string]interface{}{
		"file_name":   artifact.FileName,
		"storage_key": artifact.StorageKey,
		"client_id":   artifact.ClientID,
	})

	return &RequestDownloadOutput{
		PresignedURL: presigned.URL,
		ExpiresAt:    presigned.ExpiresAt,
		Artifact:     artifact,
	}, nil
}
