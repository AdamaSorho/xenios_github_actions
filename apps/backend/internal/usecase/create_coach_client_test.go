package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/xenios/backend/internal/domain"
)

type mockCoachClientRepo struct {
	createFunc      func(ctx context.Context, cc *domain.CoachClient) (*domain.CoachClient, error)
	listByCoachFunc func(ctx context.Context, coachID string, limit, offset int) ([]*domain.CoachClient, error)
}

func (m *mockCoachClientRepo) Create(ctx context.Context, cc *domain.CoachClient) (*domain.CoachClient, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, cc)
	}
	cc.ID = "generated-id"
	return cc, nil
}

func (m *mockCoachClientRepo) ListByCoachID(ctx context.Context, coachID string, limit, offset int) ([]*domain.CoachClient, error) {
	if m.listByCoachFunc != nil {
		return m.listByCoachFunc(ctx, coachID, limit, offset)
	}
	return nil, nil
}

func TestCreateCoachClientUseCase_Execute_HappyPath(t *testing.T) {
	// Arrange
	repo := &mockCoachClientRepo{}
	uc := NewCreateCoachClientUseCase(repo)
	input := CreateCoachClientInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
	}

	// Act
	result, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.CoachID != "coach-1" {
		t.Errorf("expected CoachID 'coach-1', got '%s'", result.CoachID)
	}
	if result.ClientID != "client-1" {
		t.Errorf("expected ClientID 'client-1', got '%s'", result.ClientID)
	}
	if result.Status != domain.CoachClientStatusActive {
		t.Errorf("expected Status '%s', got '%s'", domain.CoachClientStatusActive, result.Status)
	}
	if result.ID != "generated-id" {
		t.Errorf("expected ID 'generated-id', got '%s'", result.ID)
	}
}

func TestCreateCoachClientUseCase_Execute_EmptyCoachID(t *testing.T) {
	// Arrange
	repo := &mockCoachClientRepo{}
	uc := NewCreateCoachClientUseCase(repo)
	input := CreateCoachClientInput{
		CoachID:  "",
		ClientID: "client-1",
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
	if err.Error() != "coach_id is required" {
		t.Errorf("expected 'coach_id is required', got '%s'", err.Error())
	}
}

func TestCreateCoachClientUseCase_Execute_EmptyClientID(t *testing.T) {
	// Arrange
	repo := &mockCoachClientRepo{}
	uc := NewCreateCoachClientUseCase(repo)
	input := CreateCoachClientInput{
		CoachID:  "coach-1",
		ClientID: "",
	}

	// Act
	result, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("expected error for empty client_id")
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
	if err.Error() != "client_id is required" {
		t.Errorf("expected 'client_id is required', got '%s'", err.Error())
	}
}

func TestCreateCoachClientUseCase_Execute_SameCoachAndClient(t *testing.T) {
	// Arrange
	repo := &mockCoachClientRepo{}
	uc := NewCreateCoachClientUseCase(repo)
	input := CreateCoachClientInput{
		CoachID:  "user-1",
		ClientID: "user-1",
	}

	// Act
	result, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("expected error when coach_id equals client_id")
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestCreateCoachClientUseCase_Execute_RepoError(t *testing.T) {
	// Arrange
	repoErr := errors.New("database error")
	repo := &mockCoachClientRepo{
		createFunc: func(ctx context.Context, cc *domain.CoachClient) (*domain.CoachClient, error) {
			return nil, repoErr
		},
	}
	uc := NewCreateCoachClientUseCase(repo)
	input := CreateCoachClientInput{
		CoachID:  "coach-1",
		ClientID: "client-1",
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

func TestCreateCoachClientInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   CreateCoachClientInput
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid input",
			input:   CreateCoachClientInput{CoachID: "c1", ClientID: "cl1"},
			wantErr: false,
		},
		{
			name:    "empty coach_id",
			input:   CreateCoachClientInput{CoachID: "", ClientID: "cl1"},
			wantErr: true,
			errMsg:  "coach_id is required",
		},
		{
			name:    "empty client_id",
			input:   CreateCoachClientInput{CoachID: "c1", ClientID: ""},
			wantErr: true,
			errMsg:  "client_id is required",
		},
		{
			name:    "same IDs",
			input:   CreateCoachClientInput{CoachID: "x", ClientID: "x"},
			wantErr: true,
			errMsg:  "coach_id and client_id must be different",
		},
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
			if tc.wantErr && err != nil && err.Error() != tc.errMsg {
				t.Errorf("expected error '%s', got '%s'", tc.errMsg, err.Error())
			}
		})
	}
}
