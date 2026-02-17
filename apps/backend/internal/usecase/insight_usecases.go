package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// --- GetInsightQueueUseCase ---

// GetInsightQueueInput holds the input for querying the insight queue.
type GetInsightQueueInput struct {
	CoachID string
	Status  string
	Limit   int
	Offset  int
}

// GetInsightQueueOutput holds the paginated insight queue results.
type GetInsightQueueOutput struct {
	Insights []*entities.InsightCard `json:"insights"`
	Page     int                     `json:"page"`
	Limit    int                     `json:"limit"`
	Total    int                     `json:"total"`
}

// GetInsightQueueUseCase retrieves the coach's insight approval queue.
type GetInsightQueueUseCase struct {
	repo repository.InsightCardRepository
}

// NewGetInsightQueueUseCase creates a new GetInsightQueueUseCase.
func NewGetInsightQueueUseCase(repo repository.InsightCardRepository) *GetInsightQueueUseCase {
	return &GetInsightQueueUseCase{repo: repo}
}

// Execute retrieves insight cards for the coach's queue.
func (uc *GetInsightQueueUseCase) Execute(ctx context.Context, input GetInsightQueueInput) (*GetInsightQueueOutput, error) {
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if input.Limit <= 0 {
		input.Limit = 20
	}

	filter := repository.InsightCardFilter{
		CoachID: input.CoachID,
		Status:  input.Status,
		Limit:   input.Limit,
		Offset:  input.Offset,
	}

	insights, total, err := uc.repo.ListByCoach(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list insights: %w", err)
	}

	if insights == nil {
		insights = []*entities.InsightCard{}
	}

	page := 1
	if input.Limit > 0 {
		page = (input.Offset / input.Limit) + 1
	}

	return &GetInsightQueueOutput{
		Insights: insights,
		Page:     page,
		Limit:    input.Limit,
		Total:    total,
	}, nil
}

// --- ApproveInsightUseCase ---

// ApproveInsightUseCase handles approving an insight card.
type ApproveInsightUseCase struct {
	repo      repository.InsightCardRepository
	auditRepo repository.AuditRepository
}

// NewApproveInsightUseCase creates a new ApproveInsightUseCase.
func NewApproveInsightUseCase(repo repository.InsightCardRepository, auditRepo repository.AuditRepository) *ApproveInsightUseCase {
	return &ApproveInsightUseCase{repo: repo, auditRepo: auditRepo}
}

// Execute approves an insight card and logs an audit event.
func (uc *ApproveInsightUseCase) Execute(ctx context.Context, insightID, coachID string) (*entities.InsightCard, error) {
	insight, err := uc.repo.FindByID(ctx, insightID)
	if err != nil {
		return nil, fmt.Errorf("find insight: %w", err)
	}
	if insight == nil {
		return nil, &ValidationError{Message: "insight not found"}
	}
	if insight.CoachID != coachID {
		return nil, &AuthorizationError{Message: "not authorized to approve this insight"}
	}

	if err := insight.Approve(); err != nil {
		return nil, &InvalidTransitionError{Message: err.Error()}
	}

	if err := uc.repo.Update(ctx, insight); err != nil {
		return nil, fmt.Errorf("update insight: %w", err)
	}

	if auditErr := uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    coachID,
		Action:     "insight.approve",
		EntityType: "insight_card",
		EntityID:   insightID,
	}); auditErr != nil {
		log.Printf("audit log error: %v", auditErr)
	}

	return insight, nil
}

// --- DismissInsightUseCase ---

// DismissInsightUseCase handles dismissing an insight card.
type DismissInsightUseCase struct {
	repo      repository.InsightCardRepository
	auditRepo repository.AuditRepository
}

// NewDismissInsightUseCase creates a new DismissInsightUseCase.
func NewDismissInsightUseCase(repo repository.InsightCardRepository, auditRepo repository.AuditRepository) *DismissInsightUseCase {
	return &DismissInsightUseCase{repo: repo, auditRepo: auditRepo}
}

// Execute dismisses an insight card and logs an audit event.
func (uc *DismissInsightUseCase) Execute(ctx context.Context, insightID, coachID string) (*entities.InsightCard, error) {
	insight, err := uc.repo.FindByID(ctx, insightID)
	if err != nil {
		return nil, fmt.Errorf("find insight: %w", err)
	}
	if insight == nil {
		return nil, &ValidationError{Message: "insight not found"}
	}
	if insight.CoachID != coachID {
		return nil, &AuthorizationError{Message: "not authorized to dismiss this insight"}
	}

	if err := insight.Dismiss(); err != nil {
		return nil, &InvalidTransitionError{Message: err.Error()}
	}

	if err := uc.repo.Update(ctx, insight); err != nil {
		return nil, fmt.Errorf("update insight: %w", err)
	}

	if auditErr := uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    coachID,
		Action:     "insight.dismiss",
		EntityType: "insight_card",
		EntityID:   insightID,
	}); auditErr != nil {
		log.Printf("audit log error: %v", auditErr)
	}

	return insight, nil
}

// --- EditInsightUseCase ---

// EditInsightInput holds the input for editing an insight card.
type EditInsightInput struct {
	InsightID string
	CoachID   string
	Title     string
	Body      string
}

// EditInsightUseCase handles editing the title and body of an insight card.
type EditInsightUseCase struct {
	repo      repository.InsightCardRepository
	auditRepo repository.AuditRepository
}

// NewEditInsightUseCase creates a new EditInsightUseCase.
func NewEditInsightUseCase(repo repository.InsightCardRepository, auditRepo repository.AuditRepository) *EditInsightUseCase {
	return &EditInsightUseCase{repo: repo, auditRepo: auditRepo}
}

// Execute updates the title and body of an insight card.
func (uc *EditInsightUseCase) Execute(ctx context.Context, input EditInsightInput) (*entities.InsightCard, error) {
	insight, err := uc.repo.FindByID(ctx, input.InsightID)
	if err != nil {
		return nil, fmt.Errorf("find insight: %w", err)
	}
	if insight == nil {
		return nil, &ValidationError{Message: "insight not found"}
	}
	if insight.CoachID != input.CoachID {
		return nil, &AuthorizationError{Message: "not authorized to edit this insight"}
	}

	if err := insight.UpdateText(input.Title, input.Body); err != nil {
		return nil, &ValidationError{Message: err.Error()}
	}

	if err := uc.repo.Update(ctx, insight); err != nil {
		return nil, fmt.Errorf("update insight: %w", err)
	}

	if auditErr := uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "insight.edit",
		EntityType: "insight_card",
		EntityID:   input.InsightID,
		Metadata: map[string]interface{}{
			"title": input.Title,
		},
	}); auditErr != nil {
		log.Printf("audit log error: %v", auditErr)
	}

	return insight, nil
}

// --- ShareInsightUseCase ---

// ShareInsightUseCase handles sharing an approved insight card with the client.
type ShareInsightUseCase struct {
	repo      repository.InsightCardRepository
	auditRepo repository.AuditRepository
}

// NewShareInsightUseCase creates a new ShareInsightUseCase.
func NewShareInsightUseCase(repo repository.InsightCardRepository, auditRepo repository.AuditRepository) *ShareInsightUseCase {
	return &ShareInsightUseCase{repo: repo, auditRepo: auditRepo}
}

// Execute transitions an insight from approved to shared.
func (uc *ShareInsightUseCase) Execute(ctx context.Context, insightID, coachID string) (*entities.InsightCard, error) {
	insight, err := uc.repo.FindByID(ctx, insightID)
	if err != nil {
		return nil, fmt.Errorf("find insight: %w", err)
	}
	if insight == nil {
		return nil, &ValidationError{Message: "insight not found"}
	}
	if insight.CoachID != coachID {
		return nil, &AuthorizationError{Message: "not authorized to share this insight"}
	}

	if err := insight.Share(); err != nil {
		return nil, &InvalidTransitionError{Message: err.Error()}
	}

	if err := uc.repo.Update(ctx, insight); err != nil {
		return nil, fmt.Errorf("update insight: %w", err)
	}

	if auditErr := uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    coachID,
		Action:     "insight.share",
		EntityType: "insight_card",
		EntityID:   insightID,
	}); auditErr != nil {
		log.Printf("audit log error: %v", auditErr)
	}

	return insight, nil
}

// --- GetClientInsightsUseCase ---

// GetClientInsightsInput holds the input for querying client insights.
type GetClientInsightsInput struct {
	ClientID string
	CoachID  string
	Status   string
	Limit    int
	Offset   int
}

// GetClientInsightsOutput holds the paginated client insights results.
type GetClientInsightsOutput struct {
	Insights []*entities.InsightCard `json:"insights"`
	Page     int                     `json:"page"`
	Limit    int                     `json:"limit"`
	Total    int                     `json:"total"`
}

// GetClientInsightsUseCase retrieves insights for a specific client.
type GetClientInsightsUseCase struct {
	repo repository.InsightCardRepository
}

// NewGetClientInsightsUseCase creates a new GetClientInsightsUseCase.
func NewGetClientInsightsUseCase(repo repository.InsightCardRepository) *GetClientInsightsUseCase {
	return &GetClientInsightsUseCase{repo: repo}
}

// Execute retrieves insight cards for a specific client.
func (uc *GetClientInsightsUseCase) Execute(ctx context.Context, input GetClientInsightsInput) (*GetClientInsightsOutput, error) {
	if input.ClientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}
	if input.Limit <= 0 {
		input.Limit = 20
	}

	filter := repository.InsightCardFilter{
		ClientID: input.ClientID,
		CoachID:  input.CoachID,
		Status:   input.Status,
		Limit:    input.Limit,
		Offset:   input.Offset,
	}

	insights, total, err := uc.repo.ListByClient(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list client insights: %w", err)
	}

	if insights == nil {
		insights = []*entities.InsightCard{}
	}

	page := 1
	if input.Limit > 0 {
		page = (input.Offset / input.Limit) + 1
	}

	return &GetClientInsightsOutput{
		Insights: insights,
		Page:     page,
		Limit:    input.Limit,
		Total:    total,
	}, nil
}

// --- Error types ---

// AuthorizationError represents a forbidden action.
type AuthorizationError struct {
	Message string
}

func (e *AuthorizationError) Error() string {
	return e.Message
}

// IsAuthorizationError checks whether the given error is an AuthorizationError.
func IsAuthorizationError(err error) bool {
	_, ok := err.(*AuthorizationError)
	return ok
}

// InvalidTransitionError represents an invalid status transition.
type InvalidTransitionError struct {
	Message string
}

func (e *InvalidTransitionError) Error() string {
	return e.Message
}

// IsInvalidTransitionError checks whether the given error is an InvalidTransitionError.
func IsInvalidTransitionError(err error) bool {
	_, ok := err.(*InvalidTransitionError)
	return ok
}
