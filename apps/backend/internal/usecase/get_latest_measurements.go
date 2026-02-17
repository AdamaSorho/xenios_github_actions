package usecase

import (
	"context"
	"log"
	"time"

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

// Execute retrieves the latest measurement for each type for a client.
func (uc *GetLatestMeasurementsUseCase) Execute(ctx context.Context, coachID, clientID string) ([]*entities.LatestMeasurement, error) {
	if coachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if clientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}

	// Verify coach-client relationship
	rel, err := uc.coachClientRepo.FindByCoachAndClient(ctx, coachID, clientID)
	if err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, &AuthorizationError{Message: "access denied: no coach-client relationship"}
	}

	results, err := uc.measurementRepo.FindLatestByClientID(ctx, clientID)
	if err != nil {
		return nil, err
	}

	// Log PHI access audit event
	auditEvent := &entities.AuditEvent{
		ActorID:    coachID,
		Action:     "phi.access",
		EntityType: "measurement",
		EntityID:   clientID,
		Metadata: map[string]interface{}{
			"endpoint": "measurements/latest",
		},
		CreatedAt: time.Now(),
	}
	if err := uc.auditRepo.LogEvent(ctx, auditEvent); err != nil {
		log.Printf("failed to log audit event: %v", err)
	}

	if results == nil {
		results = []*entities.LatestMeasurement{}
	}

	return results, nil
}
