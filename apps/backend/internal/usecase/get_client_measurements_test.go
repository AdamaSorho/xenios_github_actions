package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// --- Stub repositories for measurement tests ---

type stubMeasurementRepo struct {
	findByClientResult *entities.MeasurementPage
	findByClientErr    error
	findLatestResult   []*entities.Measurement
	findLatestErr      error
	findByTypeResult   []*entities.Measurement
	findByTypeErr      error
}

func (s *stubMeasurementRepo) Create(_ context.Context, m *entities.Measurement) (*entities.Measurement, error) {
	return m, nil
}

func (s *stubMeasurementRepo) FindByClientID(_ context.Context, _ entities.MeasurementFilter) (*entities.MeasurementPage, error) {
	return s.findByClientResult, s.findByClientErr
}

func (s *stubMeasurementRepo) FindLatestByClientID(_ context.Context, _ string) ([]*entities.Measurement, error) {
	return s.findLatestResult, s.findLatestErr
}

func (s *stubMeasurementRepo) FindByType(_ context.Context, _, _ string) ([]*entities.Measurement, error) {
	return s.findByTypeResult, s.findByTypeErr
}

type stubCoachClientRepoForAuth struct {
	relationship *entities.CoachClient
	err          error
}

func (s *stubCoachClientRepoForAuth) Create(_ context.Context, _, _ string) (*entities.CoachClient, error) {
	return nil, nil
}

func (s *stubCoachClientRepoForAuth) ListByCoachID(_ context.Context, _ string, _, _ int) ([]*entities.CoachClient, error) {
	return nil, nil
}

func (s *stubCoachClientRepoForAuth) FindByCoachAndClient(_ context.Context, _, _ string) (*entities.CoachClient, error) {
	return s.relationship, s.err
}

type stubAuditRepoForMeasurements struct {
	events []*entities.AuditEvent
}

func (s *stubAuditRepoForMeasurements) LogEvent(_ context.Context, event *entities.AuditEvent) error {
	s.events = append(s.events, event)
	return nil
}

func (s *stubAuditRepoForMeasurements) Query(_ context.Context, _ entities.AuditQueryFilter) ([]*entities.AuditEvent, int, error) {
	return s.events, len(s.events), nil
}

// --- GetClientMeasurements tests ---

func TestGetClientMeasurements_Success_ReturnsMeasurements(t *testing.T) {
	now := time.Now()
	measurementRepo := &stubMeasurementRepo{
		findByClientResult: &entities.MeasurementPage{
			Measurements: []*entities.Measurement{
				{ID: "m1", ClientID: "client-1", Type: "weight", Value: 185.4, Unit: "lbs", MeasuredAt: now},
			},
			Page:  1,
			Limit: 20,
			Total: 1,
		},
	}
	ccRepo := &stubCoachClientRepoForAuth{
		relationship: &entities.CoachClient{ID: "cc1", CoachID: "coach-1", ClientID: "client-1"},
	}
	auditRepo := &stubAuditRepoForMeasurements{}

	uc := NewGetClientMeasurementsUseCase(measurementRepo, ccRepo, auditRepo)
	result, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Measurements) != 1 {
		t.Errorf("expected 1 measurement, got %d", len(result.Measurements))
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
	if len(auditRepo.events) != 1 {
		t.Errorf("expected 1 audit event, got %d", len(auditRepo.events))
	}
	if auditRepo.events[0].Action != "phi.access" {
		t.Errorf("expected phi.access action, got %s", auditRepo.events[0].Action)
	}
}

func TestGetClientMeasurements_MissingCoachID_ReturnsValidationError(t *testing.T) {
	uc := NewGetClientMeasurementsUseCase(
		&stubMeasurementRepo{},
		&stubCoachClientRepoForAuth{},
		&stubAuditRepoForMeasurements{},
	)

	_, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "",
		ClientID: "client-1",
	})

	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestGetClientMeasurements_MissingClientID_ReturnsValidationError(t *testing.T) {
	uc := NewGetClientMeasurementsUseCase(
		&stubMeasurementRepo{},
		&stubCoachClientRepoForAuth{},
		&stubAuditRepoForMeasurements{},
	)

	_, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "",
	})

	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestGetClientMeasurements_Unauthorized_ReturnsAuthorizationError(t *testing.T) {
	ccRepo := &stubCoachClientRepoForAuth{relationship: nil}
	uc := NewGetClientMeasurementsUseCase(
		&stubMeasurementRepo{},
		ccRepo,
		&stubAuditRepoForMeasurements{},
	)

	_, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

	if !IsAuthorizationError(err) {
		t.Errorf("expected AuthorizationError, got %T: %v", err, err)
	}
}

func TestGetClientMeasurements_RepoError_ReturnsError(t *testing.T) {
	ccRepo := &stubCoachClientRepoForAuth{
		relationship: &entities.CoachClient{ID: "cc1", CoachID: "coach-1", ClientID: "client-1"},
	}
	measurementRepo := &stubMeasurementRepo{
		findByClientErr: errors.New("db error"),
	}
	uc := NewGetClientMeasurementsUseCase(measurementRepo, ccRepo, &stubAuditRepoForMeasurements{})

	_, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGetClientMeasurements_DefaultPagination_Applied(t *testing.T) {
	ccRepo := &stubCoachClientRepoForAuth{
		relationship: &entities.CoachClient{ID: "cc1", CoachID: "coach-1", ClientID: "client-1"},
	}
	measurementRepo := &stubMeasurementRepo{
		findByClientResult: &entities.MeasurementPage{
			Measurements: []*entities.Measurement{},
			Page:         1,
			Limit:        20,
			Total:        0,
		},
	}
	uc := NewGetClientMeasurementsUseCase(measurementRepo, ccRepo, &stubAuditRepoForMeasurements{})

	result, err := uc.Execute(context.Background(), GetClientMeasurementsInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Page)
	}
	if result.Limit != 20 {
		t.Errorf("expected limit 20, got %d", result.Limit)
	}
}

// --- GetLatestMeasurements tests ---

func TestGetLatestMeasurements_Success_ReturnsLatest(t *testing.T) {
	now := time.Now()
	measurementRepo := &stubMeasurementRepo{
		findLatestResult: []*entities.Measurement{
			{ID: "m1", ClientID: "client-1", Type: "weight", Value: 185.4, Unit: "lbs", MeasuredAt: now},
			{ID: "m2", ClientID: "client-1", Type: "body_fat_pct", Value: 22.3, Unit: "%", MeasuredAt: now},
		},
	}
	ccRepo := &stubCoachClientRepoForAuth{
		relationship: &entities.CoachClient{ID: "cc1", CoachID: "coach-1", ClientID: "client-1"},
	}
	auditRepo := &stubAuditRepoForMeasurements{}

	uc := NewGetLatestMeasurementsUseCase(measurementRepo, ccRepo, auditRepo)
	result, err := uc.Execute(context.Background(), "coach-1", "client-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 measurements, got %d", len(result))
	}
	if len(auditRepo.events) != 1 {
		t.Errorf("expected 1 audit event, got %d", len(auditRepo.events))
	}
}

func TestGetLatestMeasurements_Unauthorized_ReturnsAuthorizationError(t *testing.T) {
	ccRepo := &stubCoachClientRepoForAuth{relationship: nil}
	uc := NewGetLatestMeasurementsUseCase(&stubMeasurementRepo{}, ccRepo, &stubAuditRepoForMeasurements{})

	_, err := uc.Execute(context.Background(), "coach-1", "client-1")

	if !IsAuthorizationError(err) {
		t.Errorf("expected AuthorizationError, got %T: %v", err, err)
	}
}

func TestGetLatestMeasurements_MissingCoachID_ReturnsValidationError(t *testing.T) {
	uc := NewGetLatestMeasurementsUseCase(&stubMeasurementRepo{}, &stubCoachClientRepoForAuth{}, &stubAuditRepoForMeasurements{})

	_, err := uc.Execute(context.Background(), "", "client-1")

	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestGetLatestMeasurements_MissingClientID_ReturnsValidationError(t *testing.T) {
	uc := NewGetLatestMeasurementsUseCase(&stubMeasurementRepo{}, &stubCoachClientRepoForAuth{}, &stubAuditRepoForMeasurements{})

	_, err := uc.Execute(context.Background(), "coach-1", "")

	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestGetLatestMeasurements_RepoError_ReturnsError(t *testing.T) {
	ccRepo := &stubCoachClientRepoForAuth{
		relationship: &entities.CoachClient{ID: "cc1", CoachID: "coach-1", ClientID: "client-1"},
	}
	measurementRepo := &stubMeasurementRepo{
		findLatestErr: errors.New("db error"),
	}
	uc := NewGetLatestMeasurementsUseCase(measurementRepo, ccRepo, &stubAuditRepoForMeasurements{})

	_, err := uc.Execute(context.Background(), "coach-1", "client-1")

	if err == nil {
		t.Error("expected error, got nil")
	}
}
