package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestGetQueueStatusUseCase_Execute_ReturnsStatus(t *testing.T) {
	expected := &entities.QueueStatus{
		Pending:    10,
		Active:     3,
		Completed:  50,
		Failed:     2,
		Expired:    1,
		DeadLetter: 0,
	}

	mock := &mockJobQueue{
		getStatusFunc: func(ctx context.Context) (*entities.QueueStatus, error) {
			return expected, nil
		},
	}

	uc := NewGetQueueStatusUseCase(mock)
	status, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.Pending != expected.Pending {
		t.Errorf("expected pending %d, got %d", expected.Pending, status.Pending)
	}
	if status.Active != expected.Active {
		t.Errorf("expected active %d, got %d", expected.Active, status.Active)
	}
	if status.Completed != expected.Completed {
		t.Errorf("expected completed %d, got %d", expected.Completed, status.Completed)
	}
	if status.Failed != expected.Failed {
		t.Errorf("expected failed %d, got %d", expected.Failed, status.Failed)
	}
	if status.Expired != expected.Expired {
		t.Errorf("expected expired %d, got %d", expected.Expired, status.Expired)
	}
	if status.DeadLetter != expected.DeadLetter {
		t.Errorf("expected dead_letter %d, got %d", expected.DeadLetter, status.DeadLetter)
	}
}

func TestGetQueueStatusUseCase_Execute_Error_PropagatesError(t *testing.T) {
	expectedErr := errors.New("database error")

	mock := &mockJobQueue{
		getStatusFunc: func(ctx context.Context) (*entities.QueueStatus, error) {
			return nil, expectedErr
		},
	}

	uc := NewGetQueueStatusUseCase(mock)
	_, err := uc.Execute(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %q, got %q", expectedErr, err)
	}
}

func TestGetQueueStatusUseCase_Execute_EmptyQueue_ReturnsZeroCounts(t *testing.T) {
	mock := &mockJobQueue{
		getStatusFunc: func(ctx context.Context) (*entities.QueueStatus, error) {
			return &entities.QueueStatus{}, nil
		},
	}

	uc := NewGetQueueStatusUseCase(mock)
	status, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.Pending != 0 || status.Active != 0 || status.Completed != 0 ||
		status.Failed != 0 || status.Expired != 0 || status.DeadLetter != 0 {
		t.Errorf("expected all zero counts, got %+v", status)
	}
}
