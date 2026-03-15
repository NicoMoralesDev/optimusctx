package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
)

func TestWatchRunnerLifecycle(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc Main() {}\n")

	layout, err := state.ResolveLayout(repoRoot)
	if err != nil {
		t.Fatalf("ResolveLayout(%q) error = %v", repoRoot, err)
	}
	if err := os.MkdirAll(layout.TmpDir, 0o755); err != nil {
		t.Fatalf("mkdir tmp dir: %v", err)
	}

	events := make(chan repository.WatchEvent, 1)
	nowTimes := []time.Time{
		time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 15, 12, 0, 1, 0, time.UTC),
		time.Date(2026, 3, 15, 12, 0, 2, 0, time.UTC),
		time.Date(2026, 3, 15, 12, 0, 3, 0, time.UTC),
	}
	nowIndex := 0
	refreshCalled := 0

	service := NewWatchService()
	service.Now = func() time.Time {
		if nowIndex >= len(nowTimes) {
			return nowTimes[len(nowTimes)-1]
		}
		value := nowTimes[nowIndex]
		nowIndex++
		return value
	}
	service.PID = func() int { return 4321 }
	service.HeartbeatInterval = 50 * time.Millisecond
	service.DefaultDebounce = 5 * time.Millisecond
	service.Observe = func(ctx context.Context, root string) (<-chan repository.WatchEvent, error) {
		if root != repoRoot {
			t.Fatalf("observer root = %q, want %q", root, repoRoot)
		}
		return events, nil
	}
	service.Refresh = func(ctx context.Context, request RefreshRequest) (RefreshResult, error) {
		refreshCalled++
		if request.StartPath != repoRoot {
			t.Fatalf("refresh StartPath = %q, want %q", request.StartPath, repoRoot)
		}
		if request.Reason != repository.RefreshReasonWatch {
			t.Fatalf("refresh Reason = %q, want %q", request.Reason, repository.RefreshReasonWatch)
		}
		return RefreshResult{Generation: 9}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := service.Run(ctx, repository.WatchRequest{
			StartPath:         repoRoot,
			HeartbeatInterval: 50 * time.Millisecond,
			DebounceWindow:    5 * time.Millisecond,
		})
		done <- err
	}()

	statusPath := filepath.Join(layout.TmpDir, repository.DefaultWatchStatusFilename)
	waitForFile(t, statusPath)
	events <- repository.WatchEvent{
		Path: "main.go",
		Op:   repository.WatchEventOpChange,
		At:   time.Date(2026, 3, 15, 12, 0, 1, 0, time.UTC),
	}
	waitForCondition(t, 2*time.Second, func() bool {
		record := loadWatchStatusRecord(t, statusPath)
		return record.LastRefreshGeneration == 9 && record.LastRefreshDoneAt != ""
	})

	record := loadWatchStatusRecord(t, statusPath)
	if record.PID != 4321 {
		t.Fatalf("record.PID = %d, want 4321", record.PID)
	}
	if record.RepoRoot != repoRoot {
		t.Fatalf("record.RepoRoot = %q, want %q", record.RepoRoot, repoRoot)
	}
	if record.LastEventAt == "" {
		t.Fatal("record.LastEventAt should be set after watch event")
	}
	if record.LastRefreshGeneration != 9 {
		t.Fatalf("record.LastRefreshGeneration = %d, want 9", record.LastRefreshGeneration)
	}
	if refreshCalled != 1 {
		t.Fatalf("refreshCalled = %d, want 1", refreshCalled)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not exit after cancel")
	}

	if _, err := os.Stat(statusPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("status file still exists after graceful stop: %v", err)
	}
}

func TestWatchRefreshUsesCanonicalPipeline(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "helper.go"), "package pkg\n")

	layout, err := state.ResolveLayout(repoRoot)
	if err != nil {
		t.Fatalf("ResolveLayout(%q) error = %v", repoRoot, err)
	}
	if err := os.MkdirAll(layout.TmpDir, 0o755); err != nil {
		t.Fatalf("mkdir tmp dir: %v", err)
	}

	events := make(chan repository.WatchEvent, 2)
	service := NewWatchService()
	service.HeartbeatInterval = 50 * time.Millisecond
	service.DefaultDebounce = 5 * time.Millisecond
	service.PID = func() int { return 4321 }
	service.Observe = func(ctx context.Context, root string) (<-chan repository.WatchEvent, error) {
		return events, nil
	}

	requests := make(chan RefreshRequest, 1)
	service.Refresh = func(ctx context.Context, request RefreshRequest) (RefreshResult, error) {
		requests <- request
		return RefreshResult{Generation: 12, FreshnessStatus: repository.FreshnessStatusFresh}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := service.Run(ctx, repository.WatchRequest{
			StartPath:         repoRoot,
			HeartbeatInterval: 50 * time.Millisecond,
			DebounceWindow:    5 * time.Millisecond,
		})
		done <- err
	}()

	statusPath := filepath.Join(layout.TmpDir, repository.DefaultWatchStatusFilename)
	waitForFile(t, statusPath)
	events <- repository.WatchEvent{Path: filepath.Join(repoRoot, "pkg", "helper.go"), Op: repository.WatchEventOpChange}

	var request RefreshRequest
	select {
	case request = <-requests:
	case <-time.After(2 * time.Second):
		t.Fatal("watch-triggered refresh was not invoked")
	}

	if request.StartPath != repoRoot {
		t.Fatalf("request.StartPath = %q, want %q", request.StartPath, repoRoot)
	}
	if request.Reason != repository.RefreshReasonWatch {
		t.Fatalf("request.Reason = %q, want %q", request.Reason, repository.RefreshReasonWatch)
	}
	if request.ForceFull {
		t.Fatal("request.ForceFull = true, want false")
	}
	if !reflect.DeepEqual(request.ChangedHint, []string{"pkg/helper.go"}) {
		t.Fatalf("request.ChangedHint = %#v, want pkg/helper.go", request.ChangedHint)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not exit after cancel")
	}
}

func TestWatchDebouncesBurstEvents(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "first.go"), "package pkg\n")
	writeRepoFile(t, filepath.Join(repoRoot, "pkg", "second.go"), "package pkg\n")

	layout, err := state.ResolveLayout(repoRoot)
	if err != nil {
		t.Fatalf("ResolveLayout(%q) error = %v", repoRoot, err)
	}
	if err := os.MkdirAll(layout.TmpDir, 0o755); err != nil {
		t.Fatalf("mkdir tmp dir: %v", err)
	}

	events := make(chan repository.WatchEvent, 4)
	service := NewWatchService()
	service.HeartbeatInterval = 50 * time.Millisecond
	service.DefaultDebounce = 20 * time.Millisecond
	service.Observe = func(ctx context.Context, root string) (<-chan repository.WatchEvent, error) {
		return events, nil
	}

	requests := make(chan RefreshRequest, 2)
	service.Refresh = func(ctx context.Context, request RefreshRequest) (RefreshResult, error) {
		requests <- request
		return RefreshResult{Generation: 3, FreshnessStatus: repository.FreshnessStatusFresh}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := service.Run(ctx, repository.WatchRequest{
			StartPath:         repoRoot,
			HeartbeatInterval: 50 * time.Millisecond,
			DebounceWindow:    20 * time.Millisecond,
		})
		done <- err
	}()

	waitForFile(t, filepath.Join(layout.TmpDir, repository.DefaultWatchStatusFilename))
	events <- repository.WatchEvent{Path: "pkg/first.go", Op: repository.WatchEventOpChange}
	events <- repository.WatchEvent{Path: "pkg/second.go", Op: repository.WatchEventOpChange}
	events <- repository.WatchEvent{Path: "pkg/first.go", Op: repository.WatchEventOpChange}

	var request RefreshRequest
	select {
	case request = <-requests:
	case <-time.After(2 * time.Second):
		t.Fatal("debounced refresh was not invoked")
	}

	select {
	case extra := <-requests:
		t.Fatalf("unexpected second refresh request: %#v", extra)
	case <-time.After(100 * time.Millisecond):
	}

	if request.ForceFull {
		t.Fatal("request.ForceFull = true, want false")
	}
	if !reflect.DeepEqual(request.ChangedHint, []string{"pkg/first.go", "pkg/second.go"}) {
		t.Fatalf("request.ChangedHint = %#v, want sorted unique hints", request.ChangedHint)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not exit after cancel")
	}
}

func TestWatchOverflowFallsBackToFullRefresh(t *testing.T) {
	repoRoot := initRepo(t)
	layout, err := state.ResolveLayout(repoRoot)
	if err != nil {
		t.Fatalf("ResolveLayout(%q) error = %v", repoRoot, err)
	}
	if err := os.MkdirAll(layout.TmpDir, 0o755); err != nil {
		t.Fatalf("mkdir tmp dir: %v", err)
	}

	events := make(chan repository.WatchEvent, 2)
	service := NewWatchService()
	service.HeartbeatInterval = 50 * time.Millisecond
	service.DefaultDebounce = 5 * time.Millisecond
	service.Observe = func(ctx context.Context, root string) (<-chan repository.WatchEvent, error) {
		return events, nil
	}

	requests := make(chan RefreshRequest, 1)
	service.Refresh = func(ctx context.Context, request RefreshRequest) (RefreshResult, error) {
		requests <- request
		return RefreshResult{Generation: 4, FreshnessStatus: repository.FreshnessStatusFresh}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := service.Run(ctx, repository.WatchRequest{
			StartPath:         repoRoot,
			HeartbeatInterval: 50 * time.Millisecond,
			DebounceWindow:    5 * time.Millisecond,
		})
		done <- err
	}()

	statusPath := filepath.Join(layout.TmpDir, repository.DefaultWatchStatusFilename)
	waitForFile(t, statusPath)
	events <- repository.WatchEvent{Path: "pkg/known.go", Op: repository.WatchEventOpChange}
	events <- repository.WatchEvent{Op: repository.WatchEventOpOverflow}

	var request RefreshRequest
	select {
	case request = <-requests:
	case <-time.After(2 * time.Second):
		t.Fatal("overflow fallback refresh was not invoked")
	}

	if !request.ForceFull {
		t.Fatal("request.ForceFull = false, want true")
	}
	if len(request.ChangedHint) != 0 {
		t.Fatalf("request.ChangedHint = %#v, want none", request.ChangedHint)
	}

	waitForCondition(t, 2*time.Second, func() bool {
		record := loadWatchStatusRecord(t, statusPath)
		return record.LastRefreshGeneration == 4 && record.LastError == ""
	})

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not exit after cancel")
	}
}

func TestWatchUncertainEventFallsBackToFullRefresh(t *testing.T) {
	repoRoot := initRepo(t)
	layout, err := state.ResolveLayout(repoRoot)
	if err != nil {
		t.Fatalf("ResolveLayout(%q) error = %v", repoRoot, err)
	}
	if err := os.MkdirAll(layout.TmpDir, 0o755); err != nil {
		t.Fatalf("mkdir tmp dir: %v", err)
	}

	events := make(chan repository.WatchEvent, 1)
	service := NewWatchService()
	service.HeartbeatInterval = 50 * time.Millisecond
	service.DefaultDebounce = 5 * time.Millisecond
	service.Observe = func(ctx context.Context, root string) (<-chan repository.WatchEvent, error) {
		return events, nil
	}

	requests := make(chan RefreshRequest, 1)
	service.Refresh = func(ctx context.Context, request RefreshRequest) (RefreshResult, error) {
		requests <- request
		return RefreshResult{Generation: 5, FreshnessStatus: repository.FreshnessStatusFresh}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := service.Run(ctx, repository.WatchRequest{
			StartPath:         repoRoot,
			HeartbeatInterval: 50 * time.Millisecond,
			DebounceWindow:    5 * time.Millisecond,
		})
		done <- err
	}()

	waitForFile(t, filepath.Join(layout.TmpDir, repository.DefaultWatchStatusFilename))
	events <- repository.WatchEvent{Path: "pkg/new-dir", Op: repository.WatchEventOpChange, Uncertain: true}

	var request RefreshRequest
	select {
	case request = <-requests:
	case <-time.After(2 * time.Second):
		t.Fatal("uncertain fallback refresh was not invoked")
	}
	if !request.ForceFull {
		t.Fatal("request.ForceFull = false, want true")
	}
	if len(request.ChangedHint) != 0 {
		t.Fatalf("request.ChangedHint = %#v, want none", request.ChangedHint)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not exit after cancel")
	}
}

func TestWatchRefreshFailureRecovery(t *testing.T) {
	repoRoot := initRepo(t)
	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc Main() {}\n")

	refresh := NewRefreshService()
	initial, err := refresh.Refresh(context.Background(), RefreshRequest{
		StartPath: repoRoot,
		Reason:    repository.RefreshReasonInit,
		ForceFull: true,
	})
	if err != nil {
		t.Fatalf("Refresh() baseline error = %v", err)
	}

	events := make(chan repository.WatchEvent, 2)
	service := NewWatchService()
	service.HeartbeatInterval = 50 * time.Millisecond
	service.DefaultDebounce = 5 * time.Millisecond
	service.Observe = func(ctx context.Context, root string) (<-chan repository.WatchEvent, error) {
		return events, nil
	}

	failNext := true
	service.Refresh = func(ctx context.Context, request RefreshRequest) (RefreshResult, error) {
		if failNext {
			failNext = false
			request.InjectFailure = func(stage string) error {
				if stage == "after_files" {
					return errors.New("forced watch failure")
				}
				return nil
			}
		}
		return refresh.Refresh(ctx, request)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := service.Run(ctx, repository.WatchRequest{
			StartPath:         repoRoot,
			HeartbeatInterval: 50 * time.Millisecond,
			DebounceWindow:    5 * time.Millisecond,
		})
		done <- err
	}()

	layout, err := state.ResolveLayout(repoRoot)
	if err != nil {
		t.Fatalf("ResolveLayout(%q) error = %v", repoRoot, err)
	}
	statusPath := filepath.Join(layout.TmpDir, repository.DefaultWatchStatusFilename)
	waitForFile(t, statusPath)

	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc Failed() {}\n")
	events <- repository.WatchEvent{Path: "main.go", Op: repository.WatchEventOpChange}
	waitForCondition(t, 2*time.Second, func() bool {
		record := loadWatchStatusRecord(t, statusPath)
		return record.LastError == "apply refresh plan: forced watch failure"
	})

	db := openStateDatabase(t, filepath.Join(layout.StateDir, "db.sqlite"))
	defer db.Close()
	assertRepositoryFreshness(t, db, repository.RefreshReasonWatch, repository.RefreshRunStatusFailed, repository.FreshnessStatusPartiallyDegraded, initial.Generation+1, initial.Generation)

	writeRepoFile(t, filepath.Join(repoRoot, "main.go"), "package main\n\nfunc Recovered() {}\n")
	events <- repository.WatchEvent{Path: "main.go", Op: repository.WatchEventOpChange}
	waitForCondition(t, 2*time.Second, func() bool {
		record := loadWatchStatusRecord(t, statusPath)
		return record.LastRefreshGeneration == initial.Generation+2 && record.LastError == ""
	})

	assertRepositoryFreshness(t, db, repository.RefreshReasonWatch, repository.RefreshRunStatusSuccess, repository.FreshnessStatusFresh, initial.Generation+2, initial.Generation+2)

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not exit after cancel")
	}
}

func TestWatchStatusStaleHeartbeat(t *testing.T) {
	repoRoot := initRepo(t)
	layout, err := state.ResolveLayout(repoRoot)
	if err != nil {
		t.Fatalf("ResolveLayout(%q) error = %v", repoRoot, err)
	}
	if err := os.MkdirAll(layout.TmpDir, 0o755); err != nil {
		t.Fatalf("mkdir tmp dir: %v", err)
	}

	statusPath := filepath.Join(layout.TmpDir, repository.DefaultWatchStatusFilename)
	record := repository.NewWatchStatusRecord(99, repoRoot, "dev", time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC))
	record = record.WithHeartbeat(repository.WatchHeartbeat{
		At:                time.Date(2026, 3, 15, 12, 0, 5, 0, time.UTC),
		LastRefreshDoneAt: time.Date(2026, 3, 15, 12, 0, 4, 0, time.UTC),
		LastError:         "stopped unexpectedly",
	})
	writeWatchStatusRecord(t, statusPath, record)

	service := NewWatchService()
	service.Now = func() time.Time { return time.Date(2026, 3, 15, 12, 0, 20, 0, time.UTC) }
	service.ProcessRunning = func(pid int) bool {
		t.Fatalf("ProcessRunning should not be checked when heartbeat is stale")
		return false
	}

	result, err := service.Status(context.Background(), repoRoot, 5*time.Second)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if result.Status != repository.WatchStatusKindStale {
		t.Fatalf("result.Status = %q, want stale", result.Status)
	}
	if result.Reason != "watch heartbeat is stale" {
		t.Fatalf("result.Reason = %q, want stale heartbeat", result.Reason)
	}

	if err := os.Remove(statusPath); err != nil {
		t.Fatalf("remove status file: %v", err)
	}
	result, err = service.Status(context.Background(), repoRoot, 5*time.Second)
	if err != nil {
		t.Fatalf("Status() after remove error = %v", err)
	}
	if result.Status != repository.WatchStatusKindAbsent {
		t.Fatalf("result.Status = %q, want absent", result.Status)
	}
	if result.Reason != "watch status file not found" {
		t.Fatalf("result.Reason = %q, want missing file reason", result.Reason)
	}
}

func loadWatchStatusRecord(t *testing.T, path string) repository.WatchStatusRecord {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %q: %v", path, err)
	}
	var record repository.WatchStatusRecord
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("decode %q: %v", path, err)
	}
	return record
}

func writeWatchStatusRecord(t *testing.T, path string, record repository.WatchStatusRecord) {
	t.Helper()

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		t.Fatalf("marshal watch status: %v", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}

func assertRepositoryFreshness(t *testing.T, db *sql.DB, reason repository.RefreshReason, status repository.RefreshRunStatus, freshness repository.FreshnessStatus, currentGeneration int64, lastGeneration int64) {
	t.Helper()

	var storedReason string
	var storedStatus string
	var storedFreshness string
	var storedCurrentGeneration int64
	var storedLastGeneration int64
	if err := db.QueryRow(`SELECT last_refresh_reason, last_refresh_status, freshness_status, current_refresh_generation, last_refresh_generation FROM repositories`).Scan(
		&storedReason,
		&storedStatus,
		&storedFreshness,
		&storedCurrentGeneration,
		&storedLastGeneration,
	); err != nil {
		t.Fatalf("query repositories freshness: %v", err)
	}
	if storedReason != string(reason) {
		t.Fatalf("last_refresh_reason = %q, want %q", storedReason, reason)
	}
	if storedStatus != string(status) {
		t.Fatalf("last_refresh_status = %q, want %q", storedStatus, status)
	}
	if storedFreshness != string(freshness) {
		t.Fatalf("freshness_status = %q, want %q", storedFreshness, freshness)
	}
	if storedCurrentGeneration != currentGeneration {
		t.Fatalf("current_refresh_generation = %d, want %d", storedCurrentGeneration, currentGeneration)
	}
	if storedLastGeneration != lastGeneration {
		t.Fatalf("last_refresh_generation = %d, want %d", storedLastGeneration, lastGeneration)
	}
}

func waitForFile(t *testing.T, path string) {
	t.Helper()
	waitForCondition(t, 2*time.Second, func() bool {
		_, err := os.Stat(path)
		return err == nil
	})
}

func waitForCondition(t *testing.T, timeout time.Duration, condition func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not satisfied before timeout")
}
