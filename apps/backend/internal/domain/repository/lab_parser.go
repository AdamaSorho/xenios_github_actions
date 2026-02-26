package repository

import "github.com/xenios/backend/internal/domain/entities"

// LabFileParser extracts lab markers from raw file content.
type LabFileParser interface {
	Parse(content []byte) ([]entities.ParsedMarker, error)
}
