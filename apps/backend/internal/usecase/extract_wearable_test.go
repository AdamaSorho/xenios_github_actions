package usecase

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
	domainrepo "github.com/xenios/backend/internal/domain/repository"
)

// mockWearableParser implements repository.WearableParser for testing.
type mockWearableParser struct {
	source       entities.WearableSource
	parseFunc    func(reader io.Reader, clientID, recordedBy string) ([]entities.Measurement, error)
	detectFunc   func(header []byte) bool
}

func (m *mockWearableParser) Source() entities.WearableSource { return m.source }
func (m *mockWearableParser) Parse(reader io.Reader, clientID, recordedBy string) ([]entities.Measurement, error) {
	if m.parseFunc != nil {
		return m.parseFunc(reader, clientID, recordedBy)
	}
	return nil, nil
}
func (m *mockWearableParser) DetectFormat(header []byte) bool {
	if m.detectFunc != nil {
		return m.detectFunc(header)
	}
	return false
}

var _ domainrepo.WearableParser = &mockWearableParser{}

func newExtractWearableTestDeps() (*ExtractWearableUseCase, *repository.InMemoryMeasurementRepository, *repository.InMemoryWearableSummaryRepository, *repository.InMemoryAuditRepository) {
	measurementRepo := repository.NewInMemoryMeasurementRepository()
	summaryRepo := repository.NewInMemoryWearableSummaryRepository()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewExtractWearableUseCase(measurementRepo, summaryRepo, auditRepo)
	return uc, measurementRepo, summaryRepo, auditRepo
}

func createTestMeasurements(clientID, recordedBy string) []entities.Measurement {
	return []entities.Measurement{
		{
			ClientID:        clientID,
			RecordedBy:      recordedBy,
			MeasurementType: entities.MeasurementTypeHRV,
			Value:           45.2,
			Unit:            "ms",
			MeasuredAt:      time.Now().AddDate(0, 0, -1),
			Source:          entities.WearableSourceWhoop,
		},
		{
			ClientID:        clientID,
			RecordedBy:      recordedBy,
			MeasurementType: entities.MeasurementTypeSleepDuration,
			Value:           7.5,
			Unit:            "hrs",
			MeasuredAt:      time.Now().AddDate(0, 0, -1),
			Source:          entities.WearableSourceWhoop,
		},
	}
}

func TestExtractWearable_ValidInput_ReturnsMeasurements(t *testing.T) {
	uc, measurementRepo, _, auditRepo := newExtractWearableTestDeps()

	testMeasurements := createTestMeasurements("client-1", "coach-1")
	parser := &mockWearableParser{
		source: entities.WearableSourceWhoop,
		parseFunc: func(reader io.Reader, clientID, recordedBy string) ([]entities.Measurement, error) {
			return testMeasurements, nil
		},
	}

	input := ExtractWearableInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		Source:     entities.WearableSourceWhoop,
		Reader:     strings.NewReader("fake csv data"),
		Parser:     parser,
		ArtifactID: "artifact-1",
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.MeasurementsInserted != 2 {
		t.Errorf("expected 2 measurements inserted, got %d", output.MeasurementsInserted)
	}

	if len(output.Averages) == 0 {
		t.Error("expected non-empty averages")
	}

	// Verify measurements were stored
	stored := measurementRepo.GetMeasurements()
	if len(stored) != 2 {
		t.Errorf("expected 2 stored measurements, got %d", len(stored))
	}

	// Verify audit event was logged
	events := auditRepo.GetEvents()
	found := false
	for _, e := range events {
		if e.Action == "artifact.extract" {
			found = true
			if e.ActorID != "coach-1" {
				t.Errorf("expected actor_id %q, got %q", "coach-1", e.ActorID)
			}
			if e.EntityID != "artifact-1" {
				t.Errorf("expected entity_id %q, got %q", "artifact-1", e.EntityID)
			}
			break
		}
	}
	if !found {
		t.Error("expected audit event for artifact.extract")
	}
}

func TestExtractWearable_DuplicateImport_SkipsDuplicates(t *testing.T) {
	uc, measurementRepo, _, _ := newExtractWearableTestDeps()

	testMeasurements := createTestMeasurements("client-1", "coach-1")
	parser := &mockWearableParser{
		source: entities.WearableSourceWhoop,
		parseFunc: func(reader io.Reader, clientID, recordedBy string) ([]entities.Measurement, error) {
			return testMeasurements, nil
		},
	}

	input := ExtractWearableInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		Source:     entities.WearableSourceWhoop,
		Reader:     strings.NewReader("fake csv data"),
		Parser:     parser,
		ArtifactID: "artifact-1",
	}

	// First import
	output1, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error on first import: %v", err)
	}
	if output1.MeasurementsInserted != 2 {
		t.Errorf("expected 2 on first import, got %d", output1.MeasurementsInserted)
	}

	// Second import (same data) - should skip duplicates
	output2, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error on second import: %v", err)
	}
	if output2.MeasurementsInserted != 0 {
		t.Errorf("expected 0 on duplicate import, got %d", output2.MeasurementsInserted)
	}

	// Total stored should still be 2
	stored := measurementRepo.GetMeasurements()
	if len(stored) != 2 {
		t.Errorf("expected 2 total stored measurements, got %d", len(stored))
	}
}

func TestExtractWearable_EmptyClientID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newExtractWearableTestDeps()

	input := ExtractWearableInput{
		ClientID:   "",
		CoachID:    "coach-1",
		Source:     entities.WearableSourceWhoop,
		Reader:     strings.NewReader("data"),
		Parser:     &mockWearableParser{source: entities.WearableSourceWhoop},
		ArtifactID: "artifact-1",
	}

	_, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for empty client_id")
	}

	var validationErr *entities.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestExtractWearable_EmptyCoachID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newExtractWearableTestDeps()

	input := ExtractWearableInput{
		ClientID:   "client-1",
		CoachID:    "",
		Source:     entities.WearableSourceWhoop,
		Reader:     strings.NewReader("data"),
		Parser:     &mockWearableParser{source: entities.WearableSourceWhoop},
		ArtifactID: "artifact-1",
	}

	_, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for empty coach_id")
	}

	var validationErr *entities.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestExtractWearable_InvalidSource_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newExtractWearableTestDeps()

	input := ExtractWearableInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		Source:     "invalid",
		Reader:     strings.NewReader("data"),
		Parser:     &mockWearableParser{source: "invalid"},
		ArtifactID: "artifact-1",
	}

	_, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for invalid source")
	}

	var validationErr *entities.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestExtractWearable_NilReader_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newExtractWearableTestDeps()

	input := ExtractWearableInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		Source:     entities.WearableSourceWhoop,
		Reader:     nil,
		Parser:     &mockWearableParser{source: entities.WearableSourceWhoop},
		ArtifactID: "artifact-1",
	}

	_, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for nil reader")
	}

	var validationErr *entities.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestExtractWearable_NilParser_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newExtractWearableTestDeps()

	input := ExtractWearableInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		Source:     entities.WearableSourceWhoop,
		Reader:     strings.NewReader("data"),
		Parser:     nil,
		ArtifactID: "artifact-1",
	}

	_, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for nil parser")
	}

	var validationErr *entities.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestExtractWearable_ParserError_ReturnsError(t *testing.T) {
	uc, _, _, _ := newExtractWearableTestDeps()

	parser := &mockWearableParser{
		source: entities.WearableSourceWhoop,
		parseFunc: func(reader io.Reader, clientID, recordedBy string) ([]entities.Measurement, error) {
			return nil, errors.New("malformed CSV")
		},
	}

	input := ExtractWearableInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		Source:     entities.WearableSourceWhoop,
		Reader:     strings.NewReader("bad data"),
		Parser:     parser,
		ArtifactID: "artifact-1",
	}

	_, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error from parser failure")
	}
	if !strings.Contains(err.Error(), "malformed CSV") {
		t.Errorf("expected error to contain 'malformed CSV', got %q", err.Error())
	}
}

func TestExtractWearable_ComputesRollingAverages(t *testing.T) {
	uc, _, _, _ := newExtractWearableTestDeps()

	// Create measurements spread over the last 7 days
	now := time.Now()
	var measurements []entities.Measurement
	for i := 0; i < 7; i++ {
		measurements = append(measurements, entities.Measurement{
			ClientID:        "client-1",
			RecordedBy:      "coach-1",
			MeasurementType: entities.MeasurementTypeHRV,
			Value:           40.0 + float64(i),
			Unit:            "ms",
			MeasuredAt:      now.AddDate(0, 0, -i),
			Source:          entities.WearableSourceWhoop,
		})
	}

	parser := &mockWearableParser{
		source: entities.WearableSourceWhoop,
		parseFunc: func(reader io.Reader, clientID, recordedBy string) ([]entities.Measurement, error) {
			return measurements, nil
		},
	}

	input := ExtractWearableInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		Source:     entities.WearableSourceWhoop,
		Reader:     strings.NewReader("fake data"),
		Parser:     parser,
		ArtifactID: "artifact-1",
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that we have averages for 7-day, 14-day, and 30-day windows
	expectedKeys := []string{"avg_hrv_ms_7d", "avg_hrv_ms_14d", "avg_hrv_ms_30d"}
	for _, key := range expectedKeys {
		if _, ok := output.Averages[key]; !ok {
			t.Errorf("expected average key %q in output", key)
		}
	}

	// The 7-day average should be the average of 40,41,42,43,44,45,46 = 43
	avg7d, ok := output.Averages["avg_hrv_ms_7d"]
	if !ok {
		t.Fatal("expected avg_hrv_ms_7d in output")
	}
	if avg7d != 43.0 {
		t.Errorf("expected 7-day HRV average 43.0, got %v", avg7d)
	}
}

func TestExtractWearable_StoresWearableSummary(t *testing.T) {
	uc, _, summaryRepo, _ := newExtractWearableTestDeps()

	testMeasurements := createTestMeasurements("client-1", "coach-1")
	parser := &mockWearableParser{
		source: entities.WearableSourceWhoop,
		parseFunc: func(reader io.Reader, clientID, recordedBy string) ([]entities.Measurement, error) {
			return testMeasurements, nil
		},
	}

	input := ExtractWearableInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		Source:     entities.WearableSourceWhoop,
		Reader:     strings.NewReader("fake data"),
		Parser:     parser,
		ArtifactID: "artifact-1",
	}

	_, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	summaries := summaryRepo.GetSummaries()
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}

	s := summaries[0]
	if s.ClientID != "client-1" {
		t.Errorf("expected client_id %q, got %q", "client-1", s.ClientID)
	}
	if s.Source != entities.WearableSourceWhoop {
		t.Errorf("expected source %q, got %q", entities.WearableSourceWhoop, s.Source)
	}
	if s.Metrics == nil {
		t.Error("expected non-nil metrics")
	}
}

func TestRoundToOneDecimal_RoundsCorrectly(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{45.24, 45.2},
		{45.25, 45.3},
		{45.0, 45.0},
		{0.0, 0.0},
		{100.99, 101.0},
	}
	for _, tt := range tests {
		got := roundToOneDecimal(tt.input)
		if got != tt.expected {
			t.Errorf("roundToOneDecimal(%f) = %f, expected %f", tt.input, got, tt.expected)
		}
	}
}
