package repository

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/domain"
)

func TestInMemoryCoachClientRepository_Create(t *testing.T) {
	// Arrange
	repo := NewInMemoryCoachClientRepository()
	cc := &domain.CoachClient{
		CoachID:  "coach-1",
		ClientID: "client-1",
		Status:   "active",
	}

	// Act
	result, err := repo.Create(context.Background(), cc)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ID == "" {
		t.Error("expected ID to be generated")
	}
	if result.CoachID != "coach-1" {
		t.Errorf("expected CoachID 'coach-1', got '%s'", result.CoachID)
	}
}

func TestInMemoryCoachClientRepository_ListByCoachID(t *testing.T) {
	// Arrange
	repo := NewInMemoryCoachClientRepository()
	repo.Create(context.Background(), &domain.CoachClient{CoachID: "coach-1", ClientID: "client-1", Status: "active"})
	repo.Create(context.Background(), &domain.CoachClient{CoachID: "coach-1", ClientID: "client-2", Status: "active"})
	repo.Create(context.Background(), &domain.CoachClient{CoachID: "coach-2", ClientID: "client-3", Status: "active"})

	// Act
	results, err := repo.ListByCoachID(context.Background(), "coach-1", 10, 0)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestInMemoryCoachClientRepository_ListByCoachID_Pagination(t *testing.T) {
	// Arrange
	repo := NewInMemoryCoachClientRepository()
	for i := 0; i < 5; i++ {
		repo.Create(context.Background(), &domain.CoachClient{
			CoachID:  "coach-1",
			ClientID: "client-" + string(rune('a'+i)),
			Status:   "active",
		})
	}

	// Act - get page 2 with limit 2
	results, err := repo.ListByCoachID(context.Background(), "coach-1", 2, 2)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results with offset 2 limit 2, got %d", len(results))
	}
}

func TestInMemoryCoachClientRepository_ListByCoachID_OffsetBeyondResults(t *testing.T) {
	// Arrange
	repo := NewInMemoryCoachClientRepository()
	repo.Create(context.Background(), &domain.CoachClient{CoachID: "coach-1", ClientID: "client-1", Status: "active"})

	// Act
	results, err := repo.ListByCoachID(context.Background(), "coach-1", 10, 100)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for offset beyond data, got %d", len(results))
	}
}

func TestInMemoryCoachClientRepository_ListByCoachID_NoResults(t *testing.T) {
	// Arrange
	repo := NewInMemoryCoachClientRepository()

	// Act
	results, err := repo.ListByCoachID(context.Background(), "nonexistent", 10, 0)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestInMemoryCoachClientRepository_ConcurrentAccess(t *testing.T) {
	// Arrange
	repo := NewInMemoryCoachClientRepository()
	done := make(chan bool, 20)

	// Act - concurrent creates and reads
	for i := 0; i < 10; i++ {
		go func() {
			repo.Create(context.Background(), &domain.CoachClient{
				CoachID:  "coach-1",
				ClientID: "client",
				Status:   "active",
			})
			done <- true
		}()
		go func() {
			repo.ListByCoachID(context.Background(), "coach-1", 10, 0)
			done <- true
		}()
	}

	// Assert - wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}
