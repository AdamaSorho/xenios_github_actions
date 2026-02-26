package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/xenios/backend/internal/adapter/middleware"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

// --- Mock use cases for insight handler ---

type mockInsightQueueUC struct {
	output *usecase.GetInsightQueueOutput
	err    error
}

func (m *mockInsightQueueUC) Execute(_ context.Context, _ usecase.GetInsightQueueInput) (*usecase.GetInsightQueueOutput, error) {
	return m.output, m.err
}

type mockClientInsightsUC struct {
	output *usecase.GetInsightQueueOutput
	err    error
}

func (m *mockClientInsightsUC) Execute(_ context.Context, _ usecase.GetClientInsightsInput) (*usecase.GetInsightQueueOutput, error) {
	return m.output, m.err
}

type mockApproveInsightUC struct {
	output *entities.InsightCard
	err    error
}

func (m *mockApproveInsightUC) Execute(_ context.Context, _ usecase.InsightActionInput) (*entities.InsightCard, error) {
	return m.output, m.err
}

type mockDismissInsightUC struct {
	output *entities.InsightCard
	err    error
}

func (m *mockDismissInsightUC) Execute(_ context.Context, _ usecase.InsightActionInput) (*entities.InsightCard, error) {
	return m.output, m.err
}

type mockEditInsightUC struct {
	output *entities.InsightCard
	err    error
}

func (m *mockEditInsightUC) Execute(_ context.Context, _ usecase.EditInsightInput) (*entities.InsightCard, error) {
	return m.output, m.err
}

type mockShareInsightUC struct {
	output *entities.InsightCard
	err    error
}

func (m *mockShareInsightUC) Execute(_ context.Context, _ usecase.InsightActionInput) (*entities.InsightCard, error) {
	return m.output, m.err
}

func defaultInsightHandler() *InsightHandler {
	return NewInsightHandler(
		&mockInsightQueueUC{
			output: &usecase.GetInsightQueueOutput{
				Insights:   []*entities.InsightCard{},
				Pagination: usecase.PaginationInfo{Page: 1, Limit: 20, Total: 0},
			},
		},
		&mockClientInsightsUC{
			output: &usecase.GetInsightQueueOutput{
				Insights:   []*entities.InsightCard{},
				Pagination: usecase.PaginationInfo{Page: 1, Limit: 20, Total: 0},
			},
		},
		&mockApproveInsightUC{
			output: &entities.InsightCard{ID: "i-1", Status: entities.InsightStatusApproved},
		},
		&mockDismissInsightUC{
			output: &entities.InsightCard{ID: "i-1", Status: entities.InsightStatusDismissed},
		},
		&mockEditInsightUC{
			output: &entities.InsightCard{ID: "i-1", Title: "Updated"},
		},
		&mockShareInsightUC{
			output: &entities.InsightCard{ID: "i-1", Status: entities.InsightStatusShared},
		},
	)
}

func withCoachAuth(r *http.Request) *http.Request {
	ctx := middleware.SetUserClaims(r.Context(), &middleware.UserClaims{
		Subject: "coach-1",
		Role:    "coach",
	})
	return r.WithContext(ctx)
}

// --- GetQueue tests ---

func TestInsightHandler_GetQueue_Success(t *testing.T) {
	h := defaultInsightHandler()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/insights/queue", nil)
	req = withCoachAuth(req)

	h.GetQueue(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestInsightHandler_GetQueue_NoAuth_Returns401(t *testing.T) {
	h := defaultInsightHandler()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/insights/queue", nil)

	h.GetQueue(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestInsightHandler_GetQueue_ValidationError_Returns422(t *testing.T) {
	h := NewInsightHandler(
		&mockInsightQueueUC{err: &usecase.ValidationError{Message: "invalid status"}},
		&mockClientInsightsUC{},
		&mockApproveInsightUC{},
		&mockDismissInsightUC{},
		&mockEditInsightUC{},
		&mockShareInsightUC{},
	)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/insights/queue?status=invalid", nil)
	req = withCoachAuth(req)

	h.GetQueue(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", w.Code)
	}
}

func TestInsightHandler_GetQueue_InternalError_Returns500(t *testing.T) {
	h := NewInsightHandler(
		&mockInsightQueueUC{err: fmt.Errorf("database error")},
		&mockClientInsightsUC{},
		&mockApproveInsightUC{},
		&mockDismissInsightUC{},
		&mockEditInsightUC{},
		&mockShareInsightUC{},
	)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/insights/queue", nil)
	req = withCoachAuth(req)

	h.GetQueue(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

// --- GetClientInsights tests ---

func TestInsightHandler_GetClientInsights_Success(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Get("/api/v1/clients/{clientID}/insights", h.GetClientInsights)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/insights", nil)
	req = withCoachAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestInsightHandler_GetClientInsights_NoAuth_Returns401(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Get("/api/v1/clients/{clientID}/insights", h.GetClientInsights)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/insights", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

// --- Approve tests ---

func TestInsightHandler_Approve_Success(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/approve", h.Approve)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1/approve", nil)
	req = withCoachAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestInsightHandler_Approve_NoAuth_Returns401(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/approve", h.Approve)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1/approve", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestInsightHandler_Approve_ValidationError_Returns422(t *testing.T) {
	h := NewInsightHandler(
		&mockInsightQueueUC{},
		&mockClientInsightsUC{},
		&mockApproveInsightUC{err: &usecase.ValidationError{Message: "cannot approve"}},
		&mockDismissInsightUC{},
		&mockEditInsightUC{},
		&mockShareInsightUC{},
	)

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/approve", h.Approve)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1/approve", nil)
	req = withCoachAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", w.Code)
	}
}

func TestInsightHandler_Approve_AuthError_Returns403(t *testing.T) {
	h := NewInsightHandler(
		&mockInsightQueueUC{},
		&mockClientInsightsUC{},
		&mockApproveInsightUC{err: &usecase.AuthenticationError{Message: "not authorized"}},
		&mockDismissInsightUC{},
		&mockEditInsightUC{},
		&mockShareInsightUC{},
	)

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/approve", h.Approve)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1/approve", nil)
	req = withCoachAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

// --- Dismiss tests ---

func TestInsightHandler_Dismiss_Success(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/dismiss", h.Dismiss)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1/dismiss", nil)
	req = withCoachAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestInsightHandler_Dismiss_NoAuth_Returns401(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/dismiss", h.Dismiss)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1/dismiss", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestInsightHandler_Dismiss_ValidationError_Returns422(t *testing.T) {
	h := NewInsightHandler(
		&mockInsightQueueUC{},
		&mockClientInsightsUC{},
		&mockApproveInsightUC{},
		&mockDismissInsightUC{err: &usecase.ValidationError{Message: "cannot dismiss"}},
		&mockEditInsightUC{},
		&mockShareInsightUC{},
	)

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/dismiss", h.Dismiss)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1/dismiss", nil)
	req = withCoachAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", w.Code)
	}
}

// --- Edit tests ---

func TestInsightHandler_Edit_Success(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}", h.Edit)

	body, _ := json.Marshal(EditInsightRequest{Title: "New Title", Body: "New body"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1", bytes.NewReader(body))
	req = withCoachAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestInsightHandler_Edit_NoAuth_Returns401(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}", h.Edit)

	body, _ := json.Marshal(EditInsightRequest{Title: "New Title"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1", bytes.NewReader(body))

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestInsightHandler_Edit_InvalidJSON_Returns400(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}", h.Edit)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1", bytes.NewReader([]byte("not json")))
	req = withCoachAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestInsightHandler_Edit_ValidationError_Returns422(t *testing.T) {
	h := NewInsightHandler(
		&mockInsightQueueUC{},
		&mockClientInsightsUC{},
		&mockApproveInsightUC{},
		&mockDismissInsightUC{},
		&mockEditInsightUC{err: &usecase.ValidationError{Message: "title or body is required"}},
		&mockShareInsightUC{},
	)

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}", h.Edit)

	body, _ := json.Marshal(EditInsightRequest{})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1", bytes.NewReader(body))
	req = withCoachAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", w.Code)
	}
}

func TestInsightHandler_Edit_AuthError_Returns403(t *testing.T) {
	h := NewInsightHandler(
		&mockInsightQueueUC{},
		&mockClientInsightsUC{},
		&mockApproveInsightUC{},
		&mockDismissInsightUC{},
		&mockEditInsightUC{err: &usecase.AuthenticationError{Message: "not authorized"}},
		&mockShareInsightUC{},
	)

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}", h.Edit)

	body, _ := json.Marshal(EditInsightRequest{Title: "X"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1", bytes.NewReader(body))
	req = withCoachAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

// --- Share tests ---

func TestInsightHandler_Share_Success(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/share", h.Share)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1/share", nil)
	req = withCoachAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestInsightHandler_Share_NoAuth_Returns401(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/share", h.Share)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1/share", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestInsightHandler_Share_ValidationError_Returns422(t *testing.T) {
	h := NewInsightHandler(
		&mockInsightQueueUC{},
		&mockClientInsightsUC{},
		&mockApproveInsightUC{},
		&mockDismissInsightUC{},
		&mockEditInsightUC{},
		&mockShareInsightUC{err: &usecase.ValidationError{Message: "cannot share draft"}},
	)

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/share", h.Share)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1/share", nil)
	req = withCoachAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", w.Code)
	}
}

func TestInsightHandler_Share_AuthError_Returns403(t *testing.T) {
	h := NewInsightHandler(
		&mockInsightQueueUC{},
		&mockClientInsightsUC{},
		&mockApproveInsightUC{},
		&mockDismissInsightUC{},
		&mockEditInsightUC{},
		&mockShareInsightUC{err: &usecase.AuthenticationError{Message: "not authorized"}},
	)

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/share", h.Share)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1/share", nil)
	req = withCoachAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}
