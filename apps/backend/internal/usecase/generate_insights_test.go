package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func newGenerateInsightsUseCase() (
	*GenerateInsightsUseCase,
	*repository.InMemoryInsightCardRepository,
	*repository.InMemoryMeasurementRepository,
	*repository.InMemoryAuditRepository,
) {
	insightRepo := repository.NewInMemoryInsightCardRepository()
	measurementRepo := repository.NewInMemoryMeasurementRepository()
	auditRepo := repository.NewInMemoryAuditRepository()

	uc := NewGenerateInsightsUseCase(insightRepo, measurementRepo, auditRepo)
	return uc, insightRepo, measurementRepo, auditRepo
}

func refFloat(f float64) *float64 { return &f }

func TestGenerateInsights_LabOutOfRange_HighLDL_CreatesInsight(t *testing.T) {
	uc, insightRepo, measurementRepo, auditRepo := newGenerateInsightsUseCase()

	measurementRepo.Add(&entities.Measurement{
		ID:           "m-1",
		ClientID:     "client-1",
		CoachID:      "coach-1",
		ArtifactID:   "artifact-1",
		Type:         entities.MeasurementTypeLab,
		MarkerName:   "LDL Cholesterol",
		Value:        142,
		Unit:         "mg/dL",
		ReferenceMax: refFloat(100),
		Flag:         entities.MeasurementFlagHigh,
		RecordedAt:   time.Now(),
		CreatedAt:    time.Now(),
	})

	input := GenerateInsightsInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "artifact-1",
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.InsightCards) != 1 {
		t.Fatalf("expected 1 insight card, got %d", len(output.InsightCards))
	}

	card := output.InsightCards[0]
	if card.Priority != entities.InsightPriorityHigh {
		t.Errorf("expected priority %q, got %q", entities.InsightPriorityHigh, card.Priority)
	}
	if card.Status != entities.InsightStatusDraft {
		t.Errorf("expected status %q, got %q", entities.InsightStatusDraft, card.Status)
	}
	if card.ClientID != "client-1" {
		t.Errorf("expected client_id %q, got %q", "client-1", card.ClientID)
	}
	if card.CoachID != "coach-1" {
		t.Errorf("expected coach_id %q, got %q", "coach-1", card.CoachID)
	}
	if len(card.Evidence) != 1 {
		t.Fatalf("expected 1 evidence ref, got %d", len(card.Evidence))
	}
	if card.Evidence[0].MeasurementID != "m-1" {
		t.Errorf("expected measurement_id %q, got %q", "m-1", card.Evidence[0].MeasurementID)
	}

	// Verify persisted
	if insightRepo.CardCount() != 1 {
		t.Errorf("expected 1 card in repo, got %d", insightRepo.CardCount())
	}

	// Verify audit event logged
	if auditRepo.EventCount() != 1 {
		t.Errorf("expected 1 audit event, got %d", auditRepo.EventCount())
	}
	events := auditRepo.GetEvents()
	if events[0].Action != "insight.generate" {
		t.Errorf("expected action %q, got %q", "insight.generate", events[0].Action)
	}
}

func TestGenerateInsights_LabCritical_HighGlucose_CreatesUrgentInsight(t *testing.T) {
	uc, insightRepo, measurementRepo, _ := newGenerateInsightsUseCase()

	measurementRepo.Add(&entities.Measurement{
		ID:           "m-2",
		ClientID:     "client-1",
		CoachID:      "coach-1",
		ArtifactID:   "artifact-1",
		Type:         entities.MeasurementTypeLab,
		MarkerName:   "Fasting Glucose",
		Value:        250,
		Unit:         "mg/dL",
		ReferenceMax: refFloat(100),
		Flag:         entities.MeasurementFlagCriticalHigh,
		RecordedAt:   time.Now(),
		CreatedAt:    time.Now(),
	})

	input := GenerateInsightsInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "artifact-1",
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.InsightCards) != 1 {
		t.Fatalf("expected 1 insight card, got %d", len(output.InsightCards))
	}

	card := output.InsightCards[0]
	if card.Priority != entities.InsightPriorityUrgent {
		t.Errorf("expected priority %q, got %q", entities.InsightPriorityUrgent, card.Priority)
	}
	if card.Category != entities.InsightCategorySafety {
		t.Errorf("expected category %q, got %q", entities.InsightCategorySafety, card.Category)
	}

	if insightRepo.CardCount() != 1 {
		t.Errorf("expected 1 card in repo, got %d", insightRepo.CardCount())
	}
}

func TestGenerateInsights_MultipleOutOfRange_CreatesMultipleInsights(t *testing.T) {
	uc, insightRepo, measurementRepo, _ := newGenerateInsightsUseCase()

	markers := []struct {
		id     string
		name   string
		value  float64
		unit   string
		refMax float64
		flag   entities.MeasurementFlag
	}{
		{"m-1", "LDL Cholesterol", 142, "mg/dL", 100, entities.MeasurementFlagHigh},
		{"m-2", "Triglycerides", 200, "mg/dL", 150, entities.MeasurementFlagHigh},
		{"m-3", "HDL Cholesterol", 35, "mg/dL", 40, entities.MeasurementFlagLow},
	}

	for _, m := range markers {
		measurementRepo.Add(&entities.Measurement{
			ID:           m.id,
			ClientID:     "client-1",
			CoachID:      "coach-1",
			ArtifactID:   "artifact-1",
			Type:         entities.MeasurementTypeLab,
			MarkerName:   m.name,
			Value:        m.value,
			Unit:         m.unit,
			ReferenceMax: refFloat(m.refMax),
			Flag:         m.flag,
			RecordedAt:   time.Now(),
			CreatedAt:    time.Now(),
		})
	}

	input := GenerateInsightsInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "artifact-1",
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.InsightCards) != 3 {
		t.Fatalf("expected 3 insight cards, got %d", len(output.InsightCards))
	}

	if insightRepo.CardCount() != 3 {
		t.Errorf("expected 3 cards in repo, got %d", insightRepo.CardCount())
	}
}

func TestGenerateInsights_AllInRange_NoInsightsCreated(t *testing.T) {
	uc, insightRepo, measurementRepo, _ := newGenerateInsightsUseCase()

	measurementRepo.Add(&entities.Measurement{
		ID:           "m-1",
		ClientID:     "client-1",
		CoachID:      "coach-1",
		ArtifactID:   "artifact-1",
		Type:         entities.MeasurementTypeLab,
		MarkerName:   "LDL Cholesterol",
		Value:        85,
		Unit:         "mg/dL",
		ReferenceMax: refFloat(100),
		Flag:         entities.MeasurementFlagNormal,
		RecordedAt:   time.Now(),
		CreatedAt:    time.Now(),
	})

	input := GenerateInsightsInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "artifact-1",
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.InsightCards) != 0 {
		t.Fatalf("expected 0 insight cards, got %d", len(output.InsightCards))
	}

	if insightRepo.CardCount() != 0 {
		t.Errorf("expected 0 cards in repo, got %d", insightRepo.CardCount())
	}
}

func TestGenerateInsights_DuplicatePrevention_SkipsExistingMeasurement(t *testing.T) {
	uc, insightRepo, measurementRepo, _ := newGenerateInsightsUseCase()

	m := &entities.Measurement{
		ID:           "m-1",
		ClientID:     "client-1",
		CoachID:      "coach-1",
		ArtifactID:   "artifact-1",
		Type:         entities.MeasurementTypeLab,
		MarkerName:   "LDL Cholesterol",
		Value:        142,
		Unit:         "mg/dL",
		ReferenceMax: refFloat(100),
		Flag:         entities.MeasurementFlagHigh,
		RecordedAt:   time.Now(),
		CreatedAt:    time.Now(),
	}
	measurementRepo.Add(m)

	input := GenerateInsightsInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "artifact-1",
	}

	// First call creates insight
	output1, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error on first call: %v", err)
	}
	if len(output1.InsightCards) != 1 {
		t.Fatalf("expected 1 insight on first call, got %d", len(output1.InsightCards))
	}

	// Second call with same data should skip (duplicate)
	output2, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if len(output2.InsightCards) != 0 {
		t.Fatalf("expected 0 insights on second call (duplicate), got %d", len(output2.InsightCards))
	}

	if insightRepo.CardCount() != 1 {
		t.Errorf("expected 1 card total in repo, got %d", insightRepo.CardCount())
	}
}

func TestGenerateInsights_HRVDeclining_CreatesMediumInsight(t *testing.T) {
	uc, insightRepo, measurementRepo, _ := newGenerateInsightsUseCase()

	now := time.Now()

	// Add HRV data for prior 7 days (avg ~60ms)
	for i := 14; i >= 8; i-- {
		measurementRepo.Add(&entities.Measurement{
			ID:         fmt.Sprintf("m-prior-%d", i),
			ClientID:   "client-1",
			CoachID:    "coach-1",
			ArtifactID: "artifact-old",
			Type:       entities.MeasurementTypeWearable,
			MarkerName: "HRV",
			Value:      60,
			Unit:       "ms",
			Flag:       entities.MeasurementFlagNormal,
			RecordedAt: now.AddDate(0, 0, -i),
			CreatedAt:  now.AddDate(0, 0, -i),
		})
	}

	// Add HRV data for recent 7 days (avg ~48ms => 20% decline)
	for i := 7; i >= 1; i-- {
		measurementRepo.Add(&entities.Measurement{
			ID:         fmt.Sprintf("m-recent-%d", i),
			ClientID:   "client-1",
			CoachID:    "coach-1",
			ArtifactID: "artifact-1",
			Type:       entities.MeasurementTypeWearable,
			MarkerName: "HRV",
			Value:      48,
			Unit:       "ms",
			Flag:       entities.MeasurementFlagNormal,
			RecordedAt: now.AddDate(0, 0, -i),
			CreatedAt:  now.AddDate(0, 0, -i),
		})
	}

	input := GenerateInsightsInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "artifact-1",
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.InsightCards) != 1 {
		t.Fatalf("expected 1 insight card for HRV decline, got %d", len(output.InsightCards))
	}

	card := output.InsightCards[0]
	if card.Priority != entities.InsightPriorityMedium {
		t.Errorf("expected priority %q, got %q", entities.InsightPriorityMedium, card.Priority)
	}
	if card.Category != entities.InsightCategoryRecovery {
		t.Errorf("expected category %q, got %q", entities.InsightCategoryRecovery, card.Category)
	}

	if insightRepo.CardCount() != 1 {
		t.Errorf("expected 1 card in repo, got %d", insightRepo.CardCount())
	}
}

func TestGenerateInsights_SleepDeclining_CreatesMediumInsight(t *testing.T) {
	uc, insightRepo, measurementRepo, _ := newGenerateInsightsUseCase()

	now := time.Now()

	// Add sleep data for recent 7 days (avg 5.8 hours)
	for i := 7; i >= 1; i-- {
		measurementRepo.Add(&entities.Measurement{
			ID:         fmt.Sprintf("m-sleep-%d", i),
			ClientID:   "client-1",
			CoachID:    "coach-1",
			ArtifactID: "artifact-1",
			Type:       entities.MeasurementTypeWearable,
			MarkerName: "Sleep Duration",
			Value:      5.8,
			Unit:       "hours",
			Flag:       entities.MeasurementFlagNormal,
			RecordedAt: now.AddDate(0, 0, -i),
			CreatedAt:  now.AddDate(0, 0, -i),
		})
	}

	input := GenerateInsightsInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "artifact-1",
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.InsightCards) != 1 {
		t.Fatalf("expected 1 insight card for sleep decline, got %d", len(output.InsightCards))
	}

	card := output.InsightCards[0]
	if card.Priority != entities.InsightPriorityMedium {
		t.Errorf("expected priority %q, got %q", entities.InsightPriorityMedium, card.Priority)
	}
	if card.Category != entities.InsightCategoryRecovery {
		t.Errorf("expected category %q, got %q", entities.InsightCategoryRecovery, card.Category)
	}

	if insightRepo.CardCount() != 1 {
		t.Errorf("expected 1 card in repo, got %d", insightRepo.CardCount())
	}
}

func TestGenerateInsights_WeightChange_CreatesMediumInsight(t *testing.T) {
	uc, insightRepo, measurementRepo, _ := newGenerateInsightsUseCase()

	now := time.Now()

	// Weight 2 weeks ago: 200 lbs
	measurementRepo.Add(&entities.Measurement{
		ID:         "m-weight-old",
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "artifact-old",
		Type:       entities.MeasurementTypeBodyComp,
		MarkerName: "Weight",
		Value:      200,
		Unit:       "lbs",
		Flag:       entities.MeasurementFlagNormal,
		RecordedAt: now.AddDate(0, 0, -13),
		CreatedAt:  now.AddDate(0, 0, -13),
	})

	// Weight now: 191.6 lbs (4.2% decrease)
	measurementRepo.Add(&entities.Measurement{
		ID:         "m-weight-new",
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "artifact-1",
		Type:       entities.MeasurementTypeBodyComp,
		MarkerName: "Weight",
		Value:      191.6,
		Unit:       "lbs",
		Flag:       entities.MeasurementFlagNormal,
		RecordedAt: now,
		CreatedAt:  now,
	})

	input := GenerateInsightsInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "artifact-1",
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.InsightCards) != 1 {
		t.Fatalf("expected 1 insight card for weight change, got %d", len(output.InsightCards))
	}

	card := output.InsightCards[0]
	if card.Priority != entities.InsightPriorityMedium {
		t.Errorf("expected priority %q, got %q", entities.InsightPriorityMedium, card.Priority)
	}

	if insightRepo.CardCount() != 1 {
		t.Errorf("expected 1 card in repo, got %d", insightRepo.CardCount())
	}
}

func TestGenerateInsights_BodyFatProgress_CreatesLowInsight(t *testing.T) {
	uc, insightRepo, measurementRepo, _ := newGenerateInsightsUseCase()

	now := time.Now()

	// Body fat 2 weeks ago: 22.3%
	measurementRepo.Add(&entities.Measurement{
		ID:         "m-bf-old",
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "artifact-old",
		Type:       entities.MeasurementTypeBodyComp,
		MarkerName: "Body Fat",
		Value:      22.3,
		Unit:       "%",
		Flag:       entities.MeasurementFlagNormal,
		RecordedAt: now.AddDate(0, 0, -13),
		CreatedAt:  now.AddDate(0, 0, -13),
	})

	// Body fat now: 21.1% (1.2% decrease)
	measurementRepo.Add(&entities.Measurement{
		ID:         "m-bf-new",
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "artifact-1",
		Type:       entities.MeasurementTypeBodyComp,
		MarkerName: "Body Fat",
		Value:      21.1,
		Unit:       "%",
		Flag:       entities.MeasurementFlagNormal,
		RecordedAt: now,
		CreatedAt:  now,
	})

	input := GenerateInsightsInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "artifact-1",
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.InsightCards) != 1 {
		t.Fatalf("expected 1 insight card for body fat progress, got %d", len(output.InsightCards))
	}

	card := output.InsightCards[0]
	if card.Priority != entities.InsightPriorityLow {
		t.Errorf("expected priority %q, got %q", entities.InsightPriorityLow, card.Priority)
	}

	if insightRepo.CardCount() != 1 {
		t.Errorf("expected 1 card in repo, got %d", insightRepo.CardCount())
	}
}

func TestGenerateInsights_EvidenceLinking_ReferencesSourceMeasurements(t *testing.T) {
	uc, _, measurementRepo, _ := newGenerateInsightsUseCase()

	measurementRepo.Add(&entities.Measurement{
		ID:           "m-123",
		ClientID:     "client-1",
		CoachID:      "coach-1",
		ArtifactID:   "artifact-456",
		Type:         entities.MeasurementTypeLab,
		MarkerName:   "LDL Cholesterol",
		Value:        142,
		Unit:         "mg/dL",
		ReferenceMax: refFloat(100),
		Flag:         entities.MeasurementFlagHigh,
		RecordedAt:   time.Now(),
		CreatedAt:    time.Now(),
	})

	input := GenerateInsightsInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "artifact-456",
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.InsightCards) != 1 {
		t.Fatalf("expected 1 insight card, got %d", len(output.InsightCards))
	}

	card := output.InsightCards[0]
	if len(card.Evidence) != 1 {
		t.Fatalf("expected 1 evidence ref, got %d", len(card.Evidence))
	}

	ev := card.Evidence[0]
	if ev.MeasurementID != "m-123" {
		t.Errorf("expected measurement_id %q, got %q", "m-123", ev.MeasurementID)
	}
	if ev.ArtifactID != "artifact-456" {
		t.Errorf("expected artifact_id %q, got %q", "artifact-456", ev.ArtifactID)
	}
	if ev.Description == "" {
		t.Error("expected non-empty evidence description")
	}
}

func TestGenerateInsights_MissingClientID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newGenerateInsightsUseCase()

	input := GenerateInsightsInput{
		CoachID:    "coach-1",
		ArtifactID: "artifact-1",
	}

	_, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected validation error for missing client_id")
	}

	var validationErr *entities.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestGenerateInsights_MissingCoachID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newGenerateInsightsUseCase()

	input := GenerateInsightsInput{
		ClientID:   "client-1",
		ArtifactID: "artifact-1",
	}

	_, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected validation error for missing coach_id")
	}

	var validationErr *entities.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestGenerateInsights_MissingArtifactID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newGenerateInsightsUseCase()

	input := GenerateInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
	}

	_, err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected validation error for missing artifact_id")
	}

	var validationErr *entities.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestGenerateInsights_NoMeasurements_ReturnsEmptyInsights(t *testing.T) {
	uc, insightRepo, _, _ := newGenerateInsightsUseCase()

	input := GenerateInsightsInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "artifact-1",
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.InsightCards) != 0 {
		t.Fatalf("expected 0 insight cards, got %d", len(output.InsightCards))
	}

	if insightRepo.CardCount() != 0 {
		t.Errorf("expected 0 cards in repo, got %d", insightRepo.CardCount())
	}
}

func TestGenerateInsights_AllCardsHaveDraftStatus(t *testing.T) {
	uc, _, measurementRepo, _ := newGenerateInsightsUseCase()

	measurementRepo.Add(&entities.Measurement{
		ID:           "m-1",
		ClientID:     "client-1",
		CoachID:      "coach-1",
		ArtifactID:   "artifact-1",
		Type:         entities.MeasurementTypeLab,
		MarkerName:   "LDL Cholesterol",
		Value:        142,
		Unit:         "mg/dL",
		ReferenceMax: refFloat(100),
		Flag:         entities.MeasurementFlagHigh,
		RecordedAt:   time.Now(),
		CreatedAt:    time.Now(),
	})

	measurementRepo.Add(&entities.Measurement{
		ID:           "m-2",
		ClientID:     "client-1",
		CoachID:      "coach-1",
		ArtifactID:   "artifact-1",
		Type:         entities.MeasurementTypeLab,
		MarkerName:   "Fasting Glucose",
		Value:        250,
		Unit:         "mg/dL",
		ReferenceMax: refFloat(100),
		Flag:         entities.MeasurementFlagCriticalHigh,
		RecordedAt:   time.Now(),
		CreatedAt:    time.Now(),
	})

	input := GenerateInsightsInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "artifact-1",
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, card := range output.InsightCards {
		if card.Status != entities.InsightStatusDraft {
			t.Errorf("card %d: expected status %q, got %q", i, entities.InsightStatusDraft, card.Status)
		}
	}
}

func TestGenerateInsights_AuditEventsLoggedForEachInsight(t *testing.T) {
	uc, _, measurementRepo, auditRepo := newGenerateInsightsUseCase()

	measurementRepo.Add(&entities.Measurement{
		ID:           "m-1",
		ClientID:     "client-1",
		CoachID:      "coach-1",
		ArtifactID:   "artifact-1",
		Type:         entities.MeasurementTypeLab,
		MarkerName:   "LDL Cholesterol",
		Value:        142,
		Unit:         "mg/dL",
		ReferenceMax: refFloat(100),
		Flag:         entities.MeasurementFlagHigh,
		RecordedAt:   time.Now(),
		CreatedAt:    time.Now(),
	})

	measurementRepo.Add(&entities.Measurement{
		ID:           "m-2",
		ClientID:     "client-1",
		CoachID:      "coach-1",
		ArtifactID:   "artifact-1",
		Type:         entities.MeasurementTypeLab,
		MarkerName:   "Triglycerides",
		Value:        200,
		Unit:         "mg/dL",
		ReferenceMax: refFloat(150),
		Flag:         entities.MeasurementFlagHigh,
		RecordedAt:   time.Now(),
		CreatedAt:    time.Now(),
	})

	input := GenerateInsightsInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "artifact-1",
	}

	_, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if auditRepo.EventCount() != 2 {
		t.Errorf("expected 2 audit events (one per insight), got %d", auditRepo.EventCount())
	}

	for _, ev := range auditRepo.GetEvents() {
		if ev.Action != "insight.generate" {
			t.Errorf("expected action %q, got %q", "insight.generate", ev.Action)
		}
		if ev.EntityType != "insight_card" {
			t.Errorf("expected entity_type %q, got %q", "insight_card", ev.EntityType)
		}
	}
}

func TestGenerateInsights_LabLowFlag_CreatesInsight(t *testing.T) {
	uc, _, measurementRepo, _ := newGenerateInsightsUseCase()

	measurementRepo.Add(&entities.Measurement{
		ID:           "m-1",
		ClientID:     "client-1",
		CoachID:      "coach-1",
		ArtifactID:   "artifact-1",
		Type:         entities.MeasurementTypeLab,
		MarkerName:   "Vitamin D",
		Value:        15,
		Unit:         "ng/mL",
		ReferenceMin: refFloat(30),
		Flag:         entities.MeasurementFlagLow,
		RecordedAt:   time.Now(),
		CreatedAt:    time.Now(),
	})

	input := GenerateInsightsInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "artifact-1",
	}

	output, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.InsightCards) != 1 {
		t.Fatalf("expected 1 insight card, got %d", len(output.InsightCards))
	}

	card := output.InsightCards[0]
	if card.Priority != entities.InsightPriorityHigh {
		t.Errorf("expected priority %q, got %q", entities.InsightPriorityHigh, card.Priority)
	}
}
