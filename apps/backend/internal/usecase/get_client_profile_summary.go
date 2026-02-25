package usecase

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GetClientProfileSummaryUseCase handles retrieving a consolidated client profile.
type GetClientProfileSummaryUseCase struct {
	measurementRepo repository.MeasurementRepository
	wearableRepo    repository.WearableSummaryRepository
	coachClientRepo repository.CoachClientRepository
	auditRepo       repository.AuditRepository
}

// NewGetClientProfileSummaryUseCase creates a new GetClientProfileSummaryUseCase.
func NewGetClientProfileSummaryUseCase(
	measurementRepo repository.MeasurementRepository,
	wearableRepo repository.WearableSummaryRepository,
	coachClientRepo repository.CoachClientRepository,
	auditRepo repository.AuditRepository,
) *GetClientProfileSummaryUseCase {
	return &GetClientProfileSummaryUseCase{
		measurementRepo: measurementRepo,
		wearableRepo:    wearableRepo,
		coachClientRepo: coachClientRepo,
		auditRepo:       auditRepo,
	}
}

// GetClientProfileSummaryInput holds the input parameters.
type GetClientProfileSummaryInput struct {
	CoachID  string
	ClientID string
}

// ProfileSummaryOutput holds the consolidated profile response.
type ProfileSummaryOutput struct {
	BodyComposition map[string]*MeasurementSummary `json:"body_composition"`
	Labs            LabsSummary                    `json:"labs"`
	Wearable        *WearableSummaryInfo           `json:"wearable"`
	Nutrition       NutritionSummary               `json:"nutrition"`
}

// MeasurementSummary holds a single measurement point for the profile summary.
type MeasurementSummary struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
	Date  string  `json:"date"`
}

// LabsSummary holds lab result summary data.
type LabsSummary struct {
	FlaggedCount int                      `json:"flagged_count"`
	LastTestDate string                   `json:"last_test_date"`
	Markers      []*LabMarker             `json:"markers"`
}

// LabMarker holds a single lab marker.
type LabMarker struct {
	Type  string  `json:"type"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
	Flag  string  `json:"flag,omitempty"`
}

// WearableSummaryInfo holds wearable average data.
type WearableSummaryInfo struct {
	Source        string  `json:"source"`
	AvgHrv7d     float64 `json:"avg_hrv_7d"`
	AvgSleep7d   float64 `json:"avg_sleep_7d"`
	AvgRecovery7d float64 `json:"avg_recovery_7d"`
}

// NutritionSummary holds nutrition averages.
type NutritionSummary struct {
	AvgCalories7d float64 `json:"avg_calories_7d"`
	AvgProtein7d  float64 `json:"avg_protein_7d"`
}

// bodyCompositionTypes are measurement types that map to body composition.
var bodyCompositionTypes = map[string]bool{
	"weight":               true,
	"body_fat_pct":         true,
	"skeletal_muscle_mass": true,
	"bmi":                  true,
}

// labTypes are measurement types that map to lab results.
var labTypes = map[string]bool{
	"ldl_cholesterol":  true,
	"hdl_cholesterol":  true,
	"total_cholesterol": true,
	"triglycerides":    true,
	"glucose":          true,
	"hba1c":            true,
	"testosterone":     true,
	"vitamin_d":        true,
	"iron":             true,
	"creatinine":       true,
	"tsh":              true,
}

// nutritionTypes are measurement types that map to nutrition data.
var nutritionTypes = map[string]bool{
	"calories": true,
	"protein":  true,
	"carbs":    true,
	"fat":      true,
}

// Execute retrieves a consolidated profile summary for a client.
func (uc *GetClientProfileSummaryUseCase) Execute(ctx context.Context, input GetClientProfileSummaryInput) (*ProfileSummaryOutput, error) {
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if input.ClientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}

	// Authorization check
	rel, err := uc.coachClientRepo.FindByCoachAndClient(ctx, input.CoachID, input.ClientID)
	if err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, &AuthorizationError{Message: "not authorized to access this client's data"}
	}

	// Get latest measurements
	latest, err := uc.measurementRepo.FindLatestByClientID(ctx, input.ClientID)
	if err != nil {
		return nil, err
	}

	// Get wearable summaries (last 7 days)
	summaries, err := uc.wearableRepo.FindByClientID(ctx, input.ClientID, 7)
	if err != nil {
		return nil, err
	}

	output := buildProfileSummary(latest, summaries)

	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "phi.profile_summary_accessed",
		EntityType: "client",
		EntityID:   input.ClientID,
	})

	return output, nil
}

func buildProfileSummary(latest []*entities.Measurement, wearableSummaries []*entities.WearableSummary) *ProfileSummaryOutput {
	output := &ProfileSummaryOutput{
		BodyComposition: make(map[string]*MeasurementSummary),
		Labs: LabsSummary{
			Markers: []*LabMarker{},
		},
		Nutrition: NutritionSummary{},
	}

	// Categorize latest measurements
	for _, m := range latest {
		if bodyCompositionTypes[m.Type] {
			output.BodyComposition[m.Type] = &MeasurementSummary{
				Value: m.Value,
				Unit:  m.Unit,
				Date:  m.MeasuredAt.Format("2006-01-02"),
			}
		}
		if labTypes[m.Type] {
			marker := &LabMarker{
				Type:  m.Type,
				Value: m.Value,
				Unit:  m.Unit,
			}
			if m.Flag != nil {
				marker.Flag = *m.Flag
				output.Labs.FlaggedCount++
			}
			output.Labs.Markers = append(output.Labs.Markers, marker)
			date := m.MeasuredAt.Format("2006-01-02")
			if date > output.Labs.LastTestDate {
				output.Labs.LastTestDate = date
			}
		}
		if m.Type == "calories" {
			output.Nutrition.AvgCalories7d = m.Value
		}
		if m.Type == "protein" {
			output.Nutrition.AvgProtein7d = m.Value
		}
	}

	// Compute wearable averages from summaries
	if len(wearableSummaries) > 0 {
		var totalHrv, totalSleep, totalRecovery float64
		var countHrv, countSleep, countRecovery int
		source := wearableSummaries[0].Source

		for _, s := range wearableSummaries {
			if v, ok := getMetricFloat(s.Metrics, "hrv"); ok {
				totalHrv += v
				countHrv++
			}
			if v, ok := getMetricFloat(s.Metrics, "sleep_hours"); ok {
				totalSleep += v
				countSleep++
			}
			if v, ok := getMetricFloat(s.Metrics, "recovery_score"); ok {
				totalRecovery += v
				countRecovery++
			}
		}

		info := &WearableSummaryInfo{Source: source}
		if countHrv > 0 {
			info.AvgHrv7d = totalHrv / float64(countHrv)
		}
		if countSleep > 0 {
			info.AvgSleep7d = totalSleep / float64(countSleep)
		}
		if countRecovery > 0 {
			info.AvgRecovery7d = totalRecovery / float64(countRecovery)
		}
		output.Wearable = info
	}

	return output
}

func getMetricFloat(metrics map[string]interface{}, key string) (float64, bool) {
	v, ok := metrics[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}
