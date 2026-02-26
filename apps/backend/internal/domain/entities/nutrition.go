package entities

import "time"

// NutritionMeasurementType represents the type of nutrition measurement.
type NutritionMeasurementType string

const (
	NutritionTypeCalories NutritionMeasurementType = "calories"
	NutritionTypeProtein  NutritionMeasurementType = "protein"
	NutritionTypeCarbs    NutritionMeasurementType = "carbs"
	NutritionTypeFat      NutritionMeasurementType = "fat"
	NutritionTypeFiber    NutritionMeasurementType = "fiber"
)

// NutritionUnit maps measurement types to their units.
var NutritionUnit = map[NutritionMeasurementType]string{
	NutritionTypeCalories: "kcal",
	NutritionTypeProtein:  "g",
	NutritionTypeCarbs:    "g",
	NutritionTypeFat:      "g",
	NutritionTypeFiber:    "g",
}

// AllNutritionTypes returns all nutrition measurement types.
func AllNutritionTypes() []NutritionMeasurementType {
	return []NutritionMeasurementType{
		NutritionTypeCalories,
		NutritionTypeProtein,
		NutritionTypeCarbs,
		NutritionTypeFat,
		NutritionTypeFiber,
	}
}

// IsValidNutritionType checks if the given type is a valid nutrition measurement type.
func IsValidNutritionType(t NutritionMeasurementType) bool {
	_, ok := NutritionUnit[t]
	return ok
}

// NutritionRow represents a single parsed row from a nutrition CSV.
type NutritionRow struct {
	Date     time.Time
	Meal     string
	Calories float64
	Protein  float64
	Carbs    float64
	Fat      float64
	Fiber    float64
}

// NutritionDailyTotal represents the summed nutritional values for a single day.
type NutritionDailyTotal struct {
	Date     time.Time
	Calories float64
	Protein  float64
	Carbs    float64
	Fat      float64
	Fiber    float64
}

// NutritionSummary holds computed rolling averages for nutrition data.
type NutritionSummary struct {
	ID         string    `json:"id"`
	ClientID   string    `json:"client_id"`
	ArtifactID string    `json:"artifact_id"`
	DaysCount  int       `json:"days_count"`
	AvgCalories7d  float64 `json:"avg_calories_7d"`
	AvgProtein7d   float64 `json:"avg_protein_7d"`
	AvgCarbs7d     float64 `json:"avg_carbs_7d"`
	AvgFat7d       float64 `json:"avg_fat_7d"`
	AvgFiber7d     float64 `json:"avg_fiber_7d"`
	AvgCalories14d float64 `json:"avg_calories_14d"`
	AvgProtein14d  float64 `json:"avg_protein_14d"`
	AvgCarbs14d    float64 `json:"avg_carbs_14d"`
	AvgFat14d      float64 `json:"avg_fat_14d"`
	AvgFiber14d    float64 `json:"avg_fiber_14d"`
	AvgCalories30d float64 `json:"avg_calories_30d"`
	AvgProtein30d  float64 `json:"avg_protein_30d"`
	AvgCarbs30d    float64 `json:"avg_carbs_30d"`
	AvgFat30d      float64 `json:"avg_fat_30d"`
	AvgFiber30d    float64 `json:"avg_fiber_30d"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Measurement represents a single stored measurement value (e.g., daily calories).
type Measurement struct {
	ID              string    `json:"id"`
	ClientID        string    `json:"client_id"`
	RecordedBy      string    `json:"recorded_by"`
	MeasurementType string    `json:"measurement_type"`
	Value           float64   `json:"value"`
	Unit            string    `json:"unit"`
	MeasuredAt      time.Time `json:"measured_at"`
	Notes           string    `json:"notes,omitempty"`
	ArtifactID      string    `json:"artifact_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// ComputeDailyTotals groups nutrition rows by date and sums per-day values.
func ComputeDailyTotals(rows []NutritionRow) []NutritionDailyTotal {
	dateMap := make(map[string]*NutritionDailyTotal)
	var order []string

	for _, r := range rows {
		key := r.Date.Format("2006-01-02")
		dt, exists := dateMap[key]
		if !exists {
			dt = &NutritionDailyTotal{Date: r.Date}
			dateMap[key] = dt
			order = append(order, key)
		}
		dt.Calories += r.Calories
		dt.Protein += r.Protein
		dt.Carbs += r.Carbs
		dt.Fat += r.Fat
		dt.Fiber += r.Fiber
	}

	result := make([]NutritionDailyTotal, 0, len(order))
	for _, key := range order {
		result = append(result, *dateMap[key])
	}
	return result
}

// ComputeAverages computes the average of daily totals for the last N days.
// If there are fewer days than N, averages all available data.
func ComputeAverages(totals []NutritionDailyTotal, days int) NutritionDailyTotal {
	if len(totals) == 0 {
		return NutritionDailyTotal{}
	}

	count := days
	if count > len(totals) {
		count = len(totals)
	}

	// Take the last 'count' entries
	subset := totals[len(totals)-count:]
	var sum NutritionDailyTotal
	for _, t := range subset {
		sum.Calories += t.Calories
		sum.Protein += t.Protein
		sum.Carbs += t.Carbs
		sum.Fat += t.Fat
		sum.Fiber += t.Fiber
	}

	n := float64(count)
	return NutritionDailyTotal{
		Calories: sum.Calories / n,
		Protein:  sum.Protein / n,
		Carbs:    sum.Carbs / n,
		Fat:      sum.Fat / n,
		Fiber:    sum.Fiber / n,
	}
}
