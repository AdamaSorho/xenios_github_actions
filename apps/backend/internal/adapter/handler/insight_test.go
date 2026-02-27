package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

// --- Mock use cases for InsightHandler ---

type mockInsightTransitionUC struct {
	output *entities.InsightCard
	err    error
}

func (m *mockInsightTransitionUC) Execute(_ context.Context, _ usecase.TransitionInput) (*entities.InsightCard, error) {
	return m.output, m.err
}

type mockEditInsightUC struct {
	output *entities.InsightCard
	err    error
}

func (m *mockEditInsightUC) Execute(_ context.Context, _ usecase.EditInsightInput) (*entities.InsightCard, error) {
	return m.output, m.err
}

type mockGetInsightQueueUC struct {
	output *usecase.InsightQueueOutput
	err    error
}

func (m *mockGetInsightQueueUC) Execute(_ context.Context, _ usecase.InsightQueueInput) (*usecase.InsightQueueOutput, error) {
	return m.output, m.err
}

type mockGetClientInsightsUC struct {
	output *usecase.InsightQueueOutput
	err    error
}

func (m *mockGetClientInsightsUC) Execute(_ context.Context, _ usecase.ClientInsightsInput) (*usecase.InsightQueueOutput, error) {
	return m.output, m.err
}

func sampleInsightCard() *entities.InsightCard {
	now := time.Now()
	return &entities.InsightCard{
		ID:         "insight-1",
		ClientID:   "client-1",
		CoachID:    "coach-1",
		ClientName: "Test Client",
		Title:      "Test Insight",
		Body:       "Test body",
		Category:   "nutrition",
		Priority:   entities.InsightPriorityMedium,
		Status:     entities.InsightStatusDraft,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func defaultInsightHandler() *InsightHandler {
	return NewInsightHandler(
		&mockInsightTransitionUC{output: sampleInsightCard()},
		&mockEditInsightUC{output: sampleInsightCard()},
		&mockGetInsightQueueUC{
			output: &usecase.InsightQueueOutput{
				Insights:   []*entities.InsightCard{sampleInsightCard()},
				Pagination: usecase.PaginationOutput{Page: 1, Limit: 20, Total: 1},
			},
		},
		&mockGetClientInsightsUC{
			output: &usecase.InsightQueueOutput{
				Insights:   []*entities.InsightCard{sampleInsightCard()},
				Pagination: usecase.PaginationOutput{Page: 1, Limit: 20, Total: 1},
			},
		},
	)
}

// --- GetQueue tests ---

func TestInsightHandler_GetQueue_Success(t *testing.T) {
	h := defaultInsightHandler()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/insights/queue?page=1&limit=20", nil)
	req = withAuth(req)

	h.GetQueue(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp usecase.InsightQueueOutput
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Insights) != 1 {
		t.Errorf("expected 1 insight, got %d", len(resp.Insights))
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

func TestInsightHandler_GetQueue_ValidationError_Returns400(t *testing.T) {
	h := NewInsightHandler(
		&mockInsightTransitionUC{},
		&mockEditInsightUC{},
		&mockGetInsightQueueUC{err: &usecase.ValidationError{Message: "coach_id is required"}},
		&mockGetClientInsightsUC{},
	)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/insights/queue", nil)
	req = withAuth(req)

	h.GetQueue(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// --- GetClientInsights tests ---

func TestInsightHandler_GetClientInsights_Success(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Get("/api/v1/clients/{clientID}/insights", h.GetClientInsights)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/insights?status=draft", nil)
	req = withAuth(req)

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

func TestInsightHandler_GetClientInsights_InvalidStatus_Returns400(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Get("/api/v1/clients/{clientID}/insights", h.GetClientInsights)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/insights?status=invalid", nil)
	req = withAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// --- Approve tests ---

func TestInsightHandler_Approve_Success(t *testing.T) {
	approved := sampleInsightCard()
	approved.Status = entities.InsightStatusApproved

	h := NewInsightHandler(
		&mockInsightTransitionUC{output: approved},
		&mockEditInsightUC{},
		&mockGetInsightQueueUC{},
		&mockGetClientInsightsUC{},
	)

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/approve", h.Approve)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/insight-1/approve", nil)
	req = withAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp entities.InsightCard
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != entities.InsightStatusApproved {
		t.Errorf("expected status 'approved', got %q", resp.Status)
	}
}

func TestInsightHandler_Approve_NoAuth_Returns401(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/approve", h.Approve)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/insight-1/approve", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestInsightHandler_Approve_TransitionError_Returns422(t *testing.T) {
	h := NewInsightHandler(
		&mockInsightTransitionUC{err: &usecase.TransitionError{Message: "invalid transition", FromStatus: "dismissed", ToStatus: "approved"}},
		&mockEditInsightUC{},
		&mockGetInsightQueueUC{},
		&mockGetClientInsightsUC{},
	)

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/approve", h.Approve)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/insight-1/approve", nil)
	req = withAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", w.Code)
	}
}

func TestInsightHandler_Approve_AuthError_Returns403(t *testing.T) {
	h := NewInsightHandler(
		&mockInsightTransitionUC{err: &usecase.AuthenticationError{Message: "not authorized"}},
		&mockEditInsightUC{},
		&mockGetInsightQueueUC{},
		&mockGetClientInsightsUC{},
	)

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/approve", h.Approve)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/insight-1/approve", nil)
	req = withAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

// --- Dismiss tests ---

func TestInsightHandler_Dismiss_Success(t *testing.T) {
	dismissed := sampleInsightCard()
	dismissed.Status = entities.InsightStatusDismissed

	h := NewInsightHandler(
		&mockInsightTransitionUC{output: dismissed},
		&mockEditInsightUC{},
		&mockGetInsightQueueUC{},
		&mockGetClientInsightsUC{},
	)

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/dismiss", h.Dismiss)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/insight-1/dismiss", nil)
	req = withAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// --- Share tests ---

func TestInsightHandler_Share_Success(t *testing.T) {
	shared := sampleInsightCard()
	shared.Status = entities.InsightStatusShared

	h := NewInsightHandler(
		&mockInsightTransitionUC{output: shared},
		&mockEditInsightUC{},
		&mockGetInsightQueueUC{},
		&mockGetClientInsightsUC{},
	)

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}/share", h.Share)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/insight-1/share", nil)
	req = withAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// --- EditInsight tests ---

func TestInsightHandler_EditInsight_Success(t *testing.T) {
	edited := sampleInsightCard()
	edited.Title = "Updated Title"
	edited.Body = "Updated Body"

	h := NewInsightHandler(
		&mockInsightTransitionUC{},
		&mockEditInsightUC{output: edited},
		&mockGetInsightQueueUC{},
		&mockGetClientInsightsUC{},
	)

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}", h.EditInsight)

	body, _ := json.Marshal(editInsightRequest{Title: "Updated Title", Body: "Updated Body"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/insight-1", bytes.NewReader(body))
	req = withAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp entities.InsightCard
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %q", resp.Title)
	}
}

func TestInsightHandler_EditInsight_NoAuth_Returns401(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}", h.EditInsight)

	body, _ := json.Marshal(editInsightRequest{Title: "T", Body: "B"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/insight-1", bytes.NewReader(body))

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestInsightHandler_EditInsight_InvalidJSON_Returns400(t *testing.T) {
	h := defaultInsightHandler()

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}", h.EditInsight)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/insight-1", bytes.NewReader([]byte("not json")))
	req = withAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestInsightHandler_EditInsight_ValidationError_Returns400(t *testing.T) {
	h := NewInsightHandler(
		&mockInsightTransitionUC{},
		&mockEditInsightUC{err: &usecase.ValidationError{Message: "title is required"}},
		&mockGetInsightQueueUC{},
		&mockGetClientInsightsUC{},
	)

	r := chi.NewRouter()
	r.Put("/api/v1/insights/{insightID}", h.EditInsight)

	body, _ := json.Marshal(editInsightRequest{Title: "", Body: "B"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/insights/insight-1", bytes.NewReader(body))
	req = withAuth(req)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// --- parsePagination tests ---

func TestParsePagination_Defaults(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	page, limit := parsePagination(req)
	if page != 1 {
		t.Errorf("expected default page 1, got %d", page)
	}
	if limit != 20 {
		t.Errorf("expected default limit 20, got %d", limit)
	}
}

func TestParsePagination_CustomValues(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?page=3&limit=50", nil)
	page, limit := parsePagination(req)
	if page != 3 {
		t.Errorf("expected page 3, got %d", page)
	}
	if limit != 50 {
		t.Errorf("expected limit 50, got %d", limit)
	}
}

func TestParsePagination_InvalidValues_UseDefaults(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?page=abc&limit=-5", nil)
	page, limit := parsePagination(req)
	if page != 1 {
		t.Errorf("expected default page 1, got %d", page)
	}
	if limit != 20 {
		t.Errorf("expected default limit 20, got %d", limit)
	}
}

func TestParsePagination_LimitExceedsMax_UseDefault(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?limit=200", nil)
	_, limit := parsePagination(req)
	if limit != 20 {
		t.Errorf("expected default limit 20 for oversized limit, got %d", limit)
	}
}
