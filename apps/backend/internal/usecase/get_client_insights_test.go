package usecase

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func newGetClientInsightsUseCase() (*GetClientInsightsUseCase, *repository.InMemoryInsightCardRepository) {
	insightRepo := repository.NewInMemoryInsightCardRepository()
	uc := NewGetClientInsightsUseCase(insightRepo)
	return uc, insightRepo
}

func TestGetClientInsights_ReturnsInsightsForClient(t *testing.T) {
	uc, insightRepo := newGetClientInsightsUseCase()
	ctx := context.Background()

	seedDraftInsight(insightRepo, "coach-1", "client-1")
	seedApprovedInsight(insightRepo, "coach-1", "client-1")
	seedDraftInsight(insightRepo, "coach-1", "client-2")

	result, err := uc.Execute(ctx, ClientInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
		Page:     1,
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Insights) != 2 {
		t.Errorf("expected 2 insights, got %d", len(result.Insights))
	}
}

func TestGetClientInsights_FiltersByStatus(t *testing.T) {
	uc, insightRepo := newGetClientInsightsUseCase()
	ctx := context.Background()

	seedDraftInsight(insightRepo, "coach-1", "client-1")
	seedApprovedInsight(insightRepo, "coach-1", "client-1")

	status := entities.InsightStatusApproved
	result, err := uc.Execute(ctx, ClientInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
		Status:   &status,
		Page:     1,
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Insights) != 1 {
		t.Errorf("expected 1 insight, got %d", len(result.Insights))
	}
	if result.Insights[0].Status != entities.InsightStatusApproved {
		t.Errorf("expected status 'approved', got %q", result.Insights[0].Status)
	}
}

func TestGetClientInsights_EmptyClientID_ReturnsValidationError(t *testing.T) {
	uc, _ := newGetClientInsightsUseCase()

	_, err := uc.Execute(context.Background(), ClientInsightsInput{
		ClientID: "",
		CoachID:  "coach-1",
		Page:     1,
		Limit:    20,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetClientInsights_EmptyCoachID_ReturnsValidationError(t *testing.T) {
	uc, _ := newGetClientInsightsUseCase()

	_, err := uc.Execute(context.Background(), ClientInsightsInput{
		ClientID: "client-1",
		CoachID:  "",
		Page:     1,
		Limit:    20,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetClientInsights_NoMatches_ReturnsEmptyList(t *testing.T) {
	uc, _ := newGetClientInsightsUseCase()

	result, err := uc.Execute(context.Background(), ClientInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
		Page:     1,
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Insights == nil {
		t.Error("expected non-nil insights slice")
	}
	if len(result.Insights) != 0 {
		t.Errorf("expected 0 insights, got %d", len(result.Insights))
	}
}

func TestGetClientInsights_Pagination(t *testing.T) {
	uc, insightRepo := newGetClientInsightsUseCase()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		seedDraftInsight(insightRepo, "coach-1", "client-1")
	}

	result, err := uc.Execute(ctx, ClientInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
		Page:     1,
		Limit:    2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Insights) != 2 {
		t.Errorf("expected 2 insights, got %d", len(result.Insights))
	}
	if result.Pagination.Total != 5 {
		t.Errorf("expected total 5, got %d", result.Pagination.Total)
	}
}

func TestGetClientInsights_DefaultPagination(t *testing.T) {
	uc, _ := newGetClientInsightsUseCase()

	result, err := uc.Execute(context.Background(), ClientInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
		Page:     0,
		Limit:    0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Pagination.Page != 1 {
		t.Errorf("expected default page 1, got %d", result.Pagination.Page)
	}
	if result.Pagination.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", result.Pagination.Limit)
	}
}
