package repository

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestDiscoveryDeterministicOrderingAndBuiltins(t *testing.T) {
	repoRoot := initDiscoveryRepo(t)
	createBuiltInFixtureTree(t, repoRoot)
	writeFile(t, filepath.Join(repoRoot, ".gitignore"), "*.tmp\n!keep.tmp\nignored-dir/\n")
	writeFile(t, filepath.Join(repoRoot, ".git", "info", "exclude"), "local-only.txt\n")
	writeFile(t, filepath.Join(repoRoot, "src", "app.go"), "package main\n")
	writeFile(t, filepath.Join(repoRoot, "keep.tmp"), "keep me\n")
	writeFile(t, filepath.Join(repoRoot, "ignored.tmp"), "ignore me\n")
	writeFile(t, filepath.Join(repoRoot, "local-only.txt"), "ignored via exclude\n")
	writeFile(t, filepath.Join(repoRoot, "ignored-dir", "nested.txt"), "ignored dir\n")

	discoveredAt := time.Date(2026, 3, 14, 20, 0, 0, 0, time.UTC)
	result, err := Discovery{
		RootPath: repoRoot,
		Matcher:  NewIgnoreMatcher(repoRoot),
		Now: func() time.Time {
			return discoveredAt
		},
	}.Walk()
	if err != nil {
		t.Fatalf("walk repository: %v", err)
	}

	if got, want := directoryPathsByReason(result.Directories, IgnoreReasonBuiltinExclusion), []string{
		".git",
		".next",
		".optimusctx",
		".turbo",
		"build",
		"coverage",
		"dist",
		"node_modules",
		"tmp",
		"vendor",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("builtin ignored directories = %v, want %v", got, want)
	}

	if got, want := includedFilePaths(result.Files), []string{".gitignore", "keep.tmp", "src/app.go"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("included file paths = %v, want %v", got, want)
	}

	if ignored := fileByPath(result.Files, "ignored.tmp"); ignored.IgnoreReason != IgnoreReasonGitIgnore {
		t.Fatalf("ignored.tmp reason = %q, want %q", ignored.IgnoreReason, IgnoreReasonGitIgnore)
	}
	if excluded := fileByPath(result.Files, "local-only.txt"); excluded.IgnoreReason != IgnoreReasonGitInfoExclude {
		t.Fatalf("local-only.txt reason = %q, want %q", excluded.IgnoreReason, IgnoreReasonGitInfoExclude)
	}
	if ignoredDir := directoryByPath(result.Directories, "ignored-dir"); ignoredDir.IgnoreReason != IgnoreReasonGitIgnore {
		t.Fatalf("ignored-dir reason = %q, want %q", ignoredDir.IgnoreReason, IgnoreReasonGitIgnore)
	}

	for _, path := range []string{"keep.tmp", "src/app.go"} {
		record := fileByPath(result.Files, path)
		if record.LastIndexedAt != discoveredAt {
			t.Fatalf("%s last indexed at = %s, want %s", path, record.LastIndexedAt, discoveredAt)
		}
		if len(record.ContentHash) != 64 {
			t.Fatalf("%s hash length = %d, want 64", path, len(record.ContentHash))
		}
	}
}

func TestIgnoreMatcherGitInfoExcludeAndNegation(t *testing.T) {
	repoRoot := initDiscoveryRepo(t)
	writeFile(t, filepath.Join(repoRoot, ".gitignore"), "*.tmp\n!keep.tmp\nignored-dir/\n")
	writeFile(t, filepath.Join(repoRoot, ".git", "info", "exclude"), "local-only.txt\n")
	writeFile(t, filepath.Join(repoRoot, "keep.tmp"), "keep me\n")
	writeFile(t, filepath.Join(repoRoot, "ignored.tmp"), "ignore me\n")
	writeFile(t, filepath.Join(repoRoot, "local-only.txt"), "local only\n")

	matcher := NewIgnoreMatcher(repoRoot)

	if match := matcher.Evaluate("ignored.tmp", false); match.Reason != IgnoreReasonGitIgnore {
		t.Fatalf("ignored.tmp reason = %q, want %q", match.Reason, IgnoreReasonGitIgnore)
	}
	if match := matcher.Evaluate("local-only.txt", false); match.Reason != IgnoreReasonGitInfoExclude {
		t.Fatalf("local-only.txt reason = %q, want %q", match.Reason, IgnoreReasonGitInfoExclude)
	}
	if match := matcher.Evaluate("keep.tmp", false); match.Status != IgnoreStatusIncluded {
		t.Fatalf("keep.tmp status = %q, want %q", match.Status, IgnoreStatusIncluded)
	}
	if match := matcher.Evaluate("ignored-dir", true); match.Reason != IgnoreReasonGitIgnore {
		t.Fatalf("ignored-dir reason = %q, want %q", match.Reason, IgnoreReasonGitIgnore)
	}
}

func TestDiscoveryDoesNotTraverseSymlinks(t *testing.T) {
	repoRoot := initDiscoveryRepo(t)
	writeFile(t, filepath.Join(repoRoot, "src", "app.go"), "package main\n")
	writeFile(t, filepath.Join(repoRoot, "real", "nested.go"), "package real\n")

	target := filepath.Join(repoRoot, "real")
	link := filepath.Join(repoRoot, "linked-dir")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	result, err := NewDiscovery(repoRoot).Walk()
	if err != nil {
		t.Fatalf("walk repository: %v", err)
	}

	record := fileByPath(result.Files, "linked-dir")
	if record.IgnoreReason != IgnoreReasonSymlinkNotTraversed {
		t.Fatalf("linked-dir reason = %q, want %q", record.IgnoreReason, IgnoreReasonSymlinkNotTraversed)
	}
	if record.IgnoreStatus != IgnoreStatusIgnored {
		t.Fatalf("linked-dir status = %q, want %q", record.IgnoreStatus, IgnoreStatusIgnored)
	}
	if hasFile(result.Files, "linked-dir/nested.go") {
		t.Fatal("symlink target should not be traversed")
	}
}

func TestDiscoveryMetadata(t *testing.T) {
	repoRoot := initDiscoveryRepo(t)
	writeFile(t, filepath.Join(repoRoot, ".gitignore"), "*.tmp\n")
	writeFile(t, filepath.Join(repoRoot, "src", "app.go"), "package main\n")
	writeFile(t, filepath.Join(repoRoot, "ignored.tmp"), "ignore me\n")

	discoveredAt := time.Date(2026, 3, 14, 21, 0, 0, 0, time.UTC)
	result, err := Discovery{
		RootPath: repoRoot,
		Matcher:  NewIgnoreMatcher(repoRoot),
		Now: func() time.Time {
			return discoveredAt
		},
	}.Walk()
	if err != nil {
		t.Fatalf("walk repository: %v", err)
	}

	if result.Repository.RootPath != repoRoot {
		t.Fatalf("repository root = %q, want %q", result.Repository.RootPath, repoRoot)
	}
	if result.Repository.DetectionMode != DetectionModeGit {
		t.Fatalf("repository detection mode = %q, want %q", result.Repository.DetectionMode, DetectionModeGit)
	}

	srcDirectory := directoryByPath(result.Directories, "src")
	if srcDirectory.ParentPath != "." {
		t.Fatalf("src parent path = %q, want .", srcDirectory.ParentPath)
	}
	if srcDirectory.IgnoreStatus != IgnoreStatusIncluded {
		t.Fatalf("src ignore status = %q, want %q", srcDirectory.IgnoreStatus, IgnoreStatusIncluded)
	}

	included := fileByPath(result.Files, "src/app.go")
	if included.DirectoryPath != "src" {
		t.Fatalf("included directory path = %q, want src", included.DirectoryPath)
	}
	if included.Extension != ".go" {
		t.Fatalf("included extension = %q, want .go", included.Extension)
	}
	if included.LanguageHint != "go" {
		t.Fatalf("included language hint = %q, want go", included.LanguageHint)
	}
	if included.SizeBytes == 0 {
		t.Fatal("included file size should be populated")
	}
	if included.ContentHash == "" {
		t.Fatal("included file hash should be populated")
	}
	if included.LastIndexedAt != discoveredAt {
		t.Fatalf("included last indexed at = %s, want %s", included.LastIndexedAt, discoveredAt)
	}
	if included.FilesystemModTime.IsZero() {
		t.Fatal("included filesystem mod time should be populated")
	}
	if included.IgnoreStatus != IgnoreStatusIncluded || included.IgnoreReason != IgnoreReasonNone {
		t.Fatalf("included ignore state = (%q, %q), want (%q, %q)", included.IgnoreStatus, included.IgnoreReason, IgnoreStatusIncluded, IgnoreReasonNone)
	}

	ignored := fileByPath(result.Files, "ignored.tmp")
	if ignored.DirectoryPath != "." {
		t.Fatalf("ignored directory path = %q, want .", ignored.DirectoryPath)
	}
	if ignored.ContentHash != "" {
		t.Fatalf("ignored hash = %q, want empty", ignored.ContentHash)
	}
	if !ignored.LastIndexedAt.IsZero() {
		t.Fatalf("ignored last indexed at = %s, want zero", ignored.LastIndexedAt)
	}
	if ignored.IgnoreReason != IgnoreReasonGitIgnore {
		t.Fatalf("ignored reason = %q, want %q", ignored.IgnoreReason, IgnoreReasonGitIgnore)
	}
}

func TestFileRecord(t *testing.T) {
	repoRoot := initDiscoveryRepo(t)
	writeFile(t, filepath.Join(repoRoot, "notes.md"), "# Notes\n")

	result, err := NewDiscovery(repoRoot).Walk()
	if err != nil {
		t.Fatalf("walk repository: %v", err)
	}

	record := fileByPath(result.Files, "notes.md")
	if record.Path != "notes.md" {
		t.Fatalf("record path = %q, want notes.md", record.Path)
	}
	if record.DirectoryPath != "." {
		t.Fatalf("record directory path = %q, want .", record.DirectoryPath)
	}
	if record.Extension != ".md" {
		t.Fatalf("record extension = %q, want .md", record.Extension)
	}
	if record.LanguageHint != "markdown" {
		t.Fatalf("record language hint = %q, want markdown", record.LanguageHint)
	}
	if len(record.ContentHash) != 64 {
		t.Fatalf("record hash length = %d, want 64", len(record.ContentHash))
	}
	if record.LastIndexedAt.IsZero() {
		t.Fatal("record last indexed at should be populated")
	}
}

func TestConditionalHashingReusesPersistedHashesOnNoOpScan(t *testing.T) {
	repoRoot := initDiscoveryRepo(t)
	writeFile(t, filepath.Join(repoRoot, "src", "app.go"), "package main\n")
	writeFile(t, filepath.Join(repoRoot, "README.md"), "# OptimusCtx\n")

	initialResult, err := NewDiscovery(repoRoot).Walk()
	if err != nil {
		t.Fatalf("initial Walk() error = %v", err)
	}

	hashCalls := 0
	secondResult, err := Discovery{
		RootPath: repoRoot,
		Matcher:  NewIgnoreMatcher(repoRoot),
		PersistedSnapshot: RepositorySnapshot{
			Files: persistedFilesFromRecords(initialResult.Files),
		},
		HashFile: func(path string) (string, error) {
			hashCalls++
			return hashFile(path)
		},
	}.Walk()
	if err != nil {
		t.Fatalf("second Walk() error = %v", err)
	}

	if hashCalls != 0 {
		t.Fatalf("hash calls = %d, want 0", hashCalls)
	}

	for _, path := range []string{"README.md", "src/app.go"} {
		first := fileByPath(initialResult.Files, path)
		second := fileByPath(secondResult.Files, path)
		if first.ContentHash != second.ContentHash {
			t.Fatalf("%s hash = %q, want %q", path, second.ContentHash, first.ContentHash)
		}
		if !first.LastIndexedAt.Equal(second.LastIndexedAt) {
			t.Fatalf("%s last indexed at = %v, want %v", path, second.LastIndexedAt, first.LastIndexedAt)
		}
	}
}

func TestConditionalHashingRehashesChangedOrReincludedFiles(t *testing.T) {
	repoRoot := initDiscoveryRepo(t)
	writeFile(t, filepath.Join(repoRoot, ".gitignore"), "*.tmp\n")
	writeFile(t, filepath.Join(repoRoot, "stable.go"), "package main\n")
	writeFile(t, filepath.Join(repoRoot, "changed.go"), "package changed\n")
	writeFile(t, filepath.Join(repoRoot, "new.tmp"), "temporary\n")

	initialResult, err := NewDiscovery(repoRoot).Walk()
	if err != nil {
		t.Fatalf("initial Walk() error = %v", err)
	}

	writeFile(t, filepath.Join(repoRoot, "changed.go"), "package changed\n\nfunc refreshed() {}\n")
	if err := os.WriteFile(filepath.Join(repoRoot, ".gitignore"), []byte(""), 0o644); err != nil {
		t.Fatalf("WriteFile(.gitignore) error = %v", err)
	}

	var hashedPaths []string
	secondResult, err := Discovery{
		RootPath: repoRoot,
		Matcher:  NewIgnoreMatcher(repoRoot),
		PersistedSnapshot: RepositorySnapshot{
			Files: persistedFilesFromRecords(initialResult.Files),
		},
		HashFile: func(path string) (string, error) {
			hashedPaths = append(hashedPaths, filepath.Base(path))
			return hashFile(path)
		},
	}.Walk()
	if err != nil {
		t.Fatalf("second Walk() error = %v", err)
	}

	if got, want := hashedPaths, []string{".gitignore", "changed.go", "new.tmp"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("hashed paths = %v, want %v", got, want)
	}

	stableBefore := fileByPath(initialResult.Files, "stable.go")
	stableAfter := fileByPath(secondResult.Files, "stable.go")
	if stableBefore.ContentHash != stableAfter.ContentHash {
		t.Fatalf("stable.go hash changed from %q to %q", stableBefore.ContentHash, stableAfter.ContentHash)
	}
	if !stableBefore.LastIndexedAt.Equal(stableAfter.LastIndexedAt) {
		t.Fatalf("stable.go last indexed at = %v, want %v", stableAfter.LastIndexedAt, stableBefore.LastIndexedAt)
	}

	reincluded := fileByPath(secondResult.Files, "new.tmp")
	if reincluded.IgnoreStatus != IgnoreStatusIncluded {
		t.Fatalf("new.tmp ignore status = %q, want %q", reincluded.IgnoreStatus, IgnoreStatusIncluded)
	}
	if reincluded.ContentHash == "" {
		t.Fatal("new.tmp should be hashed after re-inclusion")
	}
}

func TestDiscoveryExcludesRuntimeStateContents(t *testing.T) {
	repoRoot := initDiscoveryRepo(t)
	writeFile(t, filepath.Join(repoRoot, "main.go"), "package main\n")
	writeFile(t, filepath.Join(repoRoot, ".optimusctx", "state.json"), "{\"generation\":1}\n")
	writeFile(t, filepath.Join(repoRoot, ".optimusctx", "tmp", "run.log"), "ignored runtime state\n")

	result, err := NewDiscovery(repoRoot).Walk()
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	if directory := directoryByPath(result.Directories, ".optimusctx"); directory.IgnoreReason != IgnoreReasonBuiltinExclusion {
		t.Fatalf(".optimusctx ignore reason = %q, want %q", directory.IgnoreReason, IgnoreReasonBuiltinExclusion)
	}
	if hasFile(result.Files, ".optimusctx/state.json") {
		t.Fatal(".optimusctx/state.json should not be discovered as repository content")
	}
	if hasFile(result.Files, ".optimusctx/tmp/run.log") {
		t.Fatal(".optimusctx/tmp/run.log should not be discovered as repository content")
	}
	if got, want := includedFilePaths(result.Files), []string{"main.go"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("included file paths = %v, want %v", got, want)
	}
}

func TestStreamingHashing(t *testing.T) {
	repoRoot := initDiscoveryRepo(t)
	path := filepath.Join(repoRoot, "large.txt")
	content := strings.Repeat("optimusctx-streaming-hash\n", 4096)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}

	got, err := hashFile(path)
	if err != nil {
		t.Fatalf("hashFile() error = %v", err)
	}

	wantSum := sha256.Sum256([]byte(content))
	want := hex.EncodeToString(wantSum[:])
	if got != want {
		t.Fatalf("hashFile() = %q, want %q", got, want)
	}
}

func initDiscoveryRepo(t *testing.T) string {
	t.Helper()

	repoRoot := t.TempDir()
	runGitCommand(t, repoRoot, "init")
	return repoRoot
}

func createBuiltInFixtureTree(t *testing.T, repoRoot string) {
	t.Helper()

	for _, dir := range []string{
		".next",
		".optimusctx",
		".turbo",
		"build",
		"coverage",
		"dist",
		"node_modules",
		"tmp",
		"vendor",
	} {
		writeFile(t, filepath.Join(repoRoot, dir, "artifact.txt"), dir+"\n")
	}
}

func persistedFilesFromRecords(files []FileRecord) []PersistedFileSnapshotRecord {
	records := make([]PersistedFileSnapshotRecord, 0, len(files))
	for _, file := range files {
		records = append(records, PersistedFileSnapshotRecord{
			Path:              file.Path,
			DirectoryPath:     file.DirectoryPath,
			Extension:         file.Extension,
			LanguageHint:      file.LanguageHint,
			SizeBytes:         file.SizeBytes,
			ContentHash:       file.ContentHash,
			LastIndexedAt:     file.LastIndexedAt,
			FilesystemModTime: file.FilesystemModTime,
			IgnoreStatus:      file.IgnoreStatus,
			IgnoreReason:      file.IgnoreReason,
			DiscoveredAt:      file.DiscoveredAt,
			UpdatedAt:         file.DiscoveredAt,
		})
	}
	return records
}

func includedFilePaths(files []FileRecord) []string {
	paths := make([]string, 0, len(files))
	for _, file := range files {
		if file.IgnoreStatus == IgnoreStatusIncluded {
			paths = append(paths, file.Path)
		}
	}
	return paths
}

func directoryPathsByReason(directories []DirectoryRecord, reason IgnoreReason) []string {
	paths := make([]string, 0, len(directories))
	for _, directory := range directories {
		if directory.IgnoreReason == reason {
			paths = append(paths, directory.Path)
		}
	}
	return paths
}

func fileByPath(files []FileRecord, target string) FileRecord {
	for _, file := range files {
		if file.Path == target {
			return file
		}
	}
	return FileRecord{}
}

func directoryByPath(directories []DirectoryRecord, target string) DirectoryRecord {
	for _, directory := range directories {
		if directory.Path == target {
			return directory
		}
	}
	return DirectoryRecord{}
}

func hasFile(files []FileRecord, target string) bool {
	for _, file := range files {
		if file.Path == target {
			return true
		}
	}
	return false
}
