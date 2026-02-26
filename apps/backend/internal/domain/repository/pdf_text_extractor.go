package repository

import "context"

// PDFTextExtractor defines the interface for extracting text content from PDF files.
type PDFTextExtractor interface {
	// ExtractText takes raw PDF bytes and returns the extracted text content.
	ExtractText(ctx context.Context, pdfData []byte) (string, error)
}
