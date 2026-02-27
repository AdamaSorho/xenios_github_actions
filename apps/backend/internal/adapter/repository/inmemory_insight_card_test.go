package repository

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
	domainrepo "github.com/xenios/backend/internal/domain/repository"
)

// Compile-time interface compliance check.
var _ domainrepo.InsightCardRepository = &InMemoryInsightCardRepository{}

func TestInMemoryInsightCardRepository_Create_ReturnsCardWithID(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	card := &entities.InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Test Insight",
		Body:     "Test body",
		Category: entities.InsightCategoryNutrition,
		Priority: entities.InsightPriorityHigh,
		Status:   entities.InsightStatusDraft,
	}

	saved, err := repo.Create(context.Background(), card)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if saved.ID == "" {
		t.Error("expected non-empty ID")
	}
	if saved.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if saved.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}

func TestInMemoryInsightCardRepository_Create_PreservesFields(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	card := &entities.InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Elevated LDL",
		Body:     "LDL is high",
		Category: entities.InsightCategorySafety,
		Priority: entities.InsightPriorityUrgent,
		Status:   entities.InsightStatusDraft,
		Evidence: []entities.EvidenceRef{
			{MeasurementID: "meas-1", ArtifactID: "art-1", Description: "LDL: 190"},
		},
	}

	saved, err := repo.Create(context.Background(), card)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if saved.CoachID != "coach-1" {
		t.Errorf("expected CoachID coach-1, got %s", saved.CoachID)
	}
	if saved.ClientID != "client-1" {
		t.Errorf("expected ClientID client-1, got %s", saved.ClientID)
	}
	if saved.Title != "Elevated LDL" {
		t.Errorf("expected Title 'Elevated LDL', got %s", saved.Title)
	}
	if saved.Category != entities.InsightCategorySafety {
		t.Errorf("expected Category safety, got %s", saved.Category)
	}
	if saved.Priority != entities.InsightPriorityUrgent {
		t.Errorf("expected Priority urgent, got %s", saved.Priority)
	}
	if len(saved.Evidence) != 1 {
		t.Fatalf("expected 1 evidence ref, got %d", len(saved.Evidence))
	}
	if saved.Evidence[0].MeasurementID != "meas-1" {
		t.Errorf("expected evidence MeasurementID meas-1, got %s", saved.Evidence[0].MeasurementID)
	}
}

func TestInMemoryInsightCardRepository_Create_CopiesEvidence(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	evidence := []entities.EvidenceRef{
		{MeasurementID: "meas-1", ArtifactID: "art-1", Description: "original"},
	}
	card := &entities.InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Test",
		Status:   entities.InsightStatusDraft,
		Evidence: evidence,
	}

	saved, err := repo.Create(context.Background(), card)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Mutate the original slice — stored copy should be unaffected.
	evidence[0].Description = "mutated"
	if saved.Evidence[0].Description != "original" {
		t.Error("expected evidence to be a deep copy, not a reference")
	}
}

func TestInMemoryInsightCardRepository_Create_IncrementsCount(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	for i := 0; i < 3; i++ {
		_, _ = repo.Create(context.Background(), &entities.InsightCard{
			CoachID: "coach-1", ClientID: "client-1", Status: entities.InsightStatusDraft,
		})
	}
	if repo.CardCount() != 3 {
		t.Errorf("expected 3 cards, got %d", repo.CardCount())
	}
}

func TestInMemoryInsightCardRepository_FindByID_Exists_ReturnsCard(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	saved, _ := repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-1", Title: "Find Me", Status: entities.InsightStatusDraft,
	})

	found, err := repo.FindByID(context.Background(), saved.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil {
		t.Fatal("expected card, got nil")
	}
	if found.Title != "Find Me" {
		t.Errorf("expected Title 'Find Me', got %s", found.Title)
	}
}

func TestInMemoryInsightCardRepository_FindByID_NotExists_ReturnsNil(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()

	found, err := repo.FindByID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Errorf("expected nil, got %+v", found)
	}
}

func TestInMemoryInsightCardRepository_FindByClientID_FiltersCorrectly(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	_, _ = repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-1", Title: "Card A", Status: entities.InsightStatusDraft,
	})
	_, _ = repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-2", Title: "Card B", Status: entities.InsightStatusDraft,
	})
	_, _ = repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-1", Title: "Card C", Status: entities.InsightStatusDraft,
	})

	cards, err := repo.FindByClientID(context.Background(), "client-1", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cards) != 2 {
		t.Errorf("expected 2 cards for client-1, got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_FindByClientID_Pagination(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	for i := 0; i < 5; i++ {
		_, _ = repo.Create(context.Background(), &entities.InsightCard{
			CoachID: "coach-1", ClientID: "client-1", Status: entities.InsightStatusDraft,
		})
	}

	cards, err := repo.FindByClientID(context.Background(), "client-1", 2, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cards) != 2 {
		t.Errorf("expected 2 cards with limit=2 offset=1, got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_FindByClientID_OffsetBeyondLength(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	_, _ = repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-1", Status: entities.InsightStatusDraft,
	})

	cards, err := repo.FindByClientID(context.Background(), "client-1", 10, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cards) != 0 {
		t.Errorf("expected 0 cards with offset beyond length, got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_FindByStatus_FiltersCoachAndStatus(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	_, _ = repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-1", Status: entities.InsightStatusDraft,
	})
	_, _ = repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-1", Status: entities.InsightStatusApproved,
	})
	_, _ = repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-2", ClientID: "client-1", Status: entities.InsightStatusDraft,
	})

	cards, err := repo.FindByStatus(context.Background(), "coach-1", entities.InsightStatusDraft, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cards) != 1 {
		t.Errorf("expected 1 draft card for coach-1, got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_FindByStatus_Pagination(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	for i := 0; i < 5; i++ {
		_, _ = repo.Create(context.Background(), &entities.InsightCard{
			CoachID: "coach-1", ClientID: "client-1", Status: entities.InsightStatusDraft,
		})
	}

	cards, err := repo.FindByStatus(context.Background(), "coach-1", entities.InsightStatusDraft, 2, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cards) != 2 {
		t.Errorf("expected 2 cards with limit=2, got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_ExistsByEvidence_Exists_ReturnsTrue(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	_, _ = repo.Create(context.Background(), &entities.InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Status:   entities.InsightStatusDraft,
		Evidence: []entities.EvidenceRef{
			{MeasurementID: "meas-1", ArtifactID: "art-1"},
		},
	})

	exists, err := repo.ExistsByEvidence(context.Background(), "client-1", "meas-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected ExistsByEvidence to return true")
	}
}

func TestInMemoryInsightCardRepository_ExistsByEvidence_DifferentClient_ReturnsFalse(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	_, _ = repo.Create(context.Background(), &entities.InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Status:   entities.InsightStatusDraft,
		Evidence: []entities.EvidenceRef{
			{MeasurementID: "meas-1", ArtifactID: "art-1"},
		},
	})

	exists, err := repo.ExistsByEvidence(context.Background(), "client-2", "meas-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected ExistsByEvidence to return false for different client")
	}
}

func TestInMemoryInsightCardRepository_ExistsByEvidence_DifferentMeasurement_ReturnsFalse(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	_, _ = repo.Create(context.Background(), &entities.InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Status:   entities.InsightStatusDraft,
		Evidence: []entities.EvidenceRef{
			{MeasurementID: "meas-1", ArtifactID: "art-1"},
		},
	})

	exists, err := repo.ExistsByEvidence(context.Background(), "client-1", "meas-999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected ExistsByEvidence to return false for different measurement")
	}
}

func TestInMemoryInsightCardRepository_ExistsByEvidence_Empty_ReturnsFalse(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()

	exists, err := repo.ExistsByEvidence(context.Background(), "client-1", "meas-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected ExistsByEvidence to return false on empty repo")
	}
}

func TestInMemoryInsightCardRepository_UpdateStatus_ExistingCard_UpdatesStatus(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	saved, _ := repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-1", Status: entities.InsightStatusDraft,
	})

	updated, err := repo.UpdateStatus(context.Background(), saved.ID, entities.InsightStatusApproved)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != entities.InsightStatusApproved {
		t.Errorf("expected status approved, got %s", updated.Status)
	}
	if updated.UpdatedAt.Equal(saved.CreatedAt) {
		t.Error("expected UpdatedAt to change")
	}
}

func TestInMemoryInsightCardRepository_UpdateStatus_NotFound_ReturnsError(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()

	_, err := repo.UpdateStatus(context.Background(), "nonexistent", entities.InsightStatusApproved)
	if err == nil {
		t.Fatal("expected error for nonexistent card")
	}
}

func TestInMemoryInsightCardRepository_GetCards_ReturnsSnapshot(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	_, _ = repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-1", Status: entities.InsightStatusDraft,
	})
	_, _ = repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-2", Status: entities.InsightStatusDraft,
	})

	cards := repo.GetCards()
	if len(cards) != 2 {
		t.Errorf("expected 2 cards, got %d", len(cards))
	}
}

func TestInMemoryInsightCardRepository_FindByClientID_DefaultLimit(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	for i := 0; i < 55; i++ {
		_, _ = repo.Create(context.Background(), &entities.InsightCard{
			CoachID: "coach-1", ClientID: "client-1", Status: entities.InsightStatusDraft,
		})
	}

	// limit=0 should default to 50
	cards, err := repo.FindByClientID(context.Background(), "client-1", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cards) != 50 {
		t.Errorf("expected default limit of 50, got %d", len(cards))
	}
}
