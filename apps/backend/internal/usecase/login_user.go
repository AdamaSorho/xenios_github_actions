package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// LoginUserUseCase handles user authentication.
type LoginUserUseCase struct {
	userRepo     repository.UserRepository
	tokenRepo    repository.RefreshTokenRepository
	tokenService repository.TokenService
	auditRepo    repository.AuditRepository
	hasher       PasswordHasher
}

// NewLoginUserUseCase creates a new LoginUserUseCase.
func NewLoginUserUseCase(
	userRepo repository.UserRepository,
	tokenRepo repository.RefreshTokenRepository,
	tokenService repository.TokenService,
	auditRepo repository.AuditRepository,
	hasher PasswordHasher,
) *LoginUserUseCase {
	return &LoginUserUseCase{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		tokenService: tokenService,
		auditRepo:    auditRepo,
		hasher:       hasher,
	}
}

// LoginInput holds the input for user login.
type LoginInput struct {
	Email    string
	Password string
}

// LoginOutput holds the output of user login.
type LoginOutput struct {
	User   *entities.User       `json:"user"`
	Tokens *entities.AuthTokens `json:"tokens"`
}

// AuthenticationError represents invalid credentials.
type AuthenticationError struct {
	Message string
}

func (e *AuthenticationError) Error() string {
	return e.Message
}

// IsAuthenticationError checks whether the given error is an AuthenticationError.
func IsAuthenticationError(err error) bool {
	_, ok := err.(*AuthenticationError)
	return ok
}

// Execute authenticates a user and returns tokens.
func (uc *LoginUserUseCase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))

	if input.Email == "" {
		return nil, &ValidationError{Message: "email is required"}
	}
	if input.Password == "" {
		return nil, &ValidationError{Message: "password is required"}
	}

	user, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return nil, &AuthenticationError{Message: "invalid credentials"}
	}

	if err := uc.hasher.Compare(user.PasswordHash, input.Password); err != nil {
		_ = uc.auditRepo.LogEvent(ctx, user.ID, "auth.login_failed", "user", user.ID, map[string]interface{}{
			"reason": "invalid_password",
		})
		return nil, &AuthenticationError{Message: "invalid credentials"}
	}

	accessToken, err := uc.tokenService.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshTokenRaw, err := uc.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	refreshHash := uc.tokenService.HashToken(refreshTokenRaw)
	_, err = uc.tokenRepo.Create(ctx, user.ID, refreshHash, time.Now().Add(7*24*time.Hour))
	if err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	_ = uc.auditRepo.LogEvent(ctx, user.ID, "auth.login", "user", user.ID, nil)

	return &LoginOutput{
		User: user,
		Tokens: &entities.AuthTokens{
			AccessToken:  accessToken,
			RefreshToken: refreshTokenRaw,
		},
	}, nil
}
