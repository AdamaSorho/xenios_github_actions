package usecase

import (
	"context"
	"fmt"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// clientAccessChecker provides shared authorization and PHI audit logging
// for use cases that access client health data.
type clientAccessChecker struct {
	ccRepo    repository.CoachClientRepository
	auditRepo repository.AuditRepository
}

func (c *clientAccessChecker) verifyCoachClient(ctx context.Context, coachID, clientID string) error {
	rel, err := c.ccRepo.FindByCoachAndClient(ctx, coachID, clientID)
	if err != nil {
		return fmt.Errorf("check coach-client relationship: %w", err)
	}
	if rel == nil {
		return &AuthenticationError{Message: "forbidden: not authorized to access this client"}
	}
	return nil
}

func (c *clientAccessChecker) logPHIAccess(ctx context.Context, coachID, clientID, resource string) {
	_ = c.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    coachID,
		Action:     "phi.access",
		EntityType: "client",
		EntityID:   clientID,
		Metadata: map[string]interface{}{
			"resource": resource,
		},
	})
}
