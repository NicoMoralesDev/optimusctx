package buildinfo

import "testing"

func TestBuildInfoSummary(t *testing.T) {
	previousVersion := Version
	previousCommit := Commit
	previousBuildDate := BuildDate
	t.Cleanup(func() {
		Version = previousVersion
		Commit = previousCommit
		BuildDate = previousBuildDate
	})

	Version = "v1.2.3"
	Commit = "abc1234"
	BuildDate = "2026-03-16T15:00:00Z"

	info := Current()
	if info.Version != "v1.2.3" {
		t.Fatalf("info.Version = %q, want v1.2.3", info.Version)
	}
	if info.Commit != "abc1234" {
		t.Fatalf("info.Commit = %q, want abc1234", info.Commit)
	}
	if info.BuildDate != "2026-03-16T15:00:00Z" {
		t.Fatalf("info.BuildDate = %q, want 2026-03-16T15:00:00Z", info.BuildDate)
	}

	want := "optimusctx version=v1.2.3 commit=abc1234 build_date=2026-03-16T15:00:00Z"
	if got := info.Summary(); got != want {
		t.Fatalf("info.Summary() = %q, want %q", got, want)
	}
	if got := Summary(); got != want {
		t.Fatalf("Summary() = %q, want %q", got, want)
	}
}
