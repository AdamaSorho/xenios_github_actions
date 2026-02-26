package entities

import "time"

// InsightCategory classifies an insight card.
type InsightCategory string

const (
	InsightCategoryNutrition   InsightCategory = "nutrition"
	InsightCategoryRecovery    InsightCategory = "recovery"
	InsightCategoryPerformance InsightCategory = "performance"
	InsightCategorySafety      InsightCategory = "safety"
	InsightCategoryGeneral     InsightCategory = "general"
)

// InsightPriority indicates the urgency of an insight.
type InsightPriority string

const (
	InsightPriorityLow    InsightPriority = "low"
	InsightPriorityMedium InsightPriority = "medium"
	InsightPriorityHigh   InsightPriority = "high"
	InsightPriorityUrgent InsightPriority = "urgent"
)

// InsightStatus tracks the lifecycle of an insight card.
type InsightStatus string

const (
	InsightStatusDraft     InsightStatus = "draft"
	InsightStatusApproved  InsightStatus = "approved"
	InsightStatusDismissed InsightStatus = "dismissed"
	InsightStatusShared    InsightStatus = "shared"
)

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

// EvidenceRef links an insight to the source measurement or artifact it was derived from.
type EvidenceRef struct {
	MeasurementID string `json:"measurement_id"`
	ArtifactID    string `json:"artifact_id"`
	Description   string `json:"description"`
}

// IsValidInsightCategory returns true if the category is valid.
func IsValidInsightCategory(c InsightCategory) bool {
	switch c {
	case InsightCategoryNutrition, InsightCategoryRecovery,
		InsightCategoryPerformance, InsightCategorySafety, InsightCategoryGeneral:
		return true
	}
	return false
}

// IsValidInsightPriority returns true if the priority is valid.
func IsValidInsightPriority(p InsightPriority) bool {
	switch p {
	case InsightPriorityLow, InsightPriorityMedium,
		InsightPriorityHigh, InsightPriorityUrgent:
		return true
	}
	return false
}

// IsValidInsightStatus returns true if the status is valid.
func IsValidInsightStatus(s InsightStatus) bool {
	switch s {
	case InsightStatusDraft, InsightStatusApproved,
		InsightStatusDismissed, InsightStatusShared:
		return true
	}
	return false
}
