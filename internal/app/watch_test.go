package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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
