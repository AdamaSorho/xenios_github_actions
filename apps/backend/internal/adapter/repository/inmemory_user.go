package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryUserRepository is an in-memory implementation of UserRepository.
type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*entities.User
	count int
}

// NewInMemoryUserRepository creates a new InMemoryUserRepository.
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]*entities.User),
	}
}

// Create stores a new user.
func (r *InMemoryUserRepository) Create(_ context.Context, email, passwordHash, name, role string) (*entities.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.count++
	id := fmt.Sprintf("user-%d", r.count)
	now := time.Now()

	user := &entities.User{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	r.users[id] = user
	return user, nil
}

// FindByEmail returns a user by email, or nil if not found.
func (r *InMemoryUserRepository) FindByEmail(_ context.Context, email string) (*entities.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

// FindByID returns a user by ID, or nil if not found.
func (r *InMemoryUserRepository) FindByID(_ context.Context, id string) (*entities.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, nil
	}
	return user, nil
}
