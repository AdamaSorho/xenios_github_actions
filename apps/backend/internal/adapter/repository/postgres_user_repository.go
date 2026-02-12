package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// Querier abstracts the pgx query interface for testability.
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

// PostgresUserRepository implements the UserRepository interface using PostgreSQL.
type PostgresUserRepository struct {
	db Querier
}

// NewPostgresUserRepository creates a new PostgresUserRepository instance.
func NewPostgresUserRepository(db Querier) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// Create inserts a new user into the database and returns the created user
// with database-generated fields (id, created_at, updated_at).
func (r *PostgresUserRepository) Create(ctx context.Context, user *entities.User) (*entities.User, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO users (email, name, password_hash)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, name, password_hash, created_at, updated_at`,
		user.Email, user.Name, user.PasswordHash,
	)

	var created entities.User
	err := row.Scan(
		&created.ID,
		&created.Email,
		&created.Name,
		&created.PasswordHash,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, repository.ErrDuplicateEmail
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &created, nil
}

// FindByEmail retrieves a user by their email address.
// Returns nil, nil when no user is found with the given email.
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, email, name, password_hash, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	)

	var user entities.User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	return &user, nil
}

// Compile-time verification that PostgresUserRepository implements UserRepository.
var _ repository.UserRepository = (*PostgresUserRepository)(nil)
