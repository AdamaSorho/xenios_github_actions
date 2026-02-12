package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// mockUserRepo implements repository.UserRepository for testing.
type mockUserRepo struct {
	createFn      func(ctx context.Context, user *entities.User) (*entities.User, error)
	findByEmailFn func(ctx context.Context, email string) (*entities.User, error)
	findByIDFn    func(ctx context.Context, id string) (*entities.User, error)
}

func (m *mockUserRepo) Create(ctx context.Context, user *entities.User) (*entities.User, error) {
	return m.createFn(ctx, user)
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	return m.findByEmailFn(ctx, email)
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*entities.User, error) {
	return m.findByIDFn(ctx, id)
}

// Verify mockUserRepo implements UserRepository interface at compile time.
var _ repository.UserRepository = (*mockUserRepo)(nil)

func TestNewCreateUserUseCase_ValidDependencies_ReturnsUseCase(t *testing.T) {
	repo := &mockUserRepo{}
	uc := NewCreateUserUseCase(repo)

	if uc == nil {
		t.Fatal("expected non-nil use case")
	}
}

func TestCreateUser_ValidInput_ReturnsUser(t *testing.T) {
	repo := &mockUserRepo{
		findByEmailFn: func(_ context.Context, _ string) (*entities.User, error) {
			return nil, nil // no existing user
		},
		createFn: func(_ context.Context, user *entities.User) (*entities.User, error) {
			return user, nil
		},
	}

	uc := NewCreateUserUseCase(repo)
	input := CreateUserInput{
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "$2a$10$validbcrypthash",
	}

	user, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != input.Email {
		t.Errorf("expected email '%s', got '%s'", input.Email, user.Email)
	}
	if user.Name != input.Name {
		t.Errorf("expected name '%s', got '%s'", input.Name, user.Name)
	}
	if user.PasswordHash != input.PasswordHash {
		t.Errorf("expected password hash '%s', got '%s'", input.PasswordHash, user.PasswordHash)
	}
}

func TestCreateUser_DuplicateEmail_ReturnsError(t *testing.T) {
	existingUser := &entities.User{
		ID:    "existing-id",
		Email: "test@example.com",
		Name:  "Existing User",
	}
	repo := &mockUserRepo{
		findByEmailFn: func(_ context.Context, _ string) (*entities.User, error) {
			return existingUser, nil // user already exists
		},
	}

	uc := NewCreateUserUseCase(repo)
	input := CreateUserInput{
		Email:        "test@example.com",
		Name:         "New User",
		PasswordHash: "$2a$10$validbcrypthash",
	}

	_, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
	if !errors.Is(err, ErrDuplicateEmail) {
		t.Errorf("expected ErrDuplicateEmail, got %v", err)
	}
}

func TestCreateUser_EmptyEmail_ReturnsError(t *testing.T) {
	repo := &mockUserRepo{}
	uc := NewCreateUserUseCase(repo)

	input := CreateUserInput{
		Email:        "",
		Name:         "Test User",
		PasswordHash: "$2a$10$validhash",
	}

	_, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for empty email")
	}
}

func TestCreateUser_EmptyName_ReturnsError(t *testing.T) {
	repo := &mockUserRepo{}
	uc := NewCreateUserUseCase(repo)

	input := CreateUserInput{
		Email:        "test@example.com",
		Name:         "",
		PasswordHash: "$2a$10$validhash",
	}

	_, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestCreateUser_EmptyPasswordHash_ReturnsError(t *testing.T) {
	repo := &mockUserRepo{}
	uc := NewCreateUserUseCase(repo)

	input := CreateUserInput{
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "",
	}

	_, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for empty password hash")
	}
}

func TestCreateUser_RepositoryCreateError_PropagatesError(t *testing.T) {
	repoErr := errors.New("database connection failed")
	repo := &mockUserRepo{
		findByEmailFn: func(_ context.Context, _ string) (*entities.User, error) {
			return nil, nil
		},
		createFn: func(_ context.Context, _ *entities.User) (*entities.User, error) {
			return nil, repoErr
		},
	}

	uc := NewCreateUserUseCase(repo)
	input := CreateUserInput{
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "$2a$10$validhash",
	}

	_, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error from repository")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repository error to propagate, got %v", err)
	}
}

func TestCreateUser_RepositoryFindByEmailError_PropagatesError(t *testing.T) {
	repoErr := errors.New("database timeout")
	repo := &mockUserRepo{
		findByEmailFn: func(_ context.Context, _ string) (*entities.User, error) {
			return nil, repoErr
		},
	}

	uc := NewCreateUserUseCase(repo)
	input := CreateUserInput{
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "$2a$10$validhash",
	}

	_, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error from repository")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repository error to propagate, got %v", err)
	}
}

func TestCreateUser_CancelledContext_PropagatesError(t *testing.T) {
	repo := &mockUserRepo{
		findByEmailFn: func(ctx context.Context, _ string) (*entities.User, error) {
			return nil, ctx.Err()
		},
	}

	uc := NewCreateUserUseCase(repo)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	input := CreateUserInput{
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "$2a$10$validhash",
	}

	_, err := uc.Execute(ctx, input)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestCreateUser_InvalidEmail_ReturnsError(t *testing.T) {
	repo := &mockUserRepo{}
	uc := NewCreateUserUseCase(repo)

	input := CreateUserInput{
		Email:        "invalid-email",
		Name:         "Test User",
		PasswordHash: "$2a$10$validhash",
	}

	_, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for invalid email")
	}
}
