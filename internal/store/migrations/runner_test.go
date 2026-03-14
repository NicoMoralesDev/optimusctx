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
	if !reflect.DeepEqual(versions, []int{1}) {
		t.Fatalf("versions = %v, want [1]", versions)
	}

	assertTablesExist(t, db, "schema_migrations", "repositories", "directories", "files")
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
	assertIndexColumns(t, db, "files", []string{"repository_id", "ignore_status"})
	assertIndexColumns(t, db, "files", []string{"repository_id", "language"})
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
