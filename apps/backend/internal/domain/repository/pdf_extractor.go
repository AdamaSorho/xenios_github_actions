package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// PDFExtractor defines the interface for extracting InBody data from PDF content.
type PDFExtractor interface {
	// ExtractInBody parses PDF content and returns structured InBody results.
	ExtractInBody(ctx context.Context, pdfData []byte) (*entities.InBodyResult, error)
}
