package usecase

import (
	"context"
	"fmt"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GetLatestMeasurementsUseCase handles querying the latest measurement per type.
type GetLatestMeasurementsUseCase struct {
	measurementRepo repository.MeasurementRepository
	clientAccessChecker
}

// NewGetLatestMeasurementsUseCase creates a new GetLatestMeasurementsUseCase.
func NewGetLatestMeasurementsUseCase(
	measurementRepo repository.MeasurementRepository,
	ccRepo repository.CoachClientRepository,
	auditRepo repository.AuditRepository,
) *GetLatestMeasurementsUseCase {
	return &GetLatestMeasurementsUseCase{
		measurementRepo:     measurementRepo,
		clientAccessChecker: clientAccessChecker{ccRepo: ccRepo, auditRepo: auditRepo},
	}
}

// GetLatestMeasurementsOutput holds the latest measurement for each type.
type GetLatestMeasurementsOutput struct {
	Measurements []*entities.Measurement `json:"measurements"`
}

// Execute retrieves the latest measurement for each type for a client.
func (uc *GetLatestMeasurementsUseCase) Execute(ctx context.Context, coachID, clientID string) (*GetLatestMeasurementsOutput, error) {
	if coachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if clientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}

	if err := uc.verifyCoachClient(ctx, coachID, clientID); err != nil {
		return nil, err
	}

	measurements, err := uc.measurementRepo.FindLatestByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("find latest measurements: %w", err)
	}

	if measurements == nil {
		measurements = []*entities.Measurement{}
	}

	uc.logPHIAccess(ctx, coachID, clientID, "measurements_latest")

	return &GetLatestMeasurementsOutput{
		Measurements: measurements,
	}, nil
}
