package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/xenios/backend/internal/domain"
)

func TestListCoachClientsUseCase_Execute_HappyPath(t *testing.T) {
	// Arrange
	repo := &mockCoachClientRepo{
		listByCoachFunc: func(ctx context.Context, coachID string, limit, offset int) ([]*domain.CoachClient, error) {
			return []*domain.CoachClient{
				{ID: "1", CoachID: coachID, ClientID: "client-1", Status: "active"},
				{ID: "2", CoachID: coachID, ClientID: "client-2", Status: "active"},
			}, nil
		},
	}
	uc := NewListCoachClientsUseCase(repo)
	input := ListCoachClientsInput{
		CoachID: "coach-1",
		Limit:   10,
		Offset:  0,
	}

	// Act
	result, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if result[0].ClientID != "client-1" {
		t.Errorf("expected first client 'client-1', got '%s'", result[0].ClientID)
	}
}

func TestListCoachClientsUseCase_Execute_DefaultLimit(t *testing.T) {
	// Arrange
	var capturedLimit int
	repo := &mockCoachClientRepo{
		listByCoachFunc: func(ctx context.Context, coachID string, limit, offset int) ([]*domain.CoachClient, error) {
			capturedLimit = limit
			return nil, nil
		},
	}
	uc := NewListCoachClientsUseCase(repo)
	input := ListCoachClientsInput{
		CoachID: "coach-1",
		Limit:   0, // should default to 20
		Offset:  0,
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if capturedLimit != 20 {
		t.Errorf("expected default limit 20, got %d", capturedLimit)
	}
}

func TestListCoachClientsUseCase_Execute_MaxLimit(t *testing.T) {
	// Arrange
	var capturedLimit int
	repo := &mockCoachClientRepo{
		listByCoachFunc: func(ctx context.Context, coachID string, limit, offset int) ([]*domain.CoachClient, error) {
			capturedLimit = limit
			return nil, nil
		},
	}
	uc := NewListCoachClientsUseCase(repo)
	input := ListCoachClientsInput{
		CoachID: "coach-1",
		Limit:   500, // should be capped at 100
		Offset:  0,
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if capturedLimit != 100 {
		t.Errorf("expected max limit 100, got %d", capturedLimit)
	}
}

func TestListCoachClientsUseCase_Execute_EmptyCoachID(t *testing.T) {
	// Arrange
	repo := &mockCoachClientRepo{}
	uc := NewListCoachClientsUseCase(repo)
	input := ListCoachClientsInput{
		CoachID: "",
		Limit:   10,
	}

	// Act
	result, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("expected error for empty coach_id")
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestListCoachClientsUseCase_Execute_NegativeLimit(t *testing.T) {
	// Arrange
	repo := &mockCoachClientRepo{}
	uc := NewListCoachClientsUseCase(repo)
	input := ListCoachClientsInput{
		CoachID: "coach-1",
		Limit:   -1,
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("expected error for negative limit")
	}
}

func TestListCoachClientsUseCase_Execute_NegativeOffset(t *testing.T) {
	// Arrange
	repo := &mockCoachClientRepo{}
	uc := NewListCoachClientsUseCase(repo)
	input := ListCoachClientsInput{
		CoachID: "coach-1",
		Limit:   10,
		Offset:  -1,
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("expected error for negative offset")
	}
}

func TestListCoachClientsUseCase_Execute_RepoError(t *testing.T) {
	// Arrange
	repoErr := errors.New("database unavailable")
	repo := &mockCoachClientRepo{
		listByCoachFunc: func(ctx context.Context, coachID string, limit, offset int) ([]*domain.CoachClient, error) {
			return nil, repoErr
		},
	}
	uc := NewListCoachClientsUseCase(repo)
	input := ListCoachClientsInput{
		CoachID: "coach-1",
		Limit:   10,
	}

	// Act
	result, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("expected error from repository")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repo error, got: %v", err)
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestListCoachClientsUseCase_Execute_EmptyResult(t *testing.T) {
	// Arrange
	repo := &mockCoachClientRepo{
		listByCoachFunc: func(ctx context.Context, coachID string, limit, offset int) ([]*domain.CoachClient, error) {
			return []*domain.CoachClient{}, nil
		},
	}
	uc := NewListCoachClientsUseCase(repo)
	input := ListCoachClientsInput{
		CoachID: "coach-1",
		Limit:   10,
	}

	// Act
	result, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 results, got %d", len(result))
	}
}

func TestListCoachClientsInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   ListCoachClientsInput
		wantErr bool
	}{
		{"valid", ListCoachClientsInput{CoachID: "c1", Limit: 10, Offset: 0}, false},
		{"zero limit (uses default)", ListCoachClientsInput{CoachID: "c1", Limit: 0, Offset: 0}, false},
		{"zero offset", ListCoachClientsInput{CoachID: "c1", Limit: 10, Offset: 0}, false},
		{"empty coach_id", ListCoachClientsInput{CoachID: "", Limit: 10}, true},
		{"negative limit", ListCoachClientsInput{CoachID: "c1", Limit: -1}, true},
		{"negative offset", ListCoachClientsInput{CoachID: "c1", Limit: 10, Offset: -1}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.Validate()
			if tc.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
