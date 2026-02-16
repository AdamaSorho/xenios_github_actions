package repository

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryRefreshTokenRepository_Create_ReturnsToken(t *testing.T) {
	repo := NewInMemoryRefreshTokenRepository()
	expires := time.Now().Add(7 * 24 * time.Hour)

	rt, err := repo.Create(context.Background(), "user-1", "hash123", expires)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rt == nil {
		t.Fatal("expected refresh token")
	}
	if rt.ID == "" {
		t.Error("expected non-empty ID")
	}
	if rt.UserID != "user-1" {
		t.Errorf("expected user_id 'user-1', got '%s'", rt.UserID)
	}
	if rt.TokenHash != "hash123" {
		t.Errorf("expected token_hash 'hash123', got '%s'", rt.TokenHash)
	}
	if rt.Used {
		t.Error("expected Used to be false")
	}
}

func TestInMemoryRefreshTokenRepository_Create_UniqueIDs(t *testing.T) {
	repo := NewInMemoryRefreshTokenRepository()
	ctx := context.Background()
	expires := time.Now().Add(time.Hour)

	rt1, _ := repo.Create(ctx, "user-1", "hash1", expires)
	rt2, _ := repo.Create(ctx, "user-1", "hash2", expires)

	if rt1.ID == rt2.ID {
		t.Errorf("expected unique IDs, got '%s' for both", rt1.ID)
	}
}

func TestInMemoryRefreshTokenRepository_FindByTokenHash_Exists(t *testing.T) {
	repo := NewInMemoryRefreshTokenRepository()
	ctx := context.Background()

	_, _ = repo.Create(ctx, "user-1", "target-hash", time.Now().Add(time.Hour))

	found, err := repo.FindByTokenHash(ctx, "target-hash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil {
		t.Fatal("expected to find token")
	}
	if found.TokenHash != "target-hash" {
		t.Errorf("expected hash 'target-hash', got '%s'", found.TokenHash)
	}
}

func TestInMemoryRefreshTokenRepository_FindByTokenHash_NotExists(t *testing.T) {
	repo := NewInMemoryRefreshTokenRepository()

	found, err := repo.FindByTokenHash(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Error("expected nil for non-existent hash")
	}
}

func TestInMemoryRefreshTokenRepository_MarkUsed_SetsUsedFlag(t *testing.T) {
	repo := NewInMemoryRefreshTokenRepository()
	ctx := context.Background()

	rt, _ := repo.Create(ctx, "user-1", "hash1", time.Now().Add(time.Hour))

	err := repo.MarkUsed(ctx, rt.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, _ := repo.FindByTokenHash(ctx, "hash1")
	if found == nil {
		t.Fatal("expected token")
	}
	if !found.Used {
		t.Error("expected Used to be true after MarkUsed")
	}
}

func TestInMemoryRefreshTokenRepository_MarkUsed_NonExistent_NoError(t *testing.T) {
	repo := NewInMemoryRefreshTokenRepository()

	err := repo.MarkUsed(context.Background(), "nonexistent-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInMemoryRefreshTokenRepository_RevokeAllForUser_RevokesAll(t *testing.T) {
	repo := NewInMemoryRefreshTokenRepository()
	ctx := context.Background()

	_, _ = repo.Create(ctx, "user-1", "hash1", time.Now().Add(time.Hour))
	_, _ = repo.Create(ctx, "user-1", "hash2", time.Now().Add(time.Hour))
	_, _ = repo.Create(ctx, "user-2", "hash3", time.Now().Add(time.Hour))

	err := repo.RevokeAllForUser(ctx, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// user-1 tokens should be revoked
	rt1, _ := repo.FindByTokenHash(ctx, "hash1")
	if rt1.RevokedAt == nil {
		t.Error("expected user-1 token 1 to be revoked")
	}

	rt2, _ := repo.FindByTokenHash(ctx, "hash2")
	if rt2.RevokedAt == nil {
		t.Error("expected user-1 token 2 to be revoked")
	}

	// user-2 token should not be revoked
	rt3, _ := repo.FindByTokenHash(ctx, "hash3")
	if rt3.RevokedAt != nil {
		t.Error("expected user-2 token to not be revoked")
	}
}
