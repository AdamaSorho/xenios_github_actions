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

type mockCreateCoachClientUC struct {
	result *entities.CoachClient
	err    error
}

func (m *mockCreateCoachClientUC) Execute(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.result != nil {
		return m.result, nil
	}
	return &entities.CoachClient{ID: "new-id", CoachID: coachID, ClientID: clientID}, nil
}

type mockListCoachClientsUC struct {
	results []*entities.CoachClient
	err     error
}

func (m *mockListCoachClientsUC) Execute(ctx context.Context, coachID string, limit, offset int) ([]*entities.CoachClient, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.results, nil
}

func newChiContext(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestCoachClientHandler_Create_Success(t *testing.T) {
	createUC := &mockCreateCoachClientUC{}
	listUC := &mockListCoachClientsUC{}
	h := NewCoachClientHandler(createUC, listUC)

	body := `{"client_id": "client-1"}`
	req := httptest.NewRequest("POST", "/api/v1/coaches/coach-1/clients", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = newChiContext(req, map[string]string{"coachID": "coach-1"})
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	var result entities.CoachClient
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.CoachID != "coach-1" {
		t.Errorf("expected CoachID coach-1, got %s", result.CoachID)
	}
	if result.ClientID != "client-1" {
		t.Errorf("expected ClientID client-1, got %s", result.ClientID)
	}
}

func TestCoachClientHandler_Create_InvalidJSON(t *testing.T) {
	createUC := &mockCreateCoachClientUC{}
	listUC := &mockListCoachClientsUC{}
	h := NewCoachClientHandler(createUC, listUC)

	req := httptest.NewRequest("POST", "/api/v1/coaches/coach-1/clients", bytes.NewBufferString("not json"))
	req = newChiContext(req, map[string]string{"coachID": "coach-1"})
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var errResp ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Code != "INVALID_JSON" {
		t.Errorf("expected code INVALID_JSON, got %s", errResp.Code)
	}
}

func TestCoachClientHandler_Create_MissingCoachID(t *testing.T) {
	createUC := &mockCreateCoachClientUC{}
	listUC := &mockListCoachClientsUC{}
	h := NewCoachClientHandler(createUC, listUC)

	body := `{"client_id": "client-1"}`
	req := httptest.NewRequest("POST", "/api/v1/coaches//clients", bytes.NewBufferString(body))
	req = newChiContext(req, map[string]string{"coachID": ""})
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestCoachClientHandler_Create_ValidationError(t *testing.T) {
	createUC := &mockCreateCoachClientUC{
		err: &usecase.ValidationError{Message: "client_id is required"},
	}
	listUC := &mockListCoachClientsUC{}
	h := NewCoachClientHandler(createUC, listUC)

	body := `{"client_id": ""}`
	req := httptest.NewRequest("POST", "/api/v1/coaches/coach-1/clients", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = newChiContext(req, map[string]string{"coachID": "coach-1"})
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var errResp ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code VALIDATION_ERROR, got %s", errResp.Code)
	}
}

func TestCoachClientHandler_Create_InternalError(t *testing.T) {
	createUC := &mockCreateCoachClientUC{
		err: errors.New("database unavailable"),
	}
	listUC := &mockListCoachClientsUC{}
	h := NewCoachClientHandler(createUC, listUC)

	body := `{"client_id": "client-1"}`
	req := httptest.NewRequest("POST", "/api/v1/coaches/coach-1/clients", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = newChiContext(req, map[string]string{"coachID": "coach-1"})
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}

	var errResp ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Code != "INTERNAL_ERROR" {
		t.Errorf("expected code INTERNAL_ERROR, got %s", errResp.Code)
	}
	// Verify internal error details are NOT leaked
	if errResp.Error == "database unavailable" {
		t.Error("internal error message should not be leaked to client")
	}
}

func TestCoachClientHandler_List_Success(t *testing.T) {
	createUC := &mockCreateCoachClientUC{}
	listUC := &mockListCoachClientsUC{
		results: []*entities.CoachClient{
			{ID: "1", CoachID: "coach-1", ClientID: "client-a"},
			{ID: "2", CoachID: "coach-1", ClientID: "client-b"},
		},
	}
	h := NewCoachClientHandler(createUC, listUC)

	req := httptest.NewRequest("GET", "/api/v1/coaches/coach-1/clients?limit=20&offset=0", nil)
	req = newChiContext(req, map[string]string{"coachID": "coach-1"})
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	data, ok := result["data"].([]interface{})
	if !ok {
		t.Fatal("expected data array in response")
	}
	if len(data) != 2 {
		t.Errorf("expected 2 items in data, got %d", len(data))
	}
}

func TestCoachClientHandler_List_MissingCoachID(t *testing.T) {
	createUC := &mockCreateCoachClientUC{}
	listUC := &mockListCoachClientsUC{}
	h := NewCoachClientHandler(createUC, listUC)

	req := httptest.NewRequest("GET", "/api/v1/coaches//clients", nil)
	req = newChiContext(req, map[string]string{"coachID": ""})
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestCoachClientHandler_List_ValidationError(t *testing.T) {
	createUC := &mockCreateCoachClientUC{}
	listUC := &mockListCoachClientsUC{
		err: &usecase.ValidationError{Message: "limit must be non-negative"},
	}
	h := NewCoachClientHandler(createUC, listUC)

	req := httptest.NewRequest("GET", "/api/v1/coaches/coach-1/clients?limit=-1", nil)
	req = newChiContext(req, map[string]string{"coachID": "coach-1"})
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var errResp ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code VALIDATION_ERROR, got %s", errResp.Code)
	}
}

func TestCoachClientHandler_List_InternalError(t *testing.T) {
	createUC := &mockCreateCoachClientUC{}
	listUC := &mockListCoachClientsUC{
		err: errors.New("database error"),
	}
	h := NewCoachClientHandler(createUC, listUC)

	req := httptest.NewRequest("GET", "/api/v1/coaches/coach-1/clients", nil)
	req = newChiContext(req, map[string]string{"coachID": "coach-1"})
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}

	var errResp ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Code != "INTERNAL_ERROR" {
		t.Errorf("expected code INTERNAL_ERROR, got %s", errResp.Code)
	}
	if errResp.Error == "database error" {
		t.Error("internal error message should not be leaked to client")
	}
}

func TestCoachClientHandler_List_EmptyResults(t *testing.T) {
	createUC := &mockCreateCoachClientUC{}
	listUC := &mockListCoachClientsUC{
		results: nil,
	}
	h := NewCoachClientHandler(createUC, listUC)

	req := httptest.NewRequest("GET", "/api/v1/coaches/coach-1/clients", nil)
	req = newChiContext(req, map[string]string{"coachID": "coach-1"})
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&result)
	data, ok := result["data"].([]interface{})
	if !ok {
		t.Fatal("expected data array in response")
	}
	if len(data) != 0 {
		t.Errorf("expected 0 items, got %d", len(data))
	}
}
