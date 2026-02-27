package usecase

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/infrastructure/nutrition"
)

const mfpCSV = `Date,Meal,Calories,Fat (g),Protein (g),Carbs (g),Fiber (g)
2026-01-15,Breakfast,450,12,35,48,6
2026-01-15,Lunch,680,22,45,65,8
2026-01-15,Dinner,720,28,40,70,5
2026-01-15,Snack,200,8,12,20,2
`

const multiDayCSV = `Date,Meal,Calories,Fat (g),Protein (g),Carbs (g),Fiber (g)
2026-01-15,Breakfast,450,12,35,48,6
2026-01-15,Lunch,680,22,45,65,8
2026-01-15,Dinner,720,28,40,70,5
2026-01-15,Snack,200,8,12,20,2
2026-01-16,Breakfast,400,10,30,45,5
2026-01-16,Lunch,650,20,42,60,7
2026-01-16,Dinner,700,25,38,68,4
`

const genericCSV = `date,calories,protein,carbs,fat,fiber
2026-01-15,2050,132,203,70,21
2026-01-16,1750,110,173,55,16
2026-01-17,2200,145,230,80,24
`

const singleDayCSV = `date,calories,protein,carbs,fat,fiber
2026-01-15,2050,132,203,70,21
`

const badValuesCSV = `Date,Meal,Calories,Fat (g),Protein (g),Carbs (g),Fiber (g)
2026-01-15,Breakfast,450,12,35,48,6
2026-01-15,Lunch,bad_value,22,45,65,8
2026-01-15,Dinner,720,28,40,70,5
`

func newExtractNutritionUseCase() (
	*ExtractNutritionUseCase,
	*repository.InMemoryArtifactRepository,
	*repository.InMemoryFileStorage,
	*repository.InMemoryMeasurementRepository,
	*repository.InMemoryNutritionSummaryRepository,
	*repository.InMemoryAuditRepository,
) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	parser := nutrition.NewCSVParser()
	measurementRepo := repository.NewInMemoryMeasurementRepository()
	summaryRepo := repository.NewInMemoryNutritionSummaryRepository()
	auditRepo := repository.NewInMemoryAuditRepository()

	uc := NewExtractNutritionUseCase(artifactRepo, fileStorage, parser, measurementRepo, summaryRepo, auditRepo)
	return uc, artifactRepo, fileStorage, measurementRepo, summaryRepo, auditRepo
}

func createTestArtifact(t *testing.T, repo *repository.InMemoryArtifactRepository, storage *repository.InMemoryFileStorage, csvContent string) *entities.Artifact {
	t.Helper()
	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "nutrition.csv",
		FileType:    "text/csv",
		FileSize:    int64(len(csvContent)),
		StorageKey:  "client-1/document/nutrition-test.csv",
		Type:        entities.ArtifactTypeDocument,
		Status:      entities.ArtifactStatusUploaded,
		ContentType: "text/csv",
	}
	created, err := repo.Create(context.Background(), art)
	if err != nil {
		t.Fatalf("failed to create test artifact: %v", err)
	}
	storage.PutObjectWithContent(created.StorageKey, []byte(csvContent))
	return created
}

func TestExtractNutrition_HappyPath_StoresDailyMeasurements(t *testing.T) {
	uc, artifactRepo, fileStorage, measurementRepo, _, _ := newExtractNutritionUseCase()
	artifact := createTestArtifact(t, artifactRepo, fileStorage, mfpCSV)

	out, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		ClientID:   "client-1",
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.DaysProcessed != 1 {
		t.Errorf("expected 1 day processed, got %d", out.DaysProcessed)
	}

	measurements := measurementRepo.GetAll()
	// 1 day × 5 metrics (calories, protein, carbs, fat, fiber) = 5 measurements
	if len(measurements) != 5 {
		t.Errorf("expected 5 measurements, got %d", len(measurements))
	}
}

func TestExtractNutrition_MultipleDays_StoresCorrectTotals(t *testing.T) {
	uc, artifactRepo, fileStorage, measurementRepo, _, _ := newExtractNutritionUseCase()
	artifact := createTestArtifact(t, artifactRepo, fileStorage, multiDayCSV)

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

	measurements := measurementRepo.GetAll()
	// 2 days × 5 metrics = 10 measurements
	if len(measurements) != 10 {
		t.Errorf("expected 10 measurements, got %d", len(measurements))
	}

	// Check day1 calorie total: 450+680+720+200 = 2050
	var day1Cals float64
	for _, m := range measurements {
		if m.MeasurementType == "calories" && m.MeasuredAt.Day() == 15 {
			day1Cals = m.Value
			break
		}
	}
	if day1Cals != 2050 {
		t.Errorf("expected day1 calories 2050, got %.2f", day1Cals)
	}
}

func TestExtractNutrition_GenericFormat_ParsesCorrectly(t *testing.T) {
	uc, artifactRepo, fileStorage, measurementRepo, _, _ := newExtractNutritionUseCase()
	artifact := createTestArtifact(t, artifactRepo, fileStorage, genericCSV)

	out, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		ClientID:   "client-1",
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.DaysProcessed != 3 {
		t.Errorf("expected 3 days processed, got %d", out.DaysProcessed)
	}

	measurements := measurementRepo.GetAll()
	if len(measurements) != 15 {
		t.Errorf("expected 15 measurements (3 days × 5 metrics), got %d", len(measurements))
	}
}

func TestExtractNutrition_ComputesAverages_StoresSummary(t *testing.T) {
	uc, artifactRepo, fileStorage, _, summaryRepo, _ := newExtractNutritionUseCase()
	artifact := createTestArtifact(t, artifactRepo, fileStorage, genericCSV)

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		ClientID:   "client-1",
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	summaries := summaryRepo.GetByClientID("client-1")
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}

	s := summaries[0]
	// 3 days: 2050, 1750, 2200 → avg = 2000
	expectedAvg := 2000.0
	diff := s.AvgCalories7d - expectedAvg
	if diff < -0.01 || diff > 0.01 {
		t.Errorf("expected 7d avg calories %.2f, got %.2f", expectedAvg, s.AvgCalories7d)
	}
	if s.ArtifactID != artifact.ID {
		t.Errorf("expected artifact ID %q, got %q", artifact.ID, s.ArtifactID)
	}
}

func TestExtractNutrition_SingleDay_AveragesEqualTotals(t *testing.T) {
	uc, artifactRepo, fileStorage, _, summaryRepo, _ := newExtractNutritionUseCase()
	artifact := createTestArtifact(t, artifactRepo, fileStorage, singleDayCSV)

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		ClientID:   "client-1",
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	summaries := summaryRepo.GetByClientID("client-1")
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}

	s := summaries[0]
	if s.AvgCalories7d != 2050 {
		t.Errorf("expected avg calories 2050, got %.2f", s.AvgCalories7d)
	}
	if s.AvgCalories14d != 2050 {
		t.Errorf("expected 14d avg calories 2050, got %.2f", s.AvgCalories14d)
	}
	if s.AvgCalories30d != 2050 {
		t.Errorf("expected 30d avg calories 2050, got %.2f", s.AvgCalories30d)
	}
}

func TestExtractNutrition_BadValues_ProcessesValidRows(t *testing.T) {
	uc, artifactRepo, fileStorage, measurementRepo, _, _ := newExtractNutritionUseCase()
	artifact := createTestArtifact(t, artifactRepo, fileStorage, badValuesCSV)

	out, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		ClientID:   "client-1",
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.DaysProcessed != 1 {
		t.Errorf("expected 1 day processed, got %d", out.DaysProcessed)
	}

	measurements := measurementRepo.GetAll()
	// 1 day × 5 metrics = 5
	if len(measurements) != 5 {
		t.Errorf("expected 5 measurements, got %d", len(measurements))
	}
}

func TestExtractNutrition_EmptyArtifactID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _, _ := newExtractNutritionUseCase()

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
	uc, _, _, _, _, _ := newExtractNutritionUseCase()

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
	uc, _, _, _, _, _ := newExtractNutritionUseCase()

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
	uc, _, _, _, _, _ := newExtractNutritionUseCase()

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

func TestExtractNutrition_AuditEventLogged(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _, auditRepo := newExtractNutritionUseCase()
	artifact := createTestArtifact(t, artifactRepo, fileStorage, mfpCSV)

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
	found := false
	for _, e := range events {
		if e.Action == "nutrition.imported" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'nutrition.imported' audit event")
	}
}

func TestExtractNutrition_MeasurementsLinkedToArtifact(t *testing.T) {
	uc, artifactRepo, fileStorage, measurementRepo, _, _ := newExtractNutritionUseCase()
	artifact := createTestArtifact(t, artifactRepo, fileStorage, mfpCSV)

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
			t.Errorf("expected artifact_id %q, got %q", artifact.ID, m.ArtifactID)
		}
	}
}
