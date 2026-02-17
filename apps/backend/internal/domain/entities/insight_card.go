package entities

import (
	"fmt"
	"time"
)

// Insight status constants.
const (
	InsightStatusDraft     = "draft"
	InsightStatusApproved  = "approved"
	InsightStatusDismissed = "dismissed"
	InsightStatusShared    = "shared"
)

// ValidInsightStatuses enumerates the known insight card statuses.
var ValidInsightStatuses = map[string]bool{
	InsightStatusDraft:     true,
	InsightStatusApproved:  true,
	InsightStatusDismissed: true,
	InsightStatusShared:    true,
}

// IsValidInsightStatus checks if a status string is valid.
func IsValidInsightStatus(status string) bool {
	return ValidInsightStatuses[status]
}

// ValidInsightCategories enumerates the known insight categories.
var ValidInsightCategories = map[string]bool{
	"general":     true,
	"nutrition":   true,
	"recovery":    true,
	"performance": true,
	"behavior":    true,
	"safety":      true,
}

// IsValidInsightCategory checks if a category string is valid.
func IsValidInsightCategory(cat string) bool {
	return ValidInsightCategories[cat]
}

// ValidInsightPriorities enumerates the known insight priorities.
var ValidInsightPriorities = map[string]bool{
	"low":    true,
	"medium": true,
	"high":   true,
	"urgent": true,
}

// IsValidInsightPriority checks if a priority string is valid.
func IsValidInsightPriority(pri string) bool {
	return ValidInsightPriorities[pri]
}

// InsightPriorityRank returns a numeric rank for sorting by priority (higher = more urgent).
func InsightPriorityRank(priority string) int {
	switch priority {
	case "urgent":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

// InsightEvidence represents a piece of evidence supporting an insight.
type InsightEvidence struct {
	MeasurementID string `json:"measurementId"`
	Description   string `json:"description"`
}

// InsightCard represents an AI-generated coaching insight that requires coach approval.
type InsightCard struct {
	ID         string            `json:"id"`
	CoachID    string            `json:"coachId"`
	ClientID   string            `json:"clientId"`
	SessionID  *string           `json:"sessionId,omitempty"`
	Title      string            `json:"title"`
	Body       string            `json:"body"`
	Category   string            `json:"category"`
	Status     string            `json:"status"`
	Priority   string            `json:"priority"`
	Evidence   []InsightEvidence `json:"evidence,omitempty"`
	ApprovedAt *time.Time        `json:"approvedAt,omitempty"`
	SharedAt   *time.Time        `json:"sharedAt,omitempty"`
	CreatedAt  time.Time         `json:"createdAt"`
	UpdatedAt  time.Time         `json:"updatedAt"`
}

// NewInsightCard creates a new InsightCard with draft status.
func NewInsightCard(coachID, clientID, title, body, category, priority string) *InsightCard {
	now := time.Now()
	return &InsightCard{
		CoachID:   coachID,
		ClientID:  clientID,
		Title:     title,
		Body:      body,
		Category:  category,
		Status:    InsightStatusDraft,
		Priority:  priority,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Approve transitions the insight from draft to approved.
func (ic *InsightCard) Approve() error {
	if ic.Status != InsightStatusDraft {
		return fmt.Errorf("cannot approve insight with status %q: only draft insights can be approved", ic.Status)
	}
	now := time.Now()
	ic.Status = InsightStatusApproved
	ic.ApprovedAt = &now
	ic.UpdatedAt = now
	return nil
}

// Dismiss transitions the insight from draft to dismissed.
func (ic *InsightCard) Dismiss() error {
	if ic.Status != InsightStatusDraft {
		return fmt.Errorf("cannot dismiss insight with status %q: only draft insights can be dismissed", ic.Status)
	}
	ic.Status = InsightStatusDismissed
	ic.UpdatedAt = time.Now()
	return nil
}

// Share transitions the insight from approved to shared.
func (ic *InsightCard) Share() error {
	if ic.Status != InsightStatusApproved {
		return fmt.Errorf("cannot share insight with status %q: only approved insights can be shared", ic.Status)
	}
	now := time.Now()
	ic.Status = InsightStatusShared
	ic.SharedAt = &now
	ic.UpdatedAt = now
	return nil
}

// UpdateText updates the title and body of the insight.
// Only allowed for draft or approved insights.
func (ic *InsightCard) UpdateText(title, body string) error {
	if title == "" {
		return fmt.Errorf("title is required")
	}
	if body == "" {
		return fmt.Errorf("body is required")
	}
	if ic.Status == InsightStatusDismissed || ic.Status == InsightStatusShared {
		return fmt.Errorf("cannot edit insight with status %q", ic.Status)
	}
	ic.Title = title
	ic.Body = body
	ic.UpdatedAt = time.Now()
	return nil
}
