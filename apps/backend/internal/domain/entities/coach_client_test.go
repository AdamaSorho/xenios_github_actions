package entities

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCoachClient_Creation(t *testing.T) {
	now := time.Now()
	cc := CoachClient{
		ID:        "cc-1",
		CoachID:   "coach-1",
		ClientID:  "client-1",
		CreatedAt: now,
	}

	if cc.ID != "cc-1" {
		t.Errorf("expected ID cc-1, got %s", cc.ID)
	}
	if cc.CoachID != "coach-1" {
		t.Errorf("expected CoachID coach-1, got %s", cc.CoachID)
	}
	if cc.ClientID != "client-1" {
		t.Errorf("expected ClientID client-1, got %s", cc.ClientID)
	}
	if !cc.CreatedAt.Equal(now) {
		t.Errorf("expected CreatedAt %v, got %v", now, cc.CreatedAt)
	}
}

func TestCoachClient_JSONSerialization(t *testing.T) {
	cc := CoachClient{
		ID:        "cc-2",
		CoachID:   "coach-2",
		ClientID:  "client-2",
		CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(cc)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled CoachClient
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.ID != cc.ID {
		t.Errorf("expected ID %s, got %s", cc.ID, unmarshaled.ID)
	}
	if unmarshaled.CoachID != cc.CoachID {
		t.Errorf("expected CoachID %s, got %s", cc.CoachID, unmarshaled.CoachID)
	}
	if unmarshaled.ClientID != cc.ClientID {
		t.Errorf("expected ClientID %s, got %s", cc.ClientID, unmarshaled.ClientID)
	}
}

func TestCoachClient_EmptyFields(t *testing.T) {
	cc := CoachClient{}

	if cc.ID != "" {
		t.Errorf("expected empty ID, got %s", cc.ID)
	}
	if cc.CoachID != "" {
		t.Errorf("expected empty CoachID, got %s", cc.CoachID)
	}
	if cc.ClientID != "" {
		t.Errorf("expected empty ClientID, got %s", cc.ClientID)
	}
}
