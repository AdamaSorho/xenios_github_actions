package repository

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	domainrepo "github.com/xenios/backend/internal/domain/repository"
)

func TestInMemoryInsightCardRepository_Create_AssignsID(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	card := &entities.InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Test",
		Body:     "Body",
		Category: entities.InsightCategoryGeneral,
		Status:   entities.InsightStatusDraft,
		Priority: entities.InsightPriorityMedium,
	}

	err := repo.Create(context.Background(), card)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card.ID == "" {
		t.Error("expected ID to be assigned")
	}
}

func TestInMemoryInsightCardRepository_FindByID_ExistingCard(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	repo.Seed(&entities.InsightCard{
		ID:       "i1",
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Test",
		Body:     "Body",
		Category: entities.InsightCategoryGeneral,
		Status:   entities.InsightStatusDraft,
		Priority: entities.InsightPriorityMedium,
	})

	card, err := repo.FindByID(context.Background(), "i1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card == nil {
		t.Fatal("expected card, got nil")
	}
	if card.ID != "i1" {
		t.Errorf("expected ID 'i1', got %q", card.ID)
	}
}

func TestInMemoryInsightCardRepository_FindByID_NotFound(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	card, err := repo.FindByID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card != nil {
		t.Fatal("expected nil, got card")
	}
}

func TestInMemoryInsightCardRepository_Update_ExistingCard(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	repo.Seed(&entities.InsightCard{
		ID:       "i1",
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Original",
		Body:     "Body",
		Category: entities.InsightCategoryGeneral,
		Status:   entities.InsightStatusDraft,
		Priority: entities.InsightPriorityMedium,
	})

	card, _ := repo.FindByID(context.Background(), "i1")
	card.Title = "Updated"
	err := repo.Update(context.Background(), card)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := repo.FindByID(context.Background(), "i1")
	if updated.Title != "Updated" {
		t.Errorf("expected title 'Updated', got %q", updated.Title)
	}
}

func TestInMemoryInsightCardRepository_Update_NotFound(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	err := repo.Update(context.Background(), &entities.InsightCard{ID: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent card")
	}
}

func TestInMemoryInsightCardRepository_ListByCoach_FilterByCoachID(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	now := time.Now()
	repo.Seed(
		&entities.InsightCard{ID: "i1", CoachID: "coach-1", ClientID: "c1", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityHigh, CreatedAt: now},
		&entities.InsightCard{ID: "i2", CoachID: "coach-1", ClientID: "c2", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityMedium, CreatedAt: now},
		&entities.InsightCard{ID: "i3", CoachID: "coach-2", ClientID: "c3", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityLow, CreatedAt: now},
	)

	cards, total, err := repo.ListByCoach(context.Background(), domainrepo.InsightCardFilter{CoachID: "coach-1", Limit: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2 total, got %d", total)
	}
	if len(cards) != 2 {
		t.Errorf("expected 2 cards, got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_ListByCoach_FilterByStatus(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	repo.Seed(
		&entities.InsightCard{ID: "i1", CoachID: "coach-1", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityMedium},
		&entities.InsightCard{ID: "i2", CoachID: "coach-1", Status: entities.InsightStatusApproved, Priority: entities.InsightPriorityMedium},
	)

	cards, total, err := repo.ListByCoach(context.Background(), domainrepo.InsightCardFilter{CoachID: "coach-1", Status: entities.InsightStatusDraft, Limit: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 total, got %d", total)
	}
	if len(cards) != 1 {
		t.Errorf("expected 1 card, got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_ListByCoach_PrioritySorting(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	now := time.Now()
	repo.Seed(
		&entities.InsightCard{ID: "low", CoachID: "coach-1", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityLow, CreatedAt: now},
		&entities.InsightCard{ID: "urgent", CoachID: "coach-1", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityUrgent, CreatedAt: now},
		&entities.InsightCard{ID: "medium", CoachID: "coach-1", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityMedium, CreatedAt: now},
	)

	cards, _, err := repo.ListByCoach(context.Background(), domainrepo.InsightCardFilter{CoachID: "coach-1", Limit: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cards[0].ID != "urgent" {
		t.Errorf("expected first card 'urgent', got %q", cards[0].ID)
	}
}

func TestInMemoryInsightCardRepository_ListByCoach_Pagination(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	now := time.Now()
	for i := 0; i < 5; i++ {
		repo.Seed(&entities.InsightCard{
			CoachID:   "coach-1",
			Status:    entities.InsightStatusDraft,
			Priority:  entities.InsightPriorityMedium,
			CreatedAt: now,
		})
	}

	cards, total, err := repo.ListByCoach(context.Background(), domainrepo.InsightCardFilter{CoachID: "coach-1", Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(cards) != 2 {
		t.Errorf("expected 2 cards, got %d", len(cards))
	}

	// Offset past end
	cards, _, _ = repo.ListByCoach(context.Background(), domainrepo.InsightCardFilter{CoachID: "coach-1", Limit: 2, Offset: 10})
	if len(cards) != 0 {
		t.Errorf("expected 0 cards for offset past end, got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_ListByClient_FilterByClientID(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	now := time.Now()
	repo.Seed(
		&entities.InsightCard{ID: "i1", CoachID: "coach-1", ClientID: "client-1", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityMedium, CreatedAt: now},
		&entities.InsightCard{ID: "i2", CoachID: "coach-1", ClientID: "client-1", Status: entities.InsightStatusApproved, Priority: entities.InsightPriorityMedium, CreatedAt: now},
		&entities.InsightCard{ID: "i3", CoachID: "coach-1", ClientID: "client-2", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityMedium, CreatedAt: now},
	)

	cards, total, err := repo.ListByClient(context.Background(), domainrepo.InsightCardFilter{ClientID: "client-1", Limit: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2 total, got %d", total)
	}
	if len(cards) != 2 {
		t.Errorf("expected 2 cards, got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_ListByClient_Pagination(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	now := time.Now()
	for i := 0; i < 5; i++ {
		repo.Seed(&entities.InsightCard{
			CoachID:   "coach-1",
			ClientID:  "client-1",
			Status:    entities.InsightStatusDraft,
			Priority:  entities.InsightPriorityMedium,
			CreatedAt: now,
		})
	}

	cards, total, err := repo.ListByClient(context.Background(), domainrepo.InsightCardFilter{ClientID: "client-1", Limit: 3, Offset: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(cards) != 3 {
		t.Errorf("expected 3 cards, got %d", len(cards))
	}

	// Offset past end
	cards, _, _ = repo.ListByClient(context.Background(), domainrepo.InsightCardFilter{ClientID: "client-1", Limit: 2, Offset: 10})
	if len(cards) != 0 {
		t.Errorf("expected 0 cards, got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_ListByCoach_FilterByClientID(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	now := time.Now()
	repo.Seed(
		&entities.InsightCard{ID: "i1", CoachID: "coach-1", ClientID: "client-1", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityMedium, CreatedAt: now},
		&entities.InsightCard{ID: "i2", CoachID: "coach-1", ClientID: "client-2", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityMedium, CreatedAt: now},
	)

	cards, total, err := repo.ListByCoach(context.Background(), domainrepo.InsightCardFilter{CoachID: "coach-1", ClientID: "client-1", Limit: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 total, got %d", total)
	}
	if len(cards) != 1 {
		t.Errorf("expected 1 card, got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_ListByClient_FilterByCoachID(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	now := time.Now()
	repo.Seed(
		&entities.InsightCard{ID: "i1", CoachID: "coach-1", ClientID: "client-1", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityMedium, CreatedAt: now},
		&entities.InsightCard{ID: "i2", CoachID: "coach-2", ClientID: "client-1", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityMedium, CreatedAt: now},
	)

	cards, total, err := repo.ListByClient(context.Background(), domainrepo.InsightCardFilter{ClientID: "client-1", CoachID: "coach-1", Limit: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 total, got %d", total)
	}
	if len(cards) != 1 {
		t.Errorf("expected 1 card, got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_Seed_SetsID(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	repo.Seed(&entities.InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Status:   entities.InsightStatusDraft,
		Priority: entities.InsightPriorityMedium,
	})

	cards, _, _ := repo.ListByCoach(context.Background(), domainrepo.InsightCardFilter{CoachID: "coach-1", Limit: 20})
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].ID == "" {
		t.Error("expected ID to be assigned by Seed")
	}
}

func TestInMemoryInsightCardRepository_ListByCoach_DefaultLimit(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	now := time.Now()
	for i := 0; i < 25; i++ {
		repo.Seed(&entities.InsightCard{
			CoachID:   "coach-1",
			Status:    entities.InsightStatusDraft,
			Priority:  entities.InsightPriorityMedium,
			CreatedAt: now,
		})
	}

	// Test with limit 0 (should default to 20)
	cards, total, _ := repo.ListByCoach(context.Background(), domainrepo.InsightCardFilter{CoachID: "coach-1", Limit: 0})
	if total != 25 {
		t.Errorf("expected total 25, got %d", total)
	}
	if len(cards) != 20 {
		t.Errorf("expected 20 cards (default limit), got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_ListByClient_DefaultLimit(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	now := time.Now()
	for i := 0; i < 25; i++ {
		repo.Seed(&entities.InsightCard{
			CoachID:   "coach-1",
			ClientID:  "client-1",
			Status:    entities.InsightStatusDraft,
			Priority:  entities.InsightPriorityMedium,
			CreatedAt: now,
		})
	}

	cards, total, _ := repo.ListByClient(context.Background(), domainrepo.InsightCardFilter{ClientID: "client-1", Limit: 0})
	if total != 25 {
		t.Errorf("expected total 25, got %d", total)
	}
	if len(cards) != 20 {
		t.Errorf("expected 20 cards (default limit), got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_ListByClient_FilterByStatus(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	now := time.Now()
	repo.Seed(
		&entities.InsightCard{ID: "i1", CoachID: "coach-1", ClientID: "client-1", Status: entities.InsightStatusDraft, Priority: entities.InsightPriorityMedium, CreatedAt: now},
		&entities.InsightCard{ID: "i2", CoachID: "coach-1", ClientID: "client-1", Status: entities.InsightStatusApproved, Priority: entities.InsightPriorityMedium, CreatedAt: now},
	)

	cards, total, err := repo.ListByClient(context.Background(), domainrepo.InsightCardFilter{ClientID: "client-1", Status: entities.InsightStatusDraft, Limit: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 total, got %d", total)
	}
	if len(cards) != 1 {
		t.Errorf("expected 1 card, got %d", len(cards))
	}
}
