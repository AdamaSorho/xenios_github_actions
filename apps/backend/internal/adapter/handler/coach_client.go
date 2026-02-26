package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/xenios/backend/internal/domain/entities"
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
	coachID := requireURLParam(w, r, "coachID", "coach ID")
	if coachID == "" {
		return
	}

	var req CreateCoachClientRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	cc, err := h.createUC.Execute(r.Context(), coachID, req.ClientID)
	if handleUseCaseError(w, err) {
		return
	}

	_ = respondJSON(w, http.StatusCreated, cc)
}

// List handles GET /api/v1/coaches/{coachID}/clients
func (h *CoachClientHandler) List(w http.ResponseWriter, r *http.Request) {
	coachID := requireURLParam(w, r, "coachID", "coach ID")
	if coachID == "" {
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	results, err := h.listUC.Execute(r.Context(), coachID, limit, offset)
	if handleUseCaseError(w, err) {
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
