package entities

import "time"

// User represents an authenticated user in the system.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	AvatarURL    *string   `json:"avatar_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ValidRoles are the allowed user roles.
var ValidRoles = map[string]bool{
	"coach":  true,
	"client": true,
	"admin":  true,
}

// IsValidRole checks whether the given role is valid.
func IsValidRole(role string) bool {
	return ValidRoles[role]
}
