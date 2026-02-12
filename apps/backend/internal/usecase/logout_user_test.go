package usecase

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/infrastructure/auth"
	"golang.org/x/crypto/bcrypt"
)

func newLogoutUseCase() (*LogoutUserUseCase, *RegisterUserUseCase, *RefreshTokenUseCase, *repository.InMemoryAuditRepository) {
	userRepo := repository.NewInMemoryUserRepository()
	tokenRepo := repository.NewInMemoryRefreshTokenRepository()
	auditRepo := repository.NewInMemoryAuditRepository()
	tokenSvc := auth.NewJWTTokenService("test-secret", 900)
	hasher := auth.NewBcryptHasher(bcrypt.MinCost)

	registerUC := NewRegisterUserUseCase(userRepo, tokenRepo, tokenSvc, auditRepo, hasher)
	refreshUC := NewRefreshTokenUseCase(userRepo, tokenRepo, tokenSvc, auditRepo)
	logoutUC := NewLogoutUserUseCase(tokenRepo, auditRepo)
	return logoutUC, registerUC, refreshUC, auditRepo
}

func TestLogoutUser_ValidUser_RevokesTokens(t *testing.T) {
	logoutUC, registerUC, refreshUC, _ := newLogoutUseCase()
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

	// Logout
	err = logoutUC.Execute(ctx, regOut.User.ID)
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	// Trying to refresh after logout should fail (tokens revoked)
	_, err = refreshUC.Execute(ctx, regOut.Tokens.RefreshToken)
	if err == nil {
		t.Fatal("expected error after logout")
	}
}

func TestLogoutUser_EmptyUserID_ReturnsValidationError(t *testing.T) {
	logoutUC, _, _, _ := newLogoutUseCase()

	err := logoutUC.Execute(context.Background(), "")
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestLogoutUser_AuditEventLogged(t *testing.T) {
	logoutUC, registerUC, _, auditRepo := newLogoutUseCase()
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

	err = logoutUC.Execute(ctx, regOut.User.ID)
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	found := false
	for _, e := range auditRepo.Events {
		if e.Action == "auth.logout" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected auth.logout audit event")
	}
}
