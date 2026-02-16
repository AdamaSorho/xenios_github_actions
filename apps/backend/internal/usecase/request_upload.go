package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// RequestUploadUseCase handles presigned URL generation for file uploads.
type RequestUploadUseCase struct {
	artifactRepo repository.ArtifactRepository
	fileStorage  repository.FileStorageRepository
	auditRepo    repository.AuditRepository
}

// NewRequestUploadUseCase creates a new RequestUploadUseCase.
func NewRequestUploadUseCase(
	artifactRepo repository.ArtifactRepository,
	fileStorage repository.FileStorageRepository,
	auditRepo repository.AuditRepository,
) *RequestUploadUseCase {
	return &RequestUploadUseCase{
		artifactRepo: artifactRepo,
		fileStorage:  fileStorage,
		auditRepo:    auditRepo,
	}
}

// RequestUploadInput holds the input for requesting an upload URL.
type RequestUploadInput struct {
	FileName    string
	FileSize    int64
	ContentType string
	ClientID    string
	CoachID     string
}

// RequestUploadOutput holds the output of a presigned upload URL request.
type RequestUploadOutput struct {
	PresignedURL string         `json:"presigned_url"`
	ArtifactID   string         `json:"artifact_id"`
	ExpiresAt    time.Time      `json:"expires_at"`
	StorageKey   string         `json:"storage_key"`
	Artifact     *entities.Artifact `json:"artifact"`
}

// Execute validates the upload request and generates a presigned URL.
func (uc *RequestUploadUseCase) Execute(ctx context.Context, input RequestUploadInput) (*RequestUploadOutput, error) {
	if input.FileName == "" {
		return nil, &ValidationError{Message: "file_name is required"}
	}
	if input.ContentType == "" {
		return nil, &ValidationError{Message: "content_type is required"}
	}
	if input.ClientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}

	// Validate file extension
	artTypeFromExt, err := entities.ValidateFileExtension(input.FileName)
	if err != nil {
		return nil, &ValidationError{Message: err.Error()}
	}

	// Validate content type
	_, err = entities.ValidateContentType(input.ContentType)
	if err != nil {
		return nil, &ValidationError{Message: err.Error()}
	}

	// Validate file size
	if err := entities.ValidateFileSize(input.FileSize, artTypeFromExt); err != nil {
		return nil, &ValidationError{Message: err.Error()}
	}

	// Create artifact record with pending status
	artifact := &entities.Artifact{
		ClientID:    input.ClientID,
		CoachID:     input.CoachID,
		FileName:    input.FileName,
		FileType:    input.ContentType,
		FileSize:    input.FileSize,
		Type:        artTypeFromExt,
		Status:      entities.ArtifactStatusPending,
		ContentType: input.ContentType,
	}

	created, err := uc.artifactRepo.Create(ctx, artifact)
	if err != nil {
		return nil, fmt.Errorf("create artifact: %w", err)
	}

	// Build storage key and update artifact
	storageKey := entities.BuildStorageKey(input.ClientID, artTypeFromExt, created.ID, input.FileName)
	created.StorageKey = storageKey

	// Generate presigned upload URL
	expiry := time.Duration(entities.PresignedURLExpiryMinutes) * time.Minute
	presigned, err := uc.fileStorage.GenerateUploadURL(ctx, storageKey, input.ContentType, expiry)
	if err != nil {
		return nil, fmt.Errorf("generate upload url: %w", err)
	}

	_ = uc.auditRepo.LogEvent(ctx, input.CoachID, "artifact.upload_requested", "artifact", created.ID, map[string]interface{}{
		"file_name":    input.FileName,
		"file_size":    input.FileSize,
		"content_type": input.ContentType,
		"client_id":    input.ClientID,
	})

	return &RequestUploadOutput{
		PresignedURL: presigned.URL,
		ArtifactID:   created.ID,
		ExpiresAt:    presigned.ExpiresAt,
		StorageKey:   storageKey,
		Artifact:     created,
	}, nil
}
