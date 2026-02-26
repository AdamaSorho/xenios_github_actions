package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// ConfirmUploadUseCase handles upload confirmation after a client uploads to S3.
type ConfirmUploadUseCase struct {
	artifactRepo repository.ArtifactRepository
	fileStorage  repository.FileStorageRepository
	auditRepo    repository.AuditRepository
	jobQueue     repository.JobQueue
}

// NewConfirmUploadUseCase creates a new ConfirmUploadUseCase.
func NewConfirmUploadUseCase(
	artifactRepo repository.ArtifactRepository,
	fileStorage repository.FileStorageRepository,
	auditRepo repository.AuditRepository,
	jobQueue repository.JobQueue,
) *ConfirmUploadUseCase {
	return &ConfirmUploadUseCase{
		artifactRepo: artifactRepo,
		fileStorage:  fileStorage,
		auditRepo:    auditRepo,
		jobQueue:     jobQueue,
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
	JobID    string             `json:"job_id,omitempty"`
}

// ExtractionJobPayload is the JSON payload for extraction jobs.
type ExtractionJobPayload struct {
	ArtifactID      string `json:"artifact_id"`
	StorageKey      string `json:"storage_key"`
	FileName        string `json:"file_name"`
	ContentType     string `json:"content_type"`
	DocumentSubtype string `json:"document_subtype"`
	ClientID        string `json:"client_id"`
	CoachID         string `json:"coach_id"`
}

// Execute verifies the file exists in storage, classifies the file, enqueues a job, and updates the artifact.
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

	// Classify the file
	subtype := entities.ClassifyDocument(artifact.DocumentSubtype, artifact.FileName, artifact.ContentType)
	updated, err = uc.artifactRepo.UpdateDocumentSubtype(ctx, input.ArtifactID, subtype)
	if err != nil {
		return nil, fmt.Errorf("update document subtype: %w", err)
	}

	// Enqueue extraction job
	jobType := entities.JobTypeForSubtype(subtype)
	payload, err := json.Marshal(ExtractionJobPayload{
		ArtifactID:      artifact.ID,
		StorageKey:      artifact.StorageKey,
		FileName:        artifact.FileName,
		ContentType:     artifact.ContentType,
		DocumentSubtype: string(subtype),
		ClientID:        artifact.ClientID,
		CoachID:         artifact.CoachID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal job payload: %w", err)
	}

	job, err := uc.jobQueue.Enqueue(ctx, jobType, payload)
	if err != nil {
		return nil, fmt.Errorf("enqueue extraction job: %w", err)
	}

	// Log audit events
	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "artifact.upload_confirmed",
		EntityType: "artifact",
		EntityID:   input.ArtifactID,
		Metadata: map[string]interface{}{
			"storage_key": artifact.StorageKey,
			"file_name":   artifact.FileName,
			"client_id":   artifact.ClientID,
		},
	})

	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "artifact.classified",
		EntityType: "artifact",
		EntityID:   input.ArtifactID,
		Metadata: map[string]interface{}{
			"document_subtype": string(subtype),
			"job_type":         string(jobType),
			"job_id":           job.ID,
		},
	})

	return &ConfirmUploadOutput{
		Artifact: updated,
		JobID:    job.ID,
	}, nil
}
