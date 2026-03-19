package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/app"
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
			"next action: runtime is ready; point your MCP client at `optimusctx run`",
		} {
			if !strings.Contains(output, want) {
				t.Fatalf("status output missing %q:\n%s", want, output)
			}
		}
	})
}

func TestStatusCommandClientPreview(t *testing.T) {
	repoRoot := initCLIRepo(t)

	previousHealth := statusHealthService
	previousWatch := statusWatchService
	previousInstall := statusInstallPreviewService
	t.Cleanup(func() {
		statusHealthService = previousHealth
		statusWatchService = previousWatch
		statusInstallPreviewService = previousInstall
	})

	statusHealthService = func(ctx context.Context, workingDir string) (repository.HealthResult, error) {
		return repository.HealthResult{
			Repository: repository.LayeredContextEnvelope{RepositoryRoot: repoRoot, Generation: 1, Freshness: repository.FreshnessStatusFresh},
			Summary:    repository.HealthSummary{StateStatus: repository.HealthStateStatusReady},
		}, nil
	}
	statusWatchService = func(ctx context.Context, workingDir string) (repository.WatchStatusResult, error) {
		return repository.WatchStatusResult{RepositoryRoot: repoRoot, Status: repository.WatchStatusKindAbsent, Reason: "watch status file not found"}, nil
	}
	statusInstallPreviewService = func(ctx context.Context, request app.InstallRequest) (app.InstallResult, error) {
		return app.InstallResult{
			Rendered: repository.RenderedClientConfig{
				Client:     repository.SupportedClients()[0],
				ConfigPath: "/tmp/claude.json",
				Mode:       repository.RenderModePreview,
				Content:    "{\n  \"mcpServers\": {\n    \"optimusctx\": {\n      \"command\": \"optimusctx\",\n      \"args\": [\"run\"]\n    }\n  }\n}\n",
			},
		}, nil
	}

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"status", "--client", "claude-desktop"}, &stdout); err != nil {
			t.Fatalf("Execute(status --client) error = %v", err)
		}
		output := stdout.String()
		for _, want := range []string{
			"client: Claude Desktop",
			"config path: /tmp/claude.json",
			"mode: preview",
			"\"args\": [\"run\"]",
			"status: preview only",
		} {
			if !strings.Contains(output, want) {
				t.Fatalf("status client preview missing %q:\n%s", want, output)
			}
		}
	})
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
