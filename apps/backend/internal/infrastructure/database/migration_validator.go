// Package database provides database infrastructure utilities including migration validation.
package database

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// MigrationFile represents a parsed migration file.
type MigrationFile struct {
	Name      string
	Number    int
	Direction string // "up" or "down"
	Content   string
}

// MigrationValidator validates SQL migration files for correctness and best practices.
type MigrationValidator struct {
	dir   string
	files []MigrationFile
}

// ValidationError represents a migration validation failure.
type ValidationError struct {
	File    string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.File, e.Message)
}

// NewMigrationValidator creates a new validator for the given migrations directory.
func NewMigrationValidator(dir string) (*MigrationValidator, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	numberPattern := regexp.MustCompile(`^(\d+)_(.+)\.(up|down)\.sql$`)
	var files []MigrationFile

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		matches := numberPattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}

		var num int
		if _, err := fmt.Sscanf(matches[1], "%d", &num); err != nil {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", entry.Name(), err)
		}

		files = append(files, MigrationFile{
			Name:      entry.Name(),
			Number:    num,
			Direction: matches[3],
			Content:   string(content),
		})
	}

	return &MigrationValidator{dir: dir, files: files}, nil
}

// ValidateSequentialNumbering checks that migration numbers are sequential starting from 1.
func (v *MigrationValidator) ValidateSequentialNumbering() []ValidationError {
	numbers := make(map[int]bool)
	for _, f := range v.files {
		numbers[f.Number] = true
	}

	var sorted []int
	for n := range numbers {
		sorted = append(sorted, n)
	}
	sort.Ints(sorted)

	var errors []ValidationError
	for i, n := range sorted {
		expected := i + 1
		if n != expected {
			errors = append(errors, ValidationError{
				File:    fmt.Sprintf("migration_%06d", expected),
				Message: fmt.Sprintf("expected migration number %d, but found %d (gap in sequence)", expected, n),
			})
		}
	}
	return errors
}

// ValidateUpDownPairs checks that every .up.sql has a matching .down.sql and vice versa.
func (v *MigrationValidator) ValidateUpDownPairs() []ValidationError {
	ups := make(map[int]string)
	downs := make(map[int]string)

	for _, f := range v.files {
		if f.Direction == "up" {
			ups[f.Number] = f.Name
		} else {
			downs[f.Number] = f.Name
		}
	}

	var errors []ValidationError
	for num, name := range ups {
		if _, ok := downs[num]; !ok {
			errors = append(errors, ValidationError{
				File:    name,
				Message: "up migration has no matching down migration",
			})
		}
	}
	for num, name := range downs {
		if _, ok := ups[num]; !ok {
			errors = append(errors, ValidationError{
				File:    name,
				Message: "down migration has no matching up migration",
			})
		}
	}
	return errors
}

// ValidateIdempotency checks that CREATE TABLE/INDEX use IF NOT EXISTS and DROP uses IF EXISTS.
func (v *MigrationValidator) ValidateIdempotency() []ValidationError {
	var errors []ValidationError

	for _, f := range v.files {
		if f.Direction == "up" {
			if containsWithout(f.Content, "CREATE TABLE", "IF NOT EXISTS") {
				errors = append(errors, ValidationError{
					File:    f.Name,
					Message: "CREATE TABLE without IF NOT EXISTS",
				})
			}
			if containsWithout(f.Content, "CREATE INDEX", "IF NOT EXISTS") {
				errors = append(errors, ValidationError{
					File:    f.Name,
					Message: "CREATE INDEX without IF NOT EXISTS",
				})
			}
		} else {
			for _, stmt := range []string{"DROP TABLE", "DROP FUNCTION", "DROP TRIGGER", "DROP RULE", "DROP POLICY"} {
				if containsWithout(f.Content, stmt, "IF EXISTS") {
					errors = append(errors, ValidationError{
						File:    f.Name,
						Message: fmt.Sprintf("%s without IF EXISTS", stmt),
					})
				}
			}
		}
	}
	return errors
}

// ValidateTriggerGuards checks that all CREATE TRIGGER statements have DROP TRIGGER IF EXISTS guards.
func (v *MigrationValidator) ValidateTriggerGuards() []ValidationError {
	triggerPattern := regexp.MustCompile(`(?i)CREATE\s+TRIGGER\s+(\w+)\s+`)
	var errors []ValidationError

	for _, f := range v.files {
		if f.Direction != "up" {
			continue
		}
		matches := triggerPattern.FindAllStringSubmatch(f.Content, -1)
		for _, match := range matches {
			triggerName := strings.ToLower(match[1])
			dropPattern := fmt.Sprintf("drop trigger if exists %s", triggerName)
			if !strings.Contains(strings.ToLower(f.Content), dropPattern) {
				errors = append(errors, ValidationError{
					File:    f.Name,
					Message: fmt.Sprintf("CREATE TRIGGER %s lacks DROP TRIGGER IF EXISTS guard", triggerName),
				})
			}
		}
	}
	return errors
}

// ValidatePolicyGuards checks that all CREATE POLICY statements have DROP POLICY IF EXISTS guards.
func (v *MigrationValidator) ValidatePolicyGuards() []ValidationError {
	policyPattern := regexp.MustCompile(`(?i)CREATE\s+POLICY\s+(\w+)\s+ON\s+(\w+)`)
	var errors []ValidationError

	for _, f := range v.files {
		if f.Direction != "up" {
			continue
		}
		matches := policyPattern.FindAllStringSubmatch(f.Content, -1)
		for _, match := range matches {
			policyName := strings.ToLower(match[1])
			tableName := strings.ToLower(match[2])
			dropPattern := fmt.Sprintf("drop policy if exists %s on %s", policyName, tableName)
			if !strings.Contains(strings.ToLower(f.Content), dropPattern) {
				errors = append(errors, ValidationError{
					File:    f.Name,
					Message: fmt.Sprintf("CREATE POLICY %s ON %s lacks DROP POLICY IF EXISTS guard", policyName, tableName),
				})
			}
		}
	}
	return errors
}

// ValidateRLSEnabled checks that all tables have RLS enabled.
func (v *MigrationValidator) ValidateRLSEnabled() []ValidationError {
	tablePattern := regexp.MustCompile(`(?i)create\s+table\s+if\s+not\s+exists\s+(\w+)`)
	var errors []ValidationError

	for _, f := range v.files {
		if f.Direction != "up" {
			continue
		}
		sql := strings.ToLower(f.Content)
		matches := tablePattern.FindAllStringSubmatch(sql, -1)
		for _, match := range matches {
			tableName := match[1]
			rlsPattern := fmt.Sprintf("alter table %s enable row level security", tableName)
			if !strings.Contains(sql, rlsPattern) {
				errors = append(errors, ValidationError{
					File:    f.Name,
					Message: fmt.Sprintf("table '%s' does not have RLS enabled", tableName),
				})
			}
		}
	}
	return errors
}

// ValidateUUIDPrimaryKeys checks that all tables use UUID primary keys.
func (v *MigrationValidator) ValidateUUIDPrimaryKeys() []ValidationError {
	var errors []ValidationError

	for _, f := range v.files {
		if f.Direction != "up" {
			continue
		}
		sql := strings.ToLower(f.Content)
		if strings.Contains(sql, "create table") {
			if !strings.Contains(sql, "uuid primary key default gen_random_uuid()") {
				if !strings.Contains(f.Name, "rls_policies") {
					errors = append(errors, ValidationError{
						File:    f.Name,
						Message: "tables should use UUID PRIMARY KEY DEFAULT gen_random_uuid()",
					})
				}
			}
		}
	}
	return errors
}

// ValidateNoORM checks that no ORM patterns are present.
func (v *MigrationValidator) ValidateNoORM() []ValidationError {
	ormPatterns := []string{"sequelize", "typeorm", "prisma", "drizzle", "gorm"}
	var errors []ValidationError

	for _, f := range v.files {
		sql := strings.ToLower(f.Content)
		for _, pattern := range ormPatterns {
			if strings.Contains(sql, pattern) {
				errors = append(errors, ValidationError{
					File:    f.Name,
					Message: fmt.Sprintf("contains ORM pattern '%s' - must use raw SQL only", pattern),
				})
			}
		}
	}
	return errors
}

// ValidateAll runs all validation checks and returns all errors.
func (v *MigrationValidator) ValidateAll() []ValidationError {
	var allErrors []ValidationError
	allErrors = append(allErrors, v.ValidateSequentialNumbering()...)
	allErrors = append(allErrors, v.ValidateUpDownPairs()...)
	allErrors = append(allErrors, v.ValidateIdempotency()...)
	allErrors = append(allErrors, v.ValidateTriggerGuards()...)
	allErrors = append(allErrors, v.ValidatePolicyGuards()...)
	allErrors = append(allErrors, v.ValidateRLSEnabled()...)
	allErrors = append(allErrors, v.ValidateUUIDPrimaryKeys()...)
	allErrors = append(allErrors, v.ValidateNoORM()...)
	return allErrors
}

// stripSQLComments removes SQL line comments (-- ...) from the SQL string.
func stripSQLComments(sql string) string {
	var result strings.Builder
	for _, line := range strings.Split(sql, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		if idx := strings.Index(line, "--"); idx >= 0 {
			result.WriteString(line[:idx])
		} else {
			result.WriteString(line)
		}
		result.WriteString("\n")
	}
	return result.String()
}

// containsWithout checks if sql contains "keyword" not followed by "qualifier" (case-insensitive).
func containsWithout(sql, keyword, qualifier string) bool {
	lower := strings.ToLower(stripSQLComments(sql))
	kw := strings.ToLower(keyword)
	qual := strings.ToLower(qualifier)
	idx := 0
	for {
		pos := strings.Index(lower[idx:], kw)
		if pos == -1 {
			return false
		}
		absPos := idx + pos
		after := strings.TrimSpace(lower[absPos+len(kw):])
		if !strings.HasPrefix(after, qual) {
			return true
		}
		idx = absPos + len(kw)
	}
}
