package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// PostgresUserRepository implements UserRepository using PostgreSQL/Supabase.
// Uses raw SQL with pgx - NO ORMs allowed.
type PostgresUserRepository struct {
	db *pgxpool.Pool
}

// NewPostgresUserRepository creates a new PostgresUserRepository.
func NewPostgresUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	var u entities.User
	err := r.db.QueryRow(ctx,
		"SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1",
		id,
	).Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	var u entities.User
	err := r.db.QueryRow(ctx,
		"SELECT id, email, name, created_at, updated_at FROM users WHERE email = $1",
		email,
	).Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *entities.User) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO users (id, email, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
		user.ID, user.Email, user.Name, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *entities.User) error {
	_, err := r.db.Exec(ctx,
		"UPDATE users SET email = $1, name = $2, updated_at = $3 WHERE id = $4",
		user.Email, user.Name, user.UpdatedAt, user.ID,
	)
	return err
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		"DELETE FROM users WHERE id = $1",
		id,
	)
	return err
}
