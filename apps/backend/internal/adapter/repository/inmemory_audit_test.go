package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryAuditRepository_LogEvent_StoresEvent(t *testing.T) {
	repo := NewInMemoryAuditRepository()
	err := repo.LogEvent(context.Background(), &entities.AuditEvent{
		ActorID:    "user-1",
		Action:     "user.login",
		EntityType: "user",
		EntityID:   "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(repo.Events))
	}
}

func TestInMemoryAuditRepository_LogEvent_AssignsIDAndTimestamp(t *testing.T) {
	repo := NewInMemoryAuditRepository()
	err := repo.LogEvent(context.Background(), &entities.AuditEvent{
		ActorID:    "user-1",
		Action:     "user.login",
		EntityType: "user",
		EntityID:   "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.Events[0].ID == "" {
		t.Error("expected non-empty ID")
	}
	if repo.Events[0].CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestInMemoryAuditRepository_LogEvent_PreservesExistingTimestamp(t *testing.T) {
	repo := NewInMemoryAuditRepository()
	ts := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	err := repo.LogEvent(context.Background(), &entities.AuditEvent{
		ActorID:    "user-1",
		Action:     "user.login",
		EntityType: "user",
		EntityID:   "user-1",
		CreatedAt:  ts,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.Events[0].CreatedAt.Equal(ts) {
		t.Errorf("expected preserved timestamp %v, got %v", ts, repo.Events[0].CreatedAt)
	}
}

func TestInMemoryAuditRepository_Query_AllEvents(t *testing.T) {
	repo := NewInMemoryAuditRepository()
	_ = repo.LogEvent(context.Background(), &entities.AuditEvent{
		ActorID: "user-1", Action: "user.login", EntityType: "user", EntityID: "user-1",
	})
	_ = repo.LogEvent(context.Background(), &entities.AuditEvent{
		ActorID: "user-2", Action: "client.view", EntityType: "client_profile", EntityID: "client-1",
	})

	events, total, err := repo.Query(context.Background(), entities.AuditQueryFilter{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}

func TestInMemoryAuditRepository_Query_FilterByActorID(t *testing.T) {
	repo := NewInMemoryAuditRepository()
	_ = repo.LogEvent(context.Background(), &entities.AuditEvent{
		ActorID: "user-1", Action: "user.login", EntityType: "user", EntityID: "user-1",
	})
	_ = repo.LogEvent(context.Background(), &entities.AuditEvent{
		ActorID: "user-2", Action: "user.login", EntityType: "user", EntityID: "user-2",
	})

	events, total, err := repo.Query(context.Background(), entities.AuditQueryFilter{
		ActorID: "user-1",
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
}

func TestInMemoryAuditRepository_Query_Pagination(t *testing.T) {
	repo := NewInMemoryAuditRepository()
	for i := 0; i < 5; i++ {
		_ = repo.LogEvent(context.Background(), &entities.AuditEvent{
			ActorID: "user-1", Action: "user.login", EntityType: "user", EntityID: "user-1",
		})
	}

	events, total, err := repo.Query(context.Background(), entities.AuditQueryFilter{
		Limit:  2,
		Offset: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events on page, got %d", len(events))
	}
}

func TestInMemoryAuditRepository_Query_DefaultLimit(t *testing.T) {
	repo := NewInMemoryAuditRepository()
	for i := 0; i < 60; i++ {
		_ = repo.LogEvent(context.Background(), &entities.AuditEvent{
			ActorID: "user-1", Action: "user.login", EntityType: "user", EntityID: "user-1",
		})
	}

	events, total, err := repo.Query(context.Background(), entities.AuditQueryFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 60 {
		t.Errorf("expected total 60, got %d", total)
	}
	// Default limit is 50
	if len(events) != 50 {
		t.Errorf("expected 50 events with default limit, got %d", len(events))
	}
}

func TestInMemoryAuditRepository_Query_FilterByTimeRange(t *testing.T) {
	repo := NewInMemoryAuditRepository()

	now := time.Now()
	_ = repo.LogEvent(context.Background(), &entities.AuditEvent{
		ActorID: "user-1", Action: "user.login", EntityType: "user", EntityID: "user-1",
		CreatedAt: now.Add(-2 * time.Hour),
	})
	_ = repo.LogEvent(context.Background(), &entities.AuditEvent{
		ActorID: "user-1", Action: "client.view", EntityType: "client_profile", EntityID: "client-1",
		CreatedAt: now.Add(-1 * time.Hour),
	})
	_ = repo.LogEvent(context.Background(), &entities.AuditEvent{
		ActorID: "user-1", Action: "user.logout", EntityType: "user", EntityID: "user-1",
		CreatedAt: now,
	})

	from := now.Add(-90 * time.Minute)
	to := now.Add(-30 * time.Minute)
	events, total, err := repo.Query(context.Background(), entities.AuditQueryFilter{
		From:  &from,
		To:    &to,
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
}
