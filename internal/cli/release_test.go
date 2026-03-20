package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootHelpListsRelease(t *testing.T) {
	var stdout bytes.Buffer
	if err := NewRootCommand().Execute([]string{"--help"}, &stdout); err != nil {
		t.Fatalf("Execute(--help) error = %v", err)
	}
	if !strings.Contains(stdout.String(), "release   Prepare and validate a release plan") {
		t.Fatalf("help missing release command:\n%s", stdout.String())
	}
}

func TestRootExecuteReleaseHelp(t *testing.T) {
	var stdout bytes.Buffer
	if err := NewRootCommand().Execute([]string{"release", "help"}, &stdout); err != nil {
		t.Fatalf("Execute(release help) error = %v", err)
	}
	if !strings.Contains(stdout.String(), "optimusctx release <command>") {
		t.Fatalf("release help missing usage:\n%s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "prepare   Prepare and validate a release plan") {
		t.Fatalf("release help missing prepare subcommand:\n%s", stdout.String())
	}
}
