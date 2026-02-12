package entities

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// User entity validation errors.
var (
	ErrEmptyEmail        = errors.New("email must not be empty")
	ErrEmptyName         = errors.New("name must not be empty")
	ErrEmptyPasswordHash = errors.New("password hash must not be empty")
	ErrInvalidEmail      = errors.New("email format is invalid")
)

// User represents a registered user in the system.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// NewUser creates a new User entity with validation.
// The passwordHash parameter must be a pre-hashed password (e.g., bcrypt).
// Returns an error if any required field is empty or the email format is invalid.
func NewUser(email, name, passwordHash string) (*User, error) {
	email = strings.TrimSpace(email)
	name = strings.TrimSpace(name)

	if email == "" {
		return nil, ErrEmptyEmail
	}
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}
	if name == "" {
		return nil, ErrEmptyName
	}
	if passwordHash == "" {
		return nil, ErrEmptyPasswordHash
	}

	now := time.Now().UTC()
	return &User{
		ID:           generateID(),
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// isValidEmail performs basic email format validation.
// Checks for presence of @ with non-empty local part and domain.
func isValidEmail(email string) bool {
	if strings.Contains(email, " ") {
		return false
	}
	at := strings.IndexByte(email, '@')
	if at < 1 {
		return false
	}
	domain := email[at+1:]
	if domain == "" {
		return false
	}
	return true
}

// generateID creates a unique identifier for a new user.
// Uses a simple format; in production this would use UUID generation.
func generateID() string {
	return fmt.Sprintf("user-%d", time.Now().UnixNano())
}
