package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/xenios/backend/internal/usecase"
)

// UserHandler handles HTTP requests for user operations.
type UserHandler struct {
	getUserUseCase    *usecase.GetUserUseCase
	createUserUseCase *usecase.CreateUserUseCase
}

// NewUserHandler creates a new UserHandler with injected use cases.
func NewUserHandler(
	getUserUseCase *usecase.GetUserUseCase,
	createUserUseCase *usecase.CreateUserUseCase,
) *UserHandler {
	return &UserHandler{
		getUserUseCase:    getUserUseCase,
		createUserUseCase: createUserUseCase,
	}
}

// UserResponse is the JSON response format for a user.
type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// CreateUserRequest is the JSON request format for creating a user.
type CreateUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// ErrorResponse is the JSON response format for errors.
type ErrorResponse struct {
	Error string `json:"error"`
}

// GetUser handles GET /api/users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.getUserUseCase.Execute(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to retrieve user")
		return
	}
	if user == nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respondJSON(w, http.StatusOK, UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// CreateUser handles POST /api/users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.createUserUseCase.Execute(r.Context(), usecase.CreateUserInput{
		Email: req.Email,
		Name:  req.Name,
	})
	if err != nil {
		switch err {
		case usecase.ErrEmailAlreadyExists:
			respondError(w, http.StatusConflict, "email already exists")
		case usecase.ErrInvalidEmail:
			respondError(w, http.StatusBadRequest, "invalid email")
		case usecase.ErrInvalidName:
			respondError(w, http.StatusBadRequest, "name is required")
		default:
			respondError(w, http.StatusInternalServerError, "failed to create user")
		}
		return
	}

	respondJSON(w, http.StatusCreated, UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}
