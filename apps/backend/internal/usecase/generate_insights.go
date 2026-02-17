package usecase

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GenerateInsightsUseCase evaluates recently extracted measurements and creates
// draft InsightCards for any triggered rules (out-of-range labs, wearable trends,
// body composition changes).
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

// GenerateInsightsInput holds the input for generating insights.
type GenerateInsightsInput struct {
	ClientID   string
	CoachID    string
	ArtifactID string
}

// GenerateInsightsOutput holds the result of insight generation.
type GenerateInsightsOutput struct {
	InsightCards []*entities.InsightCard
}

// Execute evaluates measurements from the given artifact and generates draft insight cards.
func (uc *GenerateInsightsUseCase) Execute(ctx context.Context, input GenerateInsightsInput) (*GenerateInsightsOutput, error) {
	if input.ClientID == "" {
		return nil, entities.NewValidationError("client_id is required")
	}
	if input.CoachID == "" {
		return nil, entities.NewValidationError("coach_id is required")
	}
	if input.ArtifactID == "" {
		return nil, entities.NewValidationError("artifact_id is required")
	}

	// Fetch measurements for this artifact
	artifactMeasurements, err := uc.measurementRepo.FindByArtifactID(ctx, input.ArtifactID)
	if err != nil {
		return nil, fmt.Errorf("find measurements by artifact: %w", err)
	}

	// Fetch wider time range for trend analysis (14 days)
	now := time.Now()
	twoWeeksAgo := now.AddDate(0, 0, -14)
	allMeasurements, err := uc.measurementRepo.FindByClientID(ctx, input.ClientID, twoWeeksAgo, now)
	if err != nil {
		return nil, fmt.Errorf("find measurements by client: %w", err)
	}

	var cards []*entities.InsightCard

	// Rule 1 & 2: Lab out-of-range and critical values
	labCards, err := uc.evaluateLabRules(ctx, input, artifactMeasurements)
	if err != nil {
		return nil, err
	}
	cards = append(cards, labCards...)

	// Rule 3: HRV declining trend
	hrvCards, err := uc.evaluateHRVTrend(ctx, input, allMeasurements)
	if err != nil {
		return nil, err
	}
	cards = append(cards, hrvCards...)

	// Rule 4: Sleep declining
	sleepCards, err := uc.evaluateSleepTrend(ctx, input, allMeasurements)
	if err != nil {
		return nil, err
	}
	cards = append(cards, sleepCards...)

	// Rule 5 & 6: Body composition changes (weight and body fat)
	bodyCompCards, err := uc.evaluateBodyComposition(ctx, input, allMeasurements)
	if err != nil {
		return nil, err
	}
	cards = append(cards, bodyCompCards...)

	return &GenerateInsightsOutput{InsightCards: cards}, nil
}

// evaluateLabRules checks for out-of-range and critical lab values.
func (uc *GenerateInsightsUseCase) evaluateLabRules(ctx context.Context, input GenerateInsightsInput, measurements []*entities.Measurement) ([]*entities.InsightCard, error) {
	var cards []*entities.InsightCard

	for _, m := range measurements {
		if m.Type != entities.MeasurementTypeLab {
			continue
		}

		if !m.IsOutOfRange() && !m.IsCritical() {
			continue
		}

		// Duplicate prevention
		exists, err := uc.insightRepo.ExistsByMeasurementID(ctx, m.ID)
		if err != nil {
			return nil, fmt.Errorf("check duplicate for measurement %s: %w", m.ID, err)
		}
		if exists {
			continue
		}

		card := uc.buildLabInsightCard(input, m)

		created, err := uc.insightRepo.Create(ctx, card)
		if err != nil {
			return nil, fmt.Errorf("create insight card: %w", err)
		}

		if err := uc.logAuditEvent(ctx, input.CoachID, created); err != nil {
			return nil, err
		}

		cards = append(cards, created)
	}

	return cards, nil
}

// buildLabInsightCard constructs an insight card for an out-of-range or critical lab value.
func (uc *GenerateInsightsUseCase) buildLabInsightCard(input GenerateInsightsInput, m *entities.Measurement) *entities.InsightCard {
	var title, body string
	var priority entities.InsightPriority
	var category entities.InsightCategory

	refDesc := buildReferenceDescription(m)

	if m.IsCritical() {
		priority = entities.InsightPriorityUrgent
		category = entities.InsightCategorySafety
		title = fmt.Sprintf("Critical %s Level", m.MarkerName)
		body = fmt.Sprintf("%s at %.1f %s requires immediate medical attention%s",
			m.MarkerName, m.Value, m.Unit, refDesc)
	} else {
		priority = entities.InsightPriorityHigh
		category = entities.InsightCategoryNutrition
		title = fmt.Sprintf("Elevated %s", m.MarkerName)
		if m.Flag == entities.MeasurementFlagLow {
			title = fmt.Sprintf("Low %s", m.MarkerName)
		}
		body = fmt.Sprintf("%s at %.1f %s is outside the recommended range%s",
			m.MarkerName, m.Value, m.Unit, refDesc)
	}

	evidenceDesc := fmt.Sprintf("%s: %.1f %s", m.MarkerName, m.Value, m.Unit)
	if refDesc != "" {
		evidenceDesc += " " + refDesc[2:] // strip leading " ("
		evidenceDesc = evidenceDesc[:len(evidenceDesc)-1] // strip trailing ")"
	}

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
				ArtifactID:    m.ArtifactID,
				Description:   evidenceDesc,
			},
		},
	}
}

// buildReferenceDescription creates a human-readable reference range string.
func buildReferenceDescription(m *entities.Measurement) string {
	if m.ReferenceMin != nil && m.ReferenceMax != nil {
		return fmt.Sprintf(" (ref: %.0f-%.0f %s)", *m.ReferenceMin, *m.ReferenceMax, m.Unit)
	}
	if m.ReferenceMax != nil {
		return fmt.Sprintf(" (ref: <%.0f %s)", *m.ReferenceMax, m.Unit)
	}
	if m.ReferenceMin != nil {
		return fmt.Sprintf(" (ref: >%.0f %s)", *m.ReferenceMin, m.Unit)
	}
	return ""
}

// evaluateHRVTrend checks if HRV has declined >15% over the last 7 days compared to the prior 7 days.
func (uc *GenerateInsightsUseCase) evaluateHRVTrend(ctx context.Context, input GenerateInsightsInput, measurements []*entities.Measurement) ([]*entities.InsightCard, error) {
	now := time.Now()
	sevenDaysAgo := now.AddDate(0, 0, -7)
	fourteenDaysAgo := now.AddDate(0, 0, -14)

	var recentHRV, priorHRV []float64
	var recentMeasurementIDs []string

	for _, m := range measurements {
		if m.Type != entities.MeasurementTypeWearable || m.MarkerName != "HRV" {
			continue
		}

		if !m.RecordedAt.Before(sevenDaysAgo) {
			recentHRV = append(recentHRV, m.Value)
			recentMeasurementIDs = append(recentMeasurementIDs, m.ID)
		} else if !m.RecordedAt.Before(fourteenDaysAgo) {
			priorHRV = append(priorHRV, m.Value)
		}
	}

	if len(recentHRV) == 0 || len(priorHRV) == 0 {
		return nil, nil
	}

	recentAvg := average(recentHRV)
	priorAvg := average(priorHRV)

	if priorAvg == 0 {
		return nil, nil
	}

	declinePercent := (priorAvg - recentAvg) / priorAvg * 100

	if declinePercent <= 15 {
		return nil, nil
	}

	// Use first recent measurement for duplicate check
	exists, err := uc.insightRepo.ExistsByMeasurementID(ctx, recentMeasurementIDs[0])
	if err != nil {
		return nil, fmt.Errorf("check duplicate for HRV trend: %w", err)
	}
	if exists {
		return nil, nil
	}

	card := &entities.InsightCard{
		CoachID:  input.CoachID,
		ClientID: input.ClientID,
		Title:    "HRV Declining Trend",
		Body:     fmt.Sprintf("HRV has declined %.0f%% this week (avg %.0f ms vs prior %.0f ms) — consider reducing training volume", declinePercent, recentAvg, priorAvg),
		Category: entities.InsightCategoryRecovery,
		Priority: entities.InsightPriorityMedium,
		Status:   entities.InsightStatusDraft,
		Evidence: []entities.EvidenceRef{
			{
				MeasurementID: recentMeasurementIDs[0],
				ArtifactID:    input.ArtifactID,
				Description:   fmt.Sprintf("HRV 7-day avg: %.0f ms (prior: %.0f ms, decline: %.0f%%)", recentAvg, priorAvg, declinePercent),
			},
		},
	}

	created, err := uc.insightRepo.Create(ctx, card)
	if err != nil {
		return nil, fmt.Errorf("create HRV insight card: %w", err)
	}

	if err := uc.logAuditEvent(ctx, input.CoachID, created); err != nil {
		return nil, err
	}

	return []*entities.InsightCard{created}, nil
}

// evaluateSleepTrend checks if the 7-day average sleep is below 6 hours.
func (uc *GenerateInsightsUseCase) evaluateSleepTrend(ctx context.Context, input GenerateInsightsInput, measurements []*entities.Measurement) ([]*entities.InsightCard, error) {
	now := time.Now()
	sevenDaysAgo := now.AddDate(0, 0, -7)

	var recentSleep []float64
	var recentMeasurementIDs []string

	for _, m := range measurements {
		if m.Type != entities.MeasurementTypeWearable || m.MarkerName != "Sleep Duration" {
			continue
		}
		if !m.RecordedAt.Before(sevenDaysAgo) {
			recentSleep = append(recentSleep, m.Value)
			recentMeasurementIDs = append(recentMeasurementIDs, m.ID)
		}
	}

	if len(recentSleep) == 0 {
		return nil, nil
	}

	avgSleep := average(recentSleep)

	if avgSleep >= 6.0 {
		return nil, nil
	}

	// Duplicate check
	exists, err := uc.insightRepo.ExistsByMeasurementID(ctx, recentMeasurementIDs[0])
	if err != nil {
		return nil, fmt.Errorf("check duplicate for sleep trend: %w", err)
	}
	if exists {
		return nil, nil
	}

	card := &entities.InsightCard{
		CoachID:  input.CoachID,
		ClientID: input.ClientID,
		Title:    "Low Average Sleep Duration",
		Body:     fmt.Sprintf("Average sleep dropped to %.1f hours — sleep quality may be impacting recovery", avgSleep),
		Category: entities.InsightCategoryRecovery,
		Priority: entities.InsightPriorityMedium,
		Status:   entities.InsightStatusDraft,
		Evidence: []entities.EvidenceRef{
			{
				MeasurementID: recentMeasurementIDs[0],
				ArtifactID:    input.ArtifactID,
				Description:   fmt.Sprintf("7-day avg sleep: %.1f hours (threshold: 6.0 hours)", avgSleep),
			},
		},
	}

	created, err := uc.insightRepo.Create(ctx, card)
	if err != nil {
		return nil, fmt.Errorf("create sleep insight card: %w", err)
	}

	if err := uc.logAuditEvent(ctx, input.CoachID, created); err != nil {
		return nil, err
	}

	return []*entities.InsightCard{created}, nil
}

// evaluateBodyComposition checks for significant weight and body fat changes over 2 weeks.
func (uc *GenerateInsightsUseCase) evaluateBodyComposition(ctx context.Context, input GenerateInsightsInput, measurements []*entities.Measurement) ([]*entities.InsightCard, error) {
	var cards []*entities.InsightCard

	weightCards, err := uc.evaluateWeightChange(ctx, input, measurements)
	if err != nil {
		return nil, err
	}
	cards = append(cards, weightCards...)

	bodyFatCards, err := uc.evaluateBodyFatChange(ctx, input, measurements)
	if err != nil {
		return nil, err
	}
	cards = append(cards, bodyFatCards...)

	return cards, nil
}

// evaluateWeightChange checks for >3% weight change in 2 weeks.
func (uc *GenerateInsightsUseCase) evaluateWeightChange(ctx context.Context, input GenerateInsightsInput, measurements []*entities.Measurement) ([]*entities.InsightCard, error) {
	var oldest, newest *entities.Measurement

	for _, m := range measurements {
		if m.Type != entities.MeasurementTypeBodyComp || m.MarkerName != "Weight" {
			continue
		}
		if oldest == nil || m.RecordedAt.Before(oldest.RecordedAt) {
			oldest = m
		}
		if newest == nil || m.RecordedAt.After(newest.RecordedAt) {
			newest = m
		}
	}

	if oldest == nil || newest == nil || oldest.ID == newest.ID {
		return nil, nil
	}

	// Only trigger if the measurements span at least 7 days
	const minSpan = 7 * 24 * time.Hour
	if newest.RecordedAt.Sub(oldest.RecordedAt) < minSpan {
		return nil, nil
	}

	changePercent := math.Abs((newest.Value - oldest.Value) / oldest.Value * 100)
	if changePercent < 3.0 {
		return nil, nil
	}

	// Duplicate check
	exists, err := uc.insightRepo.ExistsByMeasurementID(ctx, newest.ID)
	if err != nil {
		return nil, fmt.Errorf("check duplicate for weight change: %w", err)
	}
	if exists {
		return nil, nil
	}

	direction := "decreased"
	if newest.Value > oldest.Value {
		direction = "increased"
	}

	card := &entities.InsightCard{
		CoachID:  input.CoachID,
		ClientID: input.ClientID,
		Title:    "Significant Weight Change",
		Body:     fmt.Sprintf("Weight %s %.1f%% in %d days (%.1f → %.1f %s) — verify this aligns with client goals", direction, changePercent, int(newest.RecordedAt.Sub(oldest.RecordedAt).Hours()/24), oldest.Value, newest.Value, newest.Unit),
		Category: entities.InsightCategoryPerformance,
		Priority: entities.InsightPriorityMedium,
		Status:   entities.InsightStatusDraft,
		Evidence: []entities.EvidenceRef{
			{
				MeasurementID: newest.ID,
				ArtifactID:    newest.ArtifactID,
				Description:   fmt.Sprintf("Weight: %.1f → %.1f %s (%.1f%% change)", oldest.Value, newest.Value, newest.Unit, changePercent),
			},
		},
	}

	created, err := uc.insightRepo.Create(ctx, card)
	if err != nil {
		return nil, fmt.Errorf("create weight insight card: %w", err)
	}

	if err := uc.logAuditEvent(ctx, input.CoachID, created); err != nil {
		return nil, err
	}

	return []*entities.InsightCard{created}, nil
}

// evaluateBodyFatChange checks for >1% body fat decrease.
func (uc *GenerateInsightsUseCase) evaluateBodyFatChange(ctx context.Context, input GenerateInsightsInput, measurements []*entities.Measurement) ([]*entities.InsightCard, error) {
	var oldest, newest *entities.Measurement

	for _, m := range measurements {
		if m.Type != entities.MeasurementTypeBodyComp || m.MarkerName != "Body Fat" {
			continue
		}
		if oldest == nil || m.RecordedAt.Before(oldest.RecordedAt) {
			oldest = m
		}
		if newest == nil || m.RecordedAt.After(newest.RecordedAt) {
			newest = m
		}
	}

	if oldest == nil || newest == nil || oldest.ID == newest.ID {
		return nil, nil
	}

	decrease := oldest.Value - newest.Value
	if decrease < 1.0 {
		return nil, nil
	}

	// Duplicate check
	exists, err := uc.insightRepo.ExistsByMeasurementID(ctx, newest.ID)
	if err != nil {
		return nil, fmt.Errorf("check duplicate for body fat change: %w", err)
	}
	if exists {
		return nil, nil
	}

	card := &entities.InsightCard{
		CoachID:  input.CoachID,
		ClientID: input.ClientID,
		Title:    "Body Fat Progress",
		Body:     fmt.Sprintf("Body fat decreased from %.1f%% to %.1f%% — great progress!", oldest.Value, newest.Value),
		Category: entities.InsightCategoryPerformance,
		Priority: entities.InsightPriorityLow,
		Status:   entities.InsightStatusDraft,
		Evidence: []entities.EvidenceRef{
			{
				MeasurementID: newest.ID,
				ArtifactID:    newest.ArtifactID,
				Description:   fmt.Sprintf("Body Fat: %.1f%% → %.1f%% (%.1f%% decrease)", oldest.Value, newest.Value, decrease),
			},
		},
	}

	created, err := uc.insightRepo.Create(ctx, card)
	if err != nil {
		return nil, fmt.Errorf("create body fat insight card: %w", err)
	}

	if err := uc.logAuditEvent(ctx, input.CoachID, created); err != nil {
		return nil, err
	}

	return []*entities.InsightCard{created}, nil
}

// logAuditEvent logs an audit event for a generated insight card.
func (uc *GenerateInsightsUseCase) logAuditEvent(ctx context.Context, coachID string, card *entities.InsightCard) error {
	event := &entities.AuditEvent{
		ActorID:    coachID,
		Action:     "insight.generate",
		EntityType: "insight_card",
		EntityID:   card.ID,
		Metadata: map[string]interface{}{
			"title":    card.Title,
			"category": string(card.Category),
			"priority": string(card.Priority),
			"client_id": card.ClientID,
		},
	}
	if err := uc.auditRepo.LogEvent(ctx, event); err != nil {
		return fmt.Errorf("log audit event: %w", err)
	}
	return nil
}

// average computes the arithmetic mean of a float64 slice.
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
