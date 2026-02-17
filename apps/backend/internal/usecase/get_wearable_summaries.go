package usecase

import (
	"context"
	"log"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GetWearableSummariesUseCase retrieves wearable summary data for a client.
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

// GetWearableSummariesInput holds the input for retrieving wearable summaries.
type GetWearableSummariesInput struct {
	CoachID  string
	ClientID string
	Source   string
	From     *time.Time
	To       *time.Time
	Page     int
	Limit    int
}

// GetWearableSummariesOutput holds the paginated output.
type GetWearableSummariesOutput struct {
	Summaries  []*entities.WearableSummary `json:"summaries"`
	Pagination PaginationInfo             `json:"pagination"`
}

// Execute retrieves wearable summaries for a client after authorization check.
func (uc *GetWearableSummariesUseCase) Execute(ctx context.Context, input GetWearableSummariesInput) (*GetWearableSummariesOutput, error) {
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if input.ClientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}

	// Verify coach-client relationship
	rel, err := uc.coachClientRepo.FindByCoachAndClient(ctx, input.CoachID, input.ClientID)
	if err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, &AuthorizationError{Message: "access denied: no coach-client relationship"}
	}

	// Apply defaults
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Limit <= 0 {
		input.Limit = defaultLimit
	}
	if input.Limit > maxLimit {
		input.Limit = maxLimit
	}

	offset := (input.Page - 1) * input.Limit

	filter := entities.WearableSummaryFilter{
		ClientID: input.ClientID,
		Source:   input.Source,
		From:     input.From,
		To:       input.To,
		Limit:    input.Limit,
		Offset:   offset,
	}

	summaries, total, err := uc.wearableRepo.FindByClientID(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Log PHI access audit event
	auditEvent := &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "phi.access",
		EntityType: "wearable_summary",
		EntityID:   input.ClientID,
		Metadata: map[string]interface{}{
			"endpoint": "wearable-summaries",
			"source":   input.Source,
		},
		CreatedAt: time.Now(),
	}
	if err := uc.auditRepo.LogEvent(ctx, auditEvent); err != nil {
		log.Printf("failed to log audit event: %v", err)
	}

	if summaries == nil {
		summaries = []*entities.WearableSummary{}
	}

	return &GetWearableSummariesOutput{
		Summaries: summaries,
		Pagination: PaginationInfo{
			Page:  input.Page,
			Limit: input.Limit,
			Total: total,
		},
	}, nil
}
