package usecase

import (
	"bytes"
	"context"
	"fmt"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// ExtractLabResultsUseCase handles parsing lab result files and storing measurements.
type ExtractLabResultsUseCase struct {
	measurementRepo repository.MeasurementRepository
	artifactRepo    repository.ArtifactRepository
	auditRepo       repository.AuditRepository
	parsers         map[string]repository.LabParser
}

// NewExtractLabResultsUseCase creates a new ExtractLabResultsUseCase.
func NewExtractLabResultsUseCase(
	measurementRepo repository.MeasurementRepository,
	artifactRepo repository.ArtifactRepository,
	auditRepo repository.AuditRepository,
	parsers map[string]repository.LabParser,
) *ExtractLabResultsUseCase {
	return &ExtractLabResultsUseCase{
		measurementRepo: measurementRepo,
		artifactRepo:    artifactRepo,
		auditRepo:       auditRepo,
		parsers:         parsers,
	}
}

// ExtractLabResultsInput holds the input for extracting lab results.
type ExtractLabResultsInput struct {
	ArtifactID  string
	CoachID     string
	ClientID    string
	Content     []byte
	ContentType string
}

// ExtractLabResultsOutput holds the result of the extraction.
type ExtractLabResultsOutput struct {
	MeasurementsCount int                      `json:"measurements_count"`
	Measurements      []entities.LabMeasurement `json:"measurements"`
}

// Execute parses lab result content, stores measurements, and logs an audit event.
func (uc *ExtractLabResultsUseCase) Execute(ctx context.Context, input ExtractLabResultsInput) (*ExtractLabResultsOutput, error) {
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	// Select parser based on content type
	parser, err := uc.selectParser(input.ContentType)
	if err != nil {
		return nil, err
	}

	// Parse the file content
	measurements, err := parser.Parse(bytes.NewReader(input.Content))
	if err != nil {
		return nil, fmt.Errorf("parse lab results: %w", err)
	}

	if len(measurements) == 0 {
		return &ExtractLabResultsOutput{
			MeasurementsCount: 0,
			Measurements:      measurements,
		}, nil
	}

	// Convert to repository inputs
	inputs := convertToMeasurementInputs(measurements, input)

	// Store measurements
	count, err := uc.measurementRepo.CreateBatch(ctx, inputs)
	if err != nil {
		return nil, fmt.Errorf("store measurements: %w", err)
	}

	// Log audit event
	uc.logAuditEvent(ctx, input, count, measurements)

	return &ExtractLabResultsOutput{
		MeasurementsCount: count,
		Measurements:      measurements,
	}, nil
}

func (uc *ExtractLabResultsUseCase) validateInput(input ExtractLabResultsInput) error {
	if input.ArtifactID == "" {
		return &ValidationError{Message: "artifact_id is required"}
	}
	if input.CoachID == "" {
		return &ValidationError{Message: "coach_id is required"}
	}
	if input.ClientID == "" {
		return &ValidationError{Message: "client_id is required"}
	}
	if len(input.Content) == 0 {
		return &ValidationError{Message: "content is required"}
	}
	if input.ContentType == "" {
		return &ValidationError{Message: "content_type is required"}
	}
	return nil
}

func (uc *ExtractLabResultsUseCase) selectParser(contentType string) (repository.LabParser, error) {
	parser, ok := uc.parsers[contentType]
	if !ok {
		return nil, &ValidationError{Message: fmt.Sprintf("unsupported content type: %s", contentType)}
	}
	return parser, nil
}

func convertToMeasurementInputs(measurements []entities.LabMeasurement, input ExtractLabResultsInput) []repository.MeasurementInput {
	inputs := make([]repository.MeasurementInput, len(measurements))
	for i, m := range measurements {
		inputs[i] = repository.MeasurementInput{
			ClientID:        input.ClientID,
			RecordedBy:      input.CoachID,
			MeasurementType: string(m.MeasurementType),
			Value:           m.Value,
			Unit:            m.Unit,
			ArtifactID:      &input.ArtifactID,
			ReferenceLow:    m.ReferenceLow,
			ReferenceHigh:   m.ReferenceHigh,
			Flag:            m.Flag,
		}
	}
	return inputs
}

func (uc *ExtractLabResultsUseCase) logAuditEvent(ctx context.Context, input ExtractLabResultsInput, count int, measurements []entities.LabMeasurement) {
	flaggedMarkers := countFlaggedMarkers(measurements)
	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "lab_results.extracted",
		EntityType: "artifact",
		EntityID:   input.ArtifactID,
		Metadata: map[string]interface{}{
			"client_id":         input.ClientID,
			"content_type":      input.ContentType,
			"measurements_count": count,
			"flagged_count":     flaggedMarkers,
		},
	})
}

func countFlaggedMarkers(measurements []entities.LabMeasurement) int {
	count := 0
	for _, m := range measurements {
		if m.Flag != nil && *m.Flag != entities.LabFlagNormal {
			count++
		}
	}
	return count
}
