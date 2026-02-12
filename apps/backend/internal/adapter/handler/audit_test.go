package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/xenios/backend/internal/adapter/middleware"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

// --- Mock use cases ---

type mockQueryAuditLogUC struct {
	output *usecase.QueryAuditLogOutput
	err    error
}

func (m *mockQueryAuditLogUC) Execute(_ context.Context, _ usecase.QueryAuditLogInput) (*usecase.QueryAuditLogOutput, error) {
	return m.output, m.err
}

func defaultAuditHandler() *AuditHandler {
	return NewAuditHandler(&mockQueryAuditLogUC{
		output: &usecase.QueryAuditLogOutput{
			Events: []*entities.AuditEvent{
				{
					ID:         "ev-1",
					ActorID:    "user-1",
					Action:     "user.login",
					EntityType: "user",
					EntityID:   "user-1",
					CreatedAt:  time.Now(),
				},
			},
			Total:  1,
			Limit:  50,
			Offset: 0,
		},
	})
}

// --- Query tests ---

func TestAuditHandler_Query_Success(t *testing.T) {
	h := defaultAuditHandler()

	req := httptest.NewRequest("GET", "/api/v1/admin/audit?limit=10", nil)
	ctx := middleware.SetUserClaims(req.Context(), &middleware.UserClaims{Subject: "admin-1", Role: "admin"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.Query(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	events, ok := resp["events"].([]interface{})
	if !ok {
		t.Fatal("expected events array in response")
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
}

func TestAuditHandler_Query_NonAdmin_Forbidden(t *testing.T) {
	h := defaultAuditHandler()

	req := httptest.NewRequest("GET", "/api/v1/admin/audit", nil)
	ctx := middleware.SetUserClaims(req.Context(), &middleware.UserClaims{Subject: "user-1", Role: "client"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.Query(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestAuditHandler_Query_CoachRole_Forbidden(t *testing.T) {
	h := defaultAuditHandler()

	req := httptest.NewRequest("GET", "/api/v1/admin/audit", nil)
	ctx := middleware.SetUserClaims(req.Context(), &middleware.UserClaims{Subject: "user-1", Role: "coach"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.Query(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestAuditHandler_Query_NoClaims_Unauthorized(t *testing.T) {
	h := defaultAuditHandler()

	req := httptest.NewRequest("GET", "/api/v1/admin/audit", nil)
	rec := httptest.NewRecorder()

	h.Query(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestAuditHandler_Query_WithFilters_PassedThrough(t *testing.T) {
	var capturedInput usecase.QueryAuditLogInput
	h := NewAuditHandler(&mockQueryAuditLogUCCapture{
		output: &usecase.QueryAuditLogOutput{
			Events: []*entities.AuditEvent{},
			Total:  0,
			Limit:  10,
			Offset: 0,
		},
		capturedInput: &capturedInput,
	})

	req := httptest.NewRequest("GET", "/api/v1/admin/audit?entity_type=user&action=user.login&actor_id=user-1&limit=10&offset=5", nil)
	ctx := middleware.SetUserClaims(req.Context(), &middleware.UserClaims{Subject: "admin-1", Role: "admin"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.Query(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if capturedInput.EntityType != "user" {
		t.Errorf("expected entity_type 'user', got '%s'", capturedInput.EntityType)
	}
	if capturedInput.Action != "user.login" {
		t.Errorf("expected action 'user.login', got '%s'", capturedInput.Action)
	}
	if capturedInput.ActorID != "user-1" {
		t.Errorf("expected actor_id 'user-1', got '%s'", capturedInput.ActorID)
	}
	if capturedInput.Limit != 10 {
		t.Errorf("expected limit 10, got %d", capturedInput.Limit)
	}
	if capturedInput.Offset != 5 {
		t.Errorf("expected offset 5, got %d", capturedInput.Offset)
	}
}

func TestAuditHandler_Query_WithTimeRange_ParsedCorrectly(t *testing.T) {
	var capturedInput usecase.QueryAuditLogInput
	h := NewAuditHandler(&mockQueryAuditLogUCCapture{
		output: &usecase.QueryAuditLogOutput{
			Events: []*entities.AuditEvent{},
			Total:  0,
			Limit:  50,
			Offset: 0,
		},
		capturedInput: &capturedInput,
	})

	req := httptest.NewRequest("GET", "/api/v1/admin/audit?from=2024-01-01T00:00:00Z&to=2024-12-31T23:59:59Z", nil)
	ctx := middleware.SetUserClaims(req.Context(), &middleware.UserClaims{Subject: "admin-1", Role: "admin"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.Query(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if capturedInput.From == nil {
		t.Fatal("expected 'from' to be parsed")
	}
	if capturedInput.To == nil {
		t.Fatal("expected 'to' to be parsed")
	}
}

func TestAuditHandler_Query_ContentType(t *testing.T) {
	h := defaultAuditHandler()

	req := httptest.NewRequest("GET", "/api/v1/admin/audit", nil)
	ctx := middleware.SetUserClaims(req.Context(), &middleware.UserClaims{Subject: "admin-1", Role: "admin"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.Query(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", ct)
	}
}

func TestAuditHandler_Query_InternalError_Returns500(t *testing.T) {
	h := NewAuditHandler(&mockQueryAuditLogUC{
		err: context.DeadlineExceeded,
	})

	req := httptest.NewRequest("GET", "/api/v1/admin/audit", nil)
	ctx := middleware.SetUserClaims(req.Context(), &middleware.UserClaims{Subject: "admin-1", Role: "admin"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.Query(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

// --- Helper mock that captures input ---

type mockQueryAuditLogUCCapture struct {
	output        *usecase.QueryAuditLogOutput
	capturedInput *usecase.QueryAuditLogInput
}

func (m *mockQueryAuditLogUCCapture) Execute(_ context.Context, input usecase.QueryAuditLogInput) (*usecase.QueryAuditLogOutput, error) {
	*m.capturedInput = input
	return m.output, nil
}
