package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xenios/backend/internal/domain/entities"
)

// PostgresWearableSummaryRepository implements WearableSummaryRepository with PostgreSQL.
type PostgresWearableSummaryRepository struct {
	db *pgxpool.Pool
}

// NewPostgresWearableSummaryRepository creates a new PostgresWearableSummaryRepository.
func NewPostgresWearableSummaryRepository(db *pgxpool.Pool) *PostgresWearableSummaryRepository {
	return &PostgresWearableSummaryRepository{db: db}
}

// Upsert creates or updates a wearable summary.
func (r *PostgresWearableSummaryRepository) Upsert(ctx context.Context, summary *entities.WearableSummary) (*entities.WearableSummary, error) {
	err := r.db.QueryRow(ctx,
		`INSERT INTO wearable_summaries (client_id, source, summary_date, metrics, synced_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 ON CONFLICT (client_id, source, summary_date)
		 DO UPDATE SET metrics = $4, synced_at = NOW()
		 RETURNING id, synced_at, created_at`,
		summary.ClientID, summary.Source, summary.SummaryDate, summary.Metrics,
	).Scan(&summary.ID, &summary.SyncedAt, &summary.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("upsert wearable summary: %w", err)
	}
	return summary, nil
}

// FindByClientID retrieves wearable summaries for a client within the last N days.
func (r *PostgresWearableSummaryRepository) FindByClientID(ctx context.Context, clientID string, days int) ([]*entities.WearableSummary, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, client_id, source, summary_date, metrics, synced_at, created_at
		 FROM wearable_summaries
		 WHERE client_id = $1 AND summary_date >= CURRENT_DATE - $2::INT
		 ORDER BY summary_date DESC`,
		clientID, days,
	)
	if err != nil {
		return nil, fmt.Errorf("query wearable summaries: %w", err)
	}
	defer rows.Close()

	var summaries []*entities.WearableSummary
	for rows.Next() {
		var s entities.WearableSummary
		if err := rows.Scan(
			&s.ID, &s.ClientID, &s.Source, &s.SummaryDate, &s.Metrics, &s.SyncedAt, &s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan wearable summary: %w", err)
		}
		summaries = append(summaries, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate wearable summaries: %w", err)
	}

	return summaries, nil
}
