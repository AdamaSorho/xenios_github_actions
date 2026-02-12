package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewMigrationValidator_ValidDir verifies that NewMigrationValidator works with the migrations dir.
func TestNewMigrationValidator_ValidDir(t *testing.T) {
	migrationsDir := findMigrationsDir(t)

	v, err := NewMigrationValidator(migrationsDir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if v == nil {
		t.Fatal("expected non-nil validator")
	}
	if len(v.files) == 0 {
		t.Error("expected migration files to be loaded")
	}
}

// TestNewMigrationValidator_InvalidDir verifies that NewMigrationValidator returns an error for invalid dir.
func TestNewMigrationValidator_InvalidDir(t *testing.T) {
	_, err := NewMigrationValidator("/nonexistent/path")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

// TestNewMigrationValidator_EmptyDir verifies that NewMigrationValidator works with an empty dir.
func TestNewMigrationValidator_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	v, err := NewMigrationValidator(tmpDir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(v.files) != 0 {
		t.Errorf("expected 0 files, got %d", len(v.files))
	}
}

// TestMigrationValidator_ValidateSequentialNumbering verifies sequential numbering detection.
func TestMigrationValidator_ValidateSequentialNumbering(t *testing.T) {
	migrationsDir := findMigrationsDir(t)

	v, err := NewMigrationValidator(migrationsDir)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	errors := v.ValidateSequentialNumbering()
	if len(errors) > 0 {
		for _, e := range errors {
			t.Errorf("sequential numbering error: %s", e.Error())
		}
	}
}

// TestMigrationValidator_ValidateSequentialNumbering_Gap verifies gap detection.
func TestMigrationValidator_ValidateSequentialNumbering_Gap(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with a gap (1, 3 - missing 2)
	writeFile(t, tmpDir, "000001_test.up.sql", "CREATE TABLE IF NOT EXISTS t1 (id UUID PRIMARY KEY);")
	writeFile(t, tmpDir, "000001_test.down.sql", "DROP TABLE IF EXISTS t1;")
	writeFile(t, tmpDir, "000003_test.up.sql", "CREATE TABLE IF NOT EXISTS t3 (id UUID PRIMARY KEY);")
	writeFile(t, tmpDir, "000003_test.down.sql", "DROP TABLE IF EXISTS t3;")

	v, err := NewMigrationValidator(tmpDir)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	errors := v.ValidateSequentialNumbering()
	if len(errors) == 0 {
		t.Error("expected errors for gap in numbering")
	}
}

// TestMigrationValidator_ValidateUpDownPairs verifies up/down pair detection.
func TestMigrationValidator_ValidateUpDownPairs(t *testing.T) {
	migrationsDir := findMigrationsDir(t)

	v, err := NewMigrationValidator(migrationsDir)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	errors := v.ValidateUpDownPairs()
	if len(errors) > 0 {
		for _, e := range errors {
			t.Errorf("up/down pair error: %s", e.Error())
		}
	}
}

// TestMigrationValidator_ValidateUpDownPairs_Missing verifies detection of missing pairs.
func TestMigrationValidator_ValidateUpDownPairs_Missing(t *testing.T) {
	tmpDir := t.TempDir()

	// Create only an up file (missing down)
	writeFile(t, tmpDir, "000001_test.up.sql", "CREATE TABLE IF NOT EXISTS t1 (id UUID PRIMARY KEY);")

	v, err := NewMigrationValidator(tmpDir)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	errors := v.ValidateUpDownPairs()
	if len(errors) == 0 {
		t.Error("expected errors for missing down migration")
	}
}

// TestMigrationValidator_ValidateIdempotency verifies idempotency checks against real migrations.
func TestMigrationValidator_ValidateIdempotency(t *testing.T) {
	migrationsDir := findMigrationsDir(t)

	v, err := NewMigrationValidator(migrationsDir)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	errors := v.ValidateIdempotency()
	if len(errors) > 0 {
		for _, e := range errors {
			t.Errorf("idempotency error: %s", e.Error())
		}
	}
}

// TestMigrationValidator_ValidateIdempotency_Fail verifies detection of non-idempotent SQL.
func TestMigrationValidator_ValidateIdempotency_Fail(t *testing.T) {
	tmpDir := t.TempDir()

	writeFile(t, tmpDir, "000001_test.up.sql", "CREATE TABLE users (id UUID PRIMARY KEY);")
	writeFile(t, tmpDir, "000001_test.down.sql", "DROP TABLE users;")

	v, err := NewMigrationValidator(tmpDir)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	errors := v.ValidateIdempotency()
	if len(errors) == 0 {
		t.Error("expected errors for non-idempotent SQL")
	}
}

// TestMigrationValidator_ValidateTriggerGuards verifies trigger guard checks.
func TestMigrationValidator_ValidateTriggerGuards(t *testing.T) {
	migrationsDir := findMigrationsDir(t)

	v, err := NewMigrationValidator(migrationsDir)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	errors := v.ValidateTriggerGuards()
	if len(errors) > 0 {
		for _, e := range errors {
			t.Errorf("trigger guard error: %s", e.Error())
		}
	}
}

// TestMigrationValidator_ValidateTriggerGuards_Missing verifies detection of missing guards.
func TestMigrationValidator_ValidateTriggerGuards_Missing(t *testing.T) {
	tmpDir := t.TempDir()

	// Trigger without DROP guard
	writeFile(t, tmpDir, "000001_test.up.sql",
		"CREATE TABLE IF NOT EXISTS t1 (id UUID PRIMARY KEY);\nCREATE TRIGGER test_trigger BEFORE UPDATE ON t1 FOR EACH ROW EXECUTE FUNCTION update_at();")
	writeFile(t, tmpDir, "000001_test.down.sql", "DROP TABLE IF EXISTS t1;")

	v, err := NewMigrationValidator(tmpDir)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	errors := v.ValidateTriggerGuards()
	if len(errors) == 0 {
		t.Error("expected errors for trigger without DROP guard")
	}
}

// TestMigrationValidator_ValidatePolicyGuards verifies policy guard checks.
func TestMigrationValidator_ValidatePolicyGuards(t *testing.T) {
	migrationsDir := findMigrationsDir(t)

	v, err := NewMigrationValidator(migrationsDir)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	errors := v.ValidatePolicyGuards()
	if len(errors) > 0 {
		for _, e := range errors {
			t.Errorf("policy guard error: %s", e.Error())
		}
	}
}

// TestMigrationValidator_ValidateRLSEnabled verifies RLS checks against real migrations.
func TestMigrationValidator_ValidateRLSEnabled(t *testing.T) {
	migrationsDir := findMigrationsDir(t)

	v, err := NewMigrationValidator(migrationsDir)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	errors := v.ValidateRLSEnabled()
	if len(errors) > 0 {
		for _, e := range errors {
			t.Errorf("RLS error: %s", e.Error())
		}
	}
}

// TestMigrationValidator_ValidateUUIDPrimaryKeys verifies UUID PK checks against real migrations.
func TestMigrationValidator_ValidateUUIDPrimaryKeys(t *testing.T) {
	migrationsDir := findMigrationsDir(t)

	v, err := NewMigrationValidator(migrationsDir)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	errors := v.ValidateUUIDPrimaryKeys()
	if len(errors) > 0 {
		for _, e := range errors {
			t.Errorf("UUID PK error: %s", e.Error())
		}
	}
}

// TestMigrationValidator_ValidateNoORM verifies no ORM pattern detection.
func TestMigrationValidator_ValidateNoORM(t *testing.T) {
	migrationsDir := findMigrationsDir(t)

	v, err := NewMigrationValidator(migrationsDir)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	errors := v.ValidateNoORM()
	if len(errors) > 0 {
		for _, e := range errors {
			t.Errorf("ORM error: %s", e.Error())
		}
	}
}

// TestMigrationValidator_ValidateNoORM_Fail verifies detection of ORM patterns.
func TestMigrationValidator_ValidateNoORM_Fail(t *testing.T) {
	tmpDir := t.TempDir()

	writeFile(t, tmpDir, "000001_test.up.sql", "-- Generated by prisma\nCREATE TABLE IF NOT EXISTS t1 (id UUID PRIMARY KEY);")
	writeFile(t, tmpDir, "000001_test.down.sql", "DROP TABLE IF EXISTS t1;")

	v, err := NewMigrationValidator(tmpDir)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	errors := v.ValidateNoORM()
	if len(errors) == 0 {
		t.Error("expected errors for ORM pattern detection")
	}
}

// TestMigrationValidator_ValidateAll verifies that ValidateAll runs all checks.
func TestMigrationValidator_ValidateAll(t *testing.T) {
	migrationsDir := findMigrationsDir(t)

	v, err := NewMigrationValidator(migrationsDir)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	errors := v.ValidateAll()
	if len(errors) > 0 {
		for _, e := range errors {
			t.Errorf("validation error: %s", e.Error())
		}
	}
}

// TestValidationError_Error verifies the error string format.
func TestValidationError_Error(t *testing.T) {
	err := ValidationError{File: "test.sql", Message: "missing IF NOT EXISTS"}
	expected := "test.sql: missing IF NOT EXISTS"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

// TestStripSQLComments_Mixed verifies comment stripping with mixed content.
func TestStripSQLComments_Mixed(t *testing.T) {
	input := "-- full line comment\nSELECT 1; -- inline\nSELECT 2;"
	result := stripSQLComments(input)
	if strings.Contains(result, "full line comment") {
		t.Error("expected full line comment to be removed")
	}
	if strings.Contains(result, "inline") {
		t.Error("expected inline comment to be removed")
	}
	if !strings.Contains(result, "SELECT 1;") || !strings.Contains(result, "SELECT 2;") {
		t.Error("expected SQL statements to be preserved")
	}
}

// TestContainsWithout_CaseInsensitive verifies case-insensitive matching.
func TestContainsWithout_CaseInsensitive(t *testing.T) {
	sql := "create table IF NOT EXISTS users (id uuid);"
	if containsWithout(sql, "CREATE TABLE", "IF NOT EXISTS") {
		t.Error("expected false for case-insensitive match with qualifier")
	}
}

// TestContainsWithout_QualifierPresent verifies that qualifier is properly detected.
func TestContainsWithout_QualifierPresent(t *testing.T) {
	sql := "DROP TABLE IF EXISTS users;"
	if containsWithout(sql, "DROP TABLE", "IF EXISTS") {
		t.Error("expected false when qualifier is present")
	}
}

// findMigrationsDir locates the migrations directory relative to the test.
func findMigrationsDir(t *testing.T) string {
	t.Helper()
	// Navigate from internal/infrastructure/database/ to migrations/
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Try walking up to find migrations directory
	candidates := []string{
		filepath.Join(dir, "..", "..", "..", "migrations"),
		filepath.Join(dir, "../../../../migrations"),
	}

	for _, candidate := range candidates {
		abs, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}

	t.Fatalf("could not find migrations directory from %s", dir)
	return ""
}

// writeFile is a test helper to create files in a temp directory.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", name, err)
	}
}

