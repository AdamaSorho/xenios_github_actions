package repository

import (
	"context"
	"fmt"

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

// Create inserts a new measurement.
func (r *PostgresMeasurementRepository) Create(ctx context.Context, m *entities.Measurement) (*entities.Measurement, error) {
	err := r.db.QueryRow(ctx,
		`INSERT INTO measurements (client_id, recorded_by, measurement_type, value, unit, measured_at, notes, artifact_id, flag, reference_low, reference_high)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING id, created_at`,
		m.ClientID, m.RecordedBy, m.Type, m.Value, m.Unit, m.MeasuredAt,
		m.Notes, m.ArtifactID, m.Flag, m.ReferenceLow, m.ReferenceHigh,
	).Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert measurement: %w", err)
	}
	return m, nil
}

// FindByClientID retrieves measurements with filtering and pagination.
func (r *PostgresMeasurementRepository) FindByClientID(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
	qb := newQueryBuilder()

	qb.addCondition("client_id", filter.ClientID)

	if filter.Type != "" {
		qb.addCondition("measurement_type", filter.Type)
	}
	if filter.From != nil {
		qb.addTimeCondition("measured_at", ">=", *filter.From)
	}
	if filter.To != nil {
		qb.addTimeCondition("measured_at", "<=", *filter.To)
	}

	// Count total
	countSQL := "SELECT COUNT(*) FROM measurements" + qb.whereClause()
	var total int
	if err := r.db.QueryRow(ctx, countSQL, qb.args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count measurements: %w", err)
	}

	// Fetch paginated results
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	idx := qb.nextParamIdx()
	selectSQL := fmt.Sprintf(
		`SELECT id, client_id, measurement_type, value, unit, measured_at,
		        artifact_id, flag, reference_low, reference_high, notes, recorded_by, created_at
		 FROM measurements%s
		 ORDER BY measured_at DESC
		 LIMIT $%d OFFSET $%d`,
		qb.whereClause(), idx, idx+1,
	)

	args := append(qb.args, limit, offset)
	rows, err := r.db.Query(ctx, selectSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query measurements: %w", err)
	}
	defer rows.Close()

	var measurements []*entities.Measurement
	for rows.Next() {
		var m entities.Measurement
		if err := rows.Scan(
			&m.ID, &m.ClientID, &m.Type, &m.Value, &m.Unit, &m.MeasuredAt,
			&m.ArtifactID, &m.Flag, &m.ReferenceLow, &m.ReferenceHigh, &m.Notes, &m.RecordedBy, &m.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan measurement: %w", err)
		}
		measurements = append(measurements, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate measurements: %w", err)
	}

	return measurements, total, nil
}

// FindLatestByClientID returns the most recent measurement for each type.
func (r *PostgresMeasurementRepository) FindLatestByClientID(ctx context.Context, clientID string) ([]*entities.Measurement, error) {
	rows, err := r.db.Query(ctx,
		`SELECT DISTINCT ON (measurement_type)
		        id, client_id, measurement_type, value, unit, measured_at,
		        artifact_id, flag, reference_low, reference_high, notes, recorded_by, created_at
		 FROM measurements
		 WHERE client_id = $1
		 ORDER BY measurement_type, measured_at DESC`,
		clientID,
	)
	if err != nil {
		return nil, fmt.Errorf("query latest measurements: %w", err)
	}
	defer rows.Close()

	var measurements []*entities.Measurement
	for rows.Next() {
		var m entities.Measurement
		if err := rows.Scan(
			&m.ID, &m.ClientID, &m.Type, &m.Value, &m.Unit, &m.MeasuredAt,
			&m.ArtifactID, &m.Flag, &m.ReferenceLow, &m.ReferenceHigh, &m.Notes, &m.RecordedBy, &m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan latest measurement: %w", err)
		}
		measurements = append(measurements, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate latest measurements: %w", err)
	}

	return measurements, nil
}

// FindByType retrieves measurements for a client filtered by measurement type.
func (r *PostgresMeasurementRepository) FindByType(ctx context.Context, clientID, measurementType string) ([]*entities.Measurement, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, client_id, measurement_type, value, unit, measured_at,
		        artifact_id, flag, reference_low, reference_high, notes, recorded_by, created_at
		 FROM measurements
		 WHERE client_id = $1 AND measurement_type = $2
		 ORDER BY measured_at DESC`,
		clientID, measurementType,
	)
	if err != nil {
		return nil, fmt.Errorf("query measurements by type: %w", err)
	}
	defer rows.Close()

	var measurements []*entities.Measurement
	for rows.Next() {
		var m entities.Measurement
		if err := rows.Scan(
			&m.ID, &m.ClientID, &m.Type, &m.Value, &m.Unit, &m.MeasuredAt,
			&m.ArtifactID, &m.Flag, &m.ReferenceLow, &m.ReferenceHigh, &m.Notes, &m.RecordedBy, &m.CreatedAt,
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
