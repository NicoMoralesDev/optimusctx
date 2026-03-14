package app

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
	"github.com/niccrow/optimusctx/internal/store/sqlite"
)

func TestRefreshService(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "README.md"), "hello\n")
	writeRepoFile(t, filepath.Join(repoRoot, "cmd", "app.go"), "package main\n")

	service := NewRefreshService()

	first, err := service.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	})
	if err != nil {
		t.Fatalf("Refresh() baseline error = %v", err)
	}

	writeRepoFile(t, filepath.Join(repoRoot, "cmd", "app.go"), "package main\n\nfunc refreshed() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "helper.go"), "package pkg\n")

	second, err := service.Refresh(context.Background(), RefreshRequest{
		StartPath: filepath.Join(repoRoot, "cmd"),
		Reason:    repository.RefreshReasonManual,
	})
	if err != nil {
		t.Fatalf("Refresh() incremental error = %v", err)
	}

	if first.RepositoryRoot != repoRoot || second.RepositoryRoot != repoRoot {
		t.Fatalf("repository roots = %q / %q, want %q", first.RepositoryRoot, second.RepositoryRoot, repoRoot)
	}
	if second.ChangedFiles != 2 {
		t.Fatalf("ChangedFiles = %d, want 2", second.ChangedFiles)
	}
	if second.UnchangedFiles != 1 {
		t.Fatalf("UnchangedFiles = %d, want 1", second.UnchangedFiles)
	}
	if second.AffectedDirectories != 3 {
		t.Fatalf("AffectedDirectories = %d, want 3", second.AffectedDirectories)
	}
	if second.Generation != first.Generation+1 {
		t.Fatalf("generation = %d, want %d", second.Generation, first.Generation+1)
	}
	if second.FreshnessStatus != repository.FreshnessStatusFresh {
		t.Fatalf("freshness status = %q, want %q", second.FreshnessStatus, repository.FreshnessStatusFresh)
	}
}

func TestNoOpRefresh(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n")

	service := NewRefreshService()

	first, err := service.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	})
	if err != nil {
		t.Fatalf("Refresh() baseline error = %v", err)
	}

	second, err := service.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonManual,
	})
	if err != nil {
		t.Fatalf("Refresh() no-op error = %v", err)
	}

	if second.ChangedFiles != 0 {
		t.Fatalf("ChangedFiles = %d, want 0", second.ChangedFiles)
	}
	if second.UnchangedFiles == 0 {
		t.Fatal("UnchangedFiles should be non-zero")
	}
	if second.AffectedDirectories != 0 {
		t.Fatalf("AffectedDirectories = %d, want 0", second.AffectedDirectories)
	}
	if second.Generation != first.Generation+1 {
		t.Fatalf("generation = %d, want %d", second.Generation, first.Generation+1)
	}
}

func TestSnapshotEquivalence(t *testing.T) {
	baselineRepo := initRepo(t)
	writeRepoFile(t, filepath.Join(baselineRepo, "README.md"), "final\n")
	writeRepoFile(t, filepath.Join(baselineRepo, "pkg", "helper.go"), "package pkg\n")
	writeRepoFile(t, filepath.Join(baselineRepo, "pkg", "main.go"), "package pkg\n")

	incrementalRepo := initRepo(t)
	writeRepoFile(t, filepath.Join(incrementalRepo, "README.md"), "start\n")
	writeRepoFile(t, filepath.Join(incrementalRepo, "pkg", "main.go"), "package pkg\n")

	service := NewRefreshService()

	if _, err := service.Refresh(context.Background(), RefreshRequest{
		StartPath: baselineRepo,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() baseline repo error = %v", err)
	}

	if _, err := service.Refresh(context.Background(), RefreshRequest{
		StartPath: incrementalRepo,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() incremental repo initial error = %v", err)
	}

	writeRepoFile(t, filepath.Join(incrementalRepo, "README.md"), "final\n")
	writeRepoFile(t, filepath.Join(incrementalRepo, "pkg", "helper.go"), "package pkg\n")

	if _, err := service.Refresh(context.Background(), RefreshRequest{
		StartPath: incrementalRepo,
		Reason:    repository.RefreshReasonManual,
	}); err != nil {
		t.Fatalf("Refresh() incremental repo follow-up error = %v", err)
	}

	baselineSnapshot := loadSnapshotForRepo(t, baselineRepo)
	incrementalSnapshot := loadSnapshotForRepo(t, incrementalRepo)

	if !reflect.DeepEqual(normalizeSnapshot(baselineSnapshot), normalizeSnapshot(incrementalSnapshot)) {
		t.Fatalf("snapshots differ:\nbaseline=%#v\nincremental=%#v", normalizeSnapshot(baselineSnapshot), normalizeSnapshot(incrementalSnapshot))
	}
}

func loadSnapshotForRepo(t *testing.T, repoRoot string) repository.RepositorySnapshot {
	t.Helper()

	layout, err := state.ResolveLayout(repoRoot)
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, err := sqlite.OpenOrCreateStore(context.Background(), layout, repository.DetectionModeGit)
	if err != nil {
		t.Fatalf("OpenOrCreateStore() error = %v", err)
	}
	defer store.Close()

	record, err := store.UpsertRepository(context.Background(), testRepoRoot(repoRoot), serviceNow())
	if err != nil {
		t.Fatalf("UpsertRepository() error = %v", err)
	}

	snapshot, err := store.LoadRepositorySnapshot(context.Background(), record.ID)
	if err != nil {
		t.Fatalf("LoadRepositorySnapshot() error = %v", err)
	}
	return snapshot
}

func normalizeSnapshot(snapshot repository.RepositorySnapshot) repository.RepositorySnapshot {
	snapshot.Repository.RepositoryID = 0
	snapshot.Repository.RootPath = ""
	snapshot.Repository.GitCommonDir = ""
	snapshot.Repository.GitHeadRef = ""
	snapshot.Repository.GitHeadCommit = ""
	snapshot.Repository.CurrentGeneration = 0
	snapshot.Repository.LastRefreshGeneration = 0
	snapshot.Repository.LastRefreshStartedAt = serviceNow()
	snapshot.Repository.LastRefreshCompletedAt = serviceNow()
	snapshot.Repository.LastRefreshReason = ""
	snapshot.Repository.LastRefreshStatus = repository.RefreshRunStatusSuccess
	for index := range snapshot.Directories {
		snapshot.Directories[index].DiscoveredAt = serviceNow()
		snapshot.Directories[index].LastRefreshedAt = serviceNow()
		snapshot.Directories[index].LastRefreshGeneration = 0
	}
	for index := range snapshot.Files {
		snapshot.Files[index].DiscoveredAt = serviceNow()
		snapshot.Files[index].UpdatedAt = serviceNow()
		snapshot.Files[index].LastIndexedAt = serviceNow()
		snapshot.Files[index].FilesystemModTime = serviceNow()
		snapshot.Files[index].LastSeenGeneration = 0
		snapshot.Files[index].RefreshRunID = 0
		snapshot.Files[index].UpdatedReason = ""
	}
	return snapshot
}

func testRepoRoot(repoRoot string) repository.RepositoryRoot {
	return repository.RepositoryRoot{
		RootPath:      repoRoot,
		DetectionMode: repository.DetectionModeGit,
		Fingerprint: repository.RepositoryFingerprint{
			RootPath:      repoRoot,
			GitCommonDir:  filepath.Join(repoRoot, ".git"),
			GitHeadRef:    "refs/heads/main",
			GitHeadCommit: "0123456789abcdef0123456789abcdef01234567",
		},
	}
}

func serviceNow() time.Time {
	return time.Date(2026, 3, 14, 22, 0, 0, 0, time.UTC)
}
