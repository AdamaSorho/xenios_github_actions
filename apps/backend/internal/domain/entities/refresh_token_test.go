package entities

import (
	"testing"
	"time"
)

func TestRefreshToken_IsExpired_NotExpired_ReturnsFalse(t *testing.T) {
	rt := &RefreshToken{
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if rt.IsExpired() {
		t.Error("expected token to not be expired")
	}
}

func TestRefreshToken_IsExpired_Expired_ReturnsTrue(t *testing.T) {
	rt := &RefreshToken{
		ExpiresAt: time.Now().Add(-time.Hour),
	}
	if !rt.IsExpired() {
		t.Error("expected token to be expired")
	}
}

func TestRefreshToken_IsRevoked_NotRevoked_ReturnsFalse(t *testing.T) {
	rt := &RefreshToken{}
	if rt.IsRevoked() {
		t.Error("expected token to not be revoked")
	}
}

func TestRefreshToken_IsRevoked_Revoked_ReturnsTrue(t *testing.T) {
	now := time.Now()
	rt := &RefreshToken{
		RevokedAt: &now,
	}
	if !rt.IsRevoked() {
		t.Error("expected token to be revoked")
	}
}

func TestRefreshToken_IsUsable_ValidToken_ReturnsTrue(t *testing.T) {
	rt := &RefreshToken{
		Used:      false,
		ExpiresAt: time.Now().Add(time.Hour),
		RevokedAt: nil,
	}
	if !rt.IsUsable() {
		t.Error("expected token to be usable")
	}
}

func TestRefreshToken_IsUsable_UsedToken_ReturnsFalse(t *testing.T) {
	rt := &RefreshToken{
		Used:      true,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if rt.IsUsable() {
		t.Error("expected used token to not be usable")
	}
}

func TestRefreshToken_IsUsable_ExpiredToken_ReturnsFalse(t *testing.T) {
	rt := &RefreshToken{
		Used:      false,
		ExpiresAt: time.Now().Add(-time.Hour),
	}
	if rt.IsUsable() {
		t.Error("expected expired token to not be usable")
	}
}

func TestRefreshToken_IsUsable_RevokedToken_ReturnsFalse(t *testing.T) {
	now := time.Now()
	rt := &RefreshToken{
		Used:      false,
		ExpiresAt: time.Now().Add(time.Hour),
		RevokedAt: &now,
	}
	if rt.IsUsable() {
		t.Error("expected revoked token to not be usable")
	}
}
