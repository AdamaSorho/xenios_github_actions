package usecase

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func newInsightTransitionUseCase() (*InsightTransitionUseCase, *repository.InMemoryInsightCardRepository, *repository.InMemoryAuditRepository) {
	insightRepo := repository.NewInMemoryInsightCardRepository()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewInsightTransitionUseCase(insightRepo, auditRepo)
	return uc, insightRepo, auditRepo
}

func seedDraftInsight(repo *repository.InMemoryInsightCardRepository, coachID, clientID string) *entities.InsightCard {
	insight, _ := repo.Create(context.Background(), &entities.InsightCard{
		CoachID:    coachID,
		ClientID:   clientID,
		ClientName: "Test Client",
		Title:      "Test Insight",
		Body:       "Test body content",
		Category:   "nutrition",
		Priority:   entities.InsightPriorityMedium,
		Status:     entities.InsightStatusDraft,
	})
	return insight
}

func seedApprovedInsight(repo *repository.InMemoryInsightCardRepository, coachID, clientID string) *entities.InsightCard {
	insight, _ := repo.Create(context.Background(), &entities.InsightCard{
		CoachID:    coachID,
		ClientID:   clientID,
		ClientName: "Test Client",
		Title:      "Approved Insight",
		Body:       "Approved body",
		Category:   "fitness",
		Priority:   entities.InsightPriorityHigh,
		Status:     entities.InsightStatusApproved,
	})
	return insight
}

func TestInsightTransition_Approve_DraftToApproved_Succeeds(t *testing.T) {
	uc, insightRepo, _ := newInsightTransitionUseCase()
	insight := seedDraftInsight(insightRepo, "coach-1", "client-1")

	result, err := uc.Execute(context.Background(), TransitionInput{
		InsightID:    insight.ID,
		CoachID:      "coach-1",
		TargetStatus: entities.InsightStatusApproved,
		AuditAction:  "insight.approved",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != entities.InsightStatusApproved {
		t.Errorf("expected status 'approved', got %q", result.Status)
	}
	if result.ApprovedAt == nil {
		t.Error("expected ApprovedAt to be set")
	}
}

func TestInsightTransition_Dismiss_DraftToDismissed_Succeeds(t *testing.T) {
	uc, insightRepo, _ := newInsightTransitionUseCase()
	insight := seedDraftInsight(insightRepo, "coach-1", "client-1")

	result, err := uc.Execute(context.Background(), TransitionInput{
		InsightID:    insight.ID,
		CoachID:      "coach-1",
		TargetStatus: entities.InsightStatusDismissed,
		AuditAction:  "insight.dismissed",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != entities.InsightStatusDismissed {
		t.Errorf("expected status 'dismissed', got %q", result.Status)
	}
}

func TestInsightTransition_Share_ApprovedToShared_Succeeds(t *testing.T) {
	uc, insightRepo, _ := newInsightTransitionUseCase()
	insight := seedApprovedInsight(insightRepo, "coach-1", "client-1")

	result, err := uc.Execute(context.Background(), TransitionInput{
		InsightID:    insight.ID,
		CoachID:      "coach-1",
		TargetStatus: entities.InsightStatusShared,
		AuditAction:  "insight.shared",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != entities.InsightStatusShared {
		t.Errorf("expected status 'shared', got %q", result.Status)
	}
	if result.SharedAt == nil {
		t.Error("expected SharedAt to be set")
	}
}

func TestInsightTransition_InvalidTransition_ReturnsTransitionError(t *testing.T) {
	uc, insightRepo, _ := newInsightTransitionUseCase()
	insight := seedDraftInsight(insightRepo, "coach-1", "client-1")

	_, err := uc.Execute(context.Background(), TransitionInput{
		InsightID:    insight.ID,
		CoachID:      "coach-1",
		TargetStatus: entities.InsightStatusShared,
		AuditAction:  "insight.shared",
	})
	if err == nil {
		t.Fatal("expected error for invalid transition")
	}
	if !IsTransitionError(err) {
		t.Errorf("expected TransitionError, got %T: %v", err, err)
	}
}

func TestInsightTransition_WrongCoach_ReturnsAuthError(t *testing.T) {
	uc, insightRepo, _ := newInsightTransitionUseCase()
	insight := seedDraftInsight(insightRepo, "coach-1", "client-1")

	_, err := uc.Execute(context.Background(), TransitionInput{
		InsightID:    insight.ID,
		CoachID:      "coach-other",
		TargetStatus: entities.InsightStatusApproved,
		AuditAction:  "insight.approved",
	})
	if err == nil {
		t.Fatal("expected error for wrong coach")
	}
	if !IsAuthenticationError(err) {
		t.Errorf("expected AuthenticationError, got %T: %v", err, err)
	}
}

func TestInsightTransition_NotFound_ReturnsValidationError(t *testing.T) {
	uc, _, _ := newInsightTransitionUseCase()

	_, err := uc.Execute(context.Background(), TransitionInput{
		InsightID:    "nonexistent",
		CoachID:      "coach-1",
		TargetStatus: entities.InsightStatusApproved,
		AuditAction:  "insight.approved",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent insight")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestInsightTransition_EmptyInsightID_ReturnsValidationError(t *testing.T) {
	uc, _, _ := newInsightTransitionUseCase()

	_, err := uc.Execute(context.Background(), TransitionInput{
		InsightID:    "",
		CoachID:      "coach-1",
		TargetStatus: entities.InsightStatusApproved,
		AuditAction:  "insight.approved",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestInsightTransition_EmptyCoachID_ReturnsValidationError(t *testing.T) {
	uc, _, _ := newInsightTransitionUseCase()

	_, err := uc.Execute(context.Background(), TransitionInput{
		InsightID:    "insight-1",
		CoachID:      "",
		TargetStatus: entities.InsightStatusApproved,
		AuditAction:  "insight.approved",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestInsightTransition_Approve_LogsAuditEvent(t *testing.T) {
	uc, insightRepo, auditRepo := newInsightTransitionUseCase()
	insight := seedDraftInsight(insightRepo, "coach-1", "client-1")

	_, err := uc.Execute(context.Background(), TransitionInput{
		InsightID:    insight.ID,
		CoachID:      "coach-1",
		TargetStatus: entities.InsightStatusApproved,
		AuditAction:  "insight.approved",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := auditRepo.GetEvents()
	if len(events) == 0 {
		t.Fatal("expected audit event to be logged")
	}
	if events[0].Action != "insight.approved" {
		t.Errorf("expected action 'insight.approved', got %q", events[0].Action)
	}
	if events[0].ActorID != "coach-1" {
		t.Errorf("expected actor 'coach-1', got %q", events[0].ActorID)
	}
}

func TestInsightTransition_DismissedInsight_CannotApprove(t *testing.T) {
	uc, insightRepo, _ := newInsightTransitionUseCase()
	insight := seedDraftInsight(insightRepo, "coach-1", "client-1")

	// First dismiss
	uc.Execute(context.Background(), TransitionInput{
		InsightID:    insight.ID,
		CoachID:      "coach-1",
		TargetStatus: entities.InsightStatusDismissed,
		AuditAction:  "insight.dismissed",
	})

	// Try to approve dismissed insight
	_, err := uc.Execute(context.Background(), TransitionInput{
		InsightID:    insight.ID,
		CoachID:      "coach-1",
		TargetStatus: entities.InsightStatusApproved,
		AuditAction:  "insight.approved",
	})
	if err == nil {
		t.Fatal("expected error for dismissed → approved transition")
	}
	if !IsTransitionError(err) {
		t.Errorf("expected TransitionError, got %T", err)
	}
}
