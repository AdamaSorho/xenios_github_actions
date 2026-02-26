package repository

import (
	"context"
	"testing"
)

// mockPDFTextExtractor verifies the interface can be implemented.
type mockPDFTextExtractor struct{}

func (m *mockPDFTextExtractor) ExtractText(_ context.Context, _ []byte) (string, error) {
	return "sample text", nil
}

// Compile-time interface check.
var _ PDFTextExtractor = &mockPDFTextExtractor{}

func TestPDFTextExtractor_InterfaceCompiles(t *testing.T) {
	var extractor PDFTextExtractor = &mockPDFTextExtractor{}
	if extractor == nil {
		t.Fatal("expected non-nil extractor")
	}
}
