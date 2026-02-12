package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xenios/backend/internal/domain/entities"
)

// JWTTokenService implements the TokenService interface using JWT.
type JWTTokenService struct {
	secret          string
	accessTokenTTL  time.Duration
}

// NewJWTTokenService creates a new JWTTokenService.
func NewJWTTokenService(secret string, accessTokenTTL time.Duration) *JWTTokenService {
	return &JWTTokenService{
		secret:         secret,
		accessTokenTTL: accessTokenTTL,
	}
}

// GenerateAccessToken creates a signed JWT access token containing user claims.
func (s *JWTTokenService) GenerateAccessToken(user *entities.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"iat":  now.Unix(),
		"exp":  now.Add(s.accessTokenTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

// GenerateRefreshToken creates a cryptographically random refresh token.
func (s *JWTTokenService) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// HashToken returns a SHA-256 hash of the token for secure storage.
func (s *JWTTokenService) HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
