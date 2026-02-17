package usecase

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// AuthorizationError represents a forbidden access attempt.
type AuthorizationError struct {
	Message string
}

func (e *AuthorizationError) Error() string {
	return e.Message
}

// IsAuthorizationError checks whether the given error is an AuthorizationError.
func IsAuthorizationError(err error) bool {
	var ae *AuthorizationError
	return errors.As(err, &ae)
}

// GetClientMeasurementsUseCase retrieves measurements for a client with authorization.
type GetClientMeasurementsUseCase struct {
	measurementRepo repository.MeasurementRepository
	coachClientRepo repository.CoachClientRepository
	auditRepo       repository.AuditRepository
}

// NewGetClientMeasurementsUseCase creates a new GetClientMeasurementsUseCase.
func NewGetClientMeasurementsUseCase(
	measurementRepo repository.MeasurementRepository,
	coachClientRepo repository.CoachClientRepository,
	auditRepo repository.AuditRepository,
) *GetClientMeasurementsUseCase {
	return &GetClientMeasurementsUseCase{
		measurementRepo: measurementRepo,
		coachClientRepo: coachClientRepo,
		auditRepo:       auditRepo,
	}
}

// GetMeasurementsInput holds the input for retrieving client measurements.
type GetMeasurementsInput struct {
	CoachID         string
	ClientID        string
	MeasurementType string
	From            *time.Time
	To              *time.Time
	Page            int
	Limit           int
}

// GetMeasurementsOutput holds the paginated output.
type GetMeasurementsOutput struct {
	Measurements []*entities.Measurement `json:"measurements"`
	Pagination   PaginationInfo          `json:"pagination"`
}

// PaginationInfo holds pagination metadata.
type PaginationInfo struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// Execute retrieves measurements for a client after authorization check.
func (uc *GetClientMeasurementsUseCase) Execute(ctx context.Context, input GetMeasurementsInput) (*GetMeasurementsOutput, error) {
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if input.ClientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}

	// Verify coach-client relationship
	rel, err := uc.coachClientRepo.FindByCoachAndClient(ctx, input.CoachID, input.ClientID)
	if err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, &AuthorizationError{Message: "access denied: no coach-client relationship"}
	}

	// Apply defaults
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Limit <= 0 {
		input.Limit = defaultLimit
	}
	if input.Limit > maxLimit {
		input.Limit = maxLimit
	}

	offset := (input.Page - 1) * input.Limit

	filter := entities.MeasurementFilter{
		ClientID:        input.ClientID,
		MeasurementType: input.MeasurementType,
		From:            input.From,
		To:              input.To,
		Limit:           input.Limit,
		Offset:          offset,
	}

	measurements, total, err := uc.measurementRepo.FindByClientID(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Log PHI access audit event
	auditEvent := &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "phi.access",
		EntityType: "measurement",
		EntityID:   input.ClientID,
		Metadata: map[string]interface{}{
			"endpoint": "measurements",
			"type":     input.MeasurementType,
		},
		CreatedAt: time.Now(),
	}
	if err := uc.auditRepo.LogEvent(ctx, auditEvent); err != nil {
		log.Printf("failed to log audit event: %v", err)
	}

	if measurements == nil {
		measurements = []*entities.Measurement{}
	}

	return &GetMeasurementsOutput{
		Measurements: measurements,
		Pagination: PaginationInfo{
			Page:  input.Page,
			Limit: input.Limit,
			Total: total,
		},
	}, nil
}
