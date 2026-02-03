package entities

import (
	"time"

	"github.com/google/uuid"
)

// User represents the core user entity in the domain layer.
// This is a pure business object with no external dependencies.
type User struct {
	ID        uuid.UUID
	Email     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser creates a new User with generated ID and timestamps.
func NewUser(email, name string) *User {
	now := time.Now()
	return &User{
		ID:        uuid.New(),
		Email:     email,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
