package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// --- Mock implementations for testing ---

type mockMeasurementRepository struct {
	createFunc        func(ctx context.Context, m *entities.Measurement) (*entities.Measurement, error)
	batchCreateFunc   func(ctx context.Context, measurements []*entities.Measurement) ([]*entities.Measurement, error)
	findByArtifactFunc func(ctx context.Context, artifactID string) ([]*entities.Measurement, error)
	created           []*entities.Measurement
}

func (m *mockMeasurementRepository) Create(ctx context.Context, measurement *entities.Measurement) (*entities.Measurement, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, measurement)
	}
	measurement.ID = fmt.Sprintf("meas-%d", len(m.created)+1)
	measurement.CreatedAt = time.Now()
	m.created = append(m.created, measurement)
	return measurement, nil
}

func (m *mockMeasurementRepository) BatchCreate(ctx context.Context, measurements []*entities.Measurement) ([]*entities.Measurement, error) {
	if m.batchCreateFunc != nil {
		return m.batchCreateFunc(ctx, measurements)
	}
	var results []*entities.Measurement
	for _, meas := range measurements {
		result, err := m.Create(ctx, meas)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (m *mockMeasurementRepository) FindByArtifactID(ctx context.Context, artifactID string) ([]*entities.Measurement, error) {
	if m.findByArtifactFunc != nil {
		return m.findByArtifactFunc(ctx, artifactID)
	}
	return nil, nil
}

type mockArtifactRepoForLab struct {
	artifacts map[string]*entities.Artifact
	updateStatusFunc func(ctx context.Context, id string, status entities.ArtifactStatus) (*entities.Artifact, error)
}

func (m *mockArtifactRepoForLab) Create(_ context.Context, a *entities.Artifact) (*entities.Artifact, error) {
	return a, nil
}

func (m *mockArtifactRepoForLab) FindByID(_ context.Context, id string) (*entities.Artifact, error) {
	if a, ok := m.artifacts[id]; ok {
		return a, nil
	}
	return nil, nil
}

func (m *mockArtifactRepoForLab) UpdateStatus(ctx context.Context, id string, status entities.ArtifactStatus) (*entities.Artifact, error) {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, id, status)
	}
	if a, ok := m.artifacts[id]; ok {
		a.Status = status
		return a, nil
	}
	return nil, fmt.Errorf("artifact not found: %s", id)
}

type mockFileStorageForLab struct {
	files map[string][]byte
}

func (m *mockFileStorageForLab) GenerateUploadURL(_ context.Context, _ string, _ string, _ time.Duration) (interface{}, error) {
	return nil, nil
}

func (m *mockFileStorageForLab) GenerateDownloadURL(_ context.Context, _ string, _ time.Duration) (interface{}, error) {
	return nil, nil
}

func (m *mockFileStorageForLab) ObjectExists(_ context.Context, key string) (bool, error) {
	_, ok := m.files[key]
	return ok, nil
}

func (m *mockFileStorageForLab) GetObject(_ context.Context, key string) ([]byte, error) {
	data, ok := m.files[key]
	if !ok {
		return nil, fmt.Errorf("object not found: %s", key)
	}
	return data, nil
}

type mockAuditRepoForLab struct {
	events []*entities.AuditEvent
}

func (m *mockAuditRepoForLab) LogEvent(_ context.Context, event *entities.AuditEvent) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockAuditRepoForLab) Query(_ context.Context, _ entities.AuditQueryFilter) ([]*entities.AuditEvent, int, error) {
	return m.events, len(m.events), nil
}

// --- Tests ---

func TestExtractLabResults_HappyPath_CSVWith10Markers(t *testing.T) {
	csvData := []byte(`Test Name,Result,Units,Reference Range
"Glucose, Fasting",98,mg/dL,70-100
Total Cholesterol,210,mg/dL,<200
LDL Cholesterol,142,mg/dL,<100
HDL Cholesterol,55,mg/dL,>40
Triglycerides,120,mg/dL,<150
Hemoglobin A1c,5.4,%,<5.7
"Testosterone, Total",650,ng/dL,300-1000
TSH,2.5,mIU/L,0.4-4.0
"Vitamin D, 25-Hydroxy",35,ng/mL,30-100
"Iron, Serum",80,mcg/dL,60-170
`)

	artifactID := "art-1"
	artifact := &entities.Artifact{
		ID:       artifactID,
		ClientID: "client-1",
		CoachID:  "coach-1",
		FileName: "bloodwork.csv",
		Status:   entities.ArtifactStatusUploaded,
		StorageKey: "client-1/document/art-1.csv",
	}

	artifactRepo := &mockArtifactRepoForLab{artifacts: map[string]*entities.Artifact{artifactID: artifact}}
	measurementRepo := &mockMeasurementRepository{}
	fileStorage := &mockFileStorageForLab{files: map[string][]byte{artifact.StorageKey: csvData}}
	auditRepo := &mockAuditRepoForLab{}

	uc := NewExtractLabResultsUseCase(artifactRepo, measurementRepo, fileStorage, auditRepo)

	input := ExtractLabResultsInput{
		ArtifactID: artifactID,
		CoachID:    "coach-1",
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.ExtractedCount != 10 {
		t.Errorf("expected 10 extracted markers, got %d", output.ExtractedCount)
	}
	if output.ArtifactID != artifactID {
		t.Errorf("expected artifact ID %s, got %s", artifactID, output.ArtifactID)
	}
	if len(measurementRepo.created) != 10 {
		t.Errorf("expected 10 measurements stored, got %d", len(measurementRepo.created))
	}

	// Verify audit event logged
	if len(auditRepo.events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(auditRepo.events))
	}
	if auditRepo.events[0].Action != "lab.extract" {
		t.Errorf("expected audit action 'lab.extract', got %s", auditRepo.events[0].Action)
	}
}

func TestExtractLabResults_OutOfRange_FlagsCorrectly(t *testing.T) {
	csvData := []byte(`Test Name,Result,Units,Reference Range
LDL Cholesterol,142,mg/dL,<100
`)

	artifactID := "art-2"
	artifact := &entities.Artifact{
		ID:       artifactID,
		ClientID: "client-1",
		CoachID:  "coach-1",
		FileName: "bloodwork.csv",
		Status:   entities.ArtifactStatusUploaded,
		StorageKey: "client-1/document/art-2.csv",
	}

	artifactRepo := &mockArtifactRepoForLab{artifacts: map[string]*entities.Artifact{artifactID: artifact}}
	measurementRepo := &mockMeasurementRepository{}
	fileStorage := &mockFileStorageForLab{files: map[string][]byte{artifact.StorageKey: csvData}}
	auditRepo := &mockAuditRepoForLab{}

	uc := NewExtractLabResultsUseCase(artifactRepo, measurementRepo, fileStorage, auditRepo)

	output, err := uc.Execute(context.Background(), ExtractLabResultsInput{
		ArtifactID: artifactID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.FlaggedCount != 1 {
		t.Errorf("expected 1 flagged marker, got %d", output.FlaggedCount)
	}

	ldl := measurementRepo.created[0]
	if ldl.Flag != entities.FlagHigh {
		t.Errorf("LDL flag: expected high, got %s", ldl.Flag)
	}
	high := 100.0
	if ldl.ReferenceHigh == nil || *ldl.ReferenceHigh != high {
		t.Errorf("LDL reference high: expected 100, got %v", ldl.ReferenceHigh)
	}
}

func TestExtractLabResults_MissingReference_FlagEmpty(t *testing.T) {
	csvData := []byte(`Test Name,Result,Units,Reference Range
LDL Cholesterol,142,mg/dL,
`)

	artifactID := "art-3"
	artifact := &entities.Artifact{
		ID:       artifactID,
		ClientID: "client-1",
		CoachID:  "coach-1",
		FileName: "bloodwork.csv",
		Status:   entities.ArtifactStatusUploaded,
		StorageKey: "client-1/document/art-3.csv",
	}

	artifactRepo := &mockArtifactRepoForLab{artifacts: map[string]*entities.Artifact{artifactID: artifact}}
	measurementRepo := &mockMeasurementRepository{}
	fileStorage := &mockFileStorageForLab{files: map[string][]byte{artifact.StorageKey: csvData}}
	auditRepo := &mockAuditRepoForLab{}

	uc := NewExtractLabResultsUseCase(artifactRepo, measurementRepo, fileStorage, auditRepo)

	output, err := uc.Execute(context.Background(), ExtractLabResultsInput{
		ArtifactID: artifactID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.FlaggedCount != 0 {
		t.Errorf("expected 0 flagged markers (no reference), got %d", output.FlaggedCount)
	}

	ldl := measurementRepo.created[0]
	if ldl.Flag != "" {
		t.Errorf("LDL flag: expected empty, got %s", ldl.Flag)
	}
}

func TestExtractLabResults_ArtifactNotFound_ReturnsError(t *testing.T) {
	artifactRepo := &mockArtifactRepoForLab{artifacts: map[string]*entities.Artifact{}}
	measurementRepo := &mockMeasurementRepository{}
	fileStorage := &mockFileStorageForLab{files: map[string][]byte{}}
	auditRepo := &mockAuditRepoForLab{}

	uc := NewExtractLabResultsUseCase(artifactRepo, measurementRepo, fileStorage, auditRepo)

	_, err := uc.Execute(context.Background(), ExtractLabResultsInput{
		ArtifactID: "nonexistent",
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent artifact")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestExtractLabResults_EmptyArtifactID_ReturnsError(t *testing.T) {
	uc := NewExtractLabResultsUseCase(
		&mockArtifactRepoForLab{artifacts: map[string]*entities.Artifact{}},
		&mockMeasurementRepository{},
		&mockFileStorageForLab{files: map[string][]byte{}},
		&mockAuditRepoForLab{},
	)

	_, err := uc.Execute(context.Background(), ExtractLabResultsInput{
		ArtifactID: "",
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error for empty artifact ID")
	}
}

func TestExtractLabResults_UnparseableFormat_ReturnsError(t *testing.T) {
	badData := []byte("This is not CSV or PDF data at all")
	artifactID := "art-bad"
	artifact := &entities.Artifact{
		ID:       artifactID,
		ClientID: "client-1",
		CoachID:  "coach-1",
		FileName: "bloodwork.csv",
		Status:   entities.ArtifactStatusUploaded,
		StorageKey: "client-1/document/art-bad.csv",
	}

	artifactRepo := &mockArtifactRepoForLab{artifacts: map[string]*entities.Artifact{artifactID: artifact}}
	measurementRepo := &mockMeasurementRepository{}
	fileStorage := &mockFileStorageForLab{files: map[string][]byte{artifact.StorageKey: badData}}
	auditRepo := &mockAuditRepoForLab{}

	uc := NewExtractLabResultsUseCase(artifactRepo, measurementRepo, fileStorage, auditRepo)

	_, err := uc.Execute(context.Background(), ExtractLabResultsInput{
		ArtifactID: artifactID,
		CoachID:    "coach-1",
	})
	if err == nil {
		t.Fatal("expected error for unparseable format")
	}
}

func TestExtractLabResults_DuplicateMarkers_StoresBoth(t *testing.T) {
	csvData := []byte(`Test Name,Result,Units,Reference Range
"Glucose, Fasting",98,mg/dL,70-100
"Glucose, Fasting",102,mg/dL,70-100
`)

	artifactID := "art-dup"
	artifact := &entities.Artifact{
		ID:       artifactID,
		ClientID: "client-1",
		CoachID:  "coach-1",
		FileName: "bloodwork.csv",
		Status:   entities.ArtifactStatusUploaded,
		StorageKey: "client-1/document/art-dup.csv",
	}

	artifactRepo := &mockArtifactRepoForLab{artifacts: map[string]*entities.Artifact{artifactID: artifact}}
	measurementRepo := &mockMeasurementRepository{}
	fileStorage := &mockFileStorageForLab{files: map[string][]byte{artifact.StorageKey: csvData}}
	auditRepo := &mockAuditRepoForLab{}

	uc := NewExtractLabResultsUseCase(artifactRepo, measurementRepo, fileStorage, auditRepo)

	output, err := uc.Execute(context.Background(), ExtractLabResultsInput{
		ArtifactID: artifactID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.ExtractedCount != 2 {
		t.Errorf("expected 2 extracted markers (duplicates), got %d", output.ExtractedCount)
	}
	if len(measurementRepo.created) != 2 {
		t.Errorf("expected 2 measurements stored, got %d", len(measurementRepo.created))
	}
}

func TestExtractLabResults_LinksMeasurementsToArtifact(t *testing.T) {
	csvData := []byte(`Test Name,Result,Units,Reference Range
Glucose,95,mg/dL,70-100
`)

	artifactID := "art-link"
	artifact := &entities.Artifact{
		ID:       artifactID,
		ClientID: "client-1",
		CoachID:  "coach-1",
		FileName: "bloodwork.csv",
		Status:   entities.ArtifactStatusUploaded,
		StorageKey: "client-1/document/art-link.csv",
	}

	artifactRepo := &mockArtifactRepoForLab{artifacts: map[string]*entities.Artifact{artifactID: artifact}}
	measurementRepo := &mockMeasurementRepository{}
	fileStorage := &mockFileStorageForLab{files: map[string][]byte{artifact.StorageKey: csvData}}
	auditRepo := &mockAuditRepoForLab{}

	uc := NewExtractLabResultsUseCase(artifactRepo, measurementRepo, fileStorage, auditRepo)

	_, err := uc.Execute(context.Background(), ExtractLabResultsInput{
		ArtifactID: artifactID,
		CoachID:    "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := measurementRepo.created[0]
	if m.ArtifactID == nil || *m.ArtifactID != artifactID {
		t.Errorf("expected measurement linked to artifact %s, got %v", artifactID, m.ArtifactID)
	}
	if m.ClientID != "client-1" {
		t.Errorf("expected client_id 'client-1', got %s", m.ClientID)
	}
	if m.RecordedBy != "coach-1" {
		t.Errorf("expected recorded_by 'coach-1', got %s", m.RecordedBy)
	}
}
