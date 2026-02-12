package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/xenios/backend/internal/domain"
	"github.com/xenios/backend/internal/usecase"
)

// CreateCoachClientUseCase defines the interface for the create coach-client use case.
type CreateCoachClientUseCase interface {
	Execute(ctx context.Context, input usecase.CreateCoachClientInput) (*domain.CoachClient, error)
}

// ListCoachClientsUseCase defines the interface for the list coach-clients use case.
type ListCoachClientsUseCase interface {
	Execute(ctx context.Context, input usecase.ListCoachClientsInput) ([]*domain.CoachClient, error)
}

// CoachClientHandler handles HTTP requests for coach-client management.
type CoachClientHandler struct {
	createUC CreateCoachClientUseCase
	listUC   ListCoachClientsUseCase
}

// NewCoachClientHandler creates a new CoachClientHandler.
func NewCoachClientHandler(createUC CreateCoachClientUseCase, listUC ListCoachClientsUseCase) *CoachClientHandler {
	return &CoachClientHandler{
		createUC: createUC,
		listUC:   listUC,
	}
}

// CreateCoachClientRequest is the JSON request body for creating a coach-client relationship.
type CreateCoachClientRequest struct {
	CoachID  string `json:"coach_id"`
	ClientID string `json:"client_id"`
}

// CreateCoachClientResponse is the JSON response for a created coach-client relationship.
type CreateCoachClientResponse struct {
	Data *domain.CoachClient `json:"data"`
}

// ListCoachClientsResponse is the JSON response for listing coach clients.
type ListCoachClientsResponse struct {
	Data []*domain.CoachClient `json:"data"`
}

// Create handles POST /api/v1/coaches/{coachID}/clients
func (h *CoachClientHandler) Create(w http.ResponseWriter, r *http.Request) {
	coachID := chi.URLParam(r, "coachID")
	if coachID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "coach ID is required", "INVALID_REQUEST", nil)
		return
	}

	var req CreateCoachClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErrorWithCode(w, http.StatusBadRequest, "invalid JSON body", "INVALID_JSON", map[string]interface{}{
			"parse_error": err.Error(),
		})
		return
	}

	input := usecase.CreateCoachClientInput{
		CoachID:  coachID,
		ClientID: req.ClientID,
	}

	result, err := h.createUC.Execute(r.Context(), input)
	if err != nil {
		respondErrorWithCode(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR", nil)
		return
	}

	respondJSON(w, http.StatusCreated, CreateCoachClientResponse{Data: result})
}

// List handles GET /api/v1/coaches/{coachID}/clients
func (h *CoachClientHandler) List(w http.ResponseWriter, r *http.Request) {
	coachID := chi.URLParam(r, "coachID")
	if coachID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "coach ID is required", "INVALID_REQUEST", nil)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	input := usecase.ListCoachClientsInput{
		CoachID: coachID,
		Limit:   limit,
		Offset:  offset,
	}

	results, err := h.listUC.Execute(r.Context(), input)
	if err != nil {
		respondErrorWithCode(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR", nil)
		return
	}

	if results == nil {
		results = []*domain.CoachClient{}
	}

	respondJSON(w, http.StatusOK, ListCoachClientsResponse{Data: results})
}
