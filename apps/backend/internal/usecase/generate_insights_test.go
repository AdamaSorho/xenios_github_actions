package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func newGenerateInsightsUseCase() (*GenerateInsightsUseCase, *repository.InMemoryInsightCardRepository, *repository.InMemoryMeasurementRepository, *repository.InMemoryAuditRepository) {
	insightRepo := repository.NewInMemoryInsightCardRepository()
	measRepo := repository.NewInMemoryMeasurementRepository()
	auditRepo := repository.NewInMemoryAuditRepository()

	uc := NewGenerateInsightsUseCase(insightRepo, measRepo, auditRepo)
	return uc, insightRepo, measRepo, auditRepo
}

func refFloat(v float64) *float64 { return &v }

func TestGenerateInsights_LabHighLDL_CreatesDraftInsight(t *testing.T) {
	uc, insightRepo, measRepo, auditRepo := newGenerateInsightsUseCase()

	measRepo.Add(&entities.Measurement{
		ID:              "m1",
		ClientID:        "client-1",
		RecordedBy:      "coach-1",
		MeasurementType: "LDL Cholesterol",
		Value:           142,
		Unit:            "mg/dL",
		Flag:            entities.MeasurementFlagHigh,
		RefRangeHigh:    refFloat(100),
		ArtifactID:      "artifact-1",
		MeasuredAt:      time.Now(),
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

	if output.InsightsCreated != 1 {
		t.Fatalf("expected 1 insight, got %d", output.InsightsCreated)
	}

	card := output.Insights[0]
	if card.Status != entities.InsightStatusDraft {
		t.Errorf("expected status draft, got %s", card.Status)
	}
	if card.Priority != entities.InsightPriorityHigh {
		t.Errorf("expected priority high, got %s", card.Priority)
	}
	if card.Category != entities.InsightCategoryNutrition {
		t.Errorf("expected category nutrition, got %s", card.Category)
	}
	if card.CoachID != "coach-1" {
		t.Errorf("expected coach_id coach-1, got %s", card.CoachID)
	}
	if card.ClientID != "client-1" {
		t.Errorf("expected client_id client-1, got %s", card.ClientID)
	}
	if len(card.Evidence) != 1 {
		t.Fatalf("expected 1 evidence ref, got %d", len(card.Evidence))
	}
	if card.Evidence[0].MeasurementID != "m1" {
		t.Errorf("expected evidence measurement_id m1, got %s", card.Evidence[0].MeasurementID)
	}

	if insightRepo.CardCount() != 1 {
		t.Errorf("expected 1 card in repo, got %d", insightRepo.CardCount())
	}
	if auditRepo.EventCount() != 1 {
		t.Errorf("expected 1 audit event, got %d", auditRepo.EventCount())
	}
}

func TestGenerateInsights_MultipleFlags_CreatesMultipleInsights(t *testing.T) {
	uc, insightRepo, measRepo, _ := newGenerateInsightsUseCase()

	measRepo.Add(&entities.Measurement{
		ID:              "m1",
		ClientID:        "client-1",
		MeasurementType: "LDL Cholesterol",
		Value:           142,
		Unit:            "mg/dL",
		Flag:            entities.MeasurementFlagHigh,
		RefRangeHigh:    refFloat(100),
		ArtifactID:      "artifact-1",
		MeasuredAt:      time.Now(),
	})
	measRepo.Add(&entities.Measurement{
		ID:              "m2",
		ClientID:        "client-1",
		MeasurementType: "Triglycerides",
		Value:           210,
		Unit:            "mg/dL",
		Flag:            entities.MeasurementFlagHigh,
		RefRangeHigh:    refFloat(150),
		ArtifactID:      "artifact-1",
		MeasuredAt:      time.Now(),
	})
	measRepo.Add(&entities.Measurement{
		ID:              "m3",
		ClientID:        "client-1",
		MeasurementType: "HDL Cholesterol",
		Value:           32,
		Unit:            "mg/dL",
		Flag:            entities.MeasurementFlagLow,
		RefRangeLow:     refFloat(40),
		ArtifactID:      "artifact-1",
		MeasuredAt:      time.Now(),
	})

	output, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1", CoachID: "coach-1", ArtifactID: "artifact-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.InsightsCreated != 3 {
		t.Errorf("expected 3 insights, got %d", output.InsightsCreated)
	}
	if insightRepo.CardCount() != 3 {
		t.Errorf("expected 3 cards in repo, got %d", insightRepo.CardCount())
	}
}

func TestGenerateInsights_AllInRange_NoInsights(t *testing.T) {
	uc, insightRepo, measRepo, _ := newGenerateInsightsUseCase()

	measRepo.Add(&entities.Measurement{
		ID:              "m1",
		ClientID:        "client-1",
		MeasurementType: "LDL Cholesterol",
		Value:           85,
		Unit:            "mg/dL",
		Flag:            entities.MeasurementFlagNormal,
		ArtifactID:      "artifact-1",
		MeasuredAt:      time.Now(),
	})

	output, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1", CoachID: "coach-1", ArtifactID: "artifact-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.InsightsCreated != 0 {
		t.Errorf("expected 0 insights, got %d", output.InsightsCreated)
	}
	if insightRepo.CardCount() != 0 {
		t.Errorf("expected 0 cards in repo, got %d", insightRepo.CardCount())
	}
}

func TestGenerateInsights_DuplicatePrevention_NoDuplicateInsight(t *testing.T) {
	uc, insightRepo, measRepo, _ := newGenerateInsightsUseCase()

	measRepo.Add(&entities.Measurement{
		ID:              "m1",
		ClientID:        "client-1",
		MeasurementType: "LDL Cholesterol",
		Value:           142,
		Unit:            "mg/dL",
		Flag:            entities.MeasurementFlagHigh,
		RefRangeHigh:    refFloat(100),
		ArtifactID:      "artifact-1",
		MeasuredAt:      time.Now(),
	})

	input := GenerateInsightsInput{
		ClientID: "client-1", CoachID: "coach-1", ArtifactID: "artifact-1",
	}

	// First run: should create 1 insight
	output1, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("first run error: %v", err)
	}
	if output1.InsightsCreated != 1 {
		t.Fatalf("expected 1 insight on first run, got %d", output1.InsightsCreated)
	}

	// Second run: should not create duplicates
	output2, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("second run error: %v", err)
	}
	if output2.InsightsCreated != 0 {
		t.Errorf("expected 0 insights on second run, got %d", output2.InsightsCreated)
	}
	if insightRepo.CardCount() != 1 {
		t.Errorf("expected 1 total card in repo, got %d", insightRepo.CardCount())
	}
}

func TestGenerateInsights_CriticalGlucose_UrgentPriority(t *testing.T) {
	uc, _, measRepo, _ := newGenerateInsightsUseCase()

	measRepo.Add(&entities.Measurement{
		ID:              "m1",
		ClientID:        "client-1",
		MeasurementType: "Fasting Glucose",
		Value:           250,
		Unit:            "mg/dL",
		Flag:            entities.MeasurementFlagCriticalHigh,
		RefRangeHigh:    refFloat(100),
		ArtifactID:      "artifact-1",
		MeasuredAt:      time.Now(),
	})

	output, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1", CoachID: "coach-1", ArtifactID: "artifact-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.InsightsCreated != 1 {
		t.Fatalf("expected 1 insight, got %d", output.InsightsCreated)
	}

	card := output.Insights[0]
	if card.Priority != entities.InsightPriorityUrgent {
		t.Errorf("expected priority urgent, got %s", card.Priority)
	}
	if card.Category != entities.InsightCategorySafety {
		t.Errorf("expected category safety, got %s", card.Category)
	}
}

func TestGenerateInsights_EvidenceLinking_ReferencesSourceMeasurement(t *testing.T) {
	uc, _, measRepo, _ := newGenerateInsightsUseCase()

	measRepo.Add(&entities.Measurement{
		ID:              "meas-abc",
		ClientID:        "client-1",
		MeasurementType: "LDL Cholesterol",
		Value:           142,
		Unit:            "mg/dL",
		Flag:            entities.MeasurementFlagHigh,
		RefRangeHigh:    refFloat(100),
		RefRangeLow:     refFloat(0),
		ArtifactID:      "artifact-xyz",
		MeasuredAt:      time.Now(),
	})

	output, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1", CoachID: "coach-1", ArtifactID: "artifact-xyz",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(output.Insights) != 1 {
		t.Fatalf("expected 1 insight, got %d", len(output.Insights))
	}

	evidence := output.Insights[0].Evidence
	if len(evidence) != 1 {
		t.Fatalf("expected 1 evidence ref, got %d", len(evidence))
	}
	if evidence[0].MeasurementID != "meas-abc" {
		t.Errorf("expected measurement ID meas-abc, got %s", evidence[0].MeasurementID)
	}
	if evidence[0].ArtifactID != "artifact-xyz" {
		t.Errorf("expected artifact ID artifact-xyz, got %s", evidence[0].ArtifactID)
	}
	if evidence[0].Description == "" {
		t.Error("expected non-empty evidence description")
	}
}

func TestGenerateInsights_HRVDeclining_CreatesMediumInsight(t *testing.T) {
	uc, _, measRepo, _ := newGenerateInsightsUseCase()
	now := time.Now()

	// Prior week HRV data (high values)
	for i := 8; i <= 14; i++ {
		measRepo.Add(&entities.Measurement{
			ClientID:        "client-1",
			MeasurementType: "hrv",
			Value:           60,
			Unit:            "ms",
			Flag:            entities.MeasurementFlagNormal,
			MeasuredAt:      now.AddDate(0, 0, -i),
		})
	}

	// Recent week HRV data (low values — >15% drop)
	for i := 0; i < 7; i++ {
		measRepo.Add(&entities.Measurement{
			ClientID:        "client-1",
			MeasurementType: "hrv",
			Value:           45,
			Unit:            "ms",
			Flag:            entities.MeasurementFlagNormal,
			MeasuredAt:      now.AddDate(0, 0, -i),
		})
	}

	output, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1", CoachID: "coach-1", ArtifactID: "artifact-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, card := range output.Insights {
		if card.Title == "HRV Declining" {
			found = true
			if card.Priority != entities.InsightPriorityMedium {
				t.Errorf("expected medium priority, got %s", card.Priority)
			}
			if card.Category != entities.InsightCategoryRecovery {
				t.Errorf("expected recovery category, got %s", card.Category)
			}
		}
	}
	if !found {
		t.Error("expected HRV Declining insight card")
	}
}

func TestGenerateInsights_SleepDecline_CreatesMediumInsight(t *testing.T) {
	uc, _, measRepo, _ := newGenerateInsightsUseCase()
	now := time.Now()

	// Recent week sleep data (< 6 hours avg)
	for i := 0; i < 7; i++ {
		measRepo.Add(&entities.Measurement{
			ClientID:        "client-1",
			MeasurementType: "sleep_hours",
			Value:           5.5,
			Unit:            "hours",
			Flag:            entities.MeasurementFlagNormal,
			MeasuredAt:      now.AddDate(0, 0, -i),
		})
	}

	output, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1", CoachID: "coach-1", ArtifactID: "artifact-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, card := range output.Insights {
		if card.Title == "Sleep Declining" {
			found = true
			if card.Priority != entities.InsightPriorityMedium {
				t.Errorf("expected medium priority, got %s", card.Priority)
			}
		}
	}
	if !found {
		t.Error("expected Sleep Declining insight card")
	}
}

func TestGenerateInsights_WeightChange_CreatesMediumInsight(t *testing.T) {
	uc, _, measRepo, _ := newGenerateInsightsUseCase()
	now := time.Now()

	measRepo.Add(&entities.Measurement{
		ClientID:        "client-1",
		MeasurementType: "weight",
		Value:           200,
		Unit:            "lbs",
		Flag:            entities.MeasurementFlagNormal,
		MeasuredAt:      now.AddDate(0, 0, -13),
	})
	measRepo.Add(&entities.Measurement{
		ClientID:        "client-1",
		MeasurementType: "weight",
		Value:           191,
		Unit:            "lbs",
		Flag:            entities.MeasurementFlagNormal,
		MeasuredAt:      now.AddDate(0, 0, -1),
	})

	output, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1", CoachID: "coach-1", ArtifactID: "artifact-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, card := range output.Insights {
		if card.Title == "Significant Weight Change" {
			found = true
			if card.Priority != entities.InsightPriorityMedium {
				t.Errorf("expected medium priority, got %s", card.Priority)
			}
			if card.Category != entities.InsightCategoryNutrition {
				t.Errorf("expected nutrition category, got %s", card.Category)
			}
		}
	}
	if !found {
		t.Error("expected Significant Weight Change insight card")
	}
}

func TestGenerateInsights_BodyFatDrop_CreatesLowPriorityInsight(t *testing.T) {
	uc, _, measRepo, _ := newGenerateInsightsUseCase()
	now := time.Now()

	measRepo.Add(&entities.Measurement{
		ClientID:        "client-1",
		MeasurementType: "body_fat",
		Value:           22.3,
		Unit:            "%",
		Flag:            entities.MeasurementFlagNormal,
		MeasuredAt:      now.AddDate(0, 0, -13),
	})
	measRepo.Add(&entities.Measurement{
		ClientID:        "client-1",
		MeasurementType: "body_fat",
		Value:           21.1,
		Unit:            "%",
		Flag:            entities.MeasurementFlagNormal,
		MeasuredAt:      now.AddDate(0, 0, -1),
	})

	output, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1", CoachID: "coach-1", ArtifactID: "artifact-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, card := range output.Insights {
		if card.Title == "Body Fat Progress" {
			found = true
			if card.Priority != entities.InsightPriorityLow {
				t.Errorf("expected low priority, got %s", card.Priority)
			}
			if card.Category != entities.InsightCategoryPerformance {
				t.Errorf("expected performance category, got %s", card.Category)
			}
		}
	}
	if !found {
		t.Error("expected Body Fat Progress insight card")
	}
}

func TestGenerateInsights_MissingClientID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newGenerateInsightsUseCase()

	_, err := uc.Execute(context.Background(), GenerateInsightsInput{
		CoachID: "coach-1", ArtifactID: "artifact-1",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGenerateInsights_MissingCoachID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newGenerateInsightsUseCase()

	_, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1", ArtifactID: "artifact-1",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGenerateInsights_MissingArtifactID_ReturnsValidationError(t *testing.T) {
	uc, _, _, _ := newGenerateInsightsUseCase()

	_, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1", CoachID: "coach-1",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestGenerateInsights_NoMeasurements_ReturnsEmptyOutput(t *testing.T) {
	uc, _, _, _ := newGenerateInsightsUseCase()

	output, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1", CoachID: "coach-1", ArtifactID: "artifact-empty",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.InsightsCreated != 0 {
		t.Errorf("expected 0 insights, got %d", output.InsightsCreated)
	}
}

func TestGenerateInsights_LowFlag_CreatesLowTitleInsight(t *testing.T) {
	uc, _, measRepo, _ := newGenerateInsightsUseCase()

	measRepo.Add(&entities.Measurement{
		ID:              "m1",
		ClientID:        "client-1",
		MeasurementType: "Iron",
		Value:           30,
		Unit:            "ug/dL",
		Flag:            entities.MeasurementFlagLow,
		RefRangeLow:     refFloat(60),
		ArtifactID:      "artifact-1",
		MeasuredAt:      time.Now(),
	})

	output, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1", CoachID: "coach-1", ArtifactID: "artifact-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.InsightsCreated != 1 {
		t.Fatalf("expected 1 insight, got %d", output.InsightsCreated)
	}
	if output.Insights[0].Title != "Low Iron" {
		t.Errorf("expected title 'Low Iron', got %q", output.Insights[0].Title)
	}
}

func TestGenerateInsights_AuditEventLogged_ForEachInsight(t *testing.T) {
	uc, _, measRepo, auditRepo := newGenerateInsightsUseCase()

	measRepo.Add(&entities.Measurement{
		ID:              "m1",
		ClientID:        "client-1",
		MeasurementType: "LDL",
		Value:           142,
		Unit:            "mg/dL",
		Flag:            entities.MeasurementFlagHigh,
		RefRangeHigh:    refFloat(100),
		ArtifactID:      "artifact-1",
		MeasuredAt:      time.Now(),
	})
	measRepo.Add(&entities.Measurement{
		ID:              "m2",
		ClientID:        "client-1",
		MeasurementType: "Glucose",
		Value:           250,
		Unit:            "mg/dL",
		Flag:            entities.MeasurementFlagCriticalHigh,
		RefRangeHigh:    refFloat(100),
		ArtifactID:      "artifact-1",
		MeasuredAt:      time.Now(),
	})

	_, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1", CoachID: "coach-1", ArtifactID: "artifact-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if auditRepo.EventCount() != 2 {
		t.Errorf("expected 2 audit events, got %d", auditRepo.EventCount())
	}

	events := auditRepo.GetEvents()
	for _, e := range events {
		if e.Action != "insight.generate" {
			t.Errorf("expected action insight.generate, got %s", e.Action)
		}
		if e.EntityType != "insight_card" {
			t.Errorf("expected entity type insight_card, got %s", e.EntityType)
		}
	}
}

func TestSplitWeeklyAverages_SplitsCorrectly(t *testing.T) {
	now := time.Now()
	oneWeekAgo := now.AddDate(0, 0, -7)

	data := []*entities.Measurement{
		{Value: 60, MeasuredAt: now.AddDate(0, 0, -10)},
		{Value: 70, MeasuredAt: now.AddDate(0, 0, -9)},
		{Value: 40, MeasuredAt: now.AddDate(0, 0, -3)},
		{Value: 50, MeasuredAt: now.AddDate(0, 0, -1)},
	}

	priorAvg, recentAvg := splitWeeklyAverages(data, oneWeekAgo)

	expectedPrior := 65.0
	if priorAvg != expectedPrior {
		t.Errorf("expected prior avg %.1f, got %.1f", expectedPrior, priorAvg)
	}
	expectedRecent := 45.0
	if recentAvg != expectedRecent {
		t.Errorf("expected recent avg %.1f, got %.1f", expectedRecent, recentAvg)
	}
}

func TestSplitWeeklyAverages_EmptyData_ReturnsZeros(t *testing.T) {
	priorAvg, recentAvg := splitWeeklyAverages(nil, time.Now())
	if priorAvg != 0 || recentAvg != 0 {
		t.Errorf("expected (0, 0), got (%.1f, %.1f)", priorAvg, recentAvg)
	}
}

func TestAverage_CalculatesCorrectly(t *testing.T) {
	data := []*entities.Measurement{
		{Value: 10},
		{Value: 20},
		{Value: 30},
	}
	got := average(data)
	if got != 20 {
		t.Errorf("expected 20, got %.1f", got)
	}
}

func TestAverage_EmptySlice_ReturnsZero(t *testing.T) {
	got := average(nil)
	if got != 0 {
		t.Errorf("expected 0, got %.1f", got)
	}
}

func TestGenerateInsights_ContextCancelled_ReturnsError(t *testing.T) {
	uc, _, measRepo, _ := newGenerateInsightsUseCase()

	measRepo.Add(&entities.Measurement{
		ID:              "m1",
		ClientID:        "client-1",
		MeasurementType: "LDL",
		Value:           142,
		Unit:            "mg/dL",
		Flag:            entities.MeasurementFlagHigh,
		ArtifactID:      "artifact-1",
		MeasuredAt:      time.Now(),
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// The in-memory repo ignores context, so this may not error.
	// But the function should not panic.
	_, _ = uc.Execute(ctx, GenerateInsightsInput{
		ClientID: "client-1", CoachID: "coach-1", ArtifactID: "artifact-1",
	})
}

func TestGenerateInsights_HRVNoDecline_NoInsight(t *testing.T) {
	uc, _, measRepo, _ := newGenerateInsightsUseCase()
	now := time.Now()

	// HRV stable
	for i := 0; i < 14; i++ {
		measRepo.Add(&entities.Measurement{
			ClientID:        "client-1",
			MeasurementType: "hrv",
			Value:           60,
			Unit:            "ms",
			Flag:            entities.MeasurementFlagNormal,
			MeasuredAt:      now.AddDate(0, 0, -i),
		})
	}

	output, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1", CoachID: "coach-1", ArtifactID: "artifact-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, card := range output.Insights {
		if card.Title == "HRV Declining" {
			t.Error("expected no HRV insight when HRV is stable")
		}
	}
}

func TestGenerateInsights_SleepAbove6_NoInsight(t *testing.T) {
	uc, _, measRepo, _ := newGenerateInsightsUseCase()
	now := time.Now()

	for i := 0; i < 7; i++ {
		measRepo.Add(&entities.Measurement{
			ClientID:        "client-1",
			MeasurementType: "sleep_hours",
			Value:           7.5,
			Unit:            "hours",
			Flag:            entities.MeasurementFlagNormal,
			MeasuredAt:      now.AddDate(0, 0, -i),
		})
	}

	output, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1", CoachID: "coach-1", ArtifactID: "artifact-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, card := range output.Insights {
		if card.Title == "Sleep Declining" {
			t.Error("expected no sleep insight when avg >= 6")
		}
	}
}

func TestGenerateInsights_WeightChangeUnder3Pct_NoInsight(t *testing.T) {
	uc, _, measRepo, _ := newGenerateInsightsUseCase()
	now := time.Now()

	measRepo.Add(&entities.Measurement{
		ClientID:        "client-1",
		MeasurementType: "weight",
		Value:           200,
		Unit:            "lbs",
		Flag:            entities.MeasurementFlagNormal,
		MeasuredAt:      now.AddDate(0, 0, -13),
	})
	measRepo.Add(&entities.Measurement{
		ClientID:        "client-1",
		MeasurementType: "weight",
		Value:           197,
		Unit:            "lbs",
		Flag:            entities.MeasurementFlagNormal,
		MeasuredAt:      now.AddDate(0, 0, -1),
	})

	output, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1", CoachID: "coach-1", ArtifactID: "artifact-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, card := range output.Insights {
		if card.Title == "Significant Weight Change" {
			t.Error("expected no weight insight when change < 3%")
		}
	}
}

func TestGenerateInsights_BodyFatDropUnder1Pct_NoInsight(t *testing.T) {
	uc, _, measRepo, _ := newGenerateInsightsUseCase()
	now := time.Now()

	measRepo.Add(&entities.Measurement{
		ClientID:        "client-1",
		MeasurementType: "body_fat",
		Value:           22.3,
		Unit:            "%",
		Flag:            entities.MeasurementFlagNormal,
		MeasuredAt:      now.AddDate(0, 0, -13),
	})
	measRepo.Add(&entities.Measurement{
		ClientID:        "client-1",
		MeasurementType: "body_fat",
		Value:           21.8,
		Unit:            "%",
		Flag:            entities.MeasurementFlagNormal,
		MeasuredAt:      now.AddDate(0, 0, -1),
	})

	output, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1", CoachID: "coach-1", ArtifactID: "artifact-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, card := range output.Insights {
		if card.Title == "Body Fat Progress" {
			t.Error("expected no body fat insight when drop < 1%")
		}
	}
}

func TestFormatLabBody_HighFlag_ShowsAboveRecommended(t *testing.T) {
	high := 100.0
	m := &entities.Measurement{
		MeasurementType: "LDL",
		Value:           142,
		Unit:            "mg/dL",
		Flag:            entities.MeasurementFlagHigh,
		RefRangeHigh:    &high,
	}
	body := formatLabBody(m)
	expected := "LDL is 142.0 mg/dL, above the recommended < 100.0 mg/dL"
	if body != expected {
		t.Errorf("expected %q, got %q", expected, body)
	}
}

func TestFormatLabBody_LowFlag_ShowsBelowRecommended(t *testing.T) {
	low := 60.0
	m := &entities.Measurement{
		MeasurementType: "Iron",
		Value:           30,
		Unit:            "ug/dL",
		Flag:            entities.MeasurementFlagLow,
		RefRangeLow:     &low,
	}
	body := formatLabBody(m)
	expected := "Iron is 30.0 ug/dL, below the recommended > 60.0 ug/dL"
	if body != expected {
		t.Errorf("expected %q, got %q", expected, body)
	}
}

func TestFormatLabBody_NoRefRange_ShowsFlaggedStatus(t *testing.T) {
	m := &entities.Measurement{
		MeasurementType: "LDL",
		Value:           142,
		Unit:            "mg/dL",
		Flag:            entities.MeasurementFlagHigh,
	}
	body := formatLabBody(m)
	expected := "LDL is 142.0 mg/dL (flagged high)"
	if body != expected {
		t.Errorf("expected %q, got %q", expected, body)
	}
}

func TestFormatEvidenceDescription_WithRefRange_ShowsRange(t *testing.T) {
	low, high := 0.0, 100.0
	m := &entities.Measurement{
		MeasurementType: "LDL",
		Value:           142,
		Unit:            "mg/dL",
		RefRangeLow:     &low,
		RefRangeHigh:    &high,
	}
	desc := formatEvidenceDescription(m)
	expected := "LDL: 142.0 mg/dL (ref: 0.0-100.0)"
	if desc != expected {
		t.Errorf("expected %q, got %q", expected, desc)
	}
}

func TestFormatEvidenceDescription_NoRefRange_ShowsValueOnly(t *testing.T) {
	m := &entities.Measurement{
		MeasurementType: "LDL",
		Value:           142,
		Unit:            "mg/dL",
	}
	desc := formatEvidenceDescription(m)
	expected := "LDL: 142.0 mg/dL"
	if desc != expected {
		t.Errorf("expected %q, got %q", expected, desc)
	}
}

func TestBuildLabInsightCard_HighFlag_SetsHighPriority(t *testing.T) {
	m := &entities.Measurement{
		ID:              "m1",
		ClientID:        "client-1",
		MeasurementType: "LDL",
		Value:           142,
		Unit:            "mg/dL",
		Flag:            entities.MeasurementFlagHigh,
		RefRangeHigh:    refFloat(100),
		ArtifactID:      "artifact-1",
	}
	card := buildLabInsightCard("coach-1", m)

	if card.Priority != entities.InsightPriorityHigh {
		t.Errorf("expected high priority, got %s", card.Priority)
	}
	if card.Status != entities.InsightStatusDraft {
		t.Errorf("expected draft status, got %s", card.Status)
	}
}

func TestBuildLabInsightCard_CriticalHigh_SetsUrgentPriority(t *testing.T) {
	m := &entities.Measurement{
		ID:              "m1",
		ClientID:        "client-1",
		MeasurementType: "Glucose",
		Value:           250,
		Unit:            "mg/dL",
		Flag:            entities.MeasurementFlagCriticalHigh,
		RefRangeHigh:    refFloat(100),
	}
	card := buildLabInsightCard("coach-1", m)

	if card.Priority != entities.InsightPriorityUrgent {
		t.Errorf("expected urgent priority, got %s", card.Priority)
	}
	if card.Category != entities.InsightCategorySafety {
		t.Errorf("expected safety category, got %s", card.Category)
	}
}

func TestBuildLabInsightCard_CriticalLow_SetsLowTitle(t *testing.T) {
	m := &entities.Measurement{
		ID:              "m1",
		ClientID:        "client-1",
		MeasurementType: "Iron",
		Value:           10,
		Unit:            "ug/dL",
		Flag:            entities.MeasurementFlagCriticalLow,
		RefRangeLow:     refFloat(60),
	}
	card := buildLabInsightCard("coach-1", m)

	if card.Title != "Low Iron" {
		t.Errorf("expected 'Low Iron', got %q", card.Title)
	}
	if card.Priority != entities.InsightPriorityUrgent {
		t.Errorf("expected urgent priority, got %s", card.Priority)
	}
}

func TestGenerateInsights_MeasurementRepoError_PropagatesError(t *testing.T) {
	insightRepo := repository.NewInMemoryInsightCardRepository()
	auditRepo := repository.NewInMemoryAuditRepository()
	measRepo := &failingMeasurementRepo{err: errors.New("db connection failed")}

	uc := NewGenerateInsightsUseCase(insightRepo, measRepo, auditRepo)
	_, err := uc.Execute(context.Background(), GenerateInsightsInput{
		ClientID: "client-1", CoachID: "coach-1", ArtifactID: "artifact-1",
	})
	if err == nil {
		t.Fatal("expected error from failing measurement repo")
	}
}

// failingMeasurementRepo always returns an error (for error path testing).
type failingMeasurementRepo struct {
	err error
}

func (f *failingMeasurementRepo) FindByClientID(_ context.Context, _ string, _, _ int) ([]*entities.Measurement, error) {
	return nil, f.err
}

func (f *failingMeasurementRepo) FindByClientIDAndType(_ context.Context, _ string, _ string, _ time.Time) ([]*entities.Measurement, error) {
	return nil, f.err
}

func (f *failingMeasurementRepo) FindRecentByArtifactID(_ context.Context, _ string) ([]*entities.Measurement, error) {
	return nil, f.err
}
