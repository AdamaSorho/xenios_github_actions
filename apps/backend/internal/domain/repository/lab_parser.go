package repository

import (
	"io"

	"github.com/xenios/backend/internal/domain/entities"
)

// LabParser defines the interface for parsing lab result files.
type LabParser interface {
	Parse(r io.Reader) ([]entities.LabMeasurement, error)
}
