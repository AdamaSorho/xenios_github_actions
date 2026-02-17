package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// PDFExtractor defines the interface for extracting structured data from PDF files.
type PDFExtractor interface {
	// ExtractInBody parses an InBody PDF and returns extracted measurements.
	// Returns an ExtractionResult which may be partial if some fields could not be extracted.
	ExtractInBody(ctx context.Context, pdfData []byte) (*entities.ExtractionResult, error)
}
