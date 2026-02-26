package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// ExtractLabResultsUseCase processes lab result files and stores extracted measurements.
type ExtractLabResultsUseCase struct {
	artifactRepo    repository.ArtifactRepository
	measurementRepo repository.MeasurementRepository
	fileReader      repository.FileContentReader
	auditRepo       repository.AuditRepository
	csvParser       repository.LabFileParser
	pdfParser       repository.LabFileParser
}

// NewExtractLabResultsUseCase creates a new ExtractLabResultsUseCase.
func NewExtractLabResultsUseCase(
	artifactRepo repository.ArtifactRepository,
	measurementRepo repository.MeasurementRepository,
	fileReader repository.FileContentReader,
	auditRepo repository.AuditRepository,
	csvParser repository.LabFileParser,
	pdfParser repository.LabFileParser,
) *ExtractLabResultsUseCase {
	return &ExtractLabResultsUseCase{
		artifactRepo:    artifactRepo,
		measurementRepo: measurementRepo,
		fileReader:      fileReader,
		auditRepo:       auditRepo,
		csvParser:       csvParser,
		pdfParser:       pdfParser,
	}
}

// ExtractLabResultsPayload is the job payload for extract_lab_results jobs.
type ExtractLabResultsPayload struct {
	ArtifactID string `json:"artifact_id"`
	ClientID   string `json:"client_id"`
	CoachID    string `json:"coach_id"`
}

// ExtractLabResultsOutput holds the result of lab result extraction.
type ExtractLabResultsOutput struct {
	MeasurementsCreated int                    `json:"measurements_created"`
	Measurements        []*entities.Measurement `json:"measurements"`
}

// Execute processes a lab result file and stores extracted measurements.
func (uc *ExtractLabResultsUseCase) Execute(ctx context.Context, job *entities.Job) error {
	var payload ExtractLabResultsPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("parse job payload: %w", err)
	}

	if payload.ArtifactID == "" {
		return &ValidationError{Message: "artifact_id is required"}
	}
	if payload.ClientID == "" {
		return &ValidationError{Message: "client_id is required"}
	}
	if payload.CoachID == "" {
		return &ValidationError{Message: "coach_id is required"}
	}

	// Look up the artifact
	artifact, err := uc.artifactRepo.FindByID(ctx, payload.ArtifactID)
	if err != nil {
		return fmt.Errorf("find artifact: %w", err)
	}
	if artifact == nil {
		return &ValidationError{Message: fmt.Sprintf("artifact not found: %s", payload.ArtifactID)}
	}

	// Read file content
	content, err := uc.fileReader.ReadContent(ctx, artifact.StorageKey)
	if err != nil {
		return fmt.Errorf("read file content: %w", err)
	}

	// Select parser based on file type
	parser, err := uc.selectParser(artifact.FileName)
	if err != nil {
		return err
	}

	// Parse markers from file content
	markers, err := parser.Parse(content)
	if err != nil {
		return fmt.Errorf("parse lab results: %w", err)
	}

	// Convert parsed markers to measurements
	now := time.Now()
	measurements := make([]*entities.Measurement, len(markers))
	for i, m := range markers {
		measType := entities.NormalizeMarkerName(m.Name)
		flag := entities.DetermineFlag(m.Value, m.ReferenceLow, m.ReferenceHigh)

		measurements[i] = &entities.Measurement{
			ClientID:        payload.ClientID,
			RecordedBy:      payload.CoachID,
			MeasurementType: measType,
			Value:           m.Value,
			Unit:            m.Unit,
			MeasuredAt:      now,
			ArtifactID:      &payload.ArtifactID,
			ReferenceLow:    m.ReferenceLow,
			ReferenceHigh:   m.ReferenceHigh,
			Flag:            flag,
		}
	}

	// Store measurements
	_, err = uc.measurementRepo.CreateBatch(ctx, measurements)
	if err != nil {
		return fmt.Errorf("store measurements: %w", err)
	}

	// Log audit event
	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    payload.CoachID,
		Action:     "lab.extract",
		EntityType: "artifact",
		EntityID:   payload.ArtifactID,
		Metadata: map[string]interface{}{
			"client_id":           payload.ClientID,
			"markers_extracted":   len(markers),
			"file_name":           artifact.FileName,
		},
	})

	return nil
}

// selectParser returns the appropriate parser based on file extension.
func (uc *ExtractLabResultsUseCase) selectParser(fileName string) (repository.LabFileParser, error) {
	lower := strings.ToLower(fileName)
	switch {
	case strings.HasSuffix(lower, ".csv"):
		return uc.csvParser, nil
	case strings.HasSuffix(lower, ".pdf"):
		return uc.pdfParser, nil
	default:
		return nil, &ValidationError{Message: fmt.Sprintf("unsupported file format: %s", fileName)}
	}
}
