package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestListCoachClients_Success(t *testing.T) {
	repo := &mockCoachClientRepo{
		listFunc: func(ctx context.Context, coachID string, limit, offset int) ([]*entities.CoachClient, error) {
			return []*entities.CoachClient{
				{ID: "1", CoachID: coachID, ClientID: "client-1", CreatedAt: time.Now()},
				{ID: "2", CoachID: coachID, ClientID: "client-2", CreatedAt: time.Now()},
			}, nil
		},
	}
	uc := NewListCoachClientsUseCase(repo)

	results, err := uc.Execute(context.Background(), "coach-1", 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestListCoachClients_EmptyCoachID(t *testing.T) {
	repo := &mockCoachClientRepo{}
	uc := NewListCoachClientsUseCase(repo)

	_, err := uc.Execute(context.Background(), "", 20, 0)
	if err == nil {
		t.Fatal("expected error for empty coach_id")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestListCoachClients_NegativeLimit(t *testing.T) {
	repo := &mockCoachClientRepo{}
	uc := NewListCoachClientsUseCase(repo)

	_, err := uc.Execute(context.Background(), "coach-1", -1, 0)
	if err == nil {
		t.Fatal("expected error for negative limit")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestListCoachClients_NegativeOffset(t *testing.T) {
	repo := &mockCoachClientRepo{}
	uc := NewListCoachClientsUseCase(repo)

	_, err := uc.Execute(context.Background(), "coach-1", 20, -1)
	if err == nil {
		t.Fatal("expected error for negative offset")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestListCoachClients_DefaultLimit(t *testing.T) {
	var capturedLimit int
	repo := &mockCoachClientRepo{
		listFunc: func(ctx context.Context, coachID string, limit, offset int) ([]*entities.CoachClient, error) {
			capturedLimit = limit
			return []*entities.CoachClient{}, nil
		},
	}
	uc := NewListCoachClientsUseCase(repo)

	_, err := uc.Execute(context.Background(), "coach-1", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedLimit != 20 {
		t.Errorf("expected default limit 20, got %d", capturedLimit)
	}
}

func TestListCoachClients_MaxLimit(t *testing.T) {
	var capturedLimit int
	repo := &mockCoachClientRepo{
		listFunc: func(ctx context.Context, coachID string, limit, offset int) ([]*entities.CoachClient, error) {
			capturedLimit = limit
			return []*entities.CoachClient{}, nil
		},
	}
	uc := NewListCoachClientsUseCase(repo)

	_, err := uc.Execute(context.Background(), "coach-1", 500, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedLimit != 100 {
		t.Errorf("expected max limit 100, got %d", capturedLimit)
	}
}

func TestListCoachClients_RepoError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	repo := &mockCoachClientRepo{
		listFunc: func(ctx context.Context, coachID string, limit, offset int) ([]*entities.CoachClient, error) {
			return nil, repoErr
		},
	}
	uc := NewListCoachClientsUseCase(repo)

	_, err := uc.Execute(context.Background(), "coach-1", 20, 0)
	if err == nil {
		t.Fatal("expected error from repository")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repository error to propagate, got %v", err)
	}
}

func TestListCoachClients_EmptyResult(t *testing.T) {
	repo := &mockCoachClientRepo{
		listFunc: func(ctx context.Context, coachID string, limit, offset int) ([]*entities.CoachClient, error) {
			return []*entities.CoachClient{}, nil
		},
	}
	uc := NewListCoachClientsUseCase(repo)

	results, err := uc.Execute(context.Background(), "coach-1", 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestListCoachClients_OffsetBeyondResults(t *testing.T) {
	repo := &mockCoachClientRepo{
		listFunc: func(ctx context.Context, coachID string, limit, offset int) ([]*entities.CoachClient, error) {
			return []*entities.CoachClient{}, nil
		},
	}
	uc := NewListCoachClientsUseCase(repo)

	results, err := uc.Execute(context.Background(), "coach-1", 20, 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for offset beyond data, got %d", len(results))
	}
}
