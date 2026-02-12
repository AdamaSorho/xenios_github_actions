package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestBcryptHasher_Hash_ProducesValidHash(t *testing.T) {
	hasher := NewBcryptHasher(bcrypt.MinCost) // Use min cost for test speed
	hash, err := hasher.Hash("password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "" {
		t.Error("expected non-empty hash")
	}
	if hash == "password123" {
		t.Error("hash should not equal plaintext")
	}
}

func TestBcryptHasher_Compare_CorrectPassword_NoError(t *testing.T) {
	hasher := NewBcryptHasher(bcrypt.MinCost)
	hash, err := hasher.Hash("correctpassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = hasher.Compare(hash, "correctpassword")
	if err != nil {
		t.Errorf("expected no error for correct password, got %v", err)
	}
}

func TestBcryptHasher_Compare_WrongPassword_ReturnsError(t *testing.T) {
	hasher := NewBcryptHasher(bcrypt.MinCost)
	hash, err := hasher.Hash("correctpassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = hasher.Compare(hash, "wrongpassword")
	if err == nil {
		t.Error("expected error for wrong password")
	}
}

func TestBcryptHasher_NewBcryptHasher_DefaultCost(t *testing.T) {
	hasher := NewBcryptHasher(0) // Below MinCost
	if hasher.cost != bcrypt.DefaultCost {
		t.Errorf("expected default cost %d, got %d", bcrypt.DefaultCost, hasher.cost)
	}
}

func TestBcryptHasher_NewBcryptHasher_CustomCost(t *testing.T) {
	hasher := NewBcryptHasher(12)
	if hasher.cost != 12 {
		t.Errorf("expected cost 12, got %d", hasher.cost)
	}
}
