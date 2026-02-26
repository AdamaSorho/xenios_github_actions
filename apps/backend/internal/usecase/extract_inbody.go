package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// ExtractInBodyPayload is the JSON payload for an extract_inbody job.
type ExtractInBodyPayload struct {
	ArtifactID string `json:"artifact_id"`
	ClientID   string `json:"client_id"`
	CoachID    string `json:"coach_id"`
}

// ExtractInBodyOutput holds the result of an InBody extraction.
type ExtractInBodyOutput struct {
	Measurements []*entities.Measurement `json:"measurements"`
	IsPartial    bool                    `json:"is_partial"`
	FieldCount   int                     `json:"field_count"`
}

// ExtractInBodyUseCase handles extracting InBody metrics from a PDF artifact.
type ExtractInBodyUseCase struct {
	artifactRepo    repository.ArtifactRepository
	measurementRepo repository.MeasurementRepository
	fileStorage     repository.FileStorageRepository
	pdfExtractor    repository.PDFExtractor
	auditRepo       repository.AuditRepository
}

// NewExtractInBodyUseCase creates a new ExtractInBodyUseCase.
func NewExtractInBodyUseCase(
	artifactRepo repository.ArtifactRepository,
	measurementRepo repository.MeasurementRepository,
	fileStorage repository.FileStorageRepository,
	pdfExtractor repository.PDFExtractor,
	auditRepo repository.AuditRepository,
) *ExtractInBodyUseCase {
	return &ExtractInBodyUseCase{
		artifactRepo:    artifactRepo,
		measurementRepo: measurementRepo,
		fileStorage:     fileStorage,
		pdfExtractor:    pdfExtractor,
		auditRepo:       auditRepo,
	}
}

// Execute processes an extract_inbody job: downloads the PDF, extracts metrics, stores measurements.
func (uc *ExtractInBodyUseCase) Execute(ctx context.Context, job *entities.Job) error {
	if job == nil {
		return &ValidationError{Message: "job is required"}
	}

	var payload ExtractInBodyPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal job payload: %w", err)
	}

	if err := uc.validatePayload(payload); err != nil {
		return err
	}

	artifact, err := uc.artifactRepo.FindByID(ctx, payload.ArtifactID)
	if err != nil {
		return fmt.Errorf("find artifact: %w", err)
	}
	if artifact == nil {
		return &ValidationError{Message: "artifact not found"}
	}

	if artifact.Status != entities.ArtifactStatusUploaded {
		return &ValidationError{Message: fmt.Sprintf("artifact status is %s, expected uploaded", artifact.Status)}
	}

	pdfData, err := uc.fileStorage.DownloadFile(ctx, artifact.StorageKey)
	if err != nil {
		uc.logAuditEvent(ctx, payload.CoachID, "artifact.extraction_failed", payload.ArtifactID, map[string]interface{}{
			"error": fmt.Sprintf("download file: %v", err),
		})
		return fmt.Errorf("download file: %w", err)
	}

	result, err := uc.pdfExtractor.ExtractInBody(ctx, pdfData)
	if err != nil {
		uc.logAuditEvent(ctx, payload.CoachID, "artifact.extraction_failed", payload.ArtifactID, map[string]interface{}{
			"error": fmt.Sprintf("extract inbody: %v", err),
		})
		return fmt.Errorf("extract inbody data: %w", err)
	}

	measurements := result.ToMeasurements(payload.ClientID, payload.CoachID, payload.ArtifactID)

	created, err := uc.measurementRepo.CreateBatch(ctx, measurements)
	if err != nil {
		uc.logAuditEvent(ctx, payload.CoachID, "artifact.extraction_failed", payload.ArtifactID, map[string]interface{}{
			"error": fmt.Sprintf("store measurements: %v", err),
		})
		return fmt.Errorf("store measurements: %w", err)
	}

	action := "artifact.extraction_success"
	if result.IsPartial {
		action = "artifact.extraction_partial"
	}

	uc.logAuditEvent(ctx, payload.CoachID, action, payload.ArtifactID, map[string]interface{}{
		"field_count":       result.FieldCount(),
		"total_fields":      entities.TotalFields,
		"is_partial":        result.IsPartial,
		"measurements_stored": len(created),
		"client_id":         payload.ClientID,
	})

	return nil
}

func (uc *ExtractInBodyUseCase) validatePayload(p ExtractInBodyPayload) error {
	if p.ArtifactID == "" {
		return &ValidationError{Message: "artifact_id is required in job payload"}
	}
	if p.ClientID == "" {
		return &ValidationError{Message: "client_id is required in job payload"}
	}
	if p.CoachID == "" {
		return &ValidationError{Message: "coach_id is required in job payload"}
	}
	return nil
}

func (uc *ExtractInBodyUseCase) logAuditEvent(ctx context.Context, actorID, action, artifactID string, metadata map[string]interface{}) {
	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    actorID,
		Action:     action,
		EntityType: "artifact",
		EntityID:   artifactID,
		Metadata:   metadata,
	})
}
