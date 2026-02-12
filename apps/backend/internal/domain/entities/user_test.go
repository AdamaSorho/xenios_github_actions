package entities

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUser_JSONSerialization_OmitsPasswordHash(t *testing.T) {
	user := User{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "secret_hash",
		CreatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("failed to marshal user: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if _, exists := result["password_hash"]; exists {
		t.Error("password_hash should not be included in JSON output")
	}

	if result["id"] != user.ID {
		t.Errorf("expected id %q, got %q", user.ID, result["id"])
	}

	if result["email"] != user.Email {
		t.Errorf("expected email %q, got %q", user.Email, result["email"])
	}

	if result["name"] != user.Name {
		t.Errorf("expected name %q, got %q", user.Name, result["name"])
	}
}

func TestUser_FieldValues(t *testing.T) {
	now := time.Now()
	user := User{
		ID:           "test-id",
		Email:        "user@example.com",
		Name:         "Jane Doe",
		PasswordHash: "hashed_password",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if user.ID != "test-id" {
		t.Errorf("expected ID 'test-id', got %q", user.ID)
	}
	if user.Email != "user@example.com" {
		t.Errorf("expected Email 'user@example.com', got %q", user.Email)
	}
	if user.Name != "Jane Doe" {
		t.Errorf("expected Name 'Jane Doe', got %q", user.Name)
	}
	if user.PasswordHash != "hashed_password" {
		t.Errorf("expected PasswordHash 'hashed_password', got %q", user.PasswordHash)
	}
	if !user.CreatedAt.Equal(now) {
		t.Errorf("expected CreatedAt %v, got %v", now, user.CreatedAt)
	}
	if !user.UpdatedAt.Equal(now) {
		t.Errorf("expected UpdatedAt %v, got %v", now, user.UpdatedAt)
	}
}
