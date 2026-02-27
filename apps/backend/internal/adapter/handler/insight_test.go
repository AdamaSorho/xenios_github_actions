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
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

type mockListInsightsUseCase struct {
	executeFunc func(ctx context.Context, coachID string, status entities.InsightStatus, limit, offset int) ([]*entities.InsightCard, error)
}

func (m *mockListInsightsUseCase) Execute(ctx context.Context, coachID string, status entities.InsightStatus, limit, offset int) ([]*entities.InsightCard, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, coachID, status, limit, offset)
	}
	return nil, nil
}

type mockUpdateInsightStatusUseCase struct {
	executeFunc func(ctx context.Context, id string, status entities.InsightStatus) (*entities.InsightCard, error)
}

func (m *mockUpdateInsightStatusUseCase) Execute(ctx context.Context, id string, status entities.InsightStatus) (*entities.InsightCard, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, id, status)
	}
	return nil, nil
}

func newInsightRequest(method, path string, body string, coachID, insightID string) *http.Request {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	rctx := chi.NewRouteContext()
	if coachID != "" {
		rctx.URLParams.Add("coachID", coachID)
	}
	if insightID != "" {
		rctx.URLParams.Add("insightID", insightID)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestInsightHandler_ListDraftInsights_Success(t *testing.T) {
	listUC := &mockListInsightsUseCase{
		executeFunc: func(ctx context.Context, coachID string, status entities.InsightStatus, limit, offset int) ([]*entities.InsightCard, error) {
			return []*entities.InsightCard{
				{ID: "insight-1", CoachID: coachID, Title: "Test", Status: entities.InsightStatusDraft},
			}, nil
		},
	}
	updateUC := &mockUpdateInsightStatusUseCase{}

	h := NewInsightHandler(listUC, updateUC)
	req := newInsightRequest(http.MethodGet, "/coaches/coach-1/insights/drafts", "", "coach-1", "")
	w := httptest.NewRecorder()

	h.ListDraftInsights(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var cards []entities.InsightCard
	if err := json.NewDecoder(w.Body).Decode(&cards); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(cards) != 1 {
		t.Errorf("expected 1 card, got %d", len(cards))
	}
}

func TestInsightHandler_ListDraftInsights_NoCards_ReturnsEmptyArray(t *testing.T) {
	listUC := &mockListInsightsUseCase{
		executeFunc: func(ctx context.Context, coachID string, status entities.InsightStatus, limit, offset int) ([]*entities.InsightCard, error) {
			return nil, nil
		},
	}
	updateUC := &mockUpdateInsightStatusUseCase{}

	h := NewInsightHandler(listUC, updateUC)
	req := newInsightRequest(http.MethodGet, "/coaches/coach-1/insights/drafts", "", "coach-1", "")
	w := httptest.NewRecorder()

	h.ListDraftInsights(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var cards []entities.InsightCard
	if err := json.NewDecoder(w.Body).Decode(&cards); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(cards) != 0 {
		t.Errorf("expected 0 cards, got %d", len(cards))
	}
}

func TestInsightHandler_ListDraftInsights_MissingCoachID_Returns400(t *testing.T) {
	listUC := &mockListInsightsUseCase{}
	updateUC := &mockUpdateInsightStatusUseCase{}

	h := NewInsightHandler(listUC, updateUC)
	req := newInsightRequest(http.MethodGet, "/coaches//insights/drafts", "", "", "")
	w := httptest.NewRecorder()

	h.ListDraftInsights(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestInsightHandler_ListDraftInsights_Error_Returns500(t *testing.T) {
	listUC := &mockListInsightsUseCase{
		executeFunc: func(ctx context.Context, coachID string, status entities.InsightStatus, limit, offset int) ([]*entities.InsightCard, error) {
			return nil, errors.New("database error")
		},
	}
	updateUC := &mockUpdateInsightStatusUseCase{}

	h := NewInsightHandler(listUC, updateUC)
	req := newInsightRequest(http.MethodGet, "/coaches/coach-1/insights/drafts", "", "coach-1", "")
	w := httptest.NewRecorder()

	h.ListDraftInsights(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestInsightHandler_UpdateInsightStatus_Success(t *testing.T) {
	updateUC := &mockUpdateInsightStatusUseCase{
		executeFunc: func(ctx context.Context, id string, status entities.InsightStatus) (*entities.InsightCard, error) {
			return &entities.InsightCard{ID: id, Status: status, CoachID: "coach-1"}, nil
		},
	}
	listUC := &mockListInsightsUseCase{}

	h := NewInsightHandler(listUC, updateUC)
	body := `{"status":"approved"}`
	req := newInsightRequest(http.MethodPut, "/insights/insight-1/status", body, "", "insight-1")
	w := httptest.NewRecorder()

	h.UpdateInsightStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var card entities.InsightCard
	if err := json.NewDecoder(w.Body).Decode(&card); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if card.Status != entities.InsightStatusApproved {
		t.Errorf("expected status approved, got %s", card.Status)
	}
}

func TestInsightHandler_UpdateInsightStatus_MissingInsightID_Returns400(t *testing.T) {
	updateUC := &mockUpdateInsightStatusUseCase{}
	listUC := &mockListInsightsUseCase{}

	h := NewInsightHandler(listUC, updateUC)
	body := `{"status":"approved"}`
	req := newInsightRequest(http.MethodPut, "/insights//status", body, "", "")
	w := httptest.NewRecorder()

	h.UpdateInsightStatus(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestInsightHandler_UpdateInsightStatus_InvalidJSON_Returns400(t *testing.T) {
	updateUC := &mockUpdateInsightStatusUseCase{}
	listUC := &mockListInsightsUseCase{}

	h := NewInsightHandler(listUC, updateUC)
	req := newInsightRequest(http.MethodPut, "/insights/insight-1/status", "not json", "", "insight-1")
	w := httptest.NewRecorder()

	h.UpdateInsightStatus(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestInsightHandler_UpdateInsightStatus_MissingStatus_Returns400(t *testing.T) {
	updateUC := &mockUpdateInsightStatusUseCase{}
	listUC := &mockListInsightsUseCase{}

	h := NewInsightHandler(listUC, updateUC)
	body := `{}`
	req := newInsightRequest(http.MethodPut, "/insights/insight-1/status", body, "", "insight-1")
	w := httptest.NewRecorder()

	h.UpdateInsightStatus(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestInsightHandler_UpdateInsightStatus_InvalidStatus_Returns400(t *testing.T) {
	updateUC := &mockUpdateInsightStatusUseCase{}
	listUC := &mockListInsightsUseCase{}

	h := NewInsightHandler(listUC, updateUC)
	body := `{"status":"invalid"}`
	req := newInsightRequest(http.MethodPut, "/insights/insight-1/status", body, "", "insight-1")
	w := httptest.NewRecorder()

	h.UpdateInsightStatus(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestInsightHandler_UpdateInsightStatus_UseCaseError_Returns500(t *testing.T) {
	updateUC := &mockUpdateInsightStatusUseCase{
		executeFunc: func(ctx context.Context, id string, status entities.InsightStatus) (*entities.InsightCard, error) {
			return nil, errors.New("database error")
		},
	}
	listUC := &mockListInsightsUseCase{}

	h := NewInsightHandler(listUC, updateUC)
	body := `{"status":"approved"}`
	req := newInsightRequest(http.MethodPut, "/insights/insight-1/status", body, "", "insight-1")
	w := httptest.NewRecorder()

	h.UpdateInsightStatus(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestInsightHandler_UpdateInsightStatus_ValidationError_Returns400(t *testing.T) {
	updateUC := &mockUpdateInsightStatusUseCase{
		executeFunc: func(ctx context.Context, id string, status entities.InsightStatus) (*entities.InsightCard, error) {
			return nil, &usecase.ValidationError{Message: "insight not found"}
		},
	}
	listUC := &mockListInsightsUseCase{}

	h := NewInsightHandler(listUC, updateUC)
	body := `{"status":"approved"}`
	req := newInsightRequest(http.MethodPut, "/insights/insight-1/status", body, "", "insight-1")
	w := httptest.NewRecorder()

	h.UpdateInsightStatus(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestNewInsightHandler_ReturnsNonNil(t *testing.T) {
	h := NewInsightHandler(&mockListInsightsUseCase{}, &mockUpdateInsightStatusUseCase{})
	if h == nil {
		t.Error("expected non-nil InsightHandler")
	}
}

func TestInsightHandler_ListDraftInsights_ContentType(t *testing.T) {
	listUC := &mockListInsightsUseCase{
		executeFunc: func(ctx context.Context, coachID string, status entities.InsightStatus, limit, offset int) ([]*entities.InsightCard, error) {
			return []*entities.InsightCard{}, nil
		},
	}
	updateUC := &mockUpdateInsightStatusUseCase{}

	h := NewInsightHandler(listUC, updateUC)
	req := newInsightRequest(http.MethodGet, "/coaches/coach-1/insights/drafts", "", "coach-1", "")
	w := httptest.NewRecorder()

	h.ListDraftInsights(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type %q, got %q", "application/json", ct)
	}
}
