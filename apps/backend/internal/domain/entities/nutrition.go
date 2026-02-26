package entities

import (
	"sort"
	"time"
)

// NutritionMeasurementType represents the type of nutrition measurement.
type NutritionMeasurementType string

const (
	NutritionCalories NutritionMeasurementType = "calories_kcal"
	NutritionProtein  NutritionMeasurementType = "protein_g"
	NutritionCarbs    NutritionMeasurementType = "carbs_g"
	NutritionFat      NutritionMeasurementType = "fat_g"
	NutritionFiber    NutritionMeasurementType = "fiber_g"
)

// AllNutritionTypes lists all valid nutrition measurement types.
var AllNutritionTypes = []NutritionMeasurementType{
	NutritionCalories,
	NutritionProtein,
	NutritionCarbs,
	NutritionFat,
	NutritionFiber,
}

// IsValidNutritionType returns true if the given type is a valid nutrition measurement type.
func IsValidNutritionType(t NutritionMeasurementType) bool {
	for _, nt := range AllNutritionTypes {
		if t == nt {
			return true
		}
	}
	return false
}

// NutritionDailyLog represents one day of parsed nutrition data from a CSV.
type NutritionDailyLog struct {
	Date     time.Time `json:"date"`
	Calories float64   `json:"calories"`
	Protein  float64   `json:"protein"`
	Carbs    float64   `json:"carbs"`
	Fat      float64   `json:"fat"`
	Fiber    float64   `json:"fiber"`
}

// NutritionAverage represents computed averages over a period.
type NutritionAverage struct {
	ID               string    `json:"id"`
	ClientID         string    `json:"client_id"`
	SourceArtifactID string    `json:"source_artifact_id,omitempty"`
	PeriodDays       int       `json:"period_days"`
	AvgCalories      float64   `json:"avg_calories"`
	AvgProtein       float64   `json:"avg_protein"`
	AvgCarbs         float64   `json:"avg_carbs"`
	AvgFat           float64   `json:"avg_fat"`
	AvgFiber         float64   `json:"avg_fiber"`
	ComputedAt       time.Time `json:"computed_at"`
	CreatedAt        time.Time `json:"created_at"`
}

// Measurement represents a single recorded measurement value.
type Measurement struct {
	ID               string    `json:"id"`
	ClientID         string    `json:"client_id"`
	RecordedBy       string    `json:"recorded_by"`
	MeasurementType  string    `json:"measurement_type"`
	Value            float64   `json:"value"`
	Unit             string    `json:"unit"`
	MeasuredAt       time.Time `json:"measured_at"`
	Notes            string    `json:"notes,omitempty"`
	SourceArtifactID string    `json:"source_artifact_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// ComputeNutritionAverages computes averages for the given period windows from daily logs.
// The logs should be sorted by date descending (most recent first).
// Returns nil for a period if there are fewer days available than the period window.
func ComputeNutritionAverages(logs []NutritionDailyLog, periods []int) map[int]*NutritionAverage {
	if len(logs) == 0 {
		return nil
	}

	// Sort by date descending (most recent first)
	sorted := make([]NutritionDailyLog, len(logs))
	copy(sorted, logs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Date.After(sorted[j].Date)
	})

	result := make(map[int]*NutritionAverage, len(periods))
	for _, period := range periods {
		n := period
		if n > len(sorted) {
			n = len(sorted)
		}
		if n == 0 {
			continue
		}

		var sumCal, sumPro, sumCarb, sumFat, sumFib float64
		for i := 0; i < n; i++ {
			sumCal += sorted[i].Calories
			sumPro += sorted[i].Protein
			sumCarb += sorted[i].Carbs
			sumFat += sorted[i].Fat
			sumFib += sorted[i].Fiber
		}

		result[period] = &NutritionAverage{
			PeriodDays:  period,
			AvgCalories: sumCal / float64(n),
			AvgProtein:  sumPro / float64(n),
			AvgCarbs:    sumCarb / float64(n),
			AvgFat:      sumFat / float64(n),
			AvgFiber:    sumFib / float64(n),
			ComputedAt:  time.Now(),
		}
	}

	return result
}

// DailyLogsToMeasurements converts daily nutrition logs into individual measurements.
func DailyLogsToMeasurements(logs []NutritionDailyLog, clientID, coachID, artifactID string) []*Measurement {
	var measurements []*Measurement
	for _, log := range logs {
		entries := []struct {
			typ  NutritionMeasurementType
			val  float64
			unit string
		}{
			{NutritionCalories, log.Calories, "kcal"},
			{NutritionProtein, log.Protein, "g"},
			{NutritionCarbs, log.Carbs, "g"},
			{NutritionFat, log.Fat, "g"},
			{NutritionFiber, log.Fiber, "g"},
		}
		for _, e := range entries {
			measurements = append(measurements, &Measurement{
				ClientID:         clientID,
				RecordedBy:       coachID,
				MeasurementType:  string(e.typ),
				Value:            e.val,
				Unit:             e.unit,
				MeasuredAt:       log.Date,
				SourceArtifactID: artifactID,
			})
		}
	}
	return measurements
}
