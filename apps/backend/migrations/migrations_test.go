package migrations

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// migrationsDir returns the absolute path to the migrations directory.
func migrationsDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine test file path")
	}
	return filepath.Dir(filename)
}

// TestMigration000002_UpFileExists verifies the up migration file exists.
func TestMigration000002_UpFileExists(t *testing.T) {
	dir := migrationsDir(t)
	path := filepath.Join(dir, "000002_add_password_hash_to_users.up.sql")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("000002_add_password_hash_to_users.up.sql does not exist")
	}
}

// TestMigration000002_DownFileExists verifies the down migration file exists.
func TestMigration000002_DownFileExists(t *testing.T) {
	dir := migrationsDir(t)
	path := filepath.Join(dir, "000002_add_password_hash_to_users.down.sql")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("000002_add_password_hash_to_users.down.sql does not exist")
	}
}

// TestMigration000002_UpContainsAlterTable verifies the up migration adds the password_hash column.
func TestMigration000002_UpContainsAlterTable(t *testing.T) {
	dir := migrationsDir(t)
	path := filepath.Join(dir, "000002_add_password_hash_to_users.up.sql")

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read up migration: %v", err)
	}

	sql := string(content)
	upper := strings.ToUpper(sql)

	// Must alter the users table
	if !strings.Contains(upper, "ALTER TABLE") {
		t.Error("up migration must contain ALTER TABLE statement")
	}

	if !strings.Contains(upper, "USERS") {
		t.Error("up migration must reference the users table")
	}

	// Must add the password_hash column
	if !strings.Contains(upper, "ADD COLUMN") {
		t.Error("up migration must contain ADD COLUMN")
	}

	if !strings.Contains(sql, "password_hash") {
		t.Error("up migration must add password_hash column")
	}

	// Must be TEXT NOT NULL
	if !strings.Contains(upper, "TEXT") {
		t.Error("up migration must use TEXT type for password_hash")
	}

	if !strings.Contains(upper, "NOT NULL") {
		t.Error("up migration must specify NOT NULL for password_hash")
	}

	// Should use IF NOT EXISTS for idempotency
	if !strings.Contains(upper, "IF NOT EXISTS") {
		t.Error("up migration should use IF NOT EXISTS for idempotency")
	}
}

// TestMigration000002_DownContainsDropColumn verifies the down migration drops the password_hash column.
func TestMigration000002_DownContainsDropColumn(t *testing.T) {
	dir := migrationsDir(t)
	path := filepath.Join(dir, "000002_add_password_hash_to_users.down.sql")

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read down migration: %v", err)
	}

	sql := string(content)
	upper := strings.ToUpper(sql)

	// Must drop the password_hash column from users table
	if !strings.Contains(upper, "ALTER TABLE") {
		t.Error("down migration must contain ALTER TABLE statement")
	}

	if !strings.Contains(upper, "USERS") {
		t.Error("down migration must reference the users table")
	}

	if !strings.Contains(upper, "DROP COLUMN") {
		t.Error("down migration must contain DROP COLUMN")
	}

	if !strings.Contains(sql, "password_hash") {
		t.Error("down migration must reference password_hash column")
	}

	// Should use IF EXISTS for safe rollback
	if !strings.Contains(upper, "IF EXISTS") {
		t.Error("down migration should use IF EXISTS for safe rollback")
	}
}

// TestMigration000002_DownPreservesTable verifies the down migration does NOT drop the entire users table.
func TestMigration000002_DownPreservesTable(t *testing.T) {
	dir := migrationsDir(t)
	path := filepath.Join(dir, "000002_add_password_hash_to_users.down.sql")

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read down migration: %v", err)
	}

	upper := strings.ToUpper(string(content))

	if strings.Contains(upper, "DROP TABLE") {
		t.Error("down migration must NOT drop the entire users table, only the password_hash column")
	}
}

// TestMigration000002_UpNoSQLInjection verifies the migration does not use string interpolation.
func TestMigration000002_UpNoSQLInjection(t *testing.T) {
	dir := migrationsDir(t)
	path := filepath.Join(dir, "000002_add_password_hash_to_users.up.sql")

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read up migration: %v", err)
	}

	sql := string(content)

	// Check for Go template syntax
	unsafePatterns := []struct {
		pattern string
		desc    string
	}{
		{"${", "shell variable interpolation"},
		{"%s", "Go fmt.Sprintf string interpolation"},
		{"%v", "Go fmt.Sprintf value interpolation"},
		{"{{", "Go template syntax"},
	}

	for _, p := range unsafePatterns {
		if strings.Contains(sql, p.pattern) {
			t.Errorf("up migration contains %s pattern (%s) — risk of SQL injection", p.pattern, p.desc)
		}
	}
}

// TestMigration000002_DownNoSQLInjection verifies the down migration does not use string interpolation.
func TestMigration000002_DownNoSQLInjection(t *testing.T) {
	dir := migrationsDir(t)
	path := filepath.Join(dir, "000002_add_password_hash_to_users.down.sql")

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read down migration: %v", err)
	}

	sql := string(content)

	unsafePatterns := []struct {
		pattern string
		desc    string
	}{
		{"${", "shell variable interpolation"},
		{"%s", "Go fmt.Sprintf string interpolation"},
		{"%v", "Go fmt.Sprintf value interpolation"},
		{"{{", "Go template syntax"},
	}

	for _, p := range unsafePatterns {
		if strings.Contains(sql, p.pattern) {
			t.Errorf("down migration contains %s pattern (%s) — risk of SQL injection", p.pattern, p.desc)
		}
	}
}
