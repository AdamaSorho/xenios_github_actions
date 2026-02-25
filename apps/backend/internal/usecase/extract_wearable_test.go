package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// ── Mocks ───────────────────────────────────────────────────────────────────

type mockMeasurementRepo struct {
	upsertBatchFunc func(ctx context.Context, measurements []entities.Measurement) (int, error)
	averageFunc     func(ctx context.Context, clientID string, source entities.WearableSource, mt entities.MeasurementType, since time.Time) (*float64, error)
}

func (m *mockMeasurementRepo) UpsertBatch(ctx context.Context, measurements []entities.Measurement) (int, error) {
	if m.upsertBatchFunc != nil {
		return m.upsertBatchFunc(ctx, measurements)
	}
	return len(measurements), nil
}

func (m *mockMeasurementRepo) Average(ctx context.Context, clientID string, source entities.WearableSource, mt entities.MeasurementType, since time.Time) (*float64, error) {
	if m.averageFunc != nil {
		return m.averageFunc(ctx, clientID, source, mt, since)
	}
	return nil, nil
}

var _ repository.MeasurementRepository = &mockMeasurementRepo{}

type mockWearableSummaryRepo struct {
	upsertFunc func(ctx context.Context, clientID string, source entities.WearableSource, metrics json.RawMessage) error
	lastMetrics json.RawMessage
}

func (m *mockWearableSummaryRepo) Upsert(ctx context.Context, clientID string, source entities.WearableSource, metrics json.RawMessage) error {
	m.lastMetrics = metrics
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, clientID, source, metrics)
	}
	return nil
}

var _ repository.WearableSummaryRepository = &mockWearableSummaryRepo{}

type mockAuditRepo struct {
	events []*entities.AuditEvent
}

func (m *mockAuditRepo) LogEvent(_ context.Context, event *entities.AuditEvent) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockAuditRepo) Query(_ context.Context, _ entities.AuditQueryFilter) ([]*entities.AuditEvent, int, error) {
	return m.events, len(m.events), nil
}

var _ repository.AuditRepository = &mockAuditRepo{}

// ── Test Helpers ────────────────────────────────────────────────────────────

func sampleMeasurements(clientID string) []entities.Measurement {
	return []entities.Measurement{
		{ClientID: clientID, Source: entities.WearableSourceWhoop, MeasurementType: entities.MeasurementTypeHRV, Value: 45.2, MeasuredAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ClientID: clientID, Source: entities.WearableSourceWhoop, MeasurementType: entities.MeasurementTypeRecovery, Value: 72, MeasuredAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ClientID: clientID, Source: entities.WearableSourceWhoop, MeasurementType: entities.MeasurementTypeSleepQuality, Value: 85, MeasuredAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
}

// ── Tests ───────────────────────────────────────────────────────────────────

func TestExtractWearableUseCase_Execute_HappyPath_ReturnsResult(t *testing.T) {
	measurementRepo := &mockMeasurementRepo{}
	summaryRepo := &mockWearableSummaryRepo{}
	auditRepo := &mockAuditRepo{}

	uc := NewExtractWearableUseCase(measurementRepo, summaryRepo, auditRepo)

	clientID := "client-123"
	measurements := sampleMeasurements(clientID)

	result, err := uc.Execute(context.Background(), clientID, measurements, entities.WearableSourceWhoop)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Source != entities.WearableSourceWhoop {
		t.Errorf("expected source %q, got %q", entities.WearableSourceWhoop, result.Source)
	}
	if result.TotalParsed != 3 {
		t.Errorf("expected total_parsed 3, got %d", result.TotalParsed)
	}
	if result.TotalInserted != 3 {
		t.Errorf("expected total_inserted 3, got %d", result.TotalInserted)
	}
	if result.DuplicatesSkipped != 0 {
		t.Errorf("expected duplicates_skipped 0, got %d", result.DuplicatesSkipped)
	}
}

func TestExtractWearableUseCase_Execute_EmptyClientID_ReturnsValidationError(t *testing.T) {
	uc := NewExtractWearableUseCase(&mockMeasurementRepo{}, &mockWearableSummaryRepo{}, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "", sampleMeasurements("x"), entities.WearableSourceWhoop)
	if err == nil {
		t.Fatal("expected error for empty client ID")
	}
	var validationErr *entities.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestExtractWearableUseCase_Execute_InvalidSource_ReturnsValidationError(t *testing.T) {
	uc := NewExtractWearableUseCase(&mockMeasurementRepo{}, &mockWearableSummaryRepo{}, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "client-1", sampleMeasurements("client-1"), "unknown_source")
	if err == nil {
		t.Fatal("expected error for invalid source")
	}
	var validationErr *entities.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestExtractWearableUseCase_Execute_EmptyMeasurements_ReturnsValidationError(t *testing.T) {
	uc := NewExtractWearableUseCase(&mockMeasurementRepo{}, &mockWearableSummaryRepo{}, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "client-1", nil, entities.WearableSourceWhoop)
	if err == nil {
		t.Fatal("expected error for empty measurements")
	}
	var validationErr *entities.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestExtractWearableUseCase_Execute_DuplicateMeasurements_ReportsSkipped(t *testing.T) {
	measurementRepo := &mockMeasurementRepo{
		upsertBatchFunc: func(_ context.Context, measurements []entities.Measurement) (int, error) {
			// Simulate 1 out of 3 inserted (2 duplicates)
			return 1, nil
		},
	}
	uc := NewExtractWearableUseCase(measurementRepo, &mockWearableSummaryRepo{}, &mockAuditRepo{})

	result, err := uc.Execute(context.Background(), "client-1", sampleMeasurements("client-1"), entities.WearableSourceWhoop)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalInserted != 1 {
		t.Errorf("expected total_inserted 1, got %d", result.TotalInserted)
	}
	if result.DuplicatesSkipped != 2 {
		t.Errorf("expected duplicates_skipped 2, got %d", result.DuplicatesSkipped)
	}
}

func TestExtractWearableUseCase_Execute_UpsertError_PropagatesError(t *testing.T) {
	expectedErr := errors.New("database unavailable")
	measurementRepo := &mockMeasurementRepo{
		upsertBatchFunc: func(_ context.Context, _ []entities.Measurement) (int, error) {
			return 0, expectedErr
		},
	}
	uc := NewExtractWearableUseCase(measurementRepo, &mockWearableSummaryRepo{}, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "client-1", sampleMeasurements("client-1"), entities.WearableSourceWhoop)
	if err == nil {
		t.Fatal("expected error from upsert")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected wrapped error, got: %v", err)
	}
}

func TestExtractWearableUseCase_Execute_AverageError_PropagatesError(t *testing.T) {
	expectedErr := errors.New("average computation failed")
	measurementRepo := &mockMeasurementRepo{
		averageFunc: func(_ context.Context, _ string, _ entities.WearableSource, _ entities.MeasurementType, _ time.Time) (*float64, error) {
			return nil, expectedErr
		},
	}
	uc := NewExtractWearableUseCase(measurementRepo, &mockWearableSummaryRepo{}, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "client-1", sampleMeasurements("client-1"), entities.WearableSourceWhoop)
	if err == nil {
		t.Fatal("expected error from average computation")
	}
}

func TestExtractWearableUseCase_Execute_SummaryUpsertError_PropagatesError(t *testing.T) {
	expectedErr := errors.New("summary save failed")
	summaryRepo := &mockWearableSummaryRepo{
		upsertFunc: func(_ context.Context, _ string, _ entities.WearableSource, _ json.RawMessage) error {
			return expectedErr
		},
	}
	uc := NewExtractWearableUseCase(&mockMeasurementRepo{}, summaryRepo, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "client-1", sampleMeasurements("client-1"), entities.WearableSourceWhoop)
	if err == nil {
		t.Fatal("expected error from summary upsert")
	}
}

func TestExtractWearableUseCase_Execute_LogsAuditEvent(t *testing.T) {
	auditRepo := &mockAuditRepo{}
	uc := NewExtractWearableUseCase(&mockMeasurementRepo{}, &mockWearableSummaryRepo{}, auditRepo)

	_, err := uc.Execute(context.Background(), "client-1", sampleMeasurements("client-1"), entities.WearableSourceWhoop)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(auditRepo.events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(auditRepo.events))
	}
	event := auditRepo.events[0]
	if event.Action != "wearable.import" {
		t.Errorf("expected action %q, got %q", "wearable.import", event.Action)
	}
	if event.EntityType != "wearable_data" {
		t.Errorf("expected entity_type %q, got %q", "wearable_data", event.EntityType)
	}
}

func TestExtractWearableUseCase_Execute_ComputesRollingAverages(t *testing.T) {
	avg := 45.0
	measurementRepo := &mockMeasurementRepo{
		averageFunc: func(_ context.Context, _ string, _ entities.WearableSource, _ entities.MeasurementType, _ time.Time) (*float64, error) {
			return &avg, nil
		},
	}
	summaryRepo := &mockWearableSummaryRepo{}
	uc := NewExtractWearableUseCase(measurementRepo, summaryRepo, &mockAuditRepo{})

	_, err := uc.Execute(context.Background(), "client-1", sampleMeasurements("client-1"), entities.WearableSourceWhoop)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summaryRepo.lastMetrics == nil {
		t.Fatal("expected summary metrics to be set")
	}

	var averages map[string]float64
	if err := json.Unmarshal(summaryRepo.lastMetrics, &averages); err != nil {
		t.Fatalf("unmarshal metrics: %v", err)
	}

	// Should contain averages for 6 metrics × 3 windows = 18 keys
	if len(averages) != 18 {
		t.Errorf("expected 18 average keys, got %d: %v", len(averages), averages)
	}

	// Verify a few specific keys
	expectedKeys := []string{"avg_hrv_7d", "avg_hrv_14d", "avg_hrv_30d", "avg_sleep_7d", "avg_recovery_7d", "avg_steps_30d"}
	for _, key := range expectedKeys {
		val, ok := averages[key]
		if !ok {
			t.Errorf("expected key %q in averages", key)
			continue
		}
		if val != 45.0 {
			t.Errorf("expected %q = 45.0, got %f", key, val)
		}
	}
}

func TestExtractWearableUseCase_Execute_AllValidSources_Succeed(t *testing.T) {
	sources := []entities.WearableSource{
		entities.WearableSourceWhoop,
		entities.WearableSourceGarmin,
		entities.WearableSourceAppleHealth,
		entities.WearableSourceOura,
		entities.WearableSourceFitbit,
	}

	for _, src := range sources {
		t.Run(string(src), func(t *testing.T) {
			uc := NewExtractWearableUseCase(&mockMeasurementRepo{}, &mockWearableSummaryRepo{}, &mockAuditRepo{})
			ms := []entities.Measurement{
				{ClientID: "client-1", Source: src, MeasurementType: entities.MeasurementTypeHRV, Value: 45, MeasuredAt: time.Now()},
			}
			result, err := uc.Execute(context.Background(), "client-1", ms, src)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Source != src {
				t.Errorf("expected source %q, got %q", src, result.Source)
			}
		})
	}
}
