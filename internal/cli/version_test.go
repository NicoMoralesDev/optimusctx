package cli

import (
	"bytes"
	"testing"

	"github.com/niccrow/optimusctx/internal/buildinfo"
)

func TestVersionCommand(t *testing.T) {
	previousVersion := buildinfo.Version
	previousCommit := buildinfo.Commit
	previousBuildDate := buildinfo.BuildDate
	t.Cleanup(func() {
		buildinfo.Version = previousVersion
		buildinfo.Commit = previousCommit
		buildinfo.BuildDate = previousBuildDate
	})

	buildinfo.Version = "v1.2.3"
	buildinfo.Commit = "abc1234"
	buildinfo.BuildDate = "2026-03-16T15:00:00Z"

	var stdout bytes.Buffer
	if err := newVersionCommand().Run(&stdout, nil); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	want := "optimusctx version=v1.2.3 commit=abc1234 build_date=2026-03-16T15:00:00Z\n"
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}
