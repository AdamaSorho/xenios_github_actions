package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xenios/backend/internal/domain/entities"
)

// PostgresNutritionRepository implements NutritionRepository with PostgreSQL.
type PostgresNutritionRepository struct {
	db *pgxpool.Pool
}

// NewPostgresNutritionRepository creates a new PostgresNutritionRepository.
func NewPostgresNutritionRepository(db *pgxpool.Pool) *PostgresNutritionRepository {
	return &PostgresNutritionRepository{db: db}
}

// SaveRecords stores a batch of daily nutrition records using a single transaction.
func (r *PostgresNutritionRepository) SaveRecords(ctx context.Context, records []*entities.NutritionRecord) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, rec := range records {
		_, err := tx.Exec(ctx,
			`INSERT INTO nutrition_records (client_id, coach_id, artifact_id, metric_type, value, unit, record_date)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			rec.ClientID,
			rec.CoachID,
			rec.ArtifactID,
			string(rec.MetricType),
			rec.Value,
			rec.Unit,
			rec.RecordDate,
		)
		if err != nil {
			return fmt.Errorf("insert nutrition record: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// UpsertSummary creates or updates a nutrition summary for a client/artifact.
func (r *PostgresNutritionRepository) UpsertSummary(ctx context.Context, summary *entities.NutritionSummary) error {
	now := time.Now()
	_, err := r.db.Exec(ctx,
		`INSERT INTO nutrition_summaries (
			client_id, artifact_id, total_days,
			avg_calories_7d, avg_protein_7d, avg_carbs_7d, avg_fat_7d, avg_fiber_7d,
			avg_calories_14d, avg_protein_14d, avg_carbs_14d, avg_fat_14d, avg_fiber_14d,
			avg_calories_30d, avg_protein_30d, avg_carbs_30d, avg_fat_30d, avg_fiber_30d,
			computed_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		ON CONFLICT (client_id, artifact_id) DO UPDATE SET
			total_days = EXCLUDED.total_days,
			avg_calories_7d = EXCLUDED.avg_calories_7d,
			avg_protein_7d = EXCLUDED.avg_protein_7d,
			avg_carbs_7d = EXCLUDED.avg_carbs_7d,
			avg_fat_7d = EXCLUDED.avg_fat_7d,
			avg_fiber_7d = EXCLUDED.avg_fiber_7d,
			avg_calories_14d = EXCLUDED.avg_calories_14d,
			avg_protein_14d = EXCLUDED.avg_protein_14d,
			avg_carbs_14d = EXCLUDED.avg_carbs_14d,
			avg_fat_14d = EXCLUDED.avg_fat_14d,
			avg_fiber_14d = EXCLUDED.avg_fiber_14d,
			avg_calories_30d = EXCLUDED.avg_calories_30d,
			avg_protein_30d = EXCLUDED.avg_protein_30d,
			avg_carbs_30d = EXCLUDED.avg_carbs_30d,
			avg_fat_30d = EXCLUDED.avg_fat_30d,
			avg_fiber_30d = EXCLUDED.avg_fiber_30d,
			computed_at = EXCLUDED.computed_at,
			updated_at = EXCLUDED.updated_at`,
		summary.ClientID,
		summary.ArtifactID,
		summary.TotalDays,
		summary.AvgCalories7d,
		summary.AvgProtein7d,
		summary.AvgCarbs7d,
		summary.AvgFat7d,
		summary.AvgFiber7d,
		summary.AvgCalories14d,
		summary.AvgProtein14d,
		summary.AvgCarbs14d,
		summary.AvgFat14d,
		summary.AvgFiber14d,
		summary.AvgCalories30d,
		summary.AvgProtein30d,
		summary.AvgCarbs30d,
		summary.AvgFat30d,
		summary.AvgFiber30d,
		summary.ComputedAt,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("upsert nutrition summary: %w", err)
	}
	return nil
}

// GetSummaryByClientID retrieves the latest nutrition summary for a client.
func (r *PostgresNutritionRepository) GetSummaryByClientID(ctx context.Context, clientID string) (*entities.NutritionSummary, error) {
	var s entities.NutritionSummary
	err := r.db.QueryRow(ctx,
		`SELECT id, client_id, artifact_id, total_days,
			avg_calories_7d, avg_protein_7d, avg_carbs_7d, avg_fat_7d, avg_fiber_7d,
			avg_calories_14d, avg_protein_14d, avg_carbs_14d, avg_fat_14d, avg_fiber_14d,
			avg_calories_30d, avg_protein_30d, avg_carbs_30d, avg_fat_30d, avg_fiber_30d,
			computed_at, created_at, updated_at
		FROM nutrition_summaries
		WHERE client_id = $1
		ORDER BY computed_at DESC
		LIMIT 1`,
		clientID,
	).Scan(
		&s.ID, &s.ClientID, &s.ArtifactID, &s.TotalDays,
		&s.AvgCalories7d, &s.AvgProtein7d, &s.AvgCarbs7d, &s.AvgFat7d, &s.AvgFiber7d,
		&s.AvgCalories14d, &s.AvgProtein14d, &s.AvgCarbs14d, &s.AvgFat14d, &s.AvgFiber14d,
		&s.AvgCalories30d, &s.AvgProtein30d, &s.AvgCarbs30d, &s.AvgFat30d, &s.AvgFiber30d,
		&s.ComputedAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("get nutrition summary: %w", err)
	}
	return &s, nil
}
