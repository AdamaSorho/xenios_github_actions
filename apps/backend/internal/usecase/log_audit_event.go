package usecase

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// LogAuditEventUseCase handles recording audit events.
type LogAuditEventUseCase struct {
	auditRepo repository.AuditRepository
}

// NewLogAuditEventUseCase creates a new LogAuditEventUseCase.
func NewLogAuditEventUseCase(auditRepo repository.AuditRepository) *LogAuditEventUseCase {
	return &LogAuditEventUseCase{auditRepo: auditRepo}
}

// LogAuditEventInput holds the input for logging an audit event.
type LogAuditEventInput struct {
	ActorID    string
	Action     string
	EntityType string
	EntityID   string
	Metadata   map[string]interface{}
	IPAddress  string
	UserAgent  string
}

// Execute validates and records an audit event.
func (uc *LogAuditEventUseCase) Execute(ctx context.Context, input LogAuditEventInput) error {
	if input.ActorID == "" {
		return &ValidationError{Message: "actor_id is required"}
	}
	if input.Action == "" {
		return &ValidationError{Message: "action is required"}
	}
	if input.EntityType == "" {
		return &ValidationError{Message: "entity_type is required"}
	}
	if input.EntityID == "" {
		return &ValidationError{Message: "entity_id is required"}
	}

	event := &entities.AuditEvent{
		ActorID:    input.ActorID,
		Action:     input.Action,
		EntityType: input.EntityType,
		EntityID:   input.EntityID,
		Metadata:   input.Metadata,
		IPAddress:  input.IPAddress,
		UserAgent:  input.UserAgent,
	}

	return uc.auditRepo.LogEvent(ctx, event)
}
