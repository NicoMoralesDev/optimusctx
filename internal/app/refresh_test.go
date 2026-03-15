package app

import (
	"context"
	"errors"
	"os"
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
	fixture := newRefreshFixture(t)

	first, err := fixture.service.Refresh(context.Background(), RefreshRequest{
		StartPath: fixture.repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	})
	if err != nil {
		t.Fatalf("Refresh() baseline error = %v", err)
	}

	second, err := fixture.service.Refresh(context.Background(), RefreshRequest{
		StartPath: fixture.repoRoot,
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

func TestTrackedMutationRefreshCounts(t *testing.T) {
	fixture := newRefreshFixture(t)

	if _, err := fixture.service.Refresh(context.Background(), RefreshRequest{
		StartPath: fixture.repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() baseline error = %v", err)
	}

	fixture.applyTrackedMutations(t)

	mutated, err := fixture.service.Refresh(context.Background(), RefreshRequest{
		StartPath: fixture.repoRoot,
		Reason:    repository.RefreshReasonManual,
	})
	if err != nil {
		t.Fatalf("Refresh() mutation error = %v", err)
	}

	if mutated.AddedFiles != 1 {
		t.Fatalf("AddedFiles = %d, want 1", mutated.AddedFiles)
	}
	if mutated.ChangedContentFiles != 2 {
		t.Fatalf("ChangedContentFiles = %d, want 2", mutated.ChangedContentFiles)
	}
	if mutated.DeletedFiles != 1 {
		t.Fatalf("DeletedFiles = %d, want 1", mutated.DeletedFiles)
	}
	if mutated.MovedFiles != 1 {
		t.Fatalf("MovedFiles = %d, want 1", mutated.MovedFiles)
	}
	if mutated.NewlyIgnoredFiles != 1 {
		t.Fatalf("NewlyIgnoredFiles = %d, want 1", mutated.NewlyIgnoredFiles)
	}
	if mutated.UnchangedFiles != 1 {
		t.Fatalf("UnchangedFiles = %d, want 1", mutated.UnchangedFiles)
	}

	noOpAfterMutation, err := fixture.service.Refresh(context.Background(), RefreshRequest{
		StartPath: fixture.repoRoot,
		Reason:    repository.RefreshReasonManual,
	})
	if err != nil {
		t.Fatalf("Refresh() no-op after mutation error = %v", err)
	}

	if noOpAfterMutation.ChangedFiles != 0 {
		t.Fatalf("ChangedFiles = %d, want 0", noOpAfterMutation.ChangedFiles)
	}
	if noOpAfterMutation.UnchangedFiles != 5 {
		t.Fatalf("UnchangedFiles = %d, want 5", noOpAfterMutation.UnchangedFiles)
	}
}

func TestRefreshServiceFailureLeavesLastGoodSnapshot(t *testing.T) {
	fixture := newRefreshFixture(t)

	initial, err := fixture.service.Refresh(context.Background(), RefreshRequest{
		StartPath: fixture.repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	})
	if err != nil {
		t.Fatalf("Refresh() baseline error = %v", err)
	}

	beforeFailure := loadSnapshotForRepo(t, fixture.repoRoot)
	writeRepoFile(t, filepath.Join(fixture.repoRoot, "main.go"), "package main\n\nfunc degraded() {}\n")

	failed, err := fixture.service.Refresh(context.Background(), RefreshRequest{
		StartPath: fixture.repoRoot,
		Reason:    repository.RefreshReasonManual,
		InjectFailure: func(stage string) error {
			if stage == "after_files" {
				return errors.New("forced after file updates")
			}
			return nil
		},
	})
	if err == nil || err.Error() != "apply refresh plan: forced after file updates" {
		t.Fatalf("Refresh() error = %v, want injected failure", err)
	}
	if failed.RepositoryRoot != fixture.repoRoot {
		t.Fatalf("RepositoryRoot = %q, want %q", failed.RepositoryRoot, fixture.repoRoot)
	}
	if failed.Generation != initial.Generation+1 {
		t.Fatalf("Generation = %d, want %d", failed.Generation, initial.Generation+1)
	}
	if failed.FreshnessStatus != repository.FreshnessStatusPartiallyDegraded {
		t.Fatalf("FreshnessStatus = %q, want %q", failed.FreshnessStatus, repository.FreshnessStatusPartiallyDegraded)
	}
	if failed.ChangedContentFiles != 1 {
		t.Fatalf("ChangedContentFiles = %d, want 1", failed.ChangedContentFiles)
	}

	afterFailure := loadSnapshotForRepo(t, fixture.repoRoot)
	if !reflect.DeepEqual(beforeFailure.Files, afterFailure.Files) {
		t.Fatalf("files changed after failed refresh:\nbefore=%#v\nafter=%#v", beforeFailure.Files, afterFailure.Files)
	}
	if !reflect.DeepEqual(beforeFailure.Directories, afterFailure.Directories) {
		t.Fatalf("directories changed after failed refresh:\nbefore=%#v\nafter=%#v", beforeFailure.Directories, afterFailure.Directories)
	}
	if afterFailure.Repository.LastRefreshGeneration != initial.Generation {
		t.Fatalf("LastRefreshGeneration = %d, want %d", afterFailure.Repository.LastRefreshGeneration, initial.Generation)
	}
	if afterFailure.Repository.CurrentGeneration != initial.Generation+1 {
		t.Fatalf("CurrentGeneration = %d, want %d", afterFailure.Repository.CurrentGeneration, initial.Generation+1)
	}
	if afterFailure.Repository.FreshnessStatus != repository.FreshnessStatusPartiallyDegraded {
		t.Fatalf("FreshnessStatus = %q, want %q", afterFailure.Repository.FreshnessStatus, repository.FreshnessStatusPartiallyDegraded)
	}

	recovered, err := fixture.service.Refresh(context.Background(), RefreshRequest{
		StartPath: fixture.repoRoot,
		Reason:    repository.RefreshReasonManual,
	})
	if err != nil {
		t.Fatalf("Refresh() recovery error = %v", err)
	}
	if recovered.Generation != initial.Generation+2 {
		t.Fatalf("Generation = %d, want %d", recovered.Generation, initial.Generation+2)
	}
	if recovered.FreshnessStatus != repository.FreshnessStatusFresh {
		t.Fatalf("FreshnessStatus = %q, want %q", recovered.FreshnessStatus, repository.FreshnessStatusFresh)
	}
	if recovered.ChangedContentFiles != 1 {
		t.Fatalf("ChangedContentFiles = %d, want 1", recovered.ChangedContentFiles)
	}

	afterRecovery := loadSnapshotForRepo(t, fixture.repoRoot)
	mainFile := persistedSnapshotFileByPath(afterRecovery.Files, "main.go")
	if mainFile.ContentHash == persistedSnapshotFileByPath(beforeFailure.Files, "main.go").ContentHash {
		t.Fatal("main.go content hash did not update after successful recovery refresh")
	}
	if afterRecovery.Repository.FreshnessStatus != repository.FreshnessStatusFresh {
		t.Fatalf("FreshnessStatus = %q, want %q", afterRecovery.Repository.FreshnessStatus, repository.FreshnessStatusFresh)
	}
	if afterRecovery.Repository.LastRefreshGeneration != recovered.Generation {
		t.Fatalf("LastRefreshGeneration = %d, want %d", afterRecovery.Repository.LastRefreshGeneration, recovered.Generation)
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

func TestExtractionPersistence(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc Alpha() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "helper.go"), "package main\n\nfunc Helper() {}\n")

	service := NewRefreshService()
	if _, err := service.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() initial error = %v", err)
	}

	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc Beta() {}\n")
	result, err := service.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonManual,
	})
	if err != nil {
		t.Fatalf("Refresh() edit error = %v", err)
	}

	artifacts := loadArtifactsForRepo(t, repoRoot)
	if artifacts.extractions["main.go"].CoverageState != repository.ExtractionCoverageStateSupported {
		t.Fatalf("main.go coverage = %q", artifacts.extractions["main.go"].CoverageState)
	}
	if !reflect.DeepEqual(symbolNamesForPath(artifacts.symbols, "main.go"), []string{"main", "Beta"}) {
		t.Fatalf("main.go symbols = %+v", artifacts.symbols["main.go"])
	}
	if artifacts.extractions["helper.go"].SourceGeneration == result.Generation {
		t.Fatalf("helper.go should remain untouched: %+v", artifacts.extractions["helper.go"])
	}
}

func TestMoveReplacesArtifacts(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "move.go"), "package main\n\nfunc MoveMe() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "stable.go"), "package main\n\nfunc Stable() {}\n")

	service := NewRefreshService()
	if _, err := service.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() initial error = %v", err)
	}

	if err := os.Rename(filepath.Join(repoRoot, "move.go"), filepath.Join(repoRoot, "moved.go")); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}
	writeRepoFile(t, filepath.Join(repoRoot, "moved.go"), "package main\n\nfunc Moved() {}\n")

	if _, err := service.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonManual,
	}); err != nil {
		t.Fatalf("Refresh() move error = %v", err)
	}

	artifacts := loadArtifactsForRepo(t, repoRoot)
	if _, ok := artifacts.extractions["move.go"]; ok {
		t.Fatalf("move.go extraction should be removed: %+v", artifacts.extractions["move.go"])
	}
	if !reflect.DeepEqual(symbolNamesForPath(artifacts.symbols, "moved.go"), []string{"main", "Moved"}) {
		t.Fatalf("moved.go symbols = %+v", artifacts.symbols["moved.go"])
	}
	if !reflect.DeepEqual(symbolNamesForPath(artifacts.symbols, "stable.go"), []string{"main", "Stable"}) {
		t.Fatalf("stable.go symbols = %+v", artifacts.symbols["stable.go"])
	}
}

func TestIgnoreTransitionRemovesArtifacts(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc Main() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "keep.go"), "package main\n\nfunc Keep() {}\n")

	service := NewRefreshService()
	if _, err := service.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() initial error = %v", err)
	}

	writeRepoFile(t, filepath.Join(repoRoot, ".gitignore"), "main.go\n")
	if _, err := service.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonManual,
	}); err != nil {
		t.Fatalf("Refresh() ignore error = %v", err)
	}

	ignored := loadArtifactsForRepo(t, repoRoot)
	if _, ok := ignored.extractions["main.go"]; ok {
		t.Fatalf("main.go extraction should be removed when ignored: %+v", ignored.extractions["main.go"])
	}

	writeRepoFile(t, filepath.Join(repoRoot, ".gitignore"), "")
	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc MainAgain() {}\n")
	if _, err := service.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonManual,
	}); err != nil {
		t.Fatalf("Refresh() reinclude error = %v", err)
	}

	reincluded := loadArtifactsForRepo(t, repoRoot)
	if !reflect.DeepEqual(symbolNamesForPath(reincluded.symbols, "main.go"), []string{"main", "MainAgain"}) {
		t.Fatalf("main.go symbols after reinclude = %+v", reincluded.symbols["main.go"])
	}
	if !reflect.DeepEqual(symbolNamesForPath(reincluded.symbols, "keep.go"), []string{"main", "Keep"}) {
		t.Fatalf("keep.go symbols = %+v", reincluded.symbols["keep.go"])
	}
}

func TestSyntaxBreakExtractionRecovery(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc Healthy() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "helper.go"), "package main\n\nfunc Helper() {}\n")

	service := NewRefreshService()
	if _, err := service.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() initial error = %v", err)
	}

	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\nfunc (\n")
	broken, err := service.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonManual,
	})
	if err != nil {
		t.Fatalf("Refresh() broken error = %v", err)
	}

	degraded := loadArtifactsForRepo(t, repoRoot)
	if degraded.extractions["main.go"].CoverageState != repository.ExtractionCoverageStateFailed {
		t.Fatalf("main.go broken coverage = %q", degraded.extractions["main.go"].CoverageState)
	}
	if broken.FailedExtractions != 1 {
		t.Fatalf("FailedExtractions = %d, want 1", broken.FailedExtractions)
	}
	if !reflect.DeepEqual(symbolNamesForPath(degraded.symbols, "helper.go"), []string{"main", "Helper"}) {
		t.Fatalf("helper.go symbols after broken refresh = %+v", degraded.symbols["helper.go"])
	}

	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc Recovered() {}\n")
	if _, err := service.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonManual,
	}); err != nil {
		t.Fatalf("Refresh() recovery error = %v", err)
	}

	recovered := loadArtifactsForRepo(t, repoRoot)
	if recovered.extractions["main.go"].CoverageState != repository.ExtractionCoverageStateSupported {
		t.Fatalf("main.go recovered coverage = %q", recovered.extractions["main.go"].CoverageState)
	}
	if !reflect.DeepEqual(symbolNamesForPath(recovered.symbols, "main.go"), []string{"main", "Recovered"}) {
		t.Fatalf("main.go recovered symbols = %+v", recovered.symbols["main.go"])
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

type repoArtifacts struct {
	extractions map[string]repository.FileExtractionRecord
	symbols     map[string][]repository.SymbolRecord
}

func loadArtifactsForRepo(t *testing.T, repoRoot string) repoArtifacts {
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

	extractions, err := store.ListFileExtractions(context.Background(), record.ID)
	if err != nil {
		t.Fatalf("ListFileExtractions() error = %v", err)
	}
	symbols, err := store.ListSymbols(context.Background(), record.ID)
	if err != nil {
		t.Fatalf("ListSymbols() error = %v", err)
	}

	result := repoArtifacts{
		extractions: make(map[string]repository.FileExtractionRecord, len(extractions)),
		symbols:     make(map[string][]repository.SymbolRecord),
	}
	for _, extraction := range extractions {
		result.extractions[extraction.Path] = extraction
	}
	for _, symbol := range symbols {
		result.symbols[symbol.Path] = append(result.symbols[symbol.Path], symbol)
	}
	return result
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

func persistedSnapshotFileByPath(files []repository.PersistedFileSnapshotRecord, path string) repository.PersistedFileSnapshotRecord {
	for _, file := range files {
		if file.Path == path {
			return file
		}
	}
	return repository.PersistedFileSnapshotRecord{}
}

func symbolNamesForPath(symbols map[string][]repository.SymbolRecord, path string) []string {
	records := symbols[path]
	names := make([]string, 0, len(records))
	for _, symbol := range records {
		names = append(names, symbol.Name)
	}
	return names
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

type refreshFixture struct {
	repoRoot string
	service  RefreshService
}

func newRefreshFixture(t *testing.T) refreshFixture {
	t.Helper()

	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, ".gitignore"), "")
	writeRepoFile(t, filepath.Join(repoRoot, "README.md"), "# Repo\n")
	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n")
	writeRepoFile(t, filepath.Join(repoRoot, "move-me.txt"), "move me\n")
	writeRepoFile(t, filepath.Join(repoRoot, "delete-me.txt"), "delete me\n")
	writeRepoFile(t, filepath.Join(repoRoot, "ignored-later.log"), "baseline log\n")

	return refreshFixture{
		repoRoot: repoRoot,
		service:  NewRefreshService(),
	}
}

func (f refreshFixture) applyTrackedMutations(t *testing.T) {
	t.Helper()

	writeRepoFile(t, filepath.Join(f.repoRoot, "main.go"), "package main\n\nfunc refreshed() {}\n")
	writeRepoFile(t, filepath.Join(f.repoRoot, "added.go"), "package main\n")
	writeRepoFile(t, filepath.Join(f.repoRoot, ".gitignore"), "*.log\n")
	if err := os.Remove(filepath.Join(f.repoRoot, "delete-me.txt")); err != nil {
		t.Fatalf("Remove(delete-me.txt) error = %v", err)
	}
	if err := os.Rename(filepath.Join(f.repoRoot, "move-me.txt"), filepath.Join(f.repoRoot, "moved.txt")); err != nil {
		t.Fatalf("Rename(move-me.txt) error = %v", err)
	}
}
