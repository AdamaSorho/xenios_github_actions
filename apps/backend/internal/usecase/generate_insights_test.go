package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func newGenerateInsightsUseCase() (
	*GenerateInsightsUseCase,
	*repository.InMemoryInsightCardRepository,
	*repository.InMemoryMeasurementRepository,
	*repository.InMemoryWearableSummaryRepository,
	*repository.InMemoryAuditRepository,
) {
	insightRepo := repository.NewInMemoryInsightCardRepository()
	measurementRepo := repository.NewInMemoryMeasurementRepository()
	wearableRepo := repository.NewInMemoryWearableSummaryRepository()
	auditRepo := repository.NewInMemoryAuditRepository()

	uc := NewGenerateInsightsUseCase(insightRepo, measurementRepo, wearableRepo, auditRepo)
	return uc, insightRepo, measurementRepo, wearableRepo, auditRepo
}

// --- Validation Tests ---

func TestGenerateInsights_EmptyClientID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _ := newGenerateInsightsUseCase()

	_, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "",
		CoachID:  "coach-1",
	})
	if err == nil {
		t.Fatal("expected error for empty client_id")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestGenerateInsights_EmptyCoachID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _, _ := newGenerateInsightsUseCase()

	_, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1",
		CoachID:  "",
	})
	if err == nil {
		t.Fatal("expected error for empty coach_id")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

// --- Lab Out-of-Range Tests ---

func TestGenerateInsights_HighLDL_CreatesDraftInsight(t *testing.T) {
	uc, insightRepo, measurementRepo, _, _ := newGenerateInsightsUseCase()

	now := time.Now()
	measurementRepo.Add(&entities.Measurement{
		ID:              "m-1",
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "ldl_cholesterol",
		Value:           142,
		Unit:            "mg/dL",
		MeasuredAt:      now,
		CreatedAt:       now,
	})

	result, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.InsightsCreated) != 1 {
		t.Fatalf("expected 1 insight, got %d", len(result.InsightsCreated))
	}

	card := result.InsightsCreated[0]
	if card.Status != entities.InsightStatusDraft {
		t.Errorf("expected status 'draft', got %q", card.Status)
	}
	if card.Priority != entities.InsightPriorityHigh {
		t.Errorf("expected priority 'high', got %q", card.Priority)
	}
	if card.ClientID != "client-1" {
		t.Errorf("expected client_id 'client-1', got %q", card.ClientID)
	}
	if card.CoachID != "coach-1" {
		t.Errorf("expected coach_id 'coach-1', got %q", card.CoachID)
	}
	if len(card.Evidence) == 0 {
		t.Fatal("expected evidence references")
	}
	if card.Evidence[0].MeasurementID != "m-1" {
		t.Errorf("expected evidence measurement_id 'm-1', got %q", card.Evidence[0].MeasurementID)
	}

	// Verify it was persisted
	all := insightRepo.GetAll()
	if len(all) != 1 {
		t.Errorf("expected 1 insight in repository, got %d", len(all))
	}
}

func TestGenerateInsights_CriticalGlucose_CreatesUrgentInsight(t *testing.T) {
	uc, _, measurementRepo, _, _ := newGenerateInsightsUseCase()

	now := time.Now()
	measurementRepo.Add(&entities.Measurement{
		ID:              "m-2",
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "fasting_glucose",
		Value:           250,
		Unit:            "mg/dL",
		MeasuredAt:      now,
		CreatedAt:       now,
	})

	result, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.InsightsCreated) != 1 {
		t.Fatalf("expected 1 insight, got %d", len(result.InsightsCreated))
	}

	card := result.InsightsCreated[0]
	if card.Priority != entities.InsightPriorityUrgent {
		t.Errorf("expected priority 'urgent', got %q", card.Priority)
	}
	if card.Category != entities.InsightCategorySafety {
		t.Errorf("expected category 'safety', got %q", card.Category)
	}
}

func TestGenerateInsights_MultipleOutOfRange_CreatesMultipleInsights(t *testing.T) {
	uc, _, measurementRepo, _, _ := newGenerateInsightsUseCase()

	now := time.Now()
	// High LDL
	measurementRepo.Add(&entities.Measurement{
		ID:              "m-1",
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "ldl_cholesterol",
		Value:           142,
		Unit:            "mg/dL",
		MeasuredAt:      now,
		CreatedAt:       now,
	})
	// High triglycerides
	measurementRepo.Add(&entities.Measurement{
		ID:              "m-2",
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "triglycerides",
		Value:           200,
		Unit:            "mg/dL",
		MeasuredAt:      now,
		CreatedAt:       now,
	})
	// High A1c
	measurementRepo.Add(&entities.Measurement{
		ID:              "m-3",
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "hemoglobin_a1c",
		Value:           6.5,
		Unit:            "%",
		MeasuredAt:      now,
		CreatedAt:       now,
	})

	result, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.InsightsCreated) != 3 {
		t.Fatalf("expected 3 insights, got %d", len(result.InsightsCreated))
	}
}

func TestGenerateInsights_AllValuesNormal_NoInsightsCreated(t *testing.T) {
	uc, _, measurementRepo, _, _ := newGenerateInsightsUseCase()

	now := time.Now()
	measurementRepo.Add(&entities.Measurement{
		ID:              "m-1",
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "ldl_cholesterol",
		Value:           80,
		Unit:            "mg/dL",
		MeasuredAt:      now,
		CreatedAt:       now,
	})
	measurementRepo.Add(&entities.Measurement{
		ID:              "m-2",
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "fasting_glucose",
		Value:           90,
		Unit:            "mg/dL",
		MeasuredAt:      now,
		CreatedAt:       now,
	})

	result, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.InsightsCreated) != 0 {
		t.Errorf("expected 0 insights for normal values, got %d", len(result.InsightsCreated))
	}
}

// --- Duplicate Prevention Tests ---

func TestGenerateInsights_DuplicateMeasurement_SkipsDuplicate(t *testing.T) {
	uc, insightRepo, measurementRepo, _, _ := newGenerateInsightsUseCase()

	now := time.Now()
	measurementRepo.Add(&entities.Measurement{
		ID:              "m-1",
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "ldl_cholesterol",
		Value:           142,
		Unit:            "mg/dL",
		MeasuredAt:      now,
		CreatedAt:       now,
	})

	// First generation
	result1, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error on first run: %v", err)
	}
	if len(result1.InsightsCreated) != 1 {
		t.Fatalf("expected 1 insight on first run, got %d", len(result1.InsightsCreated))
	}

	// Second generation with same data — should not create duplicates
	result2, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error on second run: %v", err)
	}
	if len(result2.InsightsCreated) != 0 {
		t.Errorf("expected 0 insights on duplicate run, got %d", len(result2.InsightsCreated))
	}

	// Verify total still 1
	all := insightRepo.GetAll()
	if len(all) != 1 {
		t.Errorf("expected 1 total insight, got %d", len(all))
	}
}

// --- Wearable Trend Tests ---

func TestGenerateInsights_HRVDeclining_CreatesMediumPriorityInsight(t *testing.T) {
	uc, _, _, wearableRepo, _ := newGenerateInsightsUseCase()

	now := time.Now()
	// Previous week: average HRV = 60
	for i := 14; i >= 8; i-- {
		wearableRepo.Add(&entities.WearableSummary{
			ID:          "ws-prev-" + string(rune('a'+14-i)),
			ClientID:    "client-1",
			Source:      "whoop",
			SummaryDate: now.AddDate(0, 0, -i),
			Metrics:     map[string]interface{}{"hrv": float64(60)},
			SyncedAt:    now,
			CreatedAt:   now,
		})
	}
	// Current week: average HRV = 48 (20% decline)
	for i := 7; i >= 1; i-- {
		wearableRepo.Add(&entities.WearableSummary{
			ID:          "ws-curr-" + string(rune('a'+7-i)),
			ClientID:    "client-1",
			Source:      "whoop",
			SummaryDate: now.AddDate(0, 0, -i),
			Metrics:     map[string]interface{}{"hrv": float64(48)},
			SyncedAt:    now,
			CreatedAt:   now,
		})
	}

	result, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have at least 1 HRV decline insight
	found := false
	for _, card := range result.InsightsCreated {
		if card.Category == entities.InsightCategoryRecovery && card.Priority == entities.InsightPriorityMedium {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected HRV decline insight with recovery category and medium priority")
	}
}

func TestGenerateInsights_SleepDeclining_CreatesMediumPriorityInsight(t *testing.T) {
	uc, _, _, wearableRepo, _ := newGenerateInsightsUseCase()

	now := time.Now()
	// Last 7 days: average sleep < 6 hours
	for i := 7; i >= 1; i-- {
		wearableRepo.Add(&entities.WearableSummary{
			ID:          "ws-sleep-" + string(rune('a'+7-i)),
			ClientID:    "client-1",
			Source:      "whoop",
			SummaryDate: now.AddDate(0, 0, -i),
			Metrics:     map[string]interface{}{"sleep_hours": float64(5.5)},
			SyncedAt:    now,
			CreatedAt:   now,
		})
	}

	result, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, card := range result.InsightsCreated {
		if card.Category == entities.InsightCategoryRecovery {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected sleep decline insight with recovery category")
	}
}

// --- Body Composition Tests ---

func TestGenerateInsights_WeightChange_CreatesMediumPriorityInsight(t *testing.T) {
	uc, _, measurementRepo, _, _ := newGenerateInsightsUseCase()

	now := time.Now()
	// Weight 2 weeks ago: 200 lbs
	measurementRepo.Add(&entities.Measurement{
		ID:              "m-w1",
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "weight",
		Value:           200,
		Unit:            "lbs",
		MeasuredAt:      now.AddDate(0, 0, -14),
		CreatedAt:       now.AddDate(0, 0, -14),
	})
	// Weight now: 190 lbs (5% decrease > 3% threshold)
	measurementRepo.Add(&entities.Measurement{
		ID:              "m-w2",
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "weight",
		Value:           190,
		Unit:            "lbs",
		MeasuredAt:      now,
		CreatedAt:       now,
	})

	result, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, card := range result.InsightsCreated {
		if card.Priority == entities.InsightPriorityMedium {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected weight change insight with medium priority")
	}
}

func TestGenerateInsights_BodyFatDecrease_CreatesLowPriorityInsight(t *testing.T) {
	uc, _, measurementRepo, _, _ := newGenerateInsightsUseCase()

	now := time.Now()
	// Body fat 2 weeks ago: 22.3%
	measurementRepo.Add(&entities.Measurement{
		ID:              "m-bf1",
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "body_fat_percentage",
		Value:           22.3,
		Unit:            "%",
		MeasuredAt:      now.AddDate(0, 0, -14),
		CreatedAt:       now.AddDate(0, 0, -14),
	})
	// Body fat now: 21.1% (1.2% decrease > 1% threshold)
	measurementRepo.Add(&entities.Measurement{
		ID:              "m-bf2",
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "body_fat_percentage",
		Value:           21.1,
		Unit:            "%",
		MeasuredAt:      now,
		CreatedAt:       now,
	})

	result, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, card := range result.InsightsCreated {
		if card.Priority == entities.InsightPriorityLow {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected body fat decrease insight with low priority")
	}
}

// --- Audit Logging Tests ---

func TestGenerateInsights_CreatesAuditEvent(t *testing.T) {
	uc, _, measurementRepo, _, auditRepo := newGenerateInsightsUseCase()

	now := time.Now()
	measurementRepo.Add(&entities.Measurement{
		ID:              "m-1",
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "ldl_cholesterol",
		Value:           142,
		Unit:            "mg/dL",
		MeasuredAt:      now,
		CreatedAt:       now,
	})

	_, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := auditRepo.GetEvents()
	if len(events) == 0 {
		t.Fatal("expected at least one audit event")
	}

	found := false
	for _, event := range events {
		if event.Action == "insight.generate" {
			found = true
			if event.ActorID != "system" {
				t.Errorf("expected actor_id 'system', got %q", event.ActorID)
			}
			if event.EntityType != "insight_card" {
				t.Errorf("expected entity_type 'insight_card', got %q", event.EntityType)
			}
			break
		}
	}
	if !found {
		t.Error("expected 'insight.generate' audit event")
	}
}

// --- Status Tests ---

func TestGenerateInsights_AllInsightsStartAsDraft(t *testing.T) {
	uc, _, measurementRepo, _, _ := newGenerateInsightsUseCase()

	now := time.Now()
	measurementRepo.Add(&entities.Measurement{
		ID:              "m-1",
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "ldl_cholesterol",
		Value:           142,
		Unit:            "mg/dL",
		MeasuredAt:      now,
		CreatedAt:       now,
	})
	measurementRepo.Add(&entities.Measurement{
		ID:              "m-2",
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "fasting_glucose",
		Value:           250,
		Unit:            "mg/dL",
		MeasuredAt:      now,
		CreatedAt:       now,
	})

	result, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, card := range result.InsightsCreated {
		if card.Status != entities.InsightStatusDraft {
			t.Errorf("insight %d: expected status 'draft', got %q", i, card.Status)
		}
	}
}

// --- Evidence Linking Tests ---

func TestGenerateInsights_EvidenceLinksToSourceMeasurement(t *testing.T) {
	uc, _, measurementRepo, _, _ := newGenerateInsightsUseCase()

	now := time.Now()
	measurementRepo.Add(&entities.Measurement{
		ID:              "m-specific-id",
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "ldl_cholesterol",
		Value:           142,
		Unit:            "mg/dL",
		MeasuredAt:      now,
		CreatedAt:       now,
	})

	result, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ArtifactID: "art-99",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.InsightsCreated) != 1 {
		t.Fatalf("expected 1 insight, got %d", len(result.InsightsCreated))
	}

	card := result.InsightsCreated[0]
	if len(card.Evidence) != 1 {
		t.Fatalf("expected 1 evidence ref, got %d", len(card.Evidence))
	}
	if card.Evidence[0].MeasurementID != "m-specific-id" {
		t.Errorf("expected evidence measurement_id 'm-specific-id', got %q", card.Evidence[0].MeasurementID)
	}
	if card.Evidence[0].ArtifactID != "art-99" {
		t.Errorf("expected evidence artifact_id 'art-99', got %q", card.Evidence[0].ArtifactID)
	}
}

// --- No Data Tests ---

func TestGenerateInsights_NoMeasurements_ReturnsEmptyResult(t *testing.T) {
	uc, _, _, _, _ := newGenerateInsightsUseCase()

	result, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.InsightsCreated) != 0 {
		t.Errorf("expected 0 insights, got %d", len(result.InsightsCreated))
	}
}

// --- Unknown Measurement Type Tests ---

func TestGenerateInsights_UnknownMeasurementType_NoInsight(t *testing.T) {
	uc, _, measurementRepo, _, _ := newGenerateInsightsUseCase()

	now := time.Now()
	measurementRepo.Add(&entities.Measurement{
		ID:              "m-1",
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "some_unknown_marker",
		Value:           999,
		Unit:            "units",
		MeasuredAt:      now,
		CreatedAt:       now,
	})

	result, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.InsightsCreated) != 0 {
		t.Errorf("expected 0 insights for unknown type, got %d", len(result.InsightsCreated))
	}
}

// --- HRV Stable Tests ---

func TestGenerateInsights_HRVStable_NoInsight(t *testing.T) {
	uc, _, _, wearableRepo, _ := newGenerateInsightsUseCase()

	now := time.Now()
	// Both weeks have same HRV — no decline
	for i := 14; i >= 1; i-- {
		wearableRepo.Add(&entities.WearableSummary{
			ID:          "ws-" + string(rune('a'+14-i)),
			ClientID:    "client-1",
			Source:      "whoop",
			SummaryDate: now.AddDate(0, 0, -i),
			Metrics:     map[string]interface{}{"hrv": float64(60)},
			SyncedAt:    now,
			CreatedAt:   now,
		})
	}

	result, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, card := range result.InsightsCreated {
		if card.Category == entities.InsightCategoryRecovery {
			t.Error("expected no recovery insight for stable HRV")
		}
	}
}

// --- Result Structure Tests ---

func TestGenerateInsights_ResultContainsTotalEvaluated(t *testing.T) {
	uc, _, measurementRepo, _, _ := newGenerateInsightsUseCase()

	now := time.Now()
	measurementRepo.Add(&entities.Measurement{
		ID:              "m-1",
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "ldl_cholesterol",
		Value:           80,
		Unit:            "mg/dL",
		MeasuredAt:      now,
		CreatedAt:       now,
	})

	result, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1",
		CoachID:  "coach-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalEvaluated == 0 {
		t.Error("expected TotalEvaluated > 0")
	}
}
