package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GetClientMeasurementsUseCase retrieves measurements for a client with filtering and pagination.
type GetClientMeasurementsUseCase struct {
	measurementRepo repository.MeasurementRepository
	coachClientRepo repository.CoachClientRepository
	auditRepo       repository.AuditRepository
}

// NewGetClientMeasurementsUseCase creates a new GetClientMeasurementsUseCase.
func NewGetClientMeasurementsUseCase(
	measurementRepo repository.MeasurementRepository,
	coachClientRepo repository.CoachClientRepository,
	auditRepo repository.AuditRepository,
) *GetClientMeasurementsUseCase {
	return &GetClientMeasurementsUseCase{
		measurementRepo: measurementRepo,
		coachClientRepo: coachClientRepo,
		auditRepo:       auditRepo,
	}
}

// GetClientMeasurementsInput holds the input for the use case.
type GetClientMeasurementsInput struct {
	CoachID  string
	ClientID string
	Filter   entities.MeasurementFilter
}

// Execute retrieves measurements after verifying coach-client authorization.
func (uc *GetClientMeasurementsUseCase) Execute(ctx context.Context, input GetClientMeasurementsInput) (*entities.MeasurementResult, error) {
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if input.ClientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}

	if err := uc.verifyCoachAccess(ctx, input.CoachID, input.ClientID); err != nil {
		return nil, err
	}

	if input.Filter.Page < 1 {
		input.Filter.Page = 1
	}
	if input.Filter.Limit < 1 || input.Filter.Limit > 100 {
		input.Filter.Limit = 20
	}
	input.Filter.ClientID = input.ClientID

	measurements, total, err := uc.measurementRepo.FindByClientID(ctx, input.Filter)
	if err != nil {
		return nil, fmt.Errorf("find measurements: %w", err)
	}

	uc.logPHIAccess(ctx, input.CoachID, input.ClientID, "measurements")

	if measurements == nil {
		measurements = []*entities.Measurement{}
	}

	return &entities.MeasurementResult{
		Measurements: measurements,
		Pagination: entities.Pagination{
			Page:  input.Filter.Page,
			Limit: input.Filter.Limit,
			Total: total,
		},
	}, nil
}

func (uc *GetClientMeasurementsUseCase) verifyCoachAccess(ctx context.Context, coachID, clientID string) error {
	rel, err := uc.coachClientRepo.FindByCoachAndClient(ctx, coachID, clientID)
	if err != nil {
		return fmt.Errorf("check coach-client relationship: %w", err)
	}
	if rel == nil {
		return &AuthorizationError{Message: "not authorized to access this client's data"}
	}
	return nil
}

func (uc *GetClientMeasurementsUseCase) logPHIAccess(ctx context.Context, coachID, clientID, resource string) {
	if auditErr := uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    coachID,
		Action:     "phi.access",
		EntityType: "client",
		EntityID:   clientID,
		Metadata:   map[string]interface{}{"resource": resource},
	}); auditErr != nil {
		log.Printf("audit log error: %v", auditErr)
	}
}
