package usecase

import (
	"context"
	"fmt"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

// stubCSVParser implements NutritionCSVParser for testing.
type stubCSVParser struct {
	logs []entities.NutritionDailyLog
	err  error
}

func (p *stubCSVParser) Parse(_ []byte) ([]entities.NutritionDailyLog, error) {
	return p.logs, p.err
}

func newExtractNutritionTestDeps(parser NutritionCSVParser) (
	*ExtractNutritionUseCase,
	*repository.InMemoryNutritionRepository,
	*repository.InMemoryAuditRepository,
) {
	nutritionRepo := repository.NewInMemoryNutritionRepository()
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewExtractNutritionUseCase(nutritionRepo, auditRepo, parser)
	return uc, nutritionRepo, auditRepo
}

func sampleDailyLogs() []entities.NutritionDailyLog {
	return []entities.NutritionDailyLog{
		{Calories: 2000, Protein: 150, Carbs: 200, Fat: 80, Fiber: 30},
		{Calories: 2200, Protein: 160, Carbs: 220, Fat: 90, Fiber: 25},
		{Calories: 1800, Protein: 140, Carbs: 180, Fat: 70, Fiber: 35},
	}
}

func validInput() ExtractNutritionInput {
	return ExtractNutritionInput{
		CSVData:    []byte("date,calories,protein,carbs,fat,fiber\n2024-01-15,2000,150,200,80,30\n"),
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "artifact-1",
	}
}

func TestExtractNutrition_EmptyCSV_ReturnsValidationError(t *testing.T) {
	parser := &stubCSVParser{}
	uc, _, _ := newExtractNutritionTestDeps(parser)

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		CSVData:    nil,
		ClientID:   "c1",
		CoachID:    "co1",
		ArtifactID: "a1",
	})
	if err == nil {
		t.Fatal("expected error for empty CSV")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractNutrition_EmptyClientID_ReturnsValidationError(t *testing.T) {
	parser := &stubCSVParser{}
	uc, _, _ := newExtractNutritionTestDeps(parser)

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		CSVData:    []byte("data"),
		ClientID:   "",
		CoachID:    "co1",
		ArtifactID: "a1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractNutrition_EmptyCoachID_ReturnsValidationError(t *testing.T) {
	parser := &stubCSVParser{}
	uc, _, _ := newExtractNutritionTestDeps(parser)

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		CSVData:    []byte("data"),
		ClientID:   "c1",
		CoachID:    "",
		ArtifactID: "a1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractNutrition_EmptyArtifactID_ReturnsValidationError(t *testing.T) {
	parser := &stubCSVParser{}
	uc, _, _ := newExtractNutritionTestDeps(parser)

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		CSVData:    []byte("data"),
		ClientID:   "c1",
		CoachID:    "co1",
		ArtifactID: "",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractNutrition_ParserError_ReturnsValidationError(t *testing.T) {
	parser := &stubCSVParser{err: fmt.Errorf("bad format")}
	uc, _, _ := newExtractNutritionTestDeps(parser)

	_, err := uc.Execute(context.Background(), validInput())
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractNutrition_ValidInput_CreatesMeasurements(t *testing.T) {
	parser := &stubCSVParser{logs: sampleDailyLogs()}
	uc, nutritionRepo, _ := newExtractNutritionTestDeps(parser)

	out, err := uc.Execute(context.Background(), validInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 3 days * 5 measurement types = 15
	if out.MeasurementsCreated != 15 {
		t.Errorf("expected 15 measurements, got %d", out.MeasurementsCreated)
	}

	stored := nutritionRepo.GetMeasurements()
	if len(stored) != 15 {
		t.Errorf("expected 15 stored measurements, got %d", len(stored))
	}
}

func TestExtractNutrition_ValidInput_ReturnsDaysProcessed(t *testing.T) {
	parser := &stubCSVParser{logs: sampleDailyLogs()}
	uc, _, _ := newExtractNutritionTestDeps(parser)

	out, err := uc.Execute(context.Background(), validInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.DaysProcessed != 3 {
		t.Errorf("expected 3 days, got %d", out.DaysProcessed)
	}
}

func TestExtractNutrition_ValidInput_ComputesAverages(t *testing.T) {
	parser := &stubCSVParser{logs: sampleDailyLogs()}
	uc, _, _ := newExtractNutritionTestDeps(parser)

	out, err := uc.Execute(context.Background(), validInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Averages == nil {
		t.Fatal("expected averages to be computed")
	}
	// Should have 7-day, 14-day, 30-day averages
	for _, period := range []int{7, 14, 30} {
		avg, ok := out.Averages[period]
		if !ok {
			t.Errorf("expected %d-day average", period)
			continue
		}
		if avg.PeriodDays != period {
			t.Errorf("expected period %d, got %d", period, avg.PeriodDays)
		}
	}

	// With 3 days: avg calories = (2000+2200+1800)/3 = 2000
	avg7 := out.Averages[7]
	if avg7.AvgCalories != 2000 {
		t.Errorf("expected avg calories 2000, got %f", avg7.AvgCalories)
	}
}

func TestExtractNutrition_ValidInput_StoresAverages(t *testing.T) {
	parser := &stubCSVParser{logs: sampleDailyLogs()}
	uc, nutritionRepo, _ := newExtractNutritionTestDeps(parser)

	_, err := uc.Execute(context.Background(), validInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored := nutritionRepo.GetAverages()
	if len(stored) != 3 {
		t.Fatalf("expected 3 stored averages, got %d", len(stored))
	}

	for _, avg := range stored {
		if avg.ClientID != "client-1" {
			t.Errorf("expected client_id 'client-1', got '%s'", avg.ClientID)
		}
		if avg.SourceArtifactID != "artifact-1" {
			t.Errorf("expected artifact_id 'artifact-1', got '%s'", avg.SourceArtifactID)
		}
	}
}

func TestExtractNutrition_ValidInput_LogsAuditEvent(t *testing.T) {
	parser := &stubCSVParser{logs: sampleDailyLogs()}
	uc, _, auditRepo := newExtractNutritionTestDeps(parser)

	_, err := uc.Execute(context.Background(), validInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := auditRepo.GetEvents()
	if len(events) == 0 {
		t.Fatal("expected audit event to be logged")
	}
	if events[0].Action != "nutrition.import" {
		t.Errorf("expected action 'nutrition.import', got '%s'", events[0].Action)
	}
	if events[0].EntityType != "artifact" {
		t.Errorf("expected entity_type 'artifact', got '%s'", events[0].EntityType)
	}
	if events[0].EntityID != "artifact-1" {
		t.Errorf("expected entity_id 'artifact-1', got '%s'", events[0].EntityID)
	}
}

func TestExtractNutrition_MeasurementsLinkedToArtifact(t *testing.T) {
	parser := &stubCSVParser{logs: sampleDailyLogs()}
	uc, nutritionRepo, _ := newExtractNutritionTestDeps(parser)

	_, err := uc.Execute(context.Background(), validInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored := nutritionRepo.GetMeasurements()
	for _, m := range stored {
		if m.SourceArtifactID != "artifact-1" {
			t.Errorf("expected source_artifact_id 'artifact-1', got '%s'", m.SourceArtifactID)
		}
	}
}

func TestExtractNutrition_MeasurementTypes_IncludeAllNutritionTypes(t *testing.T) {
	parser := &stubCSVParser{logs: sampleDailyLogs()[:1]} // 1 day
	uc, nutritionRepo, _ := newExtractNutritionTestDeps(parser)

	_, err := uc.Execute(context.Background(), validInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored := nutritionRepo.GetMeasurements()
	typeSet := make(map[string]bool)
	for _, m := range stored {
		typeSet[m.MeasurementType] = true
	}

	expectedTypes := []string{"calories_kcal", "protein_g", "carbs_g", "fat_g", "fiber_g"}
	for _, et := range expectedTypes {
		if !typeSet[et] {
			t.Errorf("expected measurement type '%s' to be present", et)
		}
	}
}
