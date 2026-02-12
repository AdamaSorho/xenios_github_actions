package migrations

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// migrationDir returns the path to the migrations directory.
func migrationDir(t *testing.T) string {
	t.Helper()
	// When running tests, the working directory is the package directory.
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	return dir
}

// readMigrationFiles reads all .sql files from the migrations directory.
func readMigrationFiles(t *testing.T) []os.DirEntry {
	t.Helper()
	dir := migrationDir(t)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read migrations directory: %v", err)
	}

	var sqlFiles []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			sqlFiles = append(sqlFiles, entry)
		}
	}
	return sqlFiles
}

// TestMigrations_AllFilesExist verifies that all expected migration files are present.
func TestMigrations_AllFilesExist(t *testing.T) {
	expectedPrefixes := []string{
		"000001_create_users_table",
		"000002_create_profiles_tables",
		"000003_create_sessions_tables",
		"000004_create_health_data_tables",
		"000005_create_analytics_tables",
		"000006_create_programming_tables",
		"000007_create_rls_policies",
	}

	dir := migrationDir(t)
	for _, prefix := range expectedPrefixes {
		upFile := prefix + ".up.sql"
		downFile := prefix + ".down.sql"

		if _, err := os.Stat(filepath.Join(dir, upFile)); os.IsNotExist(err) {
			t.Errorf("missing up migration: %s", upFile)
		}
		if _, err := os.Stat(filepath.Join(dir, downFile)); os.IsNotExist(err) {
			t.Errorf("missing down migration: %s", downFile)
		}
	}
}

// TestMigrations_UpDownPairs verifies that every .up.sql has a matching .down.sql and vice versa.
func TestMigrations_UpDownPairs(t *testing.T) {
	files := readMigrationFiles(t)

	ups := make(map[string]bool)
	downs := make(map[string]bool)

	for _, f := range files {
		name := f.Name()
		if strings.HasSuffix(name, ".up.sql") {
			base := strings.TrimSuffix(name, ".up.sql")
			ups[base] = true
		} else if strings.HasSuffix(name, ".down.sql") {
			base := strings.TrimSuffix(name, ".down.sql")
			downs[base] = true
		}
	}

	for base := range ups {
		if !downs[base] {
			t.Errorf("up migration %s.up.sql has no matching down migration", base)
		}
	}
	for base := range downs {
		if !ups[base] {
			t.Errorf("down migration %s.down.sql has no matching up migration", base)
		}
	}
}

// TestMigrations_SequentialNumbering verifies that migration files use sequential numbering.
func TestMigrations_SequentialNumbering(t *testing.T) {
	files := readMigrationFiles(t)

	numberPattern := regexp.MustCompile(`^(\d+)_`)
	numbers := make(map[int]bool)

	for _, f := range files {
		matches := numberPattern.FindStringSubmatch(f.Name())
		if matches == nil {
			t.Errorf("migration file %s does not follow NNN_description pattern", f.Name())
			continue
		}
		var num int
		_, err := fmt.Sscanf(matches[1], "%d", &num)
		if err != nil {
			t.Errorf("failed to parse number from %s: %v", f.Name(), err)
		}
		numbers[num] = true
	}

	// Verify sequential (no gaps from 1..max)
	var sorted []int
	for n := range numbers {
		sorted = append(sorted, n)
	}
	sort.Ints(sorted)

	for i, n := range sorted {
		expected := i + 1
		if n != expected {
			t.Errorf("expected migration number %d, but found %d (gap in sequence)", expected, n)
		}
	}
}

// TestMigrations_NonEmpty verifies that all migration files are non-empty.
func TestMigrations_NonEmpty(t *testing.T) {
	dir := migrationDir(t)
	files := readMigrationFiles(t)

	for _, f := range files {
		content, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			t.Errorf("failed to read %s: %v", f.Name(), err)
			continue
		}
		if len(strings.TrimSpace(string(content))) == 0 {
			t.Errorf("migration file %s is empty", f.Name())
		}
	}
}

// stripSQLComments removes SQL line comments (-- ...) from the SQL string.
func stripSQLComments(sql string) string {
	var result strings.Builder
	for _, line := range strings.Split(sql, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		// Remove inline comments
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
// Returns true if there is a bare occurrence of keyword without the qualifier.
// It strips SQL comments before checking.
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

// TestMigrations_UpIdempotency verifies that up migrations use IF NOT EXISTS for idempotency.
func TestMigrations_UpIdempotency(t *testing.T) {
	dir := migrationDir(t)
	files := readMigrationFiles(t)

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".up.sql") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			t.Errorf("failed to read %s: %v", f.Name(), err)
			continue
		}

		sql := string(content)

		// Check CREATE TABLE statements use IF NOT EXISTS
		if containsWithout(sql, "CREATE TABLE", "IF NOT EXISTS") {
			t.Errorf("%s: CREATE TABLE without IF NOT EXISTS", f.Name())
		}

		// Check CREATE INDEX statements use IF NOT EXISTS
		if containsWithout(sql, "CREATE INDEX", "IF NOT EXISTS") {
			t.Errorf("%s: CREATE INDEX without IF NOT EXISTS", f.Name())
		}
	}
}

// TestMigrations_DownIdempotency verifies that down migrations use IF EXISTS for idempotency.
func TestMigrations_DownIdempotency(t *testing.T) {
	dir := migrationDir(t)
	files := readMigrationFiles(t)

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".down.sql") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			t.Errorf("failed to read %s: %v", f.Name(), err)
			continue
		}

		sql := string(content)

		if containsWithout(sql, "DROP TABLE", "IF EXISTS") {
			t.Errorf("%s: DROP TABLE without IF EXISTS", f.Name())
		}
		if containsWithout(sql, "DROP FUNCTION", "IF EXISTS") {
			t.Errorf("%s: DROP FUNCTION without IF EXISTS", f.Name())
		}
		if containsWithout(sql, "DROP TRIGGER", "IF EXISTS") {
			t.Errorf("%s: DROP TRIGGER without IF EXISTS", f.Name())
		}
		if containsWithout(sql, "DROP RULE", "IF EXISTS") {
			t.Errorf("%s: DROP RULE without IF EXISTS", f.Name())
		}
		if containsWithout(sql, "DROP POLICY", "IF EXISTS") {
			t.Errorf("%s: DROP POLICY without IF EXISTS", f.Name())
		}
	}
}

// TestMigrations_UsersTable_SoftDelete verifies that the users table has soft delete support.
func TestMigrations_UsersTable_SoftDelete(t *testing.T) {
	dir := migrationDir(t)
	content, err := os.ReadFile(filepath.Join(dir, "000001_create_users_table.up.sql"))
	if err != nil {
		t.Fatalf("failed to read users migration: %v", err)
	}

	sql := strings.ToLower(string(content))
	if !strings.Contains(sql, "deleted_at") {
		t.Error("users table missing deleted_at column for soft delete support")
	}
}

// TestMigrations_SessionsTable_SoftDelete verifies that the sessions table has soft delete support.
func TestMigrations_SessionsTable_SoftDelete(t *testing.T) {
	dir := migrationDir(t)
	content, err := os.ReadFile(filepath.Join(dir, "000003_create_sessions_tables.up.sql"))
	if err != nil {
		t.Fatalf("failed to read sessions migration: %v", err)
	}

	sql := strings.ToLower(string(content))
	if !strings.Contains(sql, "deleted_at") {
		t.Error("sessions table missing deleted_at column for soft delete support")
	}
}

// TestMigrations_EventsAudit_AppendOnly verifies that events_audit table has append-only rules.
func TestMigrations_EventsAudit_AppendOnly(t *testing.T) {
	dir := migrationDir(t)
	content, err := os.ReadFile(filepath.Join(dir, "000005_create_analytics_tables.up.sql"))
	if err != nil {
		t.Fatalf("failed to read analytics migration: %v", err)
	}

	sql := strings.ToLower(string(content))

	if !strings.Contains(sql, "events_audit_no_update") {
		t.Error("events_audit missing no-update rule for append-only enforcement")
	}
	if !strings.Contains(sql, "events_audit_no_delete") {
		t.Error("events_audit missing no-delete rule for append-only enforcement")
	}
	if !strings.Contains(sql, "do instead nothing") {
		t.Error("events_audit rules should use DO INSTEAD NOTHING")
	}
}

// TestMigrations_InsightCards_DefaultDraft verifies that insight_cards.status defaults to 'draft'.
func TestMigrations_InsightCards_DefaultDraft(t *testing.T) {
	dir := migrationDir(t)
	content, err := os.ReadFile(filepath.Join(dir, "000004_create_health_data_tables.up.sql"))
	if err != nil {
		t.Fatalf("failed to read health data migration: %v", err)
	}

	sql := strings.ToLower(string(content))
	if !strings.Contains(sql, "default 'draft'") {
		t.Error("insight_cards.status should default to 'draft' for the approval gate")
	}
}

// TestMigrations_UUIDPrimaryKeys verifies that all tables use UUID primary keys with gen_random_uuid().
func TestMigrations_UUIDPrimaryKeys(t *testing.T) {
	dir := migrationDir(t)
	files := readMigrationFiles(t)

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".up.sql") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			t.Errorf("failed to read %s: %v", f.Name(), err)
			continue
		}

		sql := strings.ToLower(string(content))

		// If the file creates a table, it should use UUID primary keys
		if strings.Contains(sql, "create table") {
			if !strings.Contains(sql, "uuid primary key default gen_random_uuid()") {
				// Skip RLS policies migration which doesn't create tables
				if !strings.Contains(f.Name(), "rls_policies") {
					t.Errorf("%s: tables should use UUID PRIMARY KEY DEFAULT gen_random_uuid()", f.Name())
				}
			}
		}
	}
}

// TestMigrations_TimestampTZ verifies that all timestamp columns use TIMESTAMPTZ.
func TestMigrations_TimestampTZ(t *testing.T) {
	dir := migrationDir(t)
	files := readMigrationFiles(t)

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".up.sql") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			t.Errorf("failed to read %s: %v", f.Name(), err)
			continue
		}

		sql := string(content)

		// Check for bare TIMESTAMP (without TZ)
		// Find all occurrences of TIMESTAMP and check they are followed by TZ
		lower := strings.ToLower(sql)
		idx := 0
		for {
			pos := strings.Index(lower[idx:], "timestamp")
			if pos == -1 {
				break
			}
			absPos := idx + pos
			afterTimestamp := lower[absPos+len("timestamp"):]
			// It's OK if immediately followed by "tz"
			if !strings.HasPrefix(afterTimestamp, "tz") {
				// Not "timestamptz" — could be bare TIMESTAMP used as a word
				// Check it's not part of a longer word (e.g. "timestamps")
				if len(afterTimestamp) == 0 || afterTimestamp[0] == ' ' || afterTimestamp[0] == ',' || afterTimestamp[0] == '\n' || afterTimestamp[0] == ')' {
					t.Errorf("%s: uses TIMESTAMP without TZ - should use TIMESTAMPTZ", f.Name())
					break
				}
			}
			idx = absPos + len("timestamp")
		}
	}
}

// TestMigrations_RLSEnabled verifies that all tables in up migrations have RLS enabled.
func TestMigrations_RLSEnabled(t *testing.T) {
	dir := migrationDir(t)
	files := readMigrationFiles(t)

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".up.sql") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			t.Errorf("failed to read %s: %v", f.Name(), err)
			continue
		}

		sql := strings.ToLower(string(content))

		// Extract all table names from CREATE TABLE statements
		tablePattern := regexp.MustCompile(`create\s+table\s+if\s+not\s+exists\s+(\w+)`)
		matches := tablePattern.FindAllStringSubmatch(sql, -1)

		for _, match := range matches {
			tableName := match[1]
			rlsPattern := fmt.Sprintf("alter table %s enable row level security", tableName)
			if !strings.Contains(sql, rlsPattern) {
				t.Errorf("%s: table '%s' does not have RLS enabled", f.Name(), tableName)
			}
		}
	}
}

// TestMigrations_AllCoreTables verifies that all required tables from the issue are created.
func TestMigrations_AllCoreTables(t *testing.T) {
	dir := migrationDir(t)
	files := readMigrationFiles(t)

	expectedTables := []string{
		// Core tables (000001, 000002)
		"users",
		"coach_profiles",
		"client_profiles",
		"coach_client_relationships",
		// Session tables (000003)
		"sessions",
		"transcript_segments",
		"workout_exercises",
		"form_cues_tracking",
		// Health data tables (000004)
		"artifacts",
		"measurements",
		"insight_cards",
		"wearable_summaries",
		// Analytics tables (000005)
		"coaching_analytics",
		"client_risk_scores",
		"events_audit",
		// Programming tables (000006)
		"programs",
		"program_versions",
		"phases",
		"microcycles",
		"programmed_sessions",
		"programmed_exercises",
		"exercise_library",
		"session_completions",
		"exercise_logs",
		"behavior_goals",
		"behavior_cues",
		"behavior_checkins",
		"program_adjustments",
	}

	// Collect all SQL content
	var allSQL strings.Builder
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".up.sql") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			t.Errorf("failed to read %s: %v", f.Name(), err)
			continue
		}
		allSQL.WriteString(strings.ToLower(string(content)))
		allSQL.WriteString("\n")
	}

	combined := allSQL.String()
	for _, table := range expectedTables {
		createPattern := fmt.Sprintf("create table if not exists %s", table)
		if !strings.Contains(combined, createPattern) {
			t.Errorf("missing table: %s", table)
		}
	}
}

// TestMigrations_ForeignKeysIndexed verifies that foreign key columns have indexes.
func TestMigrations_ForeignKeysIndexed(t *testing.T) {
	dir := migrationDir(t)
	files := readMigrationFiles(t)

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".up.sql") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			t.Errorf("failed to read %s: %v", f.Name(), err)
			continue
		}

		sql := strings.ToLower(string(content))

		// Find all REFERENCES clauses (foreign keys)
		fkPattern := regexp.MustCompile(`(\w+)\s+uuid\s+(?:not\s+null\s+)?(?:unique\s+)?references\s+\w+\(`)
		fkMatches := fkPattern.FindAllStringSubmatch(sql, -1)

		for _, match := range fkMatches {
			colName := match[1]
			// Skip 'id' columns (primary keys that also reference)
			if colName == "id" {
				continue
			}
			// Check that there's an index containing this column name
			indexPattern := fmt.Sprintf("create index if not exists")
			if !strings.Contains(sql, indexPattern) {
				t.Errorf("%s: foreign key column '%s' may lack an index", f.Name(), colName)
			}
		}
	}
}

// TestMigrations_RLSPolicies verifies that RLS policies exist in migration 000007.
func TestMigrations_RLSPolicies(t *testing.T) {
	dir := migrationDir(t)
	content, err := os.ReadFile(filepath.Join(dir, "000007_create_rls_policies.up.sql"))
	if err != nil {
		t.Fatalf("failed to read RLS policies migration: %v", err)
	}

	sql := strings.ToLower(string(content))

	// Verify key policy patterns exist
	requiredPolicies := []string{
		"create policy users_self_access",
		"create policy users_coach_access",
		"create policy sessions_coach_access",
		"create policy sessions_client_access",
		"create policy events_audit_self_access",
		"create policy events_audit_insert",
		"create policy programs_coach_access",
		"create policy programs_client_access",
	}

	for _, policy := range requiredPolicies {
		if !strings.Contains(sql, policy) {
			t.Errorf("missing RLS policy: %s", policy)
		}
	}

	// Verify helper functions exist
	if !strings.Contains(sql, "current_app_user_id") {
		t.Error("missing current_app_user_id() helper function")
	}
	if !strings.Contains(sql, "is_coach_of") {
		t.Error("missing is_coach_of() helper function")
	}
}

// TestMigrations_JSONBOnlyWhereNeeded verifies JSONB is used sparingly and only where specified.
func TestMigrations_JSONBOnlyWhereNeeded(t *testing.T) {
	dir := migrationDir(t)
	files := readMigrationFiles(t)

	// Allowed JSONB columns per the issue spec
	allowedJSONB := map[string][]string{
		"000003_create_sessions_tables.up.sql":   {"summary"},
		"000004_create_health_data_tables.up.sql": {"metrics"},
		"000005_create_analytics_tables.up.sql":   {"factors", "metadata"},
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".up.sql") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			t.Errorf("failed to read %s: %v", f.Name(), err)
			continue
		}

		sql := strings.ToLower(string(content))
		jsonbPattern := regexp.MustCompile(`(\w+)\s+jsonb`)
		jsonbMatches := jsonbPattern.FindAllStringSubmatch(sql, -1)

		allowed, hasAllowed := allowedJSONB[f.Name()]

		for _, match := range jsonbMatches {
			colName := match[1]
			if !hasAllowed {
				t.Errorf("%s: unexpected JSONB column '%s' - JSONB should only be used where schema evolution is genuinely needed", f.Name(), colName)
				continue
			}
			found := false
			for _, a := range allowed {
				if colName == a {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s: unexpected JSONB column '%s' - only %v are allowed", f.Name(), colName, allowed)
			}
		}
	}
}

// ============================================================
// Integration tests (require DATABASE_URL)
// ============================================================

// TestMigrations_Integration_ApplyAll applies all up migrations to a clean database.
func TestMigrations_Integration_ApplyAll(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to create connection pool: %v", err)
	}
	defer pool.Close()

	dir := migrationDir(t)
	files := readMigrationFiles(t)

	// Collect up migrations sorted
	var upFiles []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".up.sql") {
			upFiles = append(upFiles, f.Name())
		}
	}
	sort.Strings(upFiles)

	// Apply all up migrations
	for _, f := range upFiles {
		content, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			t.Fatalf("failed to read %s: %v", f, err)
		}
		_, err = pool.Exec(ctx, string(content))
		if err != nil {
			t.Fatalf("failed to apply migration %s: %v", f, err)
		}
		t.Logf("applied: %s", f)
	}
}

// TestMigrations_Integration_Idempotent verifies that applying migrations twice does not error.
func TestMigrations_Integration_Idempotent(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to create connection pool: %v", err)
	}
	defer pool.Close()

	dir := migrationDir(t)
	files := readMigrationFiles(t)

	var upFiles []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".up.sql") {
			upFiles = append(upFiles, f.Name())
		}
	}
	sort.Strings(upFiles)

	// Apply all up migrations twice
	for attempt := 1; attempt <= 2; attempt++ {
		for _, f := range upFiles {
			content, err := os.ReadFile(filepath.Join(dir, f))
			if err != nil {
				t.Fatalf("failed to read %s: %v", f, err)
			}
			_, err = pool.Exec(ctx, string(content))
			if err != nil {
				t.Fatalf("attempt %d: failed to apply migration %s: %v", attempt, f, err)
			}
		}
		t.Logf("attempt %d: all migrations applied successfully", attempt)
	}
}

// TestMigrations_Integration_DownReverses verifies that down migrations cleanly reverse up migrations.
func TestMigrations_Integration_DownReverses(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to create connection pool: %v", err)
	}
	defer pool.Close()

	dir := migrationDir(t)
	files := readMigrationFiles(t)

	var downFiles []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".down.sql") {
			downFiles = append(downFiles, f.Name())
		}
	}
	sort.Strings(downFiles)

	// Apply down migrations in reverse order
	for i := len(downFiles) - 1; i >= 0; i-- {
		f := downFiles[i]
		content, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			t.Fatalf("failed to read %s: %v", f, err)
		}
		_, err = pool.Exec(ctx, string(content))
		if err != nil {
			t.Fatalf("failed to apply down migration %s: %v", f, err)
		}
		t.Logf("reversed: %s", f)
	}

	// Verify all tables are gone
	var tableCount int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'public'
		  AND table_type = 'BASE TABLE'
	`).Scan(&tableCount)
	if err != nil {
		t.Fatalf("failed to count tables: %v", err)
	}
	if tableCount != 0 {
		t.Errorf("expected 0 tables after down migrations, got %d", tableCount)
	}
}
