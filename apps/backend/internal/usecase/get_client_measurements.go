package usecase

import (
	"context"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GetClientMeasurementsUseCase handles retrieving measurements for a client.
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

const (
	defaultMeasurementLimit = 20
	maxMeasurementLimit     = 100
)

// GetClientMeasurementsInput holds the input parameters.
type GetClientMeasurementsInput struct {
	CoachID  string
	ClientID string
	Type     string
	From     *time.Time
	To       *time.Time
	Page     int
	Limit    int
}

// GetClientMeasurementsOutput holds the paginated result.
type GetClientMeasurementsOutput struct {
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
func (uc *GetClientMeasurementsUseCase) Execute(ctx context.Context, input GetClientMeasurementsInput) (*GetClientMeasurementsOutput, error) {
	if input.CoachID == "" {
		return nil, &ValidationError{Message: "coach_id is required"}
	}
	if input.ClientID == "" {
		return nil, &ValidationError{Message: "client_id is required"}
	}

	// Authorization: verify coach-client relationship
	rel, err := uc.coachClientRepo.FindByCoachAndClient(ctx, input.CoachID, input.ClientID)
	if err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, &AuthorizationError{Message: "not authorized to access this client's data"}
	}

	// Normalize pagination
	page := input.Page
	if page < 1 {
		page = 1
	}
	limit := input.Limit
	if limit <= 0 {
		limit = defaultMeasurementLimit
	}
	if limit > maxMeasurementLimit {
		limit = maxMeasurementLimit
	}
	offset := (page - 1) * limit

	filter := entities.MeasurementFilter{
		ClientID: input.ClientID,
		Type:     input.Type,
		From:     input.From,
		To:       input.To,
		Limit:    limit,
		Offset:   offset,
	}

	measurements, total, err := uc.measurementRepo.FindByClientID(ctx, filter)
	if err != nil {
		return nil, err
	}

	if measurements == nil {
		measurements = []*entities.Measurement{}
	}

	// Log PHI access audit event
	_ = uc.auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    input.CoachID,
		Action:     "phi.measurements_accessed",
		EntityType: "client",
		EntityID:   input.ClientID,
		Metadata: map[string]interface{}{
			"type":  input.Type,
			"page":  page,
			"limit": limit,
		},
	})

	return &GetClientMeasurementsOutput{
		Measurements: measurements,
		Pagination: PaginationInfo{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	}, nil
}
