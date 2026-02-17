package repository

import (
	"context"
	"testing"
)

func TestInMemoryUserRepository_Create_ReturnsUser(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user, err := repo.Create(context.Background(), "test@example.com", "hash123", "Test User", "client")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected user")
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", user.Email)
	}
	if user.PasswordHash != "hash123" {
		t.Errorf("expected password hash 'hash123', got '%s'", user.PasswordHash)
	}
	if user.Name != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", user.Name)
	}
	if user.Role != "client" {
		t.Errorf("expected role 'client', got '%s'", user.Role)
	}
	if user.ID == "" {
		t.Error("expected non-empty ID")
	}
	if user.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestInMemoryUserRepository_Create_UniqueIDs(t *testing.T) {
	repo := NewInMemoryUserRepository()

	u1, _ := repo.Create(context.Background(), "a@test.com", "h1", "A", "client")
	u2, _ := repo.Create(context.Background(), "b@test.com", "h2", "B", "client")

	if u1.ID == u2.ID {
		t.Errorf("expected unique IDs, got '%s' for both", u1.ID)
	}
}

func TestInMemoryUserRepository_FindByEmail_Exists(t *testing.T) {
	repo := NewInMemoryUserRepository()
	ctx := context.Background()

	_, _ = repo.Create(ctx, "test@example.com", "hash", "Test", "client")

	found, err := repo.FindByEmail(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil {
		t.Fatal("expected to find user")
	}
	if found.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", found.Email)
	}
}

func TestInMemoryUserRepository_FindByEmail_NotExists(t *testing.T) {
	repo := NewInMemoryUserRepository()

	found, err := repo.FindByEmail(context.Background(), "nobody@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Error("expected nil for non-existent email")
	}
}

func TestInMemoryUserRepository_FindByID_Exists(t *testing.T) {
	repo := NewInMemoryUserRepository()
	ctx := context.Background()

	created, _ := repo.Create(ctx, "test@example.com", "hash", "Test", "client")

	found, err := repo.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil {
		t.Fatal("expected to find user")
	}
	if found.ID != created.ID {
		t.Errorf("expected ID '%s', got '%s'", created.ID, found.ID)
	}
}

func TestInMemoryUserRepository_FindByID_NotExists(t *testing.T) {
	repo := NewInMemoryUserRepository()

	found, err := repo.FindByID(context.Background(), "nonexistent-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Error("expected nil for non-existent ID")
	}
}
