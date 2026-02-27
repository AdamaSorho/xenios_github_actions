package entities

import "testing"

func TestIsValidInsightCategory_ValidCategories_ReturnsTrue(t *testing.T) {
	valid := []InsightCategory{
		InsightCategoryNutrition, InsightCategoryRecovery,
		InsightCategoryPerformance, InsightCategorySafety, InsightCategoryGeneral,
	}
	for _, c := range valid {
		if !IsValidInsightCategory(c) {
			t.Errorf("expected %q to be valid", c)
		}
	}
}

func TestIsValidInsightCategory_InvalidCategory_ReturnsFalse(t *testing.T) {
	if IsValidInsightCategory("invalid") {
		t.Error("expected invalid category to return false")
	}
}

func TestIsValidInsightPriority_ValidPriorities_ReturnsTrue(t *testing.T) {
	valid := []InsightPriority{
		InsightPriorityLow, InsightPriorityMedium,
		InsightPriorityHigh, InsightPriorityUrgent,
	}
	for _, p := range valid {
		if !IsValidInsightPriority(p) {
			t.Errorf("expected %q to be valid", p)
		}
	}
}

func TestIsValidInsightPriority_InvalidPriority_ReturnsFalse(t *testing.T) {
	if IsValidInsightPriority("invalid") {
		t.Error("expected invalid priority to return false")
	}
}

func TestIsValidInsightStatus_ValidStatuses_ReturnsTrue(t *testing.T) {
	valid := []InsightStatus{
		InsightStatusDraft, InsightStatusApproved,
		InsightStatusDismissed, InsightStatusShared,
	}
	for _, s := range valid {
		if !IsValidInsightStatus(s) {
			t.Errorf("expected %q to be valid", s)
		}
	}
}

func TestIsValidInsightStatus_InvalidStatus_ReturnsFalse(t *testing.T) {
	if IsValidInsightStatus("invalid") {
		t.Error("expected invalid status to return false")
	}
}
