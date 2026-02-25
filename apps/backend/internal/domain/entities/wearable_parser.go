package entities

import "io"

// WearableParser defines the contract for parsing wearable data exports.
// Implementations live in the parser package; the interface lives here
// so that the use-case layer can depend on it without importing outer layers.
type WearableParser interface {
	// Source returns the wearable source identifier this parser handles.
	Source() WearableSource

	// Parse reads raw export data and returns normalised measurements.
	// The clientID is attached to every returned measurement.
	Parse(reader io.Reader, clientID string) ([]Measurement, error)

	// DetectFormat inspects the first bytes of a file and returns true
	// if this parser can handle the format.
	DetectFormat(header []byte) bool
}
