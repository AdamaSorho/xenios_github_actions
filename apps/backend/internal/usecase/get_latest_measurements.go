package usecase

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GetLatestMeasurementsUseCase handles retrieving the latest measurement per type for a client.
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

// Execute retrieves the latest measurement per type for a client.
func (uc *GetLatestMeasurementsUseCase) Execute(ctx context.Context, coachID, clientID string) ([]*entities.Measurement, error) {
	if err := validateCoachClientInput(coachID, clientID); err != nil {
		return nil, err
	}

	if err := authorizeCoachClient(ctx, uc.coachClientRepo, coachID, clientID); err != nil {
		return nil, err
	}

	result, err := uc.measurementRepo.FindLatestByClientID(ctx, clientID)
	if err != nil {
		return nil, err
	}

	logPHIAccess(ctx, uc.auditRepo, coachID, clientID, "measurements.latest")
	return result, nil
}
