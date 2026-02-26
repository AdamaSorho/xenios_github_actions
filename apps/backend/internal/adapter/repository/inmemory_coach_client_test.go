package repository

import (
	"context"
	"sync"
	"testing"

	"github.com/xenios/backend/internal/domain/repository"
)

// Verify InMemoryCoachClientRepository implements CoachClientRepository
var _ repository.CoachClientRepository = (*InMemoryCoachClientRepository)(nil)

func TestInMemoryCoachClientRepository_Create(t *testing.T) {
	repo := NewInMemoryCoachClientRepository()

	cc, err := repo.Create(context.Background(), "coach-1", "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cc.ID == "" {
		t.Error("expected non-empty ID")
	}
	if cc.CoachID != "coach-1" {
		t.Errorf("expected CoachID coach-1, got %s", cc.CoachID)
	}
	if cc.ClientID != "client-1" {
		t.Errorf("expected ClientID client-1, got %s", cc.ClientID)
	}
	if cc.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestInMemoryCoachClientRepository_ListByCoachID(t *testing.T) {
	repo := NewInMemoryCoachClientRepository()
	ctx := context.Background()

	// Create relationships for two coaches
	_, _ = repo.Create(ctx, "coach-1", "client-a")
	_, _ = repo.Create(ctx, "coach-1", "client-b")
	_, _ = repo.Create(ctx, "coach-2", "client-c")

	// List coach-1's clients
	results, err := repo.ListByCoachID(ctx, "coach-1", 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results for coach-1, got %d", len(results))
	}

	// List coach-2's clients
	results, err = repo.ListByCoachID(ctx, "coach-2", 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result for coach-2, got %d", len(results))
	}
}

func TestInMemoryCoachClientRepository_ListByCoachID_Pagination(t *testing.T) {
	repo := NewInMemoryCoachClientRepository()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_, _ = repo.Create(ctx, "coach-1", "client-"+string(rune('a'+i)))
	}

	// First page
	results, err := repo.ListByCoachID(ctx, "coach-1", 2, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// Second page
	results, err = repo.ListByCoachID(ctx, "coach-1", 2, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// Last page
	results, err = repo.ListByCoachID(ctx, "coach-1", 2, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestInMemoryCoachClientRepository_ListByCoachID_OffsetBeyondResults(t *testing.T) {
	repo := NewInMemoryCoachClientRepository()
	ctx := context.Background()

	_, _ = repo.Create(ctx, "coach-1", "client-1")

	results, err := repo.ListByCoachID(ctx, "coach-1", 20, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestInMemoryCoachClientRepository_ListByCoachID_NoResults(t *testing.T) {
	repo := NewInMemoryCoachClientRepository()

	results, err := repo.ListByCoachID(context.Background(), "nonexistent", 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestInMemoryCoachClientRepository_ConcurrentAccess(t *testing.T) {
	repo := NewInMemoryCoachClientRepository()
	ctx := context.Background()
	var wg sync.WaitGroup

	// Concurrent creates
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _ = repo.Create(ctx, "coach-1", "client-concurrent")
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = repo.ListByCoachID(ctx, "coach-1", 20, 0)
		}()
	}

	wg.Wait()

	results, err := repo.ListByCoachID(ctx, "coach-1", 100, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 10 {
		t.Errorf("expected 10 results after concurrent writes, got %d", len(results))
	}
}

func TestInMemoryCoachClientRepository_FindByCoachAndClient_Found(t *testing.T) {
	repo := NewInMemoryCoachClientRepository()
	ctx := context.Background()

	_, _ = repo.Create(ctx, "coach-1", "client-1")

	result, err := repo.FindByCoachAndClient(ctx, "coach-1", "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.CoachID != "coach-1" {
		t.Errorf("expected CoachID coach-1, got %s", result.CoachID)
	}
	if result.ClientID != "client-1" {
		t.Errorf("expected ClientID client-1, got %s", result.ClientID)
	}
}

func TestInMemoryCoachClientRepository_FindByCoachAndClient_NotFound(t *testing.T) {
	repo := NewInMemoryCoachClientRepository()
	ctx := context.Background()

	result, err := repo.FindByCoachAndClient(ctx, "coach-1", "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil result for non-existent relationship")
	}
}

func TestInMemoryCoachClientRepository_FindByCoachAndClient_WrongCoach(t *testing.T) {
	repo := NewInMemoryCoachClientRepository()
	ctx := context.Background()

	_, _ = repo.Create(ctx, "coach-1", "client-1")

	result, err := repo.FindByCoachAndClient(ctx, "coach-2", "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil result for wrong coach")
	}
}

func TestInMemoryCoachClientRepository_UniqueIDs(t *testing.T) {
	repo := NewInMemoryCoachClientRepository()
	ctx := context.Background()
	ids := make(map[string]bool)

	for i := 0; i < 100; i++ {
		cc, err := repo.Create(ctx, "coach-1", "client-1")
		if err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}
		if ids[cc.ID] {
			t.Errorf("duplicate ID generated: %s", cc.ID)
		}
		ids[cc.ID] = true
	}
}
