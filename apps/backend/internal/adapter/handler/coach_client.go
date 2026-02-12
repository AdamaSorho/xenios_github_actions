package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

// CreateCoachClientUseCase defines the interface for the create coach-client use case.
type CreateCoachClientUseCase interface {
	Execute(ctx context.Context, coachID, clientID string) (*entities.CoachClient, error)
}

// ListCoachClientsUseCase defines the interface for the list coach-clients use case.
type ListCoachClientsUseCase interface {
	Execute(ctx context.Context, coachID string, limit, offset int) ([]*entities.CoachClient, error)
}

// CoachClientHandler handles HTTP requests for coach-client relationships.
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
	ClientID string `json:"client_id"`
}

// Create handles POST /api/v1/coaches/{coachID}/clients
func (h *CoachClientHandler) Create(w http.ResponseWriter, r *http.Request) {
	coachID := chi.URLParam(r, "coachID")
	if coachID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing coach ID", "INVALID_REQUEST", nil)
		return
	}

	var req CreateCoachClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErrorWithCode(w, http.StatusBadRequest, "invalid JSON body", "INVALID_JSON", nil)
		return
	}

	cc, err := h.createUC.Execute(r.Context(), coachID, req.ClientID)
	if err != nil {
		if usecase.IsValidationError(err) {
			respondErrorWithCode(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR", nil)
			return
		}
		respondErrorWithCode(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR", nil)
		return
	}

	_ = respondJSON(w, http.StatusCreated, cc)
}

// List handles GET /api/v1/coaches/{coachID}/clients
func (h *CoachClientHandler) List(w http.ResponseWriter, r *http.Request) {
	coachID := chi.URLParam(r, "coachID")
	if coachID == "" {
		respondErrorWithCode(w, http.StatusBadRequest, "missing coach ID", "INVALID_REQUEST", nil)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	results, err := h.listUC.Execute(r.Context(), coachID, limit, offset)
	if err != nil {
		if usecase.IsValidationError(err) {
			respondErrorWithCode(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR", nil)
			return
		}
		respondErrorWithCode(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR", nil)
		return
	}

	if results == nil {
		results = []*entities.CoachClient{}
	}

	_ = respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":   results,
		"limit":  limit,
		"offset": offset,
	})
}
