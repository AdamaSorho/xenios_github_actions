package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

// InsightJobPayload is the expected payload for insight_generation jobs.
type InsightJobPayload struct {
	ClientID   string `json:"client_id"`
	CoachID    string `json:"coach_id"`
	ArtifactID string `json:"artifact_id,omitempty"`
}

// NewInsightGenerationHandler returns a JobHandler that generates insight cards
// from recently extracted health data.
func NewInsightGenerationHandler(uc *usecase.GenerateInsightsUseCase) JobHandler {
	return func(ctx context.Context, job *entities.Job) error {
		var payload InsightJobPayload
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return fmt.Errorf("parse insight generation payload: %w", err)
		}

		if payload.ClientID == "" || payload.CoachID == "" {
			return fmt.Errorf("insight generation payload requires client_id and coach_id")
		}

		result, err := uc.Execute(ctx, usecase.GenerateInsightsInput{
			ClientID:   payload.ClientID,
			CoachID:    payload.CoachID,
			ArtifactID: payload.ArtifactID,
		})
		if err != nil {
			return fmt.Errorf("generate insights: %w", err)
		}

		log.Printf("Insight generation complete for client %s: %d insights created, %d evaluated, %d duplicates skipped",
			payload.ClientID, len(result.InsightsCreated), result.TotalEvaluated, result.DuplicatesSkipped)

		return nil
	}
}
