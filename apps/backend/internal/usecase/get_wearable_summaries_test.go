package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestGetWearableSummaries_Success_ReturnsSummaries(t *testing.T) {
	wearableRepo := &mockWearableRepo{
		findByClientFunc: func(ctx context.Context, clientID string, limit, offset int) ([]*entities.WearableSummary, error) {
			return []*entities.WearableSummary{
				{ID: "ws1", ClientID: clientID, Source: "whoop", SummaryDate: "2026-01-15"},
				{ID: "ws2", ClientID: clientID, Source: "whoop", SummaryDate: "2026-01-14"},
			}, nil
		},
	}
	auditRepo := &mockAuditRepo{}
	uc := NewGetWearableSummariesUseCase(wearableRepo, authorizedCCRepo(), auditRepo)

	out, err := uc.Execute(context.Background(), "coach-1", "client-1", 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Summaries) != 2 {
		t.Errorf("expected 2 summaries, got %d", len(out.Summaries))
	}
}

func TestGetWearableSummaries_Unauthorized_ReturnsForbidden(t *testing.T) {
	wearableRepo := &mockWearableRepo{}
	auditRepo := &mockAuditRepo{}
	uc := NewGetWearableSummariesUseCase(wearableRepo, unauthorizedCCRepo(), auditRepo)

	_, err := uc.Execute(context.Background(), "coach-1", "client-1", 20, 0)
	if err == nil {
		t.Fatal("expected error for unauthorized access")
	}
	if !IsAuthenticationError(err) {
		t.Errorf("expected AuthenticationError, got %T", err)
	}
}

func TestGetWearableSummaries_EmptyCoachID_ReturnsValidationError(t *testing.T) {
	uc := NewGetWearableSummariesUseCase(&mockWearableRepo{}, authorizedCCRepo(), &mockAuditRepo{})
	_, err := uc.Execute(context.Background(), "", "client-1", 20, 0)
	if err == nil {
		t.Fatal("expected error for empty coach_id")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetWearableSummaries_EmptyClientID_ReturnsValidationError(t *testing.T) {
	uc := NewGetWearableSummariesUseCase(&mockWearableRepo{}, authorizedCCRepo(), &mockAuditRepo{})
	_, err := uc.Execute(context.Background(), "coach-1", "", 20, 0)
	if err == nil {
		t.Fatal("expected error for empty client_id")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetWearableSummaries_EmptyData_ReturnsEmptyArray(t *testing.T) {
	wearableRepo := &mockWearableRepo{
		findByClientFunc: func(ctx context.Context, clientID string, limit, offset int) ([]*entities.WearableSummary, error) {
			return nil, nil
		},
	}
	uc := NewGetWearableSummariesUseCase(wearableRepo, authorizedCCRepo(), &mockAuditRepo{})
	out, err := uc.Execute(context.Background(), "coach-1", "client-1", 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Summaries == nil {
		t.Error("expected non-nil summaries array")
	}
	if len(out.Summaries) != 0 {
		t.Errorf("expected 0 summaries, got %d", len(out.Summaries))
	}
}

func TestGetWearableSummaries_RepoError_Propagates(t *testing.T) {
	repoErr := errors.New("database unavailable")
	wearableRepo := &mockWearableRepo{
		findByClientFunc: func(ctx context.Context, clientID string, limit, offset int) ([]*entities.WearableSummary, error) {
			return nil, repoErr
		},
	}
	uc := NewGetWearableSummariesUseCase(wearableRepo, authorizedCCRepo(), &mockAuditRepo{})
	_, err := uc.Execute(context.Background(), "coach-1", "client-1", 20, 0)
	if err == nil {
		t.Fatal("expected error from repository")
	}
}

func TestGetWearableSummaries_DefaultLimit_Applied(t *testing.T) {
	var capturedLimit int
	wearableRepo := &mockWearableRepo{
		findByClientFunc: func(ctx context.Context, clientID string, limit, offset int) ([]*entities.WearableSummary, error) {
			capturedLimit = limit
			return []*entities.WearableSummary{}, nil
		},
	}
	uc := NewGetWearableSummariesUseCase(wearableRepo, authorizedCCRepo(), &mockAuditRepo{})
	_, err := uc.Execute(context.Background(), "coach-1", "client-1", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedLimit != 20 {
		t.Errorf("expected default limit 20, got %d", capturedLimit)
	}
}

func TestGetWearableSummaries_AuditEvent_Logged(t *testing.T) {
	auditRepo := &mockAuditRepo{}
	uc := NewGetWearableSummariesUseCase(&mockWearableRepo{}, authorizedCCRepo(), auditRepo)
	_, err := uc.Execute(context.Background(), "coach-1", "client-1", 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(auditRepo.events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(auditRepo.events))
	}
	if auditRepo.events[0].Action != "phi.access" {
		t.Errorf("expected phi.access action, got %s", auditRepo.events[0].Action)
	}
}
