package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xenios/backend/internal/domain/entities"
)

func TestJWTTokenService_GenerateAccessToken_ContainsClaims(t *testing.T) {
	svc := NewJWTTokenService("test-secret", 15*time.Minute)
	user := &entities.User{
		ID:   "user-123",
		Role: "coach",
	}

	tokenStr, err := svc.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse and verify the token
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("expected MapClaims")
	}

	sub, _ := claims.GetSubject()
	if sub != "user-123" {
		t.Errorf("expected sub 'user-123', got '%s'", sub)
	}

	role, _ := claims["role"].(string)
	if role != "coach" {
		t.Errorf("expected role 'coach', got '%s'", role)
	}

	exp, _ := claims.GetExpirationTime()
	if exp == nil {
		t.Fatal("expected expiration claim")
	}
	if time.Until(exp.Time) > 16*time.Minute || time.Until(exp.Time) < 14*time.Minute {
		t.Errorf("expected ~15 min expiration, got %v", time.Until(exp.Time))
	}
}

func TestJWTTokenService_GenerateAccessToken_SignedWithHMAC(t *testing.T) {
	svc := NewJWTTokenService("my-secret", 15*time.Minute)
	user := &entities.User{ID: "u1", Role: "client"}

	tokenStr, err := svc.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should fail with wrong secret
	_, err = jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte("wrong-secret"), nil
	})
	if err == nil {
		t.Error("expected error when parsing with wrong secret")
	}
}

func TestJWTTokenService_GenerateRefreshToken_UniqueTokens(t *testing.T) {
	svc := NewJWTTokenService("secret", 15*time.Minute)

	token1, err := svc.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	token2, err := svc.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token1 == token2 {
		t.Error("expected unique tokens")
	}

	if len(token1) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("expected 64-char hex string, got %d chars", len(token1))
	}
}

func TestJWTTokenService_HashToken_DeterministicHash(t *testing.T) {
	svc := NewJWTTokenService("secret", 15*time.Minute)

	hash1 := svc.HashToken("my-token")
	hash2 := svc.HashToken("my-token")

	if hash1 != hash2 {
		t.Error("expected deterministic hash")
	}
}

func TestJWTTokenService_HashToken_DifferentInputsDifferentHashes(t *testing.T) {
	svc := NewJWTTokenService("secret", 15*time.Minute)

	hash1 := svc.HashToken("token-a")
	hash2 := svc.HashToken("token-b")

	if hash1 == hash2 {
		t.Error("expected different hashes for different inputs")
	}
}
