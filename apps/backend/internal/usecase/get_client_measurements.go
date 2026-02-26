package usecase

import (
	"context"
	"fmt"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

const (
	defaultMeasurementLimit = 20
	maxMeasurementLimit     = 100
)

// GetClientMeasurementsUseCase handles querying client measurements.
type GetClientMeasurementsUseCase struct {
	measurementRepo repository.MeasurementRepository
	ccRepo          repository.CoachClientRepository
	auditRepo       repository.AuditRepository
}

// NewGetClientMeasurementsUseCase creates a new GetClientMeasurementsUseCase.
func NewGetClientMeasurementsUseCase(
	measurementRepo repository.MeasurementRepository,
	ccRepo repository.CoachClientRepository,
	auditRepo repository.AuditRepository,
) *GetClientMeasurementsUseCase {
	return &GetClientMeasurementsUseCase{
		measurementRepo: measurementRepo,
		ccRepo:          ccRepo,
		auditRepo:       auditRepo,
	}
}

// GetClientMeasurementsInput holds the input for querying measurements.
type GetClientMeasurementsInput struct {
	CoachID         string
	ClientID        string
	MeasurementType string
	From            *entities.MeasurementFilter
	Filter          entities.MeasurementFilter
}

// GetClientMeasurementsOutput holds the output for querying measurements.
type GetClientMeasurementsOutput struct {
	Measurements []*entities.Measurement `json:"measurements"`
	Pagination   PaginationOutput        `json:"pagination"`
}

// PaginationOutput holds pagination metadata.
type PaginationOutput struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// Execute retrieves measurements for a client after authorization check.
func (uc *GetClientMeasurementsUseCase) Execute(ctx context.Context, coachID string, filter entities.MeasurementFilter) (*GetClientMeasurementsOutput, error) {
	if coachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if filter.ClientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}

	if err := uc.verifyCoachClient(ctx, coachID, filter.ClientID); err != nil {
		return nil, err
	}

	if filter.Limit <= 0 {
		filter.Limit = defaultMeasurementLimit
	}
	if filter.Limit > maxMeasurementLimit {
		filter.Limit = maxMeasurementLimit
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	measurements, total, err := uc.measurementRepo.FindByClientID(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("find measurements: %w", err)
	}

	if measurements == nil {
		measurements = []*entities.Measurement{}
	}

	page := 1
	if filter.Limit > 0 {
		page = (filter.Offset / filter.Limit) + 1
	}

	uc.logPHIAccess(ctx, coachID, filter.ClientID, "measurements")

	return &GetClientMeasurementsOutput{
		Measurements: measurements,
		Pagination: PaginationOutput{
			Page:  page,
			Limit: filter.Limit,
			Total: total,
		},
	}, nil
}

func (uc *GetClientMeasurementsUseCase) verifyCoachClient(ctx context.Context, coachID, clientID string) error {
	rel, err := uc.ccRepo.FindByCoachAndClient(ctx, coachID, clientID)
	if err != nil {
		return fmt.Errorf("check coach-client relationship: %w", err)
	}
	if rel == nil {
		return &AuthenticationError{Message: "forbidden: not authorized to access this client"}
	}
	return nil
}

func (uc *GetClientMeasurementsUseCase) logPHIAccess(ctx context.Context, coachID, clientID, resource string) {
	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    coachID,
		Action:     "phi.access",
		EntityType: "client",
		EntityID:   clientID,
		Metadata: map[string]interface{}{
			"resource": resource,
		},
	})
}
