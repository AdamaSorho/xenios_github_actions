package entities

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUser_Creation(t *testing.T) {
	// Arrange & Act
	now := time.Now()
	user := User{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "$2a$10$hashedpassword",
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
	if user.PasswordHash != "$2a$10$hashedpassword" {
		t.Errorf("expected password hash to be set, got '%s'", user.PasswordHash)
	}
	if !user.CreatedAt.Equal(now) {
		t.Errorf("expected created_at '%v', got '%v'", now, user.CreatedAt)
	}
	if !user.UpdatedAt.Equal(now) {
		t.Errorf("expected updated_at '%v', got '%v'", now, user.UpdatedAt)
	}
}

func TestUser_JSONSerialization_ExcludesPasswordHash(t *testing.T) {
	// Arrange
	now := time.Now().UTC().Truncate(time.Second)
	user := User{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "$2a$10$secrethash",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Act
	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("failed to marshal user: %v", err)
	}

	jsonStr := string(data)

	// Assert - PasswordHash must NOT appear in JSON
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

	// Verify other fields ARE present
	if _, ok := rawMap["id"]; !ok {
		t.Errorf("expected 'id' field in JSON, got: %s", jsonStr)
	}
	if _, ok := rawMap["email"]; !ok {
		t.Errorf("expected 'email' field in JSON, got: %s", jsonStr)
	}
	if _, ok := rawMap["name"]; !ok {
		t.Errorf("expected 'name' field in JSON, got: %s", jsonStr)
	}
	if _, ok := rawMap["created_at"]; !ok {
		t.Errorf("expected 'created_at' field in JSON, got: %s", jsonStr)
	}
	if _, ok := rawMap["updated_at"]; !ok {
		t.Errorf("expected 'updated_at' field in JSON, got: %s", jsonStr)
	}
}

func TestUser_JSONDeserialization(t *testing.T) {
	// Arrange
	jsonData := `{"id":"abc-123","email":"user@test.com","name":"Jane Doe","created_at":"2024-01-15T10:30:00Z","updated_at":"2024-01-15T10:30:00Z"}`

	// Act
	var user User
	err := json.Unmarshal([]byte(jsonData), &user)
	if err != nil {
		t.Fatalf("failed to unmarshal user: %v", err)
	}

	// Assert
	if user.ID != "abc-123" {
		t.Errorf("expected ID 'abc-123', got '%s'", user.ID)
	}
	if user.Email != "user@test.com" {
		t.Errorf("expected email 'user@test.com', got '%s'", user.Email)
	}
	if user.Name != "Jane Doe" {
		t.Errorf("expected name 'Jane Doe', got '%s'", user.Name)
	}
	// PasswordHash should be empty since it's not in JSON
	if user.PasswordHash != "" {
		t.Errorf("expected empty password hash from JSON, got '%s'", user.PasswordHash)
	}
}

func TestUser_JSONFieldNames(t *testing.T) {
	// Arrange
	now := time.Now().UTC()
	user := User{
		ID:        "test-id",
		Email:     "test@example.com",
		Name:      "Test",
		CreatedAt: now,
		UpdatedAt: now,
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

	// Assert - verify JSON field names use snake_case
	expectedFields := []string{"id", "email", "name", "created_at", "updated_at"}
	for _, field := range expectedFields {
		if _, ok := rawMap[field]; !ok {
			t.Errorf("expected '%s' field in JSON", field)
		}
	}

	// Verify exactly 5 fields (no password_hash)
	if len(rawMap) != 5 {
		t.Errorf("expected exactly 5 JSON fields, got %d", len(rawMap))
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

func TestUser_JSONRoundTrip(t *testing.T) {
	// Arrange
	now := time.Now().UTC().Truncate(time.Second)
	original := User{
		ID:           "round-trip-id",
		Email:        "roundtrip@example.com",
		Name:         "Round Trip",
		PasswordHash: "this-should-not-survive-roundtrip",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Act - marshal and unmarshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded User
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Assert - all fields survive except PasswordHash
	if decoded.ID != original.ID {
		t.Errorf("expected ID '%s', got '%s'", original.ID, decoded.ID)
	}
	if decoded.Email != original.Email {
		t.Errorf("expected email '%s', got '%s'", original.Email, decoded.Email)
	}
	if decoded.Name != original.Name {
		t.Errorf("expected name '%s', got '%s'", original.Name, decoded.Name)
	}
	if decoded.PasswordHash != "" {
		t.Errorf("expected empty password hash after round trip, got '%s'", decoded.PasswordHash)
	}
	if !decoded.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("expected created_at '%v', got '%v'", original.CreatedAt, decoded.CreatedAt)
	}
	if !decoded.UpdatedAt.Equal(original.UpdatedAt) {
		t.Errorf("expected updated_at '%v', got '%v'", original.UpdatedAt, decoded.UpdatedAt)
	}
}
