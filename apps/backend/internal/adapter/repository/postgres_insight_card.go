package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
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

// Create inserts a new insight card and its evidence references.
func (r *PostgresInsightCardRepository) Create(ctx context.Context, card *entities.InsightCard) (*entities.InsightCard, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var saved entities.InsightCard
	err = tx.QueryRow(ctx,
		`INSERT INTO insight_cards (coach_id, client_id, title, body, category, priority, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, coach_id, client_id, title, body, category, priority, status, created_at, updated_at`,
		card.CoachID, card.ClientID, card.Title, card.Body,
		card.Category, card.Priority, card.Status,
	).Scan(
		&saved.ID, &saved.CoachID, &saved.ClientID, &saved.Title, &saved.Body,
		&saved.Category, &saved.Priority, &saved.Status, &saved.CreatedAt, &saved.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert insight card: %w", err)
	}

	for _, ev := range card.Evidence {
		_, err = tx.Exec(ctx,
			`INSERT INTO insight_evidence (insight_card_id, measurement_id, artifact_id, description)
			 VALUES ($1, $2, $3, $4)`,
			saved.ID, nilIfEmpty(ev.MeasurementID), nilIfEmpty(ev.ArtifactID), ev.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("insert evidence ref: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	saved.Evidence = card.Evidence
	return &saved, nil
}

// FindByID returns the insight card with the given ID including evidence.
func (r *PostgresInsightCardRepository) FindByID(ctx context.Context, id string) (*entities.InsightCard, error) {
	var card entities.InsightCard
	err := r.db.QueryRow(ctx,
		`SELECT id, coach_id, client_id, title, body, category, priority, status, created_at, updated_at
		 FROM insight_cards WHERE id = $1`, id,
	).Scan(
		&card.ID, &card.CoachID, &card.ClientID, &card.Title, &card.Body,
		&card.Category, &card.Priority, &card.Status, &card.CreatedAt, &card.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find insight card: %w", err)
	}

	evidence, err := r.loadEvidence(ctx, card.ID)
	if err != nil {
		return nil, err
	}
	card.Evidence = evidence

	return &card, nil
}

// FindByClientID returns insight cards for a client with pagination.
func (r *PostgresInsightCardRepository) FindByClientID(ctx context.Context, clientID string, limit, offset int) ([]*entities.InsightCard, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(ctx,
		`SELECT id, coach_id, client_id, title, body, category, priority, status, created_at, updated_at
		 FROM insight_cards WHERE client_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		clientID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("query insight cards: %w", err)
	}
	defer rows.Close()

	return r.scanCards(ctx, rows)
}

// FindByStatus returns insight cards filtered by coach and status.
func (r *PostgresInsightCardRepository) FindByStatus(ctx context.Context, coachID string, status entities.InsightStatus, limit, offset int) ([]*entities.InsightCard, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(ctx,
		`SELECT id, coach_id, client_id, title, body, category, priority, status, created_at, updated_at
		 FROM insight_cards WHERE coach_id = $1 AND status = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		coachID, status, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("query insight cards by status: %w", err)
	}
	defer rows.Close()

	return r.scanCards(ctx, rows)
}

// ExistsByEvidence checks if an insight already exists for a given client and measurement.
func (r *PostgresInsightCardRepository) ExistsByEvidence(ctx context.Context, clientID string, measurementID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM insight_cards ic
			JOIN insight_evidence ie ON ie.insight_card_id = ic.id
			WHERE ic.client_id = $1 AND ie.measurement_id = $2
		)`, clientID, measurementID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check evidence exists: %w", err)
	}
	return exists, nil
}

// UpdateStatus updates the status of an insight card.
func (r *PostgresInsightCardRepository) UpdateStatus(ctx context.Context, id string, status entities.InsightStatus) (*entities.InsightCard, error) {
	var card entities.InsightCard
	err := r.db.QueryRow(ctx,
		`UPDATE insight_cards SET status = $1, updated_at = NOW()
		 WHERE id = $2
		 RETURNING id, coach_id, client_id, title, body, category, priority, status, created_at, updated_at`,
		status, id,
	).Scan(
		&card.ID, &card.CoachID, &card.ClientID, &card.Title, &card.Body,
		&card.Category, &card.Priority, &card.Status, &card.CreatedAt, &card.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("insight card not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("update insight card status: %w", err)
	}

	evidence, err := r.loadEvidence(ctx, card.ID)
	if err != nil {
		return nil, err
	}
	card.Evidence = evidence

	return &card, nil
}

func (r *PostgresInsightCardRepository) loadEvidence(ctx context.Context, cardID string) ([]entities.EvidenceRef, error) {
	rows, err := r.db.Query(ctx,
		`SELECT COALESCE(measurement_id::TEXT, ''), COALESCE(artifact_id::TEXT, ''), description
		 FROM insight_evidence WHERE insight_card_id = $1`, cardID,
	)
	if err != nil {
		return nil, fmt.Errorf("load evidence: %w", err)
	}
	defer rows.Close()

	var evidence []entities.EvidenceRef
	for rows.Next() {
		var ev entities.EvidenceRef
		if err := rows.Scan(&ev.MeasurementID, &ev.ArtifactID, &ev.Description); err != nil {
			return nil, fmt.Errorf("scan evidence: %w", err)
		}
		evidence = append(evidence, ev)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate evidence: %w", err)
	}
	return evidence, nil
}

func (r *PostgresInsightCardRepository) scanCards(ctx context.Context, rows pgx.Rows) ([]*entities.InsightCard, error) {
	var cards []*entities.InsightCard
	for rows.Next() {
		var card entities.InsightCard
		if err := rows.Scan(
			&card.ID, &card.CoachID, &card.ClientID, &card.Title, &card.Body,
			&card.Category, &card.Priority, &card.Status, &card.CreatedAt, &card.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan insight card: %w", err)
		}
		cards = append(cards, &card)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate insight cards: %w", err)
	}

	for _, card := range cards {
		evidence, err := r.loadEvidence(ctx, card.ID)
		if err != nil {
			return nil, err
		}
		card.Evidence = evidence
	}

	return cards, nil
}
