package entities

import "time"

// MeasurementType represents the type of nutrition measurement.
type MeasurementType string

const (
	MeasurementTypeCalories MeasurementType = "calories"
	MeasurementTypeProtein  MeasurementType = "protein"
	MeasurementTypeCarbs    MeasurementType = "carbs"
	MeasurementTypeFat      MeasurementType = "fat"
	MeasurementTypeFiber    MeasurementType = "fiber"
)

// MeasurementUnit returns the unit for a given measurement type.
func MeasurementUnit(mt MeasurementType) string {
	switch mt {
	case MeasurementTypeCalories:
		return "kcal"
	default:
		return "g"
	}
}

// IsValidMeasurementType returns true if the given type is a known nutrition measurement type.
func IsValidMeasurementType(mt MeasurementType) bool {
	switch mt {
	case MeasurementTypeCalories,
		MeasurementTypeProtein,
		MeasurementTypeCarbs,
		MeasurementTypeFat,
		MeasurementTypeFiber:
		return true
	}
	return false
}

// AllMeasurementTypes returns all known nutrition measurement types.
func AllMeasurementTypes() []MeasurementType {
	return []MeasurementType{
		MeasurementTypeCalories,
		MeasurementTypeProtein,
		MeasurementTypeCarbs,
		MeasurementTypeFat,
		MeasurementTypeFiber,
	}
}

// Measurement represents a single nutrition measurement stored in the database.
type Measurement struct {
	ID              string          `json:"id"`
	ClientID        string          `json:"client_id"`
	RecordedBy      string          `json:"recorded_by"`
	MeasurementType MeasurementType `json:"measurement_type"`
	Value           float64         `json:"value"`
	Unit            string          `json:"unit"`
	MeasuredAt      time.Time       `json:"measured_at"`
	Notes           string          `json:"notes,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
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

// DailyNutrition holds the summed nutritional data for a single day.
type DailyNutrition struct {
	Date     time.Time
	Calories float64
	Protein  float64
	Carbs    float64
	Fat      float64
	Fiber    float64
}

// NutritionSummary holds rolling average data for nutrition.
type NutritionSummary struct {
	ClientID       string             `json:"client_id"`
	ArtifactID     string             `json:"artifact_id"`
	DailyTotals    []DailyNutrition   `json:"daily_totals"`
	Averages       NutritionAverages  `json:"averages"`
	TotalDays      int                `json:"total_days"`
	ProcessedAt    time.Time          `json:"processed_at"`
}

// NutritionAverages holds computed rolling averages.
type NutritionAverages struct {
	Avg7Day  *PeriodAverage `json:"avg_7d,omitempty"`
	Avg14Day *PeriodAverage `json:"avg_14d,omitempty"`
	Avg30Day *PeriodAverage `json:"avg_30d,omitempty"`
}

// PeriodAverage holds average values for a given period.
type PeriodAverage struct {
	Days     int     `json:"days"`
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Carbs    float64 `json:"carbs"`
	Fat      float64 `json:"fat"`
	Fiber    float64 `json:"fiber"`
}

// CSVFormat represents the detected format of a nutrition CSV file.
type CSVFormat string

const (
	CSVFormatMyFitnessPal CSVFormat = "myfitnesspal"
	CSVFormatGeneric      CSVFormat = "generic"
	CSVFormatUnknown      CSVFormat = "unknown"
)
