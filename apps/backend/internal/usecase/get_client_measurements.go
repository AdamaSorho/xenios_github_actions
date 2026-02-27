package usecase

import (
	"context"
	"log"

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

// GetClientMeasurementsInput holds the input for retrieving client measurements.
type GetClientMeasurementsInput struct {
	CoachID  string
	ClientID string
	Filter   entities.MeasurementFilter
}

// Execute retrieves measurements for a client, enforcing coach-client authorization.
func (uc *GetClientMeasurementsUseCase) Execute(ctx context.Context, input GetClientMeasurementsInput) (*entities.MeasurementPage, error) {
	if err := validateCoachClientInput(input.CoachID, input.ClientID); err != nil {
		return nil, err
	}

	if err := uc.authorizeCoachClient(ctx, input.CoachID, input.ClientID); err != nil {
		return nil, err
	}

	input.Filter.ClientID = input.ClientID
	applyPaginationDefaults(&input.Filter)

	result, err := uc.measurementRepo.FindByClientID(ctx, input.Filter)
	if err != nil {
		return nil, err
	}

	uc.logPHIAccess(ctx, input.CoachID, input.ClientID, "measurements")
	return result, nil
}

func (uc *GetClientMeasurementsUseCase) authorizeCoachClient(ctx context.Context, coachID, clientID string) error {
	return authorizeCoachClient(ctx, uc.coachClientRepo, coachID, clientID)
}

func (uc *GetClientMeasurementsUseCase) logPHIAccess(ctx context.Context, coachID, clientID, resource string) {
	logPHIAccess(ctx, uc.auditRepo, coachID, clientID, resource)
}

// authorizeCoachClient checks if a coach-client relationship exists.
func authorizeCoachClient(ctx context.Context, repo repository.CoachClientRepository, coachID, clientID string) error {
	rel, err := repo.FindByCoachAndClient(ctx, coachID, clientID)
	if err != nil {
		return err
	}
	if rel == nil {
		return &AuthorizationError{Message: "not authorized to access this client"}
	}
	return nil
}

// validateCoachClientInput validates that coachID and clientID are provided.
func validateCoachClientInput(coachID, clientID string) error {
	if coachID == "" {
		return &ValidationError{Message: "coach_id is required"}
	}
	if clientID == "" {
		return &ValidationError{Message: "client_id is required"}
	}
	return nil
}

// applyPaginationDefaults sets default page and limit values.
func applyPaginationDefaults(filter *entities.MeasurementFilter) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = defaultLimit
	}
	if filter.Limit > maxLimit {
		filter.Limit = maxLimit
	}
}

// logPHIAccess logs a PHI access audit event.
func logPHIAccess(ctx context.Context, auditRepo repository.AuditRepository, coachID, clientID, resource string) {
	if auditErr := auditRepo.LogEvent(ctx, &entities.AuditEvent{
		ActorID:    coachID,
		Action:     "phi.access",
		EntityType: "client",
		EntityID:   clientID,
		Metadata:   map[string]interface{}{"resource": resource},
	}); auditErr != nil {
		log.Printf("audit log error: %v", auditErr)
	}
}

// AuthorizationError indicates a failed authorization check.
type AuthorizationError struct {
	Message string
}

func (e *AuthorizationError) Error() string { return e.Message }

// IsAuthorizationError checks if an error is an AuthorizationError.
func IsAuthorizationError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*AuthorizationError)
	return ok
}
