package usecase

import (
	"context"
	"log"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GetClientProfileSummaryUseCase retrieves a consolidated view of a client's health data.
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

// Body composition measurement types.
var bodyCompTypes = map[string]bool{
	"weight":                true,
	"body_fat_pct":          true,
	"skeletal_muscle_mass":  true,
	"bmi":                   true,
	"waist_circumference":   true,
}

// Lab marker measurement types (not body composition).
var labMarkerTypes = map[string]bool{
	"ldl_cholesterol":  true,
	"hdl_cholesterol":  true,
	"total_cholesterol": true,
	"triglycerides":    true,
	"glucose":          true,
	"hba1c":            true,
	"testosterone":     true,
	"cortisol":         true,
	"vitamin_d":        true,
	"iron":             true,
	"crp":              true,
	"tsh":              true,
}

// Nutrition measurement types.
var nutritionTypes = map[string]bool{
	"calories":     true,
	"protein":      true,
	"carbohydrates": true,
	"fat":          true,
}

// Execute retrieves the client's profile summary.
func (uc *GetClientProfileSummaryUseCase) Execute(ctx context.Context, coachID, clientID string) (*entities.ProfileSummary, error) {
	if coachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if clientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}

	// Verify coach-client relationship
	rel, err := uc.coachClientRepo.FindByCoachAndClient(ctx, coachID, clientID)
	if err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, &AuthorizationError{Message: "access denied: no coach-client relationship"}
	}

	// Get latest measurements for all types
	latestMeasurements, err := uc.measurementRepo.FindLatestByClientID(ctx, clientID)
	if err != nil {
		return nil, err
	}

	// Build body composition map
	bodyComp := make(map[string]*entities.LatestMeasurement)
	var labMarkers []entities.LatestMeasurement
	var lastLabDate *string

	// Build nutrition averages from latest measurements
	var caloriesValue, proteinValue *float64

	for _, lm := range latestMeasurements {
		if bodyCompTypes[lm.MeasurementType] {
			bodyComp[lm.MeasurementType] = lm
		} else if labMarkerTypes[lm.MeasurementType] {
			labMarkers = append(labMarkers, *lm)
			dateStr := lm.MeasuredAt.Format("2006-01-02")
			if lastLabDate == nil || dateStr > *lastLabDate {
				lastLabDate = &dateStr
			}
		} else if lm.MeasurementType == "calories" {
			v := lm.Value
			caloriesValue = &v
		} else if lm.MeasurementType == "protein" {
			v := lm.Value
			proteinValue = &v
		}
	}

	// Get wearable averages
	wearableAvg, err := uc.wearableRepo.FindAverages(ctx, clientID, 7)
	if err != nil {
		return nil, err
	}

	// Build lab summary
	labSummary := &entities.LabSummary{
		FlaggedCount: 0,
		LastTestDate: lastLabDate,
		Markers:      labMarkers,
	}
	if labSummary.Markers == nil {
		labSummary.Markers = []entities.LatestMeasurement{}
	}

	// Build nutrition averages
	var nutritionAvg *entities.NutritionAverages
	if caloriesValue != nil || proteinValue != nil {
		nutritionAvg = &entities.NutritionAverages{
			AvgCalories7d: caloriesValue,
			AvgProtein7d:  proteinValue,
		}
	}

	// Log PHI access audit event
	auditEvent := &entities.AuditEvent{
		ActorID:    coachID,
		Action:     "phi.access",
		EntityType: "profile_summary",
		EntityID:   clientID,
		Metadata: map[string]interface{}{
			"endpoint": "profile-summary",
		},
		CreatedAt: time.Now(),
	}
	if err := uc.auditRepo.LogEvent(ctx, auditEvent); err != nil {
		log.Printf("failed to log audit event: %v", err)
	}

	return &entities.ProfileSummary{
		BodyComposition: bodyComp,
		Labs:            labSummary,
		Wearable:        wearableAvg,
		Nutrition:       nutritionAvg,
	}, nil
}
