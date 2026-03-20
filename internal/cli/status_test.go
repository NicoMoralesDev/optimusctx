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

func TestStatusCommandCanonicalReport(t *testing.T) {
	repoRoot := initCLIRepo(t)

	previous := statusCommandService
	t.Cleanup(func() { statusCommandService = previous })

	statusCommandService = func(ctx context.Context, workingDir string) (repository.DoctorReport, error) {
		return repository.DoctorReport{
			Identity: repository.LayeredContextRepositoryIdentity{
				RootPath:      repoRoot,
				DetectionMode: "git",
			},
			Repository: repository.LayeredContextEnvelope{
				RepositoryRoot: repoRoot,
				Generation:     7,
				Freshness:      repository.FreshnessStatusFresh,
			},
			Install: repository.DoctorInstallSection{
				BinaryVersion: "dev",
				WorkingDir:    repoRoot,
			},
			State: repository.DoctorStateSection{
				Status: repository.DoctorStatusHealthy,
				Layout: repository.HealthStateLayout{
					StateDir:     repository.HealthPathStatus{Path: repoRoot + "/.optimusctx", Exists: true},
					DatabaseFile: repository.HealthPathStatus{Path: repoRoot + "/.optimusctx/db.sqlite", Exists: true},
				},
				Metadata:        repository.HealthStateMetadata{Present: true},
				RepositoryMatch: true,
			},
			Refresh: repository.DoctorRefreshSection{
				Status: repository.DoctorStatusHealthy,
				Health: repository.HealthRefreshDiagnostics{
					Present:                true,
					LastRefreshStatus:      repository.RefreshRunStatusSuccess,
					LastRefreshCompletedAt: time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC),
				},
			},
			Watch: repository.DoctorWatchSection{
				Status:  repository.DoctorStatusHealthy,
				Summary: "runtime watch loop is running",
				Health: repository.WatchStatusResult{
					Status:     repository.WatchStatusKindRunning,
					Reason:     "watch process heartbeat is current",
					StatusPath: repoRoot + "/.optimusctx/tmp/watch-status.json",
				},
			},
			Budget: repository.DoctorBudgetSection{
				Status: repository.DoctorStatusHealthy,
				Policy: repository.BudgetEstimatePolicy{Name: "bytes_div_4_ceiling"},
				Summary: repository.BudgetAnalysisSummary{
					ReturnedCount: 1,
				},
			},
			MCPReadiness: repository.DoctorMCPReadinessSection{
				Status:           repository.DoctorStatusHealthy,
				ServerName:       repository.DefaultMCPServerName,
				ServeCommand:     repository.NewServeCommand(""),
				SnippetAvailable: true,
			},
			MCPActivity: repository.DoctorMCPActivitySection{
				Status:             repository.DoctorStatusHealthy,
				LastSessionStartAt: time.Date(2026, 3, 19, 12, 1, 0, 0, time.UTC),
				LastInitializeAt:   time.Date(2026, 3, 19, 12, 1, 1, 0, time.UTC),
				LastToolsListAt:    time.Date(2026, 3, 19, 12, 1, 2, 0, time.UTC),
				LastToolCallAt:     time.Date(2026, 3, 19, 12, 1, 3, 0, time.UTC),
				RecentToolCalls: []repository.MCPObservedToolCall{
					{Name: "optimusctx.repository_map", At: "2026-03-19T12:01:03Z"},
				},
			},
			HostMCP: repository.DoctorHostRegistrationSection{
				Status: repository.DoctorStatusHealthy,
				Hosts: []repository.DoctorHostRegistration{
					{
						Client:               repository.SupportedClient{ID: repository.ClientCodexCLI, DisplayName: "Codex CLI"},
						RegistrationState:    repository.HostRegistrationDetected,
						RegistrationEvidence: "found OptimusCtx in shared Codex config",
						RegistrationPath:     "/home/test/.codex/config.toml",
						GuidanceState:        repository.GuidanceStateConfigured,
						GuidanceEvidence:     "found managed Codex guidance block",
						GuidancePath:         "/home/test/.codex/AGENTS.md",
					},
				},
			},
			Summary: repository.DoctorSummary{
				Status: repository.DoctorStatusHealthy,
			},
		}, nil
	}

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"status"}, &stdout); err != nil {
			t.Fatalf("Execute(status) error = %v", err)
		}
		output := stdout.String()
		for _, want := range []string{
			"overall status: healthy",
			"repository root: " + repoRoot,
			"MCP Host Registration",
			"Codex CLI registration=detected guidance=configured",
			"MCP Evidence",
			"last initialize: 2026-03-19T12:01:01Z",
			"recent tool call: optimusctx.repository_map at 2026-03-19T12:01:03Z",
			"serve command: optimusctx run",
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
	if !strings.Contains(output, "optimusctx status") {
		t.Fatalf("help output missing status usage:\n%s", output)
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
	t.Run("rejects unsupported arg", func(t *testing.T) {
		var stdout bytes.Buffer
		err := NewRootCommand().Execute([]string{"status", "now"}, &stdout)
		if err == nil || err.Error() != "status does not accept argument \"now\"" {
			t.Fatalf("Execute(status now) error = %v", err)
		}
	})

	t.Run("returns service error", func(t *testing.T) {
		previous := statusCommandService
		t.Cleanup(func() { statusCommandService = previous })
		statusCommandService = func(ctx context.Context, workingDir string) (repository.DoctorReport, error) {
			return repository.DoctorReport{}, errors.New("boom")
		}
		var stdout bytes.Buffer
		err := NewRootCommand().Execute([]string{"status"}, &stdout)
		if err == nil || err.Error() != "boom" {
			t.Fatalf("Execute(status) error = %v", err)
		}
	})
}
