package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GetInsightQueueUseCase retrieves draft insight cards for a coach.
type GetInsightQueueUseCase struct {
	insightRepo repository.InsightCardRepository
}

// NewGetInsightQueueUseCase creates a new GetInsightQueueUseCase.
func NewGetInsightQueueUseCase(insightRepo repository.InsightCardRepository) *GetInsightQueueUseCase {
	return &GetInsightQueueUseCase{insightRepo: insightRepo}
}

// GetInsightQueueInput holds the input for querying the insight queue.
type GetInsightQueueInput struct {
	CoachID string
	Status  string
	Page    int
	Limit   int
}

// GetInsightQueueOutput holds the output of the insight queue query.
type GetInsightQueueOutput struct {
	Insights   []*entities.InsightCard `json:"insights"`
	Pagination PaginationInfo          `json:"pagination"`
}

// PaginationInfo holds pagination metadata.
type PaginationInfo struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// Execute retrieves the insight queue for the authenticated coach.
func (uc *GetInsightQueueUseCase) Execute(ctx context.Context, input GetInsightQueueInput) (*GetInsightQueueOutput, error) {
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if input.Status != "" && !entities.IsValidInsightStatus(input.Status) {
		return nil, &ValidationError{Message: fmt.Sprintf("invalid status: %s", input.Status)}
	}

	if input.Page <= 0 {
		input.Page = 1
	}
	if input.Limit <= 0 {
		input.Limit = 20
	}

	status := input.Status
	if status == "" {
		status = string(entities.InsightStatusDraft)
	}

	insights, total, err := uc.insightRepo.ListByCoachID(ctx, entities.InsightQueryFilter{
		CoachID: input.CoachID,
		Status:  status,
		Page:    input.Page,
		Limit:   input.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list insight queue: %w", err)
	}

	if insights == nil {
		insights = []*entities.InsightCard{}
	}

	return &GetInsightQueueOutput{
		Insights: insights,
		Pagination: PaginationInfo{
			Page:  input.Page,
			Limit: input.Limit,
			Total: total,
		},
	}, nil
}

// GetClientInsightsUseCase retrieves insight cards for a specific client.
type GetClientInsightsUseCase struct {
	insightRepo repository.InsightCardRepository
}

// NewGetClientInsightsUseCase creates a new GetClientInsightsUseCase.
func NewGetClientInsightsUseCase(insightRepo repository.InsightCardRepository) *GetClientInsightsUseCase {
	return &GetClientInsightsUseCase{insightRepo: insightRepo}
}

// GetClientInsightsInput holds the input for querying client insights.
type GetClientInsightsInput struct {
	CoachID  string
	ClientID string
	Status   string
	Page     int
	Limit    int
}

// Execute retrieves insights for a specific client.
func (uc *GetClientInsightsUseCase) Execute(ctx context.Context, input GetClientInsightsInput) (*GetInsightQueueOutput, error) {
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if input.ClientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}
	if input.Status != "" && !entities.IsValidInsightStatus(input.Status) {
		return nil, &ValidationError{Message: fmt.Sprintf("invalid status: %s", input.Status)}
	}

	if input.Page <= 0 {
		input.Page = 1
	}
	if input.Limit <= 0 {
		input.Limit = 20
	}

	insights, total, err := uc.insightRepo.ListByClientID(ctx, entities.InsightQueryFilter{
		ClientID: input.ClientID,
		Status:   input.Status,
		Page:     input.Page,
		Limit:    input.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list client insights: %w", err)
	}

	if insights == nil {
		insights = []*entities.InsightCard{}
	}

	return &GetInsightQueueOutput{
		Insights: insights,
		Pagination: PaginationInfo{
			Page:  input.Page,
			Limit: input.Limit,
			Total: total,
		},
	}, nil
}

// ApproveInsightUseCase handles approving a draft insight card.
type ApproveInsightUseCase struct {
	insightRepo repository.InsightCardRepository
	auditRepo   repository.AuditRepository
}

// NewApproveInsightUseCase creates a new ApproveInsightUseCase.
func NewApproveInsightUseCase(insightRepo repository.InsightCardRepository, auditRepo repository.AuditRepository) *ApproveInsightUseCase {
	return &ApproveInsightUseCase{insightRepo: insightRepo, auditRepo: auditRepo}
}

// InsightActionInput holds the input for insight state transitions.
type InsightActionInput struct {
	InsightID string
	CoachID   string
}

// Execute approves a draft insight card.
func (uc *ApproveInsightUseCase) Execute(ctx context.Context, input InsightActionInput) (*entities.InsightCard, error) {
	if input.InsightID == "" {
		return nil, &ValidationError{Message: "insight_id is required"}
	}
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}

	card, err := uc.insightRepo.FindByID(ctx, input.InsightID)
	if err != nil {
		return nil, fmt.Errorf("find insight: %w", err)
	}
	if card == nil {
		return nil, &ValidationError{Message: "insight not found"}
	}
	if card.CoachID != input.CoachID {
		return nil, &AuthenticationError{Message: "not authorized to approve this insight"}
	}
	if !card.CanTransitionTo(entities.InsightStatusApproved) {
		return nil, &ValidationError{Message: fmt.Sprintf("cannot approve insight with status %s", card.Status)}
	}

	now := time.Now()
	card.Status = entities.InsightStatusApproved
	card.ApprovedAt = &now

	updated, err := uc.insightRepo.Update(ctx, card)
	if err != nil {
		return nil, fmt.Errorf("update insight: %w", err)
	}

	logInsightAudit(ctx, uc.auditRepo, input.CoachID, "insight.approve", input.InsightID)

	return updated, nil
}

// DismissInsightUseCase handles dismissing a draft insight card.
type DismissInsightUseCase struct {
	insightRepo repository.InsightCardRepository
	auditRepo   repository.AuditRepository
}

// NewDismissInsightUseCase creates a new DismissInsightUseCase.
func NewDismissInsightUseCase(insightRepo repository.InsightCardRepository, auditRepo repository.AuditRepository) *DismissInsightUseCase {
	return &DismissInsightUseCase{insightRepo: insightRepo, auditRepo: auditRepo}
}

// Execute dismisses a draft insight card.
func (uc *DismissInsightUseCase) Execute(ctx context.Context, input InsightActionInput) (*entities.InsightCard, error) {
	if input.InsightID == "" {
		return nil, &ValidationError{Message: "insight_id is required"}
	}
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}

	card, err := uc.insightRepo.FindByID(ctx, input.InsightID)
	if err != nil {
		return nil, fmt.Errorf("find insight: %w", err)
	}
	if card == nil {
		return nil, &ValidationError{Message: "insight not found"}
	}
	if card.CoachID != input.CoachID {
		return nil, &AuthenticationError{Message: "not authorized to dismiss this insight"}
	}
	if !card.CanTransitionTo(entities.InsightStatusDismissed) {
		return nil, &ValidationError{Message: fmt.Sprintf("cannot dismiss insight with status %s", card.Status)}
	}

	now := time.Now()
	card.Status = entities.InsightStatusDismissed
	card.DismissedAt = &now

	updated, err := uc.insightRepo.Update(ctx, card)
	if err != nil {
		return nil, fmt.Errorf("update insight: %w", err)
	}

	logInsightAudit(ctx, uc.auditRepo, input.CoachID, "insight.reject", input.InsightID)

	return updated, nil
}

// EditInsightUseCase handles editing the title and body of an insight card.
type EditInsightUseCase struct {
	insightRepo repository.InsightCardRepository
	auditRepo   repository.AuditRepository
}

// NewEditInsightUseCase creates a new EditInsightUseCase.
func NewEditInsightUseCase(insightRepo repository.InsightCardRepository, auditRepo repository.AuditRepository) *EditInsightUseCase {
	return &EditInsightUseCase{insightRepo: insightRepo, auditRepo: auditRepo}
}

// EditInsightInput holds the input for editing an insight card.
type EditInsightInput struct {
	InsightID string
	CoachID   string
	Title     string
	Body      string
}

// Execute edits the title and/or body of an insight card.
func (uc *EditInsightUseCase) Execute(ctx context.Context, input EditInsightInput) (*entities.InsightCard, error) {
	if input.InsightID == "" {
		return nil, &ValidationError{Message: "insight_id is required"}
	}
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}

	input.Title = strings.TrimSpace(input.Title)
	input.Body = strings.TrimSpace(input.Body)

	if input.Title == "" && input.Body == "" {
		return nil, &ValidationError{Message: "title or body is required"}
	}

	card, err := uc.insightRepo.FindByID(ctx, input.InsightID)
	if err != nil {
		return nil, fmt.Errorf("find insight: %w", err)
	}
	if card == nil {
		return nil, &ValidationError{Message: "insight not found"}
	}
	if card.CoachID != input.CoachID {
		return nil, &AuthenticationError{Message: "not authorized to edit this insight"}
	}
	if card.Status.IsTerminal() {
		return nil, &ValidationError{Message: fmt.Sprintf("cannot edit insight with terminal status %s", card.Status)}
	}

	if input.Title != "" {
		card.Title = input.Title
	}
	if input.Body != "" {
		card.Body = input.Body
	}

	updated, err := uc.insightRepo.Update(ctx, card)
	if err != nil {
		return nil, fmt.Errorf("update insight: %w", err)
	}

	logInsightAudit(ctx, uc.auditRepo, input.CoachID, "insight.edit", input.InsightID)

	return updated, nil
}

// ShareInsightUseCase handles sharing an approved insight card with the client.
type ShareInsightUseCase struct {
	insightRepo repository.InsightCardRepository
	auditRepo   repository.AuditRepository
}

// NewShareInsightUseCase creates a new ShareInsightUseCase.
func NewShareInsightUseCase(insightRepo repository.InsightCardRepository, auditRepo repository.AuditRepository) *ShareInsightUseCase {
	return &ShareInsightUseCase{insightRepo: insightRepo, auditRepo: auditRepo}
}

// Execute shares an approved insight card with the client.
func (uc *ShareInsightUseCase) Execute(ctx context.Context, input InsightActionInput) (*entities.InsightCard, error) {
	if input.InsightID == "" {
		return nil, &ValidationError{Message: "insight_id is required"}
	}
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}

	card, err := uc.insightRepo.FindByID(ctx, input.InsightID)
	if err != nil {
		return nil, fmt.Errorf("find insight: %w", err)
	}
	if card == nil {
		return nil, &ValidationError{Message: "insight not found"}
	}
	if card.CoachID != input.CoachID {
		return nil, &AuthenticationError{Message: "not authorized to share this insight"}
	}
	if !card.CanTransitionTo(entities.InsightStatusShared) {
		return nil, &ValidationError{Message: fmt.Sprintf("cannot share insight with status %s", card.Status)}
	}

	now := time.Now()
	card.Status = entities.InsightStatusShared
	card.SharedAt = &now

	updated, err := uc.insightRepo.Update(ctx, card)
	if err != nil {
		return nil, fmt.Errorf("update insight: %w", err)
	}

	logInsightAudit(ctx, uc.auditRepo, input.CoachID, "insight.share", input.InsightID)

	return updated, nil
}

// logInsightAudit is a shared helper that logs an audit event for insight actions.
func logInsightAudit(ctx context.Context, auditRepo repository.AuditRepository, actorID, action, insightID string) {
	if auditErr := auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    actorID,
		Action:     action,
		EntityType: "insight",
		EntityID:   insightID,
	}); auditErr != nil {
		log.Printf("audit log error: %v", auditErr)
	}
}
