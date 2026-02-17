package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xenios/backend/internal/domain/entities"
)

// PostgresInsightCardRepository implements InsightCardRepository with PostgreSQL.
type PostgresInsightCardRepository struct {
	db *pgxpool.Pool
}

// NewPostgresInsightCardRepository creates a new PostgresInsightCardRepository.
func NewPostgresInsightCardRepository(db *pgxpool.Pool) *PostgresInsightCardRepository {
	return &PostgresInsightCardRepository{db: db}
}

// Create persists a new insight card and returns it with generated ID and timestamps.
func (r *PostgresInsightCardRepository) Create(ctx context.Context, card *entities.InsightCard) (*entities.InsightCard, error) {
	evidenceJSON, err := json.Marshal(card.Evidence)
	if err != nil {
		return nil, fmt.Errorf("marshal evidence: %w", err)
	}

	var created entities.InsightCard
	err = r.db.QueryRow(ctx,
		`INSERT INTO insight_cards (coach_id, client_id, title, body, category, priority, status, evidence)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, coach_id, client_id, title, body, category, priority, status, evidence, created_at, updated_at`,
		card.CoachID, card.ClientID, card.Title, card.Body,
		string(card.Category), string(card.Priority), string(card.Status),
		evidenceJSON,
	).Scan(
		&created.ID, &created.CoachID, &created.ClientID,
		&created.Title, &created.Body,
		&created.Category, &created.Priority, &created.Status,
		&evidenceJSON,
		&created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert insight card: %w", err)
	}

	if err := json.Unmarshal(evidenceJSON, &created.Evidence); err != nil {
		return nil, fmt.Errorf("unmarshal evidence: %w", err)
	}

	return &created, nil
}

// FindByClientID returns insight cards for a given client with pagination.
func (r *PostgresInsightCardRepository) FindByClientID(ctx context.Context, clientID string, limit, offset int) ([]*entities.InsightCard, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, coach_id, client_id, title, body, category, priority, status, evidence, created_at, updated_at
		 FROM insight_cards WHERE client_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		clientID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("query insight cards by client: %w", err)
	}
	defer rows.Close()

	return scanInsightCards(rows)
}

// FindByStatus returns insight cards matching a given status with pagination.
func (r *PostgresInsightCardRepository) FindByStatus(ctx context.Context, status entities.InsightStatus, limit, offset int) ([]*entities.InsightCard, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, coach_id, client_id, title, body, category, priority, status, evidence, created_at, updated_at
		 FROM insight_cards WHERE status = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		string(status), limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("query insight cards by status: %w", err)
	}
	defer rows.Close()

	return scanInsightCards(rows)
}

// UpdateStatus changes the status of an insight card by ID.
func (r *PostgresInsightCardRepository) UpdateStatus(ctx context.Context, id string, status entities.InsightStatus) (*entities.InsightCard, error) {
	var card entities.InsightCard
	var evidenceJSON []byte

	err := r.db.QueryRow(ctx,
		`UPDATE insight_cards SET status = $1
		 WHERE id = $2
		 RETURNING id, coach_id, client_id, title, body, category, priority, status, evidence, created_at, updated_at`,
		string(status), id,
	).Scan(
		&card.ID, &card.CoachID, &card.ClientID,
		&card.Title, &card.Body,
		&card.Category, &card.Priority, &card.Status,
		&evidenceJSON,
		&card.CreatedAt, &card.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update insight card status: %w", err)
	}

	if err := json.Unmarshal(evidenceJSON, &card.Evidence); err != nil {
		return nil, fmt.Errorf("unmarshal evidence: %w", err)
	}

	return &card, nil
}

// ExistsByMeasurementID returns true if an insight card already exists
// that references the given measurement ID in its evidence.
func (r *PostgresInsightCardRepository) ExistsByMeasurementID(ctx context.Context, measurementID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM insight_cards
			WHERE evidence @> $1::jsonb
		)`,
		fmt.Sprintf(`[{"measurement_id":"%s"}]`, measurementID),
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check insight card by measurement: %w", err)
	}
	return exists, nil
}

// scanInsightCards is a pgx rows scanner.
type pgxRows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Err() error
}

func scanInsightCards(rows pgxRows) ([]*entities.InsightCard, error) {
	var cards []*entities.InsightCard
	for rows.Next() {
		var card entities.InsightCard
		var evidenceJSON []byte
		if err := rows.Scan(
			&card.ID, &card.CoachID, &card.ClientID,
			&card.Title, &card.Body,
			&card.Category, &card.Priority, &card.Status,
			&evidenceJSON,
			&card.CreatedAt, &card.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan insight card: %w", err)
		}
		if err := json.Unmarshal(evidenceJSON, &card.Evidence); err != nil {
			return nil, fmt.Errorf("unmarshal evidence: %w", err)
		}
		cards = append(cards, &card)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate insight cards: %w", err)
	}
	return cards, nil
}
