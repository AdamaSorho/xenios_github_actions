package repository

import (
	"context"
	"fmt"
	"testing"
)

func TestGoPDFTextExtractor_EmptyData_ReturnsError(t *testing.T) {
	extractor := NewGoPDFTextExtractor()
	_, err := extractor.ExtractText(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil data")
	}

	_, err = extractor.ExtractText(context.Background(), []byte{})
	if err == nil {
		t.Error("expected error for empty data")
	}
}

func TestGoPDFTextExtractor_InvalidPDF_ReturnsError(t *testing.T) {
	extractor := NewGoPDFTextExtractor()
	_, err := extractor.ExtractText(context.Background(), []byte("not a pdf"))
	if err == nil {
		t.Error("expected error for invalid PDF data")
	}
}

func TestInMemoryPDFTextExtractor_ReturnsMappedText(t *testing.T) {
	extractor := NewInMemoryPDFTextExtractor()
	extractor.SetTextForContent("pdf-data", "Weight: 85 kg")

	text, err := extractor.ExtractText(context.Background(), []byte("pdf-data"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "Weight: 85 kg" {
		t.Errorf("expected 'Weight: 85 kg', got %q", text)
	}
}

func TestInMemoryPDFTextExtractor_DefaultText_ReturnsDefault(t *testing.T) {
	extractor := NewInMemoryPDFTextExtractor()
	extractor.DefaultText = "default text content"

	text, err := extractor.ExtractText(context.Background(), []byte("any-data"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "default text content" {
		t.Errorf("expected 'default text content', got %q", text)
	}
}

func TestInMemoryPDFTextExtractor_NoMapping_ReturnsError(t *testing.T) {
	extractor := NewInMemoryPDFTextExtractor()

	_, err := extractor.ExtractText(context.Background(), []byte("unmapped-data"))
	if err == nil {
		t.Error("expected error for unmapped data")
	}
}

func TestInMemoryPDFTextExtractor_ForceError_ReturnsError(t *testing.T) {
	extractor := NewInMemoryPDFTextExtractor()
	extractor.ForceError = fmt.Errorf("forced test error")

	_, err := extractor.ExtractText(context.Background(), []byte("any-data"))
	if err == nil {
		t.Error("expected forced error")
	}
}
