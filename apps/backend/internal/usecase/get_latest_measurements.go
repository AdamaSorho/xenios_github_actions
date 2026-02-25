package usecase

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GetLatestMeasurementsUseCase handles retrieving the latest measurement for each type.
type GetLatestMeasurementsUseCase struct {
	measurementRepo repository.MeasurementRepository
	coachClientRepo repository.CoachClientRepository
	auditRepo       repository.AuditRepository
}

// NewGetLatestMeasurementsUseCase creates a new GetLatestMeasurementsUseCase.
func NewGetLatestMeasurementsUseCase(
	measurementRepo repository.MeasurementRepository,
	coachClientRepo repository.CoachClientRepository,
	auditRepo repository.AuditRepository,
) *GetLatestMeasurementsUseCase {
	return &GetLatestMeasurementsUseCase{
		measurementRepo: measurementRepo,
		coachClientRepo: coachClientRepo,
		auditRepo:       auditRepo,
	}
}

// GetLatestMeasurementsInput holds the input parameters.
type GetLatestMeasurementsInput struct {
	CoachID  string
	ClientID string
}

// GetLatestMeasurementsOutput holds the result.
type GetLatestMeasurementsOutput struct {
	Measurements []*entities.Measurement `json:"measurements"`
}

// Execute retrieves the latest measurement for each type for a client.
func (uc *GetLatestMeasurementsUseCase) Execute(ctx context.Context, input GetLatestMeasurementsInput) (*GetLatestMeasurementsOutput, error) {
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if input.ClientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}

	// Authorization check
	rel, err := uc.coachClientRepo.FindByCoachAndClient(ctx, input.CoachID, input.ClientID)
	if err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, &AuthorizationError{Message: "not authorized to access this client's data"}
	}

	measurements, err := uc.measurementRepo.FindLatestByClientID(ctx, input.ClientID)
	if err != nil {
		return nil, err
	}

	if measurements == nil {
		measurements = []*entities.Measurement{}
	}

	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "phi.measurements_accessed",
		EntityType: "client",
		EntityID:   input.ClientID,
		Metadata: map[string]interface{}{
			"endpoint": "latest",
		},
	})

	return &GetLatestMeasurementsOutput{
		Measurements: measurements,
	}, nil
}
