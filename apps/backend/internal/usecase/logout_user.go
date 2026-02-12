package usecase

import (
	"context"
	"fmt"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// LogoutUserUseCase handles user logout by revoking all refresh tokens.
type LogoutUserUseCase struct {
	tokenRepo repository.RefreshTokenRepository
	auditRepo repository.AuditRepository
}

// NewLogoutUserUseCase creates a new LogoutUserUseCase.
func NewLogoutUserUseCase(
	tokenRepo repository.RefreshTokenRepository,
	auditRepo repository.AuditRepository,
) *LogoutUserUseCase {
	return &LogoutUserUseCase{
		tokenRepo: tokenRepo,
		auditRepo: auditRepo,
	}
}

// Execute revokes all refresh tokens for the given user.
func (uc *LogoutUserUseCase) Execute(ctx context.Context, userID string) error {
	if userID == "" {
		return &ValidationError{Message: "user_id is required"}
	}

	if err := uc.tokenRepo.RevokeAllForUser(ctx, userID); err != nil {
		return fmt.Errorf("revoke tokens: %w", err)
	}

	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    userID,
		Action:     "auth.logout",
		EntityType: "user",
		EntityID:   userID,
	})

	return nil
}
