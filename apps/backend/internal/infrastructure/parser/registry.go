package parser

import "github.com/xenios/backend/internal/domain/entities"

// Registry holds all available wearable parsers and provides source detection.
type Registry struct {
	parsers []entities.WearableParser
}

// NewRegistry creates a Registry pre-loaded with all supported parsers.
func NewRegistry() *Registry {
	return &Registry{
		parsers: []entities.WearableParser{
			NewWhoopParser(),
			NewGarminParser(),
			NewAppleHealthParser(),
			NewOuraParser(),
			NewFitbitParser(),
		},
	}
}

// Detect inspects the header bytes and returns the first matching parser, or nil.
func (r *Registry) Detect(header []byte) entities.WearableParser {
	for _, p := range r.parsers {
		if p.DetectFormat(header) {
			return p
		}
	}
	return nil
}

// ForSource returns the parser for the given source, or nil.
func (r *Registry) ForSource(src entities.WearableSource) entities.WearableParser {
	for _, p := range r.parsers {
		if p.Source() == src {
			return p
		}
	}
	return nil
}
