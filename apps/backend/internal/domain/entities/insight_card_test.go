package entities

import (
	"testing"
	"time"
)

func TestNewInsightCard_Defaults(t *testing.T) {
	ic := NewInsightCard("coach-1", "client-1", "High LDL", "LDL is 142 mg/dL", "nutrition", "high")
	if ic.CoachID != "coach-1" {
		t.Errorf("expected CoachID coach-1, got %s", ic.CoachID)
	}
	if ic.ClientID != "client-1" {
		t.Errorf("expected ClientID client-1, got %s", ic.ClientID)
	}
	if ic.Title != "High LDL" {
		t.Errorf("expected title 'High LDL', got %s", ic.Title)
	}
	if ic.Body != "LDL is 142 mg/dL" {
		t.Errorf("expected body 'LDL is 142 mg/dL', got %s", ic.Body)
	}
	if ic.Category != "nutrition" {
		t.Errorf("expected category nutrition, got %s", ic.Category)
	}
	if ic.Priority != "high" {
		t.Errorf("expected priority high, got %s", ic.Priority)
	}
	if ic.Status != InsightStatusDraft {
		t.Errorf("expected status draft, got %s", ic.Status)
	}
	if ic.ApprovedAt != nil {
		t.Error("expected nil ApprovedAt")
	}
	if ic.SharedAt != nil {
		t.Error("expected nil SharedAt")
	}
}

func TestInsightCard_Approve_FromDraft(t *testing.T) {
	ic := NewInsightCard("coach-1", "client-1", "title", "body", "general", "medium")
	err := ic.Approve()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ic.Status != InsightStatusApproved {
		t.Errorf("expected status approved, got %s", ic.Status)
	}
	if ic.ApprovedAt == nil {
		t.Error("expected non-nil ApprovedAt")
	}
}

func TestInsightCard_Approve_FromApproved_Fails(t *testing.T) {
	ic := NewInsightCard("coach-1", "client-1", "title", "body", "general", "medium")
	_ = ic.Approve()
	err := ic.Approve()
	if err == nil {
		t.Fatal("expected error when approving already-approved insight")
	}
}

func TestInsightCard_Approve_FromDismissed_Fails(t *testing.T) {
	ic := NewInsightCard("coach-1", "client-1", "title", "body", "general", "medium")
	_ = ic.Dismiss()
	err := ic.Approve()
	if err == nil {
		t.Fatal("expected error when approving dismissed insight")
	}
}

func TestInsightCard_Dismiss_FromDraft(t *testing.T) {
	ic := NewInsightCard("coach-1", "client-1", "title", "body", "general", "medium")
	err := ic.Dismiss()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ic.Status != InsightStatusDismissed {
		t.Errorf("expected status dismissed, got %s", ic.Status)
	}
}

func TestInsightCard_Dismiss_FromApproved_Fails(t *testing.T) {
	ic := NewInsightCard("coach-1", "client-1", "title", "body", "general", "medium")
	_ = ic.Approve()
	err := ic.Dismiss()
	if err == nil {
		t.Fatal("expected error when dismissing approved insight")
	}
}

func TestInsightCard_Share_FromApproved(t *testing.T) {
	ic := NewInsightCard("coach-1", "client-1", "title", "body", "general", "medium")
	_ = ic.Approve()
	err := ic.Share()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ic.Status != InsightStatusShared {
		t.Errorf("expected status shared, got %s", ic.Status)
	}
	if ic.SharedAt == nil {
		t.Error("expected non-nil SharedAt")
	}
}

func TestInsightCard_Share_FromDraft_Fails(t *testing.T) {
	ic := NewInsightCard("coach-1", "client-1", "title", "body", "general", "medium")
	err := ic.Share()
	if err == nil {
		t.Fatal("expected error when sharing draft insight")
	}
}

func TestInsightCard_Share_FromDismissed_Fails(t *testing.T) {
	ic := NewInsightCard("coach-1", "client-1", "title", "body", "general", "medium")
	_ = ic.Dismiss()
	err := ic.Share()
	if err == nil {
		t.Fatal("expected error when sharing dismissed insight")
	}
}

func TestIsValidInsightStatus(t *testing.T) {
	tests := []struct {
		status string
		valid  bool
	}{
		{"draft", true},
		{"approved", true},
		{"dismissed", true},
		{"shared", true},
		{"pending", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsValidInsightStatus(tt.status); got != tt.valid {
			t.Errorf("IsValidInsightStatus(%q) = %v, want %v", tt.status, got, tt.valid)
		}
	}
}

func TestIsValidInsightCategory(t *testing.T) {
	tests := []struct {
		cat   string
		valid bool
	}{
		{"general", true},
		{"nutrition", true},
		{"recovery", true},
		{"performance", true},
		{"behavior", true},
		{"safety", true},
		{"other", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsValidInsightCategory(tt.cat); got != tt.valid {
			t.Errorf("IsValidInsightCategory(%q) = %v, want %v", tt.cat, got, tt.valid)
		}
	}
}

func TestIsValidInsightPriority(t *testing.T) {
	tests := []struct {
		pri   string
		valid bool
	}{
		{"low", true},
		{"medium", true},
		{"high", true},
		{"urgent", true},
		{"critical", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsValidInsightPriority(tt.pri); got != tt.valid {
			t.Errorf("IsValidInsightPriority(%q) = %v, want %v", tt.pri, got, tt.valid)
		}
	}
}

func TestInsightPriorityRank(t *testing.T) {
	if InsightPriorityRank("urgent") <= InsightPriorityRank("high") {
		t.Error("urgent should rank higher than high")
	}
	if InsightPriorityRank("high") <= InsightPriorityRank("medium") {
		t.Error("high should rank higher than medium")
	}
	if InsightPriorityRank("medium") <= InsightPriorityRank("low") {
		t.Error("medium should rank higher than low")
	}
}

func TestInsightCard_Evidence(t *testing.T) {
	ic := NewInsightCard("coach-1", "client-1", "title", "body", "general", "medium")
	if ic.Evidence != nil {
		t.Error("expected nil evidence by default")
	}
	evidence := []InsightEvidence{
		{MeasurementID: "m-1", Description: "LDL-C: 142 mg/dL"},
	}
	ic.Evidence = evidence
	if len(ic.Evidence) != 1 {
		t.Errorf("expected 1 evidence item, got %d", len(ic.Evidence))
	}
}

func TestInsightCard_UpdateText(t *testing.T) {
	ic := NewInsightCard("coach-1", "client-1", "title", "body", "general", "medium")
	before := ic.UpdatedAt
	// Ensure time passes for UpdatedAt comparison
	time.Sleep(time.Millisecond)
	err := ic.UpdateText("new title", "new body")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ic.Title != "new title" {
		t.Errorf("expected title 'new title', got %s", ic.Title)
	}
	if ic.Body != "new body" {
		t.Errorf("expected body 'new body', got %s", ic.Body)
	}
	if !ic.UpdatedAt.After(before) {
		t.Error("expected UpdatedAt to be updated")
	}
}

func TestInsightCard_UpdateText_EmptyTitle_Fails(t *testing.T) {
	ic := NewInsightCard("coach-1", "client-1", "title", "body", "general", "medium")
	err := ic.UpdateText("", "new body")
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestInsightCard_UpdateText_EmptyBody_Fails(t *testing.T) {
	ic := NewInsightCard("coach-1", "client-1", "title", "body", "general", "medium")
	err := ic.UpdateText("new title", "")
	if err == nil {
		t.Fatal("expected error for empty body")
	}
}

func TestInsightCard_UpdateText_AfterDismissed_Fails(t *testing.T) {
	ic := NewInsightCard("coach-1", "client-1", "title", "body", "general", "medium")
	_ = ic.Dismiss()
	err := ic.UpdateText("new title", "new body")
	if err == nil {
		t.Fatal("expected error when editing dismissed insight")
	}
}

func TestInsightCard_UpdateText_AfterShared_Fails(t *testing.T) {
	ic := NewInsightCard("coach-1", "client-1", "title", "body", "general", "medium")
	_ = ic.Approve()
	_ = ic.Share()
	err := ic.UpdateText("new title", "new body")
	if err == nil {
		t.Fatal("expected error when editing shared insight")
	}
}
