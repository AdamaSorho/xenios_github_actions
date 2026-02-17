package usecase

import (
	"context"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// QueryAuditLogUseCase handles querying audit log entries.
type QueryAuditLogUseCase struct {
	auditRepo repository.AuditRepository
}

// NewQueryAuditLogUseCase creates a new QueryAuditLogUseCase.
func NewQueryAuditLogUseCase(auditRepo repository.AuditRepository) *QueryAuditLogUseCase {
	return &QueryAuditLogUseCase{auditRepo: auditRepo}
}

const (
	defaultAuditLimit = 50
	maxAuditLimit     = 1000
)

// QueryAuditLogInput holds the query parameters.
type QueryAuditLogInput struct {
	ActorID    string
	Action     string
	EntityType string
	EntityID   string
	From       *time.Time
	To         *time.Time
	Limit      int
	Offset     int
}

// QueryAuditLogOutput holds the query result.
type QueryAuditLogOutput struct {
	Events []*entities.AuditEvent `json:"events"`
	Total  int                    `json:"total"`
	Limit  int                    `json:"limit"`
	Offset int                    `json:"offset"`
}

// Execute retrieves audit events matching the filter.
func (uc *QueryAuditLogUseCase) Execute(ctx context.Context, input QueryAuditLogInput) (*QueryAuditLogOutput, error) {
	if input.Limit <= 0 {
		input.Limit = defaultAuditLimit
	}
	if input.Limit > maxAuditLimit {
		input.Limit = maxAuditLimit
	}
	if input.Offset < 0 {
		input.Offset = 0
	}

	filter := entities.AuditQueryFilter{
		ActorID:    input.ActorID,
		Action:     input.Action,
		EntityType: input.EntityType,
		EntityID:   input.EntityID,
		From:       input.From,
		To:         input.To,
		Limit:      input.Limit,
		Offset:     input.Offset,
	}

	events, total, err := uc.auditRepo.Query(ctx, filter)
	if err != nil {
		return nil, err
	}

	if events == nil {
		events = []*entities.AuditEvent{}
	}

	return &QueryAuditLogOutput{
		Events: events,
		Total:  total,
		Limit:  input.Limit,
		Offset: input.Offset,
	}, nil
}
