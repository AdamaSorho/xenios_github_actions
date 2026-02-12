package repository

import (
	"context"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// RefreshTokenRepository defines the interface for refresh token persistence.
type RefreshTokenRepository interface {
	Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (*entities.RefreshToken, error)
	FindByTokenHash(ctx context.Context, tokenHash string) (*entities.RefreshToken, error)
	MarkUsed(ctx context.Context, id string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}
