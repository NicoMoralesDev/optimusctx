package cli

import (
	"bytes"
	"context"
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
				Summary: "watch mode is running",
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
		"summary: watch mode is running",
		"optional: no",
		"hotspot: pkg/alpha.go tokens=120 files=1 bytes=480",
		"item: none",
		"step: none",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
}

func TestDoctorCommandRendersDegradedAndMissingSignals(t *testing.T) {
	report := repository.DoctorReport{
		Identity: repository.LayeredContextRepositoryIdentity{
			RootPath:      "/repo",
			DetectionMode: "git",
		},
		Install: repository.DoctorInstallSection{
			BinaryVersion: "dev",
			WorkingDir:    "/repo",
		},
		State: repository.DoctorStateSection{
			Status: repository.DoctorStatusMissing,
			Layout: repository.HealthStateLayout{
				StateDir:     repository.HealthPathStatus{Path: "/repo/.optimusctx"},
				DatabaseFile: repository.HealthPathStatus{Path: "/repo/.optimusctx/db.sqlite"},
			},
		},
		Refresh: repository.DoctorRefreshSection{
			Status: repository.DoctorStatusDegraded,
			Health: repository.HealthRefreshDiagnostics{
				Present:           true,
				LastRefreshStatus: repository.RefreshRunStatusFailed,
				Freshness:         repository.FreshnessStatusPartiallyDegraded,
				FreshnessReason:   "refresh failed",
			},
			LastRun: repository.DoctorRefreshRun{
				Present:       true,
				Generation:    8,
				Reason:        repository.RefreshReasonManual,
				Status:        repository.RefreshRunStatusFailed,
				FailureReason: "forced after file updates",
			},
		},
		Watch: repository.DoctorWatchSection{
			Status:  repository.DoctorStatusDegraded,
			Summary: "watch heartbeat is stale",
			Health: repository.WatchStatusResult{
				Status:     repository.WatchStatusKindStale,
				Reason:     "watch heartbeat is stale",
				StatusPath: "/repo/.optimusctx/tmp/watch-status.json",
				Record: repository.WatchStatusRecord{
					LastError: "watch observer overflowed; falling back to full refresh",
				},
			},
		},
		Structural: repository.DoctorStructuralSection{
			Status: repository.DoctorStatusDegraded,
			Summary: repository.RepositoryStructuralCoverageSummary{
				IncludedFileCount:    4,
				FilesWithCoverageGap: 3,
				FailedCount:          1,
				PartialCount:         1,
			},
			Examples: []repository.DoctorStructuralCoverageExample{
				{Path: "pkg/partial.go", CoverageState: repository.ExtractionCoverageStatePartial, CoverageReason: repository.ExtractionCoverageReasonParseError, SymbolCount: 1},
			},
		},
		Budget: repository.DoctorBudgetSection{
			Status: repository.DoctorStatusMissing,
		},
		MCPReadiness: repository.DoctorMCPReadinessSection{
			Status:           repository.DoctorStatusHealthy,
			ServerName:       repository.DefaultMCPServerName,
			ServeCommand:     repository.NewServeCommand(""),
			SnippetAvailable: true,
		},
		Summary: repository.DoctorSummary{
			Status: repository.DoctorStatusDegraded,
			Issues: []repository.DoctorIssue{
				{Section: "state", Summary: "repository state directory is not initialized", Action: "run `optimusctx init` from the repository root to create `.optimusctx/`"},
				{Section: "refresh", Summary: "last refresh failed: forced after file updates", Action: "run `optimusctx refresh` and inspect `.optimusctx/logs/` if refresh stays degraded"},
			},
		},
		RecommendedFix: []string{
			"run `optimusctx init` from the repository root to create `.optimusctx/`",
			"run `optimusctx refresh` and inspect `.optimusctx/logs/` if refresh stays degraded",
		},
	}

	output := formatDoctorReport(report)
	for _, want := range []string{
		"overall status: degraded",
		"status: missing",
		"latest run failure: forced after file updates",
		"summary: watch heartbeat is stale",
		"reason: watch heartbeat is stale",
		"gap: pkg/partial.go (partial, reason=parse_error, symbols=1)",
		"item: state: repository state directory is not initialized; next action: run `optimusctx init` from the repository root to create `.optimusctx/`",
		"step: run `optimusctx refresh` and inspect `.optimusctx/logs/` if refresh stays degraded",
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
			Summary:  "watch mode is not running; background watch is optional",
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
		"summary: watch mode is not running; background watch is optional",
		"optional: yes",
		"reason: watch status file not found because watch mode is not running",
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
			Summary:  "watch mode is not running; background watch is optional",
			Health: repository.WatchStatusResult{
				Status:     repository.WatchStatusKindAbsent,
				Reason:     "watch status file not found",
				StatusPath: "/repo/.optimusctx/tmp/watch-status.json",
			},
		},
		Summary: repository.DoctorSummary{
			Status: repository.DoctorStatusDegraded,
			Issues: []repository.DoctorIssue{
				{Section: "refresh", Summary: "repository freshness is stale", Action: "run `optimusctx refresh` and inspect `.optimusctx/logs/` if refresh stays degraded"},
			},
		},
		RecommendedFix: []string{
			"run `optimusctx refresh` and inspect `.optimusctx/logs/` if refresh stays degraded",
		},
	}

	output := formatDoctorReport(report)
	for _, want := range []string{
		"overall status: degraded",
		"freshness: stale",
		"freshness reason: workspace changed after the last successful refresh",
		"item: refresh: repository freshness is stale; next action: run `optimusctx refresh` and inspect `.optimusctx/logs/` if refresh stays degraded",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
}
