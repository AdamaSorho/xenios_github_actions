package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/xenios/backend/internal/adapter/middleware"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

// RegisterUseCase defines the interface for the register use case.
type RegisterUseCase interface {
	Execute(ctx context.Context, input usecase.RegisterInput) (*usecase.RegisterOutput, error)
}

// LoginUseCase defines the interface for the login use case.
type LoginUseCase interface {
	Execute(ctx context.Context, input usecase.LoginInput) (*usecase.LoginOutput, error)
}

// RefreshUseCase defines the interface for the refresh token use case.
type RefreshUseCase interface {
	Execute(ctx context.Context, refreshToken string) (*usecase.RefreshOutput, error)
}

// LogoutUseCase defines the interface for the logout use case.
type LogoutUseCase interface {
	Execute(ctx context.Context, userID string) error
}

// AuthHandler handles HTTP requests for authentication.
type AuthHandler struct {
	registerUC RegisterUseCase
	loginUC    LoginUseCase
	refreshUC  RefreshUseCase
	logoutUC   LogoutUseCase
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(registerUC RegisterUseCase, loginUC LoginUseCase, refreshUC RefreshUseCase, logoutUC LogoutUseCase) *AuthHandler {
	return &AuthHandler{
		registerUC: registerUC,
		loginUC:    loginUC,
		refreshUC:  refreshUC,
		logoutUC:   logoutUC,
	}
}

// RegisterRequest is the JSON request body for user registration.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

// LoginRequest is the JSON request body for user login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest is the JSON request body for token refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthResponse is the standard auth response with user and tokens.
type AuthResponse struct {
	User   *entities.User       `json:"user"`
	Tokens *entities.AuthTokens `json:"tokens"`
}

// Register handles POST /api/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErrorWithCode(w, http.StatusBadRequest, "invalid JSON body", "INVALID_JSON", nil)
		return
	}

	out, err := h.registerUC.Execute(r.Context(), usecase.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
		Role:     req.Role,
	})
	if err != nil {
		if usecase.IsValidationError(err) {
			respondErrorWithCode(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR", nil)
			return
		}
		respondErrorWithCode(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR", nil)
		return
	}

	_ = respondJSON(w, http.StatusCreated, AuthResponse{
		User:   out.User,
		Tokens: out.Tokens,
	})
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErrorWithCode(w, http.StatusBadRequest, "invalid JSON body", "INVALID_JSON", nil)
		return
	}

	out, err := h.loginUC.Execute(r.Context(), usecase.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if usecase.IsValidationError(err) {
			respondErrorWithCode(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR", nil)
			return
		}
		if usecase.IsAuthenticationError(err) {
			respondErrorWithCode(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED", nil)
			return
		}
		respondErrorWithCode(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR", nil)
		return
	}

	_ = respondJSON(w, http.StatusOK, AuthResponse{
		User:   out.User,
		Tokens: out.Tokens,
	})
}

// Refresh handles POST /api/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErrorWithCode(w, http.StatusBadRequest, "invalid JSON body", "INVALID_JSON", nil)
		return
	}

	out, err := h.refreshUC.Execute(r.Context(), req.RefreshToken)
	if err != nil {
		if usecase.IsValidationError(err) {
			respondErrorWithCode(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR", nil)
			return
		}
		if usecase.IsAuthenticationError(err) {
			respondErrorWithCode(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED", nil)
			return
		}
		respondErrorWithCode(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR", nil)
		return
	}

	_ = respondJSON(w, http.StatusOK, out.Tokens)
}

// Logout handles POST /api/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		respondErrorWithCode(w, http.StatusUnauthorized, "missing authentication", "UNAUTHORIZED", nil)
		return
	}

	err := h.logoutUC.Execute(r.Context(), claims.Subject)
	if err != nil {
		respondErrorWithCode(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR", nil)
		return
	}

	_ = respondJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}
