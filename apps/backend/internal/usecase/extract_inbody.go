package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// ExtractInBodyUseCase handles extracting body composition metrics from InBody PDF scans.
type ExtractInBodyUseCase struct {
	artifactRepo    repository.ArtifactRepository
	fileStorage     repository.FileStorageRepository
	pdfExtractor    repository.PDFTextExtractor
	inbodyParser    repository.InBodyTextParser
	measurementRepo repository.MeasurementRepository
	auditRepo       repository.AuditRepository
}

// NewExtractInBodyUseCase creates a new ExtractInBodyUseCase.
func NewExtractInBodyUseCase(
	artifactRepo repository.ArtifactRepository,
	fileStorage repository.FileStorageRepository,
	pdfExtractor repository.PDFTextExtractor,
	inbodyParser repository.InBodyTextParser,
	measurementRepo repository.MeasurementRepository,
	auditRepo repository.AuditRepository,
) *ExtractInBodyUseCase {
	return &ExtractInBodyUseCase{
		artifactRepo:    artifactRepo,
		fileStorage:     fileStorage,
		pdfExtractor:    pdfExtractor,
		inbodyParser:    inbodyParser,
		measurementRepo: measurementRepo,
		auditRepo:       auditRepo,
	}
}

// ExtractInBodyInput holds the input for extracting InBody data.
type ExtractInBodyInput struct {
	ArtifactID string
}

// ExtractInBodyOutput holds the result of the extraction.
type ExtractInBodyOutput struct {
	Measurements []*entities.Measurement `json:"measurements"`
	IsPartial    bool                    `json:"is_partial"`
	Errors       []string                `json:"errors,omitempty"`
}

// Execute downloads the InBody PDF, extracts metrics, and stores them as measurements.
func (uc *ExtractInBodyUseCase) Execute(ctx context.Context, input ExtractInBodyInput) (*ExtractInBodyOutput, error) {
	if input.ArtifactID == "" {
		return nil, &ValidationError{Message: "artifact_id is required"}
	}

	// Fetch artifact
	artifact, err := uc.artifactRepo.FindByID(ctx, input.ArtifactID)
	if err != nil {
		return nil, fmt.Errorf("find artifact: %w", err)
	}
	if artifact == nil {
		return nil, &ValidationError{Message: "artifact not found"}
	}
	if artifact.Status != entities.ArtifactStatusUploaded {
		return nil, &ValidationError{Message: fmt.Sprintf("artifact status is %s, expected uploaded", artifact.Status)}
	}

	// Download PDF from storage
	pdfData, err := uc.fileStorage.DownloadObject(ctx, artifact.StorageKey)
	if err != nil {
		uc.logAuditEvent(ctx, artifact, "artifact.extraction_failed", map[string]interface{}{
			"error": fmt.Sprintf("download failed: %v", err),
		})
		return nil, fmt.Errorf("download artifact: %w", err)
	}

	// Extract text from PDF
	text, err := uc.pdfExtractor.ExtractText(ctx, pdfData)
	if err != nil {
		uc.logAuditEvent(ctx, artifact, "artifact.extraction_failed", map[string]interface{}{
			"error": fmt.Sprintf("pdf extraction failed: %v", err),
		})
		return nil, fmt.Errorf("extract PDF text: %w", err)
	}

	// Parse InBody metrics from extracted text
	measuredAt := time.Now()
	result := uc.inbodyParser.Parse(text, artifact.ClientID, artifact.CoachID, artifact.ID, measuredAt)

	if len(result.Measurements) == 0 {
		uc.logAuditEvent(ctx, artifact, "artifact.extraction_failed", map[string]interface{}{
			"error":  "no metrics extracted",
			"errors": result.Errors,
		})
		return nil, fmt.Errorf("no InBody metrics could be extracted from PDF")
	}

	// Store measurements
	stored, err := uc.measurementRepo.CreateBatch(ctx, result.Measurements)
	if err != nil {
		uc.logAuditEvent(ctx, artifact, "artifact.extraction_failed", map[string]interface{}{
			"error": fmt.Sprintf("storage failed: %v", err),
		})
		return nil, fmt.Errorf("store measurements: %w", err)
	}

	// Update artifact status to processed
	_, _ = uc.artifactRepo.UpdateStatus(ctx, input.ArtifactID, entities.ArtifactStatusProcessed)

	// Log audit event
	auditAction := "artifact.extraction_success"
	if result.IsPartial {
		auditAction = "artifact.extraction_partial"
	}
	uc.logAuditEvent(ctx, artifact, auditAction, map[string]interface{}{
		"measurements_count": len(stored),
		"is_partial":         result.IsPartial,
		"errors":             result.Errors,
	})

	return &ExtractInBodyOutput{
		Measurements: stored,
		IsPartial:    result.IsPartial,
		Errors:       result.Errors,
	}, nil
}

func (uc *ExtractInBodyUseCase) logAuditEvent(ctx context.Context, artifact *entities.Artifact, action string, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["artifact_id"] = artifact.ID
	metadata["file_name"] = artifact.FileName
	metadata["client_id"] = artifact.ClientID

	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    artifact.CoachID,
		Action:     action,
		EntityType: "artifact",
		EntityID:   artifact.ID,
		Metadata:   metadata,
	})
}
