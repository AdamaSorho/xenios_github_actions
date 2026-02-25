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

// GetMetricFloat64 extracts a float64 metric from the metrics map.
// Returns 0 and false if the key doesn't exist or isn't a number.
func (w *WearableSummary) GetMetricFloat64(key string) (float64, bool) {
	v, ok := w.Metrics[key]
	if !ok {
		return 0, false
	}
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}
