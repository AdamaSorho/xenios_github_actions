package parser

import (
	"fmt"
	"strings"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// Registry holds all registered wearable parsers and can detect formats.
type Registry struct {
	parsers []repository.WearableParser
}

// NewRegistry creates a parser registry with all supported parsers.
func NewRegistry() *Registry {
	return &Registry{
		parsers: []repository.WearableParser{
			NewWhoopParser(),
			NewGarminParser(),
			NewAppleHealthParser(),
			NewOuraParser(),
			NewFitbitParser(),
		},
	}
}

// DetectParser examines the header bytes and returns the matching parser.
func (r *Registry) DetectParser(header []byte) (repository.WearableParser, error) {
	for _, p := range r.parsers {
		if p.DetectFormat(header) {
			return p, nil
		}
	}
	return nil, fmt.Errorf("unrecognized wearable export format")
}

// GetParser returns the parser for a specific wearable source.
func (r *Registry) GetParser(source entities.WearableSource) (repository.WearableParser, error) {
	for _, p := range r.parsers {
		if p.Source() == source {
			return p, nil
		}
	}
	return nil, fmt.Errorf("no parser for source: %s", source)
}

// csvColumnIndex returns the index of the named column in the header row, or -1 if not found.
func csvColumnIndex(headers []string, name string) int {
	normalizedName := strings.TrimSpace(strings.ToLower(name))
	for i, h := range headers {
		if strings.TrimSpace(strings.ToLower(h)) == normalizedName {
			return i
		}
	}
	return -1
}
