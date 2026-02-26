package repository

import (
	"io"

	"github.com/xenios/backend/internal/domain/entities"
)

// WearableParser extracts measurements from a wearable device export file.
type WearableParser interface {
	// Source returns the wearable source this parser handles.
	Source() entities.WearableSource
	// Parse reads the export file and returns normalized measurements.
	Parse(reader io.Reader, clientID, recordedBy string) ([]entities.Measurement, error)
	// DetectFormat returns true if the header bytes match this parser's expected format.
	DetectFormat(header []byte) bool
}
