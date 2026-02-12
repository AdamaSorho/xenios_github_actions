package repository

import (
	"context"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryRefreshTokenRepository is an in-memory implementation of RefreshTokenRepository.
type InMemoryRefreshTokenRepository struct {
	mu     sync.RWMutex
	tokens map[string]*entities.RefreshToken
	count  int
}

// NewInMemoryRefreshTokenRepository creates a new InMemoryRefreshTokenRepository.
func NewInMemoryRefreshTokenRepository() *InMemoryRefreshTokenRepository {
	return &InMemoryRefreshTokenRepository{
		tokens: make(map[string]*entities.RefreshToken),
	}
}

// Create stores a new refresh token.
func (r *InMemoryRefreshTokenRepository) Create(_ context.Context, userID, tokenHash string, expiresAt time.Time) (*entities.RefreshToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.count++
	id := "rt-" + time.Now().Format("20060102150405") + "-" + string(rune('0'+r.count))

	rt := &entities.RefreshToken{
		ID:        id,
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		Used:      false,
		CreatedAt: time.Now(),
	}

	r.tokens[id] = rt
	return rt, nil
}

// FindByTokenHash returns a refresh token by its hash, or nil if not found.
func (r *InMemoryRefreshTokenRepository) FindByTokenHash(_ context.Context, tokenHash string) (*entities.RefreshToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, rt := range r.tokens {
		if rt.TokenHash == tokenHash {
			return rt, nil
		}
	}
	return nil, nil
}

// MarkUsed marks a refresh token as used.
func (r *InMemoryRefreshTokenRepository) MarkUsed(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if rt, ok := r.tokens[id]; ok {
		rt.Used = true
	}
	return nil
}

// RevokeAllForUser revokes all refresh tokens for a given user.
func (r *InMemoryRefreshTokenRepository) RevokeAllForUser(_ context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for _, rt := range r.tokens {
		if rt.UserID == userID {
			rt.RevokedAt = &now
		}
	}
	return nil
}
