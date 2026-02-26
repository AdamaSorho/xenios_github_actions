package repository

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryInsightCardRepository_Create_AssignsID(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	card, err := repo.Create(context.Background(), &entities.InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Test",
		Body:     "Body",
		Status:   entities.InsightStatusDraft,
		Priority: entities.InsightPriorityMedium,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card.ID == "" {
		t.Error("expected ID to be assigned")
	}
	if card.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}
}

func TestInMemoryInsightCardRepository_Create_PreservesExistingID(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	card, err := repo.Create(context.Background(), &entities.InsightCard{
		ID:       "custom-id",
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Test",
		Status:   entities.InsightStatusDraft,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card.ID != "custom-id" {
		t.Errorf("expected ID 'custom-id', got %q", card.ID)
	}
}

func TestInMemoryInsightCardRepository_Create_CopiesEvidence(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	evidence := []entities.InsightEvidence{
		{MeasurementID: "m-1", Description: "test"},
	}
	card, err := repo.Create(context.Background(), &entities.InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Test",
		Status:   entities.InsightStatusDraft,
		Evidence: evidence,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(card.Evidence) != 1 {
		t.Errorf("expected 1 evidence item, got %d", len(card.Evidence))
	}
}

func TestInMemoryInsightCardRepository_FindByID_ReturnsCard(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	created, _ := repo.Create(context.Background(), &entities.InsightCard{
		ID: "ins-1", CoachID: "coach-1", ClientID: "client-1", Title: "Test", Status: entities.InsightStatusDraft,
	})

	found, err := repo.FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil {
		t.Fatal("expected to find card")
	}
	if found.Title != "Test" {
		t.Errorf("expected title 'Test', got %q", found.Title)
	}
}

func TestInMemoryInsightCardRepository_FindByID_NotFound_ReturnsNil(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	found, err := repo.FindByID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Error("expected nil for missing card")
	}
}

func TestInMemoryInsightCardRepository_Update_ModifiesCard(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	repo.Create(context.Background(), &entities.InsightCard{
		ID: "ins-1", CoachID: "coach-1", ClientID: "client-1", Title: "Old", Status: entities.InsightStatusDraft,
	})

	updated, err := repo.Update(context.Background(), &entities.InsightCard{
		ID: "ins-1", CoachID: "coach-1", ClientID: "client-1", Title: "New", Status: entities.InsightStatusApproved,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated == nil {
		t.Fatal("expected non-nil result")
	}
	if updated.Title != "New" {
		t.Errorf("expected title 'New', got %q", updated.Title)
	}
}

func TestInMemoryInsightCardRepository_Update_NotFound_ReturnsNil(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	result, err := repo.Update(context.Background(), &entities.InsightCard{ID: "nonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil for missing card")
	}
}

func TestInMemoryInsightCardRepository_Update_CopiesEvidence(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	repo.Create(context.Background(), &entities.InsightCard{
		ID: "ins-1", CoachID: "c1", ClientID: "c2", Status: entities.InsightStatusDraft,
	})
	evidence := []entities.InsightEvidence{{MeasurementID: "m-1", Description: "test"}}
	updated, err := repo.Update(context.Background(), &entities.InsightCard{
		ID: "ins-1", CoachID: "c1", ClientID: "c2", Status: entities.InsightStatusDraft, Evidence: evidence,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(updated.Evidence) != 1 {
		t.Errorf("expected 1 evidence item, got %d", len(updated.Evidence))
	}
}

func TestInMemoryInsightCardRepository_ListByCoachID_FiltersByCoach(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	repo.Create(context.Background(), &entities.InsightCard{
		ID: "ins-1", CoachID: "coach-1", ClientID: "c1", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityMedium,
	})
	repo.Create(context.Background(), &entities.InsightCard{
		ID: "ins-2", CoachID: "coach-2", ClientID: "c2", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityMedium,
	})

	cards, total, err := repo.ListByCoachID(context.Background(), entities.InsightQueryFilter{CoachID: "coach-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(cards) != 1 {
		t.Errorf("expected 1 card, got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_ListByCoachID_FiltersByStatus(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	repo.Create(context.Background(), &entities.InsightCard{
		ID: "ins-1", CoachID: "coach-1", ClientID: "c1", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityMedium,
	})
	repo.Create(context.Background(), &entities.InsightCard{
		ID: "ins-2", CoachID: "coach-1", ClientID: "c1", Status: entities.InsightStatusApproved, Priority: entities.InsightPriorityMedium,
	})

	cards, total, err := repo.ListByCoachID(context.Background(), entities.InsightQueryFilter{
		CoachID: "coach-1",
		Status:  "draft",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(cards) != 1 {
		t.Errorf("expected 1 card, got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_ListByCoachID_SortsByPriority(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	repo.Create(context.Background(), &entities.InsightCard{
		ID: "low", CoachID: "coach-1", ClientID: "c1", Priority: entities.InsightPriorityLow, Status: entities.InsightStatusDraft,
	})
	repo.Create(context.Background(), &entities.InsightCard{
		ID: "urgent", CoachID: "coach-1", ClientID: "c1", Priority: entities.InsightPriorityUrgent, Status: entities.InsightStatusDraft,
	})

	cards, _, err := repo.ListByCoachID(context.Background(), entities.InsightQueryFilter{CoachID: "coach-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
	if cards[0].Priority != entities.InsightPriorityUrgent {
		t.Errorf("expected first card to be urgent, got %s", cards[0].Priority)
	}
}

func TestInMemoryInsightCardRepository_ListByClientID_FiltersByClient(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	repo.Create(context.Background(), &entities.InsightCard{
		ID: "ins-1", CoachID: "coach-1", ClientID: "client-1", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityMedium,
	})
	repo.Create(context.Background(), &entities.InsightCard{
		ID: "ins-2", CoachID: "coach-1", ClientID: "client-2", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityMedium,
	})

	cards, total, err := repo.ListByClientID(context.Background(), entities.InsightQueryFilter{ClientID: "client-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(cards) != 1 {
		t.Errorf("expected 1 card, got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_ListByClientID_FiltersByStatus(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	repo.Create(context.Background(), &entities.InsightCard{
		ID: "ins-1", CoachID: "coach-1", ClientID: "client-1", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityMedium,
	})
	repo.Create(context.Background(), &entities.InsightCard{
		ID: "ins-2", CoachID: "coach-1", ClientID: "client-1", Status: entities.InsightStatusApproved, Priority: entities.InsightPriorityMedium,
	})

	cards, total, err := repo.ListByClientID(context.Background(), entities.InsightQueryFilter{
		ClientID: "client-1",
		Status:   "approved",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(cards) != 1 {
		t.Errorf("expected 1 card, got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_Pagination_DefaultLimit(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	for i := 0; i < 25; i++ {
		repo.Create(context.Background(), &entities.InsightCard{
			CoachID: "coach-1", ClientID: "c1", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityMedium,
		})
	}

	cards, total, err := repo.ListByCoachID(context.Background(), entities.InsightQueryFilter{
		CoachID: "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 25 {
		t.Errorf("expected total 25, got %d", total)
	}
	if len(cards) != 20 {
		t.Errorf("expected 20 cards (default limit), got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_Pagination_CustomPageAndLimit(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	for i := 0; i < 10; i++ {
		repo.Create(context.Background(), &entities.InsightCard{
			CoachID: "coach-1", ClientID: "c1", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityMedium,
		})
	}

	cards, total, err := repo.ListByCoachID(context.Background(), entities.InsightQueryFilter{
		CoachID: "coach-1",
		Page:    2,
		Limit:   3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 10 {
		t.Errorf("expected total 10, got %d", total)
	}
	if len(cards) != 3 {
		t.Errorf("expected 3 cards on page 2, got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_Pagination_BeyondTotal(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	repo.Create(context.Background(), &entities.InsightCard{
		ID: "ins-1", CoachID: "coach-1", ClientID: "c1", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityMedium,
	})

	cards, total, err := repo.ListByCoachID(context.Background(), entities.InsightQueryFilter{
		CoachID: "coach-1",
		Page:    100,
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(cards) != 0 {
		t.Errorf("expected 0 cards beyond total, got %d", len(cards))
	}
}
