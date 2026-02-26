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

type mockGetMeasurementsUC struct {
	output *usecase.GetClientMeasurementsOutput
	err    error
}

func (m *mockGetMeasurementsUC) Execute(ctx context.Context, coachID string, filter entities.MeasurementFilter) (*usecase.GetClientMeasurementsOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.output != nil {
		return m.output, nil
	}
	return &usecase.GetClientMeasurementsOutput{
		Measurements: []*entities.Measurement{},
		Pagination:   usecase.PaginationOutput{Page: 1, Limit: 20, Total: 0},
	}, nil
}

type mockGetLatestMeasurementsUC struct {
	output *usecase.GetLatestMeasurementsOutput
	err    error
}

func (m *mockGetLatestMeasurementsUC) Execute(ctx context.Context, coachID, clientID string) (*usecase.GetLatestMeasurementsOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.output != nil {
		return m.output, nil
	}
	return &usecase.GetLatestMeasurementsOutput{Measurements: []*entities.Measurement{}}, nil
}

type mockGetWearableSummariesUC struct {
	output *usecase.GetWearableSummariesOutput
	err    error
}

func (m *mockGetWearableSummariesUC) Execute(ctx context.Context, coachID, clientID string, limit, offset int) (*usecase.GetWearableSummariesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.output != nil {
		return m.output, nil
	}
	return &usecase.GetWearableSummariesOutput{Summaries: []*entities.WearableSummary{}}, nil
}

type mockGetProfileSummaryUC struct {
	output *usecase.ProfileSummaryOutput
	err    error
}

func (m *mockGetProfileSummaryUC) Execute(ctx context.Context, coachID, clientID string) (*usecase.ProfileSummaryOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.output != nil {
		return m.output, nil
	}
	return &usecase.ProfileSummaryOutput{}, nil
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

func TestMeasurementHandler_GetMeasurements_Success(t *testing.T) {
	now := time.Now()
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{output: &usecase.GetClientMeasurementsOutput{
			Measurements: []*entities.Measurement{
				{ID: "m1", MeasurementType: "weight", Value: 185.4, Unit: "lbs", MeasuredAt: now},
			},
			Pagination: usecase.PaginationOutput{Page: 1, Limit: 20, Total: 1},
		}},
		&mockGetLatestMeasurementsUC{},
		&mockGetWearableSummariesUC{},
		&mockGetProfileSummaryUC{},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements?limit=20&page=1", nil)
	req = withAuthClaims(req, "coach-1", "coach")
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.GetMeasurements(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result usecase.GetClientMeasurementsOutput
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result.Measurements) != 1 {
		t.Errorf("expected 1 measurement, got %d", len(result.Measurements))
	}
}

func TestMeasurementHandler_GetMeasurements_MissingAuth(t *testing.T) {
	h := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.GetMeasurements(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestMeasurementHandler_GetMeasurements_MissingClientID(t *testing.T) {
	h := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients//measurements", nil)
	req = withAuthClaims(req, "coach-1", "coach")
	req = newChiContext(req, map[string]string{"clientID": ""})
	rec := httptest.NewRecorder()

	h.GetMeasurements(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_GetMeasurements_Forbidden(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{err: &usecase.AuthenticationError{Message: "forbidden"}},
		&mockGetLatestMeasurementsUC{},
		&mockGetWearableSummariesUC{},
		&mockGetProfileSummaryUC{},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements", nil)
	req = withAuthClaims(req, "coach-1", "coach")
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.GetMeasurements(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestMeasurementHandler_GetMeasurements_InternalError(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{err: errors.New("database error")},
		&mockGetLatestMeasurementsUC{},
		&mockGetWearableSummariesUC{},
		&mockGetProfileSummaryUC{},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements", nil)
	req = withAuthClaims(req, "coach-1", "coach")
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.GetMeasurements(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}

	var errResp ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Error == "database error" {
		t.Error("internal error message should not be leaked to client")
	}
}

func TestMeasurementHandler_GetLatestMeasurements_Success(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{},
		&mockGetLatestMeasurementsUC{output: &usecase.GetLatestMeasurementsOutput{
			Measurements: []*entities.Measurement{
				{ID: "m1", MeasurementType: "weight", Value: 185.4, Unit: "lbs"},
			},
		}},
		&mockGetWearableSummariesUC{},
		&mockGetProfileSummaryUC{},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements/latest", nil)
	req = withAuthClaims(req, "coach-1", "coach")
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.GetLatestMeasurements(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestMeasurementHandler_GetLatestMeasurements_MissingAuth(t *testing.T) {
	h := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements/latest", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.GetLatestMeasurements(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestMeasurementHandler_GetLatestMeasurements_MissingClientID(t *testing.T) {
	h := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients//measurements/latest", nil)
	req = withAuthClaims(req, "coach-1", "coach")
	req = newChiContext(req, map[string]string{"clientID": ""})
	rec := httptest.NewRecorder()

	h.GetLatestMeasurements(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_GetLatestMeasurements_Forbidden(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{},
		&mockGetLatestMeasurementsUC{err: &usecase.AuthenticationError{Message: "forbidden"}},
		&mockGetWearableSummariesUC{},
		&mockGetProfileSummaryUC{},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements/latest", nil)
	req = withAuthClaims(req, "coach-1", "coach")
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.GetLatestMeasurements(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestMeasurementHandler_GetWearableSummaries_Success(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{},
		&mockGetLatestMeasurementsUC{},
		&mockGetWearableSummariesUC{output: &usecase.GetWearableSummariesOutput{
			Summaries: []*entities.WearableSummary{
				{ID: "ws1", Source: "whoop", SummaryDate: "2026-01-15"},
			},
		}},
		&mockGetProfileSummaryUC{},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/wearable-summaries", nil)
	req = withAuthClaims(req, "coach-1", "coach")
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.GetWearableSummaries(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestMeasurementHandler_GetWearableSummaries_MissingAuth(t *testing.T) {
	h := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/wearable-summaries", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.GetWearableSummaries(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestMeasurementHandler_GetWearableSummaries_MissingClientID(t *testing.T) {
	h := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients//wearable-summaries", nil)
	req = withAuthClaims(req, "coach-1", "coach")
	req = newChiContext(req, map[string]string{"clientID": ""})
	rec := httptest.NewRecorder()

	h.GetWearableSummaries(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_GetWearableSummaries_Forbidden(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{},
		&mockGetLatestMeasurementsUC{},
		&mockGetWearableSummariesUC{err: &usecase.AuthenticationError{Message: "forbidden"}},
		&mockGetProfileSummaryUC{},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/wearable-summaries", nil)
	req = withAuthClaims(req, "coach-1", "coach")
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.GetWearableSummaries(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestMeasurementHandler_GetProfileSummary_Success(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{},
		&mockGetLatestMeasurementsUC{},
		&mockGetWearableSummariesUC{},
		&mockGetProfileSummaryUC{output: &usecase.ProfileSummaryOutput{
			BodyComposition: usecase.BodyCompositionSummary{
				Weight: &usecase.LatestValue{Value: 185.4, Unit: "lbs", Date: "2026-01-15"},
			},
		}},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/profile-summary", nil)
	req = withAuthClaims(req, "coach-1", "coach")
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.GetProfileSummary(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result usecase.ProfileSummaryOutput
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.BodyComposition.Weight == nil {
		t.Fatal("expected weight in response")
	}
	if result.BodyComposition.Weight.Value != 185.4 {
		t.Errorf("expected weight 185.4, got %f", result.BodyComposition.Weight.Value)
	}
}

func TestMeasurementHandler_GetProfileSummary_MissingAuth(t *testing.T) {
	h := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/profile-summary", nil)
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.GetProfileSummary(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestMeasurementHandler_GetProfileSummary_MissingClientID(t *testing.T) {
	h := newMeasurementHandler()

	req := httptest.NewRequest("GET", "/api/v1/clients//profile-summary", nil)
	req = withAuthClaims(req, "coach-1", "coach")
	req = newChiContext(req, map[string]string{"clientID": ""})
	rec := httptest.NewRecorder()

	h.GetProfileSummary(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMeasurementHandler_GetProfileSummary_Forbidden(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{},
		&mockGetLatestMeasurementsUC{},
		&mockGetWearableSummariesUC{},
		&mockGetProfileSummaryUC{err: &usecase.AuthenticationError{Message: "forbidden"}},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/profile-summary", nil)
	req = withAuthClaims(req, "coach-1", "coach")
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.GetProfileSummary(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestMeasurementHandler_GetProfileSummary_InternalError(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{},
		&mockGetLatestMeasurementsUC{},
		&mockGetWearableSummariesUC{},
		&mockGetProfileSummaryUC{err: errors.New("database error")},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/profile-summary", nil)
	req = withAuthClaims(req, "coach-1", "coach")
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.GetProfileSummary(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}

	var errResp ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Error == "database error" {
		t.Error("internal error message should not be leaked to client")
	}
}

func TestPageToOffset_DefaultPage(t *testing.T) {
	offset := pageToOffset("", 20)
	if offset != 0 {
		t.Errorf("expected 0, got %d", offset)
	}
}

func TestPageToOffset_Page2(t *testing.T) {
	offset := pageToOffset("2", 20)
	if offset != 20 {
		t.Errorf("expected 20, got %d", offset)
	}
}

func TestPageToOffset_Page3_Limit10(t *testing.T) {
	offset := pageToOffset("3", 10)
	if offset != 20 {
		t.Errorf("expected 20, got %d", offset)
	}
}

func TestPageToOffset_ZeroLimit_DefaultsTo20(t *testing.T) {
	offset := pageToOffset("2", 0)
	if offset != 20 {
		t.Errorf("expected 20, got %d", offset)
	}
}

func TestMeasurementHandler_GetMeasurements_ValidationError(t *testing.T) {
	h := NewMeasurementHandler(
		&mockGetMeasurementsUC{err: &usecase.ValidationError{Message: "client_id is required"}},
		&mockGetLatestMeasurementsUC{},
		&mockGetWearableSummariesUC{},
		&mockGetProfileSummaryUC{},
	)

	req := httptest.NewRequest("GET", "/api/v1/clients/client-1/measurements", nil)
	req = withAuthClaims(req, "coach-1", "coach")
	req = newChiContext(req, map[string]string{"clientID": "client-1"})
	rec := httptest.NewRecorder()

	h.GetMeasurements(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	var errResp ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR code, got %s", errResp.Code)
	}
}
