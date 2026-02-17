package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xenios/backend/internal/domain/entities"
)

// testUserIDs holds UUIDs of test users for foreign key constraints.
type testUserIDs struct {
	user1 string
	user2 string
}

// setupAuditTestDB creates a test database connection, cleans up the events_audit table,
// and creates test users for foreign key constraints.
func setupAuditTestDB(t *testing.T) (*pgxpool.Pool, *testUserIDs, func()) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to create connection pool: %v", err)
	}

	// Clean up events_audit table before tests
	// Note: We can't use DELETE because of append-only triggers, so we truncate
	_, err = pool.Exec(ctx, "TRUNCATE TABLE events_audit CASCADE")
	if err != nil {
		pool.Close()
		t.Fatalf("failed to clean events_audit table: %v", err)
	}

	// Create test users for foreign key constraints
	var user1ID, user2ID string
	err = pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, name, role)
		VALUES ('audit-test-1@example.com', 'hash', 'Test User 1', 'client')
		ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
		RETURNING id
	`).Scan(&user1ID)
	if err != nil {
		pool.Close()
		t.Fatalf("failed to create test user 1: %v", err)
	}

	err = pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, name, role)
		VALUES ('audit-test-2@example.com', 'hash', 'Test User 2', 'client')
		ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
		RETURNING id
	`).Scan(&user2ID)
	if err != nil {
		pool.Close()
		t.Fatalf("failed to create test user 2: %v", err)
	}

	cleanup := func() {
		pool.Close()
	}

	return pool, &testUserIDs{user1: user1ID, user2: user2ID}, cleanup
}

func TestPostgresAuditRepository_LogEvent_BasicEvent(t *testing.T) {
	pool, users, cleanup := setupAuditTestDB(t)
	defer cleanup()

	repo := NewPostgresAuditRepository(pool)
	ctx := context.Background()

	event := &entities.AuditEvent{
		ActorID:    users.user1,
		Action:     "user.login",
		EntityType: "user",
		EntityID:   users.user1,
	}

	err := repo.LogEvent(ctx, event)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify the event was inserted
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM events_audit WHERE actor_id = $1", users.user1).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count events: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 event, got %d", count)
	}
}

func TestPostgresAuditRepository_LogEvent_WithMetadata(t *testing.T) {
	pool, users, cleanup := setupAuditTestDB(t)
	defer cleanup()

	repo := NewPostgresAuditRepository(pool)
	ctx := context.Background()

	event := &entities.AuditEvent{
		ActorID:    users.user1,
		Action:     "artifact.upload",
		EntityType: "artifact",
		EntityID:   users.user2,
		Metadata: map[string]interface{}{
			"file_name": "report.pdf",
			"file_size": float64(1024), // JSON numbers are float64
		},
	}

	err := repo.LogEvent(ctx, event)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify metadata was stored
	var metadata map[string]interface{}
	err = pool.QueryRow(ctx, "SELECT metadata FROM events_audit WHERE entity_id = $1", users.user2).Scan(&metadata)
	if err != nil {
		t.Fatalf("failed to query metadata: %v", err)
	}
	if metadata["file_name"] != "report.pdf" {
		t.Errorf("expected file_name 'report.pdf', got '%v'", metadata["file_name"])
	}
}

func TestPostgresAuditRepository_LogEvent_WithIPAndUserAgent(t *testing.T) {
	pool, users, cleanup := setupAuditTestDB(t)
	defer cleanup()

	repo := NewPostgresAuditRepository(pool)
	ctx := context.Background()

	event := &entities.AuditEvent{
		ActorID:    users.user1,
		Action:     "user.login",
		EntityType: "user",
		EntityID:   users.user1,
		IPAddress:  "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
	}

	err := repo.LogEvent(ctx, event)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify IP and user agent were stored
	var ipAddress, userAgent string
	err = pool.QueryRow(ctx, "SELECT ip_address::TEXT, user_agent FROM events_audit WHERE entity_id = $1 AND action = 'user.login'", users.user1).Scan(&ipAddress, &userAgent)
	if err != nil {
		t.Fatalf("failed to query IP and user agent: %v", err)
	}
	if ipAddress != "192.168.1.1" {
		t.Errorf("expected IP '192.168.1.1', got '%s'", ipAddress)
	}
	if userAgent != "Mozilla/5.0" {
		t.Errorf("expected user agent 'Mozilla/5.0', got '%s'", userAgent)
	}
}

func TestPostgresAuditRepository_Query_AllEvents(t *testing.T) {
	pool, users, cleanup := setupAuditTestDB(t)
	defer cleanup()

	repo := NewPostgresAuditRepository(pool)
	ctx := context.Background()

	// Insert test events
	for i := 0; i < 5; i++ {
		event := &entities.AuditEvent{
			ActorID:    users.user1,
			Action:     "user.login",
			EntityType: "user",
			EntityID:   users.user1,
		}
		err := repo.LogEvent(ctx, event)
		if err != nil {
			t.Fatalf("failed to log event: %v", err)
		}
	}

	// Query all events
	filter := entities.AuditQueryFilter{
		Limit: 50,
	}
	events, total, err := repo.Query(ctx, filter)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if total < 5 {
		t.Errorf("expected at least 5 events, got %d", total)
	}
	if len(events) < 5 {
		t.Errorf("expected at least 5 events returned, got %d", len(events))
	}
}

func TestPostgresAuditRepository_Query_FilterByActorID(t *testing.T) {
	pool, users, cleanup := setupAuditTestDB(t)
	defer cleanup()

	repo := NewPostgresAuditRepository(pool)
	ctx := context.Background()

	// Insert events with different actors
	repo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    users.user1,
		Action:     "test.action",
		EntityType: "test",
		EntityID:   users.user1,
	})
	repo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    users.user2,
		Action:     "test.action",
		EntityType: "test",
		EntityID:   users.user2,
	})

	// Query by actor_id
	filter := entities.AuditQueryFilter{
		ActorID: users.user1,
		Limit:   50,
	}
	events, total, err := repo.Query(ctx, filter)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if total != 1 {
		t.Errorf("expected 1 event, got %d", total)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event returned, got %d", len(events))
	}
	if events[0].ActorID != users.user1 {
		t.Errorf("expected actor_id '%s', got '%s'", users.user1, events[0].ActorID)
	}
}

func TestPostgresAuditRepository_Query_FilterByAction(t *testing.T) {
	pool, users, cleanup := setupAuditTestDB(t)
	defer cleanup()

	repo := NewPostgresAuditRepository(pool)
	ctx := context.Background()

	// Insert events with different actions
	repo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    users.user1,
		Action:     "user.login",
		EntityType: "user",
		EntityID:   users.user1,
	})
	repo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    users.user1,
		Action:     "user.logout",
		EntityType: "user",
		EntityID:   users.user1,
	})

	// Query by action
	filter := entities.AuditQueryFilter{
		Action: "user.login",
		Limit:  50,
	}
	events, total, err := repo.Query(ctx, filter)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if total < 1 {
		t.Errorf("expected at least 1 event, got %d", total)
	}
	for _, event := range events {
		if event.Action != "user.login" {
			t.Errorf("expected action 'user.login', got '%s'", event.Action)
		}
	}
}

func TestPostgresAuditRepository_Query_FilterByEntityType(t *testing.T) {
	pool, users, cleanup := setupAuditTestDB(t)
	defer cleanup()

	repo := NewPostgresAuditRepository(pool)
	ctx := context.Background()

	// Insert events with different entity types
	repo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    users.user1,
		Action:     "create",
		EntityType: "artifact",
		EntityID:   users.user1,
	})
	repo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    users.user1,
		Action:     "create",
		EntityType: "user",
		EntityID:   users.user2,
	})

	// Query by entity_type
	filter := entities.AuditQueryFilter{
		EntityType: "artifact",
		Limit:      50,
	}
	events, total, err := repo.Query(ctx, filter)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if total < 1 {
		t.Errorf("expected at least 1 event, got %d", total)
	}
	for _, event := range events {
		if event.EntityType != "artifact" {
			t.Errorf("expected entity_type 'artifact', got '%s'", event.EntityType)
		}
	}
}

func TestPostgresAuditRepository_Query_FilterByEntityID(t *testing.T) {
	pool, users, cleanup := setupAuditTestDB(t)
	defer cleanup()

	repo := NewPostgresAuditRepository(pool)
	ctx := context.Background()

	// Insert events with different entity IDs
	repo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    users.user1,
		Action:     "view",
		EntityType: "artifact",
		EntityID:   users.user1,
	})
	repo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    users.user1,
		Action:     "view",
		EntityType: "artifact",
		EntityID:   users.user2,
	})

	// Query by entity_id
	filter := entities.AuditQueryFilter{
		EntityID: users.user1,
		Limit:    50,
	}
	events, total, err := repo.Query(ctx, filter)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if total < 1 {
		t.Errorf("expected at least 1 event, got %d", total)
	}
	for _, event := range events {
		if event.EntityID != users.user1 {
			t.Errorf("expected entity_id '%s', got '%s'", users.user1, event.EntityID)
		}
	}
}

func TestPostgresAuditRepository_Query_FilterByTimeRange(t *testing.T) {
	pool, users, cleanup := setupAuditTestDB(t)
	defer cleanup()

	repo := NewPostgresAuditRepository(pool)
	ctx := context.Background()

	// Insert an event
	repo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    users.user1,
		Action:     "test.action",
		EntityType: "test",
		EntityID:   users.user1,
	})

	// Query with time range (past hour to future hour)
	now := time.Now()
	from := now.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)

	filter := entities.AuditQueryFilter{
		From:  &from,
		To:    &to,
		Limit: 50,
	}
	_, total, err := repo.Query(ctx, filter)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if total < 1 {
		t.Errorf("expected at least 1 event in time range, got %d", total)
	}

	// Query with time range in the past (should return 0)
	pastFrom := now.Add(-2 * time.Hour)
	pastTo := now.Add(-1 * time.Hour)
	filterPast := entities.AuditQueryFilter{
		From:  &pastFrom,
		To:    &pastTo,
		Limit: 50,
	}
	eventsPast, totalPast, err := repo.Query(ctx, filterPast)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if totalPast != 0 {
		t.Errorf("expected 0 events in past time range, got %d", totalPast)
	}
	if len(eventsPast) != 0 {
		t.Errorf("expected 0 events returned for past time range, got %d", len(eventsPast))
	}
}

func TestPostgresAuditRepository_Query_Pagination(t *testing.T) {
	pool, users, cleanup := setupAuditTestDB(t)
	defer cleanup()

	repo := NewPostgresAuditRepository(pool)
	ctx := context.Background()

	// Insert 10 test events
	for i := 0; i < 10; i++ {
		repo.LogEvent(ctx, &entities.AuditEvent{
			ActorID:    users.user1,
			Action:     "test.action",
			EntityType: "test",
			EntityID:   users.user1,
		})
	}

	// Query first page (limit 5, offset 0)
	filter := entities.AuditQueryFilter{
		ActorID: users.user1,
		Action:  "test.action",
		Limit:   5,
		Offset:  0,
	}
	events, total, err := repo.Query(ctx, filter)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if total != 10 {
		t.Errorf("expected total 10, got %d", total)
	}
	if len(events) != 5 {
		t.Errorf("expected 5 events in first page, got %d", len(events))
	}

	// Query second page (limit 5, offset 5)
	filter.Offset = 5
	events2, total2, err := repo.Query(ctx, filter)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if total2 != 10 {
		t.Errorf("expected total 10 on second page, got %d", total2)
	}
	if len(events2) != 5 {
		t.Errorf("expected 5 events in second page, got %d", len(events2))
	}
}

func TestPostgresAuditRepository_Query_DefaultLimit(t *testing.T) {
	pool, _, cleanup := setupAuditTestDB(t)
	defer cleanup()

	repo := NewPostgresAuditRepository(pool)
	ctx := context.Background()

	// Query with no limit (should default to 50)
	filter := entities.AuditQueryFilter{}
	_, _, err := repo.Query(ctx, filter)
	if err != nil {
		t.Fatalf("expected no error with default limit, got: %v", err)
	}
}

func TestPostgresAuditRepository_Query_NegativeOffset(t *testing.T) {
	pool, _, cleanup := setupAuditTestDB(t)
	defer cleanup()

	repo := NewPostgresAuditRepository(pool)
	ctx := context.Background()

	// Query with negative offset (should default to 0)
	filter := entities.AuditQueryFilter{
		Limit:  10,
		Offset: -5,
	}
	_, _, err := repo.Query(ctx, filter)
	if err != nil {
		t.Fatalf("expected no error with negative offset, got: %v", err)
	}
}

func TestPostgresAuditRepository_NilIfEmpty(t *testing.T) {
	// Test the nilIfEmpty helper function
	result := nilIfEmpty("")
	if result != nil {
		t.Errorf("expected nil for empty string, got %v", result)
	}

	result = nilIfEmpty("value")
	if result != "value" {
		t.Errorf("expected 'value', got %v", result)
	}
}
