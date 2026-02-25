package nutrition

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// ParseResult holds the output of parsing a nutrition CSV file.
type ParseResult struct {
	Format      entities.CSVFormat        `json:"format"`
	DailyTotals []entities.DailyNutrition `json:"daily_totals"`
	SkippedRows int                       `json:"skipped_rows"`
	TotalRows   int                       `json:"total_rows"`
}

// CSVNutritionParser implements the repository.NutritionParser interface.
type CSVNutritionParser struct{}

// NewCSVNutritionParser creates a new CSVNutritionParser.
func NewCSVNutritionParser() *CSVNutritionParser {
	return &CSVNutritionParser{}
}

// Verify interface compliance at compile time.
var _ repository.NutritionParser = (*CSVNutritionParser)(nil)

// Parse implements repository.NutritionParser.
func (p *CSVNutritionParser) Parse(data []byte) (*repository.NutritionParseResult, error) {
	result, err := Parse(data)
	if err != nil {
		return nil, err
	}
	return &repository.NutritionParseResult{
		Format:      result.Format,
		DailyTotals: result.DailyTotals,
		SkippedRows: result.SkippedRows,
		TotalRows:   result.TotalRows,
	}, nil
}

// ComputeAverages implements repository.NutritionParser.
func (p *CSVNutritionParser) ComputeAverages(dailyTotals []entities.DailyNutrition) entities.NutritionAverages {
	return ComputeAverages(dailyTotals)
}

// DailyTotalsToMeasurements implements repository.NutritionParser.
func (p *CSVNutritionParser) DailyTotalsToMeasurements(dailyTotals []entities.DailyNutrition, clientID, recordedBy string) []*entities.Measurement {
	return DailyTotalsToMeasurements(dailyTotals, clientID, recordedBy)
}

// DetectFormat examines the CSV header row to determine the file format.
func DetectFormat(data []byte) entities.CSVFormat {
	if len(data) == 0 {
		return entities.CSVFormatUnknown
	}

	reader := csv.NewReader(bytes.NewReader(data))
	header, err := reader.Read()
	if err != nil {
		return entities.CSVFormatUnknown
	}

	normalized := make([]string, len(header))
	for i, h := range header {
		normalized[i] = strings.ToLower(strings.TrimSpace(h))
	}

	if isMFPFormat(normalized) {
		return entities.CSVFormatMyFitnessPal
	}
	if isGenericFormat(normalized) {
		return entities.CSVFormatGeneric
	}
	return entities.CSVFormatUnknown
}

func isMFPFormat(headers []string) bool {
	required := map[string]bool{
		"date": false, "meal": false, "calories": false,
		"fat (g)": false, "protein (g)": false, "carbs (g)": false,
	}
	for _, h := range headers {
		if _, ok := required[h]; ok {
			required[h] = true
		}
	}
	for _, found := range required {
		if !found {
			return false
		}
	}
	return true
}

func isGenericFormat(headers []string) bool {
	required := map[string]bool{
		"date": false, "calories": false, "protein": false,
		"carbs": false, "fat": false,
	}
	for _, h := range headers {
		if _, ok := required[h]; ok {
			required[h] = true
		}
	}
	for _, found := range required {
		if !found {
			return false
		}
	}
	return true
}

// Parse reads a nutrition CSV and returns parsed daily totals.
func Parse(data []byte) (*ParseResult, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty CSV data")
	}

	format := DetectFormat(data)
	if format == entities.CSVFormatUnknown {
		return nil, fmt.Errorf("unrecognized CSV format")
	}

	switch format {
	case entities.CSVFormatMyFitnessPal:
		return parseMFP(data, format)
	case entities.CSVFormatGeneric:
		return parseGeneric(data, format)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func parseMFP(data []byte, format entities.CSVFormat) (*ParseResult, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	colIndex := buildColumnIndex(header)

	dateIdx := colIndex["date"]
	calIdx := colIndex["calories"]
	fatIdx := colIndex["fat (g)"]
	proteinIdx := colIndex["protein (g)"]
	carbsIdx := colIndex["carbs (g)"]
	fiberIdx := colIndex["fiber (g)"]

	dailyMap := make(map[string]*entities.DailyNutrition)
	skipped := 0
	total := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			skipped++
			continue
		}
		total++

		dateStr := strings.TrimSpace(record[dateIdx])
		date, err := parseDate(dateStr)
		if err != nil {
			skipped++
			continue
		}

		calories, err := parseFloat(record[calIdx])
		if err != nil {
			skipped++
			continue
		}
		fat, err := parseFloat(record[fatIdx])
		if err != nil {
			skipped++
			continue
		}
		protein, err := parseFloat(record[proteinIdx])
		if err != nil {
			skipped++
			continue
		}
		carbs, err := parseFloat(record[carbsIdx])
		if err != nil {
			skipped++
			continue
		}

		var fiber float64
		if fiberIdx >= 0 && fiberIdx < len(record) {
			fiber, _ = parseFloat(record[fiberIdx]) // fiber is optional
		}

		key := dateStr
		if existing, ok := dailyMap[key]; ok {
			existing.Calories += calories
			existing.Protein += protein
			existing.Carbs += carbs
			existing.Fat += fat
			existing.Fiber += fiber
		} else {
			dailyMap[key] = &entities.DailyNutrition{
				Date:     date,
				Calories: calories,
				Protein:  protein,
				Carbs:    carbs,
				Fat:      fat,
				Fiber:    fiber,
			}
		}
	}

	return buildResult(dailyMap, format, skipped, total), nil
}

func parseGeneric(data []byte, format entities.CSVFormat) (*ParseResult, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	colIndex := buildColumnIndex(header)

	dateIdx := colIndex["date"]
	calIdx := colIndex["calories"]
	proteinIdx := colIndex["protein"]
	carbsIdx := colIndex["carbs"]
	fatIdx := colIndex["fat"]
	fiberIdx := colIndex["fiber"]

	dailyMap := make(map[string]*entities.DailyNutrition)
	skipped := 0
	total := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			skipped++
			continue
		}
		total++

		dateStr := strings.TrimSpace(record[dateIdx])
		date, err := parseDate(dateStr)
		if err != nil {
			skipped++
			continue
		}

		calories, err := parseFloat(record[calIdx])
		if err != nil {
			skipped++
			continue
		}
		protein, err := parseFloat(record[proteinIdx])
		if err != nil {
			skipped++
			continue
		}
		carbs, err := parseFloat(record[carbsIdx])
		if err != nil {
			skipped++
			continue
		}
		fat, err := parseFloat(record[fatIdx])
		if err != nil {
			skipped++
			continue
		}

		var fiber float64
		if fiberIdx >= 0 && fiberIdx < len(record) {
			fiber, _ = parseFloat(record[fiberIdx])
		}

		key := dateStr
		if existing, ok := dailyMap[key]; ok {
			existing.Calories += calories
			existing.Protein += protein
			existing.Carbs += carbs
			existing.Fat += fat
			existing.Fiber += fiber
		} else {
			dailyMap[key] = &entities.DailyNutrition{
				Date:     date,
				Calories: calories,
				Protein:  protein,
				Carbs:    carbs,
				Fat:      fat,
				Fiber:    fiber,
			}
		}
	}

	return buildResult(dailyMap, format, skipped, total), nil
}

func buildColumnIndex(header []string) map[string]int {
	index := make(map[string]int)
	for i, h := range header {
		index[strings.ToLower(strings.TrimSpace(h))] = i
	}
	return index
}

func buildResult(dailyMap map[string]*entities.DailyNutrition, format entities.CSVFormat, skipped, total int) *ParseResult {
	dailyTotals := make([]entities.DailyNutrition, 0, len(dailyMap))
	for _, d := range dailyMap {
		dailyTotals = append(dailyTotals, *d)
	}
	sort.Slice(dailyTotals, func(i, j int) bool {
		return dailyTotals[i].Date.Before(dailyTotals[j].Date)
	})

	return &ParseResult{
		Format:      format,
		DailyTotals: dailyTotals,
		SkippedRows: skipped,
		TotalRows:   total,
	}
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

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}

// ComputeAverages calculates rolling averages for 7, 14, and 30-day periods.
// It uses the most recent N days from the daily totals (sorted by date).
func ComputeAverages(dailyTotals []entities.DailyNutrition) entities.NutritionAverages {
	n := len(dailyTotals)
	if n == 0 {
		return entities.NutritionAverages{}
	}

	// Sort by date (most recent last)
	sorted := make([]entities.DailyNutrition, len(dailyTotals))
	copy(sorted, dailyTotals)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Date.Before(sorted[j].Date)
	})

	avg7 := computePeriodAverage(sorted, 7)
	avg14 := computePeriodAverage(sorted, 14)
	avg30 := computePeriodAverage(sorted, 30)

	return entities.NutritionAverages{
		Avg7Day:  avg7,
		Avg14Day: avg14,
		Avg30Day: avg30,
	}
}

func computePeriodAverage(sorted []entities.DailyNutrition, period int) *entities.PeriodAverage {
	n := len(sorted)
	days := period
	if days > n {
		days = n
	}

	// Take the last `days` entries
	slice := sorted[n-days:]

	var sumCal, sumPro, sumCarb, sumFat, sumFib float64
	for _, d := range slice {
		sumCal += d.Calories
		sumPro += d.Protein
		sumCarb += d.Carbs
		sumFat += d.Fat
		sumFib += d.Fiber
	}

	count := float64(days)
	return &entities.PeriodAverage{
		Days:     days,
		Calories: roundTo(sumCal/count, 1),
		Protein:  roundTo(sumPro/count, 1),
		Carbs:    roundTo(sumCarb/count, 1),
		Fat:      roundTo(sumFat/count, 1),
		Fiber:    roundTo(sumFib/count, 1),
	}
}

func roundTo(val float64, places int) float64 {
	pow := math.Pow(10, float64(places))
	return math.Round(val*pow) / pow
}

// DailyTotalsToMeasurements converts daily nutrition totals into Measurement entities.
func DailyTotalsToMeasurements(dailyTotals []entities.DailyNutrition, clientID, recordedBy string) []*entities.Measurement {
	var measurements []*entities.Measurement
	for _, day := range dailyTotals {
		pairs := []struct {
			typ   entities.MeasurementType
			value float64
		}{
			{entities.MeasurementTypeCalories, day.Calories},
			{entities.MeasurementTypeProtein, day.Protein},
			{entities.MeasurementTypeCarbs, day.Carbs},
			{entities.MeasurementTypeFat, day.Fat},
			{entities.MeasurementTypeFiber, day.Fiber},
		}
		for _, p := range pairs {
			measurements = append(measurements, &entities.Measurement{
				ClientID:        clientID,
				RecordedBy:      recordedBy,
				MeasurementType: p.typ,
				Value:           p.value,
				Unit:            entities.MeasurementUnit(p.typ),
				MeasuredAt:      day.Date,
			})
		}
	}
	return measurements
}
