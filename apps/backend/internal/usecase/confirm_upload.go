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

// Execute verifies the file exists in storage, classifies it, and enqueues an extraction job.
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

	// Classify the file type
	subtype := entities.ClassifyDocumentSubtype(
		artifact.DocumentSubtype,
		artifact.FileName,
		artifact.ContentType,
	)

	// Persist the classified subtype
	updated, err = uc.artifactRepo.UpdateDocumentSubtype(ctx, input.ArtifactID, subtype)
	if err != nil {
		return nil, fmt.Errorf("update document subtype: %w", err)
	}

	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "artifact.classified",
		EntityType: "artifact",
		EntityID:   input.ArtifactID,
		Metadata: map[string]interface{}{
			"document_subtype": string(subtype),
			"file_name":        artifact.FileName,
			"content_type":     artifact.ContentType,
		},
	})

	// Enqueue the appropriate extraction job
	jobType := entities.JobTypeForSubtype(subtype)
	payload, _ := json.Marshal(map[string]interface{}{
		"artifact_id": input.ArtifactID,
		"client_id":   artifact.ClientID,
		"coach_id":    artifact.CoachID,
		"file_name":   artifact.FileName,
		"storage_key": artifact.StorageKey,
		"subtype":     string(subtype),
	})

	var jobID string
	if uc.jobQueue != nil {
		job, err := uc.jobQueue.Enqueue(ctx, jobType, payload)
		if err != nil {
			return nil, fmt.Errorf("enqueue extraction job: %w", err)
		}
		jobID = job.ID

		_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
			ActorID:    input.CoachID,
			Action:     "artifact.job_enqueued",
			EntityType: "job",
			EntityID:   jobID,
			Metadata: map[string]interface{}{
				"artifact_id":      input.ArtifactID,
				"job_type":         string(jobType),
				"document_subtype": string(subtype),
			},
		})
	}

	return &ConfirmUploadOutput{
		Artifact: updated,
		JobID:    jobID,
	}, nil
}
