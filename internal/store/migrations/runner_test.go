package migrations

import (
	"context"
	"database/sql"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMigrationRunnerAppliesFreshDatabase(t *testing.T) {
	t.Parallel()

	db := openTestDatabase(t)
	defer db.Close()

	runner, err := NewRunner()
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}

	if err := runner.Apply(context.Background(), db); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	versions := appliedVersions(t, db)
	if !reflect.DeepEqual(versions, []int{1, 2, 3, 4, 5, 6, 7}) {
		t.Fatalf("versions = %v, want [1 2 3 4 5 6 7]", versions)
	}

	assertTablesExist(t, db, "schema_migrations", "repositories", "directories", "files", "refresh_runs", "refresh_file_events", "file_extractions", "symbols", "eval_runs", "eval_steps", "eval_artifacts", "benchmark_runs", "benchmark_lane_samples", "benchmark_lane_metrics", "benchmark_evidence_bundles", "benchmark_evidence_lane_summaries", "benchmark_evidence_attributions")
}

func TestApplyMigrationsIsNoOpWhenAlreadyCurrent(t *testing.T) {
	t.Parallel()

	db := openTestDatabase(t)
	defer db.Close()

	ctx := context.Background()
	if err := Apply(ctx, db); err != nil {
		t.Fatalf("first Apply() error = %v", err)
	}
	firstVersions := appliedVersions(t, db)

	if err := Apply(ctx, db); err != nil {
		t.Fatalf("second Apply() error = %v", err)
	}
	secondVersions := appliedVersions(t, db)

	if !reflect.DeepEqual(firstVersions, secondVersions) {
		t.Fatalf("versions changed between runs: %v != %v", firstVersions, secondVersions)
	}
}

func TestApplyMigrationsCreatesRequiredIndexes(t *testing.T) {
	t.Parallel()

	db := openTestDatabase(t)
	defer db.Close()

	if err := Apply(context.Background(), db); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	assertIndexColumns(t, db, "files", []string{"repository_id", "directory_path"})
	assertIndexColumns(t, db, "files", []string{"repository_id", "path"})
	assertIndexColumns(t, db, "files", []string{"repository_id", "ignore_status"})
	assertIndexColumns(t, db, "files", []string{"repository_id", "language"})
	assertIndexColumns(t, db, "directories", []string{"repository_id", "path"})
	assertIndexColumns(t, db, "refresh_runs", []string{"repository_id", "started_at"})
	assertIndexColumns(t, db, "refresh_runs", []string{"repository_id", "status"})
	assertIndexColumns(t, db, "refresh_file_events", []string{"refresh_run_id", "path"})
	assertIndexColumns(t, db, "refresh_file_events", []string{"repository_id", "path"})
	assertIndexColumns(t, db, "file_extractions", []string{"repository_id", "path"})
	assertIndexColumns(t, db, "file_extractions", []string{"repository_id", "coverage_state"})
	assertIndexColumns(t, db, "file_extractions", []string{"repository_id", "language"})
	assertIndexColumns(t, db, "file_extractions", []string{"repository_id", "source_generation"})
	assertIndexColumns(t, db, "symbols", []string{"repository_id", "name"})
	assertIndexColumns(t, db, "symbols", []string{"repository_id", "kind"})
	assertIndexColumns(t, db, "symbols", []string{"file_id", "ordinal"})
	assertIndexColumns(t, db, "symbols", []string{"repository_id", "path", "ordinal"})
	assertIndexColumns(t, db, "symbols", []string{"repository_id", "qualified_name"})
	assertIndexColumns(t, db, "eval_runs", []string{"repository_id", "started_at"})
	assertIndexColumns(t, db, "eval_runs", []string{"repository_id", "scenario_id", "started_at"})
	assertIndexColumns(t, db, "eval_steps", []string{"eval_run_id", "ordinal"})
	assertIndexColumns(t, db, "eval_artifacts", []string{"eval_run_id", "step_id"})
	assertIndexColumns(t, db, "benchmark_runs", []string{"repository_id", "suite_id", "started_at"})
	assertIndexColumns(t, db, "benchmark_runs", []string{"repository_id", "arm_kind", "started_at"})
	assertIndexColumns(t, db, "benchmark_lane_samples", []string{"benchmark_run_id", "lane"})
	assertIndexColumns(t, db, "benchmark_lane_metrics", []string{"benchmark_lane_sample_id", "metric_name"})
}

func TestApplyMigrationsAddsRefreshStateColumns(t *testing.T) {
	t.Parallel()

	db := openTestDatabase(t)
	defer db.Close()

	if err := Apply(context.Background(), db); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	assertColumnExists(t, db, "repositories", "last_refresh_started_at")
	assertColumnExists(t, db, "repositories", "last_refresh_completed_at")
	assertColumnExists(t, db, "repositories", "last_refresh_reason")
	assertColumnExists(t, db, "repositories", "last_refresh_status")
	assertColumnExists(t, db, "repositories", "freshness_status")
	assertColumnExists(t, db, "repositories", "freshness_reason")
	assertColumnExists(t, db, "repositories", "current_refresh_generation")
	assertColumnExists(t, db, "repositories", "last_refresh_generation")
	assertColumnExists(t, db, "directories", "subtree_fingerprint")
	assertColumnExists(t, db, "directories", "included_file_count")
	assertColumnExists(t, db, "directories", "included_directory_count")
	assertColumnExists(t, db, "directories", "total_size_bytes")
	assertColumnExists(t, db, "directories", "last_refreshed_at")
	assertColumnExists(t, db, "directories", "last_refresh_generation")
	assertColumnExists(t, db, "files", "last_seen_generation")
	assertColumnExists(t, db, "files", "refresh_run_id")
	assertColumnExists(t, db, "files", "updated_reason")
	assertColumnExists(t, db, "file_extractions", "path")
	assertColumnExists(t, db, "file_extractions", "language")
	assertColumnExists(t, db, "file_extractions", "adapter_name")
	assertColumnExists(t, db, "file_extractions", "grammar_version")
	assertColumnExists(t, db, "file_extractions", "source_content_hash")
	assertColumnExists(t, db, "file_extractions", "source_generation")
	assertColumnExists(t, db, "file_extractions", "coverage_state")
	assertColumnExists(t, db, "file_extractions", "coverage_reason")
	assertColumnExists(t, db, "file_extractions", "parser_error_count")
	assertColumnExists(t, db, "file_extractions", "has_error_nodes")
	assertColumnExists(t, db, "file_extractions", "symbol_count")
	assertColumnExists(t, db, "file_extractions", "top_level_symbol_count")
	assertColumnExists(t, db, "file_extractions", "max_symbol_depth")
	assertColumnExists(t, db, "file_extractions", "extracted_at")
	assertColumnExists(t, db, "symbols", "stable_key")
	assertColumnExists(t, db, "symbols", "parent_symbol_id")
	assertColumnExists(t, db, "symbols", "qualified_name")
	assertColumnExists(t, db, "symbols", "ordinal")
	assertColumnExists(t, db, "symbols", "depth")
	assertColumnExists(t, db, "symbols", "name_start_byte")
	assertColumnExists(t, db, "symbols", "name_end_byte")
	assertColumnExists(t, db, "symbols", "signature_start_byte")
	assertColumnExists(t, db, "symbols", "signature_end_byte")
	assertColumnExists(t, db, "symbols", "is_exported")
	assertColumnExists(t, db, "eval_runs", "scenario_id")
	assertColumnExists(t, db, "eval_runs", "artifact_root")
	assertColumnExists(t, db, "eval_steps", "step_id")
	assertColumnExists(t, db, "eval_steps", "stdout_path")
	assertColumnExists(t, db, "eval_artifacts", "artifact_id")
	assertColumnExists(t, db, "eval_artifacts", "stored_path")
	assertColumnExists(t, db, "benchmark_runs", "suite_id")
	assertColumnExists(t, db, "benchmark_runs", "arm_kind")
	assertColumnExists(t, db, "benchmark_lane_samples", "lane")
	assertColumnExists(t, db, "benchmark_lane_samples", "elapsed_ms")
	assertColumnExists(t, db, "benchmark_lane_metrics", "metric_name")
	assertColumnExists(t, db, "benchmark_lane_metrics", "value_text")
}

func TestExtractionSchemaContracts(t *testing.T) {
	t.Parallel()

	db := openTestDatabase(t)
	defer db.Close()

	if err := Apply(context.Background(), db); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	assertTablesExist(t, db, "file_extractions", "symbols")
	assertColumnExists(t, db, "file_extractions", "file_id")
	assertColumnExists(t, db, "file_extractions", "repository_id")
	assertColumnExists(t, db, "file_extractions", "refresh_run_id")
	assertColumnExists(t, db, "symbols", "file_extraction_id")
	assertColumnExists(t, db, "symbols", "file_id")
	assertColumnExists(t, db, "symbols", "repository_id")

	assertIndexColumns(t, db, "file_extractions", []string{"repository_id", "path"})
	assertIndexColumns(t, db, "file_extractions", []string{"repository_id", "coverage_state"})
	assertIndexColumns(t, db, "symbols", []string{"repository_id", "name"})
	assertIndexColumns(t, db, "symbols", []string{"repository_id", "path", "ordinal"})
}

func TestMigrationRunnerRollsBackOnFailure(t *testing.T) {
	t.Parallel()

	db := openTestDatabase(t)
	defer db.Close()

	runner := Runner{
		migrations: []Migration{
			{Version: 1, Name: "good", SQL: `CREATE TABLE good_items(id INTEGER PRIMARY KEY);`},
			{Version: 2, Name: "bad", SQL: `CREATE TABLE broken(`},
		},
	}

	err := runner.Apply(context.Background(), db)
	if err == nil {
		t.Fatal("Apply() expected error, got nil")
	}

	assertTableMissing(t, db, "good_items")
	if versions := appliedVersions(t, db); len(versions) != 0 {
		t.Fatalf("versions = %v, want none recorded", versions)
	}
}

func openTestDatabase(t *testing.T) *sql.DB {
	t.Helper()

	path := filepath.Join(t.TempDir(), "test.sqlite")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}

	return db
}

func appliedVersions(t *testing.T, db *sql.DB) []int {
	t.Helper()

	rows, err := db.Query(`SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		if isMissingTableError(err) {
			return nil
		}
		t.Fatalf("Query(schema_migrations) error = %v", err)
	}
	defer rows.Close()

	var versions []int
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			t.Fatalf("rows.Scan() error = %v", err)
		}
		versions = append(versions, version)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err() = %v", err)
	}

	return versions
}

func isMissingTableError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "no such table: schema_migrations")
}

func assertTablesExist(t *testing.T, db *sql.DB, tableNames ...string) {
	t.Helper()

	for _, tableName := range tableNames {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, tableName).Scan(&count); err != nil {
			t.Fatalf("QueryRow(%q) error = %v", tableName, err)
		}
		if count != 1 {
			t.Fatalf("table %q missing", tableName)
		}
	}
}

func assertTableMissing(t *testing.T, db *sql.DB, tableName string) {
	t.Helper()

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, tableName).Scan(&count); err != nil {
		t.Fatalf("QueryRow(%q) error = %v", tableName, err)
	}
	if count != 0 {
		t.Fatalf("table %q should not exist", tableName)
	}
}

func assertIndexColumns(t *testing.T, db *sql.DB, tableName string, expected []string) {
	t.Helper()

	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type = 'index' AND tbl_name = ?`, tableName)
	if err != nil {
		t.Fatalf("Query(index list) error = %v", err)
	}
	defer rows.Close()

	var indexNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("rows.Scan() error = %v", err)
		}
		indexNames = append(indexNames, name)
	}

	for _, indexName := range indexNames {
		indexRows, err := db.Query(`SELECT name FROM pragma_index_info(?) ORDER BY seqno`, indexName)
		if err != nil {
			t.Fatalf("Query(pragma_index_info %q) error = %v", indexName, err)
		}

		var columns []string
		for indexRows.Next() {
			var column string
			if err := indexRows.Scan(&column); err != nil {
				indexRows.Close()
				t.Fatalf("indexRows.Scan() error = %v", err)
			}
			columns = append(columns, column)
		}
		if err := indexRows.Err(); err != nil {
			indexRows.Close()
			t.Fatalf("indexRows.Err() = %v", err)
		}
		indexRows.Close()

		if reflect.DeepEqual(columns, expected) {
			return
		}
	}

	t.Fatalf("index with columns %v not found on %s", expected, tableName)
}

func assertColumnExists(t *testing.T, db *sql.DB, tableName, columnName string) {
	t.Helper()

	rows, err := db.Query(`SELECT name FROM pragma_table_info(?)`, tableName)
	if err != nil {
		t.Fatalf("Query(pragma_table_info %q) error = %v", tableName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("rows.Scan() error = %v", err)
		}
		if name == columnName {
			return
		}
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err() = %v", err)
	}

	t.Fatalf("column %q missing from %s", columnName, tableName)
}
