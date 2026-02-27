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
	fileDownloader  repository.FileDownloader
	pdfExtractor    repository.PDFTextExtractor
	measurementRepo repository.MeasurementRepository
	auditRepo       repository.AuditRepository
}

// NewExtractInBodyUseCase creates a new ExtractInBodyUseCase.
func NewExtractInBodyUseCase(
	artifactRepo repository.ArtifactRepository,
	fileDownloader repository.FileDownloader,
	pdfExtractor repository.PDFTextExtractor,
	measurementRepo repository.MeasurementRepository,
	auditRepo repository.AuditRepository,
) *ExtractInBodyUseCase {
	return &ExtractInBodyUseCase{
		artifactRepo:    artifactRepo,
		fileDownloader:  fileDownloader,
		pdfExtractor:    pdfExtractor,
		measurementRepo: measurementRepo,
		auditRepo:       auditRepo,
	}
}

// ExtractInBodyInput holds the input for InBody extraction.
type ExtractInBodyInput struct {
	ArtifactID string
	CoachID    string
}

// ExtractInBodyOutput holds the result of InBody extraction.
type ExtractInBodyOutput struct {
	ArtifactID          string `json:"artifact_id"`
	MeasurementsCreated int    `json:"measurements_created"`
	Partial             bool   `json:"partial"`
}

// Execute downloads the InBody PDF, extracts metrics, and stores them as measurements.
func (uc *ExtractInBodyUseCase) Execute(ctx context.Context, input ExtractInBodyInput) (*ExtractInBodyOutput, error) {
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	artifact, err := uc.findAndValidateArtifact(ctx, input)
	if err != nil {
		return nil, err
	}

	result, err := uc.downloadAndExtract(ctx, artifact)
	if err != nil {
		uc.logFailureEvent(ctx, input, err)
		return nil, err
	}

	measurements, err := uc.storeMeasurements(ctx, artifact, input.CoachID, result)
	if err != nil {
		uc.logFailureEvent(ctx, input, err)
		return nil, fmt.Errorf("store measurements: %w", err)
	}

	if _, err := uc.artifactRepo.UpdateStatus(ctx, input.ArtifactID, entities.ArtifactStatusProcessed); err != nil {
		return nil, fmt.Errorf("update artifact status: %w", err)
	}

	uc.logSuccessEvent(ctx, input, len(measurements), result.Partial)

	return &ExtractInBodyOutput{
		ArtifactID:          input.ArtifactID,
		MeasurementsCreated: len(measurements),
		Partial:             result.Partial,
	}, nil
}

func (uc *ExtractInBodyUseCase) validateInput(input ExtractInBodyInput) error {
	if input.ArtifactID == "" {
		return &ValidationError{Message: "artifact_id is required"}
	}
	if input.CoachID == "" {
		return &ValidationError{Message: "coach_id is required"}
	}
	return nil
}

func (uc *ExtractInBodyUseCase) findAndValidateArtifact(ctx context.Context, input ExtractInBodyInput) (*entities.Artifact, error) {
	artifact, err := uc.artifactRepo.FindByID(ctx, input.ArtifactID)
	if err != nil {
		return nil, fmt.Errorf("find artifact: %w", err)
	}
	if artifact == nil {
		return nil, &ValidationError{Message: "artifact not found"}
	}
	if artifact.CoachID != input.CoachID {
		return nil, &AuthenticationError{Message: "not authorized to process this artifact"}
	}
	if artifact.Status != entities.ArtifactStatusUploaded {
		return nil, &ValidationError{Message: fmt.Sprintf("artifact status is %s, expected uploaded", artifact.Status)}
	}
	return artifact, nil
}

func (uc *ExtractInBodyUseCase) downloadAndExtract(ctx context.Context, artifact *entities.Artifact) (*entities.InBodyResult, error) {
	pdfData, err := uc.fileDownloader.DownloadFile(ctx, artifact.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("download file: %w", err)
	}

	text, err := uc.pdfExtractor.ExtractText(ctx, pdfData)
	if err != nil {
		return nil, fmt.Errorf("extract PDF text: %w", err)
	}

	result, err := entities.ParseInBodyText(text)
	if err != nil {
		return nil, fmt.Errorf("parse InBody text: %w", err)
	}

	return result, nil
}

func (uc *ExtractInBodyUseCase) storeMeasurements(
	ctx context.Context,
	artifact *entities.Artifact,
	coachID string,
	result *entities.InBodyResult,
) ([]*entities.Measurement, error) {
	now := time.Now()
	measuredAt := now
	if result.MeasuredAt != nil {
		measuredAt = *result.MeasuredAt
	}

	measurements := make([]*entities.Measurement, 0, len(result.Metrics))
	for _, metric := range result.Metrics {
		artifactID := artifact.ID
		m := &entities.Measurement{
			ClientID:        artifact.ClientID,
			RecordedBy:      coachID,
			MeasurementType: metric.Type,
			Value:           metric.Value,
			Unit:            metric.Unit,
			MeasuredAt:      measuredAt,
			ArtifactID:      &artifactID,
		}
		measurements = append(measurements, m)
	}

	return uc.measurementRepo.CreateBatch(ctx, measurements)
}

func (uc *ExtractInBodyUseCase) logSuccessEvent(ctx context.Context, input ExtractInBodyInput, count int, partial bool) {
	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "inbody.extraction_success",
		EntityType: "artifact",
		EntityID:   input.ArtifactID,
		Metadata: map[string]interface{}{
			"measurements_created": count,
			"partial":              partial,
		},
	})
}

func (uc *ExtractInBodyUseCase) logFailureEvent(ctx context.Context, input ExtractInBodyInput, err error) {
	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "inbody.extraction_failed",
		EntityType: "artifact",
		EntityID:   input.ArtifactID,
		Metadata: map[string]interface{}{
			"error": err.Error(),
		},
	})
}
