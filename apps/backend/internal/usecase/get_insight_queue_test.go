package usecase

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func newGetInsightQueueUseCase() (*GetInsightQueueUseCase, *repository.InMemoryInsightCardRepository) {
	insightRepo := repository.NewInMemoryInsightCardRepository()
	uc := NewGetInsightQueueUseCase(insightRepo)
	return uc, insightRepo
}

func TestGetInsightQueue_ReturnsOnlyDrafts(t *testing.T) {
	uc, insightRepo := newGetInsightQueueUseCase()
	ctx := context.Background()

	seedDraftInsight(insightRepo, "coach-1", "client-1")
	seedDraftInsight(insightRepo, "coach-1", "client-2")
	seedApprovedInsight(insightRepo, "coach-1", "client-1")

	result, err := uc.Execute(ctx, InsightQueueInput{
		CoachID: "coach-1",
		Page:    1,
		Limit:   20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Insights) != 2 {
		t.Errorf("expected 2 draft insights, got %d", len(result.Insights))
	}
	if result.Pagination.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Pagination.Total)
	}
}

func TestGetInsightQueue_DoesNotReturnOtherCoachInsights(t *testing.T) {
	uc, insightRepo := newGetInsightQueueUseCase()
	ctx := context.Background()

	seedDraftInsight(insightRepo, "coach-1", "client-1")
	seedDraftInsight(insightRepo, "coach-2", "client-2")

	result, err := uc.Execute(ctx, InsightQueueInput{
		CoachID: "coach-1",
		Page:    1,
		Limit:   20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Insights) != 1 {
		t.Errorf("expected 1 insight, got %d", len(result.Insights))
	}
}

func TestGetInsightQueue_EmptyQueue_ReturnsEmptyList(t *testing.T) {
	uc, _ := newGetInsightQueueUseCase()

	result, err := uc.Execute(context.Background(), InsightQueueInput{
		CoachID: "coach-1",
		Page:    1,
		Limit:   20,
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

func TestGetInsightQueue_EmptyCoachID_ReturnsValidationError(t *testing.T) {
	uc, _ := newGetInsightQueueUseCase()

	_, err := uc.Execute(context.Background(), InsightQueueInput{
		CoachID: "",
		Page:    1,
		Limit:   20,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetInsightQueue_Pagination_RespectsLimitAndPage(t *testing.T) {
	uc, insightRepo := newGetInsightQueueUseCase()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		seedDraftInsight(insightRepo, "coach-1", "client-1")
	}

	result, err := uc.Execute(ctx, InsightQueueInput{
		CoachID: "coach-1",
		Page:    1,
		Limit:   2,
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
	if result.Pagination.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Pagination.Page)
	}
}

func TestGetInsightQueue_DefaultPagination(t *testing.T) {
	uc, _ := newGetInsightQueueUseCase()

	result, err := uc.Execute(context.Background(), InsightQueueInput{
		CoachID: "coach-1",
		Page:    0,
		Limit:   0,
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

func TestGetInsightQueue_SortedByPriority(t *testing.T) {
	uc, insightRepo := newGetInsightQueueUseCase()
	ctx := context.Background()

	insightRepo.Create(ctx, &entities.InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Low Priority",
		Body:     "body",
		Category: "nutrition",
		Priority: entities.InsightPriorityLow,
		Status:   entities.InsightStatusDraft,
	})
	insightRepo.Create(ctx, &entities.InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-2",
		Title:    "Urgent Priority",
		Body:     "body",
		Category: "nutrition",
		Priority: entities.InsightPriorityUrgent,
		Status:   entities.InsightStatusDraft,
	})

	result, err := uc.Execute(ctx, InsightQueueInput{CoachID: "coach-1", Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Insights[0].Priority != entities.InsightPriorityUrgent {
		t.Errorf("expected first insight to be urgent, got %q", result.Insights[0].Priority)
	}
}
