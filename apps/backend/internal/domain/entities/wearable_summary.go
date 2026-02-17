package entities

import "time"

// WearableSummary represents aggregated data from a wearable device for a single day.
type WearableSummary struct {
	ID          string                 `json:"id"`
	ClientID    string                 `json:"client_id"`
	Source      string                 `json:"source"`
	SummaryDate time.Time              `json:"summary_date"`
	Metrics     map[string]interface{} `json:"metrics"`
	SyncedAt    time.Time              `json:"synced_at"`
	CreatedAt   time.Time              `json:"created_at"`
}

// WearableSummaryFilter holds query parameters for listing wearable summaries.
type WearableSummaryFilter struct {
	ClientID string
	Source   string
	From     *time.Time
	To       *time.Time
	Limit    int
	Offset   int
}

// WearableAverages holds computed rolling averages from wearable data.
type WearableAverages struct {
	Source        string   `json:"source"`
	AvgHRV7d     *float64 `json:"avg_hrv_7d,omitempty"`
	AvgSleep7d   *float64 `json:"avg_sleep_7d,omitempty"`
	AvgRecovery7d *float64 `json:"avg_recovery_7d,omitempty"`
}
