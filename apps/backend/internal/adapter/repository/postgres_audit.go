package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xenios/backend/internal/domain/entities"
)

// PostgresAuditRepository implements AuditRepository with PostgreSQL.
type PostgresAuditRepository struct {
	db *pgxpool.Pool
}

// NewPostgresAuditRepository creates a new PostgresAuditRepository.
func NewPostgresAuditRepository(db *pgxpool.Pool) *PostgresAuditRepository {
	return &PostgresAuditRepository{db: db}
}

// LogEvent inserts an audit event into the events_audit table.
func (r *PostgresAuditRepository) LogEvent(ctx context.Context, event *entities.AuditEvent) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO events_audit (actor_id, action, entity_type, entity_id, metadata, ip_address, user_agent)
		 VALUES ($1, $2, $3, $4, $5, $6::INET, $7)`,
		event.ActorID,
		event.Action,
		event.EntityType,
		event.EntityID,
		event.Metadata,
		nilIfEmpty(event.IPAddress),
		nilIfEmpty(event.UserAgent),
	)
	if err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}
	return nil
}

// Query retrieves audit events matching the given filter.
func (r *PostgresAuditRepository) Query(ctx context.Context, filter entities.AuditQueryFilter) ([]*entities.AuditEvent, int, error) {
	var conditions []string
	var args []interface{}
	paramIdx := 1

	if filter.ActorID != "" {
		conditions = append(conditions, fmt.Sprintf("actor_id = $%d", paramIdx))
		args = append(args, filter.ActorID)
		paramIdx++
	}
	if filter.Action != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", paramIdx))
		args = append(args, filter.Action)
		paramIdx++
	}
	if filter.EntityType != "" {
		conditions = append(conditions, fmt.Sprintf("entity_type = $%d", paramIdx))
		args = append(args, filter.EntityType)
		paramIdx++
	}
	if filter.EntityID != "" {
		conditions = append(conditions, fmt.Sprintf("entity_id = $%d", paramIdx))
		args = append(args, filter.EntityID)
		paramIdx++
	}
	if filter.From != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", paramIdx))
		args = append(args, *filter.From)
		paramIdx++
	}
	if filter.To != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", paramIdx))
		args = append(args, *filter.To)
		paramIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total matching records
	countQuery := "SELECT COUNT(*) FROM events_audit" + where
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count audit events: %w", err)
	}

	// Fetch paginated results
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	dataQuery := fmt.Sprintf(
		`SELECT id, actor_id, action, entity_type, entity_id, metadata,
		        COALESCE(ip_address::TEXT, ''), COALESCE(user_agent, ''), created_at
		 FROM events_audit%s
		 ORDER BY created_at DESC
		 LIMIT $%d OFFSET $%d`,
		where, paramIdx, paramIdx+1,
	)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query audit events: %w", err)
	}
	defer rows.Close()

	var events []*entities.AuditEvent
	for rows.Next() {
		var e entities.AuditEvent
		if err := rows.Scan(
			&e.ID, &e.ActorID, &e.Action, &e.EntityType, &e.EntityID,
			&e.Metadata, &e.IPAddress, &e.UserAgent, &e.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan audit event: %w", err)
		}
		events = append(events, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate audit events: %w", err)
	}

	return events, total, nil
}

func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
