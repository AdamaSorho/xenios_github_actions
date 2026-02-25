package entities

import (
	"testing"
)

func TestInsightCard_TransitionTo_DraftToApproved_Success(t *testing.T) {
	ic := &InsightCard{Status: InsightStatusDraft}
	err := ic.TransitionTo(InsightStatusApproved)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ic.Status != InsightStatusApproved {
		t.Errorf("expected status %q, got %q", InsightStatusApproved, ic.Status)
	}
	if ic.ApprovedAt == nil {
		t.Error("expected approved_at to be set")
	}
}

func TestInsightCard_TransitionTo_DraftToDismissed_Success(t *testing.T) {
	ic := &InsightCard{Status: InsightStatusDraft}
	err := ic.TransitionTo(InsightStatusDismissed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ic.Status != InsightStatusDismissed {
		t.Errorf("expected status %q, got %q", InsightStatusDismissed, ic.Status)
	}
}

func TestInsightCard_TransitionTo_ApprovedToShared_Success(t *testing.T) {
	ic := &InsightCard{Status: InsightStatusApproved}
	err := ic.TransitionTo(InsightStatusShared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ic.Status != InsightStatusShared {
		t.Errorf("expected status %q, got %q", InsightStatusShared, ic.Status)
	}
	if ic.SharedAt == nil {
		t.Error("expected shared_at to be set")
	}
}

func TestInsightCard_TransitionTo_DismissedToApproved_Error(t *testing.T) {
	ic := &InsightCard{Status: InsightStatusDismissed}
	err := ic.TransitionTo(InsightStatusApproved)
	if err == nil {
		t.Fatal("expected error for invalid transition")
	}
	ste, ok := err.(*StatusTransitionError)
	if !ok {
		t.Fatalf("expected StatusTransitionError, got %T", err)
	}
	if ste.From != InsightStatusDismissed || ste.To != InsightStatusApproved {
		t.Errorf("unexpected transition error: from=%q to=%q", ste.From, ste.To)
	}
}

func TestInsightCard_TransitionTo_SharedToApproved_Error(t *testing.T) {
	ic := &InsightCard{Status: InsightStatusShared}
	err := ic.TransitionTo(InsightStatusApproved)
	if err == nil {
		t.Fatal("expected error for invalid transition")
	}
}

func TestInsightCard_TransitionTo_DraftToShared_Error(t *testing.T) {
	ic := &InsightCard{Status: InsightStatusDraft}
	err := ic.TransitionTo(InsightStatusShared)
	if err == nil {
		t.Fatal("expected error for invalid transition from draft to shared")
	}
}

func TestInsightCard_TransitionTo_ApprovedToDismissed_Error(t *testing.T) {
	ic := &InsightCard{Status: InsightStatusApproved}
	err := ic.TransitionTo(InsightStatusDismissed)
	if err == nil {
		t.Fatal("expected error for invalid transition from approved to dismissed")
	}
}

func TestInsightCard_CanTransitionTo_ValidTransitions(t *testing.T) {
	tests := []struct {
		from, to string
		expected bool
	}{
		{InsightStatusDraft, InsightStatusApproved, true},
		{InsightStatusDraft, InsightStatusDismissed, true},
		{InsightStatusApproved, InsightStatusShared, true},
		{InsightStatusDraft, InsightStatusShared, false},
		{InsightStatusDismissed, InsightStatusApproved, false},
		{InsightStatusShared, InsightStatusDraft, false},
	}

	for _, tt := range tests {
		ic := &InsightCard{Status: tt.from}
		if got := ic.CanTransitionTo(tt.to); got != tt.expected {
			t.Errorf("CanTransitionTo(%q→%q) = %v, want %v", tt.from, tt.to, got, tt.expected)
		}
	}
}

func TestInsightCard_IsVisibleToClient(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{InsightStatusDraft, false},
		{InsightStatusApproved, true},
		{InsightStatusDismissed, false},
		{InsightStatusShared, true},
	}

	for _, tt := range tests {
		ic := &InsightCard{Status: tt.status}
		if got := ic.IsVisibleToClient(); got != tt.expected {
			t.Errorf("IsVisibleToClient(%q) = %v, want %v", tt.status, got, tt.expected)
		}
	}
}

func TestIsValidInsightCategory(t *testing.T) {
	valid := []string{"general", "nutrition", "recovery", "performance", "behavior", "safety"}
	for _, c := range valid {
		if !IsValidInsightCategory(c) {
			t.Errorf("expected %q to be valid", c)
		}
	}
	if IsValidInsightCategory("invalid") {
		t.Error("expected 'invalid' to be invalid")
	}
}

func TestIsValidInsightPriority(t *testing.T) {
	valid := []string{"low", "medium", "high", "urgent"}
	for _, p := range valid {
		if !IsValidInsightPriority(p) {
			t.Errorf("expected %q to be valid", p)
		}
	}
	if IsValidInsightPriority("critical") {
		t.Error("expected 'critical' to be invalid")
	}
}
