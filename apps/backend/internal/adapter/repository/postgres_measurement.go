package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xenios/backend/internal/domain/entities"
)

// PostgresMeasurementRepository implements MeasurementRepository with PostgreSQL.
type PostgresMeasurementRepository struct {
	db *pgxpool.Pool
}

// NewPostgresMeasurementRepository creates a new PostgresMeasurementRepository.
func NewPostgresMeasurementRepository(db *pgxpool.Pool) *PostgresMeasurementRepository {
	return &PostgresMeasurementRepository{db: db}
}

// FindByClientID returns measurements for a client with pagination.
func (r *PostgresMeasurementRepository) FindByClientID(ctx context.Context, clientID string, limit, offset int) ([]*entities.Measurement, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(ctx,
		`SELECT id, client_id, recorded_by, measurement_type, value, unit,
		        flag, ref_range_low, ref_range_high, COALESCE(artifact_id::TEXT, ''),
		        measured_at, COALESCE(notes, ''), created_at
		 FROM measurements
		 WHERE client_id = $1
		 ORDER BY measured_at DESC
		 LIMIT $2 OFFSET $3`,
		clientID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("query measurements: %w", err)
	}
	defer rows.Close()

	return scanMeasurements(rows)
}

// FindByClientIDAndType returns measurements for a client filtered by type and time.
func (r *PostgresMeasurementRepository) FindByClientIDAndType(ctx context.Context, clientID string, measurementType string, since time.Time) ([]*entities.Measurement, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, client_id, recorded_by, measurement_type, value, unit,
		        flag, ref_range_low, ref_range_high, COALESCE(artifact_id::TEXT, ''),
		        measured_at, COALESCE(notes, ''), created_at
		 FROM measurements
		 WHERE client_id = $1 AND measurement_type = $2 AND measured_at >= $3
		 ORDER BY measured_at ASC`,
		clientID, measurementType, since,
	)
	if err != nil {
		return nil, fmt.Errorf("query measurements by type: %w", err)
	}
	defer rows.Close()

	return scanMeasurements(rows)
}

// FindRecentByArtifactID returns measurements linked to a given artifact.
func (r *PostgresMeasurementRepository) FindRecentByArtifactID(ctx context.Context, artifactID string) ([]*entities.Measurement, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, client_id, recorded_by, measurement_type, value, unit,
		        flag, ref_range_low, ref_range_high, COALESCE(artifact_id::TEXT, ''),
		        measured_at, COALESCE(notes, ''), created_at
		 FROM measurements
		 WHERE artifact_id = $1
		 ORDER BY measured_at ASC`,
		artifactID,
	)
	if err != nil {
		return nil, fmt.Errorf("query measurements by artifact: %w", err)
	}
	defer rows.Close()

	return scanMeasurements(rows)
}

func scanMeasurements(rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Err() error
}) ([]*entities.Measurement, error) {
	var measurements []*entities.Measurement
	for rows.Next() {
		var m entities.Measurement
		if err := rows.Scan(
			&m.ID, &m.ClientID, &m.RecordedBy, &m.MeasurementType, &m.Value, &m.Unit,
			&m.Flag, &m.RefRangeLow, &m.RefRangeHigh, &m.ArtifactID,
			&m.MeasuredAt, &m.Notes, &m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan measurement: %w", err)
		}
		measurements = append(measurements, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate measurements: %w", err)
	}
	return measurements, nil
}
