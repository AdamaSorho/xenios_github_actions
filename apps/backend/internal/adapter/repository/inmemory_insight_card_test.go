package repository

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryInsightCardRepository_Create_AssignsIDAndTimestamps(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	card := &entities.InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Test Insight",
		Body:     "Test body",
		Category: entities.InsightCategoryNutrition,
		Priority: entities.InsightPriorityHigh,
		Status:   entities.InsightStatusDraft,
		Evidence: []entities.EvidenceRef{{MeasurementID: "m-1", Description: "test"}},
	}

	created, err := repo.Create(context.Background(), card)
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

func TestInMemoryInsightCardRepository_FindByClientID_FiltersCorrectly(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()

	repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-1", Title: "A", Body: "A",
		Category: entities.InsightCategoryNutrition, Priority: entities.InsightPriorityHigh,
		Status: entities.InsightStatusDraft, Evidence: []entities.EvidenceRef{{MeasurementID: "m-1"}},
	})
	repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-2", Title: "B", Body: "B",
		Category: entities.InsightCategoryRecovery, Priority: entities.InsightPriorityMedium,
		Status: entities.InsightStatusDraft, Evidence: []entities.EvidenceRef{{MeasurementID: "m-2"}},
	})
	repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-1", Title: "C", Body: "C",
		Category: entities.InsightCategorySafety, Priority: entities.InsightPriorityUrgent,
		Status: entities.InsightStatusDraft, Evidence: []entities.EvidenceRef{{MeasurementID: "m-3"}},
	})

	cards, err := repo.FindByClientID(context.Background(), "client-1", 50, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards for client-1, got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_FindByStatus_FiltersCorrectly(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()

	repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-1", Title: "A", Body: "A",
		Category: entities.InsightCategoryNutrition, Priority: entities.InsightPriorityHigh,
		Status: entities.InsightStatusDraft, Evidence: []entities.EvidenceRef{{MeasurementID: "m-1"}},
	})

	cards, err := repo.FindByStatus(context.Background(), entities.InsightStatusDraft, 50, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 draft card, got %d", len(cards))
	}

	approved, err := repo.FindByStatus(context.Background(), entities.InsightStatusApproved, 50, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(approved) != 0 {
		t.Fatalf("expected 0 approved cards, got %d", len(approved))
	}
}

func TestInMemoryInsightCardRepository_UpdateStatus_ChangesStatus(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()

	created, _ := repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-1", Title: "A", Body: "A",
		Category: entities.InsightCategoryNutrition, Priority: entities.InsightPriorityHigh,
		Status: entities.InsightStatusDraft, Evidence: []entities.EvidenceRef{{MeasurementID: "m-1"}},
	})

	updated, err := repo.UpdateStatus(context.Background(), created.ID, entities.InsightStatusApproved)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != entities.InsightStatusApproved {
		t.Errorf("expected status %q, got %q", entities.InsightStatusApproved, updated.Status)
	}
}

func TestInMemoryInsightCardRepository_UpdateStatus_NotFound_ReturnsError(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()

	_, err := repo.UpdateStatus(context.Background(), "nonexistent", entities.InsightStatusApproved)
	if err == nil {
		t.Fatal("expected error for nonexistent card")
	}
}

func TestInMemoryInsightCardRepository_ExistsByMeasurementID_Found(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()

	repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-1", Title: "A", Body: "A",
		Category: entities.InsightCategoryNutrition, Priority: entities.InsightPriorityHigh,
		Status: entities.InsightStatusDraft, Evidence: []entities.EvidenceRef{{MeasurementID: "m-1"}},
	})

	exists, err := repo.ExistsByMeasurementID(context.Background(), "m-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected ExistsByMeasurementID to return true")
	}
}

func TestInMemoryInsightCardRepository_ExistsByMeasurementID_NotFound(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()

	exists, err := repo.ExistsByMeasurementID(context.Background(), "m-999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected ExistsByMeasurementID to return false")
	}
}

func TestInMemoryInsightCardRepository_FindByClientID_Pagination(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()

	for i := 0; i < 5; i++ {
		repo.Create(context.Background(), &entities.InsightCard{
			CoachID: "coach-1", ClientID: "client-1", Title: "A", Body: "A",
			Category: entities.InsightCategoryNutrition, Priority: entities.InsightPriorityHigh,
			Status: entities.InsightStatusDraft, Evidence: []entities.EvidenceRef{{MeasurementID: "m-1"}},
		})
	}

	cards, err := repo.FindByClientID(context.Background(), "client-1", 2, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards with limit=2, got %d", len(cards))
	}

	cards, err = repo.FindByClientID(context.Background(), "client-1", 2, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards with limit=2 offset=3, got %d", len(cards))
	}
}
