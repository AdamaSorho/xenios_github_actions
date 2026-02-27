package nutrition

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// CSVParser implements repository.NutritionParser using CSV format detection.
type CSVParser struct{}

// NewCSVParser creates a new CSVParser.
func NewCSVParser() *CSVParser {
	return &CSVParser{}
}

// Parse implements repository.NutritionParser.
func (p *CSVParser) Parse(data []byte) ([]entities.NutritionEntry, error) {
	return Parse(data)
}

var _ repository.NutritionParser = &CSVParser{}

// CSVFormat identifies the nutrition CSV format.
type CSVFormat string

const (
	FormatMFP     CSVFormat = "myfitnesspal"
	FormatGeneric CSVFormat = "generic"
	FormatUnknown CSVFormat = "unknown"
)

// DetectFormat inspects the CSV headers to determine the file format.
func DetectFormat(data []byte) CSVFormat {
	if len(data) == 0 {
		return FormatUnknown
	}

	reader := csv.NewReader(bytes.NewReader(data))
	headers, err := reader.Read()
	if err != nil {
		return FormatUnknown
	}

	normalized := normalizeHeaders(headers)

	if hasMFPHeaders(normalized) {
		return FormatMFP
	}
	if hasGenericHeaders(normalized) {
		return FormatGeneric
	}
	return FormatUnknown
}

// Parse reads CSV data, auto-detects the format, and returns parsed nutrition entries.
// Invalid rows are skipped silently; the rest are parsed.
func Parse(data []byte) ([]entities.NutritionEntry, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty CSV data")
	}

	format := DetectFormat(data)
	if format == FormatUnknown {
		return nil, fmt.Errorf("unrecognized CSV format")
	}

	reader := csv.NewReader(bytes.NewReader(data))
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read CSV headers: %w", err)
	}

	colMap := buildColumnMap(normalizeHeaders(headers))

	var entries []entities.NutritionEntry
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		entry, parseErr := parseRow(row, colMap, format)
		if parseErr != nil {
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// columnMap holds the column index for each field.
type columnMap struct {
	date     int
	meal     int
	calories int
	protein  int
	carbs    int
	fat      int
	fiber    int
}

func normalizeHeaders(headers []string) []string {
	normalized := make([]string, len(headers))
	for i, h := range headers {
		normalized[i] = strings.ToLower(strings.TrimSpace(h))
	}
	return normalized
}

func hasMFPHeaders(headers []string) bool {
	required := []string{"date", "meal", "calories"}
	return containsAll(headers, required)
}

func hasGenericHeaders(headers []string) bool {
	required := []string{"date", "calories", "protein", "carbs", "fat"}
	return containsAll(headers, required)
}

func containsAll(headers []string, required []string) bool {
	set := make(map[string]bool, len(headers))
	for _, h := range headers {
		set[h] = true
	}
	for _, r := range required {
		if !matchesAnyHeader(set, r) {
			return false
		}
	}
	return true
}

func matchesAnyHeader(set map[string]bool, target string) bool {
	if set[target] {
		return true
	}
	for h := range set {
		if strings.Contains(h, target) {
			return true
		}
	}
	return false
}

func buildColumnMap(headers []string) columnMap {
	cm := columnMap{date: -1, meal: -1, calories: -1, protein: -1, carbs: -1, fat: -1, fiber: -1}
	for i, h := range headers {
		switch {
		case h == "date":
			cm.date = i
		case h == "meal":
			cm.meal = i
		case strings.Contains(h, "calories") || strings.Contains(h, "calorie"):
			cm.calories = i
		case strings.Contains(h, "protein"):
			cm.protein = i
		case strings.Contains(h, "carbs") || strings.Contains(h, "carbohydrate"):
			cm.carbs = i
		case strings.Contains(h, "fat"):
			if !strings.Contains(h, "fiber") {
				cm.fat = i
			}
		case strings.Contains(h, "fiber"):
			cm.fiber = i
		}
	}
	return cm
}

func parseRow(row []string, cm columnMap, _ CSVFormat) (entities.NutritionEntry, error) {
	var entry entities.NutritionEntry

	if cm.date < 0 || cm.date >= len(row) {
		return entry, fmt.Errorf("missing date column")
	}

	date, err := parseDate(row[cm.date])
	if err != nil {
		return entry, fmt.Errorf("parse date: %w", err)
	}
	entry.Date = date

	if cm.meal >= 0 && cm.meal < len(row) {
		entry.Meal = strings.TrimSpace(row[cm.meal])
	}

	entry.Calories, err = parseFloatField(row, cm.calories)
	if err != nil {
		return entry, fmt.Errorf("parse calories: %w", err)
	}

	entry.Protein, _ = parseFloatField(row, cm.protein)
	entry.Carbs, _ = parseFloatField(row, cm.carbs)
	entry.Fat, _ = parseFloatField(row, cm.fat)
	entry.Fiber, _ = parseFloatField(row, cm.fiber)

	return entry, nil
}

func parseFloatField(row []string, col int) (float64, error) {
	if col < 0 || col >= len(row) {
		return 0, nil
	}
	val := strings.TrimSpace(row[col])
	if val == "" {
		return 0, nil
	}
	return strconv.ParseFloat(val, 64)
}

func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"1/2/2006",
		"2006/01/02",
	}
	for _, f := range formats {
		t, err := time.Parse(f, s)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized date format: %q", s)
}
