package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// RefreshTokenUseCase handles token refresh with rotation.
type RefreshTokenUseCase struct {
	userRepo     repository.UserRepository
	tokenRepo    repository.RefreshTokenRepository
	tokenService repository.TokenService
	auditRepo    repository.AuditRepository
}

// NewRefreshTokenUseCase creates a new RefreshTokenUseCase.
func NewRefreshTokenUseCase(
	userRepo repository.UserRepository,
	tokenRepo repository.RefreshTokenRepository,
	tokenService repository.TokenService,
	auditRepo repository.AuditRepository,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		tokenService: tokenService,
		auditRepo:    auditRepo,
	}
}

// RefreshOutput holds the output of a token refresh.
type RefreshOutput struct {
	Tokens *entities.AuthTokens `json:"tokens"`
}

// Execute refreshes the access token using a valid refresh token.
// The old refresh token is marked as used (rotation), and a new one is issued.
// If a previously-used token is presented (replay attack), all tokens for the user are revoked.
func (uc *RefreshTokenUseCase) Execute(ctx context.Context, rawRefreshToken string) (*RefreshOutput, error) {
	if rawRefreshToken == "" {
		return nil, &ValidationError{Message: "refresh_token is required"}
	}

	tokenHash := uc.tokenService.HashToken(rawRefreshToken)

	storedToken, err := uc.tokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("find refresh token: %w", err)
	}
	if storedToken == nil {
		return nil, &AuthenticationError{Message: "invalid refresh token"}
	}

	// Replay attack detection: if the token was already used, revoke all tokens for this user.
	if storedToken.Used {
		_ = uc.tokenRepo.RevokeAllForUser(ctx, storedToken.UserID)
		_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
			ActorID:    storedToken.UserID,
			Action:     "auth.token_replay_detected",
			EntityType: "refresh_token",
			EntityID:   storedToken.ID,
		})
		return nil, &AuthenticationError{Message: "token reuse detected, all sessions revoked"}
	}

	if !storedToken.IsUsable() {
		return nil, &AuthenticationError{Message: "refresh token expired or revoked"}
	}

	// Mark the old token as used
	if err := uc.tokenRepo.MarkUsed(ctx, storedToken.ID); err != nil {
		return nil, fmt.Errorf("mark token used: %w", err)
	}

	user, err := uc.userRepo.FindByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return nil, &AuthenticationError{Message: "user not found"}
	}

	accessToken, err := uc.tokenService.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	newRefreshRaw, err := uc.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	newRefreshHash := uc.tokenService.HashToken(newRefreshRaw)
	_, err = uc.tokenRepo.Create(ctx, user.ID, newRefreshHash, time.Now().Add(7*24*time.Hour))
	if err != nil {
		return nil, fmt.Errorf("store new refresh token: %w", err)
	}

	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    user.ID,
		Action:     "auth.token_refreshed",
		EntityType: "refresh_token",
		EntityID:   storedToken.ID,
	})

	return &RefreshOutput{
		Tokens: &entities.AuthTokens{
			AccessToken:  accessToken,
			RefreshToken: newRefreshRaw,
		},
	}, nil
}
