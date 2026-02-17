package usecase

import (
	"context"
	"strings"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/infrastructure/parser"
)

func newExtractWearableUseCase() (*ExtractWearableUseCase, *repository.InMemoryMeasurementRepository, *repository.InMemoryWearableSummaryRepository, *repository.InMemoryAuditRepository) {
	measurementRepo := repository.NewInMemoryMeasurementRepository()
	summaryRepo := repository.NewInMemoryWearableSummaryRepository()
	auditRepo := repository.NewInMemoryAuditRepository()

	parsers := []parser.WearableParser{
		parser.NewWhoopParser(),
		parser.NewGarminParser(),
		parser.NewAppleHealthParser(),
		parser.NewOuraParser(),
		parser.NewFitbitParser(),
	}

	uc := NewExtractWearableUseCase(measurementRepo, summaryRepo, auditRepo, parsers)
	return uc, measurementRepo, summaryRepo, auditRepo
}

func TestExtractWearable_WhoopCSV_ExtractsAllMetrics(t *testing.T) {
	uc, measurementRepo, summaryRepo, _ := newExtractWearableUseCase()

	csvData := `Cycle Start Time,Cycle End Time,HRV (ms),Resting Heart Rate (bpm),Recovery Score (%),Sleep Duration (hrs),Strain Score
2024-01-01 06:00:00,2024-01-02 06:00:00,45.2,58,72,7.5,12.3
2024-01-02 06:00:00,2024-01-03 06:00:00,42.1,60,65,6.8,14.1
2024-01-03 06:00:00,2024-01-04 06:00:00,48.5,56,78,8.1,10.5`

	input := ExtractWearableInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
		Data:     strings.NewReader(csvData),
	}

	out, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Source != entities.WearableSourceWhoop {
		t.Errorf("expected source whoop, got %q", out.Source)
	}
	// 3 rows * 5 metrics = 15
	if out.MeasurementsInserted != 15 {
		t.Errorf("expected 15 measurements inserted, got %d", out.MeasurementsInserted)
	}
	if measurementRepo.Count() != 15 {
		t.Errorf("expected 15 measurements in repo, got %d", measurementRepo.Count())
	}
	if summaryRepo.Count() == 0 {
		t.Error("expected at least one wearable summary")
	}
}

func TestExtractWearable_GarminCSV_ExtractsMetrics(t *testing.T) {
	uc, measurementRepo, _, _ := newExtractWearableUseCase()

	csvData := `Date,Steps,Resting Heart Rate (bpm),HRV (ms),Sleep Duration (hrs),Stress Score
2024-01-01,8500,62,40.3,7.2,35
2024-01-02,10200,60,43.1,6.5,28`

	input := ExtractWearableInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
		Data:     strings.NewReader(csvData),
	}

	out, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Source != entities.WearableSourceGarmin {
		t.Errorf("expected source garmin, got %q", out.Source)
	}
	// 2 rows * 4 metrics = 8
	if out.MeasurementsInserted != 8 {
		t.Errorf("expected 8 measurements inserted, got %d", out.MeasurementsInserted)
	}
	if measurementRepo.Count() != 8 {
		t.Errorf("expected 8 measurements in repo, got %d", measurementRepo.Count())
	}
}

func TestExtractWearable_DuplicateImport_SkipsExisting(t *testing.T) {
	uc, measurementRepo, _, _ := newExtractWearableUseCase()

	csvData := `Cycle Start Time,Cycle End Time,HRV (ms),Resting Heart Rate (bpm),Recovery Score (%),Sleep Duration (hrs),Strain Score
2024-01-01 06:00:00,2024-01-02 06:00:00,45.2,58,72,7.5,12.3`

	input := ExtractWearableInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
		Data:     strings.NewReader(csvData),
	}

	// First import
	out1, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out1.MeasurementsInserted != 5 {
		t.Errorf("expected 5 measurements inserted first time, got %d", out1.MeasurementsInserted)
	}

	// Second import (duplicate)
	input.Data = strings.NewReader(csvData)
	out2, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error on second import: %v", err)
	}
	if out2.MeasurementsInserted != 0 {
		t.Errorf("expected 0 measurements inserted on duplicate, got %d", out2.MeasurementsInserted)
	}
	if measurementRepo.Count() != 5 {
		t.Errorf("expected total count still 5, got %d", measurementRepo.Count())
	}
}

func TestExtractWearable_UnknownFormat_ReturnsError(t *testing.T) {
	uc, _, _, _ := newExtractWearableUseCase()

	csvData := `Unknown,Columns,Here
1,2,3`

	input := ExtractWearableInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
		Data:     strings.NewReader(csvData),
	}

	_, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func TestExtractWearable_EmptyClientID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newExtractWearableUseCase()

	csvData := `Cycle Start Time,HRV (ms)
2024-01-01 06:00:00,45.2`

	input := ExtractWearableInput{
		ClientID: "",
		CoachID:  "coach-1",
		Data:     strings.NewReader(csvData),
	}

	_, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for empty client ID")
	}
	var ve *entities.ValidationError
	if !isValidationErr(err, &ve) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestExtractWearable_EmptyCoachID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newExtractWearableUseCase()

	csvData := `Cycle Start Time,HRV (ms)
2024-01-01 06:00:00,45.2`

	input := ExtractWearableInput{
		ClientID: "client-1",
		CoachID:  "",
		Data:     strings.NewReader(csvData),
	}

	_, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for empty coach ID")
	}
}

func TestExtractWearable_NilData_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newExtractWearableUseCase()

	input := ExtractWearableInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
		Data:     nil,
	}

	_, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for nil data")
	}
}

func TestExtractWearable_AuditEventLogged(t *testing.T) {
	uc, _, _, auditRepo := newExtractWearableUseCase()

	csvData := `Cycle Start Time,Cycle End Time,HRV (ms),Resting Heart Rate (bpm),Recovery Score (%),Sleep Duration (hrs),Strain Score
2024-01-01 06:00:00,2024-01-02 06:00:00,45.2,58,72,7.5,12.3`

	input := ExtractWearableInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
		Data:     strings.NewReader(csvData),
	}

	_, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if auditRepo.EventCount() == 0 {
		t.Fatal("expected audit event to be logged")
	}
	events := auditRepo.GetEvents()
	if events[0].Action != "wearable.import" {
		t.Errorf("expected action 'wearable.import', got '%s'", events[0].Action)
	}
	if events[0].ActorID != "coach-1" {
		t.Errorf("expected actor 'coach-1', got '%s'", events[0].ActorID)
	}
}

func TestExtractWearable_RollingAverages_Computed(t *testing.T) {
	uc, _, summaryRepo, _ := newExtractWearableUseCase()

	csvData := `Cycle Start Time,Cycle End Time,HRV (ms),Resting Heart Rate (bpm),Recovery Score (%),Sleep Duration (hrs),Strain Score
2024-01-01 06:00:00,2024-01-02 06:00:00,45.2,58,72,7.5,12.3
2024-01-02 06:00:00,2024-01-03 06:00:00,42.1,60,65,6.8,14.1
2024-01-03 06:00:00,2024-01-04 06:00:00,48.5,56,78,8.1,10.5`

	input := ExtractWearableInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
		Data:     strings.NewReader(csvData),
	}

	_, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	summaries := summaryRepo.GetAll()
	if len(summaries) == 0 {
		t.Fatal("expected wearable summary to be created")
	}

	// Check that rolling averages are in the metrics
	for _, s := range summaries {
		if _, ok := s.Metrics["avg_hrv_ms_7d"]; !ok {
			t.Error("expected avg_hrv_ms_7d in metrics")
		}
	}
}

// isValidationErr checks if err is a *entities.ValidationError.
func isValidationErr(err error, target **entities.ValidationError) bool {
	ve, ok := err.(*entities.ValidationError)
	if ok {
		*target = ve
	}
	return ok
}
