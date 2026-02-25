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
	measurements := []*entities.Measurement{
		{ID: "m1", ClientID: "client-1", Type: "weight", Value: 185.4, Unit: "lbs", MeasuredAt: now},
		{ID: "m2", ClientID: "client-1", Type: "body_fat_pct", Value: 22.3, Unit: "%", MeasuredAt: now},
	}

	measurementRepo := &mockMeasurementRepo{
		findLatestByClientIDFunc: func(ctx context.Context, clientID string) ([]*entities.Measurement, error) {
			if clientID != "client-1" {
				t.Errorf("expected clientID 'client-1', got '%s'", clientID)
			}
			return measurements, nil
		},
	}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetLatestMeasurementsUseCase(measurementRepo, ccRepo, auditRepo)
	out, err := uc.Execute(context.Background(), GetLatestMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Measurements) != 2 {
		t.Errorf("expected 2 measurements, got %d", len(out.Measurements))
	}
}

func TestGetLatestMeasurements_Unauthorized(t *testing.T) {
	measurementRepo := &mockMeasurementRepo{}
	ccRepo := &mockCoachClientRepo{
		findByCoachAndClient: func(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error) {
			return nil, nil
		},
	}
	auditRepo := &mockAuditRepo{}

	uc := NewGetLatestMeasurementsUseCase(measurementRepo, ccRepo, auditRepo)
	_, err := uc.Execute(context.Background(), GetLatestMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-2",
	})

	if err == nil {
		t.Fatal("expected authorization error")
	}
	if !IsAuthorizationError(err) {
		t.Errorf("expected AuthorizationError, got %T", err)
	}
}

func TestGetLatestMeasurements_EmptyCoachID(t *testing.T) {
	uc := NewGetLatestMeasurementsUseCase(&mockMeasurementRepo{}, &mockCoachClientRepo{}, &mockAuditRepo{})
	_, err := uc.Execute(context.Background(), GetLatestMeasurementsInput{
		ClientID: "client-1",
	})
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetLatestMeasurements_EmptyClientID(t *testing.T) {
	uc := NewGetLatestMeasurementsUseCase(&mockMeasurementRepo{}, &mockCoachClientRepo{}, &mockAuditRepo{})
	_, err := uc.Execute(context.Background(), GetLatestMeasurementsInput{
		CoachID: "coach-1",
	})
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetLatestMeasurements_EmptyData(t *testing.T) {
	measurementRepo := &mockMeasurementRepo{
		findLatestByClientIDFunc: func(ctx context.Context, clientID string) ([]*entities.Measurement, error) {
			return nil, nil
		},
	}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetLatestMeasurementsUseCase(measurementRepo, ccRepo, auditRepo)
	out, err := uc.Execute(context.Background(), GetLatestMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Measurements == nil {
		t.Error("expected non-nil empty measurements array")
	}
	if len(out.Measurements) != 0 {
		t.Errorf("expected 0 measurements, got %d", len(out.Measurements))
	}
}

func TestGetLatestMeasurements_RepoError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	measurementRepo := &mockMeasurementRepo{
		findLatestByClientIDFunc: func(ctx context.Context, clientID string) ([]*entities.Measurement, error) {
			return nil, repoErr
		},
	}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetLatestMeasurementsUseCase(measurementRepo, ccRepo, auditRepo)
	_, err := uc.Execute(context.Background(), GetLatestMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

	if !errors.Is(err, repoErr) {
		t.Errorf("expected repository error to propagate, got %v", err)
	}
}

func TestGetLatestMeasurements_AuditLogged(t *testing.T) {
	measurementRepo := &mockMeasurementRepo{}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetLatestMeasurementsUseCase(measurementRepo, ccRepo, auditRepo)
	_, err := uc.Execute(context.Background(), GetLatestMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(auditRepo.events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(auditRepo.events))
	}
	if auditRepo.events[0].Action != "phi.measurements_accessed" {
		t.Errorf("expected action 'phi.measurements_accessed', got '%s'", auditRepo.events[0].Action)
	}
}
