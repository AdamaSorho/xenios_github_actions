package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

type mockCoachClientRepo struct {
	createFunc          func(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error)
	listFunc            func(ctx context.Context, coachID string, limit, offset int) ([]*entities.CoachClient, error)
	findByCoachAndClient func(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error)
}

func (m *mockCoachClientRepo) Create(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, coachID, clientID)
	}
	return &entities.CoachClient{ID: "new-id", CoachID: coachID, ClientID: clientID}, nil
}

func (m *mockCoachClientRepo) ListByCoachID(ctx context.Context, coachID string, limit, offset int) ([]*entities.CoachClient, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, coachID, limit, offset)
	}
	return []*entities.CoachClient{}, nil
}

func (m *mockCoachClientRepo) FindByCoachAndClient(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error) {
	if m.findByCoachAndClient != nil {
		return m.findByCoachAndClient(ctx, coachID, clientID)
	}
	return &entities.CoachClient{ID: "rel-1", CoachID: coachID, ClientID: clientID}, nil
}

func TestCreateCoachClient_Success(t *testing.T) {
	repo := &mockCoachClientRepo{}
	uc := NewCreateCoachClientUseCase(repo)

	cc, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cc == nil {
		t.Fatal("expected non-nil result")
	}
	if cc.CoachID != "coach-1" {
		t.Errorf("expected CoachID coach-1, got %s", cc.CoachID)
	}
	if cc.ClientID != "client-1" {
		t.Errorf("expected ClientID client-1, got %s", cc.ClientID)
	}
}

func TestCreateCoachClient_EmptyCoachID(t *testing.T) {
	repo := &mockCoachClientRepo{}
	uc := NewCreateCoachClientUseCase(repo)

	_, err := uc.Execute(context.Background(), "", "client-1")
	if err == nil {
		t.Fatal("expected error for empty coach_id")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
	if err.Error() != "coach_id is required" {
		t.Errorf("expected 'coach_id is required', got %s", err.Error())
	}
}

func TestCreateCoachClient_EmptyClientID(t *testing.T) {
	repo := &mockCoachClientRepo{}
	uc := NewCreateCoachClientUseCase(repo)

	_, err := uc.Execute(context.Background(), "coach-1", "")
	if err == nil {
		t.Fatal("expected error for empty client_id")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
	if err.Error() != "client_id is required" {
		t.Errorf("expected 'client_id is required', got %s", err.Error())
	}
}

func TestCreateCoachClient_SameCoachAndClient(t *testing.T) {
	repo := &mockCoachClientRepo{}
	uc := NewCreateCoachClientUseCase(repo)

	_, err := uc.Execute(context.Background(), "user-1", "user-1")
	if err == nil {
		t.Fatal("expected error when coach_id equals client_id")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestCreateCoachClient_RepoError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	repo := &mockCoachClientRepo{
		createFunc: func(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error) {
			return nil, repoErr
		},
	}
	uc := NewCreateCoachClientUseCase(repo)

	_, err := uc.Execute(context.Background(), "coach-1", "client-1")
	if err == nil {
		t.Fatal("expected error from repository")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repository error to propagate, got %v", err)
	}
	if IsValidationError(err) {
		t.Error("repo error should not be a ValidationError")
	}
}

func TestIsValidationError_True(t *testing.T) {
	err := &ValidationError{Message: "test"}
	if !IsValidationError(err) {
		t.Error("expected true for ValidationError")
	}
}

func TestIsValidationError_False(t *testing.T) {
	err := errors.New("not a validation error")
	if IsValidationError(err) {
		t.Error("expected false for non-ValidationError")
	}
}

func TestAuthorizationError_Error(t *testing.T) {
	err := &AuthorizationError{Message: "access denied"}
	if err.Error() != "access denied" {
		t.Errorf("expected 'access denied', got %s", err.Error())
	}
}

func TestIsAuthorizationError_True(t *testing.T) {
	err := &AuthorizationError{Message: "denied"}
	if !IsAuthorizationError(err) {
		t.Error("expected true for AuthorizationError")
	}
}

func TestIsAuthorizationError_False(t *testing.T) {
	err := errors.New("not an authorization error")
	if IsAuthorizationError(err) {
		t.Error("expected false for non-AuthorizationError")
	}
}
