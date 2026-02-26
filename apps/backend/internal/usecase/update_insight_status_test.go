package usecase

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func TestUpdateInsightStatus_ValidInput_UpdatesStatus(t *testing.T) {
	insightRepo := repository.NewInMemoryInsightCardRepository()
	auditRepo := repository.NewInMemoryAuditRepository()

	card, _ := insightRepo.Create(context.Background(), &entities.InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Test",
		Body:     "body",
		Category: entities.InsightCategoryNutrition,
		Priority: entities.InsightPriorityHigh,
		Status:   entities.InsightStatusDraft,
	})

	uc := NewUpdateInsightStatusUseCase(insightRepo, auditRepo)
	updated, err := uc.Execute(context.Background(), card.ID, entities.InsightStatusApproved)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != entities.InsightStatusApproved {
		t.Errorf("expected status approved, got %s", updated.Status)
	}
}

func TestUpdateInsightStatus_ApproveAction_LogsApproveAudit(t *testing.T) {
	insightRepo := repository.NewInMemoryInsightCardRepository()
	auditRepo := repository.NewInMemoryAuditRepository()

	card, _ := insightRepo.Create(context.Background(), &entities.InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Test",
		Body:     "body",
		Category: entities.InsightCategoryNutrition,
		Priority: entities.InsightPriorityHigh,
		Status:   entities.InsightStatusDraft,
	})

	uc := NewUpdateInsightStatusUseCase(insightRepo, auditRepo)
	_, err := uc.Execute(context.Background(), card.ID, entities.InsightStatusApproved)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := auditRepo.GetEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(events))
	}
	if events[0].Action != "insight.approve" {
		t.Errorf("expected action insight.approve, got %s", events[0].Action)
	}
}

func TestUpdateInsightStatus_DismissAction_LogsRejectAudit(t *testing.T) {
	insightRepo := repository.NewInMemoryInsightCardRepository()
	auditRepo := repository.NewInMemoryAuditRepository()

	card, _ := insightRepo.Create(context.Background(), &entities.InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Test",
		Body:     "body",
		Category: entities.InsightCategoryNutrition,
		Priority: entities.InsightPriorityHigh,
		Status:   entities.InsightStatusDraft,
	})

	uc := NewUpdateInsightStatusUseCase(insightRepo, auditRepo)
	_, err := uc.Execute(context.Background(), card.ID, entities.InsightStatusDismissed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := auditRepo.GetEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(events))
	}
	if events[0].Action != "insight.reject" {
		t.Errorf("expected action insight.reject, got %s", events[0].Action)
	}
}

func TestUpdateInsightStatus_MissingID_ReturnsValidationError(t *testing.T) {
	insightRepo := repository.NewInMemoryInsightCardRepository()
	auditRepo := repository.NewInMemoryAuditRepository()

	uc := NewUpdateInsightStatusUseCase(insightRepo, auditRepo)
	_, err := uc.Execute(context.Background(), "", entities.InsightStatusApproved)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestUpdateInsightStatus_InvalidStatus_ReturnsValidationError(t *testing.T) {
	insightRepo := repository.NewInMemoryInsightCardRepository()
	auditRepo := repository.NewInMemoryAuditRepository()

	uc := NewUpdateInsightStatusUseCase(insightRepo, auditRepo)
	_, err := uc.Execute(context.Background(), "some-id", "invalid")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestUpdateInsightStatus_NotFound_ReturnsError(t *testing.T) {
	insightRepo := repository.NewInMemoryInsightCardRepository()
	auditRepo := repository.NewInMemoryAuditRepository()

	uc := NewUpdateInsightStatusUseCase(insightRepo, auditRepo)
	_, err := uc.Execute(context.Background(), "non-existent", entities.InsightStatusApproved)
	if err == nil {
		t.Fatal("expected error for non-existent card")
	}
}
