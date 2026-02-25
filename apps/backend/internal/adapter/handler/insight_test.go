package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/xenios/backend/internal/adapter/middleware"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

// --- Mock use cases ---

type mockGetInsightQueueUC struct {
	output *usecase.GetInsightQueueOutput
	err    error
}

func (m *mockGetInsightQueueUC) Execute(_ context.Context, _ usecase.GetInsightQueueInput) (*usecase.GetInsightQueueOutput, error) {
	return m.output, m.err
}

type mockGetClientInsightsUC struct {
	output *usecase.GetClientInsightsOutput
	err    error
}

func (m *mockGetClientInsightsUC) Execute(_ context.Context, _ usecase.GetClientInsightsInput) (*usecase.GetClientInsightsOutput, error) {
	return m.output, m.err
}

type mockApproveInsightUC struct {
	card *entities.InsightCard
	err  error
}

func (m *mockApproveInsightUC) Execute(_ context.Context, _ usecase.ApproveInsightInput) (*entities.InsightCard, error) {
	return m.card, m.err
}

type mockDismissInsightUC struct {
	card *entities.InsightCard
	err  error
}

func (m *mockDismissInsightUC) Execute(_ context.Context, _ usecase.DismissInsightInput) (*entities.InsightCard, error) {
	return m.card, m.err
}

type mockEditInsightUC struct {
	card *entities.InsightCard
	err  error
}

func (m *mockEditInsightUC) Execute(_ context.Context, _ usecase.EditInsightInput) (*entities.InsightCard, error) {
	return m.card, m.err
}

type mockShareInsightUC struct {
	card *entities.InsightCard
	err  error
}

func (m *mockShareInsightUC) Execute(_ context.Context, _ usecase.ShareInsightInput) (*entities.InsightCard, error) {
	return m.card, m.err
}

func defaultInsightHandler() *InsightHandler {
	return NewInsightHandler(
		&mockGetInsightQueueUC{
			output: &usecase.GetInsightQueueOutput{
				Insights: []*entities.InsightCard{},
				Total:    0,
				Limit:    20,
				Offset:   0,
			},
		},
		&mockGetClientInsightsUC{
			output: &usecase.GetClientInsightsOutput{
				Insights: []*entities.InsightCard{},
				Total:    0,
				Limit:    20,
				Offset:   0,
			},
		},
		&mockApproveInsightUC{
			card: &entities.InsightCard{ID: "i1", Status: entities.InsightStatusApproved},
		},
		&mockDismissInsightUC{
			card: &entities.InsightCard{ID: "i1", Status: entities.InsightStatusDismissed},
		},
		&mockEditInsightUC{
			card: &entities.InsightCard{ID: "i1", Title: "Updated"},
		},
		&mockShareInsightUC{
			card: &entities.InsightCard{ID: "i1", Status: entities.InsightStatusShared},
		},
	)
}

func coachCtx(ctx context.Context) context.Context {
	return middleware.SetUserClaims(ctx, &middleware.UserClaims{
		Subject: "coach-1",
		Role:    "coach",
	})
}

func clientCtx(ctx context.Context) context.Context {
	return middleware.SetUserClaims(ctx, &middleware.UserClaims{
		Subject: "client-1",
		Role:    "client",
	})
}

// --- GetQueue Tests ---

func TestInsightHandler_GetQueue_Success(t *testing.T) {
	h := defaultInsightHandler()
	req := httptest.NewRequest("GET", "/api/v1/insights/queue", nil)
	req = req.WithContext(coachCtx(req.Context()))
	rec := httptest.NewRecorder()

	h.GetQueue(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := resp["insights"]; !ok {
		t.Error("expected 'insights' field in response")
	}
	if _, ok := resp["pagination"]; !ok {
		t.Error("expected 'pagination' field in response")
	}
}

func TestInsightHandler_GetQueue_Unauthenticated_Returns401(t *testing.T) {
	h := defaultInsightHandler()
	req := httptest.NewRequest("GET", "/api/v1/insights/queue", nil)
	rec := httptest.NewRecorder()

	h.GetQueue(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestInsightHandler_GetQueue_ClientRole_Returns403(t *testing.T) {
	h := defaultInsightHandler()
	req := httptest.NewRequest("GET", "/api/v1/insights/queue", nil)
	req = req.WithContext(clientCtx(req.Context()))
	rec := httptest.NewRecorder()

	h.GetQueue(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestInsightHandler_GetQueue_WithParams(t *testing.T) {
	h := defaultInsightHandler()
	req := httptest.NewRequest("GET", "/api/v1/insights/queue?status=draft&limit=10&offset=5", nil)
	req = req.WithContext(coachCtx(req.Context()))
	rec := httptest.NewRecorder()

	h.GetQueue(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

// --- GetClientInsights Tests ---

func TestInsightHandler_GetClientInsights_Success(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Get("/api/v1/clients/{clientID}/insights", h.GetClientInsights)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/insights", nil)
	req = req.WithContext(coachCtx(req.Context()))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestInsightHandler_GetClientInsights_Unauthenticated(t *testing.T) {
	h := defaultInsightHandler()
	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/insights", nil)
	rec := httptest.NewRecorder()

	h.GetClientInsights(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

// --- Approve Tests ---

func TestInsightHandler_Approve_Success(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/approve", h.Approve)

	req := httptest.NewRequest("PUT", "/api/v1/insights/i1/approve", nil)
	req = req.WithContext(coachCtx(req.Context()))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestInsightHandler_Approve_Unauthenticated(t *testing.T) {
	h := defaultInsightHandler()
	req := httptest.NewRequest("PUT", "/api/v1/insights/i1/approve", nil)
	rec := httptest.NewRecorder()

	h.Approve(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestInsightHandler_Approve_AuthorizationError_Returns403(t *testing.T) {
	h := NewInsightHandler(
		&mockGetInsightQueueUC{},
		&mockGetClientInsightsUC{},
		&mockApproveInsightUC{err: &usecase.AuthorizationError{Message: "forbidden"}},
		&mockDismissInsightUC{},
		&mockEditInsightUC{},
		&mockShareInsightUC{},
	)

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/approve", h.Approve)

	req := httptest.NewRequest("PUT", "/api/v1/insights/i1/approve", nil)
	req = req.WithContext(coachCtx(req.Context()))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestInsightHandler_Approve_InvalidTransition_Returns422(t *testing.T) {
	h := NewInsightHandler(
		&mockGetInsightQueueUC{},
		&mockGetClientInsightsUC{},
		&mockApproveInsightUC{err: &entities.StatusTransitionError{From: "dismissed", To: "approved"}},
		&mockDismissInsightUC{},
		&mockEditInsightUC{},
		&mockShareInsightUC{},
	)

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/approve", h.Approve)

	req := httptest.NewRequest("PUT", "/api/v1/insights/i1/approve", nil)
	req = req.WithContext(coachCtx(req.Context()))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", rec.Code)
	}
}

// --- Dismiss Tests ---

func TestInsightHandler_Dismiss_Success(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/dismiss", h.Dismiss)

	req := httptest.NewRequest("PUT", "/api/v1/insights/i1/dismiss", nil)
	req = req.WithContext(coachCtx(req.Context()))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

// --- Edit Tests ---

func TestInsightHandler_Edit_Success(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}", h.Edit)

	body, _ := json.Marshal(EditInsightRequest{Title: "New Title", Body: "New Body"})
	req := httptest.NewRequest("PUT", "/api/v1/insights/i1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(coachCtx(req.Context()))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestInsightHandler_Edit_InvalidJSON_Returns400(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}", h.Edit)

	req := httptest.NewRequest("PUT", "/api/v1/insights/i1", bytes.NewReader([]byte("not json")))
	req = req.WithContext(coachCtx(req.Context()))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

// --- Share Tests ---

func TestInsightHandler_Share_Success(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/share", h.Share)

	req := httptest.NewRequest("PUT", "/api/v1/insights/i1/share", nil)
	req = req.WithContext(coachCtx(req.Context()))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestInsightHandler_Share_Unauthenticated(t *testing.T) {
	h := defaultInsightHandler()
	req := httptest.NewRequest("PUT", "/api/v1/insights/i1/share", nil)
	rec := httptest.NewRecorder()

	h.Share(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

// --- handleInsightError Tests ---

func TestHandleInsightError_ValidationError(t *testing.T) {
	rec := httptest.NewRecorder()
	handleInsightError(rec, &usecase.ValidationError{Message: "bad input"})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleInsightError_AuthorizationError(t *testing.T) {
	rec := httptest.NewRecorder()
	handleInsightError(rec, &usecase.AuthorizationError{Message: "forbidden"})
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestHandleInsightError_StatusTransitionError(t *testing.T) {
	rec := httptest.NewRecorder()
	handleInsightError(rec, &entities.StatusTransitionError{From: "draft", To: "shared"})
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rec.Code)
	}
}
