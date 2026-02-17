package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

// LabExtractJobPayload is the JSON payload for extract_lab_results jobs.
type LabExtractJobPayload struct {
	ArtifactID string `json:"artifact_id"`
	CoachID    string `json:"coach_id"`
}

// NewLabExtractJobHandler creates a worker.JobHandler for extract_lab_results jobs.
func NewLabExtractJobHandler(uc *usecase.ExtractLabResultsUseCase) func(ctx context.Context, job *entities.Job) error {
	return func(ctx context.Context, job *entities.Job) error {
		var payload LabExtractJobPayload
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return fmt.Errorf("unmarshal job payload: %w", err)
		}

		if payload.ArtifactID == "" {
			return fmt.Errorf("artifact_id is required in job payload")
		}
		if payload.CoachID == "" {
			return fmt.Errorf("coach_id is required in job payload")
		}

		output, err := uc.Execute(ctx, usecase.ExtractLabResultsInput{
			ArtifactID: payload.ArtifactID,
			CoachID:    payload.CoachID,
		})
		if err != nil {
			return fmt.Errorf("extract lab results: %w", err)
		}

		log.Printf("Lab extraction completed for artifact %s: %d markers extracted, %d flagged",
			output.ArtifactID, output.ExtractedCount, output.FlaggedCount)

		return nil
	}
}
