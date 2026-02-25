package entities

import (
	"fmt"
	"time"
)

// InsightCard represents an AI-generated coaching insight that requires
// coach approval before being visible to clients.
type InsightCard struct {
	ID         string     `json:"id"`
	CoachID    string     `json:"coach_id"`
	ClientID   string     `json:"client_id"`
	ClientName string     `json:"client_name,omitempty"`
	SessionID  string     `json:"session_id,omitempty"`
	Title      string     `json:"title"`
	Body       string     `json:"body"`
	Category   string     `json:"category"`
	Status     string     `json:"status"`
	Priority   string     `json:"priority"`
	Evidence   []Evidence `json:"evidence,omitempty"`
	ApprovedAt *time.Time `json:"approved_at,omitempty"`
	SharedAt   *time.Time `json:"shared_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// Evidence links an insight to a specific measurement or data point.
type Evidence struct {
	MeasurementID string `json:"measurement_id"`
	Description   string `json:"description"`
}

// InsightCard status constants.
const (
	InsightStatusDraft     = "draft"
	InsightStatusApproved  = "approved"
	InsightStatusDismissed = "dismissed"
	InsightStatusShared    = "shared"
)

// InsightCard category constants.
const (
	InsightCategoryGeneral     = "general"
	InsightCategoryNutrition   = "nutrition"
	InsightCategoryRecovery    = "recovery"
	InsightCategoryPerformance = "performance"
	InsightCategoryBehavior    = "behavior"
	InsightCategorySafety      = "safety"
)

// InsightCard priority constants.
const (
	InsightPriorityLow    = "low"
	InsightPriorityMedium = "medium"
	InsightPriorityHigh   = "high"
	InsightPriorityUrgent = "urgent"
)

// validStatusTransitions defines allowed status transitions.
// Key: current status, Value: set of valid target statuses.
var validStatusTransitions = map[string]map[string]bool{
	InsightStatusDraft: {
		InsightStatusApproved:  true,
		InsightStatusDismissed: true,
	},
	InsightStatusApproved: {
		InsightStatusShared: true,
	},
	// dismissed and shared are terminal states
}

// StatusTransitionError indicates an invalid status transition was attempted.
type StatusTransitionError struct {
	From string
	To   string
}

func (e *StatusTransitionError) Error() string {
	return fmt.Sprintf("invalid status transition from %q to %q", e.From, e.To)
}

// CanTransitionTo checks if a transition from the current status to the target is valid.
func (ic *InsightCard) CanTransitionTo(target string) bool {
	allowed, ok := validStatusTransitions[ic.Status]
	if !ok {
		return false
	}
	return allowed[target]
}

// TransitionTo attempts to change the insight card's status.
// Returns a StatusTransitionError if the transition is not allowed.
func (ic *InsightCard) TransitionTo(target string) error {
	if !ic.CanTransitionTo(target) {
		return &StatusTransitionError{From: ic.Status, To: target}
	}
	ic.Status = target
	now := time.Now()
	ic.UpdatedAt = now
	switch target {
	case InsightStatusApproved:
		ic.ApprovedAt = &now
	case InsightStatusShared:
		ic.SharedAt = &now
	}
	return nil
}

// IsVisibleToClient returns true if the client should be able to see this insight.
func (ic *InsightCard) IsVisibleToClient() bool {
	return ic.Status == InsightStatusApproved || ic.Status == InsightStatusShared
}

// IsValidCategory checks if the given category is valid.
func IsValidInsightCategory(category string) bool {
	switch category {
	case InsightCategoryGeneral, InsightCategoryNutrition, InsightCategoryRecovery,
		InsightCategoryPerformance, InsightCategoryBehavior, InsightCategorySafety:
		return true
	}
	return false
}

// IsValidInsightPriority checks if the given priority is valid.
func IsValidInsightPriority(priority string) bool {
	switch priority {
	case InsightPriorityLow, InsightPriorityMedium, InsightPriorityHigh, InsightPriorityUrgent:
		return true
	}
	return false
}
