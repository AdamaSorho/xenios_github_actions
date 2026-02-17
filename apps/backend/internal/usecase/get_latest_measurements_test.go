package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestGetLatestMeasurements_Success(t *testing.T) {
	now := time.Now()
	mRepo := &mockMeasurementRepo{
		findLatestByClientIDFunc: func(ctx context.Context, clientID string) ([]*entities.LatestMeasurement, error) {
			return []*entities.LatestMeasurement{
				{MeasurementType: "weight", Value: 185.4, Unit: "lbs", MeasuredAt: now},
				{MeasurementType: "body_fat_pct", Value: 22.3, Unit: "%", MeasuredAt: now},
			}, nil
		},
	}
	uc := NewGetLatestMeasurementsUseCase(mRepo, authorizedCoachClientRepo(), &mockAuditRepo{})

	results, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestGetLatestMeasurements_Unauthorized(t *testing.T) {
	ccRepo := &mockCoachClientRepo{
		findByCoachAndClient: func(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error) {
			return nil, nil
		},
	}
	uc := NewGetLatestMeasurementsUseCase(&mockMeasurementRepo{}, ccRepo, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err == nil {
		t.Fatal("expected authorization error")
	}
	if !IsAuthorizationError(err) {
		t.Errorf("expected AuthorizationError, got %T", err)
	}
}

func TestGetLatestMeasurements_EmptyCoachID(t *testing.T) {
	uc := NewGetLatestMeasurementsUseCase(&mockMeasurementRepo{}, &mockCoachClientRepo{}, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "", "client-1")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetLatestMeasurements_EmptyClientID(t *testing.T) {
	uc := NewGetLatestMeasurementsUseCase(&mockMeasurementRepo{}, &mockCoachClientRepo{}, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "coach-1", "")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetLatestMeasurements_EmptyResult(t *testing.T) {
	mRepo := &mockMeasurementRepo{
		findLatestByClientIDFunc: func(ctx context.Context, clientID string) ([]*entities.LatestMeasurement, error) {
			return nil, nil
		},
	}
	uc := NewGetLatestMeasurementsUseCase(mRepo, authorizedCoachClientRepo(), &mockAuditRepo{})

	results, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results == nil {
		t.Fatal("expected non-nil results slice")
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestGetLatestMeasurements_RepoError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	mRepo := &mockMeasurementRepo{
		findLatestByClientIDFunc: func(ctx context.Context, clientID string) ([]*entities.LatestMeasurement, error) {
			return nil, repoErr
		},
	}
	uc := NewGetLatestMeasurementsUseCase(mRepo, authorizedCoachClientRepo(), &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err == nil {
		t.Fatal("expected error from repository")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repository error to propagate, got %v", err)
	}
}

func TestGetLatestMeasurements_AuditEventLogged(t *testing.T) {
	var auditLogged bool
	auditRepo := &mockAuditRepo{
		logEventFunc: func(ctx context.Context, event *entities.AuditEvent) error {
			auditLogged = true
			if event.Action != "phi.access" {
				t.Errorf("expected action 'phi.access', got %q", event.Action)
			}
			if event.EntityType != "measurement" {
				t.Errorf("expected entity_type 'measurement', got %q", event.EntityType)
			}
			return nil
		},
	}
	mRepo := &mockMeasurementRepo{
		findLatestByClientIDFunc: func(ctx context.Context, clientID string) ([]*entities.LatestMeasurement, error) {
			return []*entities.LatestMeasurement{}, nil
		},
	}
	uc := NewGetLatestMeasurementsUseCase(mRepo, authorizedCoachClientRepo(), auditRepo)

	_, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !auditLogged {
		t.Error("expected audit event to be logged")
	}
}
