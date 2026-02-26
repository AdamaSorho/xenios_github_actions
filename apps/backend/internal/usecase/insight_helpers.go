package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// InsightActionDeps holds common dependencies for insight action use cases
// (approve, dismiss, edit, share).
type InsightActionDeps struct {
	insightRepo repository.InsightCardRepository
	auditRepo   repository.AuditRepository
}

// NewInsightActionDeps creates a new InsightActionDeps.
func NewInsightActionDeps(
	insightRepo repository.InsightCardRepository,
	auditRepo repository.AuditRepository,
) InsightActionDeps {
	return InsightActionDeps{
		insightRepo: insightRepo,
		auditRepo:   auditRepo,
	}
}

// fetchAndAuthorize validates input, fetches the card by ID, and checks
// that the requesting coach owns the insight.
func (d *InsightActionDeps) fetchAndAuthorize(ctx context.Context, insightID, coachID, action string) (*entities.InsightCard, error) {
	if insightID == "" {
		return nil, &ValidationError{Message: "insight_id is required"}
	}
	if coachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}

	card, err := d.insightRepo.FindByID(ctx, insightID)
	if err != nil {
		return nil, fmt.Errorf("find insight: %w", err)
	}
	if card == nil {
		return nil, &ValidationError{Message: "insight not found"}
	}

	if card.CoachID != coachID {
		return nil, &AuthorizationError{
			Message: fmt.Sprintf("not authorized to %s this insight", action),
		}
	}
	return card, nil
}

// updateAndAudit persists the updated card and logs an audit event.
func (d *InsightActionDeps) updateAndAudit(
	ctx context.Context,
	card *entities.InsightCard,
	action, coachID string,
	extraMetadata map[string]interface{},
) (*entities.InsightCard, error) {
	if err := d.insightRepo.Update(ctx, card); err != nil {
		return nil, fmt.Errorf("update insight: %w", err)
	}

	metadata := map[string]interface{}{
		"client_id": card.ClientID,
		"title":     card.Title,
	}
	for k, v := range extraMetadata {
		metadata[k] = v
	}

	if auditErr := d.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    coachID,
		Action:     action,
		EntityType: "insight_card",
		EntityID:   card.ID,
		Metadata:   metadata,
	}); auditErr != nil {
		log.Printf("audit log error: %v", auditErr)
	}

	return card, nil
}

// InsightListOutput holds paginated insight card results.
type InsightListOutput struct {
	Insights []*entities.InsightCard `json:"insights"`
	Total    int                     `json:"total"`
	Limit    int                     `json:"limit"`
	Offset   int                     `json:"offset"`
}

// newInsightListOutput creates a list output, ensuring a non-nil slice.
func newInsightListOutput(insights []*entities.InsightCard, total, limit, offset int) *InsightListOutput {
	if insights == nil {
		insights = []*entities.InsightCard{}
	}
	return &InsightListOutput{
		Insights: insights,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}
}

// normalizePagination clamps limit to [1, maxLimit] with a default.
func normalizePagination(limit, maxLimit, defaultLimit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}
