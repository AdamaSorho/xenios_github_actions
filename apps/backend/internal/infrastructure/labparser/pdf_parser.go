package labparser

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/xenios/backend/internal/domain/entities"
)

// FileFormat represents the detected file format.
type FileFormat string

const (
	FormatCSV     FileFormat = "csv"
	FormatPDF     FileFormat = "pdf"
	FormatUnknown FileFormat = "unknown"
)

// DetectFormat determines whether data is CSV or PDF based on content and filename.
func DetectFormat(data []byte, fileName string) FileFormat {
	// Check PDF magic bytes
	if len(data) >= 4 && string(data[:4]) == "%PDF" {
		return FormatPDF
	}

	// Check file extension
	lowerName := strings.ToLower(fileName)
	if strings.HasSuffix(lowerName, ".csv") {
		return FormatCSV
	}
	if strings.HasSuffix(lowerName, ".pdf") {
		return FormatPDF
	}

	return FormatUnknown
}

// labLinePattern matches lines like: "Glucose, Fasting  98  mg/dL  70-100"
// It looks for a known marker name followed by a numeric value, unit, and optional reference range.
var labLinePattern = regexp.MustCompile(
	`^(.+?)\s{2,}(\d+\.?\d*)\s+([\w/%]+)\s+(.+?)(?:\s+[HLhl])?$`,
)

// ParsePDFText parses lab results from extracted PDF text content.
// It scans each line for patterns matching "marker_name  value  unit  reference_range".
func ParsePDFText(data []byte) ([]entities.LabResult, error) {
	text := string(data)
	lines := strings.Split(text, "\n")

	var results []entities.LabResult

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "---") {
			continue
		}

		result, ok := parseTextLine(line)
		if ok {
			results = append(results, result)
		}
	}

	return results, nil
}

// parseTextLine tries to extract a lab result from a single text line.
func parseTextLine(line string) (entities.LabResult, bool) {
	matches := labLinePattern.FindStringSubmatch(line)
	if matches == nil {
		return entities.LabResult{}, false
	}

	rawName := strings.TrimSpace(matches[1])
	normalizedName := strings.ToLower(rawName)

	labType, known := entities.KnownLabMarkers[normalizedName]
	if !known {
		return entities.LabResult{}, false
	}

	value, err := strconv.ParseFloat(matches[2], 64)
	if err != nil {
		return entities.LabResult{}, false
	}

	unit := matches[3]
	refRangeStr := strings.TrimSpace(matches[4])

	refLow, refHigh := parseReferenceRange(refRangeStr)
	flag := entities.DetermineFlag(value, refLow, refHigh)

	return entities.LabResult{
		MarkerName:    string(labType),
		Value:         value,
		Unit:          unit,
		ReferenceLow:  refLow,
		ReferenceHigh: refHigh,
		Flag:          flag,
		RawName:       rawName,
	}, true
}
