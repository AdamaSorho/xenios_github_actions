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

// mockInsightActionUC is a generic mock for single-insight action use cases.
type mockInsightActionUC[T any] struct {
	card *entities.InsightCard
	err  error
}

func (m *mockInsightActionUC[T]) Execute(_ context.Context, _ T) (*entities.InsightCard, error) {
	return m.card, m.err
}

// emptyListOutput returns a default empty list output for mock setup.
func emptyListOutput() *usecase.InsightListOutput {
	return &usecase.InsightListOutput{
		Insights: []*entities.InsightCard{},
		Total:    0,
		Limit:    20,
		Offset:   0,
	}
}

func defaultInsightHandler() *InsightHandler {
	return NewInsightHandler(
		&mockGetInsightQueueUC{output: emptyListOutput()},
		&mockGetClientInsightsUC{output: emptyListOutput()},
		&mockInsightActionUC[usecase.ApproveInsightInput]{
			card: &entities.InsightCard{ID: "i1", Status: entities.InsightStatusApproved},
		},
		&mockInsightActionUC[usecase.DismissInsightInput]{
			card: &entities.InsightCard{ID: "i1", Status: entities.InsightStatusDismissed},
		},
		&mockInsightActionUC[usecase.EditInsightInput]{
			card: &entities.InsightCard{ID: "i1", Title: "Updated"},
		},
		&mockInsightActionUC[usecase.ShareInsightInput]{
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

// --- Action Handler Tests (table-driven for approve, dismiss, share) ---

func TestInsightHandler_Actions_Unauthenticated_Returns401(t *testing.T) {
	h := defaultInsightHandler()

	actions := []struct {
		name   string
		method string
		path   string
		handle http.HandlerFunc
	}{
		{"Approve", "PUT", "/api/v1/insights/i1/approve", h.Approve},
		{"Dismiss", "PUT", "/api/v1/insights/i1/dismiss", h.Dismiss},
		{"Share", "PUT", "/api/v1/insights/i1/share", h.Share},
	}

	for _, a := range actions {
		t.Run(a.name, func(t *testing.T) {
			req := httptest.NewRequest(a.method, a.path, nil)
			rec := httptest.NewRecorder()
			a.handle(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected 401, got %d", rec.Code)
			}
		})
	}
}

func TestInsightHandler_Actions_Success(t *testing.T) {
	h := defaultInsightHandler()

	actions := []struct {
		name    string
		path    string
		handler http.HandlerFunc
	}{
		{"Approve", "/api/v1/insights/{insightID}/approve", h.Approve},
		{"Dismiss", "/api/v1/insights/{insightID}/dismiss", h.Dismiss},
		{"Share", "/api/v1/insights/{insightID}/share", h.Share},
	}

	for _, a := range actions {
		t.Run(a.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Put(a.path, a.handler)

			req := httptest.NewRequest("PUT", "/api/v1/insights/i1"+a.path[len("/api/v1/insights/{insightID}"):], nil)
			req = req.WithContext(coachCtx(req.Context()))
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected 200, got %d", rec.Code)
			}
		})
	}
}

func TestInsightHandler_Approve_AuthorizationError_Returns403(t *testing.T) {
	h := NewInsightHandler(
		&mockGetInsightQueueUC{},
		&mockGetClientInsightsUC{},
		&mockInsightActionUC[usecase.ApproveInsightInput]{err: &usecase.AuthorizationError{Message: "forbidden"}},
		&mockInsightActionUC[usecase.DismissInsightInput]{},
		&mockInsightActionUC[usecase.EditInsightInput]{},
		&mockInsightActionUC[usecase.ShareInsightInput]{},
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
		&mockInsightActionUC[usecase.ApproveInsightInput]{err: &entities.StatusTransitionError{From: "dismissed", To: "approved"}},
		&mockInsightActionUC[usecase.DismissInsightInput]{},
		&mockInsightActionUC[usecase.EditInsightInput]{},
		&mockInsightActionUC[usecase.ShareInsightInput]{},
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

// --- handleInsightError Tests ---

func TestHandleInsightError_MapsErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{"ValidationError", &usecase.ValidationError{Message: "bad input"}, http.StatusBadRequest},
		{"AuthorizationError", &usecase.AuthorizationError{Message: "forbidden"}, http.StatusForbidden},
		{"StatusTransitionError", &entities.StatusTransitionError{From: "draft", To: "shared"}, http.StatusUnprocessableEntity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handleInsightError(rec, tt.err)
			if rec.Code != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, rec.Code)
			}
		})
	}
}
