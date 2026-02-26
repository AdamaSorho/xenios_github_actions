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
	artifactBase
	jobQueue repository.JobQueue
}

// NewConfirmUploadUseCase creates a new ConfirmUploadUseCase.
func NewConfirmUploadUseCase(
	artifactRepo repository.ArtifactRepository,
	fileStorage repository.FileStorageRepository,
	auditRepo repository.AuditRepository,
	jobQueue repository.JobQueue,
) *ConfirmUploadUseCase {
	return &ConfirmUploadUseCase{
		artifactBase: artifactBase{
			artifactRepo: artifactRepo,
			fileStorage:  fileStorage,
			auditRepo:    auditRepo,
		},
		jobQueue: jobQueue,
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

// Execute verifies the file exists in storage, updates the artifact status,
// classifies the document subtype, and enqueues an extraction job.
func (uc *ConfirmUploadUseCase) Execute(ctx context.Context, input ConfirmUploadInput) (*ConfirmUploadOutput, error) {
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

	// Classify the document subtype
	subtype := entities.ClassifyDocumentSubtype(artifact.DocumentSubtype, artifact.FileName, artifact.ContentType)
	updated, err = uc.artifactRepo.UpdateDocumentSubtype(ctx, input.ArtifactID, subtype)
	if err != nil {
		return nil, fmt.Errorf("update document subtype: %w", err)
	}

	uc.logAudit(ctx, input.CoachID, "artifact.upload_confirmed", input.ArtifactID, map[string]interface{}{
		"storage_key": artifact.StorageKey,
		"file_name":   artifact.FileName,
		"client_id":   artifact.ClientID,
	})

	// Enqueue extraction job
	jobType := entities.DocumentSubtypeToJobType(subtype)
	payload, _ := json.Marshal(map[string]string{
		"artifact_id":      artifact.ID,
		"document_subtype": string(subtype),
		"storage_key":      artifact.StorageKey,
		"file_name":        artifact.FileName,
	})

	var jobID string
	if uc.jobQueue != nil {
		job, err := uc.jobQueue.Enqueue(ctx, jobType, payload)
		if err != nil {
			return nil, fmt.Errorf("enqueue extraction job: %w", err)
		}
		jobID = job.ID
	}

	uc.logAudit(ctx, input.CoachID, "artifact.classified", input.ArtifactID, map[string]interface{}{
		"document_subtype": string(subtype),
		"job_type":         string(jobType),
		"job_id":           jobID,
	})

	return &ConfirmUploadOutput{
		Artifact: updated,
		JobID:    jobID,
	}, nil
}
