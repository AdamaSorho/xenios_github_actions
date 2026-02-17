package usecase

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

const mfpCSV = `Date,Meal,Calories,Fat (g),Protein (g),Carbs (g),Fiber (g)
2026-01-15,Breakfast,450,12,35,48,6
2026-01-15,Lunch,680,22,45,65,8
2026-01-15,Dinner,720,28,40,70,5
2026-01-15,Snack,200,8,12,20,2
2026-01-16,Breakfast,400,10,30,50,5
2026-01-16,Lunch,650,20,42,60,7
2026-01-16,Dinner,700,25,38,68,6
`

const genericCSV = `date,calories,protein,carbs,fat,fiber
2026-01-15,2050,132,203,70,21
2026-01-16,1750,110,180,55,18
2026-01-17,2010,136,190,65,23
`

const badCSV = `Date,Meal,Calories,Fat (g),Protein (g),Carbs (g),Fiber (g)
2026-01-15,Breakfast,450,12,35,48,6
2026-01-15,Lunch,abc,22,45,65,8
2026-01-15,Dinner,720,28,40,70,5
`

func newExtractNutritionUseCase() (
	*ExtractNutritionUseCase,
	*repository.InMemoryArtifactRepository,
	*repository.InMemoryFileStorage,
	*repository.InMemoryNutritionRepository,
	*repository.InMemoryAuditRepository,
) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	nutritionRepo := repository.NewInMemoryNutritionRepository()
	auditRepo := repository.NewInMemoryAuditRepository()

	uc := NewExtractNutritionUseCase(artifactRepo, fileStorage, nutritionRepo, auditRepo)
	return uc, artifactRepo, fileStorage, nutritionRepo, auditRepo
}

func createUploadedCSVArtifact(t *testing.T, repo *repository.InMemoryArtifactRepository, storage *repository.InMemoryFileStorage, csvContent string) *entities.Artifact {
	t.Helper()
	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "nutrition.csv",
		FileType:    "text/csv",
		FileSize:    int64(len(csvContent)),
		StorageKey:  "client-1/document/artifact-1.csv",
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

func TestExtractNutrition_HappyPath_MFP_StoresRecordsAndSummary(t *testing.T) {
	uc, artifactRepo, fileStorage, nutritionRepo, auditRepo := newExtractNutritionUseCase()
	artifact := createUploadedCSVArtifact(t, artifactRepo, fileStorage, mfpCSV)

	output, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.DaysProcessed != 2 {
		t.Errorf("expected 2 days processed, got %d", output.DaysProcessed)
	}
	if output.SkippedRows != 0 {
		t.Errorf("expected 0 skipped rows, got %d", output.SkippedRows)
	}

	// 2 days * 5 metrics = 10 records
	records := nutritionRepo.GetRecords()
	if len(records) != 10 {
		t.Errorf("expected 10 nutrition records, got %d", len(records))
	}

	// Summary should be stored
	summaries := nutritionRepo.GetSummaries()
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].TotalDays != 2 {
		t.Errorf("expected total_days 2, got %d", summaries[0].TotalDays)
	}

	// Audit event should be logged
	events := auditRepo.GetEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(events))
	}
	if events[0].Action != "nutrition.imported" {
		t.Errorf("expected action 'nutrition.imported', got '%s'", events[0].Action)
	}
}

func TestExtractNutrition_GenericCSV_ParsesCorrectly(t *testing.T) {
	uc, artifactRepo, fileStorage, nutritionRepo, _ := newExtractNutritionUseCase()
	artifact := createUploadedCSVArtifact(t, artifactRepo, fileStorage, genericCSV)

	output, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.DaysProcessed != 3 {
		t.Errorf("expected 3 days processed, got %d", output.DaysProcessed)
	}

	// 3 days * 5 metrics = 15 records
	records := nutritionRepo.GetRecords()
	if len(records) != 15 {
		t.Errorf("expected 15 nutrition records, got %d", len(records))
	}
}

func TestExtractNutrition_BadValues_SkipsInvalidRows(t *testing.T) {
	uc, artifactRepo, fileStorage, nutritionRepo, _ := newExtractNutritionUseCase()
	artifact := createUploadedCSVArtifact(t, artifactRepo, fileStorage, badCSV)

	output, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.SkippedRows != 1 {
		t.Errorf("expected 1 skipped row, got %d", output.SkippedRows)
	}

	// 1 day with 2 valid rows * 5 metrics = 5 records
	records := nutritionRepo.GetRecords()
	if len(records) != 5 {
		t.Errorf("expected 5 nutrition records, got %d", len(records))
	}
}

func TestExtractNutrition_EmptyArtifactID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _ := newExtractNutritionUseCase()

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: "",
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
	uc, _, _, _, _ := newExtractNutritionUseCase()

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: "art-1",
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
	uc, _, _, _, _ := newExtractNutritionUseCase()

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: "nonexistent",
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractNutrition_WrongCoach_ReturnsAuthError(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newExtractNutritionUseCase()
	artifact := createUploadedCSVArtifact(t, artifactRepo, fileStorage, genericCSV)

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		CoachID:    "different-coach",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsAuthenticationError(err) {
		t.Errorf("expected AuthenticationError, got %T", err)
	}
}

func TestExtractNutrition_PendingArtifact_ReturnsValidationError(t *testing.T) {
	uc, artifactRepo, _, _, _ := newExtractNutritionUseCase()

	// Create a pending artifact (not uploaded)
	art := &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "nutrition.csv",
		FileType:    "text/csv",
		FileSize:    100,
		StorageKey:  "client-1/document/art-1.csv",
		Type:        entities.ArtifactTypeDocument,
		Status:      entities.ArtifactStatusPending,
		ContentType: "text/csv",
	}
	created, _ := artifactRepo.Create(context.Background(), art)

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: created.ID,
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractNutrition_SingleDay_AveragesEqualTotals(t *testing.T) {
	singleDayCSV := "date,calories,protein,carbs,fat,fiber\n2026-01-15,2050,132,203,70,21\n"
	uc, artifactRepo, fileStorage, nutritionRepo, _ := newExtractNutritionUseCase()
	artifact := createUploadedCSVArtifact(t, artifactRepo, fileStorage, singleDayCSV)

	output, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.DaysProcessed != 1 {
		t.Errorf("expected 1 day processed, got %d", output.DaysProcessed)
	}

	summaries := nutritionRepo.GetSummaries()
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}

	s := summaries[0]
	if s.AvgCalories7d != 2050 {
		t.Errorf("avg_calories_7d: got %f, want 2050", s.AvgCalories7d)
	}
	if s.AvgCalories14d != 2050 {
		t.Errorf("avg_calories_14d: got %f, want 2050", s.AvgCalories14d)
	}
	if s.AvgCalories30d != 2050 {
		t.Errorf("avg_calories_30d: got %f, want 2050", s.AvgCalories30d)
	}
}

func TestExtractNutrition_OutputLinksToArtifactID(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newExtractNutritionUseCase()
	artifact := createUploadedCSVArtifact(t, artifactRepo, fileStorage, genericCSV)

	output, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.ArtifactID != artifact.ID {
		t.Errorf("output artifact_id: got %s, want %s", output.ArtifactID, artifact.ID)
	}
}

func TestExtractNutrition_AuditEventMetadata_IncludesDetails(t *testing.T) {
	uc, artifactRepo, fileStorage, _, auditRepo := newExtractNutritionUseCase()
	artifact := createUploadedCSVArtifact(t, artifactRepo, fileStorage, genericCSV)

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := auditRepo.GetEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(events))
	}
	event := events[0]
	if event.EntityType != "artifact" {
		t.Errorf("expected entity_type 'artifact', got '%s'", event.EntityType)
	}
	if event.EntityID != artifact.ID {
		t.Errorf("expected entity_id '%s', got '%s'", artifact.ID, event.EntityID)
	}
	if event.Metadata["days_processed"] == nil {
		t.Error("expected days_processed in metadata")
	}
}
