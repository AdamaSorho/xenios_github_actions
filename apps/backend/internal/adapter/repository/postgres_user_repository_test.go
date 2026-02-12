package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// mockRow implements pgx.Row for testing.
type mockRow struct {
	scanFunc func(dest ...interface{}) error
}

func (m *mockRow) Scan(dest ...interface{}) error {
	return m.scanFunc(dest...)
}

// mockQuerier implements the Querier interface for testing.
type mockQuerier struct {
	queryRowFunc func(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

func (m *mockQuerier) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return m.queryRowFunc(ctx, sql, args...)
}

// --- Create Tests ---

func TestPostgresUserRepository_Create_Success(t *testing.T) {
	// Arrange
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	expectedID := "550e8400-e29b-41d4-a716-446655440000"

	querier := &mockQuerier{
		queryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return &mockRow{
				scanFunc: func(dest ...interface{}) error {
					if len(dest) != 6 {
						t.Fatalf("expected 6 scan destinations, got %d", len(dest))
					}
					*dest[0].(*string) = expectedID
					*dest[1].(*string) = "test@example.com"
					*dest[2].(*string) = "Test User"
					*dest[3].(*string) = "hashed_password"
					*dest[4].(*time.Time) = now
					*dest[5].(*time.Time) = now
					return nil
				},
			}
		},
	}

	repo := NewPostgresUserRepository(querier)

	input := &entities.User{
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "hashed_password",
	}

	// Act
	result, err := repo.Create(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	if result.ID != expectedID {
		t.Errorf("expected ID %q, got %q", expectedID, result.ID)
	}
	if result.Email != "test@example.com" {
		t.Errorf("expected Email 'test@example.com', got %q", result.Email)
	}
	if result.Name != "Test User" {
		t.Errorf("expected Name 'Test User', got %q", result.Name)
	}
	if result.PasswordHash != "hashed_password" {
		t.Errorf("expected PasswordHash 'hashed_password', got %q", result.PasswordHash)
	}
	if !result.CreatedAt.Equal(now) {
		t.Errorf("expected CreatedAt %v, got %v", now, result.CreatedAt)
	}
	if !result.UpdatedAt.Equal(now) {
		t.Errorf("expected UpdatedAt %v, got %v", now, result.UpdatedAt)
	}
}

func TestPostgresUserRepository_Create_DuplicateEmail(t *testing.T) {
	// Arrange
	querier := &mockQuerier{
		queryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return &mockRow{
				scanFunc: func(dest ...interface{}) error {
					// Simulate PostgreSQL unique constraint violation (error code 23505)
					return &pgconn.PgError{
						Code:    "23505",
						Message: "duplicate key value violates unique constraint \"users_email_key\"",
					}
				},
			}
		},
	}

	repo := NewPostgresUserRepository(querier)

	input := &entities.User{
		Email:        "existing@example.com",
		Name:         "Test User",
		PasswordHash: "hashed_password",
	}

	// Act
	result, err := repo.Create(context.Background(), input)

	// Assert
	if result != nil {
		t.Errorf("expected nil result, got: %+v", result)
	}
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, repository.ErrDuplicateEmail) {
		t.Errorf("expected ErrDuplicateEmail, got: %v", err)
	}
}

func TestPostgresUserRepository_Create_DatabaseError(t *testing.T) {
	// Arrange
	dbErr := errors.New("connection refused")
	querier := &mockQuerier{
		queryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return &mockRow{
				scanFunc: func(dest ...interface{}) error {
					return dbErr
				},
			}
		},
	}

	repo := NewPostgresUserRepository(querier)

	input := &entities.User{
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "hashed_password",
	}

	// Act
	result, err := repo.Create(context.Background(), input)

	// Assert
	if result != nil {
		t.Errorf("expected nil result, got: %+v", result)
	}
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if errors.Is(err, repository.ErrDuplicateEmail) {
		t.Error("should not be ErrDuplicateEmail for generic database errors")
	}
}

func TestPostgresUserRepository_Create_ContextCancelled(t *testing.T) {
	// Arrange
	querier := &mockQuerier{
		queryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return &mockRow{
				scanFunc: func(dest ...interface{}) error {
					return context.Canceled
				},
			}
		},
	}

	repo := NewPostgresUserRepository(querier)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	input := &entities.User{
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "hashed_password",
	}

	// Act
	result, err := repo.Create(ctx, input)

	// Assert
	if result != nil {
		t.Errorf("expected nil result, got: %+v", result)
	}
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

// --- FindByEmail Tests ---

func TestPostgresUserRepository_FindByEmail_Found(t *testing.T) {
	// Arrange
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	expectedID := "550e8400-e29b-41d4-a716-446655440000"

	querier := &mockQuerier{
		queryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return &mockRow{
				scanFunc: func(dest ...interface{}) error {
					if len(dest) != 6 {
						t.Fatalf("expected 6 scan destinations, got %d", len(dest))
					}
					*dest[0].(*string) = expectedID
					*dest[1].(*string) = "test@example.com"
					*dest[2].(*string) = "Test User"
					*dest[3].(*string) = "hashed_password"
					*dest[4].(*time.Time) = now
					*dest[5].(*time.Time) = now
					return nil
				},
			}
		},
	}

	repo := NewPostgresUserRepository(querier)

	// Act
	result, err := repo.FindByEmail(context.Background(), "test@example.com")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	if result.ID != expectedID {
		t.Errorf("expected ID %q, got %q", expectedID, result.ID)
	}
	if result.Email != "test@example.com" {
		t.Errorf("expected Email 'test@example.com', got %q", result.Email)
	}
	if result.Name != "Test User" {
		t.Errorf("expected Name 'Test User', got %q", result.Name)
	}
	if result.PasswordHash != "hashed_password" {
		t.Errorf("expected PasswordHash 'hashed_password', got %q", result.PasswordHash)
	}
}

func TestPostgresUserRepository_FindByEmail_NotFound(t *testing.T) {
	// Arrange
	querier := &mockQuerier{
		queryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return &mockRow{
				scanFunc: func(dest ...interface{}) error {
					return pgx.ErrNoRows
				},
			}
		},
	}

	repo := NewPostgresUserRepository(querier)

	// Act
	result, err := repo.FindByEmail(context.Background(), "nonexistent@example.com")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for not found, got: %+v", result)
	}
}

func TestPostgresUserRepository_FindByEmail_DatabaseError(t *testing.T) {
	// Arrange
	dbErr := errors.New("connection refused")
	querier := &mockQuerier{
		queryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return &mockRow{
				scanFunc: func(dest ...interface{}) error {
					return dbErr
				},
			}
		},
	}

	repo := NewPostgresUserRepository(querier)

	// Act
	result, err := repo.FindByEmail(context.Background(), "test@example.com")

	// Assert
	if result != nil {
		t.Errorf("expected nil result, got: %+v", result)
	}
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestPostgresUserRepository_FindByEmail_ContextCancelled(t *testing.T) {
	// Arrange
	querier := &mockQuerier{
		queryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			return &mockRow{
				scanFunc: func(dest ...interface{}) error {
					return context.Canceled
				},
			}
		},
	}

	repo := NewPostgresUserRepository(querier)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Act
	result, err := repo.FindByEmail(ctx, "test@example.com")

	// Assert
	if result != nil {
		t.Errorf("expected nil result, got: %+v", result)
	}
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

// --- Interface Compliance Test ---

func TestPostgresUserRepository_ImplementsUserRepository(t *testing.T) {
	querier := &mockQuerier{}
	repo := NewPostgresUserRepository(querier)

	// Compile-time check that PostgresUserRepository implements UserRepository
	var _ repository.UserRepository = repo
}

// --- SQL Query Verification ---

func TestPostgresUserRepository_Create_UsesParameterizedQuery(t *testing.T) {
	// Arrange
	var capturedSQL string
	var capturedArgs []interface{}

	querier := &mockQuerier{
		queryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			capturedSQL = sql
			capturedArgs = args
			return &mockRow{
				scanFunc: func(dest ...interface{}) error {
					now := time.Now()
					*dest[0].(*string) = "test-id"
					*dest[1].(*string) = "test@example.com"
					*dest[2].(*string) = "Test"
					*dest[3].(*string) = "hash"
					*dest[4].(*time.Time) = now
					*dest[5].(*time.Time) = now
					return nil
				},
			}
		},
	}

	repo := NewPostgresUserRepository(querier)
	input := &entities.User{
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "hashed",
	}

	// Act
	_, _ = repo.Create(context.Background(), input)

	// Assert - SQL uses parameterized queries
	if capturedSQL == "" {
		t.Fatal("expected SQL to be captured")
	}
	if len(capturedArgs) != 3 {
		t.Errorf("expected 3 query arguments (email, name, password_hash), got %d", len(capturedArgs))
	}
	if capturedArgs[0] != "test@example.com" {
		t.Errorf("expected first arg to be email, got %v", capturedArgs[0])
	}
	if capturedArgs[1] != "Test User" {
		t.Errorf("expected second arg to be name, got %v", capturedArgs[1])
	}
	if capturedArgs[2] != "hashed" {
		t.Errorf("expected third arg to be password_hash, got %v", capturedArgs[2])
	}
}

func TestPostgresUserRepository_FindByEmail_UsesParameterizedQuery(t *testing.T) {
	// Arrange
	var capturedSQL string
	var capturedArgs []interface{}

	querier := &mockQuerier{
		queryRowFunc: func(ctx context.Context, sql string, args ...interface{}) pgx.Row {
			capturedSQL = sql
			capturedArgs = args
			return &mockRow{
				scanFunc: func(dest ...interface{}) error {
					return pgx.ErrNoRows
				},
			}
		},
	}

	repo := NewPostgresUserRepository(querier)

	// Act
	_, _ = repo.FindByEmail(context.Background(), "find@example.com")

	// Assert
	if capturedSQL == "" {
		t.Fatal("expected SQL to be captured")
	}
	if len(capturedArgs) != 1 {
		t.Errorf("expected 1 query argument (email), got %d", len(capturedArgs))
	}
	if capturedArgs[0] != "find@example.com" {
		t.Errorf("expected first arg to be email, got %v", capturedArgs[0])
	}
}
