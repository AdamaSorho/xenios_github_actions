package repository

import "github.com/xenios/backend/internal/domain/entities"

// TokenService defines the interface for generating and validating tokens.
// This is placed in the domain layer as an interface so use cases can
// depend on it without knowing the JWT implementation details.
type TokenService interface {
	GenerateAccessToken(user *entities.User) (string, error)
	GenerateRefreshToken() (string, error)
	HashToken(token string) string
}
