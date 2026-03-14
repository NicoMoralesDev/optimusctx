package sqlite

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
	"github.com/niccrow/optimusctx/internal/store/migrations"
)

func TestOpenOrCreateStoreInitializesEmptyDatabase(t *testing.T) {
	t.Parallel()

	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, err := OpenOrCreateStore(context.Background(), layout, repository.DetectionModeGit)
	if err != nil {
		t.Fatalf("OpenOrCreateStore() error = %v", err)
	}
	defer store.Close()

	if store.SchemaVersion() != migrations.CurrentVersion() {
		t.Fatalf("SchemaVersion() = %d, want %d", store.SchemaVersion(), migrations.CurrentVersion())
	}

	if _, err := os.Stat(layout.DatabasePath); err != nil {
		t.Fatalf("Stat(%q) error = %v", layout.DatabasePath, err)
	}

	metadata, err := layout.ReadMetadata()
	if err != nil {
		t.Fatalf("ReadMetadata() error = %v", err)
	}
	if metadata.SchemaVersion != migrations.CurrentVersion() {
		t.Fatalf("metadata schema version = %d", metadata.SchemaVersion)
	}

	var versionCount int
	if err := store.DB().QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&versionCount); err != nil {
		t.Fatalf("QueryRow(schema_migrations) error = %v", err)
	}
	if versionCount != migrations.CurrentVersion() {
		t.Fatalf("version count = %d, want %d", versionCount, migrations.CurrentVersion())
	}
}

func TestOpenOrCreateStoreIsIdempotent(t *testing.T) {
	t.Parallel()

	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	firstStore, err := OpenOrCreateStore(context.Background(), layout, repository.DetectionModeGit)
	if err != nil {
		t.Fatalf("first OpenOrCreateStore() error = %v", err)
	}

	root := repository.RepositoryRoot{
		RootPath:      layout.RepoRoot,
		DetectionMode: repository.DetectionModeGit,
		Fingerprint: repository.RepositoryFingerprint{
			RootPath:      layout.RepoRoot,
			GitCommonDir:  layout.RepoRoot + "/.git",
			GitHeadRef:    "main",
			GitHeadCommit: "0123456789abcdef0123456789abcdef01234567",
		},
	}

	record, err := firstStore.UpsertRepository(context.Background(), root, time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC))
	if err != nil {
		firstStore.Close()
		t.Fatalf("UpsertRepository() error = %v", err)
	}
	if record.ID == 0 {
		firstStore.Close()
		t.Fatal("repository record ID should be non-zero")
	}
	if err := firstStore.Close(); err != nil {
		t.Fatalf("firstStore.Close() error = %v", err)
	}

	secondStore, err := OpenOrCreateStore(context.Background(), layout, repository.DetectionModeOptimusCtxState)
	if err != nil {
		t.Fatalf("second OpenOrCreateStore() error = %v", err)
	}
	defer secondStore.Close()

	record, err = secondStore.UpsertRepository(context.Background(), root, time.Date(2026, 3, 14, 14, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("UpsertRepository() second error = %v", err)
	}
	if record.ID == 0 {
		t.Fatal("repository record ID should remain non-zero")
	}

	var repositoryCount int
	if err := secondStore.DB().QueryRow(`SELECT COUNT(*) FROM repositories`).Scan(&repositoryCount); err != nil {
		t.Fatalf("QueryRow(repositories) error = %v", err)
	}
	if repositoryCount != 1 {
		t.Fatalf("repository count = %d, want 1", repositoryCount)
	}
}

func TestSQLiteStoreReportsCorruptDatabase(t *testing.T) {
	t.Parallel()

	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}
	if _, err := layout.Ensure(repository.DetectionModeGit, 0, time.Now().UTC()); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	if err := os.WriteFile(layout.DatabasePath, []byte("not a sqlite database"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", layout.DatabasePath, err)
	}

	store, err := OpenOrCreateStore(context.Background(), layout, repository.DetectionModeGit)
	if err == nil {
		store.Close()
		t.Fatal("OpenOrCreateStore() expected error, got nil")
	}
}
