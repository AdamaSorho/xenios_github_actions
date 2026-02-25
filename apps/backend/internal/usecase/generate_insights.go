package usecase

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

const (
	// insightLookbackDays defines how far back to look for measurements.
	insightLookbackDays = 30
	// wearableLookbackDays defines how far back to look for wearable data (2 weeks for trend comparison).
	wearableLookbackDays = 15
	// bodyCompLookbackDays defines how far back to look for body composition changes.
	bodyCompLookbackDays = 15
	// hrvDeclineThreshold is the percentage decline that triggers an insight (15%).
	hrvDeclineThreshold = 0.15
	// sleepThresholdHours is the average sleep below which an insight is generated.
	sleepThresholdHours = 6.0
	// weightChangeThreshold is the percentage change that triggers an insight (3%).
	weightChangeThreshold = 0.03
	// bodyFatDecreaseThreshold is the absolute percentage point decrease that triggers an insight.
	bodyFatDecreaseThreshold = 1.0
)

// GenerateInsightsUseCase evaluates recently extracted health data and creates draft insight cards.
type GenerateInsightsUseCase struct {
	insightRepo  repository.InsightCardRepository
	measureRepo  repository.MeasurementRepository
	wearableRepo repository.WearableSummaryRepository
	auditRepo    repository.AuditRepository
}

// NewGenerateInsightsUseCase creates a new GenerateInsightsUseCase.
func NewGenerateInsightsUseCase(
	insightRepo repository.InsightCardRepository,
	measureRepo repository.MeasurementRepository,
	wearableRepo repository.WearableSummaryRepository,
	auditRepo repository.AuditRepository,
) *GenerateInsightsUseCase {
	return &GenerateInsightsUseCase{
		insightRepo:  insightRepo,
		measureRepo:  measureRepo,
		wearableRepo: wearableRepo,
		auditRepo:    auditRepo,
	}
}

// GenerateInsightsInput holds the input for insight generation.
type GenerateInsightsInput struct {
	ClientID   string
	CoachID    string
	ArtifactID string // Optional: links insights to the source artifact
}

// GenerateInsightsOutput holds the result of insight generation.
type GenerateInsightsOutput struct {
	InsightsCreated []*entities.InsightCard
	TotalEvaluated  int
	DuplicatesSkipped int
}

// Execute evaluates measurements against rules and creates draft insights.
func (uc *GenerateInsightsUseCase) Execute(ctx context.Context, input GenerateInsightsInput) (*GenerateInsightsOutput, error) {
	if input.ClientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}

	output := &GenerateInsightsOutput{}
	now := time.Now()

	// 1. Evaluate lab measurements
	labInsights, labEvaluated, labSkipped, err := uc.evaluateLabMeasurements(ctx, input, now)
	if err != nil {
		return nil, fmt.Errorf("evaluate lab measurements: %w", err)
	}
	output.InsightsCreated = append(output.InsightsCreated, labInsights...)
	output.TotalEvaluated += labEvaluated
	output.DuplicatesSkipped += labSkipped

	// 2. Evaluate wearable trends
	wearableInsights, wearableEvaluated, err := uc.evaluateWearableTrends(ctx, input, now)
	if err != nil {
		return nil, fmt.Errorf("evaluate wearable trends: %w", err)
	}
	output.InsightsCreated = append(output.InsightsCreated, wearableInsights...)
	output.TotalEvaluated += wearableEvaluated

	// 3. Evaluate body composition changes
	bodyCompInsights, bodyCompEvaluated, bodyCompSkipped, err := uc.evaluateBodyComposition(ctx, input, now)
	if err != nil {
		return nil, fmt.Errorf("evaluate body composition: %w", err)
	}
	output.InsightsCreated = append(output.InsightsCreated, bodyCompInsights...)
	output.TotalEvaluated += bodyCompEvaluated
	output.DuplicatesSkipped += bodyCompSkipped

	// 4. Log audit events for each generated insight
	for _, card := range output.InsightsCreated {
		_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
			ActorID:    "system",
			Action:     "insight.generate",
			EntityType: "insight_card",
			EntityID:   card.ID,
			Metadata: map[string]interface{}{
				"client_id": input.ClientID,
				"coach_id":  input.CoachID,
				"category":  string(card.Category),
				"priority":  string(card.Priority),
				"title":     card.Title,
			},
		})
	}

	return output, nil
}

// evaluateLabMeasurements checks recent lab measurements against reference ranges.
func (uc *GenerateInsightsUseCase) evaluateLabMeasurements(
	ctx context.Context,
	input GenerateInsightsInput,
	now time.Time,
) ([]*entities.InsightCard, int, int, error) {
	since := now.AddDate(0, 0, -insightLookbackDays)
	measurements, err := uc.measureRepo.FindRecentByClientID(ctx, input.ClientID, since)
	if err != nil {
		return nil, 0, 0, err
	}

	var insights []*entities.InsightCard
	evaluated := 0
	skipped := 0

	for _, m := range measurements {
		ref, ok := entities.LabReferenceRanges[m.MeasurementType]
		if !ok {
			continue // Unknown measurement type, skip
		}

		evaluated++

		flag := ref.EvaluateFlag(m.Value)
		if flag == entities.MeasurementFlagNormal {
			continue
		}

		// Check for duplicate
		exists, err := uc.insightRepo.ExistsByEvidence(ctx, input.ClientID, m.ID)
		if err != nil {
			return nil, evaluated, skipped, err
		}
		if exists {
			skipped++
			continue
		}

		card := uc.buildLabInsight(input, m, ref, flag)
		created, err := uc.insightRepo.Create(ctx, card)
		if err != nil {
			return nil, evaluated, skipped, fmt.Errorf("create insight card: %w", err)
		}
		insights = append(insights, created)
	}

	return insights, evaluated, skipped, nil
}

// buildLabInsight creates an InsightCard for an out-of-range lab value.
func (uc *GenerateInsightsUseCase) buildLabInsight(
	input GenerateInsightsInput,
	m *entities.Measurement,
	ref *entities.ReferenceRange,
	flag entities.MeasurementFlag,
) *entities.InsightCard {
	priority := entities.InsightPriorityHigh
	category := entities.InsightCategoryNutrition

	if flag == entities.MeasurementFlagCriticalHigh || flag == entities.MeasurementFlagCriticalLow {
		priority = entities.InsightPriorityUrgent
		category = entities.InsightCategorySafety
	}

	title := fmt.Sprintf("Elevated %s", ref.DisplayName)
	rangeDesc := fmt.Sprintf("< %.0f %s", ref.HighNormal, ref.Unit)

	switch flag {
	case entities.MeasurementFlagLow:
		title = fmt.Sprintf("Low %s", ref.DisplayName)
		rangeDesc = fmt.Sprintf("> %.0f %s", ref.LowNormal, ref.Unit)
	case entities.MeasurementFlagCriticalLow:
		title = fmt.Sprintf("Critically Low %s", ref.DisplayName)
		rangeDesc = fmt.Sprintf("> %.0f %s", ref.LowNormal, ref.Unit)
	case entities.MeasurementFlagCriticalHigh:
		title = fmt.Sprintf("Critically Elevated %s", ref.DisplayName)
	}

	body := fmt.Sprintf(
		"%s is %.1f %s, outside the recommended range of %s.",
		ref.DisplayName, m.Value, m.Unit, rangeDesc,
	)
	if flag == entities.MeasurementFlagCriticalHigh || flag == entities.MeasurementFlagCriticalLow {
		body += " This requires immediate medical attention."
	}

	evidenceDesc := fmt.Sprintf("%s: %.1f %s (ref: %s)", ref.DisplayName, m.Value, m.Unit, rangeDesc)

	return &entities.InsightCard{
		CoachID:  input.CoachID,
		ClientID: input.ClientID,
		Title:    title,
		Body:     body,
		Category: category,
		Priority: priority,
		Status:   entities.InsightStatusDraft,
		Evidence: []entities.EvidenceRef{
			{
				MeasurementID: m.ID,
				ArtifactID:    input.ArtifactID,
				Description:   evidenceDesc,
			},
		},
	}
}

// evaluateWearableTrends checks wearable data for HRV decline and sleep issues.
func (uc *GenerateInsightsUseCase) evaluateWearableTrends(
	ctx context.Context,
	input GenerateInsightsInput,
	now time.Time,
) ([]*entities.InsightCard, int, error) {
	from := now.AddDate(0, 0, -wearableLookbackDays)
	summaries, err := uc.wearableRepo.FindByClientIDAndDateRange(ctx, input.ClientID, from, now)
	if err != nil {
		return nil, 0, err
	}

	if len(summaries) == 0 {
		return nil, 0, nil
	}

	// Sort by date ascending
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].SummaryDate.Before(summaries[j].SummaryDate)
	})

	var insights []*entities.InsightCard
	evaluated := 0

	// Check HRV decline
	hrvInsight := uc.checkHRVDecline(summaries, input, now)
	if hrvInsight != nil {
		evaluated++
		created, err := uc.insightRepo.Create(ctx, hrvInsight)
		if err != nil {
			return nil, evaluated, fmt.Errorf("create HRV insight: %w", err)
		}
		insights = append(insights, created)
	} else {
		evaluated++
	}

	// Check sleep decline
	sleepInsight := uc.checkSleepDecline(summaries, input, now)
	if sleepInsight != nil {
		evaluated++
		created, err := uc.insightRepo.Create(ctx, sleepInsight)
		if err != nil {
			return nil, evaluated, fmt.Errorf("create sleep insight: %w", err)
		}
		insights = append(insights, created)
	} else {
		evaluated++
	}

	return insights, evaluated, nil
}

// checkHRVDecline checks if HRV has declined >15% comparing current week to prior week.
func (uc *GenerateInsightsUseCase) checkHRVDecline(
	summaries []*entities.WearableSummary,
	input GenerateInsightsInput,
	now time.Time,
) *entities.InsightCard {
	midpoint := now.AddDate(0, 0, -7)

	var prevWeekHRV, currWeekHRV []float64
	for _, s := range summaries {
		hrv, ok := s.GetMetricFloat64("hrv")
		if !ok {
			continue
		}
		if s.SummaryDate.Before(midpoint) {
			prevWeekHRV = append(prevWeekHRV, hrv)
		} else {
			currWeekHRV = append(currWeekHRV, hrv)
		}
	}

	if len(prevWeekHRV) == 0 || len(currWeekHRV) == 0 {
		return nil
	}

	prevAvg := average(prevWeekHRV)
	currAvg := average(currWeekHRV)

	if prevAvg == 0 {
		return nil
	}

	decline := (prevAvg - currAvg) / prevAvg
	if decline <= hrvDeclineThreshold {
		return nil
	}

	declinePct := decline * 100
	return &entities.InsightCard{
		CoachID:  input.CoachID,
		ClientID: input.ClientID,
		Title:    "HRV Declining",
		Body: fmt.Sprintf(
			"HRV has declined %.0f%% this week (from %.0f to %.0f average). Consider reducing training volume.",
			declinePct, prevAvg, currAvg,
		),
		Category: entities.InsightCategoryRecovery,
		Priority: entities.InsightPriorityMedium,
		Status:   entities.InsightStatusDraft,
	}
}

// checkSleepDecline checks if average sleep over 7 days is below threshold.
func (uc *GenerateInsightsUseCase) checkSleepDecline(
	summaries []*entities.WearableSummary,
	input GenerateInsightsInput,
	now time.Time,
) *entities.InsightCard {
	weekAgo := now.AddDate(0, 0, -7)

	var recentSleepHours []float64
	for _, s := range summaries {
		if s.SummaryDate.Before(weekAgo) {
			continue
		}
		sleep, ok := s.GetMetricFloat64("sleep_hours")
		if !ok {
			continue
		}
		recentSleepHours = append(recentSleepHours, sleep)
	}

	if len(recentSleepHours) == 0 {
		return nil
	}

	avgSleep := average(recentSleepHours)
	if avgSleep >= sleepThresholdHours {
		return nil
	}

	return &entities.InsightCard{
		CoachID:  input.CoachID,
		ClientID: input.ClientID,
		Title:    "Sleep Declining",
		Body: fmt.Sprintf(
			"Average sleep dropped to %.1f hours over the past week. Sleep quality may be impacting recovery.",
			avgSleep,
		),
		Category: entities.InsightCategoryRecovery,
		Priority: entities.InsightPriorityMedium,
		Status:   entities.InsightStatusDraft,
	}
}

// evaluateBodyComposition checks for significant weight and body fat changes.
func (uc *GenerateInsightsUseCase) evaluateBodyComposition(
	ctx context.Context,
	input GenerateInsightsInput,
	now time.Time,
) ([]*entities.InsightCard, int, int, error) {
	since := now.AddDate(0, 0, -bodyCompLookbackDays)
	var insights []*entities.InsightCard
	evaluated := 0
	skipped := 0

	// Check weight change
	weightInsight, weightEval, weightSkip, err := uc.checkWeightChange(ctx, input, since, now)
	if err != nil {
		return nil, 0, 0, err
	}
	if weightInsight != nil {
		insights = append(insights, weightInsight)
	}
	evaluated += weightEval
	skipped += weightSkip

	// Check body fat change
	bfInsight, bfEval, bfSkip, err := uc.checkBodyFatChange(ctx, input, since, now)
	if err != nil {
		return nil, evaluated, skipped, err
	}
	if bfInsight != nil {
		insights = append(insights, bfInsight)
	}
	evaluated += bfEval
	skipped += bfSkip

	return insights, evaluated, skipped, nil
}

// checkWeightChange checks for >3% weight change in the lookback period.
func (uc *GenerateInsightsUseCase) checkWeightChange(
	ctx context.Context,
	input GenerateInsightsInput,
	since, now time.Time,
) (*entities.InsightCard, int, int, error) {
	weights, err := uc.measureRepo.FindByClientIDAndType(ctx, input.ClientID, "weight", since)
	if err != nil {
		return nil, 0, 0, err
	}
	if len(weights) < 2 {
		return nil, 0, 0, nil
	}

	// Sort by measured_at ascending
	sort.Slice(weights, func(i, j int) bool {
		return weights[i].MeasuredAt.Before(weights[j].MeasuredAt)
	})

	earliest := weights[0]
	latest := weights[len(weights)-1]

	if earliest.Value == 0 {
		return nil, 1, 0, nil
	}

	changePct := math.Abs(latest.Value-earliest.Value) / earliest.Value

	if changePct <= weightChangeThreshold {
		return nil, 1, 0, nil
	}

	// Check for duplicate
	exists, err := uc.insightRepo.ExistsByEvidence(ctx, input.ClientID, latest.ID)
	if err != nil {
		return nil, 1, 0, err
	}
	if exists {
		return nil, 1, 1, nil
	}

	direction := "decreased"
	if latest.Value > earliest.Value {
		direction = "increased"
	}

	card := &entities.InsightCard{
		CoachID:  input.CoachID,
		ClientID: input.ClientID,
		Title:    "Significant Weight Change",
		Body: fmt.Sprintf(
			"Weight %s %.1f%% in the past 2 weeks (from %.1f to %.1f %s). Verify this aligns with client goals.",
			direction, changePct*100, earliest.Value, latest.Value, latest.Unit,
		),
		Category: entities.InsightCategoryPerformance,
		Priority: entities.InsightPriorityMedium,
		Status:   entities.InsightStatusDraft,
		Evidence: []entities.EvidenceRef{
			{
				MeasurementID: latest.ID,
				ArtifactID:    input.ArtifactID,
				Description:   fmt.Sprintf("Weight: %.1f %s → %.1f %s", earliest.Value, earliest.Unit, latest.Value, latest.Unit),
			},
		},
	}

	created, err := uc.insightRepo.Create(ctx, card)
	if err != nil {
		return nil, 1, 0, err
	}
	return created, 1, 0, nil
}

// checkBodyFatChange checks for >1% body fat decrease.
func (uc *GenerateInsightsUseCase) checkBodyFatChange(
	ctx context.Context,
	input GenerateInsightsInput,
	since, now time.Time,
) (*entities.InsightCard, int, int, error) {
	bfReadings, err := uc.measureRepo.FindByClientIDAndType(ctx, input.ClientID, "body_fat_percentage", since)
	if err != nil {
		return nil, 0, 0, err
	}
	if len(bfReadings) < 2 {
		return nil, 0, 0, nil
	}

	// Sort by measured_at ascending
	sort.Slice(bfReadings, func(i, j int) bool {
		return bfReadings[i].MeasuredAt.Before(bfReadings[j].MeasuredAt)
	})

	earliest := bfReadings[0]
	latest := bfReadings[len(bfReadings)-1]
	decrease := earliest.Value - latest.Value

	if decrease < bodyFatDecreaseThreshold {
		return nil, 1, 0, nil
	}

	// Check for duplicate
	exists, err := uc.insightRepo.ExistsByEvidence(ctx, input.ClientID, latest.ID)
	if err != nil {
		return nil, 1, 0, err
	}
	if exists {
		return nil, 1, 1, nil
	}

	card := &entities.InsightCard{
		CoachID:  input.CoachID,
		ClientID: input.ClientID,
		Title:    "Body Fat Progress",
		Body: fmt.Sprintf(
			"Body fat decreased from %.1f%% to %.1f%% — great progress!",
			earliest.Value, latest.Value,
		),
		Category: entities.InsightCategoryPerformance,
		Priority: entities.InsightPriorityLow,
		Status:   entities.InsightStatusDraft,
		Evidence: []entities.EvidenceRef{
			{
				MeasurementID: latest.ID,
				ArtifactID:    input.ArtifactID,
				Description:   fmt.Sprintf("Body fat: %.1f%% → %.1f%%", earliest.Value, latest.Value),
			},
		},
	}

	created, err := uc.insightRepo.Create(ctx, card)
	if err != nil {
		return nil, 1, 0, err
	}
	return created, 1, 0, nil
}

// average calculates the mean of a slice of float64 values.
func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
