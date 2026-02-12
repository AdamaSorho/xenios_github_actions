package domain

import "context"

// CoachClient represents the relationship between a coach and a client.
type CoachClient struct {
	ID       string `json:"id"`
	CoachID  string `json:"coach_id"`
	ClientID string `json:"client_id"`
	Status   string `json:"status"` // "active", "inactive"
}

// CoachClient status constants
const (
	CoachClientStatusActive   = "active"
	CoachClientStatusInactive = "inactive"
)

// CoachClientRepository defines the interface for coach-client relationship persistence.
type CoachClientRepository interface {
	// Create persists a new coach-client relationship.
	Create(ctx context.Context, cc *CoachClient) (*CoachClient, error)
	// ListByCoachID returns all clients for a given coach.
	ListByCoachID(ctx context.Context, coachID string, limit, offset int) ([]*CoachClient, error)
}
