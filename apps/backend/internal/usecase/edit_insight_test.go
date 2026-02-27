package usecase

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
)

func newEditInsightUseCase() (*EditInsightUseCase, *repository.InMemoryInsightCardRepository, *repository.InMemoryAuditRepository) {
	insightRepo := repository.NewInMemoryInsightCardRepository()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewEditInsightUseCase(insightRepo, auditRepo)
	return uc, insightRepo, auditRepo
}

func TestEditInsight_ValidInput_UpdatesTitleAndBody(t *testing.T) {
	uc, insightRepo, _ := newEditInsightUseCase()
	insight := seedDraftInsight(insightRepo, "coach-1", "client-1")

	result, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: insight.ID,
		CoachID:   "coach-1",
		Title:     "Updated Title",
		Body:      "Updated Body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %q", result.Title)
	}
	if result.Body != "Updated Body" {
		t.Errorf("expected body 'Updated Body', got %q", result.Body)
	}
}

func TestEditInsight_EmptyInsightID_ReturnsValidationError(t *testing.T) {
	uc, _, _ := newEditInsightUseCase()

	_, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: "",
		CoachID:   "coach-1",
		Title:     "Title",
		Body:      "Body",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestEditInsight_EmptyCoachID_ReturnsValidationError(t *testing.T) {
	uc, _, _ := newEditInsightUseCase()

	_, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: "insight-1",
		CoachID:   "",
		Title:     "Title",
		Body:      "Body",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestEditInsight_EmptyTitle_ReturnsValidationError(t *testing.T) {
	uc, _, _ := newEditInsightUseCase()

	_, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: "insight-1",
		CoachID:   "coach-1",
		Title:     "",
		Body:      "Body",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestEditInsight_EmptyBody_ReturnsValidationError(t *testing.T) {
	uc, _, _ := newEditInsightUseCase()

	_, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: "insight-1",
		CoachID:   "coach-1",
		Title:     "Title",
		Body:      "",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestEditInsight_NotFound_ReturnsValidationError(t *testing.T) {
	uc, _, _ := newEditInsightUseCase()

	_, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: "nonexistent",
		CoachID:   "coach-1",
		Title:     "Title",
		Body:      "Body",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestEditInsight_WrongCoach_ReturnsAuthError(t *testing.T) {
	uc, insightRepo, _ := newEditInsightUseCase()
	insight := seedDraftInsight(insightRepo, "coach-1", "client-1")

	_, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: insight.ID,
		CoachID:   "coach-other",
		Title:     "Title",
		Body:      "Body",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsAuthenticationError(err) {
		t.Errorf("expected AuthenticationError, got %T", err)
	}
}

func TestEditInsight_ApprovedInsight_ReturnsValidationError(t *testing.T) {
	uc, insightRepo, _ := newEditInsightUseCase()
	insight := seedApprovedInsight(insightRepo, "coach-1", "client-1")

	_, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: insight.ID,
		CoachID:   "coach-1",
		Title:     "Title",
		Body:      "Body",
	})
	if err == nil {
		t.Fatal("expected error for editing approved insight")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestEditInsight_LogsAuditEvent(t *testing.T) {
	uc, insightRepo, auditRepo := newEditInsightUseCase()
	insight := seedDraftInsight(insightRepo, "coach-1", "client-1")

	_, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: insight.ID,
		CoachID:   "coach-1",
		Title:     "Updated",
		Body:      "Updated Body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := auditRepo.GetEvents()
	if len(events) == 0 {
		t.Fatal("expected audit event to be logged")
	}
	if events[0].Action != "insight.edited" {
		t.Errorf("expected action 'insight.edited', got %q", events[0].Action)
	}
}
