package entities

import "time"

// RefreshToken represents a refresh token stored in the database.
type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// IsExpired returns true if the token has passed its expiration time.
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsRevoked returns true if the token has been explicitly revoked.
func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil
}

// IsUsable returns true if the token can be used for refresh.
func (rt *RefreshToken) IsUsable() bool {
	return !rt.Used && !rt.IsExpired() && !rt.IsRevoked()
}
