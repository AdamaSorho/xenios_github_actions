package usecase

import (
	"errors"
	"testing"
)

func TestAuthorizationError_Error_ReturnsMessage(t *testing.T) {
	err := &AuthorizationError{Message: "not authorized"}
	if err.Error() != "not authorized" {
		t.Errorf("expected 'not authorized', got %s", err.Error())
	}
}

func TestIsAuthorizationError_True(t *testing.T) {
	err := &AuthorizationError{Message: "forbidden"}
	if !IsAuthorizationError(err) {
		t.Error("expected true for AuthorizationError")
	}
}

func TestIsAuthorizationError_False(t *testing.T) {
	err := errors.New("some error")
	if IsAuthorizationError(err) {
		t.Error("expected false for non-AuthorizationError")
	}
}

func TestIsAuthorizationError_Nil(t *testing.T) {
	if IsAuthorizationError(nil) {
		t.Error("expected false for nil error")
	}
}
