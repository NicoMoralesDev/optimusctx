package refresh

import (
	"reflect"
	"testing"

	"github.com/niccrow/optimusctx/internal/repository"
)

func TestSubtreeFingerprint(t *testing.T) {
	current := Snapshot{
		Directories: []DirectorySnapshot{
			{Path: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg", ParentPath: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg/ignored", ParentPath: "pkg", IgnoreStatus: repository.IgnoreStatusIgnored, IgnoreReason: repository.IgnoreReasonGitIgnore},
		},
		Files: []FileSnapshot{
			{Path: "pkg/a.go", DirectoryPath: "pkg", ContentHash: "hash-a", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg/ignored/b.tmp", DirectoryPath: "pkg/ignored", IgnoreStatus: repository.IgnoreStatusIgnored},
		},
	}

	fingerprints := ComputeSubtreeFingerprints(current, Snapshot{}, []string{"pkg/ignored", "pkg", "."})

	if fingerprints["pkg/ignored"] == "" || fingerprints["pkg"] == "" || fingerprints["."] == "" {
		t.Fatalf("expected fingerprints for pkg/ignored, pkg, and root, got %#v", fingerprints)
	}
	if fingerprints["pkg"] == fingerprints["."] {
		t.Fatal("pkg and root fingerprints should differ")
	}
}

func TestFingerprintPropagation(t *testing.T) {
	persisted := Snapshot{
		Directories: []DirectorySnapshot{
			{Path: ".", IgnoreStatus: repository.IgnoreStatusIncluded, SubtreeFingerprint: "root-old"},
			{Path: "pkg", ParentPath: ".", IgnoreStatus: repository.IgnoreStatusIncluded, SubtreeFingerprint: "pkg-old"},
			{Path: "pkg/sub", ParentPath: "pkg", IgnoreStatus: repository.IgnoreStatusIncluded, SubtreeFingerprint: "sub-old"},
			{Path: "docs", ParentPath: ".", IgnoreStatus: repository.IgnoreStatusIncluded, SubtreeFingerprint: "docs-stable"},
		},
		Files: []FileSnapshot{
			{Path: "pkg/sub/file.go", DirectoryPath: "pkg/sub", ContentHash: "old-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "docs/readme.md", DirectoryPath: "docs", ContentHash: "docs-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
		},
	}
	current := Snapshot{
		Directories: []DirectorySnapshot{
			{Path: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg", ParentPath: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg/sub", ParentPath: "pkg", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "docs", ParentPath: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
		},
		Files: []FileSnapshot{
			{Path: "pkg/sub/file.go", DirectoryPath: "pkg/sub", ContentHash: "new-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "docs/readme.md", DirectoryPath: "docs", ContentHash: "docs-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
		},
	}

	fingerprints := ComputeSubtreeFingerprints(current, persisted, []string{"pkg/sub", "pkg", "."})

	if got := fingerprints["pkg/sub"]; got == "" || got == "sub-old" {
		t.Fatalf("pkg/sub fingerprint = %q, want recomputed non-empty value", got)
	}
	if got := fingerprints["pkg"]; got == "" || got == "pkg-old" {
		t.Fatalf("pkg fingerprint = %q, want recomputed non-empty value", got)
	}
	if got := fingerprints["."]; got == "" || got == "root-old" {
		t.Fatalf("root fingerprint = %q, want recomputed non-empty value", got)
	}
	if _, ok := fingerprints["docs"]; ok {
		t.Fatalf("docs fingerprint should not be recomputed, got %q", fingerprints["docs"])
	}
}

func TestAffectedDirectories(t *testing.T) {
	persisted := Snapshot{
		Directories: []DirectorySnapshot{
			{Path: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "docs", ParentPath: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg", ParentPath: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg/internal", ParentPath: "pkg", IgnoreStatus: repository.IgnoreStatusIncluded},
		},
		Files: []FileSnapshot{
			{Path: "pkg/internal/renamed.go", DirectoryPath: "pkg/internal", ContentHash: "move-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "docs/guide.md", DirectoryPath: "docs", ContentHash: "docs-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
		},
	}
	current := Snapshot{
		Directories: []DirectorySnapshot{
			{Path: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "docs", ParentPath: ".", IgnoreStatus: repository.IgnoreStatusIgnored, IgnoreReason: repository.IgnoreReasonGitIgnore},
			{Path: "pkg", ParentPath: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg/internal", ParentPath: "pkg", IgnoreStatus: repository.IgnoreStatusIncluded},
		},
		Files: []FileSnapshot{
			{Path: "pkg/internal/moved.go", DirectoryPath: "pkg/internal", ContentHash: "move-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "docs/guide.md", DirectoryPath: "docs", IgnoreStatus: repository.IgnoreStatusIgnored},
		},
	}

	diff := DiffSnapshots(current, persisted)
	affected := AffectedDirectories(current, persisted, diff)

	if got, want := affected, []string{".", "docs", "pkg", "pkg/internal"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("affected directories = %v, want %v", got, want)
	}
}
