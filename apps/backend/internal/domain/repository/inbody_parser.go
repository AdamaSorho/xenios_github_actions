package repository

import (
	"github.com/xenios/backend/internal/domain/entities"
	"time"
)

// InBodyTextParser defines the interface for parsing InBody metrics from text.
type InBodyTextParser interface {
	// Parse extracts InBody body composition metrics from raw text.
	Parse(text, clientID, recordedBy, artifactID string, measuredAt time.Time) *entities.ExtractionResult
}
