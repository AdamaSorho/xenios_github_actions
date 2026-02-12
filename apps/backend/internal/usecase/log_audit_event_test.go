package usecase

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/adapter/repository"
	"github.com/xenios/backend/internal/domain/entities"
)

func newLogAuditEventUseCase() (*LogAuditEventUseCase, *repository.InMemoryAuditRepository) {
	auditRepo := repository.NewInMemoryAuditRepository()
	uc := NewLogAuditEventUseCase(auditRepo)
	return uc, auditRepo
}

func TestLogAuditEvent_ValidEvent_LogsSuccessfully(t *testing.T) {
	uc, auditRepo := newLogAuditEventUseCase()

	err := uc.Execute(context.Background(), LogAuditEventInput{
		ActorID:    "user-1",
		Action:     "user.login",
		EntityType: "user",
		EntityID:   "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(auditRepo.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(auditRepo.Events))
	}
	if auditRepo.Events[0].Action != "user.login" {
		t.Errorf("expected action 'user.login', got '%s'", auditRepo.Events[0].Action)
	}
	if auditRepo.Events[0].ActorID != "user-1" {
		t.Errorf("expected actor_id 'user-1', got '%s'", auditRepo.Events[0].ActorID)
	}
}

func TestLogAuditEvent_WithMetadata_IncludesMetadata(t *testing.T) {
	uc, auditRepo := newLogAuditEventUseCase()

	err := uc.Execute(context.Background(), LogAuditEventInput{
		ActorID:    "user-1",
		Action:     "auth.login_failed",
		EntityType: "user",
		EntityID:   "user-1",
		Metadata:   map[string]interface{}{"reason": "invalid_password"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if auditRepo.Events[0].Metadata == nil {
		t.Fatal("expected metadata")
	}
	if auditRepo.Events[0].Metadata["reason"] != "invalid_password" {
		t.Errorf("expected reason 'invalid_password', got '%v'", auditRepo.Events[0].Metadata["reason"])
	}
}

func TestLogAuditEvent_WithIPAndUserAgent_Included(t *testing.T) {
	uc, auditRepo := newLogAuditEventUseCase()

	err := uc.Execute(context.Background(), LogAuditEventInput{
		ActorID:    "user-1",
		Action:     "client.view",
		EntityType: "client_profile",
		EntityID:   "client-1",
		IPAddress:  "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if auditRepo.Events[0].IPAddress != "192.168.1.1" {
		t.Errorf("expected IP '192.168.1.1', got '%s'", auditRepo.Events[0].IPAddress)
	}
	if auditRepo.Events[0].UserAgent != "Mozilla/5.0" {
		t.Errorf("expected User-Agent 'Mozilla/5.0', got '%s'", auditRepo.Events[0].UserAgent)
	}
}

func TestLogAuditEvent_EmptyActorID_ReturnsValidationError(t *testing.T) {
	uc, _ := newLogAuditEventUseCase()

	err := uc.Execute(context.Background(), LogAuditEventInput{
		ActorID:    "",
		Action:     "user.login",
		EntityType: "user",
		EntityID:   "user-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestLogAuditEvent_EmptyAction_ReturnsValidationError(t *testing.T) {
	uc, _ := newLogAuditEventUseCase()

	err := uc.Execute(context.Background(), LogAuditEventInput{
		ActorID:    "user-1",
		Action:     "",
		EntityType: "user",
		EntityID:   "user-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestLogAuditEvent_EmptyEntityType_ReturnsValidationError(t *testing.T) {
	uc, _ := newLogAuditEventUseCase()

	err := uc.Execute(context.Background(), LogAuditEventInput{
		ActorID:    "user-1",
		Action:     "user.login",
		EntityType: "",
		EntityID:   "user-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestLogAuditEvent_EmptyEntityID_ReturnsValidationError(t *testing.T) {
	uc, _ := newLogAuditEventUseCase()

	err := uc.Execute(context.Background(), LogAuditEventInput{
		ActorID:    "user-1",
		Action:     "user.login",
		EntityType: "user",
		EntityID:   "",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T", err)
	}
}

func TestLogAuditEvent_CreatesAuditEventEntity(t *testing.T) {
	uc, auditRepo := newLogAuditEventUseCase()

	err := uc.Execute(context.Background(), LogAuditEventInput{
		ActorID:    "user-1",
		Action:     "insight.approve",
		EntityType: "insight_card",
		EntityID:   "card-1",
		IPAddress:  "10.0.0.1",
		UserAgent:  "TestAgent",
		Metadata:   map[string]interface{}{"key": "value"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	event := auditRepo.Events[0]
	if event.EntityType != "insight_card" {
		t.Errorf("expected entity_type 'insight_card', got '%s'", event.EntityType)
	}
	if event.EntityID != "card-1" {
		t.Errorf("expected entity_id 'card-1', got '%s'", event.EntityID)
	}

	// Verify it created a proper AuditEvent
	var _ *entities.AuditEvent = event
}
