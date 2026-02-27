package entities

import "time"

// AuditEvent represents an immutable audit log entry.
type AuditEvent struct {
	ID         string                 `json:"id"`
	ActorID    string                 `json:"actor_id"`
	Action     string                 `json:"action"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// ValidAuditActions enumerates the known audit event actions.
var ValidAuditActions = map[string]bool{
	"user.login":        true,
	"user.logout":       true,
	"user.registered":   true,
	"artifact.upload":   true,
	"artifact.extract":  true,
	"insight.generate":  true,
	"insight.approve":   true,
	"insight.reject":    true,
	"summary.send":      true,
	"client.view":       true,
	"auth.login":        true,
	"auth.login_failed": true,
	"auth.logout":       true,
	"auth.token_refreshed":       true,
	"auth.token_replay_detected": true,
	"phi.access":                 true,
}

// IsValidAuditAction checks if an action string is a known audit action.
func IsValidAuditAction(action string) bool {
	return ValidAuditActions[action]
}

// AuditQueryFilter holds query parameters for listing audit events.
type AuditQueryFilter struct {
	ActorID    string
	Action     string
	EntityType string
	EntityID   string
	From       *time.Time
	To         *time.Time
	Limit      int
	Offset     int
}
