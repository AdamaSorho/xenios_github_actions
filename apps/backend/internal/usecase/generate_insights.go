package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GenerateInsightsUseCase evaluates extracted measurements and creates draft insight cards.
type GenerateInsightsUseCase struct {
	insightRepo     repository.InsightCardRepository
	measurementRepo repository.MeasurementRepository
	auditRepo       repository.AuditRepository
}

// NewGenerateInsightsUseCase creates a new GenerateInsightsUseCase.
func NewGenerateInsightsUseCase(
	insightRepo repository.InsightCardRepository,
	measurementRepo repository.MeasurementRepository,
	auditRepo repository.AuditRepository,
) *GenerateInsightsUseCase {
	return &GenerateInsightsUseCase{
		insightRepo:     insightRepo,
		measurementRepo: measurementRepo,
		auditRepo:       auditRepo,
	}
}

// GenerateInsightsInput holds the input for insight generation.
type GenerateInsightsInput struct {
	ClientID   string
	CoachID    string
	ArtifactID string
}

// GenerateInsightsOutput holds the results of insight generation.
type GenerateInsightsOutput struct {
	InsightsCreated int
	Insights        []*entities.InsightCard
}

// Execute evaluates measurements for the given artifact and creates draft insight cards.
func (uc *GenerateInsightsUseCase) Execute(ctx context.Context, input GenerateInsightsInput) (*GenerateInsightsOutput, error) {
	if input.ClientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if input.ArtifactID == "" {
		return nil, &ValidationError{Message: "artifact_id is required"}
	}

	measurements, err := uc.measurementRepo.FindRecentByArtifactID(ctx, input.ArtifactID)
	if err != nil {
		return nil, fmt.Errorf("find measurements: %w", err)
	}

	var created []*entities.InsightCard

	for _, m := range measurements {
		cards, err := uc.evaluateLabMeasurement(ctx, input.CoachID, m)
		if err != nil {
			return nil, fmt.Errorf("evaluate lab measurement: %w", err)
		}
		created = append(created, cards...)
	}

	trendCards, err := uc.evaluateWearableTrends(ctx, input.ClientID, input.CoachID)
	if err != nil {
		return nil, fmt.Errorf("evaluate wearable trends: %w", err)
	}
	created = append(created, trendCards...)

	bodyCompCards, err := uc.evaluateBodyComposition(ctx, input.ClientID, input.CoachID)
	if err != nil {
		return nil, fmt.Errorf("evaluate body composition: %w", err)
	}
	created = append(created, bodyCompCards...)

	return &GenerateInsightsOutput{
		InsightsCreated: len(created),
		Insights:        created,
	}, nil
}

func (uc *GenerateInsightsUseCase) evaluateLabMeasurement(ctx context.Context, coachID string, m *entities.Measurement) ([]*entities.InsightCard, error) {
	if !m.IsOutOfRange() {
		return nil, nil
	}

	exists, err := uc.insightRepo.ExistsByEvidence(ctx, m.ClientID, m.ID)
	if err != nil {
		return nil, fmt.Errorf("check duplicate: %w", err)
	}
	if exists {
		return nil, nil
	}

	card := buildLabInsightCard(coachID, m)

	saved, err := uc.insightRepo.Create(ctx, card)
	if err != nil {
		return nil, fmt.Errorf("create insight: %w", err)
	}

	uc.logInsightAudit(ctx, coachID, saved.ID)

	return []*entities.InsightCard{saved}, nil
}

func buildLabInsightCard(coachID string, m *entities.Measurement) *entities.InsightCard {
	priority := entities.InsightPriorityHigh
	category := entities.InsightCategoryNutrition

	if m.IsCritical() {
		priority = entities.InsightPriorityUrgent
		category = entities.InsightCategorySafety
	}

	title := fmt.Sprintf("Elevated %s", m.MeasurementType)
	if m.Flag == entities.MeasurementFlagLow || m.Flag == entities.MeasurementFlagCriticalLow {
		title = fmt.Sprintf("Low %s", m.MeasurementType)
	}

	body := formatLabBody(m)

	return &entities.InsightCard{
		CoachID:  coachID,
		ClientID: m.ClientID,
		Title:    title,
		Body:     body,
		Category: category,
		Priority: priority,
		Status:   entities.InsightStatusDraft,
		Evidence: []entities.EvidenceRef{
			{
				MeasurementID: m.ID,
				ArtifactID:    m.ArtifactID,
				Description:   formatEvidenceDescription(m),
			},
		},
	}
}

func formatLabBody(m *entities.Measurement) string {
	if m.RefRangeHigh != nil && (m.Flag == entities.MeasurementFlagHigh || m.Flag == entities.MeasurementFlagCriticalHigh) {
		return fmt.Sprintf("%s is %.1f %s, above the recommended < %.1f %s",
			m.MeasurementType, m.Value, m.Unit, *m.RefRangeHigh, m.Unit)
	}
	if m.RefRangeLow != nil && (m.Flag == entities.MeasurementFlagLow || m.Flag == entities.MeasurementFlagCriticalLow) {
		return fmt.Sprintf("%s is %.1f %s, below the recommended > %.1f %s",
			m.MeasurementType, m.Value, m.Unit, *m.RefRangeLow, m.Unit)
	}
	return fmt.Sprintf("%s is %.1f %s (flagged %s)", m.MeasurementType, m.Value, m.Unit, m.Flag)
}

func formatEvidenceDescription(m *entities.Measurement) string {
	if m.RefRangeHigh != nil || m.RefRangeLow != nil {
		low, high := 0.0, 0.0
		if m.RefRangeLow != nil {
			low = *m.RefRangeLow
		}
		if m.RefRangeHigh != nil {
			high = *m.RefRangeHigh
		}
		return fmt.Sprintf("%s: %.1f %s (ref: %.1f-%.1f)", m.MeasurementType, m.Value, m.Unit, low, high)
	}
	return fmt.Sprintf("%s: %.1f %s", m.MeasurementType, m.Value, m.Unit)
}

func (uc *GenerateInsightsUseCase) evaluateWearableTrends(ctx context.Context, clientID, coachID string) ([]*entities.InsightCard, error) {
	var created []*entities.InsightCard

	hrvCards, err := uc.evaluateHRVTrend(ctx, clientID, coachID)
	if err != nil {
		return nil, err
	}
	created = append(created, hrvCards...)

	sleepCards, err := uc.evaluateSleepTrend(ctx, clientID, coachID)
	if err != nil {
		return nil, err
	}
	created = append(created, sleepCards...)

	return created, nil
}

func (uc *GenerateInsightsUseCase) evaluateHRVTrend(ctx context.Context, clientID, coachID string) ([]*entities.InsightCard, error) {
	now := time.Now()
	twoWeeksAgo := now.AddDate(0, 0, -14)

	hrvData, err := uc.measurementRepo.FindByClientIDAndType(ctx, clientID, "hrv", twoWeeksAgo)
	if err != nil {
		return nil, fmt.Errorf("find HRV data: %w", err)
	}

	if len(hrvData) < 2 {
		return nil, nil
	}

	oneWeekAgo := now.AddDate(0, 0, -7)
	priorAvg, recentAvg := splitWeeklyAverages(hrvData, oneWeekAgo)

	if priorAvg <= 0 {
		return nil, nil
	}

	dropPercent := ((priorAvg - recentAvg) / priorAvg) * 100
	if dropPercent <= 15 {
		return nil, nil
	}

	card := &entities.InsightCard{
		CoachID:  coachID,
		ClientID: clientID,
		Title:    "HRV Declining",
		Body:     fmt.Sprintf("HRV has declined %.0f%% this week — consider reducing training volume", dropPercent),
		Category: entities.InsightCategoryRecovery,
		Priority: entities.InsightPriorityMedium,
		Status:   entities.InsightStatusDraft,
	}

	saved, err := uc.insightRepo.Create(ctx, card)
	if err != nil {
		return nil, fmt.Errorf("create HRV insight: %w", err)
	}

	uc.logInsightAudit(ctx, coachID, saved.ID)
	return []*entities.InsightCard{saved}, nil
}

func (uc *GenerateInsightsUseCase) evaluateSleepTrend(ctx context.Context, clientID, coachID string) ([]*entities.InsightCard, error) {
	now := time.Now()
	oneWeekAgo := now.AddDate(0, 0, -7)

	sleepData, err := uc.measurementRepo.FindByClientIDAndType(ctx, clientID, "sleep_hours", oneWeekAgo)
	if err != nil {
		return nil, fmt.Errorf("find sleep data: %w", err)
	}

	if len(sleepData) == 0 {
		return nil, nil
	}

	avg := average(sleepData)
	if avg >= 6 {
		return nil, nil
	}

	card := &entities.InsightCard{
		CoachID:  coachID,
		ClientID: clientID,
		Title:    "Sleep Declining",
		Body:     fmt.Sprintf("Average sleep dropped to %.1f hours — sleep quality may be impacting recovery", avg),
		Category: entities.InsightCategoryRecovery,
		Priority: entities.InsightPriorityMedium,
		Status:   entities.InsightStatusDraft,
	}

	saved, err := uc.insightRepo.Create(ctx, card)
	if err != nil {
		return nil, fmt.Errorf("create sleep insight: %w", err)
	}

	uc.logInsightAudit(ctx, coachID, saved.ID)
	return []*entities.InsightCard{saved}, nil
}

func (uc *GenerateInsightsUseCase) evaluateBodyComposition(ctx context.Context, clientID, coachID string) ([]*entities.InsightCard, error) {
	var created []*entities.InsightCard

	weightCards, err := uc.evaluateWeightChange(ctx, clientID, coachID)
	if err != nil {
		return nil, err
	}
	created = append(created, weightCards...)

	bfCards, err := uc.evaluateBodyFatProgress(ctx, clientID, coachID)
	if err != nil {
		return nil, err
	}
	created = append(created, bfCards...)

	return created, nil
}

func (uc *GenerateInsightsUseCase) evaluateWeightChange(ctx context.Context, clientID, coachID string) ([]*entities.InsightCard, error) {
	now := time.Now()
	twoWeeksAgo := now.AddDate(0, 0, -14)

	weightData, err := uc.measurementRepo.FindByClientIDAndType(ctx, clientID, "weight", twoWeeksAgo)
	if err != nil {
		return nil, fmt.Errorf("find weight data: %w", err)
	}

	if len(weightData) < 2 {
		return nil, nil
	}

	oldest := weightData[0].Value
	newest := weightData[len(weightData)-1].Value

	if oldest == 0 {
		return nil, nil
	}

	changePercent := ((newest - oldest) / oldest) * 100
	absChange := changePercent
	if absChange < 0 {
		absChange = -absChange
	}

	if absChange <= 3 {
		return nil, nil
	}

	direction := "increased"
	if changePercent < 0 {
		direction = "decreased"
	}

	card := &entities.InsightCard{
		CoachID:  coachID,
		ClientID: clientID,
		Title:    "Significant Weight Change",
		Body:     fmt.Sprintf("Weight %s %.1f%% in 2 weeks — verify this aligns with client goals", direction, absChange),
		Category: entities.InsightCategoryNutrition,
		Priority: entities.InsightPriorityMedium,
		Status:   entities.InsightStatusDraft,
	}

	saved, err := uc.insightRepo.Create(ctx, card)
	if err != nil {
		return nil, fmt.Errorf("create weight insight: %w", err)
	}

	uc.logInsightAudit(ctx, coachID, saved.ID)
	return []*entities.InsightCard{saved}, nil
}

func (uc *GenerateInsightsUseCase) evaluateBodyFatProgress(ctx context.Context, clientID, coachID string) ([]*entities.InsightCard, error) {
	now := time.Now()
	twoWeeksAgo := now.AddDate(0, 0, -14)

	bfData, err := uc.measurementRepo.FindByClientIDAndType(ctx, clientID, "body_fat", twoWeeksAgo)
	if err != nil {
		return nil, fmt.Errorf("find body fat data: %w", err)
	}

	if len(bfData) < 2 {
		return nil, nil
	}

	oldest := bfData[0].Value
	newest := bfData[len(bfData)-1].Value
	drop := oldest - newest

	if drop < 1 {
		return nil, nil
	}

	card := &entities.InsightCard{
		CoachID:  coachID,
		ClientID: clientID,
		Title:    "Body Fat Progress",
		Body:     fmt.Sprintf("Body fat decreased from %.1f%% to %.1f%% — great progress!", oldest, newest),
		Category: entities.InsightCategoryPerformance,
		Priority: entities.InsightPriorityLow,
		Status:   entities.InsightStatusDraft,
	}

	saved, err := uc.insightRepo.Create(ctx, card)
	if err != nil {
		return nil, fmt.Errorf("create body fat insight: %w", err)
	}

	uc.logInsightAudit(ctx, coachID, saved.ID)
	return []*entities.InsightCard{saved}, nil
}

func (uc *GenerateInsightsUseCase) logInsightAudit(ctx context.Context, coachID, insightID string) {
	if err := uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    coachID,
		Action:     "insight.generate",
		EntityType: "insight_card",
		EntityID:   insightID,
	}); err != nil {
		log.Printf("audit log error: %v", err)
	}
}

func splitWeeklyAverages(data []*entities.Measurement, splitDate time.Time) (priorAvg, recentAvg float64) {
	var priorSum, recentSum float64
	var priorCount, recentCount int

	for _, m := range data {
		if m.MeasuredAt.Before(splitDate) {
			priorSum += m.Value
			priorCount++
		} else {
			recentSum += m.Value
			recentCount++
		}
	}

	if priorCount > 0 {
		priorAvg = priorSum / float64(priorCount)
	}
	if recentCount > 0 {
		recentAvg = recentSum / float64(recentCount)
	}
	return
}

func average(data []*entities.Measurement) float64 {
	if len(data) == 0 {
		return 0
	}
	var sum float64
	for _, m := range data {
		sum += m.Value
	}
	return sum / float64(len(data))
}
