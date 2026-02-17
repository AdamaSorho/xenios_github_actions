package entities

import "time"

// InsightCategory represents the category of an insight.
type InsightCategory string

const (
	InsightCategoryNutrition   InsightCategory = "nutrition"
	InsightCategoryRecovery    InsightCategory = "recovery"
	InsightCategoryPerformance InsightCategory = "performance"
	InsightCategorySafety      InsightCategory = "safety"
)

// IsValidInsightCategory returns true if the category is valid.
func IsValidInsightCategory(c InsightCategory) bool {
	switch c {
	case InsightCategoryNutrition, InsightCategoryRecovery,
		InsightCategoryPerformance, InsightCategorySafety:
		return true
	}
	return false
}

// InsightPriority represents the priority level of an insight.
type InsightPriority string

const (
	InsightPriorityLow    InsightPriority = "low"
	InsightPriorityMedium InsightPriority = "medium"
	InsightPriorityHigh   InsightPriority = "high"
	InsightPriorityUrgent InsightPriority = "urgent"
)

// IsValidInsightPriority returns true if the priority is valid.
func IsValidInsightPriority(p InsightPriority) bool {
	switch p {
	case InsightPriorityLow, InsightPriorityMedium,
		InsightPriorityHigh, InsightPriorityUrgent:
		return true
	}
	return false
}

// InsightStatus represents the lifecycle status of an insight card.
type InsightStatus string

const (
	InsightStatusDraft    InsightStatus = "draft"
	InsightStatusApproved InsightStatus = "approved"
	InsightStatusRejected InsightStatus = "rejected"
)

// EvidenceRef links an insight card to the source measurement or artifact it was derived from.
type EvidenceRef struct {
	MeasurementID string `json:"measurement_id"`
	ArtifactID    string `json:"artifact_id,omitempty"`
	Description   string `json:"description"`
}

// InsightCard represents an AI-generated draft insight for a coach to review.
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

// Validate checks that all required fields are present and valid.
func (ic *InsightCard) Validate() error {
	if ic.CoachID == "" {
		return NewValidationError("coach_id is required")
	}
	if ic.ClientID == "" {
		return NewValidationError("client_id is required")
	}
	if ic.Title == "" {
		return NewValidationError("title is required")
	}
	if ic.Body == "" {
		return NewValidationError("body is required")
	}
	if !IsValidInsightCategory(ic.Category) {
		return NewValidationError("invalid category: %q", ic.Category)
	}
	if !IsValidInsightPriority(ic.Priority) {
		return NewValidationError("invalid priority: %q", ic.Priority)
	}
	if len(ic.Evidence) == 0 {
		return NewValidationError("at least one evidence reference is required")
	}
	return nil
}

// InsightTriggerType identifies what rule triggered an insight.
type InsightTriggerType string

const (
	TriggerLabOutOfRange    InsightTriggerType = "lab_out_of_range"
	TriggerLabCritical      InsightTriggerType = "lab_critical"
	TriggerHRVDeclining     InsightTriggerType = "hrv_declining"
	TriggerSleepDeclining   InsightTriggerType = "sleep_declining"
	TriggerWeightChange     InsightTriggerType = "weight_change"
	TriggerBodyFatProgress  InsightTriggerType = "body_fat_progress"
)
