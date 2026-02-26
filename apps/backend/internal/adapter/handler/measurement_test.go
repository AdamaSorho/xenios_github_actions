package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/xenios/backend/internal/adapter/middleware"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

// Mock use cases

type mockGetMeasurementsUC struct {
	output *usecase.GetClientMeasurementsOutput
	err    error
	input  usecase.GetClientMeasurementsInput
}

func (m *mockGetMeasurementsUC) Execute(ctx context.Context, input usecase.GetClientMeasurementsInput) (*usecase.GetClientMeasurementsOutput, error) {
	m.input = input
	if m.err != nil {
		return nil, m.err
	}
	if m.output != nil {
		return m.output, nil
	}
	return &usecase.GetClientMeasurementsOutput{
		Measurements: []*entities.Measurement{},
		Pagination:   usecase.PaginationInfo{Page: 1, Limit: 20, Total: 0},
	}, nil
}

type mockGetLatestMeasurementsUC struct {
	output *usecase.GetLatestMeasurementsOutput
	err    error
}

func (m *mockGetLatestMeasurementsUC) Execute(ctx context.Context, input usecase.GetLatestMeasurementsInput) (*usecase.GetLatestMeasurementsOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.output != nil {
		return m.output, nil
	}
	return &usecase.GetLatestMeasurementsOutput{
		Measurements: []*entities.Measurement{},
	}, nil
}

type mockGetWearableSummariesUC struct {
	output *usecase.GetWearableSummariesOutput
	err    error
	input  usecase.GetWearableSummariesInput
}

func (m *mockGetWearableSummariesUC) Execute(ctx context.Context, input usecase.GetWearableSummariesInput) (*usecase.GetWearableSummariesOutput, error) {
	m.input = input
	if m.err != nil {
		return nil, m.err
	}
	if m.output != nil {
		return m.output, nil
	}
	return &usecase.GetWearableSummariesOutput{
		Summaries: []*entities.WearableSummary{},
	}, nil
}

type mockGetProfileSummaryUC struct {
	output *usecase.ProfileSummaryOutput
	err    error
}

func (m *mockGetProfileSummaryUC) Execute(ctx context.Context, input usecase.GetClientProfileSummaryInput) (*usecase.ProfileSummaryOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.output != nil {
		return m.output, nil
	}
	return &usecase.ProfileSummaryOutput{
		BodyComposition: map[string]*usecase.MeasurementSummary{},
		Labs:            usecase.LabsSummary{Markers: []*usecase.LabMarker{}},
	}, nil
}

func newMeasurementHandler() (*MeasurementHandler, *mockGetMeasurementsUC, *mockGetLatestMeasurementsUC, *mockGetWearableSummariesUC, *mockGetProfileSummaryUC) {
	getMeasUC := &mockGetMeasurementsUC{}
	getLatestUC := &mockGetLatestMeasurementsUC{}
	getWearableUC := &mockGetWearableSummariesUC{}
	getProfileUC := &mockGetProfileSummaryUC{}
	h := NewMeasurementHandler(getMeasUC, getLatestUC, getWearableUC, getProfileUC)
	return h, getMeasUC, getLatestUC, getWearableUC, getProfileUC
}

func newAuthenticatedRequest(method, url string, coachID string) *http.Request {
	req := httptest.NewRequest(method, url, nil)
	ctx := middleware.SetUserClaims(req.Context(), &middleware.UserClaims{Subject: coachID, Role: "coach"})
	return req.WithContext(ctx)
}

func setChiParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// === ListMeasurements Tests ===

func TestMeasurementHandler_ListMeasurements_Success(t *testing.T) {
	h, getMeasUC, _, _, _ := newMeasurementHandler()

	now := time.Now()
	getMeasUC.output = &usecase.GetClientMeasurementsOutput{
		Measurements: []*entities.Measurement{
			{ID: "m1", Type: "weight", Value: 185.4, Unit: "lbs", MeasuredAt: now},
		},
		Pagination: usecase.PaginationInfo{Page: 1, Limit: 20, Total: 1},
	}

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/measurements?type=weight&page=1&limit=20", "coach-1")
	req = setChiParam(req, "clientID", "client-1")
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var result usecase.GetClientMeasurementsOutput
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result.Measurements) != 1 {
		t.Errorf("expected 1 measurement, got %d", len(result.Measurements))
	}
}

func TestMeasurementHandler_ListMeasurements_PassesInputCorrectly(t *testing.T) {
	h, getMeasUC, _, _, _ := newMeasurementHandler()

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/measurements?type=weight&page=2&limit=10&from=2026-01-01&to=2026-02-01", "coach-1")
	req = setChiParam(req, "clientID", "client-1")
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if getMeasUC.input.CoachID != "coach-1" {
		t.Errorf("expected coachID 'coach-1', got '%s'", getMeasUC.input.CoachID)
	}
	if getMeasUC.input.ClientID != "client-1" {
		t.Errorf("expected clientID 'client-1', got '%s'", getMeasUC.input.ClientID)
	}
	if getMeasUC.input.Type != "weight" {
		t.Errorf("expected type 'weight', got '%s'", getMeasUC.input.Type)
	}
	if getMeasUC.input.Page != 2 {
		t.Errorf("expected page 2, got %d", getMeasUC.input.Page)
	}
	if getMeasUC.input.Limit != 10 {
		t.Errorf("expected limit 10, got %d", getMeasUC.input.Limit)
	}
	if getMeasUC.input.From == nil {
		t.Error("expected from date to be set")
	}
	if getMeasUC.input.To == nil {
		t.Error("expected to date to be set")
	}
}

func TestMeasurementHandler_ListMeasurements_NoAuth(t *testing.T) {
	h, _, _, _, _ := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("clientID", "client-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ListMeasurements_MissingClientID(t *testing.T) {
	h, _, _, _, _ := newMeasurementHandler()

	req := newAuthenticatedRequest("GET", "/api/v1/clients//measurements", "coach-1")
	req = setChiParam(req, "clientID", "")
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ListMeasurements_Forbidden(t *testing.T) {
	h, getMeasUC, _, _, _ := newMeasurementHandler()
	getMeasUC.err = &usecase.AuthorizationError{Message: "not authorized"}

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/measurements", "coach-1")
	req = setChiParam(req, "clientID", "client-1")
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}

	var errResp ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Code != "FORBIDDEN" {
		t.Errorf("expected code FORBIDDEN, got %s", errResp.Code)
	}
	// Verify internal error details are NOT leaked
	if errResp.Error == "not authorized" {
		t.Error("internal authorization message should not be leaked to client")
	}
}

func TestMeasurementHandler_ListMeasurements_InternalError(t *testing.T) {
	h, getMeasUC, _, _, _ := newMeasurementHandler()
	getMeasUC.err = errors.New("database unavailable")

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/measurements", "coach-1")
	req = setChiParam(req, "clientID", "client-1")
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

// === LatestMeasurements Tests ===

func TestMeasurementHandler_LatestMeasurements_Success(t *testing.T) {
	h, _, getLatestUC, _, _ := newMeasurementHandler()

	now := time.Now()
	getLatestUC.output = &usecase.GetLatestMeasurementsOutput{
		Measurements: []*entities.Measurement{
			{ID: "m1", Type: "weight", Value: 185.4, Unit: "lbs", MeasuredAt: now},
			{ID: "m2", Type: "body_fat_pct", Value: 22.3, Unit: "%", MeasuredAt: now},
		},
	}

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/measurements/latest", "coach-1")
	req = setChiParam(req, "clientID", "client-1")
	rec := httptest.NewRecorder()

	h.LatestMeasurements(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestMeasurementHandler_LatestMeasurements_NoAuth(t *testing.T) {
	h, _, _, _, _ := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements/latest", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("clientID", "client-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()

	h.LatestMeasurements(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestMeasurementHandler_LatestMeasurements_Forbidden(t *testing.T) {
	h, _, getLatestUC, _, _ := newMeasurementHandler()
	getLatestUC.err = &usecase.AuthorizationError{Message: "not authorized"}

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/measurements/latest", "coach-1")
	req = setChiParam(req, "clientID", "client-1")
	rec := httptest.NewRecorder()

	h.LatestMeasurements(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestMeasurementHandler_LatestMeasurements_MissingClientID(t *testing.T) {
	h, _, _, _, _ := newMeasurementHandler()

	req := newAuthenticatedRequest("GET", "/api/v1/clients//measurements/latest", "coach-1")
	req = setChiParam(req, "clientID", "")
	rec := httptest.NewRecorder()

	h.LatestMeasurements(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

// === WearableSummaries Tests ===

func TestMeasurementHandler_WearableSummaries_Success(t *testing.T) {
	h, _, _, getWearableUC, _ := newMeasurementHandler()

	getWearableUC.output = &usecase.GetWearableSummariesOutput{
		Summaries: []*entities.WearableSummary{
			{ID: "w1", Source: "whoop", SummaryDate: "2026-01-15"},
		},
	}

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/wearable-summaries", "coach-1")
	req = setChiParam(req, "clientID", "client-1")
	rec := httptest.NewRecorder()

	h.WearableSummaries(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestMeasurementHandler_WearableSummaries_WithDays(t *testing.T) {
	h, _, _, getWearableUC, _ := newMeasurementHandler()

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/wearable-summaries?days=14", "coach-1")
	req = setChiParam(req, "clientID", "client-1")
	rec := httptest.NewRecorder()

	h.WearableSummaries(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if getWearableUC.input.Days != 14 {
		t.Errorf("expected days 14, got %d", getWearableUC.input.Days)
	}
}

func TestMeasurementHandler_WearableSummaries_NoAuth(t *testing.T) {
	h, _, _, _, _ := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/wearable-summaries", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("clientID", "client-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()

	h.WearableSummaries(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestMeasurementHandler_WearableSummaries_Forbidden(t *testing.T) {
	h, _, _, getWearableUC, _ := newMeasurementHandler()
	getWearableUC.err = &usecase.AuthorizationError{Message: "not authorized"}

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/wearable-summaries", "coach-1")
	req = setChiParam(req, "clientID", "client-1")
	rec := httptest.NewRecorder()

	h.WearableSummaries(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

// === ProfileSummary Tests ===

func TestMeasurementHandler_ProfileSummary_Success(t *testing.T) {
	h, _, _, _, getProfileUC := newMeasurementHandler()

	getProfileUC.output = &usecase.ProfileSummaryOutput{
		BodyComposition: map[string]*usecase.MeasurementSummary{
			"weight": {Value: 185.4, Unit: "lbs", Date: "2026-01-15"},
		},
		Labs: usecase.LabsSummary{
			FlaggedCount: 1,
			LastTestDate: "2026-01-10",
			Markers:      []*usecase.LabMarker{{Type: "ldl_cholesterol", Value: 142, Unit: "mg/dL", Flag: "high"}},
		},
		Wearable: &usecase.WearableSummaryInfo{
			Source: "whoop", AvgHrv7d: 45.2, AvgSleep7d: 7.2, AvgRecovery7d: 68,
		},
		Nutrition: usecase.NutritionSummary{AvgCalories7d: 2150, AvgProtein7d: 165},
	}

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/profile-summary", "coach-1")
	req = setChiParam(req, "clientID", "client-1")
	rec := httptest.NewRecorder()

	h.ProfileSummary(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := result["body_composition"]; !ok {
		t.Error("expected body_composition in response")
	}
	if _, ok := result["labs"]; !ok {
		t.Error("expected labs in response")
	}
	if _, ok := result["wearable"]; !ok {
		t.Error("expected wearable in response")
	}
	if _, ok := result["nutrition"]; !ok {
		t.Error("expected nutrition in response")
	}
}

func TestMeasurementHandler_ProfileSummary_NoAuth(t *testing.T) {
	h, _, _, _, _ := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/profile-summary", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("clientID", "client-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()

	h.ProfileSummary(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ProfileSummary_Forbidden(t *testing.T) {
	h, _, _, _, getProfileUC := newMeasurementHandler()
	getProfileUC.err = &usecase.AuthorizationError{Message: "not authorized"}

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/profile-summary", "coach-1")
	req = setChiParam(req, "clientID", "client-1")
	rec := httptest.NewRecorder()

	h.ProfileSummary(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ProfileSummary_MissingClientID(t *testing.T) {
	h, _, _, _, _ := newMeasurementHandler()

	req := newAuthenticatedRequest("GET", "/api/v1/clients//profile-summary", "coach-1")
	req = setChiParam(req, "clientID", "")
	rec := httptest.NewRecorder()

	h.ProfileSummary(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ProfileSummary_InternalError(t *testing.T) {
	h, _, _, _, getProfileUC := newMeasurementHandler()
	getProfileUC.err = errors.New("database unavailable")

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/profile-summary", "coach-1")
	req = setChiParam(req, "clientID", "client-1")
	rec := httptest.NewRecorder()

	h.ProfileSummary(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}

	var errResp ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Error == "database unavailable" {
		t.Error("internal error message should not be leaked to client")
	}
}

// === handleUseCaseError Tests ===

func TestMeasurementHandler_ListMeasurements_RFC3339Dates(t *testing.T) {
	h, getMeasUC, _, _, _ := newMeasurementHandler()

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/measurements?from=2026-01-15T10:00:00Z&to=2026-02-15T18:00:00Z", "coach-1")
	req = setChiParam(req, "clientID", "client-1")
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if getMeasUC.input.From == nil {
		t.Fatal("expected from date to be set")
	}
	if getMeasUC.input.To == nil {
		t.Fatal("expected to date to be set")
	}
	expectedFrom := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	if !getMeasUC.input.From.Equal(expectedFrom) {
		t.Errorf("expected from %v, got %v", expectedFrom, *getMeasUC.input.From)
	}
	expectedTo := time.Date(2026, 2, 15, 18, 0, 0, 0, time.UTC)
	if !getMeasUC.input.To.Equal(expectedTo) {
		t.Errorf("expected to %v, got %v", expectedTo, *getMeasUC.input.To)
	}
}

func TestMeasurementHandler_ListMeasurements_ValidationError(t *testing.T) {
	h, getMeasUC, _, _, _ := newMeasurementHandler()
	getMeasUC.err = &usecase.ValidationError{Message: "client_id is required"}

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/measurements", "coach-1")
	req = setChiParam(req, "clientID", "client-1")
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
	var errResp ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code VALIDATION_ERROR, got %s", errResp.Code)
	}
}

func TestMeasurementHandler_LatestMeasurements_InternalError(t *testing.T) {
	h, _, getLatestUC, _, _ := newMeasurementHandler()
	getLatestUC.err = errors.New("database unavailable")

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/measurements/latest", "coach-1")
	req = setChiParam(req, "clientID", "client-1")
	rec := httptest.NewRecorder()

	h.LatestMeasurements(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
	var errResp ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Error == "database unavailable" {
		t.Error("internal error message should not be leaked to client")
	}
}

func TestMeasurementHandler_LatestMeasurements_ValidationError(t *testing.T) {
	h, _, getLatestUC, _, _ := newMeasurementHandler()
	getLatestUC.err = &usecase.ValidationError{Message: "coach_id is required"}

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/measurements/latest", "coach-1")
	req = setChiParam(req, "clientID", "client-1")
	rec := httptest.NewRecorder()

	h.LatestMeasurements(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_WearableSummaries_MissingClientID(t *testing.T) {
	h, _, _, _, _ := newMeasurementHandler()

	req := newAuthenticatedRequest("GET", "/api/v1/clients//wearable-summaries", "coach-1")
	req = setChiParam(req, "clientID", "")
	rec := httptest.NewRecorder()

	h.WearableSummaries(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_WearableSummaries_InternalError(t *testing.T) {
	h, _, _, getWearableUC, _ := newMeasurementHandler()
	getWearableUC.err = errors.New("database unavailable")

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/wearable-summaries", "coach-1")
	req = setChiParam(req, "clientID", "client-1")
	rec := httptest.NewRecorder()

	h.WearableSummaries(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestMeasurementHandler_WearableSummaries_ValidationError(t *testing.T) {
	h, _, _, getWearableUC, _ := newMeasurementHandler()
	getWearableUC.err = &usecase.ValidationError{Message: "coach_id is required"}

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/wearable-summaries", "coach-1")
	req = setChiParam(req, "clientID", "client-1")
	rec := httptest.NewRecorder()

	h.WearableSummaries(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ProfileSummary_ValidationError(t *testing.T) {
	h, _, _, _, getProfileUC := newMeasurementHandler()
	getProfileUC.err = &usecase.ValidationError{Message: "coach_id is required"}

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/profile-summary", "coach-1")
	req = setChiParam(req, "clientID", "client-1")
	rec := httptest.NewRecorder()

	h.ProfileSummary(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestHandleUseCaseError_ValidationError(t *testing.T) {
	rec := httptest.NewRecorder()
	handleUseCaseError(rec, &usecase.ValidationError{Message: "bad input"})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleUseCaseError_AuthorizationError(t *testing.T) {
	rec := httptest.NewRecorder()
	handleUseCaseError(rec, &usecase.AuthorizationError{Message: "denied"})
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestHandleUseCaseError_InternalError(t *testing.T) {
	rec := httptest.NewRecorder()
	handleUseCaseError(rec, errors.New("something went wrong"))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}
