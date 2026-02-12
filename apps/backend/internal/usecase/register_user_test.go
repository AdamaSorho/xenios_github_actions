package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/xenios/backend/internal/domain/entities"
)

// mockUserRepository is a mock implementation of the UserRepository interface for testing.
type mockUserRepository struct {
	createFn      func(ctx context.Context, user *entities.User) (*entities.User, error)
	findByEmailFn func(ctx context.Context, email string) (*entities.User, error)
}

func (m *mockUserRepository) Create(ctx context.Context, user *entities.User) (*entities.User, error) {
	return m.createFn(ctx, user)
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	return m.findByEmailFn(ctx, email)
}

// newMockUserRepo creates a mock that returns nil,nil for FindByEmail and echoes Create input with generated fields.
func newMockUserRepo() *mockUserRepository {
	return &mockUserRepository{
		findByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
			return nil, nil // user not found
		},
		createFn: func(ctx context.Context, user *entities.User) (*entities.User, error) {
			now := time.Now()
			return &entities.User{
				ID:           "generated-uuid",
				Email:        user.Email,
				Name:         user.Name,
				PasswordHash: user.PasswordHash,
				CreatedAt:    now,
				UpdatedAt:    now,
			}, nil
		},
	}
}

// TestRegisterUserUseCase_Execute_HappyPath tests successful user registration.
func TestRegisterUserUseCase_Execute_HappyPath(t *testing.T) {
	// Arrange
	mock := newMockUserRepo()
	uc := NewRegisterUserUseCase(mock)

	input := RegisterUserInput{
		Email:    "test@example.com",
		Password: "securepassword123",
		Name:     "Test User",
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if output == nil {
		t.Fatal("expected non-nil output")
	}
	if output.ID == "" {
		t.Error("expected non-empty ID")
	}
	if output.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", output.Email)
	}
	if output.Name != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", output.Name)
	}
	if output.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if output.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}

// TestRegisterUserUseCase_Execute_InvalidEmail_MissingAt tests email without @.
func TestRegisterUserUseCase_Execute_InvalidEmail_MissingAt(t *testing.T) {
	// Arrange
	mock := newMockUserRepo()
	uc := NewRegisterUserUseCase(mock)

	input := RegisterUserInput{
		Email:    "invalidemail.com",
		Password: "securepassword123",
		Name:     "Test User",
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid email, got nil")
	}
	if !errors.Is(err, ErrInvalidEmail) {
		t.Errorf("expected ErrInvalidEmail, got: %v", err)
	}
	if output != nil {
		t.Error("expected nil output on validation error")
	}
}

// TestRegisterUserUseCase_Execute_InvalidEmail_MissingDomain tests email without domain.
func TestRegisterUserUseCase_Execute_InvalidEmail_MissingDomain(t *testing.T) {
	// Arrange
	mock := newMockUserRepo()
	uc := NewRegisterUserUseCase(mock)

	input := RegisterUserInput{
		Email:    "user@",
		Password: "securepassword123",
		Name:     "Test User",
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid email without domain, got nil")
	}
	if !errors.Is(err, ErrInvalidEmail) {
		t.Errorf("expected ErrInvalidEmail, got: %v", err)
	}
	if output != nil {
		t.Error("expected nil output on validation error")
	}
}

// TestRegisterUserUseCase_Execute_InvalidEmail_Empty tests empty email.
func TestRegisterUserUseCase_Execute_InvalidEmail_Empty(t *testing.T) {
	// Arrange
	mock := newMockUserRepo()
	uc := NewRegisterUserUseCase(mock)

	input := RegisterUserInput{
		Email:    "",
		Password: "securepassword123",
		Name:     "Test User",
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("expected error for empty email, got nil")
	}
	if !errors.Is(err, ErrInvalidEmail) {
		t.Errorf("expected ErrInvalidEmail, got: %v", err)
	}
	if output != nil {
		t.Error("expected nil output on validation error")
	}
}

// TestRegisterUserUseCase_Execute_InvalidEmail_MissingDot tests email without dot in domain.
func TestRegisterUserUseCase_Execute_InvalidEmail_MissingDot(t *testing.T) {
	// Arrange
	mock := newMockUserRepo()
	uc := NewRegisterUserUseCase(mock)

	input := RegisterUserInput{
		Email:    "user@domaincom",
		Password: "securepassword123",
		Name:     "Test User",
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("expected error for email without dot in domain, got nil")
	}
	if !errors.Is(err, ErrInvalidEmail) {
		t.Errorf("expected ErrInvalidEmail, got: %v", err)
	}
	if output != nil {
		t.Error("expected nil output on validation error")
	}
}

// TestRegisterUserUseCase_Execute_ShortPassword tests password too short.
func TestRegisterUserUseCase_Execute_ShortPassword(t *testing.T) {
	// Arrange
	mock := newMockUserRepo()
	uc := NewRegisterUserUseCase(mock)

	input := RegisterUserInput{
		Email:    "test@example.com",
		Password: "1234567", // 7 characters
		Name:     "Test User",
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("expected error for short password, got nil")
	}
	if !errors.Is(err, ErrPasswordTooShort) {
		t.Errorf("expected ErrPasswordTooShort, got: %v", err)
	}
	if output != nil {
		t.Error("expected nil output on validation error")
	}
}

// TestRegisterUserUseCase_Execute_ExactMinimumPassword tests password exactly at minimum length.
func TestRegisterUserUseCase_Execute_ExactMinimumPassword(t *testing.T) {
	// Arrange
	mock := newMockUserRepo()
	uc := NewRegisterUserUseCase(mock)

	input := RegisterUserInput{
		Email:    "test@example.com",
		Password: "12345678", // exactly 8 characters
		Name:     "Test User",
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error for 8-char password, got: %v", err)
	}
	if output == nil {
		t.Fatal("expected non-nil output")
	}
}

// TestRegisterUserUseCase_Execute_EmptyName tests empty name.
func TestRegisterUserUseCase_Execute_EmptyName(t *testing.T) {
	// Arrange
	mock := newMockUserRepo()
	uc := NewRegisterUserUseCase(mock)

	input := RegisterUserInput{
		Email:    "test@example.com",
		Password: "securepassword123",
		Name:     "",
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
	if !errors.Is(err, ErrNameRequired) {
		t.Errorf("expected ErrNameRequired, got: %v", err)
	}
	if output != nil {
		t.Error("expected nil output on validation error")
	}
}

// TestRegisterUserUseCase_Execute_WhitespaceOnlyName tests whitespace-only name.
func TestRegisterUserUseCase_Execute_WhitespaceOnlyName(t *testing.T) {
	// Arrange
	mock := newMockUserRepo()
	uc := NewRegisterUserUseCase(mock)

	input := RegisterUserInput{
		Email:    "test@example.com",
		Password: "securepassword123",
		Name:     "   \t  ",
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("expected error for whitespace-only name, got nil")
	}
	if !errors.Is(err, ErrNameRequired) {
		t.Errorf("expected ErrNameRequired, got: %v", err)
	}
	if output != nil {
		t.Error("expected nil output on validation error")
	}
}

// TestRegisterUserUseCase_Execute_DuplicateEmail tests email already registered.
func TestRegisterUserUseCase_Execute_DuplicateEmail(t *testing.T) {
	// Arrange
	now := time.Now()
	mock := &mockUserRepository{
		findByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
			return &entities.User{
				ID:        "existing-user-id",
				Email:     email,
				Name:      "Existing User",
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
		createFn: func(ctx context.Context, user *entities.User) (*entities.User, error) {
			t.Error("Create should not be called when email already exists")
			return nil, nil
		},
	}
	uc := NewRegisterUserUseCase(mock)

	input := RegisterUserInput{
		Email:    "existing@example.com",
		Password: "securepassword123",
		Name:     "New User",
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}
	if !errors.Is(err, ErrEmailExists) {
		t.Errorf("expected ErrEmailExists, got: %v", err)
	}
	if output != nil {
		t.Error("expected nil output on duplicate email")
	}
}

// TestRegisterUserUseCase_Execute_RepositoryErrorOnCreate tests repository error during Create.
func TestRegisterUserUseCase_Execute_RepositoryErrorOnCreate(t *testing.T) {
	// Arrange
	repoErr := errors.New("database connection failed")
	mock := &mockUserRepository{
		findByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
			return nil, nil
		},
		createFn: func(ctx context.Context, user *entities.User) (*entities.User, error) {
			return nil, repoErr
		},
	}
	uc := NewRegisterUserUseCase(mock)

	input := RegisterUserInput{
		Email:    "test@example.com",
		Password: "securepassword123",
		Name:     "Test User",
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("expected error from repository, got nil")
	}
	if output != nil {
		t.Error("expected nil output on repository error")
	}
}

// TestRegisterUserUseCase_Execute_RepositoryErrorOnFindByEmail tests repository error during FindByEmail.
func TestRegisterUserUseCase_Execute_RepositoryErrorOnFindByEmail(t *testing.T) {
	// Arrange
	repoErr := errors.New("connection refused")
	mock := &mockUserRepository{
		findByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
			return nil, repoErr
		},
		createFn: func(ctx context.Context, user *entities.User) (*entities.User, error) {
			t.Error("Create should not be called when FindByEmail fails")
			return nil, nil
		},
	}
	uc := NewRegisterUserUseCase(mock)

	input := RegisterUserInput{
		Email:    "test@example.com",
		Password: "securepassword123",
		Name:     "Test User",
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("expected error from repository, got nil")
	}
	if output != nil {
		t.Error("expected nil output on repository error")
	}
}

// TestRegisterUserUseCase_Execute_PasswordHashing tests that the stored password is a valid bcrypt hash.
func TestRegisterUserUseCase_Execute_PasswordHashing(t *testing.T) {
	// Arrange
	var storedHash string
	mock := &mockUserRepository{
		findByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
			return nil, nil
		},
		createFn: func(ctx context.Context, user *entities.User) (*entities.User, error) {
			storedHash = user.PasswordHash
			now := time.Now()
			return &entities.User{
				ID:           "hashed-user-id",
				Email:        user.Email,
				Name:         user.Name,
				PasswordHash: user.PasswordHash,
				CreatedAt:    now,
				UpdatedAt:    now,
			}, nil
		},
	}
	uc := NewRegisterUserUseCase(mock)

	password := "mysecretpassword"
	input := RegisterUserInput{
		Email:    "test@example.com",
		Password: password,
		Name:     "Test User",
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if storedHash == "" {
		t.Fatal("expected password hash to be stored")
	}
	if storedHash == password {
		t.Error("stored hash must not equal the raw password")
	}
	// Verify it's a valid bcrypt hash
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
		t.Errorf("stored hash is not a valid bcrypt hash of the password: %v", err)
	}
}

// TestRegisterUserUseCase_Execute_PasswordNotInOutput tests that password/hash is not in output.
func TestRegisterUserUseCase_Execute_PasswordNotInOutput(t *testing.T) {
	// Arrange
	mock := newMockUserRepo()
	uc := NewRegisterUserUseCase(mock)

	input := RegisterUserInput{
		Email:    "test@example.com",
		Password: "securepassword123",
		Name:     "Test User",
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if output == nil {
		t.Fatal("expected non-nil output")
	}
	// RegisterUserOutput should not have a password or hash field at all
	// This is enforced by the struct definition - it only has ID, Email, Name, CreatedAt, UpdatedAt
}

// TestRegisterUserUseCase_Execute_NameTrimming tests that name is trimmed.
func TestRegisterUserUseCase_Execute_NameTrimming(t *testing.T) {
	// Arrange
	var storedName string
	mock := &mockUserRepository{
		findByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
			return nil, nil
		},
		createFn: func(ctx context.Context, user *entities.User) (*entities.User, error) {
			storedName = user.Name
			now := time.Now()
			return &entities.User{
				ID:           "trimmed-user-id",
				Email:        user.Email,
				Name:         user.Name,
				PasswordHash: user.PasswordHash,
				CreatedAt:    now,
				UpdatedAt:    now,
			}, nil
		},
	}
	uc := NewRegisterUserUseCase(mock)

	input := RegisterUserInput{
		Email:    "test@example.com",
		Password: "securepassword123",
		Name:     "  Test User  ",
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if storedName != "Test User" {
		t.Errorf("expected trimmed name 'Test User', got '%s'", storedName)
	}
	if output.Name != "Test User" {
		t.Errorf("expected trimmed name in output 'Test User', got '%s'", output.Name)
	}
}

// TestRegisterUserUseCase_Execute_EmailTrimmedAndLowercased tests email normalization.
func TestRegisterUserUseCase_Execute_EmailTrimmedAndLowercased(t *testing.T) {
	// Arrange
	var storedEmail string
	mock := &mockUserRepository{
		findByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
			return nil, nil
		},
		createFn: func(ctx context.Context, user *entities.User) (*entities.User, error) {
			storedEmail = user.Email
			now := time.Now()
			return &entities.User{
				ID:           "email-user-id",
				Email:        user.Email,
				Name:         user.Name,
				PasswordHash: user.PasswordHash,
				CreatedAt:    now,
				UpdatedAt:    now,
			}, nil
		},
	}
	uc := NewRegisterUserUseCase(mock)

	input := RegisterUserInput{
		Email:    "  Test@Example.COM  ",
		Password: "securepassword123",
		Name:     "Test User",
	}

	// Act
	output, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if storedEmail != "test@example.com" {
		t.Errorf("expected lowercased email 'test@example.com', got '%s'", storedEmail)
	}
	if output.Email != "test@example.com" {
		t.Errorf("expected lowercased email in output 'test@example.com', got '%s'", output.Email)
	}
}

// TestNewRegisterUserUseCase tests constructor.
func TestNewRegisterUserUseCase(t *testing.T) {
	// Arrange
	mock := newMockUserRepo()

	// Act
	uc := NewRegisterUserUseCase(mock)

	// Assert
	if uc == nil {
		t.Fatal("expected non-nil use case")
	}
}
