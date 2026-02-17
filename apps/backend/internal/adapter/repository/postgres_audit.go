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

// auditQueryBuilder constructs parameterized SQL for querying the events_audit table.
// All filter values are passed as parameterized arguments ($1, $2, ...) to prevent injection.
type auditQueryBuilder struct {
	conditions []string
	args       []interface{}
	paramIdx   int
}

func newAuditQueryBuilder() *auditQueryBuilder {
	return &auditQueryBuilder{paramIdx: 1}
}

func (b *auditQueryBuilder) addCondition(column string, value interface{}) {
	b.conditions = append(b.conditions, fmt.Sprintf("%s = $%d", column, b.paramIdx))
	b.args = append(b.args, value)
	b.paramIdx++
}

func (b *auditQueryBuilder) addTimeCondition(column, op string, value interface{}) {
	b.conditions = append(b.conditions, fmt.Sprintf("%s %s $%d", column, op, b.paramIdx))
	b.args = append(b.args, value)
	b.paramIdx++
}

func (b *auditQueryBuilder) whereClause() string {
	if len(b.conditions) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(b.conditions, " AND ")
}

func (b *auditQueryBuilder) countSQL() string {
	var sb strings.Builder
	sb.WriteString("SELECT COUNT(*) FROM events_audit")
	sb.WriteString(b.whereClause())
	return sb.String()
}

func (b *auditQueryBuilder) selectSQL() string {
	var sb strings.Builder
	sb.WriteString("SELECT id, actor_id, action, entity_type, entity_id, metadata,")
	sb.WriteString(" COALESCE(ip_address::TEXT, ''), COALESCE(user_agent, ''), created_at")
	sb.WriteString(" FROM events_audit")
	sb.WriteString(b.whereClause())
	sb.WriteString(" ORDER BY created_at DESC")
	sb.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", b.paramIdx, b.paramIdx+1))
	return sb.String()
}

// Query retrieves audit events matching the given filter.
func (r *PostgresAuditRepository) Query(ctx context.Context, filter entities.AuditQueryFilter) ([]*entities.AuditEvent, int, error) {
	qb := newAuditQueryBuilder()

	if filter.ActorID != "" {
		qb.addCondition("actor_id", filter.ActorID)
	}
	if filter.Action != "" {
		qb.addCondition("action", filter.Action)
	}
	if filter.EntityType != "" {
		qb.addCondition("entity_type", filter.EntityType)
	}
	if filter.EntityID != "" {
		qb.addCondition("entity_id", filter.EntityID)
	}
	if filter.From != nil {
		qb.addTimeCondition("created_at", ">=", *filter.From)
	}
	if filter.To != nil {
		qb.addTimeCondition("created_at", "<=", *filter.To)
	}

	// Count total matching records
	var total int
	if err := r.db.QueryRow(ctx, qb.countSQL(), qb.args...).Scan(&total); err != nil {
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

	args := append(qb.args, limit, offset)
	rows, err := r.db.Query(ctx, qb.selectSQL(), args...)
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
