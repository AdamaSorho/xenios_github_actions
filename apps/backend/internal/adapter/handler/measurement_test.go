package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xenios/backend/internal/adapter/middleware"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

// --- Mock use cases ---

type mockMeasurementsUC struct {
	result *entities.MeasurementPage
	err    error
}

func (m *mockMeasurementsUC) Execute(_ context.Context, _ usecase.GetClientMeasurementsInput) (*entities.MeasurementPage, error) {
	return m.result, m.err
}

type mockLatestMeasurementsUC struct {
	result []*entities.Measurement
	err    error
}

func (m *mockLatestMeasurementsUC) Execute(_ context.Context, _, _ string) ([]*entities.Measurement, error) {
	return m.result, m.err
}

type mockWearableSummariesUC struct {
	result []*entities.WearableSummary
	err    error
}

func (m *mockWearableSummariesUC) Execute(_ context.Context, _, _ string, _ int) ([]*entities.WearableSummary, error) {
	return m.result, m.err
}

type mockProfileSummaryUC struct {
	result *entities.ProfileSummary
	err    error
}

func (m *mockProfileSummaryUC) Execute(_ context.Context, _, _ string) (*entities.ProfileSummary, error) {
	return m.result, m.err
}

func newMeasurementHandler(
	measUC *mockMeasurementsUC,
	latestUC *mockLatestMeasurementsUC,
	wearableUC *mockWearableSummariesUC,
	profileUC *mockProfileSummaryUC,
) *MeasurementHandler {
	return NewMeasurementHandler(measUC, latestUC, wearableUC, profileUC)
}

func requestWithClaims(r *http.Request, subject, role string) *http.Request {
	ctx := middleware.SetUserClaims(r.Context(), &middleware.UserClaims{Subject: subject, Role: role})
	return r.WithContext(ctx)
}

// --- ListMeasurements tests ---

func TestMeasurementHandler_ListMeasurements_Success(t *testing.T) {
	measUC := &mockMeasurementsUC{
		result: &entities.MeasurementPage{
			Measurements: []*entities.Measurement{
				{ID: "m1", Type: "weight", Value: 185.4, Unit: "lbs"},
			},
			Page: 1, Limit: 20, Total: 1,
		},
	}
	h := newMeasurementHandler(measUC, &mockLatestMeasurementsUC{}, &mockWearableSummariesUC{}, &mockProfileSummaryUC{})

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements?type=weight&page=1&limit=20", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = requestWithClaims(req, "coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var result entities.MeasurementPage
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result.Measurements) != 1 {
		t.Errorf("expected 1 measurement, got %d", len(result.Measurements))
	}
}

func TestMeasurementHandler_ListMeasurements_MissingClientID(t *testing.T) {
	h := newMeasurementHandler(&mockMeasurementsUC{}, &mockLatestMeasurementsUC{}, &mockWearableSummariesUC{}, &mockProfileSummaryUC{})

	req := httptest.NewRequest("GET", "/api/v1/clients//measurements", nil)
	req = newChiContext(req, map[string]string{"clientID": ""})
	req = requestWithClaims(req, "coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ListMeasurements_NoAuth(t *testing.T) {
	h := newMeasurementHandler(&mockMeasurementsUC{}, &mockLatestMeasurementsUC{}, &mockWearableSummariesUC{}, &mockProfileSummaryUC{})

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ListMeasurements_Forbidden(t *testing.T) {
	measUC := &mockMeasurementsUC{
		err: &usecase.AuthorizationError{Message: "not authorized"},
	}
	h := newMeasurementHandler(measUC, &mockLatestMeasurementsUC{}, &mockWearableSummariesUC{}, &mockProfileSummaryUC{})

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = requestWithClaims(req, "coach-1", "coach")
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
	if errResp.Error == "not authorized" {
		t.Error("internal error message should not be leaked")
	}
}

func TestMeasurementHandler_ListMeasurements_ValidationError(t *testing.T) {
	measUC := &mockMeasurementsUC{
		err: &usecase.ValidationError{Message: "client_id is required"},
	}
	h := newMeasurementHandler(measUC, &mockLatestMeasurementsUC{}, &mockWearableSummariesUC{}, &mockProfileSummaryUC{})

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = requestWithClaims(req, "coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ListMeasurements_InternalError(t *testing.T) {
	measUC := &mockMeasurementsUC{
		err: errors.New("database error"),
	}
	h := newMeasurementHandler(measUC, &mockLatestMeasurementsUC{}, &mockWearableSummariesUC{}, &mockProfileSummaryUC{})

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = requestWithClaims(req, "coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}

	var errResp ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Error == "database error" {
		t.Error("internal error message should not be leaked")
	}
}

func TestMeasurementHandler_ListMeasurements_WithDateFilter(t *testing.T) {
	measUC := &mockMeasurementsUC{
		result: &entities.MeasurementPage{
			Measurements: []*entities.Measurement{},
			Page: 1, Limit: 20, Total: 0,
		},
	}
	h := newMeasurementHandler(measUC, &mockLatestMeasurementsUC{}, &mockWearableSummariesUC{}, &mockProfileSummaryUC{})

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements?from=2026-01-01&to=2026-02-01", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = requestWithClaims(req, "coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

// --- LatestMeasurements tests ---

func TestMeasurementHandler_LatestMeasurements_Success(t *testing.T) {
	latestUC := &mockLatestMeasurementsUC{
		result: []*entities.Measurement{
			{ID: "m1", Type: "weight", Value: 185.4, Unit: "lbs"},
		},
	}
	h := newMeasurementHandler(&mockMeasurementsUC{}, latestUC, &mockWearableSummariesUC{}, &mockProfileSummaryUC{})

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements/latest", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = requestWithClaims(req, "coach-1", "coach")
	rec := httptest.NewRecorder()

	h.LatestMeasurements(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	measurements, ok := result["measurements"].([]interface{})
	if !ok {
		t.Fatal("expected measurements array in response")
	}
	if len(measurements) != 1 {
		t.Errorf("expected 1 measurement, got %d", len(measurements))
	}
}

func TestMeasurementHandler_LatestMeasurements_MissingClientID(t *testing.T) {
	h := newMeasurementHandler(&mockMeasurementsUC{}, &mockLatestMeasurementsUC{}, &mockWearableSummariesUC{}, &mockProfileSummaryUC{})

	req := httptest.NewRequest("GET", "/api/v1/clients//measurements/latest", nil)
	req = newChiContext(req, map[string]string{"clientID": ""})
	req = requestWithClaims(req, "coach-1", "coach")
	rec := httptest.NewRecorder()

	h.LatestMeasurements(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_LatestMeasurements_NoAuth(t *testing.T) {
	h := newMeasurementHandler(&mockMeasurementsUC{}, &mockLatestMeasurementsUC{}, &mockWearableSummariesUC{}, &mockProfileSummaryUC{})

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements/latest", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.LatestMeasurements(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestMeasurementHandler_LatestMeasurements_Forbidden(t *testing.T) {
	latestUC := &mockLatestMeasurementsUC{
		err: &usecase.AuthorizationError{Message: "not authorized"},
	}
	h := newMeasurementHandler(&mockMeasurementsUC{}, latestUC, &mockWearableSummariesUC{}, &mockProfileSummaryUC{})

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements/latest", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = requestWithClaims(req, "coach-1", "coach")
	rec := httptest.NewRecorder()

	h.LatestMeasurements(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestMeasurementHandler_LatestMeasurements_NilResult(t *testing.T) {
	latestUC := &mockLatestMeasurementsUC{result: nil}
	h := newMeasurementHandler(&mockMeasurementsUC{}, latestUC, &mockWearableSummariesUC{}, &mockProfileSummaryUC{})

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements/latest", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = requestWithClaims(req, "coach-1", "coach")
	rec := httptest.NewRecorder()

	h.LatestMeasurements(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&result)
	measurements, ok := result["measurements"].([]interface{})
	if !ok {
		t.Fatal("expected measurements array in response")
	}
	if len(measurements) != 0 {
		t.Errorf("expected 0 measurements, got %d", len(measurements))
	}
}

// --- WearableSummaries tests ---

func TestMeasurementHandler_WearableSummaries_Success(t *testing.T) {
	wearableUC := &mockWearableSummariesUC{
		result: []*entities.WearableSummary{
			{ID: "w1", Source: "whoop", SummaryDate: "2026-02-27"},
		},
	}
	h := newMeasurementHandler(&mockMeasurementsUC{}, &mockLatestMeasurementsUC{}, wearableUC, &mockProfileSummaryUC{})

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/wearable-summaries", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = requestWithClaims(req, "coach-1", "coach")
	rec := httptest.NewRecorder()

	h.WearableSummaries(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	summaries, ok := result["summaries"].([]interface{})
	if !ok {
		t.Fatal("expected summaries array in response")
	}
	if len(summaries) != 1 {
		t.Errorf("expected 1 summary, got %d", len(summaries))
	}
}

func TestMeasurementHandler_WearableSummaries_MissingClientID(t *testing.T) {
	h := newMeasurementHandler(&mockMeasurementsUC{}, &mockLatestMeasurementsUC{}, &mockWearableSummariesUC{}, &mockProfileSummaryUC{})

	req := httptest.NewRequest("GET", "/api/v1/clients//wearable-summaries", nil)
	req = newChiContext(req, map[string]string{"clientID": ""})
	req = requestWithClaims(req, "coach-1", "coach")
	rec := httptest.NewRecorder()

	h.WearableSummaries(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_WearableSummaries_NoAuth(t *testing.T) {
	h := newMeasurementHandler(&mockMeasurementsUC{}, &mockLatestMeasurementsUC{}, &mockWearableSummariesUC{}, &mockProfileSummaryUC{})

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/wearable-summaries", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.WearableSummaries(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestMeasurementHandler_WearableSummaries_Forbidden(t *testing.T) {
	wearableUC := &mockWearableSummariesUC{
		err: &usecase.AuthorizationError{Message: "not authorized"},
	}
	h := newMeasurementHandler(&mockMeasurementsUC{}, &mockLatestMeasurementsUC{}, wearableUC, &mockProfileSummaryUC{})

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/wearable-summaries", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = requestWithClaims(req, "coach-1", "coach")
	rec := httptest.NewRecorder()

	h.WearableSummaries(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestMeasurementHandler_WearableSummaries_NilResult(t *testing.T) {
	wearableUC := &mockWearableSummariesUC{result: nil}
	h := newMeasurementHandler(&mockMeasurementsUC{}, &mockLatestMeasurementsUC{}, wearableUC, &mockProfileSummaryUC{})

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/wearable-summaries", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = requestWithClaims(req, "coach-1", "coach")
	rec := httptest.NewRecorder()

	h.WearableSummaries(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&result)
	summaries, ok := result["summaries"].([]interface{})
	if !ok {
		t.Fatal("expected summaries array in response")
	}
	if len(summaries) != 0 {
		t.Errorf("expected 0 summaries, got %d", len(summaries))
	}
}

// --- ProfileSummary tests ---

func TestMeasurementHandler_ProfileSummary_Success(t *testing.T) {
	profileUC := &mockProfileSummaryUC{
		result: &entities.ProfileSummary{
			BodyComposition: map[string]*entities.LatestMeasurement{
				"weight": {Type: "weight", Value: 185.4, Unit: "lbs", Date: "2026-01-15"},
			},
			Labs:      &entities.LabSummary{FlaggedCount: 0, Markers: []*entities.LatestMeasurement{}},
			Wearable:  &entities.WearableAverage{Source: "whoop"},
			Nutrition: &entities.NutritionAverage{},
		},
	}
	h := newMeasurementHandler(&mockMeasurementsUC{}, &mockLatestMeasurementsUC{}, &mockWearableSummariesUC{}, profileUC)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/profile-summary", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = requestWithClaims(req, "coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ProfileSummary(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := result["bodyComposition"]; !ok {
		t.Error("expected bodyComposition in response")
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

func TestMeasurementHandler_ProfileSummary_MissingClientID(t *testing.T) {
	h := newMeasurementHandler(&mockMeasurementsUC{}, &mockLatestMeasurementsUC{}, &mockWearableSummariesUC{}, &mockProfileSummaryUC{})

	req := httptest.NewRequest("GET", "/api/v1/clients//profile-summary", nil)
	req = newChiContext(req, map[string]string{"clientID": ""})
	req = requestWithClaims(req, "coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ProfileSummary(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ProfileSummary_NoAuth(t *testing.T) {
	h := newMeasurementHandler(&mockMeasurementsUC{}, &mockLatestMeasurementsUC{}, &mockWearableSummariesUC{}, &mockProfileSummaryUC{})

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/profile-summary", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.ProfileSummary(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ProfileSummary_Forbidden(t *testing.T) {
	profileUC := &mockProfileSummaryUC{
		err: &usecase.AuthorizationError{Message: "not authorized"},
	}
	h := newMeasurementHandler(&mockMeasurementsUC{}, &mockLatestMeasurementsUC{}, &mockWearableSummariesUC{}, profileUC)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/profile-summary", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = requestWithClaims(req, "coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ProfileSummary(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ProfileSummary_InternalError(t *testing.T) {
	profileUC := &mockProfileSummaryUC{
		err: errors.New("database error"),
	}
	h := newMeasurementHandler(&mockMeasurementsUC{}, &mockLatestMeasurementsUC{}, &mockWearableSummariesUC{}, profileUC)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/profile-summary", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = requestWithClaims(req, "coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ProfileSummary(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}

	var errResp ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Error == "database error" {
		t.Error("internal error message should not be leaked")
	}
}
