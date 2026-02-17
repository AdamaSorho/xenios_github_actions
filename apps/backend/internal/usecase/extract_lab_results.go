package usecase

import (
	"context"
	"fmt"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
	"github.com/xenios/backend/internal/infrastructure/labparser"
)

// ExtractLabResultsUseCase handles parsing lab result files and storing extracted measurements.
type ExtractLabResultsUseCase struct {
	artifactRepo    repository.ArtifactRepository
	measurementRepo repository.MeasurementRepository
	fileReader      repository.FileContentReader
	auditRepo       repository.AuditRepository
}

// NewExtractLabResultsUseCase creates a new ExtractLabResultsUseCase.
func NewExtractLabResultsUseCase(
	artifactRepo repository.ArtifactRepository,
	measurementRepo repository.MeasurementRepository,
	fileReader repository.FileContentReader,
	auditRepo repository.AuditRepository,
) *ExtractLabResultsUseCase {
	return &ExtractLabResultsUseCase{
		artifactRepo:    artifactRepo,
		measurementRepo: measurementRepo,
		fileReader:      fileReader,
		auditRepo:       auditRepo,
	}
}

// ExtractLabResultsInput holds the input for lab result extraction.
type ExtractLabResultsInput struct {
	ArtifactID string
	CoachID    string
}

// ExtractLabResultsOutput holds the result of lab result extraction.
type ExtractLabResultsOutput struct {
	ArtifactID     string `json:"artifact_id"`
	ExtractedCount int    `json:"extracted_count"`
	FlaggedCount   int    `json:"flagged_count"`
}

// Execute downloads the artifact file, parses lab results, stores measurements, and logs an audit event.
func (uc *ExtractLabResultsUseCase) Execute(ctx context.Context, input ExtractLabResultsInput) (*ExtractLabResultsOutput, error) {
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

	// Download file content
	data, err := uc.fileReader.GetObject(ctx, artifact.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("download file: %w", err)
	}

	// Parse lab results based on format
	results, err := parseLabFile(data, artifact.FileName)
	if err != nil {
		return nil, fmt.Errorf("parse lab file: %w", err)
	}

	// Convert LabResults to Measurements and store
	measurements := make([]*entities.Measurement, 0, len(results))
	flaggedCount := 0

	for _, r := range results {
		m := &entities.Measurement{
			ClientID:        artifact.ClientID,
			RecordedBy:      input.CoachID,
			MeasurementType: r.MarkerName,
			Value:           r.Value,
			Unit:            r.Unit,
			ReferenceLow:    r.ReferenceLow,
			ReferenceHigh:   r.ReferenceHigh,
			Flag:            r.Flag,
			ArtifactID:      &input.ArtifactID,
		}
		measurements = append(measurements, m)

		if r.Flag != "" && r.Flag != entities.FlagNormal {
			flaggedCount++
		}
	}

	if len(measurements) > 0 {
		_, err = uc.measurementRepo.BatchCreate(ctx, measurements)
		if err != nil {
			return nil, fmt.Errorf("store measurements: %w", err)
		}
	}

	// Log audit event
	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "lab.extract",
		EntityType: "artifact",
		EntityID:   input.ArtifactID,
		Metadata: map[string]interface{}{
			"extracted_count": len(results),
			"flagged_count":   flaggedCount,
			"file_name":       artifact.FileName,
			"client_id":       artifact.ClientID,
		},
	})

	return &ExtractLabResultsOutput{
		ArtifactID:     input.ArtifactID,
		ExtractedCount: len(results),
		FlaggedCount:   flaggedCount,
	}, nil
}

// parseLabFile detects the format and parses lab results.
func parseLabFile(data []byte, fileName string) ([]entities.LabResult, error) {
	format := labparser.DetectFormat(data, fileName)

	switch format {
	case labparser.FormatCSV:
		return labparser.ParseCSV(data)
	case labparser.FormatPDF:
		return labparser.ParsePDFText(data)
	default:
		return nil, fmt.Errorf("unsupported file format for %s", fileName)
	}
}
