package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

type mockMeasurementRepo struct {
	findByClientIDFunc      func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error)
	findLatestByClientIDFunc func(ctx context.Context, clientID string) ([]*entities.Measurement, error)
	findByTypeFunc          func(ctx context.Context, clientID, measurementType string) ([]*entities.Measurement, error)
	createFunc              func(ctx context.Context, m *entities.Measurement) (*entities.Measurement, error)
}

func (m *mockMeasurementRepo) Create(ctx context.Context, measurement *entities.Measurement) (*entities.Measurement, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, measurement)
	}
	return measurement, nil
}

func (m *mockMeasurementRepo) FindByClientID(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
	if m.findByClientIDFunc != nil {
		return m.findByClientIDFunc(ctx, filter)
	}
	return []*entities.Measurement{}, 0, nil
}

func (m *mockMeasurementRepo) FindLatestByClientID(ctx context.Context, clientID string) ([]*entities.Measurement, error) {
	if m.findLatestByClientIDFunc != nil {
		return m.findLatestByClientIDFunc(ctx, clientID)
	}
	return []*entities.Measurement{}, nil
}

func (m *mockMeasurementRepo) FindByType(ctx context.Context, clientID, measurementType string) ([]*entities.Measurement, error) {
	if m.findByTypeFunc != nil {
		return m.findByTypeFunc(ctx, clientID, measurementType)
	}
	return []*entities.Measurement{}, nil
}

type mockAuditRepo struct {
	logEventFunc func(ctx context.Context, event *entities.AuditEvent) error
	events       []*entities.AuditEvent
}

func (m *mockAuditRepo) LogEvent(ctx context.Context, event *entities.AuditEvent) error {
	m.events = append(m.events, event)
	if m.logEventFunc != nil {
		return m.logEventFunc(ctx, event)
	}
	return nil
}

func (m *mockAuditRepo) Query(ctx context.Context, filter entities.AuditQueryFilter) ([]*entities.AuditEvent, int, error) {
	return []*entities.AuditEvent{}, 0, nil
}

func TestGetClientMeasurements_Success(t *testing.T) {
	now := time.Now()
	measurements := []*entities.Measurement{
		{ID: "m1", ClientID: "client-1", Type: "weight", Value: 185.4, Unit: "lbs", MeasuredAt: now},
		{ID: "m2", ClientID: "client-1", Type: "body_fat_pct", Value: 22.3, Unit: "%", MeasuredAt: now},
	}

	measurementRepo := &mockMeasurementRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			return measurements, 2, nil
		},
	}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetClientMeasurementsUseCase(measurementRepo, ccRepo, auditRepo)
	out, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

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

func TestGetClientMeasurements_WithTypeFilter(t *testing.T) {
	var capturedFilter entities.MeasurementFilter
	measurementRepo := &mockMeasurementRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			capturedFilter = filter
			return []*entities.Measurement{}, 0, nil
		},
	}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetClientMeasurementsUseCase(measurementRepo, ccRepo, auditRepo)
	_, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Type:     "weight",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedFilter.Type != "weight" {
		t.Errorf("expected filter type 'weight', got '%s'", capturedFilter.Type)
	}
}

func TestGetClientMeasurements_WithDateRange(t *testing.T) {
	var capturedFilter entities.MeasurementFilter
	measurementRepo := &mockMeasurementRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			capturedFilter = filter
			return []*entities.Measurement{}, 0, nil
		},
	}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	uc := NewGetClientMeasurementsUseCase(measurementRepo, ccRepo, auditRepo)
	_, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		From:     &from,
		To:       &to,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedFilter.From == nil || !capturedFilter.From.Equal(from) {
		t.Error("expected from date to be set")
	}
	if capturedFilter.To == nil || !capturedFilter.To.Equal(to) {
		t.Error("expected to date to be set")
	}
}

func TestGetClientMeasurements_Pagination(t *testing.T) {
	var capturedFilter entities.MeasurementFilter
	measurementRepo := &mockMeasurementRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			capturedFilter = filter
			return []*entities.Measurement{}, 100, nil
		},
	}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetClientMeasurementsUseCase(measurementRepo, ccRepo, auditRepo)
	out, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Page:     3,
		Limit:    20,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedFilter.Offset != 40 {
		t.Errorf("expected offset 40 for page 3 limit 20, got %d", capturedFilter.Offset)
	}
	if out.Pagination.Page != 3 {
		t.Errorf("expected page 3, got %d", out.Pagination.Page)
	}
	if out.Pagination.Total != 100 {
		t.Errorf("expected total 100, got %d", out.Pagination.Total)
	}
}

func TestGetClientMeasurements_DefaultPagination(t *testing.T) {
	var capturedFilter entities.MeasurementFilter
	measurementRepo := &mockMeasurementRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			capturedFilter = filter
			return []*entities.Measurement{}, 0, nil
		},
	}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetClientMeasurementsUseCase(measurementRepo, ccRepo, auditRepo)
	out, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedFilter.Limit != defaultMeasurementLimit {
		t.Errorf("expected default limit %d, got %d", defaultMeasurementLimit, capturedFilter.Limit)
	}
	if out.Pagination.Page != 1 {
		t.Errorf("expected default page 1, got %d", out.Pagination.Page)
	}
}

func TestGetClientMeasurements_MaxLimit(t *testing.T) {
	var capturedFilter entities.MeasurementFilter
	measurementRepo := &mockMeasurementRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			capturedFilter = filter
			return []*entities.Measurement{}, 0, nil
		},
	}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetClientMeasurementsUseCase(measurementRepo, ccRepo, auditRepo)
	_, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Limit:    500,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedFilter.Limit != maxMeasurementLimit {
		t.Errorf("expected max limit %d, got %d", maxMeasurementLimit, capturedFilter.Limit)
	}
}

func TestGetClientMeasurements_Unauthorized(t *testing.T) {
	measurementRepo := &mockMeasurementRepo{}
	ccRepo := &mockCoachClientRepo{
		findByCoachAndClient: func(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error) {
			return nil, nil
		},
	}
	auditRepo := &mockAuditRepo{}

	uc := NewGetClientMeasurementsUseCase(measurementRepo, ccRepo, auditRepo)
	_, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
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

func TestGetClientMeasurements_EmptyCoachID(t *testing.T) {
	uc := NewGetClientMeasurementsUseCase(&mockMeasurementRepo{}, &mockCoachClientRepo{}, &mockAuditRepo{})
	_, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		ClientID: "client-1",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetClientMeasurements_EmptyClientID(t *testing.T) {
	uc := NewGetClientMeasurementsUseCase(&mockMeasurementRepo{}, &mockCoachClientRepo{}, &mockAuditRepo{})
	_, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID: "coach-1",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGetClientMeasurements_RepoError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	measurementRepo := &mockMeasurementRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			return nil, 0, repoErr
		},
	}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetClientMeasurementsUseCase(measurementRepo, ccRepo, auditRepo)
	_, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

	if err == nil {
		t.Fatal("expected error from repository")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repository error to propagate, got %v", err)
	}
}

func TestGetClientMeasurements_EmptyData(t *testing.T) {
	measurementRepo := &mockMeasurementRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			return nil, 0, nil
		},
	}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetClientMeasurementsUseCase(measurementRepo, ccRepo, auditRepo)
	out, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
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

func TestGetClientMeasurements_AuditLogged(t *testing.T) {
	measurementRepo := &mockMeasurementRepo{}
	ccRepo := &mockCoachClientRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewGetClientMeasurementsUseCase(measurementRepo, ccRepo, auditRepo)
	_, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
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
	if auditRepo.events[0].ActorID != "coach-1" {
		t.Errorf("expected actor_id 'coach-1', got '%s'", auditRepo.events[0].ActorID)
	}
}
