package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xenios/backend/internal/domain/entities"
)

// PostgresMeasurementRepository implements MeasurementRepository using PostgreSQL.
type PostgresMeasurementRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresMeasurementRepository creates a new PostgreSQL-backed measurement store.
func NewPostgresMeasurementRepository(pool *pgxpool.Pool) *PostgresMeasurementRepository {
	return &PostgresMeasurementRepository{pool: pool}
}

func (r *PostgresMeasurementRepository) UpsertBatch(ctx context.Context, measurements []entities.Measurement) (int, error) {
	if len(measurements) == 0 {
		return 0, nil
	}

	// Use a transaction for batch insert
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	inserted := 0
	for _, m := range measurements {
		tag, err := tx.Exec(ctx,
			`INSERT INTO measurements (client_id, source, measurement_type, value, measured_at)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (client_id, source, measurement_type, measured_at) DO NOTHING`,
			m.ClientID, m.Source, m.MeasurementType, m.Value, m.MeasuredAt,
		)
		if err != nil {
			return 0, fmt.Errorf("insert measurement: %w", err)
		}
		inserted += int(tag.RowsAffected())
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return inserted, nil
}

func (r *PostgresMeasurementRepository) Average(ctx context.Context, clientID string, source entities.WearableSource, mt entities.MeasurementType, since time.Time) (*float64, error) {
	var avg *float64
	err := r.pool.QueryRow(ctx,
		`SELECT AVG(value) FROM measurements
		 WHERE client_id = $1 AND source = $2 AND measurement_type = $3 AND measured_at >= $4`,
		clientID, source, mt, since,
	).Scan(&avg)
	if err != nil {
		return nil, fmt.Errorf("query average: %w", err)
	}
	return avg, nil
}
