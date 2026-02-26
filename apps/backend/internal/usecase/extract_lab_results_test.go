package usecase

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
	domainrepo "github.com/xenios/backend/internal/domain/repository"
	"github.com/xenios/backend/internal/infrastructure/parser"
)

func newExtractLabResultsUseCase() (
	*ExtractLabResultsUseCase,
	*repository.InMemoryArtifactRepository,
	*repository.InMemoryMeasurementRepository,
	*repository.InMemoryFileContentReader,
	*repository.InMemoryAuditRepository,
) {
	artifactRepo := repository.NewInMemoryArtifactRepository()
	measurementRepo := repository.NewInMemoryMeasurementRepository()
	fileReader := repository.NewInMemoryFileContentReader()
	auditRepo := repository.NewInMemoryAuditRepository()
	csvParser := parser.NewCSVLabParser()
	pdfParser := parser.NewPDFLabParser()

	uc := NewExtractLabResultsUseCase(artifactRepo, measurementRepo, fileReader, auditRepo, csvParser, pdfParser)
	return uc, artifactRepo, measurementRepo, fileReader, auditRepo
}

func createTestArtifact(t *testing.T, repo *repository.InMemoryArtifactRepository, fileName, storageKey string) *entities.Artifact {
	t.Helper()
	art, err := repo.Create(context.Background(), &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    fileName,
		FileType:    "text/csv",
		FileSize:    1024,
		StorageKey:  storageKey,
		Type:        entities.ArtifactTypeDocument,
		Status:      entities.ArtifactStatusUploaded,
		ContentType: "text/csv",
	})
	if err != nil {
		t.Fatalf("create artifact: %v", err)
	}
	return art
}

func makePayload(t *testing.T, artifactID, clientID, coachID string) json.RawMessage {
	t.Helper()
	p := ExtractLabResultsPayload{
		ArtifactID: artifactID,
		ClientID:   clientID,
		CoachID:    coachID,
	}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return data
}

// Verify interface compliance
var _ domainrepo.MeasurementRepository = &repository.InMemoryMeasurementRepository{}
var _ domainrepo.FileContentReader = &repository.InMemoryFileContentReader{}

func TestExtractLabResults_HappyPath_CSVWith10Markers(t *testing.T) {
	uc, artifactRepo, measurementRepo, fileReader, _ := newExtractLabResultsUseCase()

	csvContent := `Test Name,Result,Units,Reference Range
Glucose,98,mg/dL,70-100
LDL Cholesterol,142,mg/dL,<100
HDL Cholesterol,55,mg/dL,>40
Triglycerides,120,mg/dL,<150
HbA1c,5.4,%,< 5.7
Testosterone,650,ng/dL,300-1000
TSH,2.1,mIU/L,0.4-4.0
Vitamin D,45,ng/mL,30-100
Iron,85,mcg/dL,60-170
Total Cholesterol,210,mg/dL,<200
`
	art := createTestArtifact(t, artifactRepo, "lab_results.csv", "client-1/document/test.csv")
	fileReader.PutContent("client-1/document/test.csv", []byte(csvContent))

	job := &entities.Job{
		ID:      "job-1",
		Type:    entities.JobTypeExtractLabResults,
		Payload: makePayload(t, art.ID, "client-1", "coach-1"),
	}

	err := uc.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	measurements := measurementRepo.GetAll()
	if len(measurements) != 10 {
		t.Fatalf("expected 10 measurements, got %d", len(measurements))
	}
}

func TestExtractLabResults_OutOfRange_LDLFlaggedHigh(t *testing.T) {
	uc, artifactRepo, measurementRepo, fileReader, _ := newExtractLabResultsUseCase()

	csvContent := `Test Name,Result,Units,Reference Range
LDL Cholesterol,142,mg/dL,<100
`
	art := createTestArtifact(t, artifactRepo, "lab.csv", "client-1/document/ldl.csv")
	fileReader.PutContent("client-1/document/ldl.csv", []byte(csvContent))

	job := &entities.Job{
		ID:      "job-2",
		Type:    entities.JobTypeExtractLabResults,
		Payload: makePayload(t, art.ID, "client-1", "coach-1"),
	}

	err := uc.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	measurements := measurementRepo.GetAll()
	if len(measurements) != 1 {
		t.Fatalf("expected 1 measurement, got %d", len(measurements))
	}
	if measurements[0].Flag == nil || *measurements[0].Flag != entities.FlagHigh {
		t.Errorf("expected flag 'high', got %v", measurements[0].Flag)
	}
	if measurements[0].MeasurementType != "ldl_cholesterol" {
		t.Errorf("expected type 'ldl_cholesterol', got %q", measurements[0].MeasurementType)
	}
}

func TestExtractLabResults_NormalValue_FlaggedNormal(t *testing.T) {
	uc, artifactRepo, measurementRepo, fileReader, _ := newExtractLabResultsUseCase()

	csvContent := `Test Name,Result,Units,Reference Range
Glucose,85,mg/dL,70-100
`
	art := createTestArtifact(t, artifactRepo, "lab.csv", "client-1/document/glucose.csv")
	fileReader.PutContent("client-1/document/glucose.csv", []byte(csvContent))

	job := &entities.Job{
		ID:      "job-3",
		Type:    entities.JobTypeExtractLabResults,
		Payload: makePayload(t, art.ID, "client-1", "coach-1"),
	}

	err := uc.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	measurements := measurementRepo.GetAll()
	if measurements[0].Flag == nil || *measurements[0].Flag != entities.FlagNormal {
		t.Errorf("expected flag 'normal', got %v", measurements[0].Flag)
	}
}

func TestExtractLabResults_MissingReference_FlagNil(t *testing.T) {
	uc, artifactRepo, measurementRepo, fileReader, _ := newExtractLabResultsUseCase()

	csvContent := `Test Name,Result,Units,Reference Range
Glucose,98,mg/dL,
`
	art := createTestArtifact(t, artifactRepo, "lab.csv", "client-1/document/noref.csv")
	fileReader.PutContent("client-1/document/noref.csv", []byte(csvContent))

	job := &entities.Job{
		ID:      "job-4",
		Type:    entities.JobTypeExtractLabResults,
		Payload: makePayload(t, art.ID, "client-1", "coach-1"),
	}

	err := uc.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	measurements := measurementRepo.GetAll()
	if measurements[0].Flag != nil {
		t.Errorf("expected nil flag when no reference range, got %v", *measurements[0].Flag)
	}
}

func TestExtractLabResults_PDFExtraction_ExtractsMarkers(t *testing.T) {
	uc, artifactRepo, measurementRepo, fileReader, _ := newExtractLabResultsUseCase()

	pdfText := "Glucose\t98\tmg/dL\t70-100\nLDL Cholesterol\t142\tmg/dL\t<100\n"

	art, err := artifactRepo.Create(context.Background(), &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "lab_results.pdf",
		FileType:    "application/pdf",
		FileSize:    2048,
		StorageKey:  "client-1/document/lab.pdf",
		Type:        entities.ArtifactTypeDocument,
		Status:      entities.ArtifactStatusUploaded,
		ContentType: "application/pdf",
	})
	if err != nil {
		t.Fatalf("create artifact: %v", err)
	}
	fileReader.PutContent("client-1/document/lab.pdf", []byte(pdfText))

	job := &entities.Job{
		ID:      "job-5",
		Type:    entities.JobTypeExtractLabResults,
		Payload: makePayload(t, art.ID, "client-1", "coach-1"),
	}

	err = uc.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	measurements := measurementRepo.GetAll()
	if len(measurements) != 2 {
		t.Fatalf("expected 2 measurements from PDF, got %d", len(measurements))
	}
}

func TestExtractLabResults_InvalidPayload_ReturnsError(t *testing.T) {
	uc, _, _, _, _ := newExtractLabResultsUseCase()

	job := &entities.Job{
		ID:      "job-6",
		Type:    entities.JobTypeExtractLabResults,
		Payload: json.RawMessage(`{invalid json`),
	}

	err := uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for invalid JSON payload")
	}
}

func TestExtractLabResults_MissingArtifactID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _ := newExtractLabResultsUseCase()

	job := &entities.Job{
		ID:      "job-7",
		Type:    entities.JobTypeExtractLabResults,
		Payload: makePayload(t, "", "client-1", "coach-1"),
	}

	err := uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for missing artifact_id")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractLabResults_MissingClientID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _ := newExtractLabResultsUseCase()

	job := &entities.Job{
		ID:      "job-8",
		Type:    entities.JobTypeExtractLabResults,
		Payload: makePayload(t, "art-1", "", "coach-1"),
	}

	err := uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for missing client_id")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractLabResults_MissingCoachID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _ := newExtractLabResultsUseCase()

	job := &entities.Job{
		ID:      "job-9",
		Type:    entities.JobTypeExtractLabResults,
		Payload: makePayload(t, "art-1", "client-1", ""),
	}

	err := uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for missing coach_id")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractLabResults_ArtifactNotFound_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _ := newExtractLabResultsUseCase()

	job := &entities.Job{
		ID:      "job-10",
		Type:    entities.JobTypeExtractLabResults,
		Payload: makePayload(t, "nonexistent-id", "client-1", "coach-1"),
	}

	err := uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for nonexistent artifact")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractLabResults_UnsupportedFormat_ReturnsValidationError(t *testing.T) {
	uc, artifactRepo, _, fileReader, _ := newExtractLabResultsUseCase()

	art, err := artifactRepo.Create(context.Background(), &entities.Artifact{
		ClientID:    "client-1",
		CoachID:     "coach-1",
		FileName:    "lab_results.json",
		FileType:    "application/json",
		FileSize:    1024,
		StorageKey:  "client-1/document/lab.json",
		Type:        entities.ArtifactTypeDocument,
		Status:      entities.ArtifactStatusUploaded,
		ContentType: "application/json",
	})
	if err != nil {
		t.Fatalf("create artifact: %v", err)
	}
	fileReader.PutContent("client-1/document/lab.json", []byte(`{"test": true}`))

	job := &entities.Job{
		ID:      "job-11",
		Type:    entities.JobTypeExtractLabResults,
		Payload: makePayload(t, art.ID, "client-1", "coach-1"),
	}

	err = uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestExtractLabResults_UnparseableCSV_ReturnsError(t *testing.T) {
	uc, artifactRepo, _, fileReader, _ := newExtractLabResultsUseCase()

	art := createTestArtifact(t, artifactRepo, "bad.csv", "client-1/document/bad.csv")
	// Content with no valid markers (non-numeric values)
	fileReader.PutContent("client-1/document/bad.csv", []byte("Name,Value\ntest,not_a_number\n"))

	job := &entities.Job{
		ID:      "job-12",
		Type:    entities.JobTypeExtractLabResults,
		Payload: makePayload(t, art.ID, "client-1", "coach-1"),
	}

	err := uc.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for unparseable CSV")
	}
}

func TestExtractLabResults_AuditEventLogged(t *testing.T) {
	uc, artifactRepo, _, fileReader, auditRepo := newExtractLabResultsUseCase()

	csvContent := `Test Name,Result,Units,Reference Range
Glucose,98,mg/dL,70-100
`
	art := createTestArtifact(t, artifactRepo, "lab.csv", "client-1/document/audit.csv")
	fileReader.PutContent("client-1/document/audit.csv", []byte(csvContent))

	job := &entities.Job{
		ID:      "job-13",
		Type:    entities.JobTypeExtractLabResults,
		Payload: makePayload(t, art.ID, "client-1", "coach-1"),
	}

	err := uc.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := auditRepo.GetEvents()
	if len(events) == 0 {
		t.Fatal("expected audit event to be logged")
	}
	if events[0].Action != "lab.extract" {
		t.Errorf("expected action 'lab.extract', got %q", events[0].Action)
	}
	if events[0].EntityType != "artifact" {
		t.Errorf("expected entity_type 'artifact', got %q", events[0].EntityType)
	}
}

func TestExtractLabResults_LinksToArtifact(t *testing.T) {
	uc, artifactRepo, measurementRepo, fileReader, _ := newExtractLabResultsUseCase()

	csvContent := `Test Name,Result,Units,Reference Range
Glucose,98,mg/dL,70-100
`
	art := createTestArtifact(t, artifactRepo, "lab.csv", "client-1/document/linked.csv")
	fileReader.PutContent("client-1/document/linked.csv", []byte(csvContent))

	job := &entities.Job{
		ID:      "job-14",
		Type:    entities.JobTypeExtractLabResults,
		Payload: makePayload(t, art.ID, "client-1", "coach-1"),
	}

	err := uc.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	measurements := measurementRepo.GetAll()
	if measurements[0].ArtifactID == nil || *measurements[0].ArtifactID != art.ID {
		t.Errorf("expected measurement linked to artifact %q", art.ID)
	}
}

func TestExtractLabResults_DuplicateMarkers_BothStored(t *testing.T) {
	uc, artifactRepo, measurementRepo, fileReader, _ := newExtractLabResultsUseCase()

	csvContent := `Test Name,Result,Units,Reference Range
Glucose,98,mg/dL,70-100
Glucose,102,mg/dL,70-100
`
	art := createTestArtifact(t, artifactRepo, "lab.csv", "client-1/document/dup.csv")
	fileReader.PutContent("client-1/document/dup.csv", []byte(csvContent))

	job := &entities.Job{
		ID:      "job-15",
		Type:    entities.JobTypeExtractLabResults,
		Payload: makePayload(t, art.ID, "client-1", "coach-1"),
	}

	err := uc.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	measurements := measurementRepo.GetAll()
	if len(measurements) != 2 {
		t.Fatalf("expected 2 measurements for duplicate markers, got %d", len(measurements))
	}
}

func TestExtractLabResults_LowValue_FlaggedLow(t *testing.T) {
	uc, artifactRepo, measurementRepo, fileReader, _ := newExtractLabResultsUseCase()

	csvContent := `Test Name,Result,Units,Reference Range
Glucose,55,mg/dL,70-100
`
	art := createTestArtifact(t, artifactRepo, "lab.csv", "client-1/document/low.csv")
	fileReader.PutContent("client-1/document/low.csv", []byte(csvContent))

	job := &entities.Job{
		ID:      "job-16",
		Type:    entities.JobTypeExtractLabResults,
		Payload: makePayload(t, art.ID, "client-1", "coach-1"),
	}

	err := uc.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	measurements := measurementRepo.GetAll()
	if measurements[0].Flag == nil || *measurements[0].Flag != entities.FlagLow {
		t.Errorf("expected flag 'low', got %v", measurements[0].Flag)
	}
}

func TestExtractLabResults_StoresReferenceRange(t *testing.T) {
	uc, artifactRepo, measurementRepo, fileReader, _ := newExtractLabResultsUseCase()

	csvContent := `Test Name,Result,Units,Reference Range
TSH,2.1,mIU/L,0.4-4.0
`
	art := createTestArtifact(t, artifactRepo, "lab.csv", "client-1/document/ref.csv")
	fileReader.PutContent("client-1/document/ref.csv", []byte(csvContent))

	job := &entities.Job{
		ID:      "job-17",
		Type:    entities.JobTypeExtractLabResults,
		Payload: makePayload(t, art.ID, "client-1", "coach-1"),
	}

	err := uc.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	measurements := measurementRepo.GetAll()
	m := measurements[0]
	if m.ReferenceLow == nil || *m.ReferenceLow != 0.4 {
		t.Errorf("expected reference_low=0.4, got %v", m.ReferenceLow)
	}
	if m.ReferenceHigh == nil || *m.ReferenceHigh != 4.0 {
		t.Errorf("expected reference_high=4.0, got %v", m.ReferenceHigh)
	}
}

func TestExtractLabResults_NormalizesMarkerNames(t *testing.T) {
	uc, artifactRepo, measurementRepo, fileReader, _ := newExtractLabResultsUseCase()

	csvContent := `Test Name,Result,Units,Reference Range
Fasting Glucose,98,mg/dL,70-100
LDL-C,142,mg/dL,<100
Hemoglobin A1C,5.4,%,< 5.7
`
	art := createTestArtifact(t, artifactRepo, "lab.csv", "client-1/document/norm.csv")
	fileReader.PutContent("client-1/document/norm.csv", []byte(csvContent))

	job := &entities.Job{
		ID:      "job-18",
		Type:    entities.JobTypeExtractLabResults,
		Payload: makePayload(t, art.ID, "client-1", "coach-1"),
	}

	err := uc.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	measurements := measurementRepo.GetAll()
	expectedTypes := []string{"fasting_glucose", "ldl_cholesterol", "hba1c"}
	for i, m := range measurements {
		if m.MeasurementType != expectedTypes[i] {
			t.Errorf("measurement %d: expected type %q, got %q", i, expectedTypes[i], m.MeasurementType)
		}
	}
}
