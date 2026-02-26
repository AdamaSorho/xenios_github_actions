package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xenios/backend/internal/domain/entities"
)

// PDFLabParser extracts lab markers from simple PDF text content.
// It handles PDFs with tabular structure where text has been extracted to lines.
type PDFLabParser struct{}

// NewPDFLabParser creates a new PDFLabParser.
func NewPDFLabParser() *PDFLabParser {
	return &PDFLabParser{}
}

// Parse extracts lab markers from PDF content.
// For simple PDFs, it extracts text between BT/ET operators and parses lab marker patterns.
// Falls back to treating content as plain text lines if no PDF operators are found.
func (p *PDFLabParser) Parse(content []byte) ([]entities.ParsedMarker, error) {
	text := extractTextFromPDF(content)
	if text == "" {
		return nil, fmt.Errorf("parse pdf: no text content found")
	}

	markers := parseTextLines(text)
	if len(markers) == 0 {
		return nil, fmt.Errorf("parse pdf: no valid markers found in text")
	}

	return markers, nil
}

// extractTextFromPDF extracts readable text from PDF content.
// Handles two cases:
// 1. Raw PDF with text operators (Tj, TJ, etc.)
// 2. Pre-extracted text (plain text lines)
func extractTextFromPDF(content []byte) string {
	s := string(content)

	// Check if this looks like a binary PDF
	if strings.HasPrefix(s, "%PDF") {
		return extractFromPDFOperators(s)
	}

	// Treat as pre-extracted text
	return s
}

// extractFromPDFOperators extracts text from PDF text operators.
// Handles simple uncompressed PDF text streams with (text) Tj operators.
func extractFromPDFOperators(s string) string {
	var lines []string
	var currentLine strings.Builder

	i := 0
	for i < len(s) {
		// Look for text string objects: (text)
		if s[i] == '(' {
			i++
			var textBuilder strings.Builder
			depth := 1
			for i < len(s) && depth > 0 {
				if s[i] == '\\' && i+1 < len(s) {
					i++ // skip escaped character
					textBuilder.WriteByte(s[i])
				} else if s[i] == '(' {
					depth++
					textBuilder.WriteByte(s[i])
				} else if s[i] == ')' {
					depth--
					if depth > 0 {
						textBuilder.WriteByte(s[i])
					}
				} else {
					textBuilder.WriteByte(s[i])
				}
				i++
			}
			currentLine.WriteString(textBuilder.String())
			continue
		}

		// Look for newline indicators in PDF (BT = begin text, ET = end text)
		if i+2 <= len(s) && s[i] == 'E' && s[i+1] == 'T' {
			line := strings.TrimSpace(currentLine.String())
			if line != "" {
				lines = append(lines, line)
			}
			currentLine.Reset()
			i += 2
			continue
		}

		i++
	}

	// Flush remaining
	line := strings.TrimSpace(currentLine.String())
	if line != "" {
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// parseTextLines parses extracted text lines looking for lab marker patterns.
// Expects lines with format: MarkerName    Value    Unit    ReferenceRange
func parseTextLines(text string) []entities.ParsedMarker {
	lines := strings.Split(text, "\n")
	var markers []entities.ParsedMarker

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		marker := parseLabLine(line)
		if marker != nil {
			markers = append(markers, *marker)
		}
	}

	return markers
}

// parseLabLine attempts to extract a lab marker from a single text line.
// Tries multiple delimiters: tab, multiple spaces, pipe.
func parseLabLine(line string) *entities.ParsedMarker {
	// Try tab-delimited first
	if marker := tryDelimited(line, "\t"); marker != nil {
		return marker
	}

	// Try pipe-delimited
	if marker := tryDelimited(line, "|"); marker != nil {
		return marker
	}

	// Try whitespace-delimited (2+ spaces)
	if marker := tryWhitespaceDelimited(line); marker != nil {
		return marker
	}

	return nil
}

// tryDelimited parses a line using a specific delimiter.
func tryDelimited(line, delim string) *entities.ParsedMarker {
	parts := strings.Split(line, delim)
	if len(parts) < 2 {
		return nil
	}

	// Trim all parts
	trimmed := make([]string, len(parts))
	for i, p := range parts {
		trimmed[i] = strings.TrimSpace(p)
	}

	return buildMarkerFromParts(trimmed)
}

// tryWhitespaceDelimited splits on 2+ consecutive spaces.
func tryWhitespaceDelimited(line string) *entities.ParsedMarker {
	var parts []string
	current := strings.Builder{}

	prevSpace := false
	for _, ch := range line {
		if ch == ' ' {
			if prevSpace {
				s := strings.TrimSpace(current.String())
				if s != "" {
					parts = append(parts, s)
				}
				current.Reset()
				prevSpace = false
				continue
			}
			prevSpace = true
			current.WriteRune(ch)
		} else {
			prevSpace = false
			current.WriteRune(ch)
		}
	}

	s := strings.TrimSpace(current.String())
	if s != "" {
		parts = append(parts, s)
	}

	if len(parts) < 2 {
		return nil
	}

	return buildMarkerFromParts(parts)
}

// buildMarkerFromParts creates a ParsedMarker from ordered parts [name, value, unit?, refRange?].
func buildMarkerFromParts(parts []string) *entities.ParsedMarker {
	if len(parts) < 2 {
		return nil
	}

	name := parts[0]
	if name == "" {
		return nil
	}

	// Skip header-like rows
	nameLower := strings.ToLower(name)
	if nameLower == "test" || nameLower == "test name" || nameLower == "analyte" || nameLower == "component" || nameLower == "marker" {
		return nil
	}

	value, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil
	}

	unit := ""
	if len(parts) >= 3 {
		unit = parts[2]
	}

	var refLow, refHigh *float64
	if len(parts) >= 4 {
		refLow, refHigh = entities.ParseReferenceRange(parts[3])
	}

	return &entities.ParsedMarker{
		Name:          name,
		Value:         value,
		Unit:          unit,
		ReferenceLow:  refLow,
		ReferenceHigh: refHigh,
	}
}
