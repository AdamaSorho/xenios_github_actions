package entities

import "time"

// NutritionMetricType represents the type of nutrition metric.
type NutritionMetricType string

const (
	NutritionMetricCalories NutritionMetricType = "calories"
	NutritionMetricProtein  NutritionMetricType = "protein"
	NutritionMetricCarbs    NutritionMetricType = "carbs"
	NutritionMetricFat      NutritionMetricType = "fat"
	NutritionMetricFiber    NutritionMetricType = "fiber"
)

// NutritionMetricUnit maps each metric type to its unit.
var NutritionMetricUnit = map[NutritionMetricType]string{
	NutritionMetricCalories: "kcal",
	NutritionMetricProtein:  "g",
	NutritionMetricCarbs:    "g",
	NutritionMetricFat:      "g",
	NutritionMetricFiber:    "g",
}

// IsValidNutritionMetricType returns true if the given metric type is known.
func IsValidNutritionMetricType(mt NutritionMetricType) bool {
	_, ok := NutritionMetricUnit[mt]
	return ok
}

// NutritionRecord represents a single day's nutrition data for one metric.
type NutritionRecord struct {
	ID         string              `json:"id"`
	ClientID   string              `json:"client_id"`
	CoachID    string              `json:"coach_id"`
	ArtifactID string              `json:"artifact_id"`
	MetricType NutritionMetricType `json:"metric_type"`
	Value      float64             `json:"value"`
	Unit       string              `json:"unit"`
	RecordDate time.Time           `json:"record_date"`
	CreatedAt  time.Time           `json:"created_at"`
}

// NutritionDailyTotal represents summed nutrition data for a single day.
type NutritionDailyTotal struct {
	Date     time.Time `json:"date"`
	Calories float64   `json:"calories"`
	Protein  float64   `json:"protein"`
	Carbs    float64   `json:"carbs"`
	Fat      float64   `json:"fat"`
	Fiber    float64   `json:"fiber"`
}

// NutritionSummary holds computed rolling averages for nutrition data.
type NutritionSummary struct {
	ID              string    `json:"id"`
	ClientID        string    `json:"client_id"`
	ArtifactID      string    `json:"artifact_id"`
	TotalDays       int       `json:"total_days"`
	AvgCalories7d   float64   `json:"avg_calories_7d"`
	AvgProtein7d    float64   `json:"avg_protein_7d"`
	AvgCarbs7d      float64   `json:"avg_carbs_7d"`
	AvgFat7d        float64   `json:"avg_fat_7d"`
	AvgFiber7d      float64   `json:"avg_fiber_7d"`
	AvgCalories14d  float64   `json:"avg_calories_14d"`
	AvgProtein14d   float64   `json:"avg_protein_14d"`
	AvgCarbs14d     float64   `json:"avg_carbs_14d"`
	AvgFat14d       float64   `json:"avg_fat_14d"`
	AvgFiber14d     float64   `json:"avg_fiber_14d"`
	AvgCalories30d  float64   `json:"avg_calories_30d"`
	AvgProtein30d   float64   `json:"avg_protein_30d"`
	AvgCarbs30d     float64   `json:"avg_carbs_30d"`
	AvgFat30d       float64   `json:"avg_fat_30d"`
	AvgFiber30d     float64   `json:"avg_fiber_30d"`
	ComputedAt      time.Time `json:"computed_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ComputeAverages calculates 7/14/30-day averages from daily totals.
// Daily totals must be sorted by date ascending.
func ComputeAverages(dailyTotals []NutritionDailyTotal) NutritionSummary {
	n := len(dailyTotals)
	summary := NutritionSummary{TotalDays: n}
	if n == 0 {
		return summary
	}

	avg := func(totals []NutritionDailyTotal) (cal, prot, carbs, fat, fiber float64) {
		if len(totals) == 0 {
			return
		}
		for _, t := range totals {
			cal += t.Calories
			prot += t.Protein
			carbs += t.Carbs
			fat += t.Fat
			fiber += t.Fiber
		}
		cnt := float64(len(totals))
		return cal / cnt, prot / cnt, carbs / cnt, fat / cnt, fiber / cnt
	}

	// Use last N days (from the end of the sorted slice)
	window := func(days int) []NutritionDailyTotal {
		if days >= n {
			return dailyTotals
		}
		return dailyTotals[n-days:]
	}

	summary.AvgCalories7d, summary.AvgProtein7d, summary.AvgCarbs7d, summary.AvgFat7d, summary.AvgFiber7d = avg(window(7))
	summary.AvgCalories14d, summary.AvgProtein14d, summary.AvgCarbs14d, summary.AvgFat14d, summary.AvgFiber14d = avg(window(14))
	summary.AvgCalories30d, summary.AvgProtein30d, summary.AvgCarbs30d, summary.AvgFat30d, summary.AvgFiber30d = avg(window(30))

	return summary
}
