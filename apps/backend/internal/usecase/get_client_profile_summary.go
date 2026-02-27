package usecase

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// bodyCompTypes lists measurement types that belong to body composition.
var bodyCompTypes = map[string]bool{
	"weight":               true,
	"body_fat_pct":         true,
	"skeletal_muscle_mass": true,
	"bmi":                  true,
}

// nutritionTypes lists measurement types that belong to nutrition.
var nutritionTypes = map[string]bool{
	"calories": true,
	"protein":  true,
}

const wearableLookbackDays = 7

// GetClientProfileSummaryUseCase handles building a consolidated client profile.
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

// Execute builds and returns a consolidated profile summary for a client.
func (uc *GetClientProfileSummaryUseCase) Execute(ctx context.Context, coachID, clientID string) (*entities.ProfileSummary, error) {
	if err := validateCoachClientInput(coachID, clientID); err != nil {
		return nil, err
	}

	if err := authorizeCoachClient(ctx, uc.coachClientRepo, coachID, clientID); err != nil {
		return nil, err
	}

	latestMeasurements, err := uc.measurementRepo.FindLatestByClientID(ctx, clientID)
	if err != nil {
		return nil, err
	}

	wearableSummaries, err := uc.wearableRepo.FindByClientID(ctx, clientID, wearableLookbackDays)
	if err != nil {
		return nil, err
	}

	summary := buildProfileSummary(latestMeasurements, wearableSummaries)

	logPHIAccess(ctx, uc.auditRepo, coachID, clientID, "profile_summary")
	return summary, nil
}

func buildProfileSummary(measurements []*entities.Measurement, wearables []*entities.WearableSummary) *entities.ProfileSummary {
	summary := &entities.ProfileSummary{
		BodyComposition: make(map[string]*entities.LatestMeasurement),
		Labs:            &entities.LabSummary{Markers: make([]*entities.LatestMeasurement, 0)},
		Wearable:        &entities.WearableAverage{},
		Nutrition:       &entities.NutritionAverage{},
	}

	for _, m := range measurements {
		lm := &entities.LatestMeasurement{
			Type:  m.Type,
			Value: m.Value,
			Unit:  m.Unit,
			Date:  m.MeasuredAt.Format("2006-01-02"),
		}

		if bodyCompTypes[m.Type] {
			summary.BodyComposition[m.Type] = lm
		} else if nutritionTypes[m.Type] {
			populateNutritionAvg(summary.Nutrition, m)
		} else {
			addLabMarker(summary.Labs, m, lm)
		}
	}

	populateWearableAvg(summary.Wearable, wearables)
	return summary
}

func addLabMarker(labs *entities.LabSummary, m *entities.Measurement, lm *entities.LatestMeasurement) {
	if m.Flag != "" {
		lm.Type = m.Type
		labs.Markers = append(labs.Markers, lm)
		labs.FlaggedCount++
	}
	date := m.MeasuredAt.Format("2006-01-02")
	if labs.LastTestDate == "" || date > labs.LastTestDate {
		labs.LastTestDate = date
	}
}

func populateNutritionAvg(n *entities.NutritionAverage, m *entities.Measurement) {
	v := m.Value
	switch m.Type {
	case "calories":
		n.AvgCalories7d = &v
	case "protein":
		n.AvgProtein7d = &v
	}
}

func populateWearableAvg(w *entities.WearableAverage, summaries []*entities.WearableSummary) {
	if len(summaries) == 0 {
		return
	}

	w.Source = summaries[0].Source

	var hrvSum, sleepSum, recoverySum float64
	var hrvCount, sleepCount, recoveryCount int

	for _, s := range summaries {
		if v, ok := extractFloat(s.Metrics, "hrv"); ok {
			hrvSum += v
			hrvCount++
		}
		if v, ok := extractFloat(s.Metrics, "sleep_hours"); ok {
			sleepSum += v
			sleepCount++
		}
		if v, ok := extractFloat(s.Metrics, "recovery_score"); ok {
			recoverySum += v
			recoveryCount++
		}
	}

	if hrvCount > 0 {
		avg := hrvSum / float64(hrvCount)
		w.AvgHRV7d = &avg
	}
	if sleepCount > 0 {
		avg := sleepSum / float64(sleepCount)
		w.AvgSleep7d = &avg
	}
	if recoveryCount > 0 {
		avg := recoverySum / float64(recoveryCount)
		w.AvgRecovery7d = &avg
	}
}

func extractFloat(metrics map[string]interface{}, key string) (float64, bool) {
	v, ok := metrics[key]
	if !ok {
		return 0, false
	}
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}
