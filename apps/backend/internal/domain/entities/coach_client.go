package entities

import "time"

// CoachClient represents a coaching relationship between a coach and a client.
type CoachClient struct {
	ID        string    `json:"id"`
	CoachID   string    `json:"coach_id"`
	ClientID  string    `json:"client_id"`
	CreatedAt time.Time `json:"created_at"`
}
