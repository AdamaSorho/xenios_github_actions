package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func seedInsightRepo() (*repository.InMemoryInsightCardRepository, []*entities.InsightCard) {
	repo := repository.NewInMemoryInsightCardRepository()
	ctx := context.Background()

	cards := []*entities.InsightCard{
		{
			ClientID:   "client-1",
			CoachID:    "coach-1",
			ClientName: "Alice",
			Title:      "High LDL",
			Body:       "LDL is elevated",
			Category:   entities.InsightCategoryNutrition,
			Priority:   entities.InsightPriorityHigh,
			Status:     entities.InsightStatusDraft,
		},
		{
			ClientID:   "client-2",
			CoachID:    "coach-1",
			ClientName: "Bob",
			Title:      "Sleep quality",
			Body:       "Poor sleep patterns",
			Category:   entities.InsightCategorySleep,
			Priority:   entities.InsightPriorityMedium,
			Status:     entities.InsightStatusDraft,
		},
		{
			ClientID:   "client-1",
			CoachID:    "coach-1",
			ClientName: "Alice",
			Title:      "Exercise deficit",
			Body:       "Below target",
			Category:   entities.InsightCategoryExercise,
			Priority:   entities.InsightPriorityUrgent,
			Status:     entities.InsightStatusDraft,
		},
		{
			ClientID:   "client-3",
			CoachID:    "coach-2",
			ClientName: "Charlie",
			Title:      "Stress high",
			Body:       "Elevated cortisol",
			Category:   entities.InsightCategoryStress,
			Priority:   entities.InsightPriorityLow,
			Status:     entities.InsightStatusDraft,
		},
	}

	var created []*entities.InsightCard
	for _, c := range cards {
		result, _ := repo.Create(ctx, c)
		created = append(created, result)
		// Small delay to ensure different CreatedAt for sort testing
		time.Sleep(time.Millisecond)
	}
	return repo, created
}

// --- GetInsightQueueUseCase tests ---

func TestGetInsightQueue_ValidInput_ReturnsDrafts(t *testing.T) {
	repo, _ := seedInsightRepo()
	uc := NewGetInsightQueueUseCase(repo)

	out, err := uc.Execute(context.Background(), GetInsightQueueInput{
		CoachID: "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Insights) != 3 {
		t.Errorf("expected 3 drafts for coach-1, got %d", len(out.Insights))
	}
	if out.Pagination.Total != 3 {
		t.Errorf("expected total 3, got %d", out.Pagination.Total)
	}
}

func TestGetInsightQueue_EmptyCoachID_ReturnsError(t *testing.T) {
	repo, _ := seedInsightRepo()
	uc := NewGetInsightQueueUseCase(repo)

	_, err := uc.Execute(context.Background(), GetInsightQueueInput{})
	if err == nil {
		t.Fatal("expected error for empty coach_id")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetInsightQueue_InvalidStatus_ReturnsError(t *testing.T) {
	repo, _ := seedInsightRepo()
	uc := NewGetInsightQueueUseCase(repo)

	_, err := uc.Execute(context.Background(), GetInsightQueueInput{
		CoachID: "coach-1",
		Status:  "invalid",
	})
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetInsightQueue_NoInsights_ReturnsEmptyList(t *testing.T) {
	repo := repository.NewInMemoryInsightCardRepository()
	uc := NewGetInsightQueueUseCase(repo)

	out, err := uc.Execute(context.Background(), GetInsightQueueInput{
		CoachID: "coach-no-insights",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Insights) != 0 {
		t.Errorf("expected 0 insights, got %d", len(out.Insights))
	}
}

func TestGetInsightQueue_SortedByPriorityThenRecency(t *testing.T) {
	repo, _ := seedInsightRepo()
	uc := NewGetInsightQueueUseCase(repo)

	out, err := uc.Execute(context.Background(), GetInsightQueueInput{
		CoachID: "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out.Insights) < 2 {
		t.Fatal("need at least 2 insights to test sort")
	}

	// First should be urgent priority
	if out.Insights[0].Priority != entities.InsightPriorityUrgent {
		t.Errorf("expected first insight to be urgent, got %s", out.Insights[0].Priority)
	}
}

func TestGetInsightQueue_Pagination_RespectsLimits(t *testing.T) {
	repo, _ := seedInsightRepo()
	uc := NewGetInsightQueueUseCase(repo)

	out, err := uc.Execute(context.Background(), GetInsightQueueInput{
		CoachID: "coach-1",
		Page:    1,
		Limit:   2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Insights) != 2 {
		t.Errorf("expected 2 insights on page 1, got %d", len(out.Insights))
	}
	if out.Pagination.Total != 3 {
		t.Errorf("expected total 3, got %d", out.Pagination.Total)
	}
}

func TestGetInsightQueue_DefaultsPageAndLimit(t *testing.T) {
	repo, _ := seedInsightRepo()
	uc := NewGetInsightQueueUseCase(repo)

	out, err := uc.Execute(context.Background(), GetInsightQueueInput{
		CoachID: "coach-1",
		Page:    0,
		Limit:   0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Pagination.Page != 1 {
		t.Errorf("expected default page 1, got %d", out.Pagination.Page)
	}
	if out.Pagination.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", out.Pagination.Limit)
	}
}

// --- GetClientInsightsUseCase tests ---

func TestGetClientInsights_ValidInput_ReturnsInsights(t *testing.T) {
	repo, _ := seedInsightRepo()
	uc := NewGetClientInsightsUseCase(repo)

	out, err := uc.Execute(context.Background(), GetClientInsightsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Insights) != 2 {
		t.Errorf("expected 2 insights for client-1, got %d", len(out.Insights))
	}
}

func TestGetClientInsights_EmptyClientID_ReturnsError(t *testing.T) {
	repo, _ := seedInsightRepo()
	uc := NewGetClientInsightsUseCase(repo)

	_, err := uc.Execute(context.Background(), GetClientInsightsInput{
		CoachID: "coach-1",
	})
	if err == nil {
		t.Fatal("expected error for empty client_id")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetClientInsights_EmptyCoachID_ReturnsError(t *testing.T) {
	repo, _ := seedInsightRepo()
	uc := NewGetClientInsightsUseCase(repo)

	_, err := uc.Execute(context.Background(), GetClientInsightsInput{
		ClientID: "client-1",
	})
	if err == nil {
		t.Fatal("expected error for empty coach_id")
	}
}

func TestGetClientInsights_InvalidStatus_ReturnsError(t *testing.T) {
	repo, _ := seedInsightRepo()
	uc := NewGetClientInsightsUseCase(repo)

	_, err := uc.Execute(context.Background(), GetClientInsightsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Status:   "bogus",
	})
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

// --- ApproveInsightUseCase tests ---

func TestApproveInsight_ValidDraft_ReturnsApproved(t *testing.T) {
	repo, cards := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewApproveInsightUseCase(repo, auditRepo)

	result, err := uc.Execute(context.Background(), InsightActionInput{
		InsightID: cards[0].ID,
		CoachID:   "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != entities.InsightStatusApproved {
		t.Errorf("expected approved, got %s", result.Status)
	}
	if result.ApprovedAt == nil {
		t.Error("expected approved_at to be set")
	}
}

func TestApproveInsight_EmptyInsightID_ReturnsError(t *testing.T) {
	repo, _ := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewApproveInsightUseCase(repo, auditRepo)

	_, err := uc.Execute(context.Background(), InsightActionInput{CoachID: "coach-1"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestApproveInsight_EmptyCoachID_ReturnsError(t *testing.T) {
	repo, cards := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewApproveInsightUseCase(repo, auditRepo)

	_, err := uc.Execute(context.Background(), InsightActionInput{InsightID: cards[0].ID})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestApproveInsight_NotFound_ReturnsError(t *testing.T) {
	repo, _ := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewApproveInsightUseCase(repo, auditRepo)

	_, err := uc.Execute(context.Background(), InsightActionInput{
		InsightID: "nonexistent",
		CoachID:   "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestApproveInsight_WrongCoach_ReturnsAuthError(t *testing.T) {
	repo, cards := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewApproveInsightUseCase(repo, auditRepo)

	_, err := uc.Execute(context.Background(), InsightActionInput{
		InsightID: cards[0].ID,
		CoachID:   "coach-2",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsAuthenticationError(err) {
		t.Errorf("expected AuthenticationError, got %T", err)
	}
}

func TestApproveInsight_AlreadyDismissed_ReturnsError(t *testing.T) {
	repo, cards := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()

	// Dismiss first
	dismissUC := NewDismissInsightUseCase(repo, auditRepo)
	_, _ = dismissUC.Execute(context.Background(), InsightActionInput{
		InsightID: cards[0].ID,
		CoachID:   "coach-1",
	})

	// Try to approve dismissed
	approveUC := NewApproveInsightUseCase(repo, auditRepo)
	_, err := approveUC.Execute(context.Background(), InsightActionInput{
		InsightID: cards[0].ID,
		CoachID:   "coach-1",
	})
	if err == nil {
		t.Fatal("expected error for dismissed insight")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

// --- DismissInsightUseCase tests ---

func TestDismissInsight_ValidDraft_ReturnsDismissed(t *testing.T) {
	repo, cards := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewDismissInsightUseCase(repo, auditRepo)

	result, err := uc.Execute(context.Background(), InsightActionInput{
		InsightID: cards[0].ID,
		CoachID:   "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != entities.InsightStatusDismissed {
		t.Errorf("expected dismissed, got %s", result.Status)
	}
	if result.DismissedAt == nil {
		t.Error("expected dismissed_at to be set")
	}
}

func TestDismissInsight_EmptyInsightID_ReturnsError(t *testing.T) {
	repo, _ := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewDismissInsightUseCase(repo, auditRepo)

	_, err := uc.Execute(context.Background(), InsightActionInput{CoachID: "coach-1"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDismissInsight_WrongCoach_ReturnsAuthError(t *testing.T) {
	repo, cards := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewDismissInsightUseCase(repo, auditRepo)

	_, err := uc.Execute(context.Background(), InsightActionInput{
		InsightID: cards[0].ID,
		CoachID:   "coach-wrong",
	})
	if !IsAuthenticationError(err) {
		t.Errorf("expected AuthenticationError, got %T", err)
	}
}

func TestDismissInsight_NotFound_ReturnsError(t *testing.T) {
	repo, _ := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewDismissInsightUseCase(repo, auditRepo)

	_, err := uc.Execute(context.Background(), InsightActionInput{
		InsightID: "nonexistent",
		CoachID:   "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDismissInsight_AlreadyApproved_ReturnsError(t *testing.T) {
	repo, cards := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()

	// Approve first
	approveUC := NewApproveInsightUseCase(repo, auditRepo)
	_, _ = approveUC.Execute(context.Background(), InsightActionInput{
		InsightID: cards[1].ID,
		CoachID:   "coach-1",
	})

	// Try to dismiss approved
	dismissUC := NewDismissInsightUseCase(repo, auditRepo)
	_, err := dismissUC.Execute(context.Background(), InsightActionInput{
		InsightID: cards[1].ID,
		CoachID:   "coach-1",
	})
	if err == nil {
		t.Fatal("expected error for approved insight")
	}
}

// --- EditInsightUseCase tests ---

func TestEditInsight_ValidInput_UpdatesTitleAndBody(t *testing.T) {
	repo, cards := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewEditInsightUseCase(repo, auditRepo)

	result, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: cards[0].ID,
		CoachID:   "coach-1",
		Title:     "Updated Title",
		Body:      "Updated body text",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %s", result.Title)
	}
	if result.Body != "Updated body text" {
		t.Errorf("expected body 'Updated body text', got %s", result.Body)
	}
}

func TestEditInsight_OnlyTitle_UpdatesTitle(t *testing.T) {
	repo, cards := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewEditInsightUseCase(repo, auditRepo)

	originalBody := cards[0].Body
	result, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: cards[0].ID,
		CoachID:   "coach-1",
		Title:     "New Title",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Title != "New Title" {
		t.Errorf("expected title 'New Title', got %s", result.Title)
	}
	if result.Body != originalBody {
		t.Error("body should not have changed")
	}
}

func TestEditInsight_EmptyBoth_ReturnsError(t *testing.T) {
	repo, cards := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewEditInsightUseCase(repo, auditRepo)

	_, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: cards[0].ID,
		CoachID:   "coach-1",
	})
	if err == nil {
		t.Fatal("expected error for empty title and body")
	}
}

func TestEditInsight_WrongCoach_ReturnsAuthError(t *testing.T) {
	repo, cards := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewEditInsightUseCase(repo, auditRepo)

	_, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: cards[0].ID,
		CoachID:   "coach-wrong",
		Title:     "New Title",
	})
	if !IsAuthenticationError(err) {
		t.Errorf("expected AuthenticationError, got %T", err)
	}
}

func TestEditInsight_TerminalStatus_ReturnsError(t *testing.T) {
	repo, cards := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()

	// Dismiss the insight
	dismissUC := NewDismissInsightUseCase(repo, auditRepo)
	_, _ = dismissUC.Execute(context.Background(), InsightActionInput{
		InsightID: cards[0].ID,
		CoachID:   "coach-1",
	})

	// Try to edit dismissed
	editUC := NewEditInsightUseCase(repo, auditRepo)
	_, err := editUC.Execute(context.Background(), EditInsightInput{
		InsightID: cards[0].ID,
		CoachID:   "coach-1",
		Title:     "New Title",
	})
	if err == nil {
		t.Fatal("expected error for terminal status")
	}
}

func TestEditInsight_NotFound_ReturnsError(t *testing.T) {
	repo, _ := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewEditInsightUseCase(repo, auditRepo)

	_, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: "nonexistent",
		CoachID:   "coach-1",
		Title:     "New Title",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEditInsight_EmptyInsightID_ReturnsError(t *testing.T) {
	repo, _ := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewEditInsightUseCase(repo, auditRepo)

	_, err := uc.Execute(context.Background(), EditInsightInput{
		CoachID: "coach-1",
		Title:   "X",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEditInsight_EmptyCoachID_ReturnsError(t *testing.T) {
	repo, cards := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewEditInsightUseCase(repo, auditRepo)

	_, err := uc.Execute(context.Background(), EditInsightInput{
		InsightID: cards[0].ID,
		Title:     "X",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- ShareInsightUseCase tests ---

func TestShareInsight_ApprovedInsight_ReturnsShared(t *testing.T) {
	repo, cards := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()

	// Approve first
	approveUC := NewApproveInsightUseCase(repo, auditRepo)
	_, _ = approveUC.Execute(context.Background(), InsightActionInput{
		InsightID: cards[0].ID,
		CoachID:   "coach-1",
	})

	// Share
	shareUC := NewShareInsightUseCase(repo, auditRepo)
	result, err := shareUC.Execute(context.Background(), InsightActionInput{
		InsightID: cards[0].ID,
		CoachID:   "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != entities.InsightStatusShared {
		t.Errorf("expected shared, got %s", result.Status)
	}
	if result.SharedAt == nil {
		t.Error("expected shared_at to be set")
	}
}

func TestShareInsight_DraftInsight_ReturnsError(t *testing.T) {
	repo, cards := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewShareInsightUseCase(repo, auditRepo)

	_, err := uc.Execute(context.Background(), InsightActionInput{
		InsightID: cards[0].ID,
		CoachID:   "coach-1",
	})
	if err == nil {
		t.Fatal("expected error for draft insight")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestShareInsight_WrongCoach_ReturnsAuthError(t *testing.T) {
	repo, cards := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()

	// Approve first
	approveUC := NewApproveInsightUseCase(repo, auditRepo)
	_, _ = approveUC.Execute(context.Background(), InsightActionInput{
		InsightID: cards[0].ID,
		CoachID:   "coach-1",
	})

	shareUC := NewShareInsightUseCase(repo, auditRepo)
	_, err := shareUC.Execute(context.Background(), InsightActionInput{
		InsightID: cards[0].ID,
		CoachID:   "coach-wrong",
	})
	if !IsAuthenticationError(err) {
		t.Errorf("expected AuthenticationError, got %T", err)
	}
}

func TestShareInsight_NotFound_ReturnsError(t *testing.T) {
	repo, _ := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewShareInsightUseCase(repo, auditRepo)

	_, err := uc.Execute(context.Background(), InsightActionInput{
		InsightID: "nonexistent",
		CoachID:   "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestShareInsight_EmptyInsightID_ReturnsError(t *testing.T) {
	repo, _ := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewShareInsightUseCase(repo, auditRepo)

	_, err := uc.Execute(context.Background(), InsightActionInput{
		CoachID: "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestShareInsight_EmptyCoachID_ReturnsError(t *testing.T) {
	repo, cards := seedInsightRepo()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewShareInsightUseCase(repo, auditRepo)

	_, err := uc.Execute(context.Background(), InsightActionInput{
		InsightID: cards[0].ID,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- InMemoryInsightCardRepository tests ---

func TestInMemoryInsightCardRepository_Create_ReturnsCard(t *testing.T) {
	repo := repository.NewInMemoryInsightCardRepository()
	card, err := repo.Create(context.Background(), &entities.InsightCard{
		ClientID: "c1",
		CoachID:  "co1",
		Title:    "Test",
		Status:   entities.InsightStatusDraft,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card.ID == "" {
		t.Error("expected non-empty ID")
	}
	if card.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestInMemoryInsightCardRepository_FindByID_Found(t *testing.T) {
	repo := repository.NewInMemoryInsightCardRepository()
	created, _ := repo.Create(context.Background(), &entities.InsightCard{
		ClientID: "c1",
		CoachID:  "co1",
		Title:    "Test",
		Status:   entities.InsightStatusDraft,
	})

	found, err := repo.FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil {
		t.Fatal("expected non-nil card")
	}
	if found.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, found.ID)
	}
}

func TestInMemoryInsightCardRepository_FindByID_NotFound(t *testing.T) {
	repo := repository.NewInMemoryInsightCardRepository()
	found, err := repo.FindByID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Error("expected nil card")
	}
}

func TestInMemoryInsightCardRepository_Update_ReturnsUpdated(t *testing.T) {
	repo := repository.NewInMemoryInsightCardRepository()
	created, _ := repo.Create(context.Background(), &entities.InsightCard{
		ClientID: "c1",
		CoachID:  "co1",
		Title:    "Original",
		Status:   entities.InsightStatusDraft,
	})

	created.Title = "Updated"
	updated, err := repo.Update(context.Background(), created)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Title != "Updated" {
		t.Errorf("expected title 'Updated', got %s", updated.Title)
	}
}

func TestInMemoryInsightCardRepository_Update_NotFound(t *testing.T) {
	repo := repository.NewInMemoryInsightCardRepository()
	result, err := repo.Update(context.Background(), &entities.InsightCard{ID: "nonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil for non-existent card")
	}
}

func TestInMemoryInsightCardRepository_ListByCoachID_Filtered(t *testing.T) {
	repo, _ := seedInsightRepo()

	results, total, err := repo.ListByCoachID(context.Background(), entities.InsightQueryFilter{
		CoachID: "coach-1",
		Status:  "draft",
		Page:    1,
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 3 {
		t.Errorf("expected 3 total, got %d", total)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestInMemoryInsightCardRepository_ListByClientID_Filtered(t *testing.T) {
	repo, _ := seedInsightRepo()

	results, total, err := repo.ListByClientID(context.Background(), entities.InsightQueryFilter{
		ClientID: "client-1",
		Page:     1,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2 total, got %d", total)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}
