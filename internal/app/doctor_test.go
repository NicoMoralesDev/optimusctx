package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
)

func TestDoctorReportSections(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.go"), "package pkg\n\nfunc Alpha() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "beta.go"), "package pkg\n\nfunc Beta() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "docs", "guide.md"), strings.Repeat("guide\n", 20))

	if _, err := NewRefreshService().Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	now := time.Date(2026, 3, 15, 18, 0, 0, 0, time.UTC)
	writeDoctorWatchStatus(t, repoRoot, repository.NewWatchStatusRecord(4242, repoRoot, "dev", now))

	service := NewDoctorService()
	service.Getwd = func() (string, error) { return repoRoot, nil }
	service.WatchService.Now = func() time.Time { return now.Add(2 * time.Second) }
	service.WatchService.ProcessRunning = func(pid int) bool { return pid == 4242 }

	report, err := service.Doctor(context.Background(), repoRoot, repository.DoctorRequest{BudgetLimit: 3})
	if err != nil {
		t.Fatalf("Doctor() error = %v", err)
	}

	if report.Summary.Status != repository.DoctorStatusHealthy {
		t.Fatalf("Summary.Status = %q, want %q", report.Summary.Status, repository.DoctorStatusHealthy)
	}
	if report.State.Status != repository.DoctorStatusHealthy {
		t.Fatalf("State.Status = %q, want healthy", report.State.Status)
	}
	if report.Refresh.Status != repository.DoctorStatusHealthy {
		t.Fatalf("Refresh.Status = %q, want healthy", report.Refresh.Status)
	}
	if report.Watch.Status != repository.DoctorStatusHealthy {
		t.Fatalf("Watch.Status = %q, want healthy", report.Watch.Status)
	}
	if report.Structural.Status != repository.DoctorStatusHealthy {
		t.Fatalf("Structural.Status = %q, want healthy", report.Structural.Status)
	}
	if report.Budget.Status != repository.DoctorStatusHealthy {
		t.Fatalf("Budget.Status = %q, want healthy", report.Budget.Status)
	}
	if report.MCPReadiness.Status != repository.DoctorStatusHealthy {
		t.Fatalf("MCPReadiness.Status = %q, want healthy", report.MCPReadiness.Status)
	}
	if report.Refresh.LastRun.Status != repository.RefreshRunStatusSuccess {
		t.Fatalf("Refresh.LastRun.Status = %q, want success", report.Refresh.LastRun.Status)
	}
	if len(report.Budget.Hotspots) == 0 {
		t.Fatal("Budget.Hotspots should not be empty")
	}
	if report.MCPReadiness.SnippetDocument.MCPServers[repository.DefaultMCPServerName].Command != repository.DefaultServeCommandName {
		t.Fatalf("MCP command = %+v, want %q", report.MCPReadiness.SnippetDocument.MCPServers, repository.DefaultServeCommandName)
	}
}

func TestDoctorDetectsDegradedState(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "healthy.go"), "package pkg\n\nfunc Healthy() {}\n")
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "partial.go"), "package pkg\n\nfunc Recovered() {}\nfunc (\n")
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "failed.go"), "package pkg\nfunc (\n")
	writeRepoFile(t, filepath.Join(repoRoot, "notes.txt"), "plain text\n")

	refresh := NewRefreshService()
	initial, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	})
	if err != nil {
		t.Fatalf("Refresh() baseline error = %v", err)
	}

	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "healthy.go"), "package pkg\n\nfunc Healthy() {}\nfunc Mutated() {}\n")
	failed, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonManual,
		InjectFailure: func(stage string) error {
			if stage == "after_files" {
				return errors.New("forced after file updates")
			}
			return nil
		},
	})
	if err == nil {
		t.Fatal("Refresh() error = nil, want injected failure")
	}
	if failed.FreshnessStatus != repository.FreshnessStatusPartiallyDegraded {
		t.Fatalf("failed freshness = %q, want %q", failed.FreshnessStatus, repository.FreshnessStatusPartiallyDegraded)
	}
	if failed.Generation != initial.Generation+1 {
		t.Fatalf("failed generation = %d, want %d", failed.Generation, initial.Generation+1)
	}

	staleAt := time.Date(2026, 3, 15, 17, 0, 0, 0, time.UTC)
	record := repository.NewWatchStatusRecord(777, repoRoot, "dev", staleAt)
	record = record.WithHeartbeat(repository.WatchHeartbeat{
		At:                    staleAt,
		LastRefreshDoneAt:     staleAt,
		LastRefreshGeneration: initial.Generation,
		LastError:             "watch observer overflowed; falling back to full refresh",
	})
	writeDoctorWatchStatus(t, repoRoot, record)

	service := NewDoctorService()
	service.Getwd = func() (string, error) { return repoRoot, nil }
	service.WatchService.Now = func() time.Time { return staleAt.Add(20 * time.Second) }
	service.WatchService.ProcessRunning = func(pid int) bool { return true }

	report, err := service.Doctor(context.Background(), repoRoot, repository.DoctorRequest{BudgetLimit: 2})
	if err != nil {
		t.Fatalf("Doctor() error = %v", err)
	}

	if report.Summary.Status != repository.DoctorStatusDegraded {
		t.Fatalf("Summary.Status = %q, want degraded", report.Summary.Status)
	}
	if report.Refresh.Status != repository.DoctorStatusDegraded {
		t.Fatalf("Refresh.Status = %q, want degraded", report.Refresh.Status)
	}
	if report.Refresh.Health.Freshness != repository.FreshnessStatusPartiallyDegraded {
		t.Fatalf("Refresh.Health.Freshness = %q, want partially_degraded", report.Refresh.Health.Freshness)
	}
	if report.Refresh.LastRun.Status != repository.RefreshRunStatusFailed {
		t.Fatalf("Refresh.LastRun.Status = %q, want failed", report.Refresh.LastRun.Status)
	}
	if !strings.Contains(report.Refresh.LastRun.FailureReason, "forced after file updates") {
		t.Fatalf("Refresh.LastRun.FailureReason = %q", report.Refresh.LastRun.FailureReason)
	}
	if report.Watch.Status != repository.DoctorStatusDegraded {
		t.Fatalf("Watch.Status = %q, want degraded", report.Watch.Status)
	}
	if report.Watch.Health.Status != repository.WatchStatusKindStale {
		t.Fatalf("Watch.Health.Status = %q, want stale", report.Watch.Health.Status)
	}
	if report.Structural.Status != repository.DoctorStatusDegraded {
		t.Fatalf("Structural.Status = %q, want degraded", report.Structural.Status)
	}
	if report.Structural.Summary.PartialCount == 0 || report.Structural.Summary.FailedCount == 0 {
		t.Fatalf("Structural summary = %+v, want partial and failed counts", report.Structural.Summary)
	}
	if len(report.Structural.Examples) == 0 {
		t.Fatal("Structural.Examples should not be empty")
	}
	if len(report.Summary.Issues) < 3 {
		t.Fatalf("Issues = %+v, want refresh/watch/structural issues", report.Summary.Issues)
	}
	if len(report.RecommendedFix) == 0 {
		t.Fatal("RecommendedFix should not be empty")
	}
}

func TestDoctorHealthyWithoutWatch(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.go"), "package pkg\n\nfunc Alpha() {}\n")

	if _, err := NewRefreshService().Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	}); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	service := NewDoctorService()
	service.Getwd = func() (string, error) { return repoRoot, nil }

	report, err := service.Doctor(context.Background(), repoRoot, repository.DoctorRequest{BudgetLimit: 2})
	if err != nil {
		t.Fatalf("Doctor() error = %v", err)
	}

	if report.Summary.Status != repository.DoctorStatusHealthy {
		t.Fatalf("Summary.Status = %q, want healthy", report.Summary.Status)
	}
	if report.Watch.Status != repository.DoctorStatusHealthy {
		t.Fatalf("Watch.Status = %q, want healthy", report.Watch.Status)
	}
	if report.Watch.Health.Status != repository.WatchStatusKindAbsent {
		t.Fatalf("Watch.Health.Status = %q, want absent", report.Watch.Health.Status)
	}
	if !report.Watch.Optional {
		t.Fatal("Watch.Optional = false, want true")
	}
	if got, want := report.Watch.Summary, "watch mode is not running; background watch is optional"; got != want {
		t.Fatalf("Watch.Summary = %q, want %q", got, want)
	}
	if len(report.Summary.Issues) != 0 {
		t.Fatalf("Issues = %+v, want none", report.Summary.Issues)
	}
}

func TestDoctorDetectsStaleWatch(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "alpha.go"), "package pkg\n\nfunc Alpha() {}\n")

	refreshed, err := NewRefreshService().Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	})
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	staleAt := time.Date(2026, 3, 15, 18, 0, 0, 0, time.UTC)
	record := repository.NewWatchStatusRecord(777, repoRoot, "dev", staleAt)
	record = record.WithHeartbeat(repository.WatchHeartbeat{
		At:                    staleAt,
		LastRefreshDoneAt:     staleAt,
		LastRefreshGeneration: refreshed.Generation,
	})
	writeDoctorWatchStatus(t, repoRoot, record)

	service := NewDoctorService()
	service.Getwd = func() (string, error) { return repoRoot, nil }
	service.WatchService.Now = func() time.Time { return staleAt.Add(20 * time.Second) }
	service.WatchService.ProcessRunning = func(pid int) bool { return true }

	report, err := service.Doctor(context.Background(), repoRoot, repository.DoctorRequest{BudgetLimit: 2})
	if err != nil {
		t.Fatalf("Doctor() error = %v", err)
	}

	if report.Summary.Status != repository.DoctorStatusDegraded {
		t.Fatalf("Summary.Status = %q, want degraded", report.Summary.Status)
	}
	if report.Watch.Status != repository.DoctorStatusDegraded {
		t.Fatalf("Watch.Status = %q, want degraded", report.Watch.Status)
	}
	if report.Watch.Health.Status != repository.WatchStatusKindStale {
		t.Fatalf("Watch.Health.Status = %q, want stale", report.Watch.Health.Status)
	}
	if report.Watch.Optional {
		t.Fatal("Watch.Optional = true, want false")
	}
}

func writeDoctorWatchStatus(t *testing.T, repoRoot string, record repository.WatchStatusRecord) {
	t.Helper()

	layout, err := state.ResolveLayout(repoRoot)
	if err != nil {
		t.Fatalf("ResolveLayout(%q) error = %v", repoRoot, err)
	}
	if err := os.MkdirAll(layout.TmpDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", layout.TmpDir, err)
	}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent(record) error = %v", err)
	}
	data = append(data, '\n')
	statusPath := filepath.Join(layout.TmpDir, repository.DefaultWatchStatusFilename)
	if err := os.WriteFile(statusPath, data, 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", statusPath, err)
	}
}
