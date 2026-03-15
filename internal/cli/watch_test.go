package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/niccrow/optimusctx/internal/repository"
)

func TestWatchCommand(t *testing.T) {
	repoRoot := initCLIRepo(t)

	t.Run("root help lists watch", func(t *testing.T) {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"--help"}, &stdout); err != nil {
			t.Fatalf("Execute(--help) error = %v", err)
		}
		assertContains(t, stdout.String(), "watch     Run or inspect the optional repository watch process")
	})

	t.Run("run delegates to watch service", func(t *testing.T) {
		previous := watchRunCommandService
		t.Cleanup(func() {
			watchRunCommandService = previous
		})

		called := false
		watchRunCommandService = func(ctx context.Context, workingDir string) (repository.WatchRunResult, error) {
			called = true
			if workingDir != repoRoot {
				t.Fatalf("workingDir = %q, want %q", workingDir, repoRoot)
			}
			return repository.WatchRunResult{RepositoryRoot: repoRoot}, nil
		}

		withWorkingDirectory(t, repoRoot, func() {
			var stdout bytes.Buffer
			if err := NewRootCommand().Execute([]string{"watch", "run"}, &stdout); err != nil {
				t.Fatalf("Execute(watch run) error = %v", err)
			}
			if !called {
				t.Fatal("watch run service was not called")
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
		})
	})

	t.Run("status renders watch state", func(t *testing.T) {
		previous := watchStatusCommandService
		t.Cleanup(func() {
			watchStatusCommandService = previous
		})

		watchStatusCommandService = func(ctx context.Context, workingDir string) (repository.WatchStatusResult, error) {
			if workingDir != repoRoot {
				t.Fatalf("workingDir = %q, want %q", workingDir, repoRoot)
			}
			return repository.WatchStatusResult{
				RepositoryRoot: repoRoot,
				StatusPath:     repoRoot + "/.optimusctx/tmp/watch-status.json",
				Status:         repository.WatchStatusKindStale,
				Reason:         "watch heartbeat is stale",
				Record: repository.WatchStatusRecord{
					PID:                   42,
					StartedAt:             "2026-03-15T12:00:00Z",
					LastHeartbeatAt:       "2026-03-15T12:01:00Z",
					LastEventAt:           "2026-03-15T12:01:01Z",
					LastRefreshDoneAt:     "2026-03-15T12:01:02Z",
					LastRefreshGeneration: 7,
					LastError:             "forced refresh failure",
				},
			}, nil
		}

		withWorkingDirectory(t, repoRoot, func() {
			var stdout bytes.Buffer
			if err := NewRootCommand().Execute([]string{"watch", "status"}, &stdout); err != nil {
				t.Fatalf("Execute(watch status) error = %v", err)
			}
			output := stdout.String()
			assertContains(t, output, "repository root: "+repoRoot)
			assertContains(t, output, "status: stale")
			assertContains(t, output, "reason: watch heartbeat is stale")
			assertContains(t, output, "pid: 42")
			assertContains(t, output, "last refresh generation: 7")
			assertContains(t, output, "last error: forced refresh failure")
		})
	})
}

func TestWatchCommandErrors(t *testing.T) {
	repoRoot := initCLIRepo(t)

	t.Run("watch requires subcommand", func(t *testing.T) {
		var stdout bytes.Buffer
		err := NewRootCommand().Execute([]string{"watch"}, &stdout)
		if err == nil || err.Error() != "watch requires a subcommand" {
			t.Fatalf("Execute(watch) error = %v", err)
		}
		assertContains(t, stdout.String(), "optimusctx watch <command>")
	})

	t.Run("rejects unsupported run flag", func(t *testing.T) {
		var stdout bytes.Buffer
		err := NewRootCommand().Execute([]string{"watch", "run", "--once"}, &stdout)
		if err == nil || err.Error() != "watch run does not accept flags; got \"--once\"" {
			t.Fatalf("Execute(watch run --once) error = %v", err)
		}
		if stdout.Len() != 0 {
			t.Fatalf("stdout = %q, want empty", stdout.String())
		}
	})

	t.Run("rejects unsupported status argument", func(t *testing.T) {
		var stdout bytes.Buffer
		err := NewRootCommand().Execute([]string{"watch", "status", "now"}, &stdout)
		if err == nil || err.Error() != "watch status does not accept arguments; got \"now\"" {
			t.Fatalf("Execute(watch status now) error = %v", err)
		}
		if stdout.Len() != 0 {
			t.Fatalf("stdout = %q, want empty", stdout.String())
		}
	})

	t.Run("returns status service error", func(t *testing.T) {
		previous := watchStatusCommandService
		t.Cleanup(func() {
			watchStatusCommandService = previous
		})
		watchStatusCommandService = func(ctx context.Context, workingDir string) (repository.WatchStatusResult, error) {
			return repository.WatchStatusResult{}, errors.New("boom")
		}

		withWorkingDirectory(t, repoRoot, func() {
			var stdout bytes.Buffer
			err := NewRootCommand().Execute([]string{"watch", "status"}, &stdout)
			if err == nil || err.Error() != "boom" {
				t.Fatalf("Execute(watch status) error = %v, want boom", err)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
		})
	})
}

func TestRenderWatchValue(t *testing.T) {
	if got := renderWatchValue(""); got != "n/a" {
		t.Fatalf("renderWatchValue(\"\") = %q, want n/a", got)
	}
	if got := renderWatchValue("value"); got != "value" {
		t.Fatalf("renderWatchValue(value) = %q, want value", got)
	}
}

func TestFormatWatchStatusAbsentDefaults(t *testing.T) {
	output := formatWatchStatus(repository.WatchStatusResult{
		RepositoryRoot: "/repo",
		StatusPath:     "/repo/.optimusctx/tmp/watch-status.json",
		Status:         repository.WatchStatusKindAbsent,
		Reason:         "watch status file not found",
	})
	for _, fragment := range []string{
		"status: absent",
		"pid: 0",
		"started at: n/a",
		"last error: n/a",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("output = %q, want fragment %q", output, fragment)
		}
	}
}
