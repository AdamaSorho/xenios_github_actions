package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// body composition measurement types
var bodyCompTypes = map[string]bool{
	"weight":               true,
	"body_fat_pct":         true,
	"skeletal_muscle_mass": true,
}

// lab measurement types
var labTypes = map[string]bool{
	"ldl_cholesterol":   true,
	"hdl_cholesterol":   true,
	"total_cholesterol": true,
	"triglycerides":     true,
	"fasting_glucose":   true,
	"hba1c":             true,
	"testosterone":      true,
	"cortisol":          true,
	"vitamin_d":         true,
	"iron":              true,
	"creatinine":        true,
}

// nutrition measurement types
var nutritionTypes = map[string]bool{
	"calories": true,
	"protein":  true,
}

// GetClientProfileSummaryUseCase returns a consolidated profile summary.
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

// GetClientProfileSummaryInput holds the input for the use case.
type GetClientProfileSummaryInput struct {
	CoachID  string
	ClientID string
}

// Execute returns a consolidated profile summary after verifying authorization.
func (uc *GetClientProfileSummaryUseCase) Execute(ctx context.Context, input GetClientProfileSummaryInput) (*entities.ProfileSummary, error) {
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if input.ClientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}

	rel, err := uc.coachClientRepo.FindByCoachAndClient(ctx, input.CoachID, input.ClientID)
	if err != nil {
		return nil, fmt.Errorf("check coach-client relationship: %w", err)
	}
	if rel == nil {
		return nil, &AuthorizationError{Message: "not authorized to access this client's data"}
	}

	latestMeasurements, err := uc.measurementRepo.FindLatestByClientID(ctx, input.ClientID)
	if err != nil {
		return nil, fmt.Errorf("find latest measurements: %w", err)
	}

	wearableSummaries, err := uc.wearableRepo.FindByClientID(ctx, input.ClientID, 7)
	if err != nil {
		return nil, fmt.Errorf("find wearable summaries: %w", err)
	}

	summary := uc.buildProfileSummary(latestMeasurements, wearableSummaries)

	if auditErr := uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "phi.access",
		EntityType: "client",
		EntityID:   input.ClientID,
		Metadata:   map[string]interface{}{"resource": "profile_summary"},
	}); auditErr != nil {
		log.Printf("audit log error: %v", auditErr)
	}

	return summary, nil
}

func (uc *GetClientProfileSummaryUseCase) buildProfileSummary(
	measurements []*entities.LatestMeasurement,
	wearables []*entities.WearableSummary,
) *entities.ProfileSummary {
	bodyComp := make(map[string]*entities.LatestMeasurement)
	labMarkers := make([]*entities.LatestMeasurement, 0)
	flaggedCount := 0
	var lastTestDate *string
	var avgCalories, avgProtein *float64

	for _, m := range measurements {
		if bodyCompTypes[m.MeasurementType] {
			bodyComp[m.MeasurementType] = m
		}
		if labTypes[m.MeasurementType] {
			labMarkers = append(labMarkers, m)
			if m.Flag != nil && *m.Flag != "normal" {
				flaggedCount++
			}
			dateStr := m.MeasuredAt.Format("2006-01-02")
			if lastTestDate == nil || dateStr > *lastTestDate {
				lastTestDate = &dateStr
			}
		}
		if m.MeasurementType == "calories" {
			avgCalories = &m.Value
		}
		if m.MeasurementType == "protein" {
			avgProtein = &m.Value
		}
	}

	wearableAvg := buildWearableAverages(wearables)

	return &entities.ProfileSummary{
		BodyComposition: bodyComp,
		Labs: &entities.LabSummary{
			FlaggedCount: flaggedCount,
			LastTestDate: lastTestDate,
			Markers:      labMarkers,
		},
		Wearable: wearableAvg,
		Nutrition: &entities.NutritionAverages{
			AvgCalories7d: avgCalories,
			AvgProtein7d:  avgProtein,
		},
	}
}

func buildWearableAverages(wearables []*entities.WearableSummary) *entities.WearableAverages {
	if len(wearables) == 0 {
		return &entities.WearableAverages{}
	}

	var source *string
	var hrvSum, sleepSum, recoverySum float64
	var hrvCount, sleepCount, recoveryCount int

	for _, w := range wearables {
		if source == nil {
			s := w.Source
			source = &s
		}
		if hrv, ok := extractFloat(w.Metrics, "hrv"); ok {
			hrvSum += hrv
			hrvCount++
		}
		if sleep, ok := extractFloat(w.Metrics, "sleep_hours"); ok {
			sleepSum += sleep
			sleepCount++
		}
		if recovery, ok := extractFloat(w.Metrics, "recovery_score"); ok {
			recoverySum += recovery
			recoveryCount++
		}
	}

	avg := &entities.WearableAverages{Source: source}
	if hrvCount > 0 {
		v := hrvSum / float64(hrvCount)
		avg.AvgHRV7d = &v
	}
	if sleepCount > 0 {
		v := sleepSum / float64(sleepCount)
		avg.AvgSleep7d = &v
	}
	if recoveryCount > 0 {
		v := recoverySum / float64(recoveryCount)
		avg.AvgRecovery7d = &v
	}

	return avg
}

func extractFloat(metrics map[string]interface{}, key string) (float64, bool) {
	val, ok := metrics[key]
	if !ok {
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}
