package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestGetClientMeasurements_Success_ReturnsMeasurements(t *testing.T) {
	now := time.Now()
	measRepo := &mockMeasurementRepo{
		findByClientFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			return []*entities.Measurement{
				{ID: "m1", ClientID: "client-1", MeasurementType: "weight", Value: 185.4, Unit: "lbs", MeasuredAt: now},
				{ID: "m2", ClientID: "client-1", MeasurementType: "body_fat_pct", Value: 22.3, Unit: "%", MeasuredAt: now},
			}, 2, nil
		},
	}
	auditRepo := &mockAuditRepo{}
	uc := NewGetClientMeasurementsUseCase(measRepo, authorizedCCRepo(), auditRepo)

	filter := entities.MeasurementFilter{ClientID: "client-1", Limit: 20, Offset: 0}
	out, err := uc.Execute(context.Background(), "coach-1", filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Measurements) != 2 {
		t.Errorf("expected 2 measurements, got %d", len(out.Measurements))
	}
	if out.Pagination.Total != 2 {
		t.Errorf("expected total 2, got %d", out.Pagination.Total)
	}
	if out.Pagination.Page != 1 {
		t.Errorf("expected page 1, got %d", out.Pagination.Page)
	}
}

func TestGetClientMeasurements_FilterByType_ReturnsFiltered(t *testing.T) {
	var capturedFilter entities.MeasurementFilter
	measRepo := &mockMeasurementRepo{
		findByClientFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			capturedFilter = filter
			return []*entities.Measurement{
				{ID: "m1", ClientID: "client-1", MeasurementType: "weight", Value: 185.4, Unit: "lbs"},
			}, 1, nil
		},
	}
	auditRepo := &mockAuditRepo{}
	uc := NewGetClientMeasurementsUseCase(measRepo, authorizedCCRepo(), auditRepo)

	filter := entities.MeasurementFilter{ClientID: "client-1", MeasurementType: "weight", Limit: 20}
	out, err := uc.Execute(context.Background(), "coach-1", filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedFilter.MeasurementType != "weight" {
		t.Errorf("expected filter type weight, got %s", capturedFilter.MeasurementType)
	}
	if len(out.Measurements) != 1 {
		t.Errorf("expected 1 measurement, got %d", len(out.Measurements))
	}
}

func TestGetClientMeasurements_DateRange_PassesFilter(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	var capturedFilter entities.MeasurementFilter
	measRepo := &mockMeasurementRepo{
		findByClientFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			capturedFilter = filter
			return []*entities.Measurement{}, 0, nil
		},
	}
	auditRepo := &mockAuditRepo{}
	uc := NewGetClientMeasurementsUseCase(measRepo, authorizedCCRepo(), auditRepo)

	filter := entities.MeasurementFilter{ClientID: "client-1", From: &from, To: &to, Limit: 20}
	_, err := uc.Execute(context.Background(), "coach-1", filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedFilter.From == nil || !capturedFilter.From.Equal(from) {
		t.Error("expected From filter to be passed through")
	}
	if capturedFilter.To == nil || !capturedFilter.To.Equal(to) {
		t.Error("expected To filter to be passed through")
	}
}

func TestGetClientMeasurements_Unauthorized_Returns403(t *testing.T) {
	measRepo := &mockMeasurementRepo{}
	auditRepo := &mockAuditRepo{}
	uc := NewGetClientMeasurementsUseCase(measRepo, unauthorizedCCRepo(), auditRepo)

	filter := entities.MeasurementFilter{ClientID: "client-1", Limit: 20}
	_, err := uc.Execute(context.Background(), "coach-1", filter)
	if err == nil {
		t.Fatal("expected error for unauthorized access")
	}
	if !IsAuthenticationError(err) {
		t.Errorf("expected AuthenticationError, got %T", err)
	}
}

func TestGetClientMeasurements_EmptyCoachID_ReturnsValidationError(t *testing.T) {
	measRepo := &mockMeasurementRepo{}
	auditRepo := &mockAuditRepo{}
	uc := NewGetClientMeasurementsUseCase(measRepo, authorizedCCRepo(), auditRepo)

	filter := entities.MeasurementFilter{ClientID: "client-1"}
	_, err := uc.Execute(context.Background(), "", filter)
	if err == nil {
		t.Fatal("expected error for empty coach_id")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetClientMeasurements_EmptyClientID_ReturnsValidationError(t *testing.T) {
	measRepo := &mockMeasurementRepo{}
	auditRepo := &mockAuditRepo{}
	uc := NewGetClientMeasurementsUseCase(measRepo, authorizedCCRepo(), auditRepo)

	filter := entities.MeasurementFilter{ClientID: ""}
	_, err := uc.Execute(context.Background(), "coach-1", filter)
	if err == nil {
		t.Fatal("expected error for empty client_id")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetClientMeasurements_DefaultLimit_Applied(t *testing.T) {
	var capturedFilter entities.MeasurementFilter
	measRepo := &mockMeasurementRepo{
		findByClientFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			capturedFilter = filter
			return []*entities.Measurement{}, 0, nil
		},
	}
	auditRepo := &mockAuditRepo{}
	uc := NewGetClientMeasurementsUseCase(measRepo, authorizedCCRepo(), auditRepo)

	filter := entities.MeasurementFilter{ClientID: "client-1", Limit: 0}
	_, err := uc.Execute(context.Background(), "coach-1", filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedFilter.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", capturedFilter.Limit)
	}
}

func TestGetClientMeasurements_MaxLimit_Capped(t *testing.T) {
	var capturedFilter entities.MeasurementFilter
	measRepo := &mockMeasurementRepo{
		findByClientFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			capturedFilter = filter
			return []*entities.Measurement{}, 0, nil
		},
	}
	auditRepo := &mockAuditRepo{}
	uc := NewGetClientMeasurementsUseCase(measRepo, authorizedCCRepo(), auditRepo)

	filter := entities.MeasurementFilter{ClientID: "client-1", Limit: 500}
	_, err := uc.Execute(context.Background(), "coach-1", filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedFilter.Limit != 100 {
		t.Errorf("expected max limit 100, got %d", capturedFilter.Limit)
	}
}

func TestGetClientMeasurements_Pagination_Page2(t *testing.T) {
	measRepo := &mockMeasurementRepo{
		findByClientFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			return []*entities.Measurement{
				{ID: "m21", ClientID: "client-1"},
			}, 100, nil
		},
	}
	auditRepo := &mockAuditRepo{}
	uc := NewGetClientMeasurementsUseCase(measRepo, authorizedCCRepo(), auditRepo)

	filter := entities.MeasurementFilter{ClientID: "client-1", Limit: 20, Offset: 20}
	out, err := uc.Execute(context.Background(), "coach-1", filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Pagination.Page != 2 {
		t.Errorf("expected page 2, got %d", out.Pagination.Page)
	}
	if out.Pagination.Total != 100 {
		t.Errorf("expected total 100, got %d", out.Pagination.Total)
	}
}

func TestGetClientMeasurements_EmptyResult_ReturnsEmptyArray(t *testing.T) {
	measRepo := &mockMeasurementRepo{
		findByClientFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			return nil, 0, nil
		},
	}
	auditRepo := &mockAuditRepo{}
	uc := NewGetClientMeasurementsUseCase(measRepo, authorizedCCRepo(), auditRepo)

	filter := entities.MeasurementFilter{ClientID: "client-1", Limit: 20}
	out, err := uc.Execute(context.Background(), "coach-1", filter)
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

func TestGetClientMeasurements_RepoError_Propagates(t *testing.T) {
	repoErr := errors.New("database unavailable")
	measRepo := &mockMeasurementRepo{
		findByClientFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			return nil, 0, repoErr
		},
	}
	auditRepo := &mockAuditRepo{}
	uc := NewGetClientMeasurementsUseCase(measRepo, authorizedCCRepo(), auditRepo)

	filter := entities.MeasurementFilter{ClientID: "client-1", Limit: 20}
	_, err := uc.Execute(context.Background(), "coach-1", filter)
	if err == nil {
		t.Fatal("expected error from repository")
	}
}

func TestGetClientMeasurements_AuditEvent_Logged(t *testing.T) {
	measRepo := &mockMeasurementRepo{
		findByClientFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			return []*entities.Measurement{}, 0, nil
		},
	}
	auditRepo := &mockAuditRepo{}
	uc := NewGetClientMeasurementsUseCase(measRepo, authorizedCCRepo(), auditRepo)

	filter := entities.MeasurementFilter{ClientID: "client-1", Limit: 20}
	_, err := uc.Execute(context.Background(), "coach-1", filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(auditRepo.events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(auditRepo.events))
	}
	if auditRepo.events[0].Action != "phi.access" {
		t.Errorf("expected phi.access action, got %s", auditRepo.events[0].Action)
	}
	if auditRepo.events[0].ActorID != "coach-1" {
		t.Errorf("expected actor coach-1, got %s", auditRepo.events[0].ActorID)
	}
}
