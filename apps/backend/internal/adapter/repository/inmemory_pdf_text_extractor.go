package repository

import (
	"context"
	"fmt"
)

// InMemoryPDFTextExtractor is a test implementation of PDFTextExtractor.
// It returns pre-configured text for given PDF data, allowing tests to control the extracted text.
type InMemoryPDFTextExtractor struct {
	// TextByContent maps raw PDF content (as string key) to extracted text.
	TextByContent map[string]string
	// DefaultText is returned when no specific mapping exists. If empty, returns an error.
	DefaultText string
	// ForceError causes ExtractText to always return this error.
	ForceError error
}

// NewInMemoryPDFTextExtractor creates a new InMemoryPDFTextExtractor.
func NewInMemoryPDFTextExtractor() *InMemoryPDFTextExtractor {
	return &InMemoryPDFTextExtractor{
		TextByContent: make(map[string]string),
	}
}

// ExtractText returns pre-configured text or an error.
func (e *InMemoryPDFTextExtractor) ExtractText(_ context.Context, pdfData []byte) (string, error) {
	if e.ForceError != nil {
		return "", e.ForceError
	}

	key := string(pdfData)
	if text, ok := e.TextByContent[key]; ok {
		return text, nil
	}

	if e.DefaultText != "" {
		return e.DefaultText, nil
	}

	return "", fmt.Errorf("no text mapping for PDF data")
}

// SetTextForContent maps specific PDF content to extracted text (test helper).
func (e *InMemoryPDFTextExtractor) SetTextForContent(pdfContent string, extractedText string) {
	e.TextByContent[pdfContent] = extractedText
}
