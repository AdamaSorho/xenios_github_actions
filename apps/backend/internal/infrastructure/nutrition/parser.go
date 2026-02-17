package nutrition

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// CSVFormat represents the detected format of a nutrition CSV file.
type CSVFormat int

const (
	FormatUnknown       CSVFormat = iota
	FormatMyFitnessPal
	FormatGeneric
)

// String returns a human-readable name for the CSV format.
func (f CSVFormat) String() string {
	switch f {
	case FormatMyFitnessPal:
		return "MyFitnessPal"
	case FormatGeneric:
		return "Generic"
	default:
		return "Unknown"
	}
}

// ParseResult holds the result of parsing a nutrition CSV file.
type ParseResult struct {
	Format      CSVFormat
	DailyTotals []entities.NutritionDailyTotal
	TotalRows   int
	SkippedRows int
}

// DetectFormat inspects a CSV header row and returns the detected format.
func DetectFormat(header []string) CSVFormat {
	normalized := make([]string, len(header))
	for i, h := range header {
		normalized[i] = strings.ToLower(strings.TrimSpace(h))
	}

	headerStr := strings.Join(normalized, ",")

	// MyFitnessPal: has "meal" column and "(g)" suffixed columns
	if strings.Contains(headerStr, "meal") && strings.Contains(headerStr, "fat (g)") {
		return FormatMyFitnessPal
	}

	// Generic: has date, calories, protein, carbs, fat columns
	hasDate := false
	hasCalories := false
	hasProtein := false
	hasCarbs := false
	hasFat := false
	for _, h := range normalized {
		switch h {
		case "date":
			hasDate = true
		case "calories":
			hasCalories = true
		case "protein":
			hasProtein = true
		case "carbs":
			hasCarbs = true
		case "fat":
			hasFat = true
		}
	}
	if hasDate && hasCalories && hasProtein && hasCarbs && hasFat {
		return FormatGeneric
	}

	return FormatUnknown
}

// ParseCSV reads and parses a nutrition CSV file, grouping by date and summing values.
func ParseCSV(r io.Reader) (*ParseResult, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read csv header: %w", err)
	}

	format := DetectFormat(header)
	if format == FormatUnknown {
		return nil, fmt.Errorf("unrecognized CSV format; header: %v", header)
	}

	colIndex := buildColumnIndex(header, format)

	dailyMap := make(map[time.Time]*entities.NutritionDailyTotal)
	totalRows := 0
	skippedRows := 0

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			skippedRows++
			totalRows++
			continue
		}
		totalRows++

		parsed, parseErr := parseRow(row, colIndex, format)
		if parseErr != nil {
			skippedRows++
			continue
		}

		day, ok := dailyMap[parsed.date]
		if !ok {
			day = &entities.NutritionDailyTotal{Date: parsed.date}
			dailyMap[parsed.date] = day
		}
		day.Calories += parsed.calories
		day.Protein += parsed.protein
		day.Carbs += parsed.carbs
		day.Fat += parsed.fat
		day.Fiber += parsed.fiber
	}

	// Sort by date
	dailyTotals := make([]entities.NutritionDailyTotal, 0, len(dailyMap))
	for _, d := range dailyMap {
		dailyTotals = append(dailyTotals, *d)
	}
	sort.Slice(dailyTotals, func(i, j int) bool {
		return dailyTotals[i].Date.Before(dailyTotals[j].Date)
	})

	return &ParseResult{
		Format:      format,
		DailyTotals: dailyTotals,
		TotalRows:   totalRows,
		SkippedRows: skippedRows,
	}, nil
}

// columnIndex maps logical fields to CSV column positions.
type columnIndex struct {
	date     int
	calories int
	protein  int
	carbs    int
	fat      int
	fiber    int
}

func buildColumnIndex(header []string, format CSVFormat) columnIndex {
	idx := columnIndex{date: -1, calories: -1, protein: -1, carbs: -1, fat: -1, fiber: -1}

	for i, h := range header {
		h = strings.ToLower(strings.TrimSpace(h))
		switch format {
		case FormatMyFitnessPal:
			switch h {
			case "date":
				idx.date = i
			case "calories":
				idx.calories = i
			case "protein (g)":
				idx.protein = i
			case "carbs (g)":
				idx.carbs = i
			case "fat (g)":
				idx.fat = i
			case "fiber (g)":
				idx.fiber = i
			}
		case FormatGeneric:
			switch h {
			case "date":
				idx.date = i
			case "calories":
				idx.calories = i
			case "protein":
				idx.protein = i
			case "carbs":
				idx.carbs = i
			case "fat":
				idx.fat = i
			case "fiber":
				idx.fiber = i
			}
		}
	}

	return idx
}

// parsedRow holds the parsed values from a single CSV row.
type parsedRow struct {
	date     time.Time
	calories float64
	protein  float64
	carbs    float64
	fat      float64
	fiber    float64
}

func parseRow(row []string, idx columnIndex, _ CSVFormat) (*parsedRow, error) {
	if idx.date < 0 || idx.date >= len(row) {
		return nil, fmt.Errorf("date column missing")
	}

	dateStr := strings.TrimSpace(row[idx.date])
	parsedDate, err := parseDate(dateStr)
	if err != nil {
		return nil, fmt.Errorf("parse date %q: %w", dateStr, err)
	}

	parsed := &parsedRow{date: parsedDate}

	parsed.calories, err = parseFloat(row, idx.calories)
	if err != nil {
		return nil, fmt.Errorf("parse calories: %w", err)
	}
	parsed.protein, err = parseFloat(row, idx.protein)
	if err != nil {
		return nil, fmt.Errorf("parse protein: %w", err)
	}
	parsed.carbs, err = parseFloat(row, idx.carbs)
	if err != nil {
		return nil, fmt.Errorf("parse carbs: %w", err)
	}
	parsed.fat, err = parseFloat(row, idx.fat)
	if err != nil {
		return nil, fmt.Errorf("parse fat: %w", err)
	}

	// Fiber is optional (may not be present in all formats)
	if idx.fiber >= 0 && idx.fiber < len(row) {
		parsed.fiber, _ = strconv.ParseFloat(strings.TrimSpace(row[idx.fiber]), 64)
	}

	return parsed, nil
}

func parseFloat(row []string, idx int) (float64, error) {
	if idx < 0 || idx >= len(row) {
		return 0, nil
	}
	s := strings.TrimSpace(row[idx])
	if s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
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
	return time.Time{}, fmt.Errorf("unrecognized date format: %s", s)
}

// ToNutritionRecords converts daily totals to individual NutritionRecord entries.
func ToNutritionRecords(dailyTotals []entities.NutritionDailyTotal, clientID, coachID, artifactID string) []*entities.NutritionRecord {
	var records []*entities.NutritionRecord
	for _, day := range dailyTotals {
		metrics := []struct {
			metricType entities.NutritionMetricType
			value      float64
		}{
			{entities.NutritionMetricCalories, day.Calories},
			{entities.NutritionMetricProtein, day.Protein},
			{entities.NutritionMetricCarbs, day.Carbs},
			{entities.NutritionMetricFat, day.Fat},
			{entities.NutritionMetricFiber, day.Fiber},
		}
		for _, m := range metrics {
			records = append(records, &entities.NutritionRecord{
				ClientID:   clientID,
				CoachID:    coachID,
				ArtifactID: artifactID,
				MetricType: m.metricType,
				Value:      m.value,
				Unit:       entities.NutritionMetricUnit[m.metricType],
				RecordDate: day.Date,
			})
		}
	}
	return records
}
