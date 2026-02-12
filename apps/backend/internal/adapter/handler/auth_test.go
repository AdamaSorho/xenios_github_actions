package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xenios/backend/internal/adapter/middleware"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/usecase"
)

// --- Mock use cases ---

type mockRegisterUC struct {
	output *usecase.RegisterOutput
	err    error
}

func (m *mockRegisterUC) Execute(_ context.Context, _ usecase.RegisterInput) (*usecase.RegisterOutput, error) {
	return m.output, m.err
}

type mockLoginUC struct {
	output *usecase.LoginOutput
	err    error
}

func (m *mockLoginUC) Execute(_ context.Context, _ usecase.LoginInput) (*usecase.LoginOutput, error) {
	return m.output, m.err
}

type mockRefreshUC struct {
	output *usecase.RefreshOutput
	err    error
}

func (m *mockRefreshUC) Execute(_ context.Context, _ string) (*usecase.RefreshOutput, error) {
	return m.output, m.err
}

type mockLogoutUC struct {
	err error
}

func (m *mockLogoutUC) Execute(_ context.Context, _ string) error {
	return m.err
}

func defaultAuthHandler() *AuthHandler {
	return NewAuthHandler(
		&mockRegisterUC{
			output: &usecase.RegisterOutput{
				User:   &entities.User{ID: "u1", Email: "test@example.com", Name: "Test", Role: "client"},
				Tokens: &entities.AuthTokens{AccessToken: "access-jwt", RefreshToken: "refresh-opaque"},
			},
		},
		&mockLoginUC{
			output: &usecase.LoginOutput{
				User:   &entities.User{ID: "u1", Email: "test@example.com", Name: "Test", Role: "client"},
				Tokens: &entities.AuthTokens{AccessToken: "access-jwt", RefreshToken: "refresh-opaque"},
			},
		},
		&mockRefreshUC{
			output: &usecase.RefreshOutput{
				Tokens: &entities.AuthTokens{AccessToken: "new-access", RefreshToken: "new-refresh"},
			},
		},
		&mockLogoutUC{},
	)
}

// --- Register tests ---

func TestAuthHandler_Register_Success(t *testing.T) {
	h := defaultAuthHandler()

	body, _ := json.Marshal(RegisterRequest{
		Email:    "test@example.com",
		Password: "securepassword",
		Name:     "Test",
		Role:     "client",
	})

	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	var resp AuthResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.User == nil {
		t.Fatal("expected user in response")
	}
	if resp.Tokens == nil {
		t.Fatal("expected tokens in response")
	}
	if resp.Tokens.AccessToken != "access-jwt" {
		t.Errorf("expected access token 'access-jwt', got '%s'", resp.Tokens.AccessToken)
	}
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	h := defaultAuthHandler()

	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader([]byte("{invalid")))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestAuthHandler_Register_ValidationError(t *testing.T) {
	h := NewAuthHandler(
		&mockRegisterUC{err: &usecase.ValidationError{Message: "email is required"}},
		&mockLoginUC{},
		&mockRefreshUC{},
		&mockLogoutUC{},
	)

	body, _ := json.Marshal(RegisterRequest{})
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code 'VALIDATION_ERROR', got '%s'", resp.Code)
	}
}

// --- Login tests ---

func TestAuthHandler_Login_Success(t *testing.T) {
	h := defaultAuthHandler()

	body, _ := json.Marshal(LoginRequest{
		Email:    "test@example.com",
		Password: "securepassword",
	})

	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp AuthResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.User == nil {
		t.Fatal("expected user in response")
	}
	if resp.Tokens == nil {
		t.Fatal("expected tokens in response")
	}
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	h := defaultAuthHandler()

	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader([]byte("{bad")))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	h := NewAuthHandler(
		&mockRegisterUC{},
		&mockLoginUC{err: &usecase.AuthenticationError{Message: "invalid credentials"}},
		&mockRefreshUC{},
		&mockLogoutUC{},
	)

	body, _ := json.Marshal(LoginRequest{Email: "a@b.com", Password: "wrong"})
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}

	var resp ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Code != "UNAUTHORIZED" {
		t.Errorf("expected code 'UNAUTHORIZED', got '%s'", resp.Code)
	}
}

func TestAuthHandler_Login_ValidationError(t *testing.T) {
	h := NewAuthHandler(
		&mockRegisterUC{},
		&mockLoginUC{err: &usecase.ValidationError{Message: "email is required"}},
		&mockRefreshUC{},
		&mockLogoutUC{},
	)

	body, _ := json.Marshal(LoginRequest{})
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

// --- Refresh tests ---

func TestAuthHandler_Refresh_Success(t *testing.T) {
	h := defaultAuthHandler()

	body, _ := json.Marshal(RefreshRequest{RefreshToken: "some-refresh-token"})
	req := httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Refresh(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp entities.AuthTokens
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.AccessToken != "new-access" {
		t.Errorf("expected access token 'new-access', got '%s'", resp.AccessToken)
	}
}

func TestAuthHandler_Refresh_InvalidJSON(t *testing.T) {
	h := defaultAuthHandler()

	req := httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewReader([]byte("{bad")))
	rec := httptest.NewRecorder()

	h.Refresh(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestAuthHandler_Refresh_InvalidToken(t *testing.T) {
	h := NewAuthHandler(
		&mockRegisterUC{},
		&mockLoginUC{},
		&mockRefreshUC{err: &usecase.AuthenticationError{Message: "invalid refresh token"}},
		&mockLogoutUC{},
	)

	body, _ := json.Marshal(RefreshRequest{RefreshToken: "invalid"})
	req := httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Refresh(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

// --- Logout tests ---

func TestAuthHandler_Logout_Success(t *testing.T) {
	h := defaultAuthHandler()

	req := httptest.NewRequest("POST", "/api/auth/logout", nil)
	// Add user claims to context (simulating authenticated request)
	ctx := middleware.SetUserClaims(req.Context(), &middleware.UserClaims{Subject: "user-123"})
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	h.Logout(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestAuthHandler_Logout_NoClaims(t *testing.T) {
	h := defaultAuthHandler()

	req := httptest.NewRequest("POST", "/api/auth/logout", nil)
	rec := httptest.NewRecorder()

	h.Logout(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestAuthHandler_Register_ContentType(t *testing.T) {
	h := defaultAuthHandler()

	body, _ := json.Marshal(RegisterRequest{
		Email:    "test@example.com",
		Password: "securepassword",
		Name:     "Test",
		Role:     "client",
	})

	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", ct)
	}
}
