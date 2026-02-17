package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestAsyncAuditRepository_LogEvent_QueuesAndProcesses(t *testing.T) {
	inner := NewInMemoryAuditRepository()
	async := NewAsyncAuditRepository(inner, 100)
	async.Start()
	defer async.Stop()

	err := async.LogEvent(context.Background(), &entities.AuditEvent{
		ActorID:    "user-1",
		Action:     "user.login",
		EntityType: "user",
		EntityID:   "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wait for async processing
	time.Sleep(100 * time.Millisecond)

	events, total, err := inner.Query(context.Background(), entities.AuditQueryFilter{Limit: 100})
	if err != nil {
		t.Fatalf("unexpected error querying events: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected 1 event, got %d", total)
	}
	if events[0].Action != "user.login" {
		t.Errorf("expected action 'user.login', got '%s'", events[0].Action)
	}
}

func TestAsyncAuditRepository_LogEvent_MultipleEvents_AllProcessed(t *testing.T) {
	inner := NewInMemoryAuditRepository()
	async := NewAsyncAuditRepository(inner, 100)
	async.Start()

	for i := 0; i < 10; i++ {
		_ = async.LogEvent(context.Background(), &entities.AuditEvent{
			ActorID:    "user-1",
			Action:     "user.login",
			EntityType: "user",
			EntityID:   "user-1",
		})
	}

	async.Stop() // Stop waits for all events to drain

	events, total, err := inner.Query(context.Background(), entities.AuditQueryFilter{Limit: 100})
	if err != nil {
		t.Fatalf("unexpected error querying events: %v", err)
	}
	if total != 10 {
		t.Errorf("expected 10 events, got %d", total)
	}
	if len(events) != 10 {
		t.Errorf("expected 10 events in result, got %d", len(events))
	}
}

func TestAsyncAuditRepository_Stop_DrainsBuffer(t *testing.T) {
	inner := NewInMemoryAuditRepository()
	async := NewAsyncAuditRepository(inner, 100)
	async.Start()

	for i := 0; i < 5; i++ {
		_ = async.LogEvent(context.Background(), &entities.AuditEvent{
			ActorID:    "user-1",
			Action:     "artifact.upload",
			EntityType: "artifact",
			EntityID:   "art-1",
		})
	}

	async.Stop()

	events, total, err := inner.Query(context.Background(), entities.AuditQueryFilter{Limit: 100})
	if err != nil {
		t.Fatalf("unexpected error querying events: %v", err)
	}
	if total != 5 {
		t.Errorf("expected 5 events after drain, got %d", total)
	}
	if len(events) != 5 {
		t.Errorf("expected 5 events in result, got %d", len(events))
	}
}

func TestAsyncAuditRepository_Query_DelegatesSynchronously(t *testing.T) {
	inner := NewInMemoryAuditRepository()
	_ = inner.LogEvent(context.Background(), &entities.AuditEvent{
		ActorID:    "user-1",
		Action:     "user.login",
		EntityType: "user",
		EntityID:   "user-1",
	})

	async := NewAsyncAuditRepository(inner, 100)
	async.Start()
	defer async.Stop()

	events, total, err := async.Query(context.Background(), entities.AuditQueryFilter{
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

func TestAsyncAuditRepository_ConcurrentLogging_ThreadSafe(t *testing.T) {
	inner := NewInMemoryAuditRepository()
	async := NewAsyncAuditRepository(inner, 1000)
	async.Start()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = async.LogEvent(context.Background(), &entities.AuditEvent{
				ActorID:    "user-1",
				Action:     "user.login",
				EntityType: "user",
				EntityID:   "user-1",
			})
		}()
	}
	wg.Wait()
	async.Stop()

	events, total, err := inner.Query(context.Background(), entities.AuditQueryFilter{Limit: 200})
	if err != nil {
		t.Fatalf("unexpected error querying events: %v", err)
	}
	if total != 100 {
		t.Errorf("expected 100 events, got %d", total)
	}
	if len(events) != 100 {
		t.Errorf("expected 100 events in result, got %d", len(events))
	}
}

func TestAsyncAuditRepository_BufferFull_FallsBackToSync(t *testing.T) {
	inner := NewInMemoryAuditRepository()
	// Very small buffer to trigger fallback
	async := NewAsyncAuditRepository(inner, 1)
	// Don't start the processor — events will pile up and fall back to sync

	// First event fills the buffer
	_ = async.LogEvent(context.Background(), &entities.AuditEvent{
		ActorID:    "user-1",
		Action:     "user.login",
		EntityType: "user",
		EntityID:   "user-1",
	})

	// Second event should fall back to sync (buffer full, no processor running)
	err := async.LogEvent(context.Background(), &entities.AuditEvent{
		ActorID:    "user-2",
		Action:     "user.login",
		EntityType: "user",
		EntityID:   "user-2",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The sync fallback should have written directly
	events, total, err := inner.Query(context.Background(), entities.AuditQueryFilter{Limit: 100})
	if err != nil {
		t.Fatalf("unexpected error querying events: %v", err)
	}
	if total < 1 {
		t.Error("expected at least one event to be logged synchronously")
	}
	if len(events) < 1 {
		t.Error("expected at least one event in result")
	}
}

func TestAsyncAuditRepository_DoubleStart_NoPanic(t *testing.T) {
	inner := NewInMemoryAuditRepository()
	async := NewAsyncAuditRepository(inner, 100)
	async.Start()
	async.Start() // Should not panic
	async.Stop()
}

func TestAsyncAuditRepository_DoubleStop_NoPanic(t *testing.T) {
	inner := NewInMemoryAuditRepository()
	async := NewAsyncAuditRepository(inner, 100)
	async.Start()
	async.Stop()
	async.Stop() // Should not panic
}
