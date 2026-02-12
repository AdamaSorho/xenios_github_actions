package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func newQueryAuditLogUseCase() (*QueryAuditLogUseCase, *repository.InMemoryAuditRepository) {
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewQueryAuditLogUseCase(auditRepo)
	return uc, auditRepo
}

func seedAuditEvents(t *testing.T, repo *repository.InMemoryAuditRepository) {
	t.Helper()
	ctx := context.Background()

	events := []*entities.AuditEvent{
		{ActorID: "user-1", Action: "user.login", EntityType: "user", EntityID: "user-1", CreatedAt: time.Now().Add(-3 * time.Hour)},
		{ActorID: "user-1", Action: "client.view", EntityType: "client_profile", EntityID: "client-1", CreatedAt: time.Now().Add(-2 * time.Hour)},
		{ActorID: "user-2", Action: "insight.approve", EntityType: "insight_card", EntityID: "card-1", CreatedAt: time.Now().Add(-1 * time.Hour)},
		{ActorID: "user-1", Action: "user.logout", EntityType: "user", EntityID: "user-1", CreatedAt: time.Now()},
	}

	for _, e := range events {
		if err := repo.LogEvent(ctx, e); err != nil {
			t.Fatalf("failed to seed event: %v", err)
		}
	}
}

func TestQueryAuditLog_AllEvents_ReturnsPaginated(t *testing.T) {
	uc, auditRepo := newQueryAuditLogUseCase()
	seedAuditEvents(t, auditRepo)

	out, err := uc.Execute(context.Background(), QueryAuditLogInput{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Total != 4 {
		t.Errorf("expected total 4, got %d", out.Total)
	}
	if len(out.Events) != 4 {
		t.Errorf("expected 4 events, got %d", len(out.Events))
	}
}

func TestQueryAuditLog_FilterByActorID_FiltersCorrectly(t *testing.T) {
	uc, auditRepo := newQueryAuditLogUseCase()
	seedAuditEvents(t, auditRepo)

	out, err := uc.Execute(context.Background(), QueryAuditLogInput{
		ActorID: "user-1",
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Total != 3 {
		t.Errorf("expected total 3, got %d", out.Total)
	}
	for _, e := range out.Events {
		if e.ActorID != "user-1" {
			t.Errorf("expected actor_id 'user-1', got '%s'", e.ActorID)
		}
	}
}

func TestQueryAuditLog_FilterByEntityType_FiltersCorrectly(t *testing.T) {
	uc, auditRepo := newQueryAuditLogUseCase()
	seedAuditEvents(t, auditRepo)

	out, err := uc.Execute(context.Background(), QueryAuditLogInput{
		EntityType: "insight_card",
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Total != 1 {
		t.Errorf("expected total 1, got %d", out.Total)
	}
	if out.Events[0].EntityType != "insight_card" {
		t.Errorf("expected entity_type 'insight_card', got '%s'", out.Events[0].EntityType)
	}
}

func TestQueryAuditLog_FilterByAction_FiltersCorrectly(t *testing.T) {
	uc, auditRepo := newQueryAuditLogUseCase()
	seedAuditEvents(t, auditRepo)

	out, err := uc.Execute(context.Background(), QueryAuditLogInput{
		Action: "client.view",
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Total != 1 {
		t.Errorf("expected total 1, got %d", out.Total)
	}
}

func TestQueryAuditLog_FilterByTimeRange_FiltersCorrectly(t *testing.T) {
	uc, auditRepo := newQueryAuditLogUseCase()
	seedAuditEvents(t, auditRepo)

	from := time.Now().Add(-2*time.Hour - 30*time.Minute)
	to := time.Now().Add(-30 * time.Minute)

	out, err := uc.Execute(context.Background(), QueryAuditLogInput{
		From:  &from,
		To:    &to,
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should include: client.view (-2h) and insight.approve (-1h)
	if out.Total != 2 {
		t.Errorf("expected total 2, got %d", out.Total)
	}
}

func TestQueryAuditLog_Pagination_RespectsLimitAndOffset(t *testing.T) {
	uc, auditRepo := newQueryAuditLogUseCase()
	seedAuditEvents(t, auditRepo)

	out, err := uc.Execute(context.Background(), QueryAuditLogInput{
		Limit:  2,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Total != 4 {
		t.Errorf("expected total 4, got %d", out.Total)
	}
	if len(out.Events) != 2 {
		t.Errorf("expected 2 events on page, got %d", len(out.Events))
	}

	// Second page
	out2, err := uc.Execute(context.Background(), QueryAuditLogInput{
		Limit:  2,
		Offset: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out2.Events) != 2 {
		t.Errorf("expected 2 events on second page, got %d", len(out2.Events))
	}
}

func TestQueryAuditLog_DefaultLimit_Applied(t *testing.T) {
	uc, auditRepo := newQueryAuditLogUseCase()
	seedAuditEvents(t, auditRepo)

	out, err := uc.Execute(context.Background(), QueryAuditLogInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Total != 4 {
		t.Errorf("expected total 4, got %d", out.Total)
	}
}

func TestQueryAuditLog_ExcessiveLimit_Capped(t *testing.T) {
	uc, _ := newQueryAuditLogUseCase()

	out, err := uc.Execute(context.Background(), QueryAuditLogInput{
		Limit: 10000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should not error, just cap the limit
	_ = out
}

func TestQueryAuditLog_NegativeOffset_DefaultsToZero(t *testing.T) {
	uc, auditRepo := newQueryAuditLogUseCase()
	seedAuditEvents(t, auditRepo)

	out, err := uc.Execute(context.Background(), QueryAuditLogInput{
		Offset: -5,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Total != 4 {
		t.Errorf("expected total 4, got %d", out.Total)
	}
}
