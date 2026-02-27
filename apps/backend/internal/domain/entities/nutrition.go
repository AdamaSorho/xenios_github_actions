package entities

import "time"

// NutritionMetric represents the type of nutritional measurement.
type NutritionMetric string

const (
	NutritionMetricCalories NutritionMetric = "calories"
	NutritionMetricProtein  NutritionMetric = "protein"
	NutritionMetricCarbs    NutritionMetric = "carbs"
	NutritionMetricFat      NutritionMetric = "fat"
	NutritionMetricFiber    NutritionMetric = "fiber"
)

// NutritionUnit maps each metric to its standard unit.
var NutritionUnit = map[NutritionMetric]string{
	NutritionMetricCalories: "kcal",
	NutritionMetricProtein:  "g",
	NutritionMetricCarbs:    "g",
	NutritionMetricFat:      "g",
	NutritionMetricFiber:    "g",
}

// AllNutritionMetrics is an ordered list of all valid nutrition metrics.
var AllNutritionMetrics = []NutritionMetric{
	NutritionMetricCalories,
	NutritionMetricProtein,
	NutritionMetricCarbs,
	NutritionMetricFat,
	NutritionMetricFiber,
}

// IsValidNutritionMetric returns true if the metric is one of the known nutrition metrics.
func IsValidNutritionMetric(m NutritionMetric) bool {
	_, ok := NutritionUnit[m]
	return ok
}

// NutritionEntry represents a single parsed row from a nutrition CSV.
type NutritionEntry struct {
	Date     time.Time
	Meal     string
	Calories float64
	Protein  float64
	Carbs    float64
	Fat      float64
	Fiber    float64
}

// DailyNutrition represents the aggregated nutrition data for a single day.
type DailyNutrition struct {
	Date     time.Time
	Calories float64
	Protein  float64
	Carbs    float64
	Fat      float64
	Fiber    float64
}

// NutritionSummary holds computed rolling averages for a client.
type NutritionSummary struct {
	ID         string    `json:"id"`
	ClientID   string    `json:"client_id"`
	ArtifactID string    `json:"artifact_id"`
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
	TotalDays      int     `json:"total_days"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ComputeDailyTotals groups nutrition entries by date and sums their macros.
func ComputeDailyTotals(entries []NutritionEntry) []DailyNutrition {
	if len(entries) == 0 {
		return nil
	}

	dateMap := make(map[string]*DailyNutrition)
	var dateOrder []string

	for _, e := range entries {
		key := e.Date.Format("2006-01-02")
		d, exists := dateMap[key]
		if !exists {
			d = &DailyNutrition{Date: e.Date}
			dateMap[key] = d
			dateOrder = append(dateOrder, key)
		}
		d.Calories += e.Calories
		d.Protein += e.Protein
		d.Carbs += e.Carbs
		d.Fat += e.Fat
		d.Fiber += e.Fiber
	}

	result := make([]DailyNutrition, 0, len(dateOrder))
	for _, key := range dateOrder {
		result = append(result, *dateMap[key])
	}
	return result
}

// ComputeAverages computes the rolling average for the last n days of data.
// If fewer than n days are available, it averages all available data.
func ComputeAverages(dailyData []DailyNutrition, n int) DailyNutrition {
	if len(dailyData) == 0 || n <= 0 {
		return DailyNutrition{}
	}

	count := n
	if count > len(dailyData) {
		count = len(dailyData)
	}

	start := len(dailyData) - count
	subset := dailyData[start:]

	var sum DailyNutrition
	for _, d := range subset {
		sum.Calories += d.Calories
		sum.Protein += d.Protein
		sum.Carbs += d.Carbs
		sum.Fat += d.Fat
		sum.Fiber += d.Fiber
	}

	fc := float64(count)
	return DailyNutrition{
		Calories: sum.Calories / fc,
		Protein:  sum.Protein / fc,
		Carbs:    sum.Carbs / fc,
		Fat:      sum.Fat / fc,
		Fiber:    sum.Fiber / fc,
	}
}
