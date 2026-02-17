package usecase

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// PasswordHasher abstracts password hashing so the use case does not depend
// on bcrypt directly.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

// RegisterUserUseCase handles user registration.
type RegisterUserUseCase struct {
	userRepo     repository.UserRepository
	tokenRepo    repository.RefreshTokenRepository
	tokenService repository.TokenService
	auditRepo    repository.AuditRepository
	hasher       PasswordHasher
}

// NewRegisterUserUseCase creates a new RegisterUserUseCase.
func NewRegisterUserUseCase(
	userRepo repository.UserRepository,
	tokenRepo repository.RefreshTokenRepository,
	tokenService repository.TokenService,
	auditRepo repository.AuditRepository,
	hasher PasswordHasher,
) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		tokenService: tokenService,
		auditRepo:    auditRepo,
		hasher:       hasher,
	}
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// RegisterInput holds the input for user registration.
type RegisterInput struct {
	Email    string
	Password string
	Name     string
	Role     string
}

// RegisterOutput holds the output of user registration.
type RegisterOutput struct {
	User   *entities.User       `json:"user"`
	Tokens *entities.AuthTokens `json:"tokens"`
}

// Execute registers a new user and returns tokens.
func (uc *RegisterUserUseCase) Execute(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.Name = strings.TrimSpace(input.Name)

	if input.Email == "" {
		return nil, &ValidationError{Message: "email is required"}
	}
	if !emailRegex.MatchString(input.Email) {
		return nil, &ValidationError{Message: "invalid email format"}
	}
	if input.Password == "" {
		return nil, &ValidationError{Message: "password is required"}
	}
	if len(input.Password) < 8 {
		return nil, &ValidationError{Message: "password must be at least 8 characters"}
	}
	if input.Name == "" {
		return nil, &ValidationError{Message: "name is required"}
	}
	if input.Role == "" {
		input.Role = "client"
	}
	if !entities.IsValidRole(input.Role) {
		return nil, &ValidationError{Message: fmt.Sprintf("invalid role: %s", input.Role)}
	}

	existing, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("check existing user: %w", err)
	}
	if existing != nil {
		return nil, &ValidationError{Message: "email already registered"}
	}

	hash, err := uc.hasher.Hash(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := uc.userRepo.Create(ctx, input.Email, hash, input.Name, input.Role)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	accessToken, err := uc.tokenService.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshTokenRaw, err := uc.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	refreshHash := uc.tokenService.HashToken(refreshTokenRaw)
	_, err = uc.tokenRepo.Create(ctx, user.ID, refreshHash, time.Now().Add(7*24*time.Hour))
	if err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	if auditErr := uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    user.ID,
		Action:     "user.registered",
		EntityType: "user",
		EntityID:   user.ID,
	}); auditErr != nil {
		log.Printf("audit log error: %v", auditErr)
	}

	return &RegisterOutput{
		User: user,
		Tokens: &entities.AuthTokens{
			AccessToken:  accessToken,
			RefreshToken: refreshTokenRaw,
		},
	}, nil
}
