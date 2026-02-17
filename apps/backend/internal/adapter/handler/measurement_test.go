package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/xenios/backend/internal/adapter/middleware"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

// Mock use cases

type mockGetMeasurementsUC struct {
	output *usecase.GetMeasurementsOutput
	err    error
}

func (m *mockGetMeasurementsUC) Execute(ctx context.Context, input usecase.GetMeasurementsInput) (*usecase.GetMeasurementsOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.output != nil {
		return m.output, nil
	}
	return &usecase.GetMeasurementsOutput{
		Measurements: []*entities.Measurement{},
		Pagination:   usecase.PaginationInfo{Page: 1, Limit: 20, Total: 0},
	}, nil
}

type mockGetLatestMeasurementsUC struct {
	results []*entities.LatestMeasurement
	err     error
}

func (m *mockGetLatestMeasurementsUC) Execute(ctx context.Context, coachID, clientID string) ([]*entities.LatestMeasurement, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.results != nil {
		return m.results, nil
	}
	return []*entities.LatestMeasurement{}, nil
}

type mockGetWearableSummariesUC struct {
	output *usecase.GetWearableSummariesOutput
	err    error
}

func (m *mockGetWearableSummariesUC) Execute(ctx context.Context, input usecase.GetWearableSummariesInput) (*usecase.GetWearableSummariesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.output != nil {
		return m.output, nil
	}
	return &usecase.GetWearableSummariesOutput{
		Summaries:  []*entities.WearableSummary{},
		Pagination: usecase.PaginationInfo{Page: 1, Limit: 20, Total: 0},
	}, nil
}

type mockGetProfileSummaryUC struct {
	result *entities.ProfileSummary
	err    error
}

func (m *mockGetProfileSummaryUC) Execute(ctx context.Context, coachID, clientID string) (*entities.ProfileSummary, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.result != nil {
		return m.result, nil
	}
	return &entities.ProfileSummary{
		BodyComposition: map[string]*entities.LatestMeasurement{},
		Labs:            &entities.LabSummary{Markers: []entities.LatestMeasurement{}},
	}, nil
}

func newMeasurementHandler() *MeasurementHandler {
	return NewMeasurementHandler(
		&mockGetMeasurementsUC{},
		&mockGetLatestMeasurementsUC{},
		&mockGetWearableSummariesUC{},
		&mockGetProfileSummaryUC{},
	)
}

func withAuthClaims(r *http.Request, subject, role string) *http.Request {
	ctx := middleware.SetUserClaims(r.Context(), &middleware.UserClaims{Subject: subject, Role: role})
	return r.WithContext(ctx)
}

// Tests for ListMeasurements

func TestMeasurementHandler_ListMeasurements_Success(t *testing.T) {
	now := time.Now()
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{
			output: &usecase.GetMeasurementsOutput{
				Measurements: []*entities.Measurement{
					{ID: "m1", ClientID: "client-1", MeasurementType: "weight", Value: 185.4, Unit: "lbs", MeasuredAt: now},
				},
				Pagination: usecase.PaginationInfo{Page: 1, Limit: 20, Total: 1},
			},
		},
		&mockGetLatestMeasurementsUC{},
		&mockGetWearableSummariesUC{},
		&mockGetProfileSummaryUC{},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements?page=1&limit=20", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = withAuthClaims(req,"coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

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

func TestMeasurementHandler_ListMeasurements_MissingClientID(t *testing.T) {
	h := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients//measurements", nil)
	req = newChiContext(req, map[string]string{"clientID": ""})
	req = withAuthClaims(req,"coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ListMeasurements_MissingAuth(t *testing.T) {
	h := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ListMeasurements_Forbidden(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{
			err: &usecase.AuthorizationError{Message: "access denied"},
		},
		&mockGetLatestMeasurementsUC{},
		&mockGetWearableSummariesUC{},
		&mockGetProfileSummaryUC{},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = withAuthClaims(req,"coach-1", "coach")
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
	// Internal error details should not leak
	if errResp.Error == "access denied: no coach-client relationship" {
		t.Error("internal error details should not be leaked")
	}
}

func TestMeasurementHandler_ListMeasurements_ValidationError(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{
			err: &usecase.ValidationError{Message: "client_id is required"},
		},
		&mockGetLatestMeasurementsUC{},
		&mockGetWearableSummariesUC{},
		&mockGetProfileSummaryUC{},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = withAuthClaims(req,"coach-1", "coach")
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

func TestMeasurementHandler_ListMeasurements_InternalError(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{
			err: errors.New("database unavailable"),
		},
		&mockGetLatestMeasurementsUC{},
		&mockGetWearableSummariesUC{},
		&mockGetProfileSummaryUC{},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = withAuthClaims(req,"coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}

	var errResp ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Code != "INTERNAL_ERROR" {
		t.Errorf("expected code INTERNAL_ERROR, got %s", errResp.Code)
	}
	if errResp.Error == "database unavailable" {
		t.Error("internal error message should not be leaked to client")
	}
}

func TestMeasurementHandler_ListMeasurements_InvalidFromDate(t *testing.T) {
	h := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements?from=not-a-date", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = withAuthClaims(req,"coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ListMeasurements_InvalidToDate(t *testing.T) {
	h := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements?to=not-a-date", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = withAuthClaims(req,"coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ListMeasurements_WithTypeFilter(t *testing.T) {
	var capturedInput usecase.GetMeasurementsInput
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{
			output: &usecase.GetMeasurementsOutput{
				Measurements: []*entities.Measurement{},
				Pagination:   usecase.PaginationInfo{Page: 1, Limit: 20, Total: 0},
			},
		},
		&mockGetLatestMeasurementsUC{},
		&mockGetWearableSummariesUC{},
		&mockGetProfileSummaryUC{},
	)

	// Override with a capturing mock
	capturingMock := &capturingGetMeasurementsUC{}
	h.getMeasurementsUC = capturingMock

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements?type=weight", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = withAuthClaims(req,"coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ListMeasurements(rec, req)

	capturedInput = capturingMock.lastInput
	if capturedInput.MeasurementType != "weight" {
		t.Errorf("expected type filter 'weight', got %q", capturedInput.MeasurementType)
	}
}

type capturingGetMeasurementsUC struct {
	lastInput usecase.GetMeasurementsInput
}

func (m *capturingGetMeasurementsUC) Execute(ctx context.Context, input usecase.GetMeasurementsInput) (*usecase.GetMeasurementsOutput, error) {
	m.lastInput = input
	return &usecase.GetMeasurementsOutput{
		Measurements: []*entities.Measurement{},
		Pagination:   usecase.PaginationInfo{Page: 1, Limit: 20, Total: 0},
	}, nil
}

// Tests for LatestMeasurements

func TestMeasurementHandler_LatestMeasurements_Success(t *testing.T) {
	now := time.Now()
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{},
		&mockGetLatestMeasurementsUC{
			results: []*entities.LatestMeasurement{
				{MeasurementType: "weight", Value: 185.4, Unit: "lbs", MeasuredAt: now},
				{MeasurementType: "body_fat_pct", Value: 22.3, Unit: "%", MeasuredAt: now},
			},
		},
		&mockGetWearableSummariesUC{},
		&mockGetProfileSummaryUC{},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements/latest", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = withAuthClaims(req,"coach-1", "coach")
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
	if len(measurements) != 2 {
		t.Errorf("expected 2 measurements, got %d", len(measurements))
	}
}

func TestMeasurementHandler_LatestMeasurements_MissingClientID(t *testing.T) {
	h := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients//measurements/latest", nil)
	req = newChiContext(req, map[string]string{"clientID": ""})
	req = withAuthClaims(req,"coach-1", "coach")
	rec := httptest.NewRecorder()

	h.LatestMeasurements(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_LatestMeasurements_MissingAuth(t *testing.T) {
	h := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements/latest", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.LatestMeasurements(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestMeasurementHandler_LatestMeasurements_Forbidden(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{},
		&mockGetLatestMeasurementsUC{
			err: &usecase.AuthorizationError{Message: "access denied"},
		},
		&mockGetWearableSummariesUC{},
		&mockGetProfileSummaryUC{},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements/latest", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = withAuthClaims(req,"coach-1", "coach")
	rec := httptest.NewRecorder()

	h.LatestMeasurements(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

// Tests for ListWearableSummaries

func TestMeasurementHandler_ListWearableSummaries_Success(t *testing.T) {
	now := time.Now()
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{},
		&mockGetLatestMeasurementsUC{},
		&mockGetWearableSummariesUC{
			output: &usecase.GetWearableSummariesOutput{
				Summaries: []*entities.WearableSummary{
					{ID: "ws1", ClientID: "client-1", Source: "whoop", SummaryDate: now},
				},
				Pagination: usecase.PaginationInfo{Page: 1, Limit: 20, Total: 1},
			},
		},
		&mockGetProfileSummaryUC{},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/wearable-summaries", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = withAuthClaims(req,"coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ListWearableSummaries(rec, req)

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

func TestMeasurementHandler_ListWearableSummaries_MissingClientID(t *testing.T) {
	h := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients//wearable-summaries", nil)
	req = newChiContext(req, map[string]string{"clientID": ""})
	req = withAuthClaims(req,"coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ListWearableSummaries(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ListWearableSummaries_MissingAuth(t *testing.T) {
	h := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/wearable-summaries", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.ListWearableSummaries(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ListWearableSummaries_Forbidden(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{},
		&mockGetLatestMeasurementsUC{},
		&mockGetWearableSummariesUC{
			err: &usecase.AuthorizationError{Message: "access denied"},
		},
		&mockGetProfileSummaryUC{},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/wearable-summaries", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = withAuthClaims(req,"coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ListWearableSummaries(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ListWearableSummaries_InvalidFromDate(t *testing.T) {
	h := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/wearable-summaries?from=bad", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = withAuthClaims(req,"coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ListWearableSummaries(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

// Tests for ProfileSummary

func TestMeasurementHandler_ProfileSummary_Success(t *testing.T) {
	now := time.Now()
	hrv := 45.2
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{},
		&mockGetLatestMeasurementsUC{},
		&mockGetWearableSummariesUC{},
		&mockGetProfileSummaryUC{
			result: &entities.ProfileSummary{
				BodyComposition: map[string]*entities.LatestMeasurement{
					"weight": {MeasurementType: "weight", Value: 185.4, Unit: "lbs", MeasuredAt: now},
				},
				Labs: &entities.LabSummary{
					FlaggedCount: 0,
					Markers:      []entities.LatestMeasurement{},
				},
				Wearable: &entities.WearableAverages{
					Source:   "whoop",
					AvgHRV7d: &hrv,
				},
			},
		},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/profile-summary", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = withAuthClaims(req,"coach-1", "coach")
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
}

func TestMeasurementHandler_ProfileSummary_MissingClientID(t *testing.T) {
	h := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients//profile-summary", nil)
	req = newChiContext(req, map[string]string{"clientID": ""})
	req = withAuthClaims(req,"coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ProfileSummary(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ProfileSummary_MissingAuth(t *testing.T) {
	h := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/profile-summary", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.ProfileSummary(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ProfileSummary_Forbidden(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{},
		&mockGetLatestMeasurementsUC{},
		&mockGetWearableSummariesUC{},
		&mockGetProfileSummaryUC{
			err: &usecase.AuthorizationError{Message: "access denied"},
		},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/profile-summary", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = withAuthClaims(req,"coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ProfileSummary(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}
}

func TestMeasurementHandler_ProfileSummary_InternalError(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{},
		&mockGetLatestMeasurementsUC{},
		&mockGetWearableSummariesUC{},
		&mockGetProfileSummaryUC{
			err: errors.New("database unavailable"),
		},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/profile-summary", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	req = withAuthClaims(req,"coach-1", "coach")
	rec := httptest.NewRecorder()

	h.ProfileSummary(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}

	var errResp ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Code != "INTERNAL_ERROR" {
		t.Errorf("expected code INTERNAL_ERROR, got %s", errResp.Code)
	}
	if errResp.Error == "database unavailable" {
		t.Error("internal error message should not be leaked to client")
	}
}
