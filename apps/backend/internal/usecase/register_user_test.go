package usecase

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/infrastructure/auth"
	"golang.org/x/crypto/bcrypt"
)

func newRegisterUseCase() (*RegisterUserUseCase, *repository.InMemoryUserRepository, *repository.InMemoryRefreshTokenRepository, *repository.InMemoryAuditRepository) {
	userRepo := repository.NewInMemoryUserRepository()
	tokenRepo := repository.NewInMemoryRefreshTokenRepository()
	auditRepo := repository.NewInMemoryAuditRepository()
	tokenSvc := auth.NewJWTTokenService("test-secret", 900)
	hasher := auth.NewBcryptHasher(bcrypt.MinCost)

	uc := NewRegisterUserUseCase(userRepo, tokenRepo, tokenSvc, auditRepo, hasher)
	return uc, userRepo, tokenRepo, auditRepo
}

func TestRegisterUser_ValidInput_CreatesUser(t *testing.T) {
	uc, _, _, _ := newRegisterUseCase()

	out, err := uc.Execute(context.Background(), RegisterInput{
		Email:    "test@example.com",
		Password: "securepassword",
		Name:     "Test User",
		Role:     "client",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.User == nil {
		t.Fatal("expected user to be created")
	}
	if out.User.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", out.User.Email)
	}
	if out.User.Name != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", out.User.Name)
	}
	if out.User.Role != "client" {
		t.Errorf("expected role 'client', got '%s'", out.User.Role)
	}
}

func TestRegisterUser_ValidInput_ReturnsTokens(t *testing.T) {
	uc, _, _, _ := newRegisterUseCase()

	out, err := uc.Execute(context.Background(), RegisterInput{
		Email:    "test@example.com",
		Password: "securepassword",
		Name:     "Test User",
		Role:     "client",
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
}

func TestRegisterUser_EmptyEmail_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newRegisterUseCase()

	_, err := uc.Execute(context.Background(), RegisterInput{
		Email:    "",
		Password: "securepassword",
		Name:     "Test",
		Role:     "client",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestRegisterUser_InvalidEmail_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newRegisterUseCase()

	_, err := uc.Execute(context.Background(), RegisterInput{
		Email:    "not-an-email",
		Password: "securepassword",
		Name:     "Test",
		Role:     "client",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestRegisterUser_ShortPassword_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newRegisterUseCase()

	_, err := uc.Execute(context.Background(), RegisterInput{
		Email:    "test@example.com",
		Password: "short",
		Name:     "Test",
		Role:     "client",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestRegisterUser_EmptyPassword_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newRegisterUseCase()

	_, err := uc.Execute(context.Background(), RegisterInput{
		Email:    "test@example.com",
		Password: "",
		Name:     "Test",
		Role:     "client",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestRegisterUser_EmptyName_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newRegisterUseCase()

	_, err := uc.Execute(context.Background(), RegisterInput{
		Email:    "test@example.com",
		Password: "securepassword",
		Name:     "",
		Role:     "client",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestRegisterUser_InvalidRole_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newRegisterUseCase()

	_, err := uc.Execute(context.Background(), RegisterInput{
		Email:    "test@example.com",
		Password: "securepassword",
		Name:     "Test",
		Role:     "superadmin",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestRegisterUser_DuplicateEmail_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newRegisterUseCase()
	ctx := context.Background()

	_, err := uc.Execute(ctx, RegisterInput{
		Email:    "dup@example.com",
		Password: "securepassword",
		Name:     "First",
		Role:     "client",
	})
	if err != nil {
		t.Fatalf("unexpected error on first register: %v", err)
	}

	_, err = uc.Execute(ctx, RegisterInput{
		Email:    "dup@example.com",
		Password: "anotherpassword",
		Name:     "Second",
		Role:     "client",
	})
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestRegisterUser_DefaultRole_SetsClient(t *testing.T) {
	uc, _, _, _ := newRegisterUseCase()

	out, err := uc.Execute(context.Background(), RegisterInput{
		Email:    "test@example.com",
		Password: "securepassword",
		Name:     "Test",
		Role:     "", // Empty should default to "client"
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.User.Role != "client" {
		t.Errorf("expected default role 'client', got '%s'", out.User.Role)
	}
}

func TestRegisterUser_EmailNormalized_Lowercase(t *testing.T) {
	uc, _, _, _ := newRegisterUseCase()

	out, err := uc.Execute(context.Background(), RegisterInput{
		Email:    "  Test@Example.COM  ",
		Password: "securepassword",
		Name:     "Test",
		Role:     "client",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.User.Email != "test@example.com" {
		t.Errorf("expected normalized email 'test@example.com', got '%s'", out.User.Email)
	}
}

func TestRegisterUser_AuditEventLogged(t *testing.T) {
	uc, _, _, auditRepo := newRegisterUseCase()

	_, err := uc.Execute(context.Background(), RegisterInput{
		Email:    "audit@example.com",
		Password: "securepassword",
		Name:     "Audit Test",
		Role:     "client",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(auditRepo.Events) == 0 {
		t.Fatal("expected audit event to be logged")
	}
	if auditRepo.Events[0].Action != "user.registered" {
		t.Errorf("expected action 'user.registered', got '%s'", auditRepo.Events[0].Action)
	}
}

func TestRegisterUser_CoachRole_Accepted(t *testing.T) {
	uc, _, _, _ := newRegisterUseCase()

	out, err := uc.Execute(context.Background(), RegisterInput{
		Email:    "coach@example.com",
		Password: "securepassword",
		Name:     "Coach Test",
		Role:     "coach",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.User.Role != "coach" {
		t.Errorf("expected role 'coach', got '%s'", out.User.Role)
	}
}
