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

// --- mock use cases ---

type mockGetMeasurementsUC struct {
	result *entities.MeasurementResult
	err    error
}

func (m *mockGetMeasurementsUC) Execute(_ context.Context, _ usecase.GetClientMeasurementsInput) (*entities.MeasurementResult, error) {
	return m.result, m.err
}

type mockGetLatestUC struct {
	result []*entities.LatestMeasurement
	err    error
}

func (m *mockGetLatestUC) Execute(_ context.Context, _ usecase.GetLatestMeasurementsInput) ([]*entities.LatestMeasurement, error) {
	return m.result, m.err
}

type mockGetWearableSumUC struct {
	result []*entities.WearableSummary
	err    error
}

func (m *mockGetWearableSumUC) Execute(_ context.Context, _ usecase.GetWearableSummariesInput) ([]*entities.WearableSummary, error) {
	return m.result, m.err
}

type mockGetProfileSummaryUC struct {
	result *entities.ProfileSummary
	err    error
}

func (m *mockGetProfileSummaryUC) Execute(_ context.Context, _ usecase.GetClientProfileSummaryInput) (*entities.ProfileSummary, error) {
	return m.result, m.err
}

// --- helpers ---

func newMeasurementHandler() *MeasurementHandler {
	return NewMeasurementHandler(
		&mockGetMeasurementsUC{result: &entities.MeasurementResult{
			Measurements: []*entities.Measurement{},
			Pagination:   entities.Pagination{Page: 1, Limit: 20, Total: 0},
		}},
		&mockGetLatestUC{result: []*entities.LatestMeasurement{}},
		&mockGetWearableSumUC{result: []*entities.WearableSummary{}},
		&mockGetProfileSummaryUC{result: &entities.ProfileSummary{
			BodyComposition: map[string]*entities.LatestMeasurement{},
			Labs:            &entities.LabSummary{Markers: []*entities.LatestMeasurement{}},
			Wearable:        &entities.WearableAverages{},
			Nutrition:       &entities.NutritionAverages{},
		}},
	)
}

func newAuthenticatedRequest(method, url string) *http.Request {
	req := httptest.NewRequest(method, url, nil)
	ctx := middleware.SetUserClaims(req.Context(), &middleware.UserClaims{Subject: "coach-1", Role: "coach"})
	return req.WithContext(ctx)
}

func withChiParams(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// --- ListMeasurements tests ---

func TestMeasurementHandler_ListMeasurements_Success(t *testing.T) {
	now := time.Now()
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{result: &entities.MeasurementResult{
			Measurements: []*entities.Measurement{
				{ID: "m-1", MeasurementType: "weight", Value: 185.4, Unit: "lbs", MeasuredAt: now},
			},
			Pagination: entities.Pagination{Page: 1, Limit: 20, Total: 1},
		}},
		&mockGetLatestUC{}, &mockGetWearableSumUC{}, &mockGetProfileSummaryUC{},
	)

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/measurements?page=1&limit=20")
	req = withChiParams(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var result entities.MeasurementResult
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(result.Measurements) != 1 {
		t.Errorf("expected 1 measurement, got %d", len(result.Measurements))
	}
}

func TestMeasurementHandler_ListMeasurements_Unauthorized(t *testing.T) {
	h := newMeasurementHandler()
	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements", nil)
	req = withChiParams(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ListMeasurements_MissingClientID(t *testing.T) {
	h := newMeasurementHandler()
	req := newAuthenticatedRequest("GET", "/api/v1/clients//measurements")
	req = withChiParams(req, map[string]string{"clientID": ""})
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ListMeasurements_Forbidden(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{err: &usecase.AuthorizationError{Message: "not authorized"}},
		&mockGetLatestUC{}, &mockGetWearableSumUC{}, &mockGetProfileSummaryUC{},
	)

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/measurements")
	req = withChiParams(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
	var errResp ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Code != "FORBIDDEN" {
		t.Errorf("expected code FORBIDDEN, got %s", errResp.Code)
	}
}

func TestMeasurementHandler_ListMeasurements_InternalError(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{err: errors.New("db error")},
		&mockGetLatestUC{}, &mockGetWearableSumUC{}, &mockGetProfileSummaryUC{},
	)

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/measurements")
	req = withChiParams(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
	var errResp ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Error == "db error" {
		t.Error("internal error should not be leaked")
	}
}

func TestMeasurementHandler_ListMeasurements_ValidationError(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{err: &usecase.ValidationError{Message: "invalid input"}},
		&mockGetLatestUC{}, &mockGetWearableSumUC{}, &mockGetProfileSummaryUC{},
	)

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/measurements")
	req = withChiParams(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// --- LatestMeasurements tests ---

func TestMeasurementHandler_LatestMeasurements_Success(t *testing.T) {
	now := time.Now()
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{},
		&mockGetLatestUC{result: []*entities.LatestMeasurement{
			{MeasurementType: "weight", Value: 183.0, Unit: "lbs", MeasuredAt: now},
		}},
		&mockGetWearableSumUC{}, &mockGetProfileSummaryUC{},
	)

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/measurements/latest")
	req = withChiParams(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.LatestMeasurements(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var result map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&result)
	measurements, ok := result["measurements"].([]interface{})
	if !ok {
		t.Fatal("expected measurements array")
	}
	if len(measurements) != 1 {
		t.Errorf("expected 1, got %d", len(measurements))
	}
}

func TestMeasurementHandler_LatestMeasurements_Unauthorized(t *testing.T) {
	h := newMeasurementHandler()
	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements/latest", nil)
	req = withChiParams(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.LatestMeasurements(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestMeasurementHandler_LatestMeasurements_MissingClientID(t *testing.T) {
	h := newMeasurementHandler()
	req := newAuthenticatedRequest("GET", "/api/v1/clients//measurements/latest")
	req = withChiParams(req, map[string]string{"clientID": ""})
	rec := httptest.NewRecorder()

	h.LatestMeasurements(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_LatestMeasurements_Forbidden(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{},
		&mockGetLatestUC{err: &usecase.AuthorizationError{Message: "not authorized"}},
		&mockGetWearableSumUC{}, &mockGetProfileSummaryUC{},
	)

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/measurements/latest")
	req = withChiParams(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.LatestMeasurements(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

// --- WearableSummaries tests ---

func TestMeasurementHandler_WearableSummaries_Success(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{}, &mockGetLatestUC{},
		&mockGetWearableSumUC{result: []*entities.WearableSummary{
			{ID: "ws-1", Source: "whoop", SummaryDate: "2026-01-15"},
		}},
		&mockGetProfileSummaryUC{},
	)

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/wearable-summaries")
	req = withChiParams(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.WearableSummaries(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var result map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&result)
	summaries, ok := result["summaries"].([]interface{})
	if !ok {
		t.Fatal("expected summaries array")
	}
	if len(summaries) != 1 {
		t.Errorf("expected 1, got %d", len(summaries))
	}
}

func TestMeasurementHandler_WearableSummaries_Unauthorized(t *testing.T) {
	h := newMeasurementHandler()
	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/wearable-summaries", nil)
	req = withChiParams(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.WearableSummaries(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestMeasurementHandler_WearableSummaries_MissingClientID(t *testing.T) {
	h := newMeasurementHandler()
	req := newAuthenticatedRequest("GET", "/api/v1/clients//wearable-summaries")
	req = withChiParams(req, map[string]string{"clientID": ""})
	rec := httptest.NewRecorder()

	h.WearableSummaries(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_WearableSummaries_Forbidden(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{}, &mockGetLatestUC{},
		&mockGetWearableSumUC{err: &usecase.AuthorizationError{Message: "not authorized"}},
		&mockGetProfileSummaryUC{},
	)

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/wearable-summaries")
	req = withChiParams(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.WearableSummaries(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestMeasurementHandler_WearableSummaries_InternalError(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{}, &mockGetLatestUC{},
		&mockGetWearableSumUC{err: errors.New("db error")},
		&mockGetProfileSummaryUC{},
	)

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/wearable-summaries")
	req = withChiParams(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.WearableSummaries(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

// --- ProfileSummary tests ---

func TestMeasurementHandler_ProfileSummary_Success(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{}, &mockGetLatestUC{}, &mockGetWearableSumUC{},
		&mockGetProfileSummaryUC{result: &entities.ProfileSummary{
			BodyComposition: map[string]*entities.LatestMeasurement{
				"weight": {MeasurementType: "weight", Value: 185.4, Unit: "lbs", MeasuredAt: time.Now()},
			},
			Labs:      &entities.LabSummary{FlaggedCount: 0, Markers: []*entities.LatestMeasurement{}},
			Wearable:  &entities.WearableAverages{},
			Nutrition: &entities.NutritionAverages{},
		}},
	)

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/profile-summary")
	req = withChiParams(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.ProfileSummary(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var result map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&result)
	if result["bodyComposition"] == nil {
		t.Error("expected bodyComposition in response")
	}
}

func TestMeasurementHandler_ProfileSummary_Unauthorized(t *testing.T) {
	h := newMeasurementHandler()
	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/profile-summary", nil)
	req = withChiParams(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.ProfileSummary(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ProfileSummary_MissingClientID(t *testing.T) {
	h := newMeasurementHandler()
	req := newAuthenticatedRequest("GET", "/api/v1/clients//profile-summary")
	req = withChiParams(req, map[string]string{"clientID": ""})
	rec := httptest.NewRecorder()

	h.ProfileSummary(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ProfileSummary_Forbidden(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{}, &mockGetLatestUC{}, &mockGetWearableSumUC{},
		&mockGetProfileSummaryUC{err: &usecase.AuthorizationError{Message: "not authorized"}},
	)

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/profile-summary")
	req = withChiParams(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.ProfileSummary(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ProfileSummary_InternalError(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{}, &mockGetLatestUC{}, &mockGetWearableSumUC{},
		&mockGetProfileSummaryUC{err: errors.New("db error")},
	)

	req := newAuthenticatedRequest("GET", "/api/v1/clients/client-1/profile-summary")
	req = withChiParams(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.ProfileSummary(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
	var errResp ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Error == "db error" {
		t.Error("internal error should not be leaked")
	}
}

// --- handleUseCaseError tests ---

func TestHandleUseCaseError_AuthenticationError(t *testing.T) {
	rec := httptest.NewRecorder()
	handleUseCaseError(rec, &usecase.AuthenticationError{Message: "invalid credentials"})

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

// --- parseMeasurementFilter tests ---

func TestParseMeasurementFilter_WithDateRange(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?type=weight&from=2026-01-01&to=2026-02-01&page=2&limit=10", nil)
	filter := parseMeasurementFilter(req)

	if filter.MeasurementType != "weight" {
		t.Errorf("expected type weight, got %s", filter.MeasurementType)
	}
	if filter.Page != 2 {
		t.Errorf("expected page 2, got %d", filter.Page)
	}
	if filter.Limit != 10 {
		t.Errorf("expected limit 10, got %d", filter.Limit)
	}
	if filter.From == nil {
		t.Error("expected from to be set")
	}
	if filter.To == nil {
		t.Error("expected to to be set")
	}
}

func TestParseMeasurementFilter_RFC3339Dates(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?from=2026-01-01T00:00:00Z&to=2026-02-01T23:59:59Z", nil)
	filter := parseMeasurementFilter(req)

	if filter.From == nil {
		t.Error("expected RFC3339 from date to be parsed")
	}
	if filter.To == nil {
		t.Error("expected RFC3339 to date to be parsed")
	}
}

func TestParseMeasurementFilter_NoParams(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	filter := parseMeasurementFilter(req)

	if filter.MeasurementType != "" {
		t.Errorf("expected empty type, got %s", filter.MeasurementType)
	}
	if filter.From != nil {
		t.Error("expected nil from")
	}
	if filter.To != nil {
		t.Error("expected nil to")
	}
}
