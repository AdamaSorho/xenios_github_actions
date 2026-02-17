package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// ExtractInBodyUseCase handles extraction of InBody PDF data into measurements.
type ExtractInBodyUseCase struct {
	artifactRepo    repository.ArtifactRepository
	fileStorage     repository.FileStorageRepository
	measurementRepo repository.MeasurementRepository
	auditRepo       repository.AuditRepository
	pdfExtractor    repository.PDFExtractor
}

// NewExtractInBodyUseCase creates a new ExtractInBodyUseCase.
func NewExtractInBodyUseCase(
	artifactRepo repository.ArtifactRepository,
	fileStorage repository.FileStorageRepository,
	measurementRepo repository.MeasurementRepository,
	auditRepo repository.AuditRepository,
	pdfExtractor repository.PDFExtractor,
) *ExtractInBodyUseCase {
	return &ExtractInBodyUseCase{
		artifactRepo:    artifactRepo,
		fileStorage:     fileStorage,
		measurementRepo: measurementRepo,
		auditRepo:       auditRepo,
		pdfExtractor:    pdfExtractor,
	}
}

// ExtractInBodyInput holds the input for an InBody extraction.
type ExtractInBodyInput struct {
	ArtifactID string
	CoachID    string
	MeasuredAt *time.Time
}

// ExtractInBodyOutput holds the result of an InBody extraction.
type ExtractInBodyOutput struct {
	ArtifactID   string                  `json:"artifact_id"`
	Measurements []*entities.Measurement `json:"measurements"`
	Partial      bool                    `json:"partial"`
	Errors       []string                `json:"errors,omitempty"`
}

// ExtractInBodyJobPayload is the JSON payload for extract_inbody jobs.
type ExtractInBodyJobPayload struct {
	ArtifactID string     `json:"artifact_id"`
	CoachID    string     `json:"coach_id"`
	MeasuredAt *time.Time `json:"measured_at,omitempty"`
}

// Execute runs the InBody PDF extraction pipeline.
func (uc *ExtractInBodyUseCase) Execute(ctx context.Context, input ExtractInBodyInput) (*ExtractInBodyOutput, error) {
	if input.ArtifactID == "" {
		return nil, &ValidationError{Message: "artifact_id is required"}
	}
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}

	// Fetch artifact metadata
	artifact, err := uc.artifactRepo.FindByID(ctx, input.ArtifactID)
	if err != nil {
		return nil, fmt.Errorf("find artifact: %w", err)
	}
	if artifact == nil {
		return nil, &ValidationError{Message: "artifact not found"}
	}

	// Verify ownership
	if artifact.CoachID != input.CoachID {
		return nil, &AuthenticationError{Message: "not authorized to extract this artifact"}
	}

	// Verify artifact is uploaded
	if artifact.Status != entities.ArtifactStatusUploaded {
		return nil, &ValidationError{Message: fmt.Sprintf("artifact status is %s, expected uploaded", artifact.Status)}
	}

	// Download the file
	pdfData, err := uc.fileStorage.DownloadFile(ctx, artifact.StorageKey)
	if err != nil {
		uc.logExtractionFailed(ctx, input, artifact, fmt.Sprintf("download failed: %v", err))
		return nil, fmt.Errorf("download file: %w", err)
	}

	// Extract InBody data
	extractResult, err := uc.pdfExtractor.ExtractInBody(ctx, pdfData)
	if err != nil {
		uc.logExtractionFailed(ctx, input, artifact, fmt.Sprintf("extraction failed: %v", err))
		return nil, fmt.Errorf("extract inbody data: %w", err)
	}

	// Set metadata on extracted measurements
	measuredAt := time.Now()
	if input.MeasuredAt != nil {
		measuredAt = *input.MeasuredAt
	}

	for _, m := range extractResult.Measurements {
		m.ClientID = artifact.ClientID
		m.RecordedBy = input.CoachID
		m.ArtifactID = input.ArtifactID
		m.MeasuredAt = measuredAt
	}

	// Store measurements
	stored, err := uc.measurementRepo.CreateBatch(ctx, extractResult.Measurements)
	if err != nil {
		uc.logExtractionFailed(ctx, input, artifact, fmt.Sprintf("storage failed: %v", err))
		return nil, fmt.Errorf("store measurements: %w", err)
	}

	// Log audit event
	if extractResult.Partial {
		uc.logExtractionPartial(ctx, input, artifact, stored, extractResult.Errors)
	} else {
		uc.logExtractionSuccess(ctx, input, artifact, stored)
	}

	return &ExtractInBodyOutput{
		ArtifactID:   input.ArtifactID,
		Measurements: stored,
		Partial:      extractResult.Partial,
		Errors:       extractResult.Errors,
	}, nil
}

func (uc *ExtractInBodyUseCase) logExtractionSuccess(ctx context.Context, input ExtractInBodyInput, artifact *entities.Artifact, measurements []*entities.Measurement) {
	types := make([]string, len(measurements))
	for i, m := range measurements {
		types[i] = string(m.MeasurementType)
	}
	if err := uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "artifact.extraction_success",
		EntityType: "artifact",
		EntityID:   input.ArtifactID,
		Metadata: map[string]interface{}{
			"client_id":         artifact.ClientID,
			"measurement_count": len(measurements),
			"measurement_types": types,
		},
	}); err != nil {
		log.Printf("audit log error: %v", err)
	}
}

func (uc *ExtractInBodyUseCase) logExtractionPartial(ctx context.Context, input ExtractInBodyInput, artifact *entities.Artifact, measurements []*entities.Measurement, errors []string) {
	types := make([]string, len(measurements))
	for i, m := range measurements {
		types[i] = string(m.MeasurementType)
	}
	if err := uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "artifact.extraction_partial",
		EntityType: "artifact",
		EntityID:   input.ArtifactID,
		Metadata: map[string]interface{}{
			"client_id":         artifact.ClientID,
			"measurement_count": len(measurements),
			"measurement_types": types,
			"errors":            errors,
		},
	}); err != nil {
		log.Printf("audit log error: %v", err)
	}
}

func (uc *ExtractInBodyUseCase) logExtractionFailed(ctx context.Context, input ExtractInBodyInput, artifact *entities.Artifact, reason string) {
	if err := uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "artifact.extraction_failed",
		EntityType: "artifact",
		EntityID:   input.ArtifactID,
		Metadata: map[string]interface{}{
			"client_id": artifact.ClientID,
			"reason":    reason,
		},
	}); err != nil {
		log.Printf("audit log error: %v", err)
	}
}

// NewExtractInBodyJobHandler creates a worker.JobHandler for extract_inbody jobs.
func NewExtractInBodyJobHandler(uc *ExtractInBodyUseCase) func(ctx context.Context, job *entities.Job) error {
	return func(ctx context.Context, job *entities.Job) error {
		var payload ExtractInBodyJobPayload
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return fmt.Errorf("unmarshal job payload: %w", err)
		}

		input := ExtractInBodyInput{
			ArtifactID: payload.ArtifactID,
			CoachID:    payload.CoachID,
			MeasuredAt: payload.MeasuredAt,
		}

		_, err := uc.Execute(ctx, input)
		return err
	}
}
