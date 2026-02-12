package domain

import (
	"encoding/json"
	"testing"
)

func TestCoachClient_Creation(t *testing.T) {
	cc := CoachClient{
		ID:       "rel-1",
		CoachID:  "coach-1",
		ClientID: "client-1",
		Status:   CoachClientStatusActive,
	}

	if cc.ID != "rel-1" {
		t.Errorf("expected ID 'rel-1', got '%s'", cc.ID)
	}
	if cc.CoachID != "coach-1" {
		t.Errorf("expected CoachID 'coach-1', got '%s'", cc.CoachID)
	}
	if cc.ClientID != "client-1" {
		t.Errorf("expected ClientID 'client-1', got '%s'", cc.ClientID)
	}
	if cc.Status != CoachClientStatusActive {
		t.Errorf("expected Status '%s', got '%s'", CoachClientStatusActive, cc.Status)
	}
}

func TestCoachClient_JSONSerialization(t *testing.T) {
	cc := CoachClient{
		ID:       "rel-1",
		CoachID:  "coach-1",
		ClientID: "client-1",
		Status:   CoachClientStatusActive,
	}

	data, err := json.Marshal(cc)
	if err != nil {
		t.Fatalf("failed to marshal CoachClient: %v", err)
	}

	var decoded CoachClient
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal CoachClient: %v", err)
	}

	if decoded.ID != cc.ID {
		t.Errorf("expected ID '%s', got '%s'", cc.ID, decoded.ID)
	}
	if decoded.CoachID != cc.CoachID {
		t.Errorf("expected CoachID '%s', got '%s'", cc.CoachID, decoded.CoachID)
	}
	if decoded.ClientID != cc.ClientID {
		t.Errorf("expected ClientID '%s', got '%s'", cc.ClientID, decoded.ClientID)
	}
	if decoded.Status != cc.Status {
		t.Errorf("expected Status '%s', got '%s'", cc.Status, decoded.Status)
	}
}

func TestCoachClient_JSONFieldNames(t *testing.T) {
	cc := CoachClient{
		ID:       "id-1",
		CoachID:  "coach-1",
		ClientID: "client-1",
		Status:   "active",
	}

	data, err := json.Marshal(cc)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMap); err != nil {
		t.Fatalf("failed to unmarshal as raw map: %v", err)
	}

	expectedFields := []string{"id", "coach_id", "client_id", "status"}
	for _, field := range expectedFields {
		if _, ok := rawMap[field]; !ok {
			t.Errorf("expected '%s' field in JSON", field)
		}
	}
}

func TestCoachClientConstants(t *testing.T) {
	if CoachClientStatusActive != "active" {
		t.Errorf("expected CoachClientStatusActive 'active', got '%s'", CoachClientStatusActive)
	}
	if CoachClientStatusInactive != "inactive" {
		t.Errorf("expected CoachClientStatusInactive 'inactive', got '%s'", CoachClientStatusInactive)
	}
}

func TestCoachClient_ZeroValue(t *testing.T) {
	var cc CoachClient
	if cc.ID != "" {
		t.Errorf("expected empty ID, got '%s'", cc.ID)
	}
	if cc.CoachID != "" {
		t.Errorf("expected empty CoachID, got '%s'", cc.CoachID)
	}
	if cc.ClientID != "" {
		t.Errorf("expected empty ClientID, got '%s'", cc.ClientID)
	}
	if cc.Status != "" {
		t.Errorf("expected empty Status, got '%s'", cc.Status)
	}
}
