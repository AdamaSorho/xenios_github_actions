package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

func newTestInsight(coachID, clientID, status, priority string) *entities.InsightCard {
	ic := entities.NewInsightCard(coachID, clientID, "Test Title", "Test Body", "general", priority)
	ic.Status = status
	return ic
}

func TestInMemoryInsightCard_Create_And_FindByID(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	ctx := context.Background()

	ic := newTestInsight("coach-1", "client-1", "draft", "medium")
	if err := repo.Create(ctx, ic); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ic.ID == "" {
		t.Fatal("expected ID to be set")
	}

	found, err := repo.FindByID(ctx, ic.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil {
		t.Fatal("expected non-nil result")
	}
	if found.CoachID != "coach-1" {
		t.Errorf("expected CoachID coach-1, got %s", found.CoachID)
	}
}

func TestInMemoryInsightCard_FindByID_NotFound(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	found, err := repo.FindByID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Error("expected nil for nonexistent ID")
	}
}

func TestInMemoryInsightCard_Update(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	ctx := context.Background()

	ic := newTestInsight("coach-1", "client-1", "draft", "medium")
	_ = repo.Create(ctx, ic)

	ic.Title = "Updated Title"
	if err := repo.Update(ctx, ic); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, _ := repo.FindByID(ctx, ic.ID)
	if found.Title != "Updated Title" {
		t.Errorf("expected 'Updated Title', got %s", found.Title)
	}
}

func TestInMemoryInsightCard_Update_NotFound(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	ic := &entities.InsightCard{ID: "nonexistent"}
	err := repo.Update(context.Background(), ic)
	if err == nil {
		t.Fatal("expected error for nonexistent insight")
	}
}

func TestInMemoryInsightCard_ListByCoach_FilterByStatus(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	ctx := context.Background()

	// Create insights with different statuses
	draft := newTestInsight("coach-1", "client-1", "draft", "medium")
	approved := newTestInsight("coach-1", "client-2", "approved", "high")
	other := newTestInsight("coach-2", "client-3", "draft", "low")

	_ = repo.Create(ctx, draft)
	_ = repo.Create(ctx, approved)
	_ = repo.Create(ctx, other)

	// List only draft for coach-1
	results, total, err := repo.ListByCoach(ctx, repository.InsightCardFilter{
		CoachID: "coach-1",
		Status:  "draft",
		Limit:   20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestInMemoryInsightCard_ListByCoach_AllStatuses(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	ctx := context.Background()

	_ = repo.Create(ctx, newTestInsight("coach-1", "client-1", "draft", "medium"))
	_ = repo.Create(ctx, newTestInsight("coach-1", "client-2", "approved", "high"))

	results, total, err := repo.ListByCoach(ctx, repository.InsightCardFilter{
		CoachID: "coach-1",
		Limit:   20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestInMemoryInsightCard_ListByCoach_Pagination(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		ic := newTestInsight("coach-1", "client-1", "draft", "medium")
		// Offset creation times so order is deterministic
		ic.CreatedAt = time.Now().Add(time.Duration(i) * time.Second)
		_ = repo.Create(ctx, ic)
	}

	results, total, _ := repo.ListByCoach(ctx, repository.InsightCardFilter{
		CoachID: "coach-1",
		Status:  "draft",
		Limit:   2,
		Offset:  0,
	})
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results on page 1, got %d", len(results))
	}

	results2, _, _ := repo.ListByCoach(ctx, repository.InsightCardFilter{
		CoachID: "coach-1",
		Status:  "draft",
		Limit:   2,
		Offset:  4,
	})
	if len(results2) != 1 {
		t.Errorf("expected 1 result on last page, got %d", len(results2))
	}
}

func TestInMemoryInsightCard_ListByCoach_PrioritySorting(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	ctx := context.Background()

	low := newTestInsight("coach-1", "client-1", "draft", "low")
	urgent := newTestInsight("coach-1", "client-2", "draft", "urgent")
	high := newTestInsight("coach-1", "client-3", "draft", "high")

	_ = repo.Create(ctx, low)
	_ = repo.Create(ctx, urgent)
	_ = repo.Create(ctx, high)

	results, _, _ := repo.ListByCoach(ctx, repository.InsightCardFilter{
		CoachID: "coach-1",
		Status:  "draft",
		Limit:   20,
	})

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	// Urgent should be first
	if results[0].Priority != "urgent" {
		t.Errorf("expected first result to be urgent, got %s", results[0].Priority)
	}
	if results[1].Priority != "high" {
		t.Errorf("expected second result to be high, got %s", results[1].Priority)
	}
	if results[2].Priority != "low" {
		t.Errorf("expected third result to be low, got %s", results[2].Priority)
	}
}

func TestInMemoryInsightCard_ListByClient(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	ctx := context.Background()

	_ = repo.Create(ctx, newTestInsight("coach-1", "client-1", "draft", "medium"))
	_ = repo.Create(ctx, newTestInsight("coach-1", "client-1", "approved", "high"))
	_ = repo.Create(ctx, newTestInsight("coach-1", "client-2", "draft", "low"))

	results, total, err := repo.ListByClient(ctx, repository.InsightCardFilter{
		ClientID: "client-1",
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestInMemoryInsightCard_ListByClient_FilterStatus(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	ctx := context.Background()

	_ = repo.Create(ctx, newTestInsight("coach-1", "client-1", "draft", "medium"))
	_ = repo.Create(ctx, newTestInsight("coach-1", "client-1", "approved", "high"))

	results, total, _ := repo.ListByClient(ctx, repository.InsightCardFilter{
		ClientID: "client-1",
		Status:   "approved",
		Limit:    20,
	})
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}
