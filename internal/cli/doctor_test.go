package cli

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
)

func TestDoctorCommand(t *testing.T) {
	original := doctorCommandService
	t.Cleanup(func() {
		doctorCommandService = original
	})

	doctorCommandService = func(ctx context.Context, workingDir string) (repository.DoctorReport, error) {
		return repository.DoctorReport{
			Identity: repository.LayeredContextRepositoryIdentity{
				RootPath:      "/repo",
				DetectionMode: "git",
			},
			Repository: repository.LayeredContextEnvelope{
				RepositoryRoot: "/repo",
				Generation:     7,
				Freshness:      repository.FreshnessStatusFresh,
			},
			Install: repository.DoctorInstallSection{
				Status:        repository.DoctorStatusHealthy,
				BinaryVersion: "dev",
				WorkingDir:    "/repo",
			},
			State: repository.DoctorStateSection{
				Status: repository.DoctorStatusHealthy,
				Layout: repository.HealthStateLayout{
					StateDir:     repository.HealthPathStatus{Path: "/repo/.optimusctx", Exists: true},
					DatabaseFile: repository.HealthPathStatus{Path: "/repo/.optimusctx/db.sqlite", Exists: true},
				},
				Metadata:        repository.HealthStateMetadata{Present: true},
				RepositoryMatch: true,
			},
			Refresh: repository.DoctorRefreshSection{
				Status: repository.DoctorStatusHealthy,
				Health: repository.HealthRefreshDiagnostics{
					Present:                true,
					LastRefreshStatus:      repository.RefreshRunStatusSuccess,
					LastRefreshCompletedAt: time.Date(2026, 3, 15, 18, 0, 0, 0, time.UTC),
				},
				LastRun: repository.DoctorRefreshRun{
					Present:    true,
					Generation: 7,
					Reason:     repository.RefreshReasonManual,
					Status:     repository.RefreshRunStatusSuccess,
				},
			},
			Watch: repository.DoctorWatchSection{
				Status:  repository.DoctorStatusHealthy,
				Summary: "runtime watch loop is running",
				Health: repository.WatchStatusResult{
					Status:     repository.WatchStatusKindRunning,
					Reason:     "watch process heartbeat is current",
					StatusPath: "/repo/.optimusctx/tmp/watch-status.json",
					Record: repository.WatchStatusRecord{
						LastHeartbeatAt: "2026-03-15T18:00:00Z",
					},
				},
			},
			Structural: repository.DoctorStructuralSection{
				Status: repository.DoctorStatusHealthy,
				Summary: repository.RepositoryStructuralCoverageSummary{
					IncludedFileCount:    5,
					FilesWithCoverageGap: 0,
				},
			},
			Budget: repository.DoctorBudgetSection{
				Status: repository.DoctorStatusHealthy,
				Policy: repository.BudgetEstimatePolicy{Name: "bytes_div_4_ceiling"},
				Summary: repository.BudgetAnalysisSummary{
					ReturnedCount: 2,
				},
				Hotspots: []repository.BudgetHotspot{
					{Path: "pkg/alpha.go", EstimatedTokens: 120, IncludedFileCount: 1, TotalSizeBytes: 480},
				},
			},
			MCPReadiness: repository.DoctorMCPReadinessSection{
				Status:           repository.DoctorStatusHealthy,
				ServerName:       repository.DefaultMCPServerName,
				ServeCommand:     repository.NewServeCommand(""),
				SnippetAvailable: true,
			},
			Summary: repository.DoctorSummary{
				Status: repository.DoctorStatusHealthy,
			},
		}, nil
	}

	var stdout bytes.Buffer
	if err := newDoctorCommand().Run(&stdout, nil); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := stdout.String()
	for _, want := range []string{
		"overall status: healthy",
		"repository root: /repo",
		"watch state: running",
		"summary: runtime watch loop is running",
		"optional: no",
		"serve command: optimusctx run",
		"hotspot: pkg/alpha.go tokens=120 files=1 bytes=480",
		"item: none",
		"step: none",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
}

func TestDoctorCommandRendersSupportedClientMCPAction(t *testing.T) {
	report := repository.DoctorReport{
		Identity: repository.LayeredContextRepositoryIdentity{
			RootPath:      "/repo",
			DetectionMode: "git",
		},
		Install: repository.DoctorInstallSection{
			BinaryVersion: "dev",
			WorkingDir:    "/repo",
		},
		MCPReadiness: repository.DoctorMCPReadinessSection{
			Status:       repository.DoctorStatusDegraded,
			ServerName:   repository.DefaultMCPServerName,
			ServeCommand: repository.NewServeCommand(""),
		},
		Summary: repository.DoctorSummary{
			Status: repository.DoctorStatusDegraded,
			Issues: []repository.DoctorIssue{
				{
					Section: "mcp",
					Summary: "snippet preview could not be rendered",
					Action:  "use `optimusctx status --client <client> [--write]` for claude-desktop, claude-cli, codex-app, or codex-cli to validate or register the MCP contract",
				},
			},
		},
		RecommendedFix: []string{
			"use `optimusctx status --client <client> [--write]` for claude-desktop, claude-cli, codex-app, or codex-cli to validate or register the MCP contract",
		},
	}

	output := formatDoctorReport(report)
	for _, want := range []string{
		"serve command: optimusctx run",
		"item: mcp: snippet preview could not be rendered; next action: use `optimusctx status --client <client> [--write]` for claude-desktop, claude-cli, codex-app, or codex-cli to validate or register the MCP contract",
		"step: use `optimusctx status --client <client> [--write]` for claude-desktop, claude-cli, codex-app, or codex-cli to validate or register the MCP contract",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
}

func TestDoctorCommandHealthyWithoutWatch(t *testing.T) {
	report := repository.DoctorReport{
		Identity: repository.LayeredContextRepositoryIdentity{
			RootPath:      "/repo",
			DetectionMode: "git",
		},
		Install: repository.DoctorInstallSection{
			BinaryVersion: "dev",
			WorkingDir:    "/repo",
		},
		Watch: repository.DoctorWatchSection{
			Status:   repository.DoctorStatusHealthy,
			Optional: true,
			Summary:  "runtime watch loop is not running",
			Health: repository.WatchStatusResult{
				Status:     repository.WatchStatusKindAbsent,
				Reason:     "watch status file not found",
				StatusPath: "/repo/.optimusctx/tmp/watch-status.json",
			},
		},
		Summary: repository.DoctorSummary{
			Status: repository.DoctorStatusHealthy,
		},
	}

	output := formatDoctorReport(report)
	for _, want := range []string{
		"overall status: healthy",
		"status: healthy",
		"watch state: absent",
		"summary: runtime watch loop is not running",
		"optional: yes",
		"reason: watch status file not found because `optimusctx run` is not active",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
}

func TestDoctorCommandRendersStaleFreshnessSignals(t *testing.T) {
	report := repository.DoctorReport{
		Identity: repository.LayeredContextRepositoryIdentity{
			RootPath:      "/repo",
			DetectionMode: "git",
		},
		Repository: repository.LayeredContextEnvelope{
			RepositoryRoot: "/repo",
			Generation:     7,
			Freshness:      repository.FreshnessStatusStale,
		},
		Install: repository.DoctorInstallSection{
			BinaryVersion: "dev",
			WorkingDir:    "/repo",
		},
		Refresh: repository.DoctorRefreshSection{
			Status: repository.DoctorStatusDegraded,
			Health: repository.HealthRefreshDiagnostics{
				Present:                true,
				LastRefreshStatus:      repository.RefreshRunStatusSuccess,
				Freshness:              repository.FreshnessStatusStale,
				FreshnessReason:        "workspace changed after the last successful refresh",
				LastRefreshCompletedAt: time.Date(2026, 3, 15, 18, 0, 0, 0, time.UTC),
			},
			LastRun: repository.DoctorRefreshRun{
				Present:    true,
				Generation: 7,
				Reason:     repository.RefreshReasonManual,
				Status:     repository.RefreshRunStatusSuccess,
			},
		},
		Watch: repository.DoctorWatchSection{
			Status:   repository.DoctorStatusHealthy,
			Optional: true,
			Summary:  "runtime watch loop is not running",
			Health: repository.WatchStatusResult{
				Status:     repository.WatchStatusKindAbsent,
				Reason:     "watch status file not found",
				StatusPath: "/repo/.optimusctx/tmp/watch-status.json",
			},
		},
		Summary: repository.DoctorSummary{
			Status: repository.DoctorStatusDegraded,
			Issues: []repository.DoctorIssue{
				{Section: "refresh", Summary: "repository freshness is stale", Action: "run `optimusctx run` and inspect `.optimusctx/logs/` if refresh stays degraded"},
			},
		},
		RecommendedFix: []string{
			"run `optimusctx run` and inspect `.optimusctx/logs/` if refresh stays degraded",
		},
	}

	output := formatDoctorReport(report)
	for _, want := range []string{
		"overall status: degraded",
		"freshness: stale",
		"freshness reason: workspace changed after the last successful refresh",
		"item: refresh: repository freshness is stale; next action: run `optimusctx run` and inspect `.optimusctx/logs/` if refresh stays degraded",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
}

func TestDoctorSmokeVerification(t *testing.T) {
	repoRoot := initCLIRepo(t)
	writeCLIFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc main() {}\n")

	withWorkingDirectory(t, repoRoot, func() {
		var initStdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"init"}, &initStdout); err != nil {
			t.Fatalf("Execute(init) error = %v", err)
		}

		var doctorStdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"doctor"}, &doctorStdout); err != nil {
			t.Fatalf("Execute(doctor) error = %v", err)
		}

		output := doctorStdout.String()
		for _, want := range []string{
			"overall status: healthy",
			"runtime version: dev",
			"freshness: fresh",
			"snippet available: yes",
			"item: none",
		} {
			if !strings.Contains(output, want) {
				t.Fatalf("doctor output missing %q:\n%s", want, output)
			}
		}
	})
}
