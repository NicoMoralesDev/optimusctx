package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
)

func TestStatusCommandBasic(t *testing.T) {
	repoRoot := initCLIRepo(t)

	previousHealth := statusHealthService
	previousWatch := statusWatchService
	t.Cleanup(func() {
		statusHealthService = previousHealth
		statusWatchService = previousWatch
	})

	statusHealthService = func(ctx context.Context, workingDir string) (repository.HealthResult, error) {
		return repository.HealthResult{
			Repository: repository.LayeredContextEnvelope{RepositoryRoot: repoRoot, Generation: 7, Freshness: repository.FreshnessStatusFresh},
			Summary:    repository.HealthSummary{Initialized: true, StateStatus: repository.HealthStateStatusReady, RepositoryRegistered: true},
			Refresh: repository.HealthRefreshDiagnostics{
				LastRefreshCompletedAt: time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC),
			},
		}, nil
	}
	statusWatchService = func(ctx context.Context, workingDir string) (repository.WatchStatusResult, error) {
		return repository.WatchStatusResult{RepositoryRoot: repoRoot, Status: repository.WatchStatusKindRunning, Reason: "watch process heartbeat is current"}, nil
	}

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"status"}, &stdout); err != nil {
			t.Fatalf("Execute(status) error = %v", err)
		}
		output := stdout.String()
		for _, want := range []string{
			"repository root: " + repoRoot,
			"runtime status: running",
			"state status: ready",
			"freshness: fresh",
			"generation: 7",
			"watch status: running",
			"mcp command: optimusctx run",
			"supported clients: claude-desktop, claude-cli, codex-app, codex-cli",
			"next action: runtime is ready; rerun `optimusctx init --client <client> [--write]` for claude-desktop, claude-cli, codex-app, or codex-cli to preview or register MCP. Registered MCP clients should launch `optimusctx run` automatically, and manual `optimusctx run` remains the direct/debug path.",
		} {
			if !strings.Contains(output, want) {
				t.Fatalf("status output missing %q:\n%s", want, output)
			}
		}
	})
}

func TestStatusCommandHelp(t *testing.T) {
	var stdout bytes.Buffer
	if err := NewRootCommand().Execute([]string{"status", "--help"}, &stdout); err != nil {
		t.Fatalf("Execute(status --help) error = %v", err)
	}

	output := stdout.String()
	want := "optimusctx status"
	if !strings.Contains(output, want) {
		t.Fatalf("help output missing %q:\n%s", want, output)
	}
	if strings.Contains(output, "--client") {
		t.Fatalf("help output should not advertise registration flags:\n%s", output)
	}
}

func TestStatusCommandRejectsClientRegistrationFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "client", args: []string{"status", "--client", "claude-desktop"}, want: "status does not accept flag \"--client\""},
		{name: "config", args: []string{"status", "--config", "/tmp/config.json"}, want: "status does not accept flag \"--config\""},
		{name: "binary", args: []string{"status", "--binary", "/tmp/optimusctx"}, want: "status does not accept flag \"--binary\""},
		{name: "scope", args: []string{"status", "--scope", "project"}, want: "status does not accept flag \"--scope\""},
		{name: "write", args: []string{"status", "--write"}, want: "status does not accept flag \"--write\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			err := NewRootCommand().Execute(tt.args, &stdout)
			if err == nil || err.Error() != tt.want {
				t.Fatalf("Execute(%v) error = %v, want %q", tt.args, err, tt.want)
			}
		})
	}
}

func TestStatusCommandErrors(t *testing.T) {
	repoRoot := initCLIRepo(t)
	_ = repoRoot

	t.Run("rejects unsupported arg", func(t *testing.T) {
		var stdout bytes.Buffer
		err := NewRootCommand().Execute([]string{"status", "now"}, &stdout)
		if err == nil || err.Error() != "status does not accept argument \"now\"" {
			t.Fatalf("Execute(status now) error = %v", err)
		}
	})

	t.Run("returns health error", func(t *testing.T) {
		previousHealth := statusHealthService
		previousWatch := statusWatchService
		t.Cleanup(func() {
			statusHealthService = previousHealth
			statusWatchService = previousWatch
		})
		statusHealthService = func(ctx context.Context, workingDir string) (repository.HealthResult, error) {
			return repository.HealthResult{}, errors.New("boom")
		}
		statusWatchService = func(ctx context.Context, workingDir string) (repository.WatchStatusResult, error) {
			return repository.WatchStatusResult{}, nil
		}
		var stdout bytes.Buffer
		err := NewRootCommand().Execute([]string{"status"}, &stdout)
		if err == nil || err.Error() != "boom" {
			t.Fatalf("Execute(status) error = %v", err)
		}
	})
}
