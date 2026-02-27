package usecase

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GetWearableSummariesUseCase handles retrieving wearable summaries for a client.
type GetWearableSummariesUseCase struct {
	wearableRepo    repository.WearableSummaryRepository
	coachClientRepo repository.CoachClientRepository
	auditRepo       repository.AuditRepository
}

// NewGetWearableSummariesUseCase creates a new GetWearableSummariesUseCase.
func NewGetWearableSummariesUseCase(
	wearableRepo repository.WearableSummaryRepository,
	coachClientRepo repository.CoachClientRepository,
	auditRepo repository.AuditRepository,
) *GetWearableSummariesUseCase {
	return &GetWearableSummariesUseCase{
		wearableRepo:    wearableRepo,
		coachClientRepo: coachClientRepo,
		auditRepo:       auditRepo,
	}
}

// Execute retrieves wearable summaries for a client.
func (uc *GetWearableSummariesUseCase) Execute(ctx context.Context, coachID, clientID string, limit int) ([]*entities.WearableSummary, error) {
	if err := validateCoachClientInput(coachID, clientID); err != nil {
		return nil, err
	}

	if err := authorizeCoachClient(ctx, uc.coachClientRepo, coachID, clientID); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	result, err := uc.wearableRepo.FindByClientID(ctx, clientID, limit)
	if err != nil {
		return nil, err
	}

	logPHIAccess(ctx, uc.auditRepo, coachID, clientID, "wearable_summaries")
	return result, nil
}
