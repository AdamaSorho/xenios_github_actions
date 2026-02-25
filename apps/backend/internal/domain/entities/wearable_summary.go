package entities

import (
	"encoding/json"
	"time"
)

// WearableSummary stores rolling averages for a client/source combination.
type WearableSummary struct {
	ID        string          `json:"id"`
	ClientID  string          `json:"client_id"`
	Source    WearableSource  `json:"source"`
	Metrics   json.RawMessage `json:"metrics"`
	UpdatedAt time.Time       `json:"updated_at"`
	CreatedAt time.Time       `json:"created_at"`
}

// RollingAverages holds the computed rolling averages for display/storage.
type RollingAverages struct {
	AvgHRV7d        *float64 `json:"avg_hrv_7d,omitempty"`
	AvgHRV14d       *float64 `json:"avg_hrv_14d,omitempty"`
	AvgHRV30d       *float64 `json:"avg_hrv_30d,omitempty"`
	AvgSleep7d      *float64 `json:"avg_sleep_7d,omitempty"`
	AvgSleep14d     *float64 `json:"avg_sleep_14d,omitempty"`
	AvgSleep30d     *float64 `json:"avg_sleep_30d,omitempty"`
	AvgRecovery7d   *float64 `json:"avg_recovery_7d,omitempty"`
	AvgRecovery14d  *float64 `json:"avg_recovery_14d,omitempty"`
	AvgRecovery30d  *float64 `json:"avg_recovery_30d,omitempty"`
	AvgRestingHR7d  *float64 `json:"avg_resting_hr_7d,omitempty"`
	AvgRestingHR14d *float64 `json:"avg_resting_hr_14d,omitempty"`
	AvgRestingHR30d *float64 `json:"avg_resting_hr_30d,omitempty"`
	AvgSteps7d      *float64 `json:"avg_steps_7d,omitempty"`
	AvgSteps14d     *float64 `json:"avg_steps_14d,omitempty"`
	AvgSteps30d     *float64 `json:"avg_steps_30d,omitempty"`
	AvgStrain7d     *float64 `json:"avg_strain_7d,omitempty"`
	AvgStrain14d    *float64 `json:"avg_strain_14d,omitempty"`
	AvgStrain30d    *float64 `json:"avg_strain_30d,omitempty"`
}
