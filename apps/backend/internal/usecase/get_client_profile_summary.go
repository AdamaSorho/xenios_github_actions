package usecase

import (
	"context"
	"fmt"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GetClientProfileSummaryUseCase handles building a consolidated client profile.
type GetClientProfileSummaryUseCase struct {
	measurementRepo repository.MeasurementRepository
	wearableRepo    repository.WearableSummaryRepository
	ccRepo          repository.CoachClientRepository
	auditRepo       repository.AuditRepository
}

// NewGetClientProfileSummaryUseCase creates a new GetClientProfileSummaryUseCase.
func NewGetClientProfileSummaryUseCase(
	measurementRepo repository.MeasurementRepository,
	wearableRepo repository.WearableSummaryRepository,
	ccRepo repository.CoachClientRepository,
	auditRepo repository.AuditRepository,
) *GetClientProfileSummaryUseCase {
	return &GetClientProfileSummaryUseCase{
		measurementRepo: measurementRepo,
		wearableRepo:    wearableRepo,
		ccRepo:          ccRepo,
		auditRepo:       auditRepo,
	}
}

// LatestValue holds a single measurement's latest value.
type LatestValue struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
	Date  string  `json:"date"`
}

// LabMarker holds a flagged lab result.
type LabMarker struct {
	Type  string  `json:"type"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
	Flag  string  `json:"flag"`
}

// BodyCompositionSummary holds latest body composition measurements.
type BodyCompositionSummary struct {
	Weight             *LatestValue `json:"weight,omitempty"`
	BodyFatPct         *LatestValue `json:"body_fat_pct,omitempty"`
	SkeletalMuscleMass *LatestValue `json:"skeletal_muscle_mass,omitempty"`
}

// LabsSummary holds a summary of lab results.
type LabsSummary struct {
	FlaggedCount int         `json:"flagged_count"`
	LastTestDate string      `json:"last_test_date,omitempty"`
	Markers      []LabMarker `json:"markers"`
}

// WearableSummaryOutput holds aggregated wearable data.
type WearableSummaryOutput struct {
	Source        string  `json:"source,omitempty"`
	AvgHRV7d     float64 `json:"avg_hrv_7d"`
	AvgSleep7d   float64 `json:"avg_sleep_7d"`
	AvgRecovery7d float64 `json:"avg_recovery_7d"`
}

// NutritionSummaryOutput holds averaged nutrition data.
type NutritionSummaryOutput struct {
	AvgCalories7d float64 `json:"avg_calories_7d"`
	AvgProtein7d  float64 `json:"avg_protein_7d"`
}

// ProfileSummaryOutput holds the consolidated client profile summary.
type ProfileSummaryOutput struct {
	BodyComposition BodyCompositionSummary `json:"body_composition"`
	Labs            LabsSummary            `json:"labs"`
	Wearable        WearableSummaryOutput  `json:"wearable"`
	Nutrition       NutritionSummaryOutput `json:"nutrition"`
}

// Execute builds a consolidated profile summary for a client.
func (uc *GetClientProfileSummaryUseCase) Execute(ctx context.Context, coachID, clientID string) (*ProfileSummaryOutput, error) {
	if coachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if clientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}

	if err := uc.verifyCoachClient(ctx, coachID, clientID); err != nil {
		return nil, err
	}

	latestMeasurements, err := uc.measurementRepo.FindLatestByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("find latest measurements: %w", err)
	}

	wearables, err := uc.wearableRepo.FindByClientID(ctx, clientID, 7, 0)
	if err != nil {
		return nil, fmt.Errorf("find wearable summaries: %w", err)
	}

	summary := &ProfileSummaryOutput{
		BodyComposition: buildBodyComposition(latestMeasurements),
		Labs:            buildLabsSummary(latestMeasurements),
		Wearable:        buildWearableSummary(wearables),
		Nutrition:       buildNutritionSummary(latestMeasurements),
	}

	uc.logPHIAccess(ctx, coachID, clientID, "profile_summary")

	return summary, nil
}

func buildBodyComposition(measurements []*entities.Measurement) BodyCompositionSummary {
	bc := BodyCompositionSummary{}
	for _, m := range measurements {
		lv := &LatestValue{
			Value: m.Value,
			Unit:  m.Unit,
			Date:  m.MeasuredAt.Format("2006-01-02"),
		}
		switch m.MeasurementType {
		case "weight":
			bc.Weight = lv
		case "body_fat_pct":
			bc.BodyFatPct = lv
		case "skeletal_muscle_mass":
			bc.SkeletalMuscleMass = lv
		}
	}
	return bc
}

func buildLabsSummary(measurements []*entities.Measurement) LabsSummary {
	ls := LabsSummary{Markers: []LabMarker{}}
	for _, m := range measurements {
		if m.Flag == "" {
			continue
		}
		ls.Markers = append(ls.Markers, LabMarker{
			Type:  m.MeasurementType,
			Value: m.Value,
			Unit:  m.Unit,
			Flag:  m.Flag,
		})
		ls.FlaggedCount++
		date := m.MeasuredAt.Format("2006-01-02")
		if ls.LastTestDate == "" || date > ls.LastTestDate {
			ls.LastTestDate = date
		}
	}
	return ls
}

func buildWearableSummary(wearables []*entities.WearableSummary) WearableSummaryOutput {
	ws := WearableSummaryOutput{}
	if len(wearables) == 0 {
		return ws
	}

	ws.Source = wearables[0].Source
	var hrvSum, sleepSum, recoverySum float64
	var hrvCount, sleepCount, recoveryCount int

	for _, w := range wearables {
		if v, ok := extractFloat(w.Metrics, "hrv"); ok {
			hrvSum += v
			hrvCount++
		}
		if v, ok := extractFloat(w.Metrics, "sleep_hours"); ok {
			sleepSum += v
			sleepCount++
		}
		if v, ok := extractFloat(w.Metrics, "recovery_score"); ok {
			recoverySum += v
			recoveryCount++
		}
	}

	if hrvCount > 0 {
		ws.AvgHRV7d = hrvSum / float64(hrvCount)
	}
	if sleepCount > 0 {
		ws.AvgSleep7d = sleepSum / float64(sleepCount)
	}
	if recoveryCount > 0 {
		ws.AvgRecovery7d = recoverySum / float64(recoveryCount)
	}

	return ws
}

func buildNutritionSummary(measurements []*entities.Measurement) NutritionSummaryOutput {
	ns := NutritionSummaryOutput{}
	for _, m := range measurements {
		switch m.MeasurementType {
		case "calories":
			ns.AvgCalories7d = m.Value
		case "protein":
			ns.AvgProtein7d = m.Value
		}
	}
	return ns
}

func extractFloat(metrics map[string]interface{}, key string) (float64, bool) {
	v, ok := metrics[key]
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

func (uc *GetClientProfileSummaryUseCase) verifyCoachClient(ctx context.Context, coachID, clientID string) error {
	rel, err := uc.ccRepo.FindByCoachAndClient(ctx, coachID, clientID)
	if err != nil {
		return fmt.Errorf("check coach-client relationship: %w", err)
	}
	if rel == nil {
		return &AuthenticationError{Message: "forbidden: not authorized to access this client"}
	}
	return nil
}

func (uc *GetClientProfileSummaryUseCase) logPHIAccess(ctx context.Context, coachID, clientID, resource string) {
	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    coachID,
		Action:     "phi.access",
		EntityType: "client",
		EntityID:   clientID,
		Metadata: map[string]interface{}{
			"resource": resource,
		},
	})
}
