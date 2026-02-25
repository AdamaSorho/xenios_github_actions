package repository

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryInsightCardRepository_Create_AssignsID(t *testing.T) {
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

	created, err := repo.Create(context.Background(), card)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == "" {
		t.Error("expected ID to be assigned")
	}
	if created.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestInMemoryInsightCardRepository_Create_PreservesFields(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	card := &entities.InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Elevated LDL",
		Body:     "LDL is high",
		Category: entities.InsightCategoryNutrition,
		Priority: entities.InsightPriorityHigh,
		Status:   entities.InsightStatusDraft,
		Evidence: []entities.EvidenceRef{
			{MeasurementID: "m-1", Description: "LDL: 142"},
		},
	}

	created, err := repo.Create(context.Background(), card)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.Title != "Elevated LDL" {
		t.Errorf("expected title 'Elevated LDL', got %q", created.Title)
	}
	if len(created.Evidence) != 1 {
		t.Fatalf("expected 1 evidence ref, got %d", len(created.Evidence))
	}
	if created.Evidence[0].MeasurementID != "m-1" {
		t.Errorf("expected measurement_id 'm-1', got %q", created.Evidence[0].MeasurementID)
	}
}

func TestInMemoryInsightCardRepository_FindByClientID_ReturnsMatching(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	_, _ = repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-1", Title: "Insight 1",
		Category: entities.InsightCategoryNutrition, Priority: entities.InsightPriorityHigh,
		Status: entities.InsightStatusDraft,
	})
	_, _ = repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-2", Title: "Insight 2",
		Category: entities.InsightCategoryNutrition, Priority: entities.InsightPriorityHigh,
		Status: entities.InsightStatusDraft,
	})
	_, _ = repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-1", Title: "Insight 3",
		Category: entities.InsightCategoryRecovery, Priority: entities.InsightPriorityMedium,
		Status: entities.InsightStatusDraft,
	})

	results, err := repo.FindByClientID(context.Background(), "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestInMemoryInsightCardRepository_FindByStatus_ReturnsMatching(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	_, _ = repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-1", Title: "Draft",
		Category: entities.InsightCategoryNutrition, Priority: entities.InsightPriorityHigh,
		Status: entities.InsightStatusDraft,
	})

	created, _ := repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-1", Title: "To Approve",
		Category: entities.InsightCategoryNutrition, Priority: entities.InsightPriorityHigh,
		Status: entities.InsightStatusDraft,
	})
	_, _ = repo.UpdateStatus(context.Background(), created.ID, entities.InsightStatusApproved)

	drafts, err := repo.FindByStatus(context.Background(), entities.InsightStatusDraft)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(drafts) != 1 {
		t.Errorf("expected 1 draft, got %d", len(drafts))
	}
}

func TestInMemoryInsightCardRepository_UpdateStatus_UpdatesCard(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	created, _ := repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-1", Title: "Test",
		Category: entities.InsightCategoryNutrition, Priority: entities.InsightPriorityHigh,
		Status: entities.InsightStatusDraft,
	})

	updated, err := repo.UpdateStatus(context.Background(), created.ID, entities.InsightStatusApproved)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != entities.InsightStatusApproved {
		t.Errorf("expected status 'approved', got %q", updated.Status)
	}
}

func TestInMemoryInsightCardRepository_UpdateStatus_NotFound_ReturnsError(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	_, err := repo.UpdateStatus(context.Background(), "nonexistent", entities.InsightStatusApproved)
	if err == nil {
		t.Error("expected error for non-existent card")
	}
}

func TestInMemoryInsightCardRepository_ExistsByEvidence_ReturnsTrueWhenExists(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	_, _ = repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-1", Title: "Test",
		Category: entities.InsightCategoryNutrition, Priority: entities.InsightPriorityHigh,
		Status: entities.InsightStatusDraft,
		Evidence: []entities.EvidenceRef{
			{MeasurementID: "m-1", Description: "test"},
		},
	})

	exists, err := repo.ExistsByEvidence(context.Background(), "client-1", "m-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected exists to be true")
	}
}

func TestInMemoryInsightCardRepository_ExistsByEvidence_ReturnsFalseWhenNotExists(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	exists, err := repo.ExistsByEvidence(context.Background(), "client-1", "m-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected exists to be false")
	}
}

func TestInMemoryInsightCardRepository_ExistsByEvidence_DifferentClient_ReturnsFalse(t *testing.T) {
	repo := NewInMemoryInsightCardRepository()
	_, _ = repo.Create(context.Background(), &entities.InsightCard{
		CoachID: "coach-1", ClientID: "client-1", Title: "Test",
		Category: entities.InsightCategoryNutrition, Priority: entities.InsightPriorityHigh,
		Status: entities.InsightStatusDraft,
		Evidence: []entities.EvidenceRef{
			{MeasurementID: "m-1", Description: "test"},
		},
	})

	exists, err := repo.ExistsByEvidence(context.Background(), "client-2", "m-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected exists to be false for different client")
	}
}
