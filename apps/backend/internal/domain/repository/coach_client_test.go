package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

// mockCoachClientRepository is a mock implementation for testing the interface.
type mockCoachClientRepository struct {
	shouldError bool
}

func (m *mockCoachClientRepository) Create(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error) {
	if m.shouldError {
		return nil, errors.New("create error")
	}
	return &entities.CoachClient{
		ID:       "mock-id",
		CoachID:  coachID,
		ClientID: clientID,
	}, nil
}

func (m *mockCoachClientRepository) ListByCoachID(ctx context.Context, coachID string, limit, offset int) ([]*entities.CoachClient, error) {
	if m.shouldError {
		return nil, errors.New("list error")
	}
	return []*entities.CoachClient{}, nil
}

func TestCoachClientRepository_Interface(t *testing.T) {
	var _ CoachClientRepository = (*mockCoachClientRepository)(nil)
}

func TestCoachClientRepository_MockCreate(t *testing.T) {
	repo := &mockCoachClientRepository{}
	cc, err := repo.Create(context.Background(), "coach-1", "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cc.CoachID != "coach-1" {
		t.Errorf("expected CoachID coach-1, got %s", cc.CoachID)
	}
}

func TestCoachClientRepository_MockCreateError(t *testing.T) {
	repo := &mockCoachClientRepository{shouldError: true}
	cc, err := repo.Create(context.Background(), "coach-1", "client-1")
	if err == nil {
		t.Error("expected error")
	}
	if cc != nil {
		t.Error("expected nil result on error")
	}
}

func TestCoachClientRepository_MockList(t *testing.T) {
	repo := &mockCoachClientRepository{}
	results, err := repo.ListByCoachID(context.Background(), "coach-1", 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results == nil {
		t.Error("expected non-nil results")
	}
}
