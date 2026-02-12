package entities

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewUser_ValidInputs_ReturnsUser(t *testing.T) {
	// Arrange
	email := "test@example.com"
	name := "Test User"
	passwordHash := "$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ01234"

	// Act
	user, err := NewUser(email, name, passwordHash)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != email {
		t.Errorf("expected email '%s', got '%s'", email, user.Email)
	}
	if user.Name != name {
		t.Errorf("expected name '%s', got '%s'", name, user.Name)
	}
	if user.PasswordHash != passwordHash {
		t.Errorf("expected password hash '%s', got '%s'", passwordHash, user.PasswordHash)
	}
	if user.ID == "" {
		t.Error("expected non-empty ID")
	}
	if user.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if user.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}

func TestNewUser_EmptyEmail_ReturnsError(t *testing.T) {
	// Arrange & Act
	_, err := NewUser("", "Test User", "$2a$10$validhash")

	// Assert
	if err == nil {
		t.Fatal("expected error for empty email")
	}
	if err != ErrEmptyEmail {
		t.Errorf("expected ErrEmptyEmail, got %v", err)
	}
}

func TestNewUser_EmptyName_ReturnsError(t *testing.T) {
	// Arrange & Act
	_, err := NewUser("test@example.com", "", "$2a$10$validhash")

	// Assert
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if err != ErrEmptyName {
		t.Errorf("expected ErrEmptyName, got %v", err)
	}
}

func TestNewUser_EmptyPasswordHash_ReturnsError(t *testing.T) {
	// Arrange & Act
	_, err := NewUser("test@example.com", "Test User", "")

	// Assert
	if err == nil {
		t.Fatal("expected error for empty password hash")
	}
	if err != ErrEmptyPasswordHash {
		t.Errorf("expected ErrEmptyPasswordHash, got %v", err)
	}
}

func TestNewUser_InvalidEmail_ReturnsError(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{"no at sign", "testexample.com"},
		{"no domain", "test@"},
		{"no local part", "@example.com"},
		{"spaces", "test @example.com"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewUser(tc.email, "Test User", "$2a$10$validhash")
			if err == nil {
				t.Errorf("expected error for invalid email '%s'", tc.email)
			}
			if err != ErrInvalidEmail {
				t.Errorf("expected ErrInvalidEmail, got %v", err)
			}
		})
	}
}

func TestNewUser_WhitespaceOnlyInputs_ReturnsError(t *testing.T) {
	// Whitespace-only email
	_, err := NewUser("   ", "Test User", "$2a$10$validhash")
	if err == nil {
		t.Error("expected error for whitespace-only email")
	}

	// Whitespace-only name
	_, err = NewUser("test@example.com", "   ", "$2a$10$validhash")
	if err == nil {
		t.Error("expected error for whitespace-only name")
	}
}

func TestUser_JSONSerialization(t *testing.T) {
	// Arrange
	now := time.Now().UTC().Truncate(time.Second)
	user := User{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ01234",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Act
	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("failed to marshal user: %v", err)
	}

	var decoded User
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal user: %v", err)
	}

	// Assert
	if decoded.ID != user.ID {
		t.Errorf("expected ID '%s', got '%s'", user.ID, decoded.ID)
	}
	if decoded.Email != user.Email {
		t.Errorf("expected email '%s', got '%s'", user.Email, decoded.Email)
	}
	if decoded.Name != user.Name {
		t.Errorf("expected name '%s', got '%s'", user.Name, decoded.Name)
	}
}

func TestUser_JSONFieldNames(t *testing.T) {
	// Arrange
	user := User{
		ID:           "test-id",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "$2a$10$hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Act
	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("failed to marshal user: %v", err)
	}

	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMap); err != nil {
		t.Fatalf("failed to unmarshal as raw map: %v", err)
	}

	// Assert - password_hash should never appear in JSON output
	if _, ok := rawMap["password_hash"]; ok {
		t.Error("password_hash should not be serialized to JSON")
	}

	expectedFields := []string{"id", "email", "name", "created_at", "updated_at"}
	for _, field := range expectedFields {
		if _, ok := rawMap[field]; !ok {
			t.Errorf("expected '%s' field in JSON", field)
		}
	}
}

func TestUser_PasswordHashNotInJSON(t *testing.T) {
	// Arrange
	user := User{
		ID:           "test-id",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "$2a$10$secrethash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Act
	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("failed to marshal user: %v", err)
	}

	jsonStr := string(data)

	// Assert - password hash value should not appear in JSON
	if containsString(jsonStr, "$2a$10$secrethash") {
		t.Error("password hash value must not appear in JSON output")
	}
}

func TestUser_ZeroValue(t *testing.T) {
	// Arrange & Act
	var user User

	// Assert
	if user.ID != "" {
		t.Errorf("expected empty ID for zero value, got '%s'", user.ID)
	}
	if user.Email != "" {
		t.Errorf("expected empty email for zero value, got '%s'", user.Email)
	}
	if user.Name != "" {
		t.Errorf("expected empty name for zero value, got '%s'", user.Name)
	}
	if user.PasswordHash != "" {
		t.Errorf("expected empty password hash for zero value, got '%s'", user.PasswordHash)
	}
	if !user.CreatedAt.IsZero() {
		t.Error("expected zero CreatedAt for zero value")
	}
	if !user.UpdatedAt.IsZero() {
		t.Error("expected zero UpdatedAt for zero value")
	}
}

// containsString is a helper to check if a string contains a substring.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
