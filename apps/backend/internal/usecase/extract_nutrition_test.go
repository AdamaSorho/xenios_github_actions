package usecase

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/infrastructure/nutrition"
)

const sampleMFPCSV = `Date,Meal,Calories,Fat (g),Protein (g),Carbs (g),Fiber (g)
2026-01-15,Breakfast,450,12,35,48,6
2026-01-15,Lunch,680,22,45,65,8
2026-01-15,Dinner,720,28,40,70,5
2026-01-15,Snack,200,8,12,20,2
`

const sampleMultiDayCSV = `Date,Meal,Calories,Fat (g),Protein (g),Carbs (g),Fiber (g)
2026-01-15,Breakfast,450,12,35,48,6
2026-01-15,Lunch,680,22,45,65,8
2026-01-16,Breakfast,500,15,40,50,7
2026-01-16,Lunch,700,25,42,68,9
`

const sampleGenericCSV = `date,calories,protein,carbs,fat,fiber
2026-01-15,2050,132,203,70,21
2026-01-16,2200,150,210,75,25
`

const sampleCSVWithBadRow = `Date,Meal,Calories,Fat (g),Protein (g),Carbs (g),Fiber (g)
2026-01-15,Breakfast,450,12,35,48,6
2026-01-15,Lunch,abc,22,45,65,8
2026-01-15,Dinner,720,28,40,70,5
`

func newExtractNutritionDeps() (
	*ExtractNutritionUseCase,
	*repository.InMemoryArtifactRepository,
	*repository.InMemoryFileStorage,
	*repository.InMemoryMeasurementRepository,
	*repository.InMemoryNutritionSummaryRepository,
	*repository.InMemoryAuditRepository,
) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	measurementRepo := repository.NewInMemoryMeasurementRepository()
	summaryRepo := repository.NewInMemoryNutritionSummaryRepository()
	auditRepo := repository.NewInMemoryAuditRepository()

	parser := nutrition.NewCSVParser()
	uc := NewExtractNutritionUseCase(artifactRepo, fileStorage, measurementRepo, summaryRepo, auditRepo, parser)
	return uc, artifactRepo, fileStorage, measurementRepo, summaryRepo, auditRepo
}

func createTestArtifactWithCSV(t *testing.T, artifactRepo *repository.InMemoryArtifactRepository, fileStorage *repository.InMemoryFileStorage, csvContent string) *entities.Artifact {
	t.Helper()
	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "nutrition.csv",
		FileType:    "text/csv",
		FileSize:    int64(len(csvContent)),
		StorageKey:  "client-1/document/test-nutrition.csv",
		Type:        entities.ArtifactTypeDocument,
		Status:      entities.ArtifactStatusUploaded,
		ContentType: "text/csv",
	}
	created, err := artifactRepo.Create(context.Background(), art)
	if err != nil {
		t.Fatalf("failed to create test artifact: %v", err)
	}
	fileStorage.PutObjectWithContent(created.StorageKey, []byte(csvContent))
	return created
}

func TestExtractNutrition_ValidMFPCSV_StoresMeasurements(t *testing.T) {
	uc, artifactRepo, fileStorage, measurementRepo, _, _ := newExtractNutritionDeps()
	artifact := createTestArtifactWithCSV(t, artifactRepo, fileStorage, sampleMFPCSV)

	out, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		ClientID:   "client-1",
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.RowsProcessed != 4 {
		t.Errorf("expected 4 rows processed, got %d", out.RowsProcessed)
	}

	// 1 day x 5 measurement types = 5 measurements
	measurements := measurementRepo.GetAll()
	if len(measurements) != 5 {
		t.Errorf("expected 5 measurements (1 day x 5 types), got %d", len(measurements))
	}
}

func TestExtractNutrition_MultiDay_StoresDailyTotals(t *testing.T) {
	uc, artifactRepo, fileStorage, measurementRepo, _, _ := newExtractNutritionDeps()
	artifact := createTestArtifactWithCSV(t, artifactRepo, fileStorage, sampleMultiDayCSV)

	out, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		ClientID:   "client-1",
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.DaysProcessed != 2 {
		t.Errorf("expected 2 days processed, got %d", out.DaysProcessed)
	}

	// 2 days x 5 measurement types = 10 measurements
	measurements := measurementRepo.GetAll()
	if len(measurements) != 10 {
		t.Errorf("expected 10 measurements (2 days x 5 types), got %d", len(measurements))
	}
}

func TestExtractNutrition_MultiDay_ComputesDailyTotalsCorrectly(t *testing.T) {
	uc, artifactRepo, fileStorage, measurementRepo, _, _ := newExtractNutritionDeps()
	artifact := createTestArtifactWithCSV(t, artifactRepo, fileStorage, sampleMFPCSV)

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		ClientID:   "client-1",
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that the daily total for calories is 450+680+720+200 = 2050
	measurements := measurementRepo.GetAll()
	for _, m := range measurements {
		if m.MeasurementType == "calories" {
			if m.Value != 2050 {
				t.Errorf("expected daily calories total 2050, got %g", m.Value)
			}
			if m.Unit != "kcal" {
				t.Errorf("expected unit kcal, got %s", m.Unit)
			}
		}
	}
}

func TestExtractNutrition_GenericCSV_ParsesCorrectly(t *testing.T) {
	uc, artifactRepo, fileStorage, measurementRepo, _, _ := newExtractNutritionDeps()
	artifact := createTestArtifactWithCSV(t, artifactRepo, fileStorage, sampleGenericCSV)

	out, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		ClientID:   "client-1",
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.RowsProcessed != 2 {
		t.Errorf("expected 2 rows processed, got %d", out.RowsProcessed)
	}
	if out.DaysProcessed != 2 {
		t.Errorf("expected 2 days processed, got %d", out.DaysProcessed)
	}

	measurements := measurementRepo.GetAll()
	if len(measurements) != 10 {
		t.Errorf("expected 10 measurements (2 days x 5 types), got %d", len(measurements))
	}
}

func TestExtractNutrition_BadRows_SkipsAndContinues(t *testing.T) {
	uc, artifactRepo, fileStorage, measurementRepo, _, _ := newExtractNutritionDeps()
	artifact := createTestArtifactWithCSV(t, artifactRepo, fileStorage, sampleCSVWithBadRow)

	out, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		ClientID:   "client-1",
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.RowsProcessed != 2 {
		t.Errorf("expected 2 rows processed, got %d", out.RowsProcessed)
	}
	if out.RowsSkipped != 1 {
		t.Errorf("expected 1 row skipped, got %d", out.RowsSkipped)
	}

	// 1 day x 5 types = 5 measurements (both valid rows are same day)
	measurements := measurementRepo.GetAll()
	if len(measurements) != 5 {
		t.Errorf("expected 5 measurements, got %d", len(measurements))
	}
}

func TestExtractNutrition_StoresNutritionSummary(t *testing.T) {
	uc, artifactRepo, fileStorage, _, summaryRepo, _ := newExtractNutritionDeps()
	artifact := createTestArtifactWithCSV(t, artifactRepo, fileStorage, sampleGenericCSV)

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		ClientID:   "client-1",
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	summaries := summaryRepo.GetAll()
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	summary := summaries[0]
	if summary.ClientID != "client-1" {
		t.Errorf("expected client_id client-1, got %s", summary.ClientID)
	}
	if summary.ArtifactID != artifact.ID {
		t.Errorf("expected artifact_id %s, got %s", artifact.ID, summary.ArtifactID)
	}
	// avg over 2 days: (2050+2200)/2 = 2125
	expectedAvgCal := 2125.0
	if summary.AvgCalories7d != expectedAvgCal {
		t.Errorf("expected avg_calories_7d %g, got %g", expectedAvgCal, summary.AvgCalories7d)
	}
}

func TestExtractNutrition_LogsAuditEvent(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _, auditRepo := newExtractNutritionDeps()
	artifact := createTestArtifactWithCSV(t, artifactRepo, fileStorage, sampleMFPCSV)

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		ClientID:   "client-1",
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := auditRepo.GetEvents()
	if len(events) == 0 {
		t.Fatal("expected audit event to be logged")
	}
	if events[0].Action != "nutrition.csv_imported" {
		t.Errorf("expected action 'nutrition.csv_imported', got '%s'", events[0].Action)
	}
	if events[0].EntityType != "artifact" {
		t.Errorf("expected entity_type 'artifact', got '%s'", events[0].EntityType)
	}
}

func TestExtractNutrition_EmptyArtifactID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _, _ := newExtractNutritionDeps()

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: "",
		ClientID:   "client-1",
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractNutrition_EmptyClientID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _, _ := newExtractNutritionDeps()

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: "art-1",
		ClientID:   "",
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractNutrition_EmptyCoachID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _, _ := newExtractNutritionDeps()

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: "art-1",
		ClientID:   "client-1",
		CoachID:    "",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractNutrition_ArtifactNotFound_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _, _ := newExtractNutritionDeps()

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: "nonexistent",
		ClientID:   "client-1",
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractNutrition_ArtifactNotUploaded_ReturnsValidationError(t *testing.T) {
	uc, artifactRepo, _, _, _, _ := newExtractNutritionDeps()

	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "nutrition.csv",
		FileType:    "text/csv",
		FileSize:    100,
		StorageKey:  "client-1/document/test.csv",
		Type:        entities.ArtifactTypeDocument,
		Status:      entities.ArtifactStatusPending,
		ContentType: "text/csv",
	}
	created, err := artifactRepo.Create(context.Background(), art)
	if err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}

	_, err = uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: created.ID,
		ClientID:   "client-1",
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractNutrition_LinksToArtifactID(t *testing.T) {
	uc, artifactRepo, fileStorage, measurementRepo, _, _ := newExtractNutritionDeps()
	artifact := createTestArtifactWithCSV(t, artifactRepo, fileStorage, sampleGenericCSV)

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		ClientID:   "client-1",
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	measurements := measurementRepo.GetAll()
	for _, m := range measurements {
		if m.ArtifactID != artifact.ID {
			t.Errorf("expected measurement to link to artifact %s, got %s", artifact.ID, m.ArtifactID)
		}
	}
}

func TestExtractNutrition_SingleDay_AveragesEqualDailyTotal(t *testing.T) {
	csv := `date,calories,protein,carbs,fat,fiber
2026-01-15,2100,165,220,78,28
`
	uc, artifactRepo, fileStorage, _, summaryRepo, _ := newExtractNutritionDeps()
	artifact := createTestArtifactWithCSV(t, artifactRepo, fileStorage, csv)

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		ClientID:   "client-1",
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	summaries := summaryRepo.GetAll()
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	s := summaries[0]
	if s.AvgCalories7d != 2100 {
		t.Errorf("expected avg_calories_7d 2100, got %g", s.AvgCalories7d)
	}
	if s.AvgCalories14d != 2100 {
		t.Errorf("expected avg_calories_14d 2100, got %g", s.AvgCalories14d)
	}
	if s.AvgCalories30d != 2100 {
		t.Errorf("expected avg_calories_30d 2100, got %g", s.AvgCalories30d)
	}
}
