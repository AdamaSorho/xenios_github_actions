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
)

// CSVFormat represents the detected CSV format.
type CSVFormat string

const (
	FormatMyFitnessPal CSVFormat = "myfitnesspal"
	FormatGeneric      CSVFormat = "generic"
	FormatUnknown      CSVFormat = "unknown"
)

// CSVParser parses nutrition CSV files in various formats.
type CSVParser struct{}

// NewCSVParser creates a new CSVParser.
func NewCSVParser() *CSVParser {
	return &CSVParser{}
}

// Parse reads CSV data, detects its format, and returns aggregated daily nutrition logs.
func (p *CSVParser) Parse(data []byte) ([]entities.NutritionDailyLog, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty CSV data")
	}

	reader := csv.NewReader(bytes.NewReader(data))
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read CSV headers: %w", err)
	}

	normalized := normalizeHeaders(headers)
	format, colMap := detectFormat(normalized)
	if format == FormatUnknown {
		return nil, fmt.Errorf("unrecognized CSV format: headers %v", headers)
	}

	dailyTotals := make(map[string]*entities.NutritionDailyLog)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read CSV row: %w", err)
		}

		log, err := parseRow(record, colMap)
		if err != nil {
			continue // skip unparseable rows
		}

		dateKey := log.Date.Format("2006-01-02")
		if existing, ok := dailyTotals[dateKey]; ok {
			existing.Calories += log.Calories
			existing.Protein += log.Protein
			existing.Carbs += log.Carbs
			existing.Fat += log.Fat
			existing.Fiber += log.Fiber
		} else {
			entry := *log
			dailyTotals[dateKey] = &entry
		}
	}

	if len(dailyTotals) == 0 {
		return nil, fmt.Errorf("no valid nutrition data found in CSV")
	}

	result := make([]entities.NutritionDailyLog, 0, len(dailyTotals))
	for _, v := range dailyTotals {
		result = append(result, *v)
	}

	return result, nil
}

// columnMap holds the column indices for each nutrition field.
type columnMap struct {
	date     int
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

func detectFormat(headers []string) (CSVFormat, *columnMap) {
	cm := &columnMap{date: -1, calories: -1, protein: -1, carbs: -1, fat: -1, fiber: -1}

	for i, h := range headers {
		switch {
		case h == "date":
			cm.date = i
		case h == "calories" || h == "calories (kcal)":
			cm.calories = i
		case h == "protein" || h == "protein (g)":
			cm.protein = i
		case h == "carbs" || h == "carbohydrates (g)" || h == "carbs (g)" || h == "total carbohydrates (g)":
			cm.carbs = i
		case h == "fat" || h == "fat (g)" || h == "total fat (g)":
			cm.fat = i
		case h == "fiber" || h == "fiber (g)":
			cm.fiber = i
		}
	}

	if cm.date < 0 || cm.calories < 0 {
		return FormatUnknown, nil
	}

	// Determine format: MFP typically has "fat (g)" style headers
	isMFP := false
	for _, h := range headers {
		if strings.Contains(h, "(g)") || strings.Contains(h, "(kcal)") {
			isMFP = true
			break
		}
	}

	if isMFP {
		return FormatMyFitnessPal, cm
	}
	return FormatGeneric, cm
}

func parseRow(record []string, cm *columnMap) (*entities.NutritionDailyLog, error) {
	if cm.date >= len(record) || cm.calories >= len(record) {
		return nil, fmt.Errorf("row too short")
	}

	date, err := parseDate(record[cm.date])
	if err != nil {
		return nil, fmt.Errorf("parse date: %w", err)
	}

	calories, err := parseNumeric(safeGet(record, cm.calories))
	if err != nil {
		return nil, fmt.Errorf("parse calories: %w", err)
	}

	protein, _ := parseNumeric(safeGet(record, cm.protein))
	carbs, _ := parseNumeric(safeGet(record, cm.carbs))
	fat, _ := parseNumeric(safeGet(record, cm.fat))
	fiber, _ := parseNumeric(safeGet(record, cm.fiber))

	return &entities.NutritionDailyLog{
		Date:     date,
		Calories: calories,
		Protein:  protein,
		Carbs:    carbs,
		Fat:      fat,
		Fiber:    fiber,
	}, nil
}

func safeGet(record []string, idx int) string {
	if idx < 0 || idx >= len(record) {
		return ""
	}
	return record[idx]
}

var dateFormats = []string{
	"2006-01-02",
	"01/02/2006",
	"1/2/2006",
	"Jan 2, 2006",
	"January 2, 2006",
	"2006/01/02",
	"02-Jan-2006",
}

func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(strings.Trim(s, "\""))
	for _, format := range dateFormats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse date: %q", s)
}

// parseNumeric parses a string that may contain commas (e.g. "2,150") into a float64.
func parseNumeric(s string) (float64, error) {
	s = strings.TrimSpace(strings.Trim(s, "\""))
	if s == "" || s == "--" || s == "N/A" {
		return 0, nil
	}
	// Remove thousands separator commas
	s = strings.ReplaceAll(s, ",", "")
	return strconv.ParseFloat(s, 64)
}
