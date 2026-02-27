package entities

import (
	"testing"
	"time"
)

func TestCanTransition_DraftToApproved_ReturnsTrue(t *testing.T) {
	if !CanTransition(InsightStatusDraft, InsightStatusApproved) {
		t.Error("expected draft → approved to be valid")
	}
}

func TestCanTransition_DraftToDismissed_ReturnsTrue(t *testing.T) {
	if !CanTransition(InsightStatusDraft, InsightStatusDismissed) {
		t.Error("expected draft → dismissed to be valid")
	}
}

func TestCanTransition_ApprovedToShared_ReturnsTrue(t *testing.T) {
	if !CanTransition(InsightStatusApproved, InsightStatusShared) {
		t.Error("expected approved → shared to be valid")
	}
}

func TestCanTransition_DismissedToApproved_ReturnsFalse(t *testing.T) {
	if CanTransition(InsightStatusDismissed, InsightStatusApproved) {
		t.Error("expected dismissed → approved to be invalid")
	}
}

func TestCanTransition_SharedToApproved_ReturnsFalse(t *testing.T) {
	if CanTransition(InsightStatusShared, InsightStatusApproved) {
		t.Error("expected shared → approved to be invalid")
	}
}

func TestCanTransition_DraftToShared_ReturnsFalse(t *testing.T) {
	if CanTransition(InsightStatusDraft, InsightStatusShared) {
		t.Error("expected draft → shared to be invalid (must go through approved)")
	}
}

func TestCanTransition_ApprovedToDismissed_ReturnsFalse(t *testing.T) {
	if CanTransition(InsightStatusApproved, InsightStatusDismissed) {
		t.Error("expected approved → dismissed to be invalid")
	}
}

func TestValidateTransition_ValidTransition_ReturnsNil(t *testing.T) {
	err := ValidateTransition(InsightStatusDraft, InsightStatusApproved)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateTransition_InvalidTransition_ReturnsError(t *testing.T) {
	err := ValidateTransition(InsightStatusDismissed, InsightStatusApproved)
	if err == nil {
		t.Error("expected error for invalid transition")
	}
}

func TestInsightCard_IsVisibleToClient_Approved_ReturnsTrue(t *testing.T) {
	ic := &InsightCard{Status: InsightStatusApproved}
	if !ic.IsVisibleToClient() {
		t.Error("expected approved insight to be visible to client")
	}
}

func TestInsightCard_IsVisibleToClient_Shared_ReturnsTrue(t *testing.T) {
	ic := &InsightCard{Status: InsightStatusShared}
	if !ic.IsVisibleToClient() {
		t.Error("expected shared insight to be visible to client")
	}
}

func TestInsightCard_IsVisibleToClient_Draft_ReturnsFalse(t *testing.T) {
	ic := &InsightCard{Status: InsightStatusDraft}
	if ic.IsVisibleToClient() {
		t.Error("expected draft insight to NOT be visible to client")
	}
}

func TestInsightCard_IsVisibleToClient_Dismissed_ReturnsFalse(t *testing.T) {
	ic := &InsightCard{Status: InsightStatusDismissed}
	if ic.IsVisibleToClient() {
		t.Error("expected dismissed insight to NOT be visible to client")
	}
}

func TestPriorityRank_UrgentIsHighest(t *testing.T) {
	if PriorityRank(InsightPriorityUrgent) >= PriorityRank(InsightPriorityHigh) {
		t.Error("expected urgent to rank higher (lower number) than high")
	}
}

func TestPriorityRank_Ordering(t *testing.T) {
	urgent := PriorityRank(InsightPriorityUrgent)
	high := PriorityRank(InsightPriorityHigh)
	medium := PriorityRank(InsightPriorityMedium)
	low := PriorityRank(InsightPriorityLow)

	if !(urgent < high && high < medium && medium < low) {
		t.Errorf("expected urgent < high < medium < low, got %d, %d, %d, %d", urgent, high, medium, low)
	}
}

func TestPriorityRank_UnknownPriority_ReturnsHighNumber(t *testing.T) {
	rank := PriorityRank(InsightPriority("unknown"))
	if rank < 10 {
		t.Errorf("expected unknown priority to have high rank, got %d", rank)
	}
}

func TestValidInsightStatuses_ContainsAllStatuses(t *testing.T) {
	statuses := []InsightStatus{InsightStatusDraft, InsightStatusApproved, InsightStatusDismissed, InsightStatusShared}
	for _, s := range statuses {
		if !ValidInsightStatuses[s] {
			t.Errorf("expected %q to be a valid status", s)
		}
	}
}

func TestValidInsightPriorities_ContainsAllPriorities(t *testing.T) {
	priorities := []InsightPriority{InsightPriorityUrgent, InsightPriorityHigh, InsightPriorityMedium, InsightPriorityLow}
	for _, p := range priorities {
		if !ValidInsightPriorities[p] {
			t.Errorf("expected %q to be a valid priority", p)
		}
	}
}

func TestInsightCard_TimestampFields(t *testing.T) {
	now := time.Now()
	ic := &InsightCard{
		ApprovedAt: &now,
		SharedAt:   nil,
	}
	if ic.ApprovedAt == nil {
		t.Error("expected ApprovedAt to be set")
	}
	if ic.SharedAt != nil {
		t.Error("expected SharedAt to be nil")
	}
}
