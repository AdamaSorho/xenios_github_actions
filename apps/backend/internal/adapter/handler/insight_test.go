package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/xenios/backend/internal/adapter/middleware"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

// --- Mock use cases for insight handler ---

type mockGetInsightQueueUC struct {
	output *usecase.GetInsightQueueOutput
	err    error
}

func (m *mockGetInsightQueueUC) Execute(_ context.Context, _ usecase.GetInsightQueueInput) (*usecase.GetInsightQueueOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.output, nil
}

type mockApproveInsightUC struct {
	output *entities.InsightCard
	err    error
}

func (m *mockApproveInsightUC) Execute(_ context.Context, _, _ string) (*entities.InsightCard, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.output, nil
}

type mockDismissInsightUC struct {
	output *entities.InsightCard
	err    error
}

func (m *mockDismissInsightUC) Execute(_ context.Context, _, _ string) (*entities.InsightCard, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.output, nil
}

type mockEditInsightUC struct {
	output *entities.InsightCard
	err    error
}

func (m *mockEditInsightUC) Execute(_ context.Context, _ usecase.EditInsightInput) (*entities.InsightCard, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.output, nil
}

type mockShareInsightUC struct {
	output *entities.InsightCard
	err    error
}

func (m *mockShareInsightUC) Execute(_ context.Context, _, _ string) (*entities.InsightCard, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.output, nil
}

type mockGetClientInsightsUC struct {
	output *usecase.GetClientInsightsOutput
	err    error
}

func (m *mockGetClientInsightsUC) Execute(_ context.Context, _ usecase.GetClientInsightsInput) (*usecase.GetClientInsightsOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.output, nil
}

func newInsightHandler(
	getQueue *mockGetInsightQueueUC,
	approve *mockApproveInsightUC,
	dismiss *mockDismissInsightUC,
	edit *mockEditInsightUC,
	share *mockShareInsightUC,
	getClient *mockGetClientInsightsUC,
) *InsightHandler {
	return NewInsightHandler(getQueue, approve, dismiss, edit, share, getClient)
}

func insightWithAuth(req *http.Request, userID, role string) *http.Request {
	ctx := middleware.SetUserClaims(req.Context(), &middleware.UserClaims{Subject: userID, Role: role})
	return req.WithContext(ctx)
}

func insightWithChiParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func insightWithAuthAndParam(req *http.Request, userID, role, paramKey, paramValue string) *http.Request {
	req = insightWithAuth(req, userID, role)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(paramKey, paramValue)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// --- GetQueue tests ---

func TestInsightHandler_GetQueue_Success(t *testing.T) {
	h := newInsightHandler(
		&mockGetInsightQueueUC{output: &usecase.GetInsightQueueOutput{
			Insights: []*entities.InsightCard{{ID: "i-1", Status: "draft"}},
			Total:    1, Page: 1, Limit: 20,
		}},
		nil, nil, nil, nil, nil,
	)

	req := httptest.NewRequest("GET", "/api/v1/insights/queue?status=draft", nil)
	req = insightWithAuth(req, "coach-1", "coach")
	rec := httptest.NewRecorder()

	h.GetQueue(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result usecase.GetInsightQueueOutput
	_ = json.NewDecoder(rec.Body).Decode(&result)
	if len(result.Insights) != 1 {
		t.Errorf("expected 1 insight, got %d", len(result.Insights))
	}
}

func TestInsightHandler_GetQueue_Unauthenticated(t *testing.T) {
	h := newInsightHandler(&mockGetInsightQueueUC{}, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/api/v1/insights/queue", nil)
	rec := httptest.NewRecorder()

	h.GetQueue(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestInsightHandler_GetQueue_InternalError(t *testing.T) {
	h := newInsightHandler(
		&mockGetInsightQueueUC{err: errors.New("db error")},
		nil, nil, nil, nil, nil,
	)

	req := httptest.NewRequest("GET", "/api/v1/insights/queue", nil)
	req = insightWithAuth(req, "coach-1", "coach")
	rec := httptest.NewRecorder()

	h.GetQueue(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

// --- Approve tests ---

func TestInsightHandler_Approve_Success(t *testing.T) {
	h := newInsightHandler(
		nil,
		&mockApproveInsightUC{output: &entities.InsightCard{ID: "i-1", Status: "approved"}},
		nil, nil, nil, nil,
	)

	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1/approve", nil)
	req = insightWithAuthAndParam(req, "coach-1", "coach", "insightID", "i-1")
	rec := httptest.NewRecorder()

	h.Approve(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestInsightHandler_Approve_Unauthenticated(t *testing.T) {
	h := newInsightHandler(nil, &mockApproveInsightUC{}, nil, nil, nil, nil)

	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1/approve", nil)
	req = insightWithChiParam(req, "insightID", "i-1")
	rec := httptest.NewRecorder()

	h.Approve(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestInsightHandler_Approve_MissingInsightID(t *testing.T) {
	h := newInsightHandler(nil, &mockApproveInsightUC{}, nil, nil, nil, nil)

	req := httptest.NewRequest("PUT", "/api/v1/insights//approve", nil)
	req = insightWithAuthAndParam(req, "coach-1", "coach", "insightID", "")
	rec := httptest.NewRecorder()

	h.Approve(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestInsightHandler_Approve_Forbidden(t *testing.T) {
	h := newInsightHandler(
		nil,
		&mockApproveInsightUC{err: &usecase.AuthorizationError{Message: "not your insight"}},
		nil, nil, nil, nil,
	)

	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1/approve", nil)
	req = insightWithAuthAndParam(req, "coach-2", "coach", "insightID", "i-1")
	rec := httptest.NewRecorder()

	h.Approve(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestInsightHandler_Approve_InvalidTransition(t *testing.T) {
	h := newInsightHandler(
		nil,
		&mockApproveInsightUC{err: &usecase.InvalidTransitionError{Message: "already dismissed"}},
		nil, nil, nil, nil,
	)

	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1/approve", nil)
	req = insightWithAuthAndParam(req, "coach-1", "coach", "insightID", "i-1")
	rec := httptest.NewRecorder()

	h.Approve(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rec.Code)
	}
}

// --- Dismiss tests ---

func TestInsightHandler_Dismiss_Success(t *testing.T) {
	h := newInsightHandler(
		nil, nil,
		&mockDismissInsightUC{output: &entities.InsightCard{ID: "i-1", Status: "dismissed"}},
		nil, nil, nil,
	)

	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1/dismiss", nil)
	req = insightWithAuthAndParam(req, "coach-1", "coach", "insightID", "i-1")
	rec := httptest.NewRecorder()

	h.Dismiss(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// --- Edit tests ---

func TestInsightHandler_Edit_Success(t *testing.T) {
	h := newInsightHandler(
		nil, nil, nil,
		&mockEditInsightUC{output: &entities.InsightCard{ID: "i-1", Title: "Updated", Body: "Updated body"}},
		nil, nil,
	)

	body := `{"title": "Updated", "body": "Updated body"}`
	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = insightWithAuthAndParam(req, "coach-1", "coach", "insightID", "i-1")
	rec := httptest.NewRecorder()

	h.Edit(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result entities.InsightCard
	_ = json.NewDecoder(rec.Body).Decode(&result)
	if result.Title != "Updated" {
		t.Errorf("expected 'Updated', got %s", result.Title)
	}
}

func TestInsightHandler_Edit_InvalidJSON(t *testing.T) {
	h := newInsightHandler(nil, nil, nil, &mockEditInsightUC{}, nil, nil)

	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1", bytes.NewBufferString("not json"))
	req = insightWithAuthAndParam(req, "coach-1", "coach", "insightID", "i-1")
	rec := httptest.NewRecorder()

	h.Edit(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// --- Share tests ---

func TestInsightHandler_Share_Success(t *testing.T) {
	h := newInsightHandler(
		nil, nil, nil, nil,
		&mockShareInsightUC{output: &entities.InsightCard{ID: "i-1", Status: "shared"}},
		nil,
	)

	req := httptest.NewRequest("PUT", "/api/v1/insights/i-1/share", nil)
	req = insightWithAuthAndParam(req, "coach-1", "coach", "insightID", "i-1")
	rec := httptest.NewRecorder()

	h.Share(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// --- GetClientInsights tests ---

func TestInsightHandler_GetClientInsights_Success(t *testing.T) {
	h := newInsightHandler(
		nil, nil, nil, nil, nil,
		&mockGetClientInsightsUC{output: &usecase.GetClientInsightsOutput{
			Insights: []*entities.InsightCard{{ID: "i-1"}},
			Total:    1, Page: 1, Limit: 20,
		}},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/insights", nil)
	req = insightWithAuthAndParam(req, "coach-1", "coach", "clientID", "client-1")
	rec := httptest.NewRecorder()

	h.GetClientInsights(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestInsightHandler_GetClientInsights_MissingClientID(t *testing.T) {
	h := newInsightHandler(nil, nil, nil, nil, nil, &mockGetClientInsightsUC{})

	req := httptest.NewRequest("GET", "/api/v1/clients//insights", nil)
	req = insightWithAuthAndParam(req, "coach-1", "coach", "clientID", "")
	rec := httptest.NewRecorder()

	h.GetClientInsights(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
