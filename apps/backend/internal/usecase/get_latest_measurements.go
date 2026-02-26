package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GetLatestMeasurementsUseCase retrieves the most recent measurement for each type.
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

// GetLatestMeasurementsInput holds the input for the use case.
type GetLatestMeasurementsInput struct {
	CoachID  string
	ClientID string
}

// Execute retrieves the latest measurement for each type after verifying authorization.
func (uc *GetLatestMeasurementsUseCase) Execute(ctx context.Context, input GetLatestMeasurementsInput) ([]*entities.LatestMeasurement, error) {
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if input.ClientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}

	rel, err := uc.coachClientRepo.FindByCoachAndClient(ctx, input.CoachID, input.ClientID)
	if err != nil {
		return nil, fmt.Errorf("check coach-client relationship: %w", err)
	}
	if rel == nil {
		return nil, &AuthorizationError{Message: "not authorized to access this client's data"}
	}

	results, err := uc.measurementRepo.FindLatestByClientID(ctx, input.ClientID)
	if err != nil {
		return nil, fmt.Errorf("find latest measurements: %w", err)
	}

	if auditErr := uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "phi.access",
		EntityType: "client",
		EntityID:   input.ClientID,
		Metadata:   map[string]interface{}{"resource": "measurements_latest"},
	}); auditErr != nil {
		log.Printf("audit log error: %v", auditErr)
	}

	if results == nil {
		results = []*entities.LatestMeasurement{}
	}

	return results, nil
}
