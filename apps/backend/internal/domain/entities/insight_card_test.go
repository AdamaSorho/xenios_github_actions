package entities

import (
	"testing"
	"time"
)

func TestCanTransitionTo_DraftToApproved_ReturnsTrue(t *testing.T) {
	card := &InsightCard{Status: InsightStatusDraft}
	if !card.CanTransitionTo(InsightStatusApproved) {
		t.Error("expected draft -> approved to be valid")
	}
}

func TestCanTransitionTo_DraftToDismissed_ReturnsTrue(t *testing.T) {
	card := &InsightCard{Status: InsightStatusDraft}
	if !card.CanTransitionTo(InsightStatusDismissed) {
		t.Error("expected draft -> dismissed to be valid")
	}
}

func TestCanTransitionTo_ApprovedToShared_ReturnsTrue(t *testing.T) {
	card := &InsightCard{Status: InsightStatusApproved}
	if !card.CanTransitionTo(InsightStatusShared) {
		t.Error("expected approved -> shared to be valid")
	}
}

func TestCanTransitionTo_DismissedToAny_ReturnsFalse(t *testing.T) {
	card := &InsightCard{Status: InsightStatusDismissed}
	for _, target := range []InsightStatus{InsightStatusDraft, InsightStatusApproved, InsightStatusShared} {
		if card.CanTransitionTo(target) {
			t.Errorf("expected dismissed -> %s to be invalid", target)
		}
	}
}

func TestCanTransitionTo_SharedToAny_ReturnsFalse(t *testing.T) {
	card := &InsightCard{Status: InsightStatusShared}
	for _, target := range []InsightStatus{InsightStatusDraft, InsightStatusApproved, InsightStatusDismissed} {
		if card.CanTransitionTo(target) {
			t.Errorf("expected shared -> %s to be invalid", target)
		}
	}
}

func TestCanTransitionTo_DraftToShared_ReturnsFalse(t *testing.T) {
	card := &InsightCard{Status: InsightStatusDraft}
	if card.CanTransitionTo(InsightStatusShared) {
		t.Error("expected draft -> shared to be invalid")
	}
}

func TestCanTransitionTo_ApprovedToDraft_ReturnsFalse(t *testing.T) {
	card := &InsightCard{Status: InsightStatusApproved}
	if card.CanTransitionTo(InsightStatusDraft) {
		t.Error("expected approved -> draft to be invalid")
	}
}

func TestIsValidInsightStatus_ValidStatuses_ReturnsTrue(t *testing.T) {
	for _, s := range []string{"draft", "approved", "dismissed", "shared"} {
		if !IsValidInsightStatus(s) {
			t.Errorf("expected %q to be valid", s)
		}
	}
}

func TestIsValidInsightStatus_InvalidStatus_ReturnsFalse(t *testing.T) {
	if IsValidInsightStatus("invalid") {
		t.Error("expected 'invalid' to be invalid status")
	}
}

func TestIsValidInsightPriority_ValidPriorities_ReturnsTrue(t *testing.T) {
	for _, p := range []string{"urgent", "high", "medium", "low"} {
		if !IsValidInsightPriority(p) {
			t.Errorf("expected %q to be valid priority", p)
		}
	}
}

func TestIsValidInsightPriority_InvalidPriority_ReturnsFalse(t *testing.T) {
	if IsValidInsightPriority("invalid") {
		t.Error("expected 'invalid' to be invalid priority")
	}
}

func TestIsValidInsightCategory_ValidCategories_ReturnsTrue(t *testing.T) {
	for _, c := range []string{"nutrition", "exercise", "sleep", "stress", "general"} {
		if !IsValidInsightCategory(c) {
			t.Errorf("expected %q to be valid category", c)
		}
	}
}

func TestIsValidInsightCategory_InvalidCategory_ReturnsFalse(t *testing.T) {
	if IsValidInsightCategory("invalid") {
		t.Error("expected 'invalid' to be invalid category")
	}
}

func TestInsightStatus_IsTerminal_DismissedAndShared(t *testing.T) {
	if !InsightStatusDismissed.IsTerminal() {
		t.Error("expected dismissed to be terminal")
	}
	if !InsightStatusShared.IsTerminal() {
		t.Error("expected shared to be terminal")
	}
}

func TestInsightStatus_IsTerminal_DraftAndApprovedAreNot(t *testing.T) {
	if InsightStatusDraft.IsTerminal() {
		t.Error("expected draft to not be terminal")
	}
	if InsightStatusApproved.IsTerminal() {
		t.Error("expected approved to not be terminal")
	}
}

func TestInsightPriority_SortOrder(t *testing.T) {
	if InsightPriorityUrgent.SortOrder() >= InsightPriorityHigh.SortOrder() {
		t.Error("urgent should sort before high")
	}
	if InsightPriorityHigh.SortOrder() >= InsightPriorityMedium.SortOrder() {
		t.Error("high should sort before medium")
	}
	if InsightPriorityMedium.SortOrder() >= InsightPriorityLow.SortOrder() {
		t.Error("medium should sort before low")
	}
}

func TestInsightPriority_SortOrder_UnknownPriority(t *testing.T) {
	p := InsightPriority("unknown")
	if p.SortOrder() <= InsightPriorityLow.SortOrder() {
		t.Error("unknown priority should sort after low")
	}
}

func TestInsightCard_Fields(t *testing.T) {
	now := time.Now()
	card := &InsightCard{
		ID:         "insight-1",
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ClientName: "Test Client",
		Title:      "Test Title",
		Body:       "Test body text",
		Category:   InsightCategoryNutrition,
		Priority:   InsightPriorityHigh,
		Status:     InsightStatusDraft,
		Evidence: []InsightEvidence{
			{MeasurementID: "m-1", Description: "LDL high"},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if card.ID != "insight-1" {
		t.Errorf("expected ID 'insight-1', got %s", card.ID)
	}
	if card.ApprovedAt != nil {
		t.Error("expected ApprovedAt to be nil for draft")
	}
	if len(card.Evidence) != 1 {
		t.Errorf("expected 1 evidence item, got %d", len(card.Evidence))
	}
}
