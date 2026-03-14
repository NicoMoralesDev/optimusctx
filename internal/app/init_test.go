package app

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niccrow/optimusctx/internal/repository"

	_ "modernc.org/sqlite"
)

func TestInitServicePersistsRepositoryInventory(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, ".gitignore"), "*.tmp\nignored-dir/\n")
	writeRepoFile(t, filepath.Join(repoRoot, "cmd", "app.go"), "package main\n")
	writeRepoFile(t, filepath.Join(repoRoot, "ignored.tmp"), "ignored\n")
	writeRepoFile(t, filepath.Join(repoRoot, "ignored-dir", "nested.txt"), "ignored dir\n")
	writeRepoFile(t, filepath.Join(repoRoot, "dist", "bundle.js"), "console.log('ignored');\n")

	service := NewInitService()

	result, err := service.Init(context.Background(), filepath.Join(repoRoot, "cmd"))
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if result.RepositoryRoot != repoRoot {
		t.Fatalf("RepositoryRoot = %q, want %q", result.RepositoryRoot, repoRoot)
	}
	if result.StatePath != filepath.Join(repoRoot, ".optimusctx") {
		t.Fatalf("StatePath = %q", result.StatePath)
	}
	if result.SchemaVersion == 0 {
		t.Fatal("SchemaVersion should be non-zero")
	}
	if result.FileCount < 2 {
		t.Fatalf("FileCount = %d, want at least 2", result.FileCount)
	}
	if result.IncludedFiles == 0 {
		t.Fatal("IncludedFiles should be non-zero")
	}
	if result.IgnoredFiles == 0 {
		t.Fatal("IgnoredFiles should be non-zero")
	}

	db := openStateDatabase(t, filepath.Join(result.StatePath, "db.sqlite"))
	defer db.Close()

	assertRepositoryRow(t, db, repoRoot)
	assertIndexedFileRow(t, db, "cmd/app.go")
	assertIgnoredFileRow(t, db, "ignored.tmp", string(repository.IgnoreReasonGitIgnore))
	assertIgnoredDirectoryRow(t, db, "ignored-dir", string(repository.IgnoreReasonGitIgnore))
	assertIgnoredDirectoryRow(t, db, "dist", string(repository.IgnoreReasonBuiltinExclusion))
}

func TestInitServiceIsIdempotent(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, ".gitignore"), "*.tmp\n")
	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n")
	writeRepoFile(t, filepath.Join(repoRoot, "ignored.tmp"), "ignored\n")

	service := NewInitService()

	firstResult, err := service.Init(context.Background(), repoRoot)
	if err != nil {
		t.Fatalf("first Init() error = %v", err)
	}
	secondResult, err := service.Init(context.Background(), filepath.Join(repoRoot, "nested"))
	if err != nil {
		t.Fatalf("second Init() error = %v", err)
	}

	if firstResult.SchemaVersion != secondResult.SchemaVersion {
		t.Fatalf("SchemaVersion changed from %d to %d", firstResult.SchemaVersion, secondResult.SchemaVersion)
	}
	if firstResult.FileCount != secondResult.FileCount {
		t.Fatalf("FileCount changed from %d to %d", firstResult.FileCount, secondResult.FileCount)
	}

	db := openStateDatabase(t, filepath.Join(repoRoot, ".optimusctx", "db.sqlite"))
	defer db.Close()

	assertCount(t, db, `SELECT COUNT(*) FROM repositories`, 1)
	assertCount(t, db, `SELECT COUNT(*) FROM files`, secondResult.FileCount)
	assertCount(t, db, `SELECT COUNT(*) FROM directories`, secondResult.DirectoryCount)
}

func TestInitWorkflowReturnsRepositoryNotFound(t *testing.T) {
	service := NewInitService()

	_, err := service.Init(context.Background(), t.TempDir())
	if err == nil {
		t.Fatal("Init() expected error, got nil")
	}
	if !contains(err.Error(), repository.ErrRepositoryNotFound.Error()) {
		t.Fatalf("Init() error = %v, want repository not found", err)
	}
}

func TestInitUsesRefreshBaseline(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n")

	realRefresh := NewRefreshService()
	var seen RefreshRequest
	service := NewInitService()
	service.Refresh = func(ctx context.Context, request RefreshRequest) (RefreshResult, error) {
		seen = request
		return realRefresh.Refresh(ctx, request)
	}

	result, err := service.Init(context.Background(), filepath.Join(repoRoot, "nested"))
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if seen.Reason != repository.RefreshReasonInit {
		t.Fatalf("refresh reason = %q, want %q", seen.Reason, repository.RefreshReasonInit)
	}
	if !seen.ForceFull {
		t.Fatal("init should force the baseline refresh path")
	}
	if seen.StartPath != repoRoot {
		t.Fatalf("refresh start path = %q, want %q", seen.StartPath, repoRoot)
	}
	if result.FileCount == 0 || result.IncludedFiles == 0 {
		t.Fatalf("unexpected init result: %+v", result)
	}
}

func initRepo(t *testing.T) string {
	t.Helper()

	repoRoot := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, output)
	}

	if err := os.MkdirAll(filepath.Join(repoRoot, ".git", "info"), 0o755); err != nil {
		t.Fatalf("mkdir git info: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repoRoot, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	return repoRoot
}

func writeRepoFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}

func openStateDatabase(t *testing.T, path string) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open(%q) error = %v", path, err)
	}
	return db
}

func assertRepositoryRow(t *testing.T, db *sql.DB, rootPath string) {
	t.Helper()

	var storedRoot string
	var detectionMode string
	if err := db.QueryRow(`SELECT root_path, detection_mode FROM repositories`).Scan(&storedRoot, &detectionMode); err != nil {
		t.Fatalf("query repositories: %v", err)
	}
	if storedRoot != rootPath {
		t.Fatalf("stored root_path = %q, want %q", storedRoot, rootPath)
	}
	if detectionMode != repository.DetectionModeGit {
		t.Fatalf("detection_mode = %q, want %q", detectionMode, repository.DetectionModeGit)
	}
}

func assertIndexedFileRow(t *testing.T, db *sql.DB, path string) {
	t.Helper()

	var language string
	var sizeBytes int64
	var contentHash sql.NullString
	var lastIndexedAt sql.NullString
	var ignoreStatus string
	var ignoreReason sql.NullString
	if err := db.QueryRow(`SELECT language, size_bytes, content_hash, last_indexed_at, ignore_status, ignore_reason FROM files WHERE path = ?`, path).
		Scan(&language, &sizeBytes, &contentHash, &lastIndexedAt, &ignoreStatus, &ignoreReason); err != nil {
		t.Fatalf("query indexed file %q: %v", path, err)
	}
	if language != "go" {
		t.Fatalf("language = %q, want go", language)
	}
	if sizeBytes == 0 {
		t.Fatal("size_bytes should be non-zero")
	}
	if !contentHash.Valid || len(contentHash.String) != 64 {
		t.Fatalf("content_hash = %#v, want sha256", contentHash)
	}
	if !lastIndexedAt.Valid {
		t.Fatal("last_indexed_at should be present")
	}
	if ignoreStatus != string(repository.IgnoreStatusIncluded) {
		t.Fatalf("ignore_status = %q, want %q", ignoreStatus, repository.IgnoreStatusIncluded)
	}
	if ignoreReason.Valid {
		t.Fatalf("ignore_reason = %#v, want NULL for included file", ignoreReason)
	}
}

func assertIgnoredFileRow(t *testing.T, db *sql.DB, path string, reason string) {
	t.Helper()

	var ignoreStatus string
	var ignoreReason sql.NullString
	var contentHash sql.NullString
	var lastIndexedAt sql.NullString
	if err := db.QueryRow(`SELECT ignore_status, ignore_reason, content_hash, last_indexed_at FROM files WHERE path = ?`, path).
		Scan(&ignoreStatus, &ignoreReason, &contentHash, &lastIndexedAt); err != nil {
		t.Fatalf("query ignored file %q: %v", path, err)
	}
	if ignoreStatus != string(repository.IgnoreStatusIgnored) {
		t.Fatalf("ignore_status = %q, want %q", ignoreStatus, repository.IgnoreStatusIgnored)
	}
	if !ignoreReason.Valid || ignoreReason.String != reason {
		t.Fatalf("ignore_reason = %#v, want %q", ignoreReason, reason)
	}
	if contentHash.Valid {
		t.Fatalf("content_hash = %#v, want NULL", contentHash)
	}
	if lastIndexedAt.Valid {
		t.Fatalf("last_indexed_at = %#v, want NULL", lastIndexedAt)
	}
}

func assertIgnoredDirectoryRow(t *testing.T, db *sql.DB, path string, reason string) {
	t.Helper()

	var ignoreStatus string
	var ignoreReason sql.NullString
	if err := db.QueryRow(`SELECT ignore_status, ignore_reason FROM directories WHERE path = ?`, path).
		Scan(&ignoreStatus, &ignoreReason); err != nil {
		t.Fatalf("query ignored directory %q: %v", path, err)
	}
	if ignoreStatus != string(repository.IgnoreStatusIgnored) {
		t.Fatalf("ignore_status = %q, want %q", ignoreStatus, repository.IgnoreStatusIgnored)
	}
	if !ignoreReason.Valid || ignoreReason.String != reason {
		t.Fatalf("ignore_reason = %#v, want %q", ignoreReason, reason)
	}
}

func assertCount(t *testing.T, db *sql.DB, query string, want int) {
	t.Helper()

	var got int
	if err := db.QueryRow(query).Scan(&got); err != nil {
		t.Fatalf("count query %q: %v", query, err)
	}
	if got != want {
		t.Fatalf("count for %q = %d, want %d", query, got, want)
	}
}

func contains(haystack string, needle string) bool {
	return strings.Contains(haystack, needle)
}
