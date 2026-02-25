package entities

import "time"

// InsightCategory categorizes the type of insight.
type InsightCategory string

const (
	InsightCategoryGeneral     InsightCategory = "general"
	InsightCategoryNutrition   InsightCategory = "nutrition"
	InsightCategoryRecovery    InsightCategory = "recovery"
	InsightCategoryPerformance InsightCategory = "performance"
	InsightCategoryBehavior    InsightCategory = "behavior"
	InsightCategorySafety      InsightCategory = "safety"
)

// InsightPriority defines the urgency level of an insight.
type InsightPriority string

const (
	InsightPriorityLow    InsightPriority = "low"
	InsightPriorityMedium InsightPriority = "medium"
	InsightPriorityHigh   InsightPriority = "high"
	InsightPriorityUrgent InsightPriority = "urgent"
)

// InsightStatus represents the lifecycle state of an insight card.
type InsightStatus string

const (
	InsightStatusDraft     InsightStatus = "draft"
	InsightStatusApproved  InsightStatus = "approved"
	InsightStatusDismissed InsightStatus = "dismissed"
	InsightStatusShared    InsightStatus = "shared"
)

// EvidenceRef links an insight card to the source data it was derived from.
type EvidenceRef struct {
	MeasurementID string `json:"measurement_id"`
	ArtifactID    string `json:"artifact_id,omitempty"`
	Description   string `json:"description"`
}

// InsightCard represents an AI-generated coaching insight that requires coach approval.
type InsightCard struct {
	ID        string          `json:"id"`
	CoachID   string          `json:"coach_id"`
	ClientID  string          `json:"client_id"`
	Title     string          `json:"title"`
	Body      string          `json:"body"`
	Category  InsightCategory `json:"category"`
	Priority  InsightPriority `json:"priority"`
	Status    InsightStatus   `json:"status"`
	Evidence  []EvidenceRef   `json:"evidence"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// IsValidInsightCategory returns true if the category is one of the known types.
func IsValidInsightCategory(c InsightCategory) bool {
	switch c {
	case InsightCategoryGeneral, InsightCategoryNutrition, InsightCategoryRecovery,
		InsightCategoryPerformance, InsightCategoryBehavior, InsightCategorySafety:
		return true
	}
	return false
}

// IsValidInsightPriority returns true if the priority is one of the known levels.
func IsValidInsightPriority(p InsightPriority) bool {
	switch p {
	case InsightPriorityLow, InsightPriorityMedium, InsightPriorityHigh, InsightPriorityUrgent:
		return true
	}
	return false
}

// IsValidInsightStatus returns true if the status is one of the known states.
func IsValidInsightStatus(s InsightStatus) bool {
	switch s {
	case InsightStatusDraft, InsightStatusApproved, InsightStatusDismissed, InsightStatusShared:
		return true
	}
	return false
}
