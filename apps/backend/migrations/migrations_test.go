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

// readMigrationSQL reads the SQL content of a specific migration file.
func readMigrationSQL(t *testing.T, filename string) string {
	t.Helper()
	dir := migrationDir(t)
	content, err := os.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		t.Fatalf("failed to read %s: %v", filename, err)
	}
	return string(content)
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
		"000008_create_job_queue",
		"000009_create_refresh_tokens",
		"000011_audit_trigger",
		"000012_add_document_subtype",
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

// TestMigrations_TriggersIdempotent verifies that CREATE TRIGGER statements have
// DROP TRIGGER IF EXISTS guards for idempotency.
func TestMigrations_TriggersIdempotent(t *testing.T) {
	dir := migrationDir(t)
	files := readMigrationFiles(t)

	triggerPattern := regexp.MustCompile(`(?i)CREATE\s+TRIGGER\s+(\w+)\s+`)

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
		matches := triggerPattern.FindAllStringSubmatch(sql, -1)

		for _, match := range matches {
			triggerName := strings.ToLower(match[1])
			dropPattern := fmt.Sprintf("drop trigger if exists %s", triggerName)
			if !strings.Contains(strings.ToLower(sql), dropPattern) {
				t.Errorf("%s: CREATE TRIGGER %s lacks DROP TRIGGER IF EXISTS guard", f.Name(), triggerName)
			}
		}
	}
}

// TestMigrations_PoliciesIdempotent verifies that CREATE POLICY statements have
// DROP POLICY IF EXISTS guards for idempotency.
func TestMigrations_PoliciesIdempotent(t *testing.T) {
	dir := migrationDir(t)
	files := readMigrationFiles(t)

	policyPattern := regexp.MustCompile(`(?i)CREATE\s+POLICY\s+(\w+)\s+ON\s+(\w+)`)

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
		matches := policyPattern.FindAllStringSubmatch(sql, -1)

		for _, match := range matches {
			policyName := strings.ToLower(match[1])
			tableName := strings.ToLower(match[2])
			dropPattern := fmt.Sprintf("drop policy if exists %s on %s", policyName, tableName)
			if !strings.Contains(strings.ToLower(sql), dropPattern) {
				t.Errorf("%s: CREATE POLICY %s ON %s lacks DROP POLICY IF EXISTS guard", f.Name(), policyName, tableName)
			}
		}
	}
}

// TestMigrations_RulesIdempotent verifies that CREATE RULE statements use
// CREATE OR REPLACE RULE for idempotency.
func TestMigrations_RulesIdempotent(t *testing.T) {
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
		lower := strings.ToLower(stripSQLComments(sql))

		// Find bare CREATE RULE (without OR REPLACE)
		idx := 0
		for {
			pos := strings.Index(lower[idx:], "create rule")
			if pos == -1 {
				break
			}
			absPos := idx + pos
			// Check that it's actually CREATE OR REPLACE RULE
			before := lower[max(0, absPos-20):absPos]
			if !strings.Contains(before, "or replace") {
				// Check if "or replace" comes right after "create"
				afterCreate := strings.TrimSpace(lower[absPos+len("create"):])
				if !strings.HasPrefix(afterCreate, "or replace rule") {
					t.Errorf("%s: CREATE RULE without OR REPLACE - should use CREATE OR REPLACE RULE for idempotency", f.Name())
					break
				}
			}
			idx = absPos + len("create rule")
		}
	}
}

// TestMigrations_UsersTable_SoftDelete verifies that the users table has soft delete support.
func TestMigrations_UsersTable_SoftDelete(t *testing.T) {
	sql := strings.ToLower(readMigrationSQL(t, "000001_create_users_table.up.sql"))
	if !strings.Contains(sql, "deleted_at") {
		t.Error("users table missing deleted_at column for soft delete support")
	}
}

// TestMigrations_SessionsTable_SoftDelete verifies that the sessions table has soft delete support.
func TestMigrations_SessionsTable_SoftDelete(t *testing.T) {
	sql := strings.ToLower(readMigrationSQL(t, "000003_create_sessions_tables.up.sql"))
	if !strings.Contains(sql, "deleted_at") {
		t.Error("sessions table missing deleted_at column for soft delete support")
	}
}

// TestMigrations_EventsAudit_AppendOnly verifies that events_audit table has append-only rules.
func TestMigrations_EventsAudit_AppendOnly(t *testing.T) {
	sql := strings.ToLower(readMigrationSQL(t, "000005_create_analytics_tables.up.sql"))

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

// TestMigrations_AuditTrigger_ReplacesRules verifies that migration 000010 replaces
// rules with trigger-based append-only enforcement.
func TestMigrations_AuditTrigger_ReplacesRules(t *testing.T) {
	sql := strings.ToLower(readMigrationSQL(t, "000011_audit_trigger.up.sql"))

	if !strings.Contains(sql, "drop rule if exists events_audit_no_update") {
		t.Error("000010 should drop the no-update rule")
	}
	if !strings.Contains(sql, "drop rule if exists events_audit_no_delete") {
		t.Error("000010 should drop the no-delete rule")
	}
	if !strings.Contains(sql, "prevent_audit_mutation") {
		t.Error("000010 should create prevent_audit_mutation trigger function")
	}
	if !strings.Contains(sql, "raise exception") {
		t.Error("000010 trigger should raise an exception on mutation attempts")
	}
	if !strings.Contains(sql, "before update") {
		t.Error("000010 should create a BEFORE UPDATE trigger")
	}
	if !strings.Contains(sql, "before delete") {
		t.Error("000010 should create a BEFORE DELETE trigger")
	}
}

// TestMigrations_AuditTrigger_Down_RestoresRules verifies that the down migration restores rules.
func TestMigrations_AuditTrigger_Down_RestoresRules(t *testing.T) {
	sql := strings.ToLower(readMigrationSQL(t, "000011_audit_trigger.down.sql"))

	if !strings.Contains(sql, "drop trigger if exists events_audit_no_update") {
		t.Error("000010 down should drop the update trigger")
	}
	if !strings.Contains(sql, "drop trigger if exists events_audit_no_delete") {
		t.Error("000010 down should drop the delete trigger")
	}
	if !strings.Contains(sql, "drop function if exists prevent_audit_mutation") {
		t.Error("000010 down should drop the trigger function")
	}
	if !strings.Contains(sql, "do instead nothing") {
		t.Error("000010 down should restore DO INSTEAD NOTHING rules")
	}
}

// TestMigrations_EventsAudit_ActorIdOnDeleteRestrict verifies that events_audit.actor_id
// uses ON DELETE RESTRICT to prevent user deletion when audit entries exist.
func TestMigrations_EventsAudit_ActorIdOnDeleteRestrict(t *testing.T) {
	sql := strings.ToLower(readMigrationSQL(t, "000005_create_analytics_tables.up.sql"))

	// actor_id should have an explicit ON DELETE action
	if !strings.Contains(sql, "on delete restrict") {
		t.Error("events_audit.actor_id should use ON DELETE RESTRICT to preserve audit log integrity")
	}
}

// TestMigrations_InsightCards_DefaultDraft verifies that insight_cards.status defaults to 'draft'.
func TestMigrations_InsightCards_DefaultDraft(t *testing.T) {
	sql := strings.ToLower(readMigrationSQL(t, "000004_create_health_data_tables.up.sql"))
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
		// Job queue tables (000008)
		"jobs",
		"jobs_dead_letter",
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
			// Skip columns with UNIQUE constraint (already indexed implicitly)
			fkLine := match[0]
			if strings.Contains(fkLine, "unique") {
				continue
			}
			// Check that there's an index definition containing this FK column name
			indexColPattern := fmt.Sprintf("create index if not exists idx_\\w+_%s", regexp.QuoteMeta(colName))
			indexByCol := regexp.MustCompile(indexColPattern)
			// Also check for index definitions that reference the column in their ON clause
			indexOnPattern := fmt.Sprintf("on \\w+\\(%s", regexp.QuoteMeta(colName))
			indexByOn := regexp.MustCompile(indexOnPattern)
			if !indexByCol.MatchString(sql) && !indexByOn.MatchString(sql) {
				t.Errorf("%s: foreign key column '%s' may lack an index", f.Name(), colName)
			}
		}
	}
}

// TestMigrations_RLSPolicies verifies that RLS policies exist in migration 000007.
func TestMigrations_RLSPolicies(t *testing.T) {
	sql := strings.ToLower(readMigrationSQL(t, "000007_create_rls_policies.up.sql"))

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
		"000003_create_sessions_tables.up.sql":     {"summary"},
		"000004_create_health_data_tables.up.sql":  {"metrics"},
		"000005_create_analytics_tables.up.sql":    {"factors", "metadata"},
		"000008_create_job_queue.up.sql":           {"payload", "details"},
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

// TestMigrations_CheckConstraints verifies that enum-like columns use CHECK constraints.
func TestMigrations_CheckConstraints(t *testing.T) {
	dir := migrationDir(t)
	files := readMigrationFiles(t)

	// Known columns that should have CHECK constraints
	checkedColumns := map[string][]string{
		"000001_create_users_table.up.sql":       {"role"},
		"000003_create_sessions_tables.up.sql":   {"session_type", "status", "speaker", "cue_type"},
		"000004_create_health_data_tables.up.sql": {"category", "status", "priority"},
		"000006_create_programming_tables.up.sql": {"category", "difficulty", "status", "intensity_level", "session_type", "cue_type", "adjustment_type"},
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".up.sql") {
			continue
		}

		cols, hasCols := checkedColumns[f.Name()]
		if !hasCols {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			t.Errorf("failed to read %s: %v", f.Name(), err)
			continue
		}

		sql := strings.ToLower(string(content))
		for _, col := range cols {
			// Check that the column has a CHECK constraint with IN clause
			checkPattern := fmt.Sprintf("%s in (", col)
			if !strings.Contains(sql, checkPattern) {
				t.Errorf("%s: column '%s' should have a CHECK constraint with IN clause", f.Name(), col)
			}
		}
	}
}

// TestMigrations_ParameterizedQueries verifies that no raw string interpolation is used in SQL.
func TestMigrations_ParameterizedQueries(t *testing.T) {
	dir := migrationDir(t)
	files := readMigrationFiles(t)

	for _, f := range files {
		content, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			t.Errorf("failed to read %s: %v", f.Name(), err)
			continue
		}

		sql := string(content)

		// SQL migrations should not contain Go-style format strings or JS template literals
		if strings.Contains(sql, "fmt.Sprintf") || strings.Contains(sql, "${") {
			t.Errorf("%s: contains string interpolation patterns - use parameterized queries", f.Name())
		}
	}
}

// TestMigrations_CoachingAnalytics_UniqueConstraint verifies the unique constraint on coaching_analytics.
func TestMigrations_CoachingAnalytics_UniqueConstraint(t *testing.T) {
	sql := strings.ToLower(readMigrationSQL(t, "000005_create_analytics_tables.up.sql"))

	if !strings.Contains(sql, "unique(coach_id, client_id, period_start, period_end)") {
		t.Error("coaching_analytics should have UNIQUE constraint on (coach_id, client_id, period_start, period_end)")
	}
}

// TestMigrations_SessionCompletions_SessionIdIndex verifies session_completions has index on session_id.
func TestMigrations_SessionCompletions_SessionIdIndex(t *testing.T) {
	sql := strings.ToLower(readMigrationSQL(t, "000006_create_programming_tables.up.sql"))

	if !strings.Contains(sql, "idx_session_completions_session_id") {
		t.Error("session_completions should have an index on session_id")
	}
}

// TestMigrations_ProgramVersions_CreatedByIndex verifies program_versions has index on created_by.
func TestMigrations_ProgramVersions_CreatedByIndex(t *testing.T) {
	sql := strings.ToLower(readMigrationSQL(t, "000006_create_programming_tables.up.sql"))

	if !strings.Contains(sql, "idx_program_versions_created_by") {
		t.Error("program_versions should have an index on created_by FK column")
	}
}

// TestMigrations_ExerciseLibrary_OnDeleteSetNull verifies exercise_library.created_by has ON DELETE SET NULL.
func TestMigrations_ExerciseLibrary_OnDeleteSetNull(t *testing.T) {
	sql := strings.ToLower(readMigrationSQL(t, "000006_create_programming_tables.up.sql"))

	// Find the exercise_library created_by line
	lines := strings.Split(sql, "\n")
	for _, line := range lines {
		if strings.Contains(line, "created_by") && strings.Contains(line, "references users(id)") {
			if !strings.Contains(line, "on delete set null") {
				t.Error("exercise_library.created_by should use ON DELETE SET NULL")
			}
			return
		}
	}
}

// TestMigrations_HelperFunctions_ErrorHandling verifies that helper functions handle errors.
func TestMigrations_HelperFunctions_ErrorHandling(t *testing.T) {
	sql := strings.ToLower(readMigrationSQL(t, "000007_create_rls_policies.up.sql"))

	// current_app_user_id should handle exceptions
	if !strings.Contains(sql, "exception") {
		t.Error("current_app_user_id() should have EXCEPTION handler for error resilience")
	}
	if !strings.Contains(sql, "when others then return null") {
		t.Error("current_app_user_id() should return NULL on exception")
	}
}

// TestMigrations_DownMigrations_ReverseOrder verifies that down migrations drop in reverse dependency order.
func TestMigrations_DownMigrations_ReverseOrder(t *testing.T) {
	// In 000007.down.sql, policies should be dropped before functions
	sql := readMigrationSQL(t, "000007_create_rls_policies.down.sql")
	lower := strings.ToLower(sql)

	funcDropPos := strings.Index(lower, "drop function")
	policyDropPos := strings.LastIndex(lower, "drop policy")

	if funcDropPos < policyDropPos {
		t.Error("000007.down.sql: functions should be dropped after policies")
	}
}

// TestMigrations_NoORMPatterns verifies that no ORM patterns are used in migrations.
func TestMigrations_NoORMPatterns(t *testing.T) {
	dir := migrationDir(t)
	files := readMigrationFiles(t)

	ormPatterns := []string{
		"sequelize",
		"typeorm",
		"prisma",
		"drizzle",
		"gorm",
		"migrate.up",
		"migrate.down",
	}

	for _, f := range files {
		content, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			t.Errorf("failed to read %s: %v", f.Name(), err)
			continue
		}

		sql := strings.ToLower(string(content))
		for _, pattern := range ormPatterns {
			if strings.Contains(sql, pattern) {
				t.Errorf("%s: contains ORM pattern '%s' - must use raw SQL only", f.Name(), pattern)
			}
		}
	}
}

// TestMigrations_FileSize_Reasonable verifies migration files are not suspiciously small or large.
func TestMigrations_FileSize_Reasonable(t *testing.T) {
	dir := migrationDir(t)
	files := readMigrationFiles(t)

	for _, f := range files {
		info, err := os.Stat(filepath.Join(dir, f.Name()))
		if err != nil {
			t.Errorf("failed to stat %s: %v", f.Name(), err)
			continue
		}

		// Migration files should have meaningful content (>10 bytes)
		if info.Size() < 10 {
			t.Errorf("%s: file is suspiciously small (%d bytes)", f.Name(), info.Size())
		}

		// Migration files should not be excessively large (>100KB)
		if info.Size() > 100*1024 {
			t.Errorf("%s: file is suspiciously large (%d bytes)", f.Name(), info.Size())
		}
	}
}

// TestMigrations_EventsAuditEntityTimeIndex verifies composite index for time-range audit queries.
func TestMigrations_EventsAuditEntityTimeIndex(t *testing.T) {
	sql := strings.ToLower(readMigrationSQL(t, "000005_create_analytics_tables.up.sql"))

	if !strings.Contains(sql, "idx_events_audit_entity_time") {
		t.Error("events_audit should have composite index on (entity_type, entity_id, created_at) for time-range queries")
	}
}

// TestStripSQLComments_EmptyInput verifies stripSQLComments handles empty input.
func TestStripSQLComments_EmptyInput(t *testing.T) {
	result := stripSQLComments("")
	if strings.TrimSpace(result) != "" {
		t.Errorf("expected empty result for empty input, got %q", result)
	}
}

// TestStripSQLComments_OnlyComments verifies stripSQLComments handles comment-only input.
func TestStripSQLComments_OnlyComments(t *testing.T) {
	input := "-- this is a comment\n-- another comment"
	result := stripSQLComments(input)
	if strings.TrimSpace(result) != "" {
		t.Errorf("expected empty result for comment-only input, got %q", result)
	}
}

// TestStripSQLComments_InlineComments verifies stripSQLComments removes inline comments.
func TestStripSQLComments_InlineComments(t *testing.T) {
	input := "SELECT 1; -- inline comment\nSELECT 2;"
	result := stripSQLComments(input)
	if !strings.Contains(result, "SELECT 1;") {
		t.Error("expected SQL before inline comment to be preserved")
	}
	if strings.Contains(result, "inline comment") {
		t.Error("expected inline comment to be removed")
	}
	if !strings.Contains(result, "SELECT 2;") {
		t.Error("expected SQL after comment line to be preserved")
	}
}

// TestContainsWithout_NoOccurrence verifies containsWithout returns false when keyword is absent.
func TestContainsWithout_NoOccurrence(t *testing.T) {
	sql := "ALTER TABLE users ADD COLUMN name TEXT;"
	if containsWithout(sql, "CREATE TABLE", "IF NOT EXISTS") {
		t.Error("expected false when keyword is not present")
	}
}

// TestContainsWithout_WithQualifier verifies containsWithout returns false when qualifier is present.
func TestContainsWithout_WithQualifier(t *testing.T) {
	sql := "CREATE TABLE IF NOT EXISTS users (id UUID);"
	if containsWithout(sql, "CREATE TABLE", "IF NOT EXISTS") {
		t.Error("expected false when qualifier is present")
	}
}

// TestContainsWithout_WithoutQualifier verifies containsWithout returns true when qualifier is missing.
func TestContainsWithout_WithoutQualifier(t *testing.T) {
	sql := "CREATE TABLE users (id UUID);"
	if !containsWithout(sql, "CREATE TABLE", "IF NOT EXISTS") {
		t.Error("expected true when qualifier is missing")
	}
}

// TestContainsWithout_CommentsIgnored verifies containsWithout ignores keywords in comments.
func TestContainsWithout_CommentsIgnored(t *testing.T) {
	sql := "-- CREATE TABLE users\nCREATE TABLE IF NOT EXISTS users (id UUID);"
	if containsWithout(sql, "CREATE TABLE", "IF NOT EXISTS") {
		t.Error("expected false when bare keyword is only in a comment")
	}
}

// TestContainsWithout_MultipleOccurrences verifies detection when one occurrence lacks qualifier.
func TestContainsWithout_MultipleOccurrences(t *testing.T) {
	sql := "CREATE TABLE IF NOT EXISTS users (id UUID);\nCREATE TABLE roles (id UUID);"
	if !containsWithout(sql, "CREATE TABLE", "IF NOT EXISTS") {
		t.Error("expected true when second occurrence lacks qualifier")
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

	// Verify all tables are gone (except schema_migrations from migrate tool)
	var tableCount int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'public'
		  AND table_type = 'BASE TABLE'
		  AND table_name != 'schema_migrations'
	`).Scan(&tableCount)
	if err != nil {
		t.Fatalf("failed to count tables: %v", err)
	}
	if tableCount != 0 {
		// Print which tables remain for debugging
		rows, _ := pool.Query(ctx, `
			SELECT table_name
			FROM information_schema.tables
			WHERE table_schema = 'public'
			  AND table_type = 'BASE TABLE'
			  AND table_name != 'schema_migrations'
		`)
		var remaining []string
		for rows.Next() {
			var name string
			rows.Scan(&name)
			remaining = append(remaining, name)
		}
		rows.Close()
		t.Errorf("expected 0 tables after down migrations, got %d: %v", tableCount, remaining)
	}
}
