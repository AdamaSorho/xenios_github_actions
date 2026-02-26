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
	domainrepo "github.com/xenios/backend/internal/domain/repository"
)

// CSVFormat represents the detected CSV format.
type CSVFormat string

const (
	FormatMyFitnessPal CSVFormat = "myfitnesspal"
	FormatGeneric      CSVFormat = "generic"
	FormatUnknown      CSVFormat = "unknown"
)

// ParseResult holds the result of parsing a nutrition CSV.
type ParseResult struct {
	Format      CSVFormat
	Rows        []entities.NutritionRow
	RowsParsed  int
	RowsSkipped int
	Errors      []string
}

// CSVParser implements the NutritionParser interface.
type CSVParser struct{}

// NewCSVParser creates a new CSVParser.
func NewCSVParser() *CSVParser {
	return &CSVParser{}
}

// Parse implements repository.NutritionParser.
func (p *CSVParser) Parse(data []byte) (*domainrepo.NutritionParseResult, error) {
	result, err := ParseCSV(data)
	if err != nil {
		return nil, err
	}
	return &domainrepo.NutritionParseResult{
		Format:      string(result.Format),
		Rows:        result.Rows,
		RowsParsed:  result.RowsParsed,
		RowsSkipped: result.RowsSkipped,
		Errors:      result.Errors,
	}, nil
}

// ParseCSV parses raw CSV bytes into nutrition rows. It auto-detects the format
// (MyFitnessPal or generic) based on the header row.
func ParseCSV(data []byte) (*ParseResult, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, fmt.Errorf("empty CSV data")
	}

	reader := csv.NewReader(bytes.NewReader(data))
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read CSV header: %w", err)
	}

	format := detectFormat(header)
	if format == FormatUnknown {
		return nil, fmt.Errorf("unrecognized CSV format: headers %v", header)
	}

	colMap := buildColumnMap(header)
	result := &ParseResult{Format: format}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.RowsSkipped++
			result.Errors = append(result.Errors, fmt.Sprintf("read error: %v", err))
			continue
		}

		row, parseErr := parseRow(format, colMap, record)
		if parseErr != nil {
			result.RowsSkipped++
			result.Errors = append(result.Errors, parseErr.Error())
			continue
		}

		result.Rows = append(result.Rows, *row)
		result.RowsParsed++
	}

	return result, nil
}

// detectFormat determines the CSV format from the header row.
func detectFormat(headers []string) CSVFormat {
	normalized := make(map[string]bool, len(headers))
	for _, h := range headers {
		normalized[strings.ToLower(strings.TrimSpace(h))] = true
	}

	// MFP format: Date, Meal, Calories, Fat (g), Protein (g), Carbs (g), Fiber (g)
	if normalized["date"] && normalized["meal"] && normalized["calories"] &&
		normalized["fat (g)"] && normalized["protein (g)"] && normalized["carbs (g)"] {
		return FormatMyFitnessPal
	}

	// Generic format: date, calories, protein, carbs, fat
	if normalized["date"] && normalized["calories"] && normalized["protein"] &&
		normalized["carbs"] && normalized["fat"] {
		return FormatGeneric
	}

	return FormatUnknown
}

// columnMap maps normalized column names to their index.
type columnMap map[string]int

func buildColumnMap(headers []string) columnMap {
	m := make(columnMap, len(headers))
	for i, h := range headers {
		m[strings.ToLower(strings.TrimSpace(h))] = i
	}
	return m
}

func (cm columnMap) get(record []string, key string) string {
	idx, ok := cm[key]
	if !ok || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}

func parseRow(format CSVFormat, cols columnMap, record []string) (*entities.NutritionRow, error) {
	dateStr := cols.get(record, "date")
	date, err := parseDate(dateStr)
	if err != nil {
		return nil, fmt.Errorf("row date %q: %w", dateStr, err)
	}

	row := &entities.NutritionRow{Date: date}

	switch format {
	case FormatMyFitnessPal:
		row.Meal = cols.get(record, "meal")
		row.Calories, err = parseFloat(cols.get(record, "calories"))
		if err != nil {
			return nil, fmt.Errorf("row calories: %w", err)
		}
		row.Fat, err = parseFloat(cols.get(record, "fat (g)"))
		if err != nil {
			return nil, fmt.Errorf("row fat: %w", err)
		}
		row.Protein, err = parseFloat(cols.get(record, "protein (g)"))
		if err != nil {
			return nil, fmt.Errorf("row protein: %w", err)
		}
		row.Carbs, err = parseFloat(cols.get(record, "carbs (g)"))
		if err != nil {
			return nil, fmt.Errorf("row carbs: %w", err)
		}
		fiberStr := cols.get(record, "fiber (g)")
		if fiberStr != "" {
			row.Fiber, err = parseFloat(fiberStr)
			if err != nil {
				return nil, fmt.Errorf("row fiber: %w", err)
			}
		}

	case FormatGeneric:
		row.Calories, err = parseFloat(cols.get(record, "calories"))
		if err != nil {
			return nil, fmt.Errorf("row calories: %w", err)
		}
		row.Protein, err = parseFloat(cols.get(record, "protein"))
		if err != nil {
			return nil, fmt.Errorf("row protein: %w", err)
		}
		row.Carbs, err = parseFloat(cols.get(record, "carbs"))
		if err != nil {
			return nil, fmt.Errorf("row carbs: %w", err)
		}
		row.Fat, err = parseFloat(cols.get(record, "fat"))
		if err != nil {
			return nil, fmt.Errorf("row fat: %w", err)
		}
		fiberStr := cols.get(record, "fiber")
		if fiberStr != "" {
			row.Fiber, err = parseFloat(fiberStr)
			if err != nil {
				return nil, fmt.Errorf("row fiber: %w", err)
			}
		}
	}

	return row, nil
}

func parseDate(s string) (time.Time, error) {
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

func parseFloat(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number %q: %w", s, err)
	}
	return v, nil
}
