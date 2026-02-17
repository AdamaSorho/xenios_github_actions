package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

// insightJobPayload is the expected JSON payload for insight generation jobs.
type insightJobPayload struct {
	ClientID   string `json:"client_id"`
	CoachID    string `json:"coach_id"`
	ArtifactID string `json:"artifact_id"`
}

// InsightGenerationHandler processes insight_generation jobs from the job queue.
type InsightGenerationHandler struct {
	generateInsightsUC *usecase.GenerateInsightsUseCase
}

// NewInsightGenerationHandler creates a new InsightGenerationHandler.
func NewInsightGenerationHandler(generateInsightsUC *usecase.GenerateInsightsUseCase) *InsightGenerationHandler {
	return &InsightGenerationHandler{
		generateInsightsUC: generateInsightsUC,
	}
}

// Handle processes a single insight_generation job.
func (h *InsightGenerationHandler) Handle(ctx context.Context, job *entities.Job) error {
	var payload insightJobPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal insight job payload: %w", err)
	}

	if payload.ClientID == "" || payload.CoachID == "" || payload.ArtifactID == "" {
		return fmt.Errorf("insight job payload missing required fields: client_id, coach_id, artifact_id")
	}

	input := usecase.GenerateInsightsInput{
		ClientID:   payload.ClientID,
		CoachID:    payload.CoachID,
		ArtifactID: payload.ArtifactID,
	}

	output, err := h.generateInsightsUC.Execute(ctx, input)
	if err != nil {
		return fmt.Errorf("generate insights: %w", err)
	}

	log.Printf("Generated %d insight cards for client %s from artifact %s",
		len(output.InsightCards), payload.ClientID, payload.ArtifactID)

	return nil
}
