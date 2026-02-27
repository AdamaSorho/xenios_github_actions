package entities

import (
	"fmt"
	"time"
)

// InsightStatus represents the lifecycle status of an InsightCard.
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

// ValidInsightStatuses enumerates all valid insight statuses.
var ValidInsightStatuses = map[InsightStatus]bool{
	InsightStatusDraft:     true,
	InsightStatusApproved:  true,
	InsightStatusDismissed: true,
	InsightStatusShared:    true,
}

// ValidInsightPriorities enumerates all valid insight priorities.
var ValidInsightPriorities = map[InsightPriority]bool{
	InsightPriorityUrgent: true,
	InsightPriorityHigh:   true,
	InsightPriorityMedium: true,
	InsightPriorityLow:    true,
}

// priorityOrder defines sort order for priorities (lower = higher priority).
var priorityOrder = map[InsightPriority]int{
	InsightPriorityUrgent: 0,
	InsightPriorityHigh:   1,
	InsightPriorityMedium: 2,
	InsightPriorityLow:    3,
}

// PriorityRank returns the numeric rank of a priority for sorting.
func PriorityRank(p InsightPriority) int {
	if rank, ok := priorityOrder[p]; ok {
		return rank
	}
	return 99
}

// Evidence links an InsightCard to a measurement or data point.
type Evidence struct {
	MeasurementID string `json:"measurement_id"`
	Description   string `json:"description"`
}

// InsightCard represents an AI-generated insight for a client.
type InsightCard struct {
	ID          string          `json:"id"`
	ClientID    string          `json:"client_id"`
	CoachID     string          `json:"coach_id"`
	ClientName  string          `json:"client_name"`
	Title       string          `json:"title"`
	Body        string          `json:"body"`
	Category    string          `json:"category"`
	Priority    InsightPriority `json:"priority"`
	Status      InsightStatus   `json:"status"`
	Evidence    []Evidence      `json:"evidence,omitempty"`
	ApprovedAt  *time.Time      `json:"approved_at,omitempty"`
	DismissedAt *time.Time      `json:"dismissed_at,omitempty"`
	SharedAt    *time.Time      `json:"shared_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// validTransitions defines the allowed status transitions.
var validTransitions = map[InsightStatus]map[InsightStatus]bool{
	InsightStatusDraft:    {InsightStatusApproved: true, InsightStatusDismissed: true},
	InsightStatusApproved: {InsightStatusShared: true},
}

// CanTransition checks whether a status transition is valid.
func CanTransition(from, to InsightStatus) bool {
	targets, ok := validTransitions[from]
	if !ok {
		return false
	}
	return targets[to]
}

// ValidateTransition returns an error if the status transition is invalid.
func ValidateTransition(from, to InsightStatus) error {
	if !CanTransition(from, to) {
		return fmt.Errorf("invalid status transition from %q to %q", from, to)
	}
	return nil
}

// IsVisibleToClient returns true if the insight should be visible to clients.
func (ic *InsightCard) IsVisibleToClient() bool {
	return ic.Status == InsightStatusApproved || ic.Status == InsightStatusShared
}
