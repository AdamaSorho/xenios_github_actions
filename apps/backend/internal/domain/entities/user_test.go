package entities

import "testing"

func TestIsValidRole_Coach_ReturnsTrue(t *testing.T) {
	if !IsValidRole("coach") {
		t.Error("expected 'coach' to be a valid role")
	}
}

func TestIsValidRole_Client_ReturnsTrue(t *testing.T) {
	if !IsValidRole("client") {
		t.Error("expected 'client' to be a valid role")
	}
}

func TestIsValidRole_Admin_ReturnsTrue(t *testing.T) {
	if !IsValidRole("admin") {
		t.Error("expected 'admin' to be a valid role")
	}
}

func TestIsValidRole_Invalid_ReturnsFalse(t *testing.T) {
	if IsValidRole("superuser") {
		t.Error("expected 'superuser' to be an invalid role")
	}
}

func TestIsValidRole_Empty_ReturnsFalse(t *testing.T) {
	if IsValidRole("") {
		t.Error("expected empty string to be an invalid role")
	}
}
