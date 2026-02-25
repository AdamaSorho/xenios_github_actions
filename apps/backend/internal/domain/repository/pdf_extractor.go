package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// PDFExtractor defines the interface for extracting structured data from PDF files.
type PDFExtractor interface {
	ExtractInBody(ctx context.Context, pdfContent []byte) (*entities.InBodyResult, error)
}
