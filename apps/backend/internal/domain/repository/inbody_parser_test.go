package repository

import (
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// mockInBodyTextParser verifies the interface can be implemented.
type mockInBodyTextParser struct{}

func (m *mockInBodyTextParser) Parse(_, _, _, _ string, _ time.Time) *entities.ExtractionResult {
	return &entities.ExtractionResult{}
}

// Compile-time interface check.
var _ InBodyTextParser = &mockInBodyTextParser{}

func TestInBodyTextParser_InterfaceCompiles(t *testing.T) {
	var parser InBodyTextParser = &mockInBodyTextParser{}
	if parser == nil {
		t.Fatal("expected non-nil parser")
	}
}
