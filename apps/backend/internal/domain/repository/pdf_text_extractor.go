package repository

import "context"

// PDFTextExtractor defines the interface for extracting text content from PDF data.
type PDFTextExtractor interface {
	ExtractText(ctx context.Context, pdfData []byte) (string, error)
}
