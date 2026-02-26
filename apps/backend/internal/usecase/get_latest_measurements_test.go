package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestGetLatestMeasurements_Success_ReturnsLatest(t *testing.T) {
	now := time.Now()
	measRepo := &mockMeasurementRepo{
		findLatestFunc: func(ctx context.Context, clientID string) ([]*entities.Measurement, error) {
			return []*entities.Measurement{
				{ID: "m1", ClientID: clientID, MeasurementType: "weight", Value: 185.4, Unit: "lbs", MeasuredAt: now},
				{ID: "m2", ClientID: clientID, MeasurementType: "body_fat_pct", Value: 22.3, Unit: "%", MeasuredAt: now},
			}, nil
		},
	}
	auditRepo := &mockAuditRepo{}
	uc := NewGetLatestMeasurementsUseCase(measRepo, authorizedCCRepo(), auditRepo)

	out, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Measurements) != 2 {
		t.Errorf("expected 2 measurements, got %d", len(out.Measurements))
	}
}

func TestGetLatestMeasurements_Unauthorized_ReturnsForbidden(t *testing.T) {
	measRepo := &mockMeasurementRepo{}
	auditRepo := &mockAuditRepo{}
	uc := NewGetLatestMeasurementsUseCase(measRepo, unauthorizedCCRepo(), auditRepo)

	_, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err == nil {
		t.Fatal("expected error for unauthorized access")
	}
	if !IsAuthenticationError(err) {
		t.Errorf("expected AuthenticationError, got %T", err)
	}
}

func TestGetLatestMeasurements_EmptyCoachID_ReturnsValidationError(t *testing.T) {
	uc := NewGetLatestMeasurementsUseCase(&mockMeasurementRepo{}, authorizedCCRepo(), &mockAuditRepo{})
	_, err := uc.Execute(context.Background(), "", "client-1")
	if err == nil {
		t.Fatal("expected error for empty coach_id")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetLatestMeasurements_EmptyClientID_ReturnsValidationError(t *testing.T) {
	uc := NewGetLatestMeasurementsUseCase(&mockMeasurementRepo{}, authorizedCCRepo(), &mockAuditRepo{})
	_, err := uc.Execute(context.Background(), "coach-1", "")
	if err == nil {
		t.Fatal("expected error for empty client_id")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetLatestMeasurements_EmptyData_ReturnsEmptyArray(t *testing.T) {
	measRepo := &mockMeasurementRepo{
		findLatestFunc: func(ctx context.Context, clientID string) ([]*entities.Measurement, error) {
			return nil, nil
		},
	}
	uc := NewGetLatestMeasurementsUseCase(measRepo, authorizedCCRepo(), &mockAuditRepo{})
	out, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Measurements == nil {
		t.Error("expected non-nil measurements array")
	}
	if len(out.Measurements) != 0 {
		t.Errorf("expected 0 measurements, got %d", len(out.Measurements))
	}
}

func TestGetLatestMeasurements_RepoError_Propagates(t *testing.T) {
	repoErr := errors.New("database unavailable")
	measRepo := &mockMeasurementRepo{
		findLatestFunc: func(ctx context.Context, clientID string) ([]*entities.Measurement, error) {
			return nil, repoErr
		},
	}
	uc := NewGetLatestMeasurementsUseCase(measRepo, authorizedCCRepo(), &mockAuditRepo{})
	_, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err == nil {
		t.Fatal("expected error from repository")
	}
}

func TestGetLatestMeasurements_AuditEvent_Logged(t *testing.T) {
	auditRepo := &mockAuditRepo{}
	uc := NewGetLatestMeasurementsUseCase(&mockMeasurementRepo{}, authorizedCCRepo(), auditRepo)
	_, err := uc.Execute(context.Background(), "coach-1", "client-1")
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
