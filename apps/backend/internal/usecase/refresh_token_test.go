package usecase

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/infrastructure/auth"
	"golang.org/x/crypto/bcrypt"
)

func newRefreshUseCase() (*RefreshTokenUseCase, *RegisterUserUseCase, *LoginUserUseCase, *repository.InMemoryAuditRepository) {
	userRepo := repository.NewInMemoryUserRepository()
	tokenRepo := repository.NewInMemoryRefreshTokenRepository()
	auditRepo := repository.NewInMemoryAuditRepository()
	tokenSvc := auth.NewJWTTokenService("test-secret", 900)
	hasher := auth.NewBcryptHasher(bcrypt.MinCost)

	registerUC := NewRegisterUserUseCase(userRepo, tokenRepo, tokenSvc, auditRepo, hasher)
	loginUC := NewLoginUserUseCase(userRepo, tokenRepo, tokenSvc, auditRepo, hasher)
	refreshUC := NewRefreshTokenUseCase(userRepo, tokenRepo, tokenSvc, auditRepo)
	return refreshUC, registerUC, loginUC, auditRepo
}

func TestRefreshToken_ValidToken_ReturnsNewTokens(t *testing.T) {
	refreshUC, registerUC, _, _ := newRefreshUseCase()
	ctx := context.Background()

	regOut, err := registerUC.Execute(ctx, RegisterInput{
		Email:    "user@example.com",
		Password: "securepassword",
		Name:     "Test User",
		Role:     "client",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	out, err := refreshUC.Execute(ctx, regOut.Tokens.RefreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if out.Tokens == nil {
		t.Fatal("expected tokens")
	}
	if out.Tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if out.Tokens.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	// New refresh token should be different from old
	if out.Tokens.RefreshToken == regOut.Tokens.RefreshToken {
		t.Error("expected rotated refresh token to be different")
	}
}

func TestRefreshToken_EmptyToken_ReturnsValidationError(t *testing.T) {
	refreshUC, _, _, _ := newRefreshUseCase()

	_, err := refreshUC.Execute(context.Background(), "")
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestRefreshToken_InvalidToken_ReturnsAuthError(t *testing.T) {
	refreshUC, _, _, _ := newRefreshUseCase()

	_, err := refreshUC.Execute(context.Background(), "non-existent-token")
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsAuthenticationError(err) {
		t.Errorf("expected AuthenticationError, got %T: %v", err, err)
	}
}

func TestRefreshToken_TokenReuse_RevokesAllTokens(t *testing.T) {
	refreshUC, registerUC, _, auditRepo := newRefreshUseCase()
	ctx := context.Background()

	regOut, err := registerUC.Execute(ctx, RegisterInput{
		Email:    "user@example.com",
		Password: "securepassword",
		Name:     "Test User",
		Role:     "client",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// First refresh should succeed
	_, err = refreshUC.Execute(ctx, regOut.Tokens.RefreshToken)
	if err != nil {
		t.Fatalf("first refresh failed: %v", err)
	}

	// Second use of same token should trigger replay detection
	_, err = refreshUC.Execute(ctx, regOut.Tokens.RefreshToken)
	if err == nil {
		t.Fatal("expected error for token reuse")
	}
	if !IsAuthenticationError(err) {
		t.Errorf("expected AuthenticationError, got %T: %v", err, err)
	}

	// Check audit log for replay detection
	found := false
	for _, e := range auditRepo.Events {
		if e.Action == "auth.token_replay_detected" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected auth.token_replay_detected audit event")
	}
}

func TestRefreshToken_RotatedToken_Works(t *testing.T) {
	refreshUC, registerUC, _, _ := newRefreshUseCase()
	ctx := context.Background()

	regOut, err := registerUC.Execute(ctx, RegisterInput{
		Email:    "user@example.com",
		Password: "securepassword",
		Name:     "Test User",
		Role:     "client",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// First refresh
	out1, err := refreshUC.Execute(ctx, regOut.Tokens.RefreshToken)
	if err != nil {
		t.Fatalf("first refresh failed: %v", err)
	}

	// Second refresh with the new rotated token
	out2, err := refreshUC.Execute(ctx, out1.Tokens.RefreshToken)
	if err != nil {
		t.Fatalf("second refresh failed: %v", err)
	}
	if out2.Tokens.AccessToken == "" {
		t.Error("expected non-empty access token from second refresh")
	}
}

func TestRefreshToken_AuditEventLogged(t *testing.T) {
	refreshUC, registerUC, _, auditRepo := newRefreshUseCase()
	ctx := context.Background()

	regOut, err := registerUC.Execute(ctx, RegisterInput{
		Email:    "user@example.com",
		Password: "securepassword",
		Name:     "Test User",
		Role:     "client",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	_, err = refreshUC.Execute(ctx, regOut.Tokens.RefreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}

	found := false
	for _, e := range auditRepo.Events {
		if e.Action == "auth.token_refreshed" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected auth.token_refreshed audit event")
	}
}
