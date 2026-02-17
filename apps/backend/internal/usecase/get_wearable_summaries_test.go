package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

type mockWearableSummaryRepo struct {
	upsertFunc        func(ctx context.Context, summary *entities.WearableSummary) (*entities.WearableSummary, error)
	findByClientIDFunc func(ctx context.Context, filter entities.WearableSummaryFilter) ([]*entities.WearableSummary, int, error)
	findAveragesFunc  func(ctx context.Context, clientID string, days int) (*entities.WearableAverages, error)
}

func (m *mockWearableSummaryRepo) Upsert(ctx context.Context, summary *entities.WearableSummary) (*entities.WearableSummary, error) {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, summary)
	}
	return summary, nil
}

func (m *mockWearableSummaryRepo) FindByClientID(ctx context.Context, filter entities.WearableSummaryFilter) ([]*entities.WearableSummary, int, error) {
	if m.findByClientIDFunc != nil {
		return m.findByClientIDFunc(ctx, filter)
	}
	return []*entities.WearableSummary{}, 0, nil
}

func (m *mockWearableSummaryRepo) FindAverages(ctx context.Context, clientID string, days int) (*entities.WearableAverages, error) {
	if m.findAveragesFunc != nil {
		return m.findAveragesFunc(ctx, clientID, days)
	}
	return nil, nil
}

func TestGetWearableSummaries_Success(t *testing.T) {
	now := time.Now()
	wRepo := &mockWearableSummaryRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.WearableSummaryFilter) ([]*entities.WearableSummary, int, error) {
			return []*entities.WearableSummary{
				{ID: "ws1", ClientID: "client-1", Source: "whoop", SummaryDate: now, Metrics: map[string]interface{}{"hrv": 45.2}},
			}, 1, nil
		},
	}
	uc := NewGetWearableSummariesUseCase(wRepo, authorizedCoachClientRepo(), &mockAuditRepo{})

	output, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Page:     1,
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(output.Summaries) != 1 {
		t.Errorf("expected 1 summary, got %d", len(output.Summaries))
	}
	if output.Pagination.Total != 1 {
		t.Errorf("expected total 1, got %d", output.Pagination.Total)
	}
}

func TestGetWearableSummaries_Unauthorized(t *testing.T) {
	ccRepo := &mockCoachClientRepo{
		findByCoachAndClient: func(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error) {
			return nil, nil
		},
	}
	uc := NewGetWearableSummariesUseCase(&mockWearableSummaryRepo{}, ccRepo, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Page:     1,
		Limit:    20,
	})
	if err == nil {
		t.Fatal("expected authorization error")
	}
	if !IsAuthorizationError(err) {
		t.Errorf("expected AuthorizationError, got %T", err)
	}
}

func TestGetWearableSummaries_EmptyCoachID(t *testing.T) {
	uc := NewGetWearableSummariesUseCase(&mockWearableSummaryRepo{}, &mockCoachClientRepo{}, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		CoachID:  "",
		ClientID: "client-1",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetWearableSummaries_EmptyClientID(t *testing.T) {
	uc := NewGetWearableSummariesUseCase(&mockWearableSummaryRepo{}, &mockCoachClientRepo{}, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		CoachID:  "coach-1",
		ClientID: "",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetWearableSummaries_EmptyResult(t *testing.T) {
	wRepo := &mockWearableSummaryRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.WearableSummaryFilter) ([]*entities.WearableSummary, int, error) {
			return nil, 0, nil
		},
	}
	uc := NewGetWearableSummariesUseCase(wRepo, authorizedCoachClientRepo(), &mockAuditRepo{})

	output, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Page:     1,
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Summaries == nil {
		t.Fatal("expected non-nil summaries slice")
	}
	if len(output.Summaries) != 0 {
		t.Errorf("expected 0 summaries, got %d", len(output.Summaries))
	}
}

func TestGetWearableSummaries_RepoError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	wRepo := &mockWearableSummaryRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.WearableSummaryFilter) ([]*entities.WearableSummary, int, error) {
			return nil, 0, repoErr
		},
	}
	uc := NewGetWearableSummariesUseCase(wRepo, authorizedCoachClientRepo(), &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Page:     1,
		Limit:    20,
	})
	if err == nil {
		t.Fatal("expected error from repository")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repository error to propagate, got %v", err)
	}
}

func TestGetWearableSummaries_FilterBySource(t *testing.T) {
	var capturedFilter entities.WearableSummaryFilter
	wRepo := &mockWearableSummaryRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.WearableSummaryFilter) ([]*entities.WearableSummary, int, error) {
			capturedFilter = filter
			return []*entities.WearableSummary{}, 0, nil
		},
	}
	uc := NewGetWearableSummariesUseCase(wRepo, authorizedCoachClientRepo(), &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Source:   "whoop",
		Page:     1,
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedFilter.Source != "whoop" {
		t.Errorf("expected source filter 'whoop', got %q", capturedFilter.Source)
	}
}

func TestGetWearableSummaries_AuditEventLogged(t *testing.T) {
	var auditLogged bool
	auditRepo := &mockAuditRepo{
		logEventFunc: func(ctx context.Context, event *entities.AuditEvent) error {
			auditLogged = true
			if event.Action != "phi.access" {
				t.Errorf("expected action 'phi.access', got %q", event.Action)
			}
			if event.EntityType != "wearable_summary" {
				t.Errorf("expected entity_type 'wearable_summary', got %q", event.EntityType)
			}
			return nil
		},
	}
	wRepo := &mockWearableSummaryRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.WearableSummaryFilter) ([]*entities.WearableSummary, int, error) {
			return []*entities.WearableSummary{}, 0, nil
		},
	}
	uc := NewGetWearableSummariesUseCase(wRepo, authorizedCoachClientRepo(), auditRepo)

	_, err := uc.Execute(context.Background(), GetWearableSummariesInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Page:     1,
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !auditLogged {
		t.Error("expected audit event to be logged")
	}
}
