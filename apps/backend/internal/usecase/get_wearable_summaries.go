package usecase

import (
	"context"
	"fmt"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GetWearableSummariesUseCase handles querying wearable summaries for a client.
type GetWearableSummariesUseCase struct {
	wearableRepo repository.WearableSummaryRepository
	ccRepo       repository.CoachClientRepository
	auditRepo    repository.AuditRepository
}

// NewGetWearableSummariesUseCase creates a new GetWearableSummariesUseCase.
func NewGetWearableSummariesUseCase(
	wearableRepo repository.WearableSummaryRepository,
	ccRepo repository.CoachClientRepository,
	auditRepo repository.AuditRepository,
) *GetWearableSummariesUseCase {
	return &GetWearableSummariesUseCase{
		wearableRepo: wearableRepo,
		ccRepo:       ccRepo,
		auditRepo:    auditRepo,
	}
}

// GetWearableSummariesOutput holds the wearable summary result.
type GetWearableSummariesOutput struct {
	Summaries []*entities.WearableSummary `json:"summaries"`
}

// Execute retrieves wearable summaries for a client.
func (uc *GetWearableSummariesUseCase) Execute(ctx context.Context, coachID, clientID string, limit, offset int) (*GetWearableSummariesOutput, error) {
	if coachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if clientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}

	if err := uc.verifyCoachClient(ctx, coachID, clientID); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = defaultMeasurementLimit
	}
	if limit > maxMeasurementLimit {
		limit = maxMeasurementLimit
	}
	if offset < 0 {
		offset = 0
	}

	summaries, err := uc.wearableRepo.FindByClientID(ctx, clientID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find wearable summaries: %w", err)
	}

	if summaries == nil {
		summaries = []*entities.WearableSummary{}
	}

	uc.logPHIAccess(ctx, coachID, clientID, "wearable_summaries")

	return &GetWearableSummariesOutput{
		Summaries: summaries,
	}, nil
}

func (uc *GetWearableSummariesUseCase) verifyCoachClient(ctx context.Context, coachID, clientID string) error {
	rel, err := uc.ccRepo.FindByCoachAndClient(ctx, coachID, clientID)
	if err != nil {
		return fmt.Errorf("check coach-client relationship: %w", err)
	}
	if rel == nil {
		return &AuthenticationError{Message: "forbidden: not authorized to access this client"}
	}
	return nil
}

func (uc *GetWearableSummariesUseCase) logPHIAccess(ctx context.Context, coachID, clientID, resource string) {
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
