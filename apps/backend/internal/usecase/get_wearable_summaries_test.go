package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

type mockWearableRepo struct {
	findByClientIDFunc func(ctx context.Context, clientID string, days int) ([]*entities.WearableSummary, error)
	upsertFunc         func(ctx context.Context, summary *entities.WearableSummary) (*entities.WearableSummary, error)
}

func (m *mockWearableRepo) Upsert(ctx context.Context, summary *entities.WearableSummary) (*entities.WearableSummary, error) {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, summary)
	}
	return summary, nil
}

func (m *mockWearableRepo) FindByClientID(ctx context.Context, clientID string, days int) ([]*entities.WearableSummary, error) {
	if m.findByClientIDFunc != nil {
		return m.findByClientIDFunc(ctx, clientID, days)
	}
	return []*entities.WearableSummary{}, nil
}

func TestGetWearableSummaries_Success(t *testing.T) {
	summaries := []*entities.WearableSummary{
		{ID: "w1", ClientID: "client-1", Source: "whoop", SummaryDate: "2026-01-15", Metrics: map[string]interface{}{"hrv": 45.0}, SyncedAt: time.Now()},
	}

	wearableRepo := &mockWearableRepo{
		findByClientIDFunc: func(ctx context.Context, clientID string, days int) ([]*entities.WearableSummary, error) {
			return summaries, nil
		},
	}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetWearableSummariesUseCase(wearableRepo, ccRepo, auditRepo)
	out, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Summaries) != 1 {
		t.Errorf("expected 1 summary, got %d", len(out.Summaries))
	}
}

func TestGetWearableSummaries_DefaultDays(t *testing.T) {
	var capturedDays int
	wearableRepo := &mockWearableRepo{
		findByClientIDFunc: func(ctx context.Context, clientID string, days int) ([]*entities.WearableSummary, error) {
			capturedDays = days
			return []*entities.WearableSummary{}, nil
		},
	}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetWearableSummariesUseCase(wearableRepo, ccRepo, auditRepo)
	_, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedDays != defaultWearableDays {
		t.Errorf("expected default days %d, got %d", defaultWearableDays, capturedDays)
	}
}

func TestGetWearableSummaries_CustomDays(t *testing.T) {
	var capturedDays int
	wearableRepo := &mockWearableRepo{
		findByClientIDFunc: func(ctx context.Context, clientID string, days int) ([]*entities.WearableSummary, error) {
			capturedDays = days
			return []*entities.WearableSummary{}, nil
		},
	}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetWearableSummariesUseCase(wearableRepo, ccRepo, auditRepo)
	_, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Days:     14,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedDays != 14 {
		t.Errorf("expected days 14, got %d", capturedDays)
	}
}

func TestGetWearableSummaries_Unauthorized(t *testing.T) {
	wearableRepo := &mockWearableRepo{}
	ccRepo := &mockCoachClientRepo{
		findByCoachAndClient: func(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error) {
			return nil, nil
		},
	}
	auditRepo := &mockAuditRepo{}

	uc := NewGetWearableSummariesUseCase(wearableRepo, ccRepo, auditRepo)
	_, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		CoachID:  "coach-1",
		ClientID: "client-2",
	})

	if !IsAuthorizationError(err) {
		t.Errorf("expected AuthorizationError, got %T", err)
	}
}

func TestGetWearableSummaries_EmptyCoachID(t *testing.T) {
	uc := NewGetWearableSummariesUseCase(&mockWearableRepo{}, &mockCoachClientRepo{}, &mockAuditRepo{})
	_, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		ClientID: "client-1",
	})
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetWearableSummaries_EmptyClientID(t *testing.T) {
	uc := NewGetWearableSummariesUseCase(&mockWearableRepo{}, &mockCoachClientRepo{}, &mockAuditRepo{})
	_, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		CoachID: "coach-1",
	})
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetWearableSummaries_RepoError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	wearableRepo := &mockWearableRepo{
		findByClientIDFunc: func(ctx context.Context, clientID string, days int) ([]*entities.WearableSummary, error) {
			return nil, repoErr
		},
	}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetWearableSummariesUseCase(wearableRepo, ccRepo, auditRepo)
	_, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

	if !errors.Is(err, repoErr) {
		t.Errorf("expected repository error, got %v", err)
	}
}

func TestGetWearableSummaries_EmptyData(t *testing.T) {
	wearableRepo := &mockWearableRepo{
		findByClientIDFunc: func(ctx context.Context, clientID string, days int) ([]*entities.WearableSummary, error) {
			return nil, nil
		},
	}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetWearableSummariesUseCase(wearableRepo, ccRepo, auditRepo)
	out, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Summaries == nil {
		t.Error("expected non-nil empty summaries array")
	}
}

func TestGetWearableSummaries_AuditLogged(t *testing.T) {
	wearableRepo := &mockWearableRepo{}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetWearableSummariesUseCase(wearableRepo, ccRepo, auditRepo)
	_, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(auditRepo.events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(auditRepo.events))
	}
	if auditRepo.events[0].Action != "phi.wearable_data_accessed" {
		t.Errorf("expected action 'phi.wearable_data_accessed', got '%s'", auditRepo.events[0].Action)
	}
}
