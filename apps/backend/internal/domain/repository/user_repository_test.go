package repository

import (
	"errors"
	"testing"
)

func TestErrDuplicateEmail_IsError(t *testing.T) {
	if ErrDuplicateEmail == nil {
		t.Fatal("ErrDuplicateEmail should not be nil")
	}

	if ErrDuplicateEmail.Error() != "email already exists" {
		t.Errorf("expected error message 'email already exists', got %q", ErrDuplicateEmail.Error())
	}
}

func TestErrDuplicateEmail_CanBeWrapped(t *testing.T) {
	wrapped := errors.New("wrapped: " + ErrDuplicateEmail.Error())
	if wrapped == nil {
		t.Fatal("wrapped error should not be nil")
	}
}

func TestErrDuplicateEmail_ErrorsIs(t *testing.T) {
	err := ErrDuplicateEmail
	if !errors.Is(err, ErrDuplicateEmail) {
		t.Error("errors.Is should match ErrDuplicateEmail")
	}
}
