package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

type stubWearableRepo struct {
	findByClientResult []*entities.WearableSummary
	findByClientErr    error
}

func (s *stubWearableRepo) Upsert(_ context.Context, summary *entities.WearableSummary) (*entities.WearableSummary, error) {
	return summary, nil
}

func (s *stubWearableRepo) FindByClientID(_ context.Context, _ string, _ int) ([]*entities.WearableSummary, error) {
	return s.findByClientResult, s.findByClientErr
}

func TestGetWearableSummaries_Success_ReturnsSummaries(t *testing.T) {
	now := time.Now()
	wearableRepo := &stubWearableRepo{
		findByClientResult: []*entities.WearableSummary{
			{ID: "w1", ClientID: "client-1", Source: "whoop", SummaryDate: "2026-02-27", Metrics: map[string]interface{}{"hrv": 45.0}, SyncedAt: now},
		},
	}
	ccRepo := &stubCoachClientRepoForAuth{
		relationship: &entities.CoachClient{ID: "cc1", CoachID: "coach-1", ClientID: "client-1"},
	}
	auditRepo := &stubAuditRepoForMeasurements{}

	uc := NewGetWearableSummariesUseCase(wearableRepo, ccRepo, auditRepo)
	result, err := uc.Execute(context.Background(), "coach-1", "client-1", 7)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 summary, got %d", len(result))
	}
	if len(auditRepo.events) != 1 {
		t.Errorf("expected 1 audit event, got %d", len(auditRepo.events))
	}
}

func TestGetWearableSummaries_Unauthorized_ReturnsAuthorizationError(t *testing.T) {
	ccRepo := &stubCoachClientRepoForAuth{relationship: nil}
	uc := NewGetWearableSummariesUseCase(&stubWearableRepo{}, ccRepo, &stubAuditRepoForMeasurements{})

	_, err := uc.Execute(context.Background(), "coach-1", "client-1", 7)

	if !IsAuthorizationError(err) {
		t.Errorf("expected AuthorizationError, got %T: %v", err, err)
	}
}

func TestGetWearableSummaries_MissingCoachID_ReturnsValidationError(t *testing.T) {
	uc := NewGetWearableSummariesUseCase(&stubWearableRepo{}, &stubCoachClientRepoForAuth{}, &stubAuditRepoForMeasurements{})

	_, err := uc.Execute(context.Background(), "", "client-1", 7)

	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestGetWearableSummaries_MissingClientID_ReturnsValidationError(t *testing.T) {
	uc := NewGetWearableSummariesUseCase(&stubWearableRepo{}, &stubCoachClientRepoForAuth{}, &stubAuditRepoForMeasurements{})

	_, err := uc.Execute(context.Background(), "coach-1", "", 7)

	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestGetWearableSummaries_DefaultLimit_Applied(t *testing.T) {
	wearableRepo := &stubWearableRepo{
		findByClientResult: []*entities.WearableSummary{},
	}
	ccRepo := &stubCoachClientRepoForAuth{
		relationship: &entities.CoachClient{ID: "cc1", CoachID: "coach-1", ClientID: "client-1"},
	}

	uc := NewGetWearableSummariesUseCase(wearableRepo, ccRepo, &stubAuditRepoForMeasurements{})
	result, err := uc.Execute(context.Background(), "coach-1", "client-1", 0)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestGetWearableSummaries_RepoError_ReturnsError(t *testing.T) {
	ccRepo := &stubCoachClientRepoForAuth{
		relationship: &entities.CoachClient{ID: "cc1", CoachID: "coach-1", ClientID: "client-1"},
	}
	wearableRepo := &stubWearableRepo{
		findByClientErr: errors.New("db error"),
	}
	uc := NewGetWearableSummariesUseCase(wearableRepo, ccRepo, &stubAuditRepoForMeasurements{})

	_, err := uc.Execute(context.Background(), "coach-1", "client-1", 7)

	if err == nil {
		t.Error("expected error, got nil")
	}
}
