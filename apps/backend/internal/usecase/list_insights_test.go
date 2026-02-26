package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func TestListInsights_ValidCoachAndStatus_ReturnsCards(t *testing.T) {
	insightRepo := repository.NewInMemoryInsightCardRepository()

	insightRepo.Create(context.Background(), &entities.InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Test Insight",
		Body:     "Test body",
		Category: entities.InsightCategoryNutrition,
		Priority: entities.InsightPriorityHigh,
		Status:   entities.InsightStatusDraft,
	})

	uc := NewListInsightsUseCase(insightRepo)
	cards, err := uc.Execute(context.Background(), "coach-1", entities.InsightStatusDraft, 50, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cards) != 1 {
		t.Errorf("expected 1 card, got %d", len(cards))
	}
}

func TestListInsights_NoCards_ReturnsEmptySlice(t *testing.T) {
	insightRepo := repository.NewInMemoryInsightCardRepository()
	uc := NewListInsightsUseCase(insightRepo)

	cards, err := uc.Execute(context.Background(), "coach-1", entities.InsightStatusDraft, 50, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cards) != 0 {
		t.Errorf("expected 0 cards, got %d", len(cards))
	}
}

func TestListInsights_MissingCoachID_ReturnsValidationError(t *testing.T) {
	insightRepo := repository.NewInMemoryInsightCardRepository()
	uc := NewListInsightsUseCase(insightRepo)

	_, err := uc.Execute(context.Background(), "", entities.InsightStatusDraft, 50, 0)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestListInsights_InvalidStatus_ReturnsValidationError(t *testing.T) {
	insightRepo := repository.NewInMemoryInsightCardRepository()
	uc := NewListInsightsUseCase(insightRepo)

	_, err := uc.Execute(context.Background(), "coach-1", "invalid", 50, 0)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestListInsights_FiltersByStatus_ReturnsOnlyMatching(t *testing.T) {
	insightRepo := repository.NewInMemoryInsightCardRepository()

	insightRepo.Create(context.Background(), &entities.InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Draft",
		Body:     "body",
		Category: entities.InsightCategoryGeneral,
		Priority: entities.InsightPriorityMedium,
		Status:   entities.InsightStatusDraft,
	})

	// Create an approved card
	card, _ := insightRepo.Create(context.Background(), &entities.InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Approved",
		Body:     "body",
		Category: entities.InsightCategoryGeneral,
		Priority: entities.InsightPriorityMedium,
		Status:   entities.InsightStatusDraft,
	})
	insightRepo.UpdateStatus(context.Background(), card.ID, entities.InsightStatusApproved)

	uc := NewListInsightsUseCase(insightRepo)
	drafts, err := uc.Execute(context.Background(), "coach-1", entities.InsightStatusDraft, 50, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(drafts) != 1 {
		t.Errorf("expected 1 draft, got %d", len(drafts))
	}
}

// Ensure the compile-time checks for time usage don't break
var _ = time.Now
