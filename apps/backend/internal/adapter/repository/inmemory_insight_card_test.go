package repository

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

func newTestInsight(coachID, clientID string, status entities.InsightStatus, priority entities.InsightPriority) *entities.InsightCard {
	return &entities.InsightCard{
		CoachID:    coachID,
		ClientID:   clientID,
		ClientName: "Test Client",
		Title:      "Test Insight",
		Body:       "Test body",
		Category:   "nutrition",
		Priority:   priority,
		Status:     status,
	}
}

func TestInMemoryInsightCard_Create_AssignsIDAndTimestamps(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	insight := newTestInsight("coach-1", "client-1", entities.InsightStatusDraft, entities.InsightPriorityMedium)

	created, err := repo.Create(context.Background(), insight)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == "" {
		t.Error("expected non-empty ID")
	}
	if created.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if created.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}

func TestInMemoryInsightCard_Create_WithExistingID_UsesProvidedID(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	insight := newTestInsight("coach-1", "client-1", entities.InsightStatusDraft, entities.InsightPriorityMedium)
	insight.ID = "custom-id"

	created, err := repo.Create(context.Background(), insight)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != "custom-id" {
		t.Errorf("expected ID 'custom-id', got %q", created.ID)
	}
}

func TestInMemoryInsightCard_FindByID_Exists_ReturnsInsight(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	insight := newTestInsight("coach-1", "client-1", entities.InsightStatusDraft, entities.InsightPriorityMedium)
	created, _ := repo.Create(context.Background(), insight)

	found, err := repo.FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil {
		t.Fatal("expected insight to be found")
	}
	if found.ID != created.ID {
		t.Errorf("expected ID %q, got %q", created.ID, found.ID)
	}
}

func TestInMemoryInsightCard_FindByID_NotExists_ReturnsNil(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()

	found, err := repo.FindByID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Error("expected nil for nonexistent insight")
	}
}

func TestInMemoryInsightCard_ListByCoachIDAndStatus_FiltersByCoachAndStatus(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	ctx := context.Background()

	repo.Create(ctx, newTestInsight("coach-1", "client-1", entities.InsightStatusDraft, entities.InsightPriorityMedium))
	repo.Create(ctx, newTestInsight("coach-1", "client-2", entities.InsightStatusDraft, entities.InsightPriorityHigh))
	repo.Create(ctx, newTestInsight("coach-1", "client-1", entities.InsightStatusApproved, entities.InsightPriorityMedium))
	repo.Create(ctx, newTestInsight("coach-2", "client-3", entities.InsightStatusDraft, entities.InsightPriorityMedium))

	results, total, err := repo.ListByCoachIDAndStatus(ctx, "coach-1", entities.InsightStatusDraft, 20, 0)
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

func TestInMemoryInsightCard_ListByCoachIDAndStatus_SortsByPriorityThenRecency(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	ctx := context.Background()

	repo.Create(ctx, newTestInsight("coach-1", "client-1", entities.InsightStatusDraft, entities.InsightPriorityLow))
	repo.Create(ctx, newTestInsight("coach-1", "client-2", entities.InsightStatusDraft, entities.InsightPriorityUrgent))
	repo.Create(ctx, newTestInsight("coach-1", "client-3", entities.InsightStatusDraft, entities.InsightPriorityHigh))

	results, _, err := repo.ListByCoachIDAndStatus(ctx, "coach-1", entities.InsightStatusDraft, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Priority != entities.InsightPriorityUrgent {
		t.Errorf("expected first result to be urgent, got %q", results[0].Priority)
	}
	if results[1].Priority != entities.InsightPriorityHigh {
		t.Errorf("expected second result to be high, got %q", results[1].Priority)
	}
}

func TestInMemoryInsightCard_ListByCoachIDAndStatus_Pagination(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		repo.Create(ctx, newTestInsight("coach-1", "client-1", entities.InsightStatusDraft, entities.InsightPriorityMedium))
	}

	results, total, err := repo.ListByCoachIDAndStatus(ctx, "coach-1", entities.InsightStatusDraft, 2, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestInMemoryInsightCard_ListByClientID_FiltersByClient(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	ctx := context.Background()

	repo.Create(ctx, newTestInsight("coach-1", "client-1", entities.InsightStatusDraft, entities.InsightPriorityMedium))
	repo.Create(ctx, newTestInsight("coach-1", "client-1", entities.InsightStatusApproved, entities.InsightPriorityHigh))
	repo.Create(ctx, newTestInsight("coach-1", "client-2", entities.InsightStatusDraft, entities.InsightPriorityMedium))

	results, total, err := repo.ListByClientID(ctx, "client-1", nil, 20, 0)
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

func TestInMemoryInsightCard_ListByClientID_FiltersByStatus(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	ctx := context.Background()

	repo.Create(ctx, newTestInsight("coach-1", "client-1", entities.InsightStatusDraft, entities.InsightPriorityMedium))
	repo.Create(ctx, newTestInsight("coach-1", "client-1", entities.InsightStatusApproved, entities.InsightPriorityHigh))

	status := entities.InsightStatusApproved
	results, total, err := repo.ListByClientID(ctx, "client-1", &status, 20, 0)
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

func TestInMemoryInsightCard_UpdateStatus_SetsStatusAndTimestamp(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	ctx := context.Background()

	created, _ := repo.Create(ctx, newTestInsight("coach-1", "client-1", entities.InsightStatusDraft, entities.InsightPriorityMedium))

	updated, err := repo.UpdateStatus(ctx, created.ID, entities.InsightStatusApproved)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != entities.InsightStatusApproved {
		t.Errorf("expected status 'approved', got %q", updated.Status)
	}
	if updated.ApprovedAt == nil {
		t.Error("expected ApprovedAt to be set")
	}
}

func TestInMemoryInsightCard_UpdateStatus_Dismissed_SetsTimestamp(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	ctx := context.Background()

	created, _ := repo.Create(ctx, newTestInsight("coach-1", "client-1", entities.InsightStatusDraft, entities.InsightPriorityMedium))

	updated, err := repo.UpdateStatus(ctx, created.ID, entities.InsightStatusDismissed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.DismissedAt == nil {
		t.Error("expected DismissedAt to be set")
	}
}

func TestInMemoryInsightCard_UpdateStatus_Shared_SetsTimestamp(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	ctx := context.Background()

	created, _ := repo.Create(ctx, newTestInsight("coach-1", "client-1", entities.InsightStatusApproved, entities.InsightPriorityMedium))

	updated, err := repo.UpdateStatus(ctx, created.ID, entities.InsightStatusShared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.SharedAt == nil {
		t.Error("expected SharedAt to be set")
	}
}

func TestInMemoryInsightCard_UpdateStatus_NotFound_ReturnsNil(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()

	result, err := repo.UpdateStatus(context.Background(), "nonexistent", entities.InsightStatusApproved)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil for nonexistent insight")
	}
}

func TestInMemoryInsightCard_UpdateContent_UpdatesTitleAndBody(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	ctx := context.Background()

	created, _ := repo.Create(ctx, newTestInsight("coach-1", "client-1", entities.InsightStatusDraft, entities.InsightPriorityMedium))

	updated, err := repo.UpdateContent(ctx, created.ID, "New Title", "New Body")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Title != "New Title" {
		t.Errorf("expected title 'New Title', got %q", updated.Title)
	}
	if updated.Body != "New Body" {
		t.Errorf("expected body 'New Body', got %q", updated.Body)
	}
}

func TestInMemoryInsightCard_UpdateContent_NotFound_ReturnsNil(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()

	result, err := repo.UpdateContent(context.Background(), "nonexistent", "T", "B")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil for nonexistent insight")
	}
}

func TestInMemoryInsightCard_CountByCoachIDAndStatus_ReturnsCount(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	ctx := context.Background()

	repo.Create(ctx, newTestInsight("coach-1", "client-1", entities.InsightStatusDraft, entities.InsightPriorityMedium))
	repo.Create(ctx, newTestInsight("coach-1", "client-2", entities.InsightStatusDraft, entities.InsightPriorityHigh))
	repo.Create(ctx, newTestInsight("coach-1", "client-1", entities.InsightStatusApproved, entities.InsightPriorityMedium))

	count, err := repo.CountByCoachIDAndStatus(ctx, "coach-1", entities.InsightStatusDraft)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}

func TestInMemoryInsightCard_CountByCoachIDAndStatus_NoMatches_ReturnsZero(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()

	count, err := repo.CountByCoachIDAndStatus(context.Background(), "coach-1", entities.InsightStatusDraft)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
}
