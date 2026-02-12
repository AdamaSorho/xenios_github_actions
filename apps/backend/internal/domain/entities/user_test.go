package entities

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUser_Creation_AllFieldsPopulated(t *testing.T) {
	// Arrange
	now := time.Now().UTC()

	// Act
	user := User{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "$2a$10$abcdefghijklmnopqrstuv",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Assert
	if user.ID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected ID '550e8400-e29b-41d4-a716-446655440000', got '%s'", user.ID)
	}

	if user.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", user.Email)
	}

	if user.Name != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", user.Name)
	}

	if user.PasswordHash != "$2a$10$abcdefghijklmnopqrstuv" {
		t.Errorf("expected password hash to be set, got '%s'", user.PasswordHash)
	}

	if !user.CreatedAt.Equal(now) {
		t.Errorf("expected created_at '%v', got '%v'", now, user.CreatedAt)
	}

	if !user.UpdatedAt.Equal(now) {
		t.Errorf("expected updated_at '%v', got '%v'", now, user.UpdatedAt)
	}
}

func TestUser_JSONSerialization_PasswordHashExcluded(t *testing.T) {
	// Arrange
	user := User{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "secret_hash_should_not_appear",
		CreatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	// Act
	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("failed to marshal user: %v", err)
	}

	jsonStr := string(data)

	// Assert - PasswordHash must NOT appear in JSON output
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMap); err != nil {
		t.Fatalf("failed to unmarshal as raw map: %v", err)
	}

	if _, ok := rawMap["password_hash"]; ok {
		t.Errorf("password_hash should NOT be in JSON output, got: %s", jsonStr)
	}

	if _, ok := rawMap["PasswordHash"]; ok {
		t.Errorf("PasswordHash should NOT be in JSON output, got: %s", jsonStr)
	}

	// Verify the hash value doesn't appear anywhere in the JSON string
	if contains(jsonStr, "secret_hash_should_not_appear") {
		t.Errorf("password hash value found in JSON output: %s", jsonStr)
	}
}

func TestUser_JSONSerialization_RoundTrip(t *testing.T) {
	// Arrange
	user := User{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "some_hash",
		CreatedAt:    time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC),
	}

	// Act
	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("failed to marshal user: %v", err)
	}

	var decoded User
	err = json.Unmarshal(data, &decoded)
	if err != nil {
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

	// PasswordHash should be empty after round-trip (excluded from JSON)
	if decoded.PasswordHash != "" {
		t.Errorf("expected empty password hash after JSON round-trip, got '%s'", decoded.PasswordHash)
	}

	if !decoded.CreatedAt.Equal(user.CreatedAt) {
		t.Errorf("expected created_at '%v', got '%v'", user.CreatedAt, decoded.CreatedAt)
	}

	if !decoded.UpdatedAt.Equal(user.UpdatedAt) {
		t.Errorf("expected updated_at '%v', got '%v'", user.UpdatedAt, decoded.UpdatedAt)
	}
}

func TestUser_JSONFieldNames(t *testing.T) {
	// Arrange
	user := User{
		ID:        "test-id",
		Email:     "test@example.com",
		Name:      "Test",
		CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
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

	// Assert - verify JSON field names
	expectedFields := []string{"id", "email", "name", "created_at", "updated_at"}
	for _, field := range expectedFields {
		if _, ok := rawMap[field]; !ok {
			t.Errorf("expected '%s' field in JSON, got keys: %v", field, keys(rawMap))
		}
	}

	// Should have exactly 5 fields (no password_hash)
	if len(rawMap) != 5 {
		t.Errorf("expected 5 JSON fields, got %d: %v", len(rawMap), keys(rawMap))
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
		t.Errorf("expected zero created_at for zero value, got '%v'", user.CreatedAt)
	}

	if !user.UpdatedAt.IsZero() {
		t.Errorf("expected zero updated_at for zero value, got '%v'", user.UpdatedAt)
	}
}

// Helper functions for test assertions
func contains(s, substr string) bool {
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

func keys(m map[string]json.RawMessage) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}
