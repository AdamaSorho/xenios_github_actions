package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// ExtractInBodyUseCase handles extraction of InBody PDF data into measurements.
type ExtractInBodyUseCase struct {
	artifactRepo    repository.ArtifactRepository
	measurementRepo repository.MeasurementRepository
	fileDownloader  repository.FileDownloader
	pdfExtractor    repository.PDFExtractor
	auditRepo       repository.AuditRepository
}

// NewExtractInBodyUseCase creates a new ExtractInBodyUseCase.
func NewExtractInBodyUseCase(
	artifactRepo repository.ArtifactRepository,
	measurementRepo repository.MeasurementRepository,
	fileDownloader repository.FileDownloader,
	pdfExtractor repository.PDFExtractor,
	auditRepo repository.AuditRepository,
) *ExtractInBodyUseCase {
	return &ExtractInBodyUseCase{
		artifactRepo:    artifactRepo,
		measurementRepo: measurementRepo,
		fileDownloader:  fileDownloader,
		pdfExtractor:    pdfExtractor,
		auditRepo:       auditRepo,
	}
}

// ExtractInBodyPayload is the expected JSON payload for extract_inbody jobs.
type ExtractInBodyPayload struct {
	ArtifactID string `json:"artifact_id"`
}

// ExtractInBodyOutput holds the result of the extraction.
type ExtractInBodyOutput struct {
	Measurements []*entities.Measurement `json:"measurements"`
	IsPartial    bool                    `json:"is_partial"`
	FieldCount   int                     `json:"field_count"`
}

// Execute processes an extract_inbody job: downloads PDF, extracts metrics, stores measurements.
func (uc *ExtractInBodyUseCase) Execute(ctx context.Context, job *entities.Job) (*ExtractInBodyOutput, error) {
	// Parse job payload
	var payload ExtractInBodyPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return nil, fmt.Errorf("invalid job payload: %w", err)
	}
	if payload.ArtifactID == "" {
		return nil, &ValidationError{Message: "artifact_id is required in job payload"}
	}

	// Fetch the artifact
	artifact, err := uc.artifactRepo.FindByID(ctx, payload.ArtifactID)
	if err != nil {
		return nil, fmt.Errorf("find artifact: %w", err)
	}
	if artifact == nil {
		return nil, &ValidationError{Message: "artifact not found"}
	}

	// Verify artifact is uploaded and is a document
	if artifact.Status != entities.ArtifactStatusUploaded {
		return nil, &ValidationError{Message: fmt.Sprintf("artifact status is %s, expected uploaded", artifact.Status)}
	}
	if artifact.Type != entities.ArtifactTypeDocument {
		return nil, &ValidationError{Message: fmt.Sprintf("artifact type is %s, expected document", artifact.Type)}
	}

	// Download the PDF file
	pdfContent, err := uc.fileDownloader.Download(ctx, artifact.StorageKey)
	if err != nil {
		uc.logAuditEvent(ctx, artifact, "extraction.inbody_failed", map[string]interface{}{
			"error": fmt.Sprintf("download failed: %v", err),
		})
		return nil, fmt.Errorf("download file: %w", err)
	}

	// Extract InBody metrics
	result, err := uc.pdfExtractor.ExtractInBody(ctx, pdfContent)
	if err != nil {
		uc.logAuditEvent(ctx, artifact, "extraction.inbody_failed", map[string]interface{}{
			"error": fmt.Sprintf("extraction failed: %v", err),
		})
		return nil, fmt.Errorf("extract inbody data: %w", err)
	}

	// Check if extraction yielded any results
	if result.IsEmpty() {
		uc.logAuditEvent(ctx, artifact, "extraction.inbody_failed", map[string]interface{}{
			"error": "no metrics could be extracted from the PDF",
		})
		return nil, fmt.Errorf("no metrics could be extracted from the PDF")
	}

	// Build measurement records from extracted data
	now := time.Now()
	measuredAt := now
	if result.MeasuredAt != nil {
		measuredAt = *result.MeasuredAt
	}

	isPartial := result.IsPartial()
	measurements := buildMeasurements(result, artifact, measuredAt, isPartial)

	// Store measurements
	stored, err := uc.measurementRepo.CreateBatch(ctx, measurements)
	if err != nil {
		uc.logAuditEvent(ctx, artifact, "extraction.inbody_failed", map[string]interface{}{
			"error": fmt.Sprintf("store measurements failed: %v", err),
		})
		return nil, fmt.Errorf("store measurements: %w", err)
	}

	// Log success or partial success audit event
	action := "extraction.inbody_success"
	if isPartial {
		action = "extraction.inbody_partial"
	}
	uc.logAuditEvent(ctx, artifact, action, map[string]interface{}{
		"fields_extracted": result.ExtractedFieldCount(),
		"total_fields":     entities.TotalExpectedFields,
		"is_partial":       isPartial,
	})

	return &ExtractInBodyOutput{
		Measurements: stored,
		IsPartial:    isPartial,
		FieldCount:   result.ExtractedFieldCount(),
	}, nil
}

func (uc *ExtractInBodyUseCase) logAuditEvent(ctx context.Context, artifact *entities.Artifact, action string, metadata map[string]interface{}) {
	metadata["artifact_id"] = artifact.ID
	metadata["client_id"] = artifact.ClientID
	metadata["file_name"] = artifact.FileName
	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    "system",
		Action:     action,
		EntityType: "artifact",
		EntityID:   artifact.ID,
		Metadata:   metadata,
	})
}

func buildMeasurements(result *entities.InBodyResult, artifact *entities.Artifact, measuredAt time.Time, isPartial bool) []*entities.Measurement {
	var measurements []*entities.Measurement

	addMeasurement := func(mt entities.MeasurementType, value *float64, unit string) {
		if value == nil {
			return
		}
		measurements = append(measurements, &entities.Measurement{
			ClientID:          artifact.ClientID,
			RecordedBy:        "system",
			MeasurementType:   mt,
			Value:             *value,
			Unit:              unit,
			MeasuredAt:        measuredAt,
			ArtifactID:        artifact.ID,
			PartialExtraction: isPartial,
		})
	}

	addMeasurement(entities.MeasurementTypeWeight, result.Weight, result.WeightUnit)
	addMeasurement(entities.MeasurementTypeBodyFatPct, result.BodyFatPct, "%")
	addMeasurement(entities.MeasurementTypeSkeletalMuscleMass, result.SkeletalMuscleMass, result.SMMUnit)
	addMeasurement(entities.MeasurementTypeBMR, result.BMR, "kcal")
	addMeasurement(entities.MeasurementTypeTotalBodyWater, result.TotalBodyWater, "L")
	addMeasurement(entities.MeasurementTypeLeanBodyMass, result.LeanBodyMass, result.LBMUnit)

	return measurements
}
