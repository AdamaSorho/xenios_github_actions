package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTAuth_MissingAuthHeader(t *testing.T) {
	handler := JWTAuth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}

	var body map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body["code"] != "UNAUTHORIZED" {
		t.Errorf("expected code UNAUTHORIZED, got %v", body["code"])
	}
}

func TestJWTAuth_InvalidFormat(t *testing.T) {
	handler := JWTAuth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic abc123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestJWTAuth_EmptyToken(t *testing.T) {
	handler := JWTAuth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer ")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	handler := JWTAuth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestJWTAuth_ValidToken(t *testing.T) {
	secret := "test-secret-key"

	// Create a valid token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	var capturedClaims *UserClaims
	handler := JWTAuth(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedClaims = GetUserClaims(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if capturedClaims == nil {
		t.Fatal("expected non-nil user claims")
	}
	if capturedClaims.Subject != "user-123" {
		t.Errorf("expected subject user-123, got %s", capturedClaims.Subject)
	}
}

func TestJWTAuth_ValidToken_ExtractsRole(t *testing.T) {
	secret := "test-secret-key"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  "user-123",
		"role": "coach",
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	var capturedClaims *UserClaims
	handler := JWTAuth(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedClaims = GetUserClaims(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if capturedClaims == nil {
		t.Fatal("expected non-nil user claims")
	}
	if capturedClaims.Role != "coach" {
		t.Errorf("expected role 'coach', got '%s'", capturedClaims.Role)
	}
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	secret := "test-secret-key"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(-time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	handler := JWTAuth(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for expired token")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestJWTAuth_WrongSecret(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte("correct-secret"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	handler := JWTAuth("wrong-secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestJWTAuth_NoSecret_AcceptsToken(t *testing.T) {
	// When no secret is configured, tokens are accepted without verification (dev mode)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "dev-user",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte("any-key"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	var capturedClaims *UserClaims
	handler := JWTAuth("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedClaims = GetUserClaims(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if capturedClaims == nil {
		t.Fatal("expected non-nil claims")
	}
	if capturedClaims.Subject != "dev-user" {
		t.Errorf("expected subject dev-user, got %s", capturedClaims.Subject)
	}
}

func TestJWTAuth_NoSecret_ExtractsRole(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  "dev-user",
		"role": "admin",
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte("any-key"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	var capturedClaims *UserClaims
	handler := JWTAuth("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedClaims = GetUserClaims(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if capturedClaims == nil {
		t.Fatal("expected non-nil claims")
	}
	if capturedClaims.Role != "admin" {
		t.Errorf("expected role 'admin', got '%s'", capturedClaims.Role)
	}
}

func TestJWTAuth_BearerCaseInsensitive(t *testing.T) {
	secret := "test-secret"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user-1",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, _ := token.SignedString([]byte(secret))

	handler := JWTAuth(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "bearer "+tokenStr)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestGetUserClaims_EmptyContext(t *testing.T) {
	claims := GetUserClaims(context.Background())
	if claims != nil {
		t.Errorf("expected nil claims for empty context, got %v", claims)
	}
}

func TestSetUserClaims_RoundTrip(t *testing.T) {
	ctx := context.Background()
	expected := &UserClaims{Subject: "user-42", Role: "coach"}
	ctx = SetUserClaims(ctx, expected)

	got := GetUserClaims(ctx)
	if got == nil {
		t.Fatal("expected non-nil claims")
	}
	if got.Subject != "user-42" {
		t.Errorf("expected subject 'user-42', got '%s'", got.Subject)
	}
	if got.Role != "coach" {
		t.Errorf("expected role 'coach', got '%s'", got.Role)
	}
}

func TestJWTAuth_ResponseContentType(t *testing.T) {
	handler := JWTAuth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}
}
