package entities

import (
	"errors"
	"testing"
)

func TestInsightCard_Validate_ValidCard_ReturnsNil(t *testing.T) {
	card := &InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Elevated LDL",
		Body:     "LDL-C is 142 mg/dL",
		Category: InsightCategoryNutrition,
		Priority: InsightPriorityHigh,
		Evidence: []EvidenceRef{{MeasurementID: "m-1", Description: "LDL: 142"}},
	}

	if err := card.Validate(); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestInsightCard_Validate_MissingCoachID_ReturnsError(t *testing.T) {
	card := &InsightCard{
		ClientID: "client-1",
		Title:    "Test",
		Body:     "Test",
		Category: InsightCategoryNutrition,
		Priority: InsightPriorityHigh,
		Evidence: []EvidenceRef{{MeasurementID: "m-1", Description: "test"}},
	}

	err := card.Validate()
	if err == nil {
		t.Fatal("expected error for missing coach_id")
	}
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestInsightCard_Validate_MissingClientID_ReturnsError(t *testing.T) {
	card := &InsightCard{
		CoachID:  "coach-1",
		Title:    "Test",
		Body:     "Test",
		Category: InsightCategoryNutrition,
		Priority: InsightPriorityHigh,
		Evidence: []EvidenceRef{{MeasurementID: "m-1", Description: "test"}},
	}

	err := card.Validate()
	if err == nil {
		t.Fatal("expected error for missing client_id")
	}
}

func TestInsightCard_Validate_MissingTitle_ReturnsError(t *testing.T) {
	card := &InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Body:     "Test",
		Category: InsightCategoryNutrition,
		Priority: InsightPriorityHigh,
		Evidence: []EvidenceRef{{MeasurementID: "m-1", Description: "test"}},
	}

	err := card.Validate()
	if err == nil {
		t.Fatal("expected error for missing title")
	}
}

func TestInsightCard_Validate_MissingBody_ReturnsError(t *testing.T) {
	card := &InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Test",
		Category: InsightCategoryNutrition,
		Priority: InsightPriorityHigh,
		Evidence: []EvidenceRef{{MeasurementID: "m-1", Description: "test"}},
	}

	err := card.Validate()
	if err == nil {
		t.Fatal("expected error for missing body")
	}
}

func TestInsightCard_Validate_InvalidCategory_ReturnsError(t *testing.T) {
	card := &InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Test",
		Body:     "Test",
		Category: "invalid",
		Priority: InsightPriorityHigh,
		Evidence: []EvidenceRef{{MeasurementID: "m-1", Description: "test"}},
	}

	err := card.Validate()
	if err == nil {
		t.Fatal("expected error for invalid category")
	}
}

func TestInsightCard_Validate_InvalidPriority_ReturnsError(t *testing.T) {
	card := &InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Test",
		Body:     "Test",
		Category: InsightCategoryNutrition,
		Priority: "invalid",
		Evidence: []EvidenceRef{{MeasurementID: "m-1", Description: "test"}},
	}

	err := card.Validate()
	if err == nil {
		t.Fatal("expected error for invalid priority")
	}
}

func TestInsightCard_Validate_EmptyEvidence_ReturnsError(t *testing.T) {
	card := &InsightCard{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Title:    "Test",
		Body:     "Test",
		Category: InsightCategoryNutrition,
		Priority: InsightPriorityHigh,
		Evidence: []EvidenceRef{},
	}

	err := card.Validate()
	if err == nil {
		t.Fatal("expected error for empty evidence")
	}
}

func TestIsValidInsightCategory_AllValidCategories(t *testing.T) {
	validCategories := []InsightCategory{
		InsightCategoryNutrition,
		InsightCategoryRecovery,
		InsightCategoryPerformance,
		InsightCategorySafety,
	}

	for _, c := range validCategories {
		if !IsValidInsightCategory(c) {
			t.Errorf("expected %q to be valid", c)
		}
	}
}

func TestIsValidInsightCategory_InvalidCategory(t *testing.T) {
	if IsValidInsightCategory("invalid") {
		t.Error("expected 'invalid' to be invalid category")
	}
}

func TestIsValidInsightPriority_AllValidPriorities(t *testing.T) {
	validPriorities := []InsightPriority{
		InsightPriorityLow,
		InsightPriorityMedium,
		InsightPriorityHigh,
		InsightPriorityUrgent,
	}

	for _, p := range validPriorities {
		if !IsValidInsightPriority(p) {
			t.Errorf("expected %q to be valid", p)
		}
	}
}

func TestIsValidInsightPriority_InvalidPriority(t *testing.T) {
	if IsValidInsightPriority("invalid") {
		t.Error("expected 'invalid' to be invalid priority")
	}
}
