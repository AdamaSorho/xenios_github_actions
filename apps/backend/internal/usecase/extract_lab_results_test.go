package usecase

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
	domainrepo "github.com/xenios/backend/internal/domain/repository"
)

// stubLabParser implements repository.LabParser for testing.
type stubLabParser struct {
	measurements []entities.LabMeasurement
	err          error
}

func (p *stubLabParser) Parse(_ io.Reader) ([]entities.LabMeasurement, error) {
	if p.err != nil {
		return nil, p.err
	}
	return p.measurements, nil
}

func newExtractTestDeps() (*repository.InMemoryMeasurementRepository, *repository.InMemoryArtifactRepository, *repository.InMemoryAuditRepository) {
	return repository.NewInMemoryMeasurementRepository(),
		repository.NewInMemoryArtifactRepository(),
		repository.NewInMemoryAuditRepository()
}

func TestExtractLabResults_ValidCSV_ExtractsAndStores(t *testing.T) {
	measRepo, artRepo, auditRepo := newExtractTestDeps()
	normalFlag := entities.LabFlagNormal
	highFlag := entities.LabFlagHigh

	parser := &stubLabParser{
		measurements: []entities.LabMeasurement{
			{
				MeasurementType: entities.LabTypeFastingGlucose,
				Value:           98,
				Unit:            "mg/dL",
				ReferenceLow:    entities.FloatPtr(70),
				ReferenceHigh:   entities.FloatPtr(100),
				Flag:            &normalFlag,
			},
			{
				MeasurementType: entities.LabTypeLDLCholesterol,
				Value:           142,
				Unit:            "mg/dL",
				ReferenceHigh:   entities.FloatPtr(100),
				Flag:            &highFlag,
			},
		},
	}

	parsers := map[string]domainrepo.LabParser{
		"text/csv": parser,
	}

	uc := NewExtractLabResultsUseCase(measRepo, artRepo, auditRepo, parsers)

	input := ExtractLabResultsInput{
		ArtifactID:  "art-1",
		CoachID:     "coach-1",
		ClientID:    "client-1",
		Content:     []byte("csv content"),
		ContentType: "text/csv",
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.MeasurementsCount != 2 {
		t.Errorf("count = %d, want 2", output.MeasurementsCount)
	}

	stored := measRepo.All()
	if len(stored) != 2 {
		t.Fatalf("stored %d, want 2", len(stored))
	}

	// Verify first measurement
	if stored[0].MeasurementType != "fasting_glucose" {
		t.Errorf("type = %s, want fasting_glucose", stored[0].MeasurementType)
	}
	if stored[0].ClientID != "client-1" {
		t.Errorf("client_id = %s, want client-1", stored[0].ClientID)
	}
	if stored[0].RecordedBy != "coach-1" {
		t.Errorf("recorded_by = %s, want coach-1", stored[0].RecordedBy)
	}
	if stored[0].ArtifactID == nil || *stored[0].ArtifactID != "art-1" {
		t.Errorf("artifact_id = %v, want art-1", stored[0].ArtifactID)
	}
}

func TestExtractLabResults_OutOfRange_FlaggedCorrectly(t *testing.T) {
	measRepo, artRepo, auditRepo := newExtractTestDeps()
	highFlag := entities.LabFlagHigh

	parser := &stubLabParser{
		measurements: []entities.LabMeasurement{
			{
				MeasurementType: entities.LabTypeLDLCholesterol,
				Value:           142,
				Unit:            "mg/dL",
				ReferenceHigh:   entities.FloatPtr(100),
				Flag:            &highFlag,
			},
		},
	}

	parsers := map[string]domainrepo.LabParser{"text/csv": parser}
	uc := NewExtractLabResultsUseCase(measRepo, artRepo, auditRepo, parsers)

	output, err := uc.Execute(context.Background(), ExtractLabResultsInput{
		ArtifactID:  "art-1",
		CoachID:     "coach-1",
		ClientID:    "client-1",
		Content:     []byte("csv"),
		ContentType: "text/csv",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.Measurements[0].Flag == nil || *output.Measurements[0].Flag != entities.LabFlagHigh {
		t.Errorf("flag = %v, want high", output.Measurements[0].Flag)
	}

	stored := measRepo.All()
	if stored[0].Flag == nil || *stored[0].Flag != entities.LabFlagHigh {
		t.Errorf("stored flag = %v, want high", stored[0].Flag)
	}
}

func TestExtractLabResults_MissingArtifactID_ReturnsError(t *testing.T) {
	measRepo, artRepo, auditRepo := newExtractTestDeps()
	parsers := map[string]domainrepo.LabParser{"text/csv": &stubLabParser{}}

	uc := NewExtractLabResultsUseCase(measRepo, artRepo, auditRepo, parsers)

	_, err := uc.Execute(context.Background(), ExtractLabResultsInput{
		CoachID:     "coach-1",
		ClientID:    "client-1",
		Content:     []byte("csv"),
		ContentType: "text/csv",
	})

	if !IsValidationError(err) {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestExtractLabResults_MissingCoachID_ReturnsError(t *testing.T) {
	measRepo, artRepo, auditRepo := newExtractTestDeps()
	parsers := map[string]domainrepo.LabParser{"text/csv": &stubLabParser{}}

	uc := NewExtractLabResultsUseCase(measRepo, artRepo, auditRepo, parsers)

	_, err := uc.Execute(context.Background(), ExtractLabResultsInput{
		ArtifactID:  "art-1",
		ClientID:    "client-1",
		Content:     []byte("csv"),
		ContentType: "text/csv",
	})

	if !IsValidationError(err) {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestExtractLabResults_MissingClientID_ReturnsError(t *testing.T) {
	measRepo, artRepo, auditRepo := newExtractTestDeps()
	parsers := map[string]domainrepo.LabParser{"text/csv": &stubLabParser{}}

	uc := NewExtractLabResultsUseCase(measRepo, artRepo, auditRepo, parsers)

	_, err := uc.Execute(context.Background(), ExtractLabResultsInput{
		ArtifactID:  "art-1",
		CoachID:     "coach-1",
		Content:     []byte("csv"),
		ContentType: "text/csv",
	})

	if !IsValidationError(err) {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestExtractLabResults_EmptyContent_ReturnsError(t *testing.T) {
	measRepo, artRepo, auditRepo := newExtractTestDeps()
	parsers := map[string]domainrepo.LabParser{"text/csv": &stubLabParser{}}

	uc := NewExtractLabResultsUseCase(measRepo, artRepo, auditRepo, parsers)

	_, err := uc.Execute(context.Background(), ExtractLabResultsInput{
		ArtifactID:  "art-1",
		CoachID:     "coach-1",
		ClientID:    "client-1",
		Content:     []byte{},
		ContentType: "text/csv",
	})

	if !IsValidationError(err) {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestExtractLabResults_UnsupportedContentType_ReturnsError(t *testing.T) {
	measRepo, artRepo, auditRepo := newExtractTestDeps()
	parsers := map[string]domainrepo.LabParser{"text/csv": &stubLabParser{}}

	uc := NewExtractLabResultsUseCase(measRepo, artRepo, auditRepo, parsers)

	_, err := uc.Execute(context.Background(), ExtractLabResultsInput{
		ArtifactID:  "art-1",
		CoachID:     "coach-1",
		ClientID:    "client-1",
		Content:     []byte("content"),
		ContentType: "application/xml",
	})

	if !IsValidationError(err) {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestExtractLabResults_ParserError_ReturnsError(t *testing.T) {
	measRepo, artRepo, auditRepo := newExtractTestDeps()

	parser := &stubLabParser{err: errors.New("parse failure")}
	parsers := map[string]domainrepo.LabParser{"text/csv": parser}

	uc := NewExtractLabResultsUseCase(measRepo, artRepo, auditRepo, parsers)

	_, err := uc.Execute(context.Background(), ExtractLabResultsInput{
		ArtifactID:  "art-1",
		CoachID:     "coach-1",
		ClientID:    "client-1",
		Content:     []byte("bad content"),
		ContentType: "text/csv",
	})

	if err == nil {
		t.Fatal("expected error for parser failure")
	}
}

func TestExtractLabResults_EmptyParsedResults_ReturnsZero(t *testing.T) {
	measRepo, artRepo, auditRepo := newExtractTestDeps()

	parser := &stubLabParser{measurements: []entities.LabMeasurement{}}
	parsers := map[string]domainrepo.LabParser{"text/csv": parser}

	uc := NewExtractLabResultsUseCase(measRepo, artRepo, auditRepo, parsers)

	output, err := uc.Execute(context.Background(), ExtractLabResultsInput{
		ArtifactID:  "art-1",
		CoachID:     "coach-1",
		ClientID:    "client-1",
		Content:     []byte("csv with no recognized markers"),
		ContentType: "text/csv",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.MeasurementsCount != 0 {
		t.Errorf("count = %d, want 0", output.MeasurementsCount)
	}

	stored := measRepo.All()
	if len(stored) != 0 {
		t.Errorf("stored %d, want 0", len(stored))
	}
}

func TestExtractLabResults_AuditEventLogged(t *testing.T) {
	measRepo, artRepo, auditRepo := newExtractTestDeps()
	normalFlag := entities.LabFlagNormal

	parser := &stubLabParser{
		measurements: []entities.LabMeasurement{
			{
				MeasurementType: entities.LabTypeFastingGlucose,
				Value:           98,
				Unit:            "mg/dL",
				Flag:            &normalFlag,
			},
		},
	}

	parsers := map[string]domainrepo.LabParser{"text/csv": parser}
	uc := NewExtractLabResultsUseCase(measRepo, artRepo, auditRepo, parsers)

	_, err := uc.Execute(context.Background(), ExtractLabResultsInput{
		ArtifactID:  "art-1",
		CoachID:     "coach-1",
		ClientID:    "client-1",
		Content:     []byte("csv"),
		ContentType: "text/csv",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify audit event was logged
	events, _, err := auditRepo.Query(context.Background(), entities.AuditQueryFilter{
		Action: "lab_results.extracted",
	})
	if err != nil {
		t.Fatalf("query audit: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("audit events = %d, want 1", len(events))
	}
	if events[0].EntityID != "art-1" {
		t.Errorf("entity_id = %s, want art-1", events[0].EntityID)
	}
	if events[0].ActorID != "coach-1" {
		t.Errorf("actor_id = %s, want coach-1", events[0].ActorID)
	}
}

func TestExtractLabResults_MissingContentType_ReturnsError(t *testing.T) {
	measRepo, artRepo, auditRepo := newExtractTestDeps()
	parsers := map[string]domainrepo.LabParser{"text/csv": &stubLabParser{}}

	uc := NewExtractLabResultsUseCase(measRepo, artRepo, auditRepo, parsers)

	_, err := uc.Execute(context.Background(), ExtractLabResultsInput{
		ArtifactID: "art-1",
		CoachID:    "coach-1",
		ClientID:   "client-1",
		Content:    []byte("csv"),
	})

	if !IsValidationError(err) {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestExtractLabResults_MeasurementInputsLinkedToArtifact(t *testing.T) {
	measRepo, artRepo, auditRepo := newExtractTestDeps()

	parser := &stubLabParser{
		measurements: []entities.LabMeasurement{
			{
				MeasurementType: entities.LabTypeTSH,
				Value:           2.1,
				Unit:            "mIU/L",
				ReferenceLow:    entities.FloatPtr(0.4),
				ReferenceHigh:   entities.FloatPtr(4.0),
			},
		},
	}

	parsers := map[string]domainrepo.LabParser{"text/csv": parser}
	uc := NewExtractLabResultsUseCase(measRepo, artRepo, auditRepo, parsers)

	_, err := uc.Execute(context.Background(), ExtractLabResultsInput{
		ArtifactID:  "art-42",
		CoachID:     "coach-1",
		ClientID:    "client-1",
		Content:     []byte("csv"),
		ContentType: "text/csv",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored := measRepo.All()
	if len(stored) != 1 {
		t.Fatalf("stored %d, want 1", len(stored))
	}
	if stored[0].ArtifactID == nil || *stored[0].ArtifactID != "art-42" {
		t.Errorf("artifact_id = %v, want art-42", stored[0].ArtifactID)
	}
	if stored[0].ReferenceLow == nil || *stored[0].ReferenceLow != 0.4 {
		t.Errorf("reference_low = %v, want 0.4", stored[0].ReferenceLow)
	}
	if stored[0].ReferenceHigh == nil || *stored[0].ReferenceHigh != 4.0 {
		t.Errorf("reference_high = %v, want 4.0", stored[0].ReferenceHigh)
	}
}

func TestExtractLabResults_FlaggedCount_CorrectInAudit(t *testing.T) {
	measRepo, artRepo, auditRepo := newExtractTestDeps()
	normalFlag := entities.LabFlagNormal
	highFlag := entities.LabFlagHigh
	lowFlag := entities.LabFlagLow

	parser := &stubLabParser{
		measurements: []entities.LabMeasurement{
			{MeasurementType: entities.LabTypeFastingGlucose, Value: 98, Unit: "mg/dL", Flag: &normalFlag},
			{MeasurementType: entities.LabTypeLDLCholesterol, Value: 142, Unit: "mg/dL", Flag: &highFlag},
			{MeasurementType: entities.LabTypeVitaminD, Value: 15, Unit: "ng/mL", Flag: &lowFlag},
		},
	}

	parsers := map[string]domainrepo.LabParser{"text/csv": parser}
	uc := NewExtractLabResultsUseCase(measRepo, artRepo, auditRepo, parsers)

	_, err := uc.Execute(context.Background(), ExtractLabResultsInput{
		ArtifactID:  "art-1",
		CoachID:     "coach-1",
		ClientID:    "client-1",
		Content:     []byte("csv"),
		ContentType: "text/csv",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events, _, _ := auditRepo.Query(context.Background(), entities.AuditQueryFilter{Action: "lab_results.extracted"})
	if len(events) != 1 {
		t.Fatalf("audit events = %d, want 1", len(events))
	}

	flaggedCount, ok := events[0].Metadata["flagged_count"]
	if !ok {
		t.Fatal("missing flagged_count in metadata")
	}
	if flaggedCount != 2 {
		t.Errorf("flagged_count = %v, want 2", flaggedCount)
	}
}
