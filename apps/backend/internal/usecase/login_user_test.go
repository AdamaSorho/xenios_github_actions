package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/adapter/repository"
	"golang.org/x/crypto/bcrypt"
)

func newLoginUseCase() (*LoginUserUseCase, *RegisterUserUseCase, *repository.InMemoryAuditRepository) {
	userRepo := repository.NewInMemoryUserRepository()
	tokenRepo := repository.NewInMemoryRefreshTokenRepository()
	auditRepo := repository.NewInMemoryAuditRepository()
	tokenSvc := newStubTokenService("test-secret", 900*time.Second)
	hasher := newStubHasher(bcrypt.MinCost)

	registerUC := NewRegisterUserUseCase(userRepo, tokenRepo, tokenSvc, auditRepo, hasher)
	loginUC := NewLoginUserUseCase(userRepo, tokenRepo, tokenSvc, auditRepo, hasher)
	return loginUC, registerUC, auditRepo
}

func registerTestUser(t *testing.T, registerUC *RegisterUserUseCase) {
	t.Helper()
	_, err := registerUC.Execute(context.Background(), RegisterInput{
		Email:    "user@example.com",
		Password: "securepassword",
		Name:     "Test User",
		Role:     "client",
	})
	if err != nil {
		t.Fatalf("failed to register test user: %v", err)
	}
}

func TestLoginUser_ValidCredentials_ReturnsTokens(t *testing.T) {
	loginUC, registerUC, _ := newLoginUseCase()
	registerTestUser(t, registerUC)

	out, err := loginUC.Execute(context.Background(), LoginInput{
		Email:    "user@example.com",
		Password: "securepassword",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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
	if out.User == nil {
		t.Fatal("expected user")
	}
	if out.User.Email != "user@example.com" {
		t.Errorf("expected email 'user@example.com', got '%s'", out.User.Email)
	}
}

func TestLoginUser_WrongPassword_ReturnsAuthError(t *testing.T) {
	loginUC, registerUC, _ := newLoginUseCase()
	registerTestUser(t, registerUC)

	_, err := loginUC.Execute(context.Background(), LoginInput{
		Email:    "user@example.com",
		Password: "wrongpassword",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsAuthenticationError(err) {
		t.Errorf("expected AuthenticationError, got %T: %v", err, err)
	}
}

func TestLoginUser_NonExistentEmail_ReturnsAuthError(t *testing.T) {
	loginUC, _, _ := newLoginUseCase()

	_, err := loginUC.Execute(context.Background(), LoginInput{
		Email:    "nobody@example.com",
		Password: "somepassword",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsAuthenticationError(err) {
		t.Errorf("expected AuthenticationError, got %T: %v", err, err)
	}
}

func TestLoginUser_EmptyEmail_ReturnsValidationError(t *testing.T) {
	loginUC, _, _ := newLoginUseCase()

	_, err := loginUC.Execute(context.Background(), LoginInput{
		Email:    "",
		Password: "securepassword",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestLoginUser_EmptyPassword_ReturnsValidationError(t *testing.T) {
	loginUC, _, _ := newLoginUseCase()

	_, err := loginUC.Execute(context.Background(), LoginInput{
		Email:    "user@example.com",
		Password: "",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestLoginUser_CaseInsensitiveEmail(t *testing.T) {
	loginUC, registerUC, _ := newLoginUseCase()
	registerTestUser(t, registerUC)

	out, err := loginUC.Execute(context.Background(), LoginInput{
		Email:    "  USER@EXAMPLE.COM  ",
		Password: "securepassword",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.User == nil {
		t.Fatal("expected user")
	}
}

func TestLoginUser_FailedLogin_AuditEventLogged(t *testing.T) {
	loginUC, registerUC, auditRepo := newLoginUseCase()
	registerTestUser(t, registerUC)

	_, _ = loginUC.Execute(context.Background(), LoginInput{
		Email:    "user@example.com",
		Password: "wrongpassword",
	})

	found := false
	for _, e := range auditRepo.GetEvents() {
		if e.Action == "auth.login_failed" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected auth.login_failed audit event")
	}
}

func TestLoginUser_SuccessfulLogin_AuditEventLogged(t *testing.T) {
	loginUC, registerUC, auditRepo := newLoginUseCase()
	registerTestUser(t, registerUC)

	_, err := loginUC.Execute(context.Background(), LoginInput{
		Email:    "user@example.com",
		Password: "securepassword",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, e := range auditRepo.GetEvents() {
		if e.Action == "auth.login" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected auth.login audit event")
	}
}
