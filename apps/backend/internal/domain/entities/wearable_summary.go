package entities

import "time"

// WearableSummary represents aggregated wearable data for a client on a given date.
type WearableSummary struct {
	ID          string                 `json:"id"`
	ClientID    string                 `json:"client_id"`
	Source      string                 `json:"source"`
	SummaryDate string                 `json:"summary_date"`
	Metrics     map[string]interface{} `json:"metrics"`
	SyncedAt    time.Time              `json:"synced_at"`
	CreatedAt   time.Time              `json:"created_at"`
}

// ProfileSummary is a consolidated view of a client's health data.
type ProfileSummary struct {
	BodyComposition map[string]*LatestMeasurement `json:"bodyComposition"`
	Labs            *LabSummary                   `json:"labs"`
	Wearable        *WearableAverages             `json:"wearable"`
	Nutrition       *NutritionAverages            `json:"nutrition"`
}

// LabSummary contains a summary of lab results.
type LabSummary struct {
	FlaggedCount int                  `json:"flaggedCount"`
	LastTestDate *string              `json:"lastTestDate"`
	Markers      []*LatestMeasurement `json:"markers"`
}

// WearableAverages contains rolling averages from wearable data.
type WearableAverages struct {
	Source        *string  `json:"source"`
	AvgHRV7d      *float64 `json:"avgHrv7d"`
	AvgSleep7d    *float64 `json:"avgSleep7d"`
	AvgRecovery7d *float64 `json:"avgRecovery7d"`
}

// NutritionAverages contains rolling averages of nutrition data.
type NutritionAverages struct {
	AvgCalories7d *float64 `json:"avgCalories7d"`
	AvgProtein7d  *float64 `json:"avgProtein7d"`
}
