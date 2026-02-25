package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

// --- Test helpers ---

func newInsightTestDeps() (*repository.InMemoryInsightCardRepository, *repository.InMemoryAuditRepository) {
	return repository.NewInMemoryInsightCardRepository(), repository.NewInMemoryAuditRepository()
}

func seedDraftInsight(repo *repository.InMemoryInsightCardRepository, id, coachID, clientID, priority string) *entities.InsightCard {
	card := &entities.InsightCard{
		ID:        id,
		CoachID:   coachID,
		ClientID:  clientID,
		Title:     "Test Insight " + id,
		Body:      "Test body for " + id,
		Category:  entities.InsightCategoryNutrition,
		Status:    entities.InsightStatusDraft,
		Priority:  priority,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Seed(card)
	return card
}

// --- GetInsightQueueUseCase Tests ---

func TestGetInsightQueue_ValidCoach_ReturnsDraftInsights(t *testing.T) {
	insightRepo, _ := newInsightTestDeps()
	uc := NewGetInsightQueueUseCase(insightRepo)

	seedDraftInsight(insightRepo, "i1", "coach-1", "client-1", entities.InsightPriorityHigh)
	seedDraftInsight(insightRepo, "i2", "coach-1", "client-2", entities.InsightPriorityMedium)
	seedDraftInsight(insightRepo, "i3", "coach-2", "client-3", entities.InsightPriorityLow)

	out, err := uc.Execute(context.Background(), GetInsightQueueInput{
		CoachID: "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Insights) != 2 {
		t.Errorf("expected 2 insights, got %d", len(out.Insights))
	}
	if out.Total != 2 {
		t.Errorf("expected total 2, got %d", out.Total)
	}
}

func TestGetInsightQueue_EmptyCoachID_ReturnsValidationError(t *testing.T) {
	insightRepo, _ := newInsightTestDeps()
	uc := NewGetInsightQueueUseCase(insightRepo)

	_, err := uc.Execute(context.Background(), GetInsightQueueInput{CoachID: ""})
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestGetInsightQueue_PriorityOrdering_UrgentFirst(t *testing.T) {
	insightRepo, _ := newInsightTestDeps()
	uc := NewGetInsightQueueUseCase(insightRepo)

	seedDraftInsight(insightRepo, "low", "coach-1", "client-1", entities.InsightPriorityLow)
	seedDraftInsight(insightRepo, "urgent", "coach-1", "client-1", entities.InsightPriorityUrgent)
	seedDraftInsight(insightRepo, "medium", "coach-1", "client-1", entities.InsightPriorityMedium)

	out, err := uc.Execute(context.Background(), GetInsightQueueInput{CoachID: "coach-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Insights) != 3 {
		t.Fatalf("expected 3 insights, got %d", len(out.Insights))
	}
	if out.Insights[0].ID != "urgent" {
		t.Errorf("expected first insight to be urgent, got %s", out.Insights[0].ID)
	}
}

func TestGetInsightQueue_EmptyQueue_ReturnsEmptyList(t *testing.T) {
	insightRepo, _ := newInsightTestDeps()
	uc := NewGetInsightQueueUseCase(insightRepo)

	out, err := uc.Execute(context.Background(), GetInsightQueueInput{CoachID: "coach-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Insights) != 0 {
		t.Errorf("expected 0 insights, got %d", len(out.Insights))
	}
}

func TestGetInsightQueue_StatusFilter_ReturnsFilteredInsights(t *testing.T) {
	insightRepo, _ := newInsightTestDeps()
	uc := NewGetInsightQueueUseCase(insightRepo)

	seedDraftInsight(insightRepo, "i1", "coach-1", "client-1", entities.InsightPriorityHigh)
	// Create an approved insight
	approved := &entities.InsightCard{
		ID:       "i2",
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Approved",
		Body:     "body",
		Category: entities.InsightCategoryGeneral,
		Status:   entities.InsightStatusApproved,
		Priority: entities.InsightPriorityMedium,
	}
	insightRepo.Seed(approved)

	out, err := uc.Execute(context.Background(), GetInsightQueueInput{
		CoachID: "coach-1",
		Status:  entities.InsightStatusApproved,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Insights) != 1 {
		t.Errorf("expected 1 approved insight, got %d", len(out.Insights))
	}
}

// --- GetClientInsightsUseCase Tests ---

func TestGetClientInsights_ValidInput_ReturnsInsights(t *testing.T) {
	insightRepo, _ := newInsightTestDeps()
	uc := NewGetClientInsightsUseCase(insightRepo)

	seedDraftInsight(insightRepo, "i1", "coach-1", "client-1", entities.InsightPriorityHigh)
	seedDraftInsight(insightRepo, "i2", "coach-1", "client-1", entities.InsightPriorityMedium)
	seedDraftInsight(insightRepo, "i3", "coach-1", "client-2", entities.InsightPriorityLow)

	out, err := uc.Execute(context.Background(), GetClientInsightsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Insights) != 2 {
		t.Errorf("expected 2 insights, got %d", len(out.Insights))
	}
}

func TestGetClientInsights_EmptyClientID_ReturnsValidationError(t *testing.T) {
	insightRepo, _ := newInsightTestDeps()
	uc := NewGetClientInsightsUseCase(insightRepo)

	_, err := uc.Execute(context.Background(), GetClientInsightsInput{
		CoachID:  "coach-1",
		ClientID: "",
	})
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestGetClientInsights_EmptyCoachID_ReturnsValidationError(t *testing.T) {
	insightRepo, _ := newInsightTestDeps()
	uc := NewGetClientInsightsUseCase(insightRepo)

	_, err := uc.Execute(context.Background(), GetClientInsightsInput{
		CoachID:  "",
		ClientID: "client-1",
	})
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestGetClientInsights_StatusFilter_ReturnsFiltered(t *testing.T) {
	insightRepo, _ := newInsightTestDeps()
	uc := NewGetClientInsightsUseCase(insightRepo)

	seedDraftInsight(insightRepo, "i1", "coach-1", "client-1", entities.InsightPriorityHigh)
	insightRepo.Seed(&entities.InsightCard{
		ID:       "i2",
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Approved",
		Body:     "body",
		Category: entities.InsightCategoryGeneral,
		Status:   entities.InsightStatusApproved,
		Priority: entities.InsightPriorityMedium,
	})

	out, err := uc.Execute(context.Background(), GetClientInsightsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Status:   entities.InsightStatusDraft,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Insights) != 1 {
		t.Errorf("expected 1 draft insight, got %d", len(out.Insights))
	}
}

// --- ApproveInsightUseCase Tests ---

func TestApproveInsight_ValidDraft_ReturnsApproved(t *testing.T) {
	insightRepo, auditRepo := newInsightTestDeps()
	uc := NewApproveInsightUseCase(insightRepo, auditRepo)

	seedDraftInsight(insightRepo, "i1", "coach-1", "client-1", entities.InsightPriorityHigh)

	card, err := uc.Execute(context.Background(), ApproveInsightInput{
		InsightID: "i1",
		CoachID:   "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card.Status != entities.InsightStatusApproved {
		t.Errorf("expected status approved, got %s", card.Status)
	}
	if card.ApprovedAt == nil {
		t.Error("expected approved_at to be set")
	}

	// Verify audit event
	if auditRepo.EventCount() != 1 {
		t.Errorf("expected 1 audit event, got %d", auditRepo.EventCount())
	}
}

func TestApproveInsight_EmptyInsightID_ReturnsValidationError(t *testing.T) {
	insightRepo, auditRepo := newInsightTestDeps()
	uc := NewApproveInsightUseCase(insightRepo, auditRepo)

	_, err := uc.Execute(context.Background(), ApproveInsightInput{
		InsightID: "",
		CoachID:   "coach-1",
	})
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestApproveInsight_NotFound_ReturnsValidationError(t *testing.T) {
	insightRepo, auditRepo := newInsightTestDeps()
	uc := NewApproveInsightUseCase(insightRepo, auditRepo)

	_, err := uc.Execute(context.Background(), ApproveInsightInput{
		InsightID: "nonexistent",
		CoachID:   "coach-1",
	})
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestApproveInsight_WrongCoach_ReturnsAuthorizationError(t *testing.T) {
	insightRepo, auditRepo := newInsightTestDeps()
	uc := NewApproveInsightUseCase(insightRepo, auditRepo)

	seedDraftInsight(insightRepo, "i1", "coach-1", "client-1", entities.InsightPriorityHigh)

	_, err := uc.Execute(context.Background(), ApproveInsightInput{
		InsightID: "i1",
		CoachID:   "coach-2",
	})
	if !IsAuthorizationError(err) {
		t.Errorf("expected AuthorizationError, got %T: %v", err, err)
	}
}

func TestApproveInsight_AlreadyDismissed_ReturnsStatusTransitionError(t *testing.T) {
	insightRepo, auditRepo := newInsightTestDeps()
	uc := NewApproveInsightUseCase(insightRepo, auditRepo)

	insightRepo.Seed(&entities.InsightCard{
		ID:       "i1",
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Dismissed",
		Body:     "body",
		Category: entities.InsightCategoryGeneral,
		Status:   entities.InsightStatusDismissed,
		Priority: entities.InsightPriorityMedium,
	})

	_, err := uc.Execute(context.Background(), ApproveInsightInput{
		InsightID: "i1",
		CoachID:   "coach-1",
	})
	if err == nil {
		t.Fatal("expected error for invalid transition")
	}
	if _, ok := err.(*entities.StatusTransitionError); !ok {
		t.Errorf("expected StatusTransitionError, got %T: %v", err, err)
	}
}

// --- DismissInsightUseCase Tests ---

func TestDismissInsight_ValidDraft_ReturnsDismissed(t *testing.T) {
	insightRepo, auditRepo := newInsightTestDeps()
	uc := NewDismissInsightUseCase(insightRepo, auditRepo)

	seedDraftInsight(insightRepo, "i1", "coach-1", "client-1", entities.InsightPriorityHigh)

	card, err := uc.Execute(context.Background(), DismissInsightInput{
		InsightID: "i1",
		CoachID:   "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card.Status != entities.InsightStatusDismissed {
		t.Errorf("expected status dismissed, got %s", card.Status)
	}

	if auditRepo.EventCount() != 1 {
		t.Errorf("expected 1 audit event, got %d", auditRepo.EventCount())
	}
}

func TestDismissInsight_WrongCoach_ReturnsAuthorizationError(t *testing.T) {
	insightRepo, auditRepo := newInsightTestDeps()
	uc := NewDismissInsightUseCase(insightRepo, auditRepo)

	seedDraftInsight(insightRepo, "i1", "coach-1", "client-1", entities.InsightPriorityHigh)

	_, err := uc.Execute(context.Background(), DismissInsightInput{
		InsightID: "i1",
		CoachID:   "coach-2",
	})
	if !IsAuthorizationError(err) {
		t.Errorf("expected AuthorizationError, got %T: %v", err, err)
	}
}

func TestDismissInsight_AlreadyApproved_ReturnsStatusTransitionError(t *testing.T) {
	insightRepo, auditRepo := newInsightTestDeps()
	uc := NewDismissInsightUseCase(insightRepo, auditRepo)

	insightRepo.Seed(&entities.InsightCard{
		ID:       "i1",
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Approved",
		Body:     "body",
		Category: entities.InsightCategoryGeneral,
		Status:   entities.InsightStatusApproved,
		Priority: entities.InsightPriorityMedium,
	})

	_, err := uc.Execute(context.Background(), DismissInsightInput{
		InsightID: "i1",
		CoachID:   "coach-1",
	})
	if err == nil {
		t.Fatal("expected error for invalid transition")
	}
}

// --- EditInsightUseCase Tests ---

func TestEditInsight_ValidInput_UpdatesTitleAndBody(t *testing.T) {
	insightRepo, auditRepo := newInsightTestDeps()
	uc := NewEditInsightUseCase(insightRepo, auditRepo)

	seedDraftInsight(insightRepo, "i1", "coach-1", "client-1", entities.InsightPriorityHigh)

	card, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: "i1",
		CoachID:   "coach-1",
		Title:     "Updated Title",
		Body:      "Updated Body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %q", card.Title)
	}
	if card.Body != "Updated Body" {
		t.Errorf("expected body 'Updated Body', got %q", card.Body)
	}

	if auditRepo.EventCount() != 1 {
		t.Errorf("expected 1 audit event, got %d", auditRepo.EventCount())
	}
}

func TestEditInsight_TitleOnly_UpdatesTitle(t *testing.T) {
	insightRepo, auditRepo := newInsightTestDeps()
	uc := NewEditInsightUseCase(insightRepo, auditRepo)

	seedDraftInsight(insightRepo, "i1", "coach-1", "client-1", entities.InsightPriorityHigh)

	card, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: "i1",
		CoachID:   "coach-1",
		Title:     "New Title",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card.Title != "New Title" {
		t.Errorf("expected title 'New Title', got %q", card.Title)
	}
}

func TestEditInsight_EmptyTitleAndBody_ReturnsValidationError(t *testing.T) {
	insightRepo, auditRepo := newInsightTestDeps()
	uc := NewEditInsightUseCase(insightRepo, auditRepo)

	seedDraftInsight(insightRepo, "i1", "coach-1", "client-1", entities.InsightPriorityHigh)

	_, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: "i1",
		CoachID:   "coach-1",
		Title:     "",
		Body:      "",
	})
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestEditInsight_DismissedInsight_ReturnsValidationError(t *testing.T) {
	insightRepo, auditRepo := newInsightTestDeps()
	uc := NewEditInsightUseCase(insightRepo, auditRepo)

	insightRepo.Seed(&entities.InsightCard{
		ID:       "i1",
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Dismissed",
		Body:     "body",
		Category: entities.InsightCategoryGeneral,
		Status:   entities.InsightStatusDismissed,
		Priority: entities.InsightPriorityMedium,
	})

	_, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: "i1",
		CoachID:   "coach-1",
		Title:     "New Title",
	})
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestEditInsight_WrongCoach_ReturnsAuthorizationError(t *testing.T) {
	insightRepo, auditRepo := newInsightTestDeps()
	uc := NewEditInsightUseCase(insightRepo, auditRepo)

	seedDraftInsight(insightRepo, "i1", "coach-1", "client-1", entities.InsightPriorityHigh)

	_, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: "i1",
		CoachID:   "coach-2",
		Title:     "New Title",
	})
	if !IsAuthorizationError(err) {
		t.Errorf("expected AuthorizationError, got %T: %v", err, err)
	}
}

// --- ShareInsightUseCase Tests ---

func TestShareInsight_ApprovedInsight_ReturnsShared(t *testing.T) {
	insightRepo, auditRepo := newInsightTestDeps()
	uc := NewShareInsightUseCase(insightRepo, auditRepo)

	insightRepo.Seed(&entities.InsightCard{
		ID:       "i1",
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Approved Insight",
		Body:     "body",
		Category: entities.InsightCategoryGeneral,
		Status:   entities.InsightStatusApproved,
		Priority: entities.InsightPriorityMedium,
	})

	card, err := uc.Execute(context.Background(), ShareInsightInput{
		InsightID: "i1",
		CoachID:   "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card.Status != entities.InsightStatusShared {
		t.Errorf("expected status shared, got %s", card.Status)
	}
	if card.SharedAt == nil {
		t.Error("expected shared_at to be set")
	}

	if auditRepo.EventCount() != 1 {
		t.Errorf("expected 1 audit event, got %d", auditRepo.EventCount())
	}
}

func TestShareInsight_DraftInsight_ReturnsStatusTransitionError(t *testing.T) {
	insightRepo, auditRepo := newInsightTestDeps()
	uc := NewShareInsightUseCase(insightRepo, auditRepo)

	seedDraftInsight(insightRepo, "i1", "coach-1", "client-1", entities.InsightPriorityHigh)

	_, err := uc.Execute(context.Background(), ShareInsightInput{
		InsightID: "i1",
		CoachID:   "coach-1",
	})
	if err == nil {
		t.Fatal("expected error for sharing draft insight")
	}
}

func TestShareInsight_WrongCoach_ReturnsAuthorizationError(t *testing.T) {
	insightRepo, auditRepo := newInsightTestDeps()
	uc := NewShareInsightUseCase(insightRepo, auditRepo)

	insightRepo.Seed(&entities.InsightCard{
		ID:       "i1",
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Approved Insight",
		Body:     "body",
		Category: entities.InsightCategoryGeneral,
		Status:   entities.InsightStatusApproved,
		Priority: entities.InsightPriorityMedium,
	})

	_, err := uc.Execute(context.Background(), ShareInsightInput{
		InsightID: "i1",
		CoachID:   "coach-2",
	})
	if !IsAuthorizationError(err) {
		t.Errorf("expected AuthorizationError, got %T: %v", err, err)
	}
}

// --- Integration: Edit then Approve flow ---

func TestEditThenApprove_HappyPath(t *testing.T) {
	insightRepo, auditRepo := newInsightTestDeps()
	editUC := NewEditInsightUseCase(insightRepo, auditRepo)
	approveUC := NewApproveInsightUseCase(insightRepo, auditRepo)

	seedDraftInsight(insightRepo, "i1", "coach-1", "client-1", entities.InsightPriorityHigh)

	// Edit
	edited, err := editUC.Execute(context.Background(), EditInsightInput{
		InsightID: "i1",
		CoachID:   "coach-1",
		Title:     "Edited Title",
		Body:      "Edited Body",
	})
	if err != nil {
		t.Fatalf("edit error: %v", err)
	}

	// Approve
	approved, err := approveUC.Execute(context.Background(), ApproveInsightInput{
		InsightID: "i1",
		CoachID:   "coach-1",
	})
	if err != nil {
		t.Fatalf("approve error: %v", err)
	}

	if approved.Title != edited.Title {
		t.Errorf("expected title %q, got %q", edited.Title, approved.Title)
	}
	if approved.Status != entities.InsightStatusApproved {
		t.Errorf("expected status approved, got %s", approved.Status)
	}

	// 2 audit events: edit + approve
	if auditRepo.EventCount() != 2 {
		t.Errorf("expected 2 audit events, got %d", auditRepo.EventCount())
	}
}
