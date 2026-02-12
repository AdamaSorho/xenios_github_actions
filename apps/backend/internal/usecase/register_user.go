package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// RegisterUserInput contains the data needed to register a new user.
type RegisterUserInput struct {
	Email    string
	Password string
	Name     string
}

// RegisterUserOutput contains the registered user data without sensitive fields.
type RegisterUserOutput struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RegisterUserUseCase orchestrates user registration.
type RegisterUserUseCase struct {
	userRepo repository.UserRepository
}

// NewRegisterUserUseCase creates a new RegisterUserUseCase instance.
func NewRegisterUserUseCase(userRepo repository.UserRepository) *RegisterUserUseCase {
	return &RegisterUserUseCase{userRepo: userRepo}
}

// Execute performs the user registration flow:
// 1. Validate input
// 2. Check for duplicate email
// 3. Hash password with bcrypt
// 4. Create user via repository
// 5. Return output without password hash
func (uc *RegisterUserUseCase) Execute(ctx context.Context, input RegisterUserInput) (*RegisterUserOutput, error) {
	// Normalize input
	email := strings.ToLower(strings.TrimSpace(input.Email))
	name := strings.TrimSpace(input.Name)

	// Step 1: Validate input
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if len(input.Password) < 8 {
		return nil, ErrPasswordTooShort
	}
	if name == "" {
		return nil, ErrNameRequired
	}

	// Step 2: Check for duplicate email
	existingUser, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("checking email availability: %w", err)
	}
	if existingUser != nil {
		return nil, ErrEmailExists
	}

	// Step 3: Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	// Step 4: Create user entity and persist
	user := &entities.User{
		Email:        email,
		Name:         name,
		PasswordHash: string(hashedPassword),
	}

	createdUser, err := uc.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	// Step 5: Map to output (exclude password hash)
	return &RegisterUserOutput{
		ID:        createdUser.ID,
		Email:     createdUser.Email,
		Name:      createdUser.Name,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
	}, nil
}

// validateEmail checks that the email contains @ and has a dot in the domain part.
func validateEmail(email string) error {
	atIndex := strings.Index(email, "@")
	if atIndex < 1 {
		return ErrInvalidEmail
	}
	domain := email[atIndex+1:]
	if domain == "" || !strings.Contains(domain, ".") {
		return ErrInvalidEmail
	}
	return nil
}
