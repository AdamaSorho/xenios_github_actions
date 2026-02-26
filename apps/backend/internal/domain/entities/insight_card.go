package entities

import "time"

// InsightStatus represents the lifecycle state of an InsightCard.
type InsightStatus string

const (
	InsightStatusDraft    InsightStatus = "draft"
	InsightStatusApproved InsightStatus = "approved"
	InsightStatusDismissed InsightStatus = "dismissed"
	InsightStatusShared   InsightStatus = "shared"
)

// InsightPriority represents the urgency level of an InsightCard.
type InsightPriority string

const (
	InsightPriorityUrgent InsightPriority = "urgent"
	InsightPriorityHigh   InsightPriority = "high"
	InsightPriorityMedium InsightPriority = "medium"
	InsightPriorityLow    InsightPriority = "low"
)

// InsightCategory represents the topic area of an InsightCard.
type InsightCategory string

const (
	InsightCategoryNutrition InsightCategory = "nutrition"
	InsightCategoryExercise  InsightCategory = "exercise"
	InsightCategorySleep     InsightCategory = "sleep"
	InsightCategoryStress    InsightCategory = "stress"
	InsightCategoryGeneral   InsightCategory = "general"
)

// InsightEvidence links an insight to the measurement that supports it.
type InsightEvidence struct {
	MeasurementID string `json:"measurement_id"`
	Description   string `json:"description"`
}

// InsightCard represents an AI-generated insight for a client, subject to coach approval.
type InsightCard struct {
	ID          string            `json:"id"`
	ClientID    string            `json:"client_id"`
	CoachID     string            `json:"coach_id"`
	ClientName  string            `json:"client_name"`
	Title       string            `json:"title"`
	Body        string            `json:"body"`
	Category    InsightCategory   `json:"category"`
	Priority    InsightPriority   `json:"priority"`
	Status      InsightStatus     `json:"status"`
	Evidence    []InsightEvidence `json:"evidence,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	ApprovedAt  *time.Time        `json:"approved_at,omitempty"`
	DismissedAt *time.Time        `json:"dismissed_at,omitempty"`
	SharedAt    *time.Time        `json:"shared_at,omitempty"`
}

// validTransitions maps each status to its allowed next statuses.
var validTransitions = map[InsightStatus][]InsightStatus{
	InsightStatusDraft:    {InsightStatusApproved, InsightStatusDismissed},
	InsightStatusApproved: {InsightStatusShared},
}

// CanTransitionTo returns true if the transition from the current status to the target is valid.
func (c *InsightCard) CanTransitionTo(target InsightStatus) bool {
	allowed, ok := validTransitions[c.Status]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == target {
			return true
		}
	}
	return false
}

// IsValidInsightStatus checks whether a status string is a known InsightStatus.
func IsValidInsightStatus(s string) bool {
	switch InsightStatus(s) {
	case InsightStatusDraft, InsightStatusApproved, InsightStatusDismissed, InsightStatusShared:
		return true
	}
	return false
}

// IsValidInsightPriority checks whether a priority string is a known InsightPriority.
func IsValidInsightPriority(p string) bool {
	switch InsightPriority(p) {
	case InsightPriorityUrgent, InsightPriorityHigh, InsightPriorityMedium, InsightPriorityLow:
		return true
	}
	return false
}

// IsValidInsightCategory checks whether a category string is a known InsightCategory.
func IsValidInsightCategory(c string) bool {
	switch InsightCategory(c) {
	case InsightCategoryNutrition, InsightCategoryExercise, InsightCategorySleep, InsightCategoryStress, InsightCategoryGeneral:
		return true
	}
	return false
}

// IsTerminalStatus returns true if the status is a terminal state (dismissed, shared).
func (s InsightStatus) IsTerminal() bool {
	switch s {
	case InsightStatusDismissed, InsightStatusShared:
		return true
	}
	return false
}

// PrioritySortOrder returns a numeric sort value for priority (lower = higher priority).
func (p InsightPriority) SortOrder() int {
	switch p {
	case InsightPriorityUrgent:
		return 0
	case InsightPriorityHigh:
		return 1
	case InsightPriorityMedium:
		return 2
	case InsightPriorityLow:
		return 3
	default:
		return 4
	}
}

// InsightQueryFilter holds query parameters for listing insight cards.
type InsightQueryFilter struct {
	CoachID  string
	ClientID string
	Status   string
	Page     int
	Limit    int
}
