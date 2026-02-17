package parser

import (
	"io"

	"github.com/xenios/backend/internal/domain/entities"
)

// WearableParser parses wearable export data and produces normalized measurements.
type WearableParser interface {
	// Source returns the wearable source identifier.
	Source() entities.WearableSource

	// Parse reads data from the reader and returns normalized measurements.
	Parse(reader io.Reader) ([]entities.WearableMeasurement, error)

	// DetectFormat checks if the header bytes match this parser's expected format.
	DetectFormat(header []byte) bool
}
