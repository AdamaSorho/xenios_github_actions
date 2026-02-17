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
	findLatestByClientIDFunc func(ctx context.Context, clientID string) ([]*entities.LatestMeasurement, error)
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

func (m *mockMeasurementRepo) FindLatestByClientID(ctx context.Context, clientID string) ([]*entities.LatestMeasurement, error) {
	if m.findLatestByClientIDFunc != nil {
		return m.findLatestByClientIDFunc(ctx, clientID)
	}
	return []*entities.LatestMeasurement{}, nil
}

func (m *mockMeasurementRepo) FindByType(ctx context.Context, clientID, measurementType string) ([]*entities.Measurement, error) {
	if m.findByTypeFunc != nil {
		return m.findByTypeFunc(ctx, clientID, measurementType)
	}
	return []*entities.Measurement{}, nil
}

type mockAuditRepo struct {
	logEventFunc func(ctx context.Context, event *entities.AuditEvent) error
	queryFunc    func(ctx context.Context, filter entities.AuditQueryFilter) ([]*entities.AuditEvent, int, error)
}

func (m *mockAuditRepo) LogEvent(ctx context.Context, event *entities.AuditEvent) error {
	if m.logEventFunc != nil {
		return m.logEventFunc(ctx, event)
	}
	return nil
}

func (m *mockAuditRepo) Query(ctx context.Context, filter entities.AuditQueryFilter) ([]*entities.AuditEvent, int, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, filter)
	}
	return []*entities.AuditEvent{}, 0, nil
}

func newMeasurementsUseCase(mRepo *mockMeasurementRepo, ccRepo *mockCoachClientRepo, aRepo *mockAuditRepo) *GetClientMeasurementsUseCase {
	return NewGetClientMeasurementsUseCase(mRepo, ccRepo, aRepo)
}

func authorizedCoachClientRepo() *mockCoachClientRepo {
	return &mockCoachClientRepo{
		findByCoachAndClient: func(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error) {
			return &entities.CoachClient{ID: "rel-1", CoachID: coachID, ClientID: clientID}, nil
		},
	}
}

func TestGetClientMeasurements_Success(t *testing.T) {
	now := time.Now()
	mRepo := &mockMeasurementRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			return []*entities.Measurement{
				{ID: "m1", ClientID: "client-1", MeasurementType: "weight", Value: 185.4, Unit: "lbs", MeasuredAt: now},
				{ID: "m2", ClientID: "client-1", MeasurementType: "body_fat_pct", Value: 22.3, Unit: "%", MeasuredAt: now},
			}, 2, nil
		},
	}
	uc := newMeasurementsUseCase(mRepo, authorizedCoachClientRepo(), &mockAuditRepo{})

	output, err := uc.Execute(context.Background(), GetMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Page:     1,
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(output.Measurements) != 2 {
		t.Errorf("expected 2 measurements, got %d", len(output.Measurements))
	}
	if output.Pagination.Total != 2 {
		t.Errorf("expected total 2, got %d", output.Pagination.Total)
	}
	if output.Pagination.Page != 1 {
		t.Errorf("expected page 1, got %d", output.Pagination.Page)
	}
}

func TestGetClientMeasurements_FilterByType(t *testing.T) {
	var capturedFilter entities.MeasurementFilter
	mRepo := &mockMeasurementRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			capturedFilter = filter
			return []*entities.Measurement{}, 0, nil
		},
	}
	uc := newMeasurementsUseCase(mRepo, authorizedCoachClientRepo(), &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), GetMeasurementsInput{
		CoachID:         "coach-1",
		ClientID:        "client-1",
		MeasurementType: "weight",
		Page:            1,
		Limit:           20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedFilter.MeasurementType != "weight" {
		t.Errorf("expected filter type 'weight', got %q", capturedFilter.MeasurementType)
	}
}

func TestGetClientMeasurements_FilterByDateRange(t *testing.T) {
	var capturedFilter entities.MeasurementFilter
	mRepo := &mockMeasurementRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			capturedFilter = filter
			return []*entities.Measurement{}, 0, nil
		},
	}
	uc := newMeasurementsUseCase(mRepo, authorizedCoachClientRepo(), &mockAuditRepo{})

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	_, err := uc.Execute(context.Background(), GetMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		From:     &from,
		To:       &to,
		Page:     1,
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedFilter.From == nil || !capturedFilter.From.Equal(from) {
		t.Errorf("expected from filter to be set")
	}
	if capturedFilter.To == nil || !capturedFilter.To.Equal(to) {
		t.Errorf("expected to filter to be set")
	}
}

func TestGetClientMeasurements_Unauthorized(t *testing.T) {
	ccRepo := &mockCoachClientRepo{
		findByCoachAndClient: func(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error) {
			return nil, nil // No relationship
		},
	}
	uc := newMeasurementsUseCase(&mockMeasurementRepo{}, ccRepo, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), GetMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Page:     1,
		Limit:    20,
	})
	if err == nil {
		t.Fatal("expected authorization error")
	}
	if !IsAuthorizationError(err) {
		t.Errorf("expected AuthorizationError, got %T: %v", err, err)
	}
}

func TestGetClientMeasurements_EmptyCoachID(t *testing.T) {
	uc := newMeasurementsUseCase(&mockMeasurementRepo{}, &mockCoachClientRepo{}, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), GetMeasurementsInput{
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

func TestGetClientMeasurements_EmptyClientID(t *testing.T) {
	uc := newMeasurementsUseCase(&mockMeasurementRepo{}, &mockCoachClientRepo{}, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), GetMeasurementsInput{
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

func TestGetClientMeasurements_DefaultPagination(t *testing.T) {
	var capturedFilter entities.MeasurementFilter
	mRepo := &mockMeasurementRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			capturedFilter = filter
			return []*entities.Measurement{}, 0, nil
		},
	}
	uc := newMeasurementsUseCase(mRepo, authorizedCoachClientRepo(), &mockAuditRepo{})

	output, err := uc.Execute(context.Background(), GetMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Page:     0,
		Limit:    0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedFilter.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", capturedFilter.Limit)
	}
	if output.Pagination.Page != 1 {
		t.Errorf("expected default page 1, got %d", output.Pagination.Page)
	}
}

func TestGetClientMeasurements_MaxLimit(t *testing.T) {
	var capturedFilter entities.MeasurementFilter
	mRepo := &mockMeasurementRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			capturedFilter = filter
			return []*entities.Measurement{}, 0, nil
		},
	}
	uc := newMeasurementsUseCase(mRepo, authorizedCoachClientRepo(), &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), GetMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Page:     1,
		Limit:    500,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedFilter.Limit != 100 {
		t.Errorf("expected max limit 100, got %d", capturedFilter.Limit)
	}
}

func TestGetClientMeasurements_Pagination(t *testing.T) {
	var capturedFilter entities.MeasurementFilter
	mRepo := &mockMeasurementRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			capturedFilter = filter
			return []*entities.Measurement{}, 100, nil
		},
	}
	uc := newMeasurementsUseCase(mRepo, authorizedCoachClientRepo(), &mockAuditRepo{})

	output, err := uc.Execute(context.Background(), GetMeasurementsInput{
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
	if output.Pagination.Page != 3 {
		t.Errorf("expected page 3, got %d", output.Pagination.Page)
	}
	if output.Pagination.Total != 100 {
		t.Errorf("expected total 100, got %d", output.Pagination.Total)
	}
}

func TestGetClientMeasurements_EmptyResult(t *testing.T) {
	mRepo := &mockMeasurementRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			return nil, 0, nil
		},
	}
	uc := newMeasurementsUseCase(mRepo, authorizedCoachClientRepo(), &mockAuditRepo{})

	output, err := uc.Execute(context.Background(), GetMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Page:     1,
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Measurements == nil {
		t.Fatal("expected non-nil measurements slice")
	}
	if len(output.Measurements) != 0 {
		t.Errorf("expected 0 measurements, got %d", len(output.Measurements))
	}
}

func TestGetClientMeasurements_RepoError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	mRepo := &mockMeasurementRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			return nil, 0, repoErr
		},
	}
	uc := newMeasurementsUseCase(mRepo, authorizedCoachClientRepo(), &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), GetMeasurementsInput{
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

func TestGetClientMeasurements_AuditEventLogged(t *testing.T) {
	var auditLogged bool
	auditRepo := &mockAuditRepo{
		logEventFunc: func(ctx context.Context, event *entities.AuditEvent) error {
			auditLogged = true
			if event.Action != "phi.access" {
				t.Errorf("expected action 'phi.access', got %q", event.Action)
			}
			if event.ActorID != "coach-1" {
				t.Errorf("expected actor_id 'coach-1', got %q", event.ActorID)
			}
			if event.EntityType != "measurement" {
				t.Errorf("expected entity_type 'measurement', got %q", event.EntityType)
			}
			if event.EntityID != "client-1" {
				t.Errorf("expected entity_id 'client-1', got %q", event.EntityID)
			}
			return nil
		},
	}
	mRepo := &mockMeasurementRepo{
		findByClientIDFunc: func(ctx context.Context, filter entities.MeasurementFilter) ([]*entities.Measurement, int, error) {
			return []*entities.Measurement{}, 0, nil
		},
	}
	uc := newMeasurementsUseCase(mRepo, authorizedCoachClientRepo(), auditRepo)

	_, err := uc.Execute(context.Background(), GetMeasurementsInput{
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

func TestIsAuthorizationError_True(t *testing.T) {
	err := &AuthorizationError{Message: "test"}
	if !IsAuthorizationError(err) {
		t.Error("expected true for AuthorizationError")
	}
}

func TestIsAuthorizationError_False(t *testing.T) {
	err := errors.New("not an authorization error")
	if IsAuthorizationError(err) {
		t.Error("expected false for non-AuthorizationError")
	}
}
