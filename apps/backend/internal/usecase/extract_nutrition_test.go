package usecase

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/infrastructure/nutrition"
)

func nutritionTestdataPath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "infrastructure", "nutrition", "testdata", name)
}

func readNutritionTestFile(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(nutritionTestdataPath(name))
	if err != nil {
		t.Fatalf("failed to read test file %s: %v", name, err)
	}
	return data
}

func newExtractNutritionTestDeps() (
	*ExtractNutritionUseCase,
	*repository.InMemoryArtifactRepository,
	*repository.InMemoryFileStorage,
	*repository.InMemoryMeasurementRepository,
	*repository.InMemoryAuditRepository,
) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	fileStorage := repository.NewInMemoryFileStorage()
	measurementRepo := repository.NewInMemoryMeasurementRepository()
	auditRepo := repository.NewInMemoryAuditRepository()

	parser := nutrition.NewCSVNutritionParser()
	uc := NewExtractNutritionUseCase(artifactRepo, fileStorage, measurementRepo, auditRepo, parser)
	return uc, artifactRepo, fileStorage, measurementRepo, auditRepo
}

func createCSVArtifact(t *testing.T, repo *repository.InMemoryArtifactRepository, storage *repository.InMemoryFileStorage, csvContent []byte) *entities.Artifact {
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
	created, err := repo.Create(context.Background(), art)
	if err != nil {
		t.Fatalf("failed to create test artifact: %v", err)
	}
	storage.PutObjectWithContent(created.StorageKey, csvContent)
	return created
}

func TestExtractNutrition_MFP30Days_StoresDailyMeasurements(t *testing.T) {
	uc, artifactRepo, fileStorage, measurementRepo, _ := newExtractNutritionTestDeps()
	csvData := readNutritionTestFile(t, "myfitnesspal_30days.csv")
	artifact := createCSVArtifact(t, artifactRepo, fileStorage, csvData)

	out, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 30 days * 5 measurement types = 150 measurements
	measurements := measurementRepo.GetAll()
	if len(measurements) != 150 {
		t.Errorf("expected 150 measurements, got %d", len(measurements))
	}

	if out.TotalDays != 30 {
		t.Errorf("expected 30 total days, got %d", out.TotalDays)
	}
	if out.Format != entities.CSVFormatMyFitnessPal {
		t.Errorf("expected format %s, got %s", entities.CSVFormatMyFitnessPal, out.Format)
	}
}

func TestExtractNutrition_SingleDay_AveragesEqualDailyTotals(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newExtractNutritionTestDeps()
	csvData := readNutritionTestFile(t, "single_day.csv")
	artifact := createCSVArtifact(t, artifactRepo, fileStorage, csvData)

	out, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Averages.Avg7Day == nil {
		t.Fatal("expected 7-day average")
	}
	if out.Averages.Avg7Day.Calories != 2050 {
		t.Errorf("expected avg calories 2050, got %.0f", out.Averages.Avg7Day.Calories)
	}
}

func TestExtractNutrition_30Days_ComputesAllAverages(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newExtractNutritionTestDeps()
	csvData := readNutritionTestFile(t, "myfitnesspal_30days.csv")
	artifact := createCSVArtifact(t, artifactRepo, fileStorage, csvData)

	out, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Averages.Avg7Day == nil {
		t.Fatal("expected 7-day average")
	}
	if out.Averages.Avg14Day == nil {
		t.Fatal("expected 14-day average")
	}
	if out.Averages.Avg30Day == nil {
		t.Fatal("expected 30-day average")
	}
	if out.Averages.Avg7Day.Days != 7 {
		t.Errorf("expected 7-day average to cover 7 days, got %d", out.Averages.Avg7Day.Days)
	}
	if out.Averages.Avg30Day.Days != 30 {
		t.Errorf("expected 30-day average to cover 30 days, got %d", out.Averages.Avg30Day.Days)
	}
}

func TestExtractNutrition_GenericCSV_ParsesCorrectly(t *testing.T) {
	uc, artifactRepo, fileStorage, measurementRepo, _ := newExtractNutritionTestDeps()
	csvData := readNutritionTestFile(t, "generic_nutrition.csv")
	artifact := createCSVArtifact(t, artifactRepo, fileStorage, csvData)

	out, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Format != entities.CSVFormatGeneric {
		t.Errorf("expected format %s, got %s", entities.CSVFormatGeneric, out.Format)
	}
	// 3 days * 5 types = 15
	measurements := measurementRepo.GetAll()
	if len(measurements) != 15 {
		t.Errorf("expected 15 measurements, got %d", len(measurements))
	}
}

func TestExtractNutrition_BadValues_SkipsInvalidRows(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newExtractNutritionTestDeps()
	csvData := readNutritionTestFile(t, "bad_values.csv")
	artifact := createCSVArtifact(t, artifactRepo, fileStorage, csvData)

	out, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.SkippedRows != 2 {
		t.Errorf("expected 2 skipped rows, got %d", out.SkippedRows)
	}
}

func TestExtractNutrition_EmptyArtifactID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _ := newExtractNutritionTestDeps()

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
	uc, _, _, _, _ := newExtractNutritionTestDeps()

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
	uc, _, _, _, _ := newExtractNutritionTestDeps()

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
	uc, artifactRepo, fileStorage, _, _ := newExtractNutritionTestDeps()
	csvData := readNutritionTestFile(t, "single_day.csv")
	artifact := createCSVArtifact(t, artifactRepo, fileStorage, csvData)

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

func TestExtractNutrition_AuditEventLogged(t *testing.T) {
	uc, artifactRepo, fileStorage, _, auditRepo := newExtractNutritionTestDeps()
	csvData := readNutritionTestFile(t, "single_day.csv")
	artifact := createCSVArtifact(t, artifactRepo, fileStorage, csvData)

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
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
		if e.Action == "nutrition.import" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'nutrition.import' audit event")
	}
}

func TestExtractNutrition_LinksArtifactID(t *testing.T) {
	uc, artifactRepo, fileStorage, _, _ := newExtractNutritionTestDeps()
	csvData := readNutritionTestFile(t, "single_day.csv")
	artifact := createCSVArtifact(t, artifactRepo, fileStorage, csvData)

	out, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ArtifactID != artifact.ID {
		t.Errorf("expected artifact_id %s, got %s", artifact.ID, out.ArtifactID)
	}
}

func TestExtractNutrition_MeasurementsHaveCorrectClientID(t *testing.T) {
	uc, artifactRepo, fileStorage, measurementRepo, _ := newExtractNutritionTestDeps()
	csvData := readNutritionTestFile(t, "single_day.csv")
	artifact := createCSVArtifact(t, artifactRepo, fileStorage, csvData)

	_, err := uc.Execute(context.Background(), ExtractNutritionInput{
		ArtifactID: artifact.ID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	measurements := measurementRepo.GetAll()
	for _, m := range measurements {
		if m.ClientID != "client-1" {
			t.Errorf("expected client_id 'client-1', got '%s'", m.ClientID)
		}
	}
}
