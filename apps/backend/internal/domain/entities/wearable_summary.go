package entities

import "time"

// WearableSummary represents aggregated data from a wearable device for a single day.
type WearableSummary struct {
	ID          string                 `json:"id"`
	ClientID    string                 `json:"client_id"`
	Source      string                 `json:"source"`
	SummaryDate string                 `json:"summary_date"`
	Metrics     map[string]interface{} `json:"metrics"`
	SyncedAt    time.Time              `json:"synced_at"`
	CreatedAt   time.Time              `json:"created_at"`
}
