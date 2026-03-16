package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResolveLayout(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()

	layout, err := ResolveLayout(repoRoot)
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	if layout.RepoRoot != repoRoot {
		t.Fatalf("RepoRoot = %q, want %q", layout.RepoRoot, repoRoot)
	}

	if layout.StateDir != filepath.Join(repoRoot, DirectoryName) {
		t.Fatalf("StateDir = %q", layout.StateDir)
	}
	if layout.DatabasePath != filepath.Join(repoRoot, DirectoryName, DatabaseFilename) {
		t.Fatalf("DatabasePath = %q", layout.DatabasePath)
	}
	if layout.MetadataPath != filepath.Join(repoRoot, DirectoryName, MetadataFilename) {
		t.Fatalf("MetadataPath = %q", layout.MetadataPath)
	}
	if layout.EvalDir != filepath.Join(repoRoot, DirectoryName, "eval") {
		t.Fatalf("EvalDir = %q", layout.EvalDir)
	}
	if layout.EvalRunDir(12) != filepath.Join(repoRoot, DirectoryName, "eval", "run-000012") {
		t.Fatalf("EvalRunDir = %q", layout.EvalRunDir(12))
	}
}

func TestLayoutEnsureCreatesMetadata(t *testing.T) {
	t.Parallel()

	layout, err := ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	now := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	metadata, err := layout.Ensure("git", 1, now)
	if err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	for _, dir := range []string{layout.StateDir, layout.EvalDir, layout.LogsDir, layout.TmpDir} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("Stat(%q) error = %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("%q is not a directory", dir)
		}
	}

	stored, err := layout.ReadMetadata()
	if err != nil {
		t.Fatalf("ReadMetadata() error = %v", err)
	}

	if metadata != stored {
		t.Fatalf("stored metadata mismatch: %#v != %#v", stored, metadata)
	}

	if stored.FormatVersion != CurrentFormatVersion {
		t.Fatalf("FormatVersion = %d", stored.FormatVersion)
	}
	if stored.RepoRoot != layout.RepoRoot {
		t.Fatalf("RepoRoot = %q", stored.RepoRoot)
	}
	if stored.RepoDetectionMode != "git" {
		t.Fatalf("RepoDetectionMode = %q", stored.RepoDetectionMode)
	}
	if stored.CreatedAt == "" || stored.UpdatedAt == "" || stored.RuntimeVersion == "" {
		t.Fatalf("metadata fields missing: %#v", stored)
	}
	if stored.SchemaVersion != 1 {
		t.Fatalf("SchemaVersion = %d", stored.SchemaVersion)
	}
}

func TestLayoutEnsureIsIdempotent(t *testing.T) {
	t.Parallel()

	layout, err := ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	firstTime := time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)
	first, err := layout.Ensure("git", 1, firstTime)
	if err != nil {
		t.Fatalf("Ensure() first error = %v", err)
	}

	secondTime := firstTime.Add(2 * time.Hour)
	second, err := layout.Ensure("existing-state", 3, secondTime)
	if err != nil {
		t.Fatalf("Ensure() second error = %v", err)
	}

	if second.CreatedAt != first.CreatedAt {
		t.Fatalf("CreatedAt changed from %q to %q", first.CreatedAt, second.CreatedAt)
	}
	if second.UpdatedAt != secondTime.Format(time.RFC3339) {
		t.Fatalf("UpdatedAt = %q", second.UpdatedAt)
	}
	if second.RepoRoot != layout.RepoRoot {
		t.Fatalf("RepoRoot = %q", second.RepoRoot)
	}
	if second.RepoDetectionMode != "existing-state" {
		t.Fatalf("RepoDetectionMode = %q", second.RepoDetectionMode)
	}
	if second.SchemaVersion != 3 {
		t.Fatalf("SchemaVersion = %d", second.SchemaVersion)
	}
}
