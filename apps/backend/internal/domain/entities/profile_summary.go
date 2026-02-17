package entities

// ProfileSummary is a consolidated view of a client's health data.
type ProfileSummary struct {
	BodyComposition map[string]*LatestMeasurement `json:"body_composition"`
	Labs            *LabSummary                   `json:"labs"`
	Wearable        *WearableAverages             `json:"wearable,omitempty"`
	Nutrition       *NutritionAverages            `json:"nutrition,omitempty"`
}

// LabSummary holds summary information about lab results.
type LabSummary struct {
	FlaggedCount int                `json:"flagged_count"`
	LastTestDate *string            `json:"last_test_date,omitempty"`
	Markers      []LatestMeasurement `json:"markers"`
}

// NutritionAverages holds rolling nutrition averages.
type NutritionAverages struct {
	AvgCalories7d *float64 `json:"avg_calories_7d,omitempty"`
	AvgProtein7d  *float64 `json:"avg_protein_7d,omitempty"`
}
