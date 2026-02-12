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
	"github.com/xenios/backend/internal/domain"
	"github.com/xenios/backend/internal/usecase"
)

type mockCreateCoachClientUC struct {
	executeFunc func(ctx context.Context, input usecase.CreateCoachClientInput) (*domain.CoachClient, error)
}

func (m *mockCreateCoachClientUC) Execute(ctx context.Context, input usecase.CreateCoachClientInput) (*domain.CoachClient, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, input)
	}
	return &domain.CoachClient{
		ID:       "rel-1",
		CoachID:  input.CoachID,
		ClientID: input.ClientID,
		Status:   "active",
	}, nil
}

type mockListCoachClientsUC struct {
	executeFunc func(ctx context.Context, input usecase.ListCoachClientsInput) ([]*domain.CoachClient, error)
}

func (m *mockListCoachClientsUC) Execute(ctx context.Context, input usecase.ListCoachClientsInput) ([]*domain.CoachClient, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, input)
	}
	return []*domain.CoachClient{}, nil
}

// helper to create a request with chi URL params
func newRequestWithChiParams(method, url string, body *bytes.Buffer, params map[string]string) *http.Request {
	if body == nil {
		body = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, url, body)
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestCoachClientHandler_Create_Success(t *testing.T) {
	// Arrange
	createUC := &mockCreateCoachClientUC{}
	listUC := &mockListCoachClientsUC{}
	handler := NewCoachClientHandler(createUC, listUC)

	body, _ := json.Marshal(CreateCoachClientRequest{ClientID: "client-1"})
	req := newRequestWithChiParams(http.MethodPost, "/api/v1/coaches/coach-1/clients",
		bytes.NewBuffer(body), map[string]string{"coachID": "coach-1"})
	rec := httptest.NewRecorder()

	// Act
	handler.Create(rec, req)

	// Assert
	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	var response CreateCoachClientResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Data == nil {
		t.Fatal("expected non-nil data")
	}
	if response.Data.CoachID != "coach-1" {
		t.Errorf("expected coach_id 'coach-1', got '%s'", response.Data.CoachID)
	}
	if response.Data.ClientID != "client-1" {
		t.Errorf("expected client_id 'client-1', got '%s'", response.Data.ClientID)
	}
}

func TestCoachClientHandler_Create_InvalidJSON(t *testing.T) {
	// Arrange
	createUC := &mockCreateCoachClientUC{}
	listUC := &mockListCoachClientsUC{}
	handler := NewCoachClientHandler(createUC, listUC)

	req := newRequestWithChiParams(http.MethodPost, "/api/v1/coaches/coach-1/clients",
		bytes.NewBuffer([]byte("not json")), map[string]string{"coachID": "coach-1"})
	rec := httptest.NewRecorder()

	// Act
	handler.Create(rec, req)

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Code != "INVALID_JSON" {
		t.Errorf("expected code 'INVALID_JSON', got '%s'", response.Code)
	}
}

func TestCoachClientHandler_Create_MissingCoachID(t *testing.T) {
	// Arrange
	createUC := &mockCreateCoachClientUC{}
	listUC := &mockListCoachClientsUC{}
	handler := NewCoachClientHandler(createUC, listUC)

	body, _ := json.Marshal(CreateCoachClientRequest{ClientID: "client-1"})
	req := newRequestWithChiParams(http.MethodPost, "/api/v1/coaches//clients",
		bytes.NewBuffer(body), map[string]string{"coachID": ""})
	rec := httptest.NewRecorder()

	// Act
	handler.Create(rec, req)

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestCoachClientHandler_Create_UseCaseError(t *testing.T) {
	// Arrange
	createUC := &mockCreateCoachClientUC{
		executeFunc: func(ctx context.Context, input usecase.CreateCoachClientInput) (*domain.CoachClient, error) {
			return nil, errors.New("client_id is required")
		},
	}
	listUC := &mockListCoachClientsUC{}
	handler := NewCoachClientHandler(createUC, listUC)

	body, _ := json.Marshal(CreateCoachClientRequest{ClientID: ""})
	req := newRequestWithChiParams(http.MethodPost, "/api/v1/coaches/coach-1/clients",
		bytes.NewBuffer(body), map[string]string{"coachID": "coach-1"})
	rec := httptest.NewRecorder()

	// Act
	handler.Create(rec, req)

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code 'VALIDATION_ERROR', got '%s'", response.Code)
	}
}

func TestCoachClientHandler_List_Success(t *testing.T) {
	// Arrange
	listUC := &mockListCoachClientsUC{
		executeFunc: func(ctx context.Context, input usecase.ListCoachClientsInput) ([]*domain.CoachClient, error) {
			return []*domain.CoachClient{
				{ID: "1", CoachID: input.CoachID, ClientID: "client-1", Status: "active"},
				{ID: "2", CoachID: input.CoachID, ClientID: "client-2", Status: "active"},
			}, nil
		},
	}
	createUC := &mockCreateCoachClientUC{}
	handler := NewCoachClientHandler(createUC, listUC)

	req := newRequestWithChiParams(http.MethodGet, "/api/v1/coaches/coach-1/clients?limit=10&offset=0",
		nil, map[string]string{"coachID": "coach-1"})
	rec := httptest.NewRecorder()

	// Act
	handler.List(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var response ListCoachClientsResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Data) != 2 {
		t.Errorf("expected 2 clients, got %d", len(response.Data))
	}
}

func TestCoachClientHandler_List_EmptyResult(t *testing.T) {
	// Arrange
	listUC := &mockListCoachClientsUC{
		executeFunc: func(ctx context.Context, input usecase.ListCoachClientsInput) ([]*domain.CoachClient, error) {
			return nil, nil
		},
	}
	createUC := &mockCreateCoachClientUC{}
	handler := NewCoachClientHandler(createUC, listUC)

	req := newRequestWithChiParams(http.MethodGet, "/api/v1/coaches/coach-1/clients",
		nil, map[string]string{"coachID": "coach-1"})
	rec := httptest.NewRecorder()

	// Act
	handler.List(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var response ListCoachClientsResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Data == nil {
		t.Error("expected non-nil data array (empty, not null)")
	}
	if len(response.Data) != 0 {
		t.Errorf("expected 0 clients, got %d", len(response.Data))
	}
}

func TestCoachClientHandler_List_MissingCoachID(t *testing.T) {
	// Arrange
	createUC := &mockCreateCoachClientUC{}
	listUC := &mockListCoachClientsUC{}
	handler := NewCoachClientHandler(createUC, listUC)

	req := newRequestWithChiParams(http.MethodGet, "/api/v1/coaches//clients",
		nil, map[string]string{"coachID": ""})
	rec := httptest.NewRecorder()

	// Act
	handler.List(rec, req)

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestCoachClientHandler_List_UseCaseError(t *testing.T) {
	// Arrange
	listUC := &mockListCoachClientsUC{
		executeFunc: func(ctx context.Context, input usecase.ListCoachClientsInput) ([]*domain.CoachClient, error) {
			return nil, errors.New("coach_id is required")
		},
	}
	createUC := &mockCreateCoachClientUC{}
	handler := NewCoachClientHandler(createUC, listUC)

	req := newRequestWithChiParams(http.MethodGet, "/api/v1/coaches/coach-1/clients",
		nil, map[string]string{"coachID": "coach-1"})
	rec := httptest.NewRecorder()

	// Act
	handler.List(rec, req)

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestCoachClientHandler_List_ParsesQueryParams(t *testing.T) {
	// Arrange
	var capturedLimit, capturedOffset int
	listUC := &mockListCoachClientsUC{
		executeFunc: func(ctx context.Context, input usecase.ListCoachClientsInput) ([]*domain.CoachClient, error) {
			capturedLimit = input.Limit
			capturedOffset = input.Offset
			return []*domain.CoachClient{}, nil
		},
	}
	createUC := &mockCreateCoachClientUC{}
	handler := NewCoachClientHandler(createUC, listUC)

	req := newRequestWithChiParams(http.MethodGet, "/api/v1/coaches/coach-1/clients?limit=25&offset=50",
		nil, map[string]string{"coachID": "coach-1"})
	rec := httptest.NewRecorder()

	// Act
	handler.List(rec, req)

	// Assert
	if capturedLimit != 25 {
		t.Errorf("expected limit 25, got %d", capturedLimit)
	}
	if capturedOffset != 50 {
		t.Errorf("expected offset 50, got %d", capturedOffset)
	}
}

func TestCoachClientHandler_List_InvalidQueryParams(t *testing.T) {
	// Arrange - non-numeric query params should default to 0
	var capturedLimit, capturedOffset int
	listUC := &mockListCoachClientsUC{
		executeFunc: func(ctx context.Context, input usecase.ListCoachClientsInput) ([]*domain.CoachClient, error) {
			capturedLimit = input.Limit
			capturedOffset = input.Offset
			return []*domain.CoachClient{}, nil
		},
	}
	createUC := &mockCreateCoachClientUC{}
	handler := NewCoachClientHandler(createUC, listUC)

	req := newRequestWithChiParams(http.MethodGet, "/api/v1/coaches/coach-1/clients?limit=abc&offset=xyz",
		nil, map[string]string{"coachID": "coach-1"})
	rec := httptest.NewRecorder()

	// Act
	handler.List(rec, req)

	// Assert - strconv.Atoi returns 0 for invalid strings
	if capturedLimit != 0 {
		t.Errorf("expected limit 0 for invalid param, got %d", capturedLimit)
	}
	if capturedOffset != 0 {
		t.Errorf("expected offset 0 for invalid param, got %d", capturedOffset)
	}
}

func TestNewCoachClientHandler_NotNil(t *testing.T) {
	createUC := &mockCreateCoachClientUC{}
	listUC := &mockListCoachClientsUC{}
	handler := NewCoachClientHandler(createUC, listUC)
	if handler == nil {
		t.Error("expected non-nil handler")
	}
}
