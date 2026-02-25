package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xenios/backend/internal/domain/entities"
)

// PostgresWearableSummaryRepository implements WearableSummaryRepository using PostgreSQL.
type PostgresWearableSummaryRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresWearableSummaryRepository creates a new PostgreSQL-backed wearable summary store.
func NewPostgresWearableSummaryRepository(pool *pgxpool.Pool) *PostgresWearableSummaryRepository {
	return &PostgresWearableSummaryRepository{pool: pool}
}

func (r *PostgresWearableSummaryRepository) Upsert(ctx context.Context, clientID string, source entities.WearableSource, metrics json.RawMessage) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO wearable_summaries (client_id, source, metrics, updated_at)
		 VALUES ($1, $2, $3, now())
		 ON CONFLICT (client_id, source)
		 DO UPDATE SET metrics = $3, updated_at = now()`,
		clientID, source, metrics,
	)
	if err != nil {
		return fmt.Errorf("upsert summary: %w", err)
	}
	return nil
}
