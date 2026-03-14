package refresh

import (
	"reflect"
	"testing"

	"github.com/niccrow/optimusctx/internal/repository"
)

func TestRefreshDiff(t *testing.T) {
	persisted := Snapshot{
		Directories: []DirectorySnapshot{
			{Path: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg", ParentPath: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
		},
		Files: []FileSnapshot{
			{Path: "pkg/changed.go", DirectoryPath: "pkg", ContentHash: "old-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg/deleted.go", DirectoryPath: "pkg", ContentHash: "deleted-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg/stable.go", DirectoryPath: "pkg", ContentHash: "stable-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
		},
	}
	current := Snapshot{
		Directories: []DirectorySnapshot{
			{Path: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg", ParentPath: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
		},
		Files: []FileSnapshot{
			{Path: "pkg/added.go", DirectoryPath: "pkg", ContentHash: "added-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg/changed.go", DirectoryPath: "pkg", ContentHash: "new-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg/stable.go", DirectoryPath: "pkg", ContentHash: "stable-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
		},
	}

	diff := DiffSnapshots(current, persisted)

	if got, want := paths(diff.Unchanged), []string{"pkg/stable.go"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unchanged = %v, want %v", got, want)
	}
	if got, want := paths(diff.Added), []string{"pkg/added.go"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("added = %v, want %v", got, want)
	}
	if got, want := paths(diff.Changed), []string{"pkg/changed.go"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("changed = %v, want %v", got, want)
	}
	if got, want := paths(diff.Deleted), []string{"pkg/deleted.go"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("deleted = %v, want %v", got, want)
	}
}

func TestMoveDetection(t *testing.T) {
	persisted := Snapshot{
		Directories: []DirectorySnapshot{
			{Path: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg", ParentPath: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
		},
		Files: []FileSnapshot{
			{Path: "pkg/renamed.go", DirectoryPath: "pkg", ContentHash: "move-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg/duplicate-a.go", DirectoryPath: "pkg", ContentHash: "dup-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg/duplicate-b.go", DirectoryPath: "pkg", ContentHash: "dup-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
		},
	}
	current := Snapshot{
		Directories: []DirectorySnapshot{
			{Path: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg", ParentPath: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
		},
		Files: []FileSnapshot{
			{Path: "pkg/moved.go", DirectoryPath: "pkg", ContentHash: "move-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg/duplicate-c.go", DirectoryPath: "pkg", ContentHash: "dup-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg/duplicate-d.go", DirectoryPath: "pkg", ContentHash: "dup-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
		},
	}

	diff := DiffSnapshots(current, persisted)

	if len(diff.Moved) != 1 {
		t.Fatalf("moved count = %d, want 1", len(diff.Moved))
	}
	if diff.Moved[0].PreviousPath != "pkg/renamed.go" || diff.Moved[0].Path != "pkg/moved.go" {
		t.Fatalf("move = %+v", diff.Moved[0])
	}
	if got, want := paths(diff.Added), []string{"pkg/duplicate-c.go", "pkg/duplicate-d.go"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("added = %v, want %v", got, want)
	}
	if got, want := paths(diff.Deleted), []string{"pkg/duplicate-a.go", "pkg/duplicate-b.go"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("deleted = %v, want %v", got, want)
	}
}

func TestIgnoreTransitions(t *testing.T) {
	persisted := Snapshot{
		Directories: []DirectorySnapshot{
			{Path: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg", ParentPath: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
		},
		Files: []FileSnapshot{
			{Path: "pkg/ignored-now.go", DirectoryPath: "pkg", ContentHash: "same-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg/reincluded.go", DirectoryPath: "pkg", IgnoreStatus: repository.IgnoreStatusIgnored},
		},
	}
	current := Snapshot{
		Directories: []DirectorySnapshot{
			{Path: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: "pkg", ParentPath: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
		},
		Files: []FileSnapshot{
			{Path: "pkg/ignored-now.go", DirectoryPath: "pkg", IgnoreStatus: repository.IgnoreStatusIgnored},
			{Path: "pkg/reincluded.go", DirectoryPath: "pkg", ContentHash: "restored-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
		},
	}

	diff := DiffSnapshots(current, persisted)

	if got, want := paths(diff.NewlyIgnored), []string{"pkg/ignored-now.go"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("newly ignored = %v, want %v", got, want)
	}
	if got, want := paths(diff.Reincluded), []string{"pkg/reincluded.go"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("reincluded = %v, want %v", got, want)
	}
}

func TestRuntimeStateExcludedFromRefreshCounts(t *testing.T) {
	persisted := Snapshot{
		Directories: []DirectorySnapshot{
			{Path: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: ".optimusctx", ParentPath: ".", IgnoreStatus: repository.IgnoreStatusIgnored, IgnoreReason: repository.IgnoreReasonBuiltinExclusion},
		},
		Files: []FileSnapshot{
			{Path: "main.go", DirectoryPath: ".", ContentHash: "stable-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
		},
	}
	current := Snapshot{
		Directories: []DirectorySnapshot{
			{Path: ".", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: ".optimusctx", ParentPath: ".", IgnoreStatus: repository.IgnoreStatusIgnored, IgnoreReason: repository.IgnoreReasonBuiltinExclusion},
		},
		Files: []FileSnapshot{
			{Path: "main.go", DirectoryPath: ".", ContentHash: "stable-hash", IgnoreStatus: repository.IgnoreStatusIncluded},
			{Path: ".optimusctx/state.json", DirectoryPath: ".optimusctx", IgnoreStatus: repository.IgnoreStatusIgnored, IgnoreReason: repository.IgnoreReasonBuiltinExclusion},
			{Path: ".optimusctx/db.sqlite", DirectoryPath: ".optimusctx", IgnoreStatus: repository.IgnoreStatusIgnored, IgnoreReason: repository.IgnoreReasonBuiltinExclusion},
		},
	}

	diff := DiffSnapshots(current, persisted)

	if got, want := paths(diff.Unchanged), []string{"main.go"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unchanged = %v, want %v", got, want)
	}
	if len(diff.Added) != 0 || len(diff.Changed) != 0 || len(diff.Deleted) != 0 || len(diff.Moved) != 0 || len(diff.NewlyIgnored) != 0 {
		t.Fatalf("runtime state should not affect refresh counts: %+v", diff)
	}
}

func paths(changes []FileChange) []string {
	result := make([]string, 0, len(changes))
	for _, change := range changes {
		result = append(result, change.Path)
	}
	return result
}
