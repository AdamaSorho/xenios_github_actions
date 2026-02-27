package repository

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
)

// GoPDFTextExtractor extracts text from PDFs using the ledongthuc/pdf library.
type GoPDFTextExtractor struct{}

// NewGoPDFTextExtractor creates a new GoPDFTextExtractor.
func NewGoPDFTextExtractor() *GoPDFTextExtractor {
	return &GoPDFTextExtractor{}
}

// ExtractText extracts all text content from a PDF byte slice.
func (e *GoPDFTextExtractor) ExtractText(_ context.Context, pdfData []byte) (string, error) {
	if len(pdfData) == 0 {
		return "", fmt.Errorf("empty PDF data")
	}

	tmpFile, err := os.CreateTemp("", "inbody-*.pdf")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(pdfData); err != nil {
		return "", fmt.Errorf("write temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("close temp file: %w", err)
	}

	f, reader, err := pdf.Open(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("open PDF: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	for i := 1; i <= reader.NumPage(); i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}

	result := strings.TrimSpace(buf.String())
	if result == "" {
		return "", fmt.Errorf("no text content found in PDF")
	}

	return result, nil
}
