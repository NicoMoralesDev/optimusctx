package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/niccrow/optimusctx/internal/buildinfo"
	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
)

const (
	defaultWatchHeartbeatInterval = 2 * time.Second
	defaultWatchPollInterval      = 2 * time.Second
	defaultWatchDebounceWindow    = 750 * time.Millisecond
	defaultWatchStaleAfter        = 10 * time.Second
)

type WatchService struct {
	Locator           repository.Locator
	ResolveLayout     func(string) (state.Layout, error)
	Refresh           func(context.Context, RefreshRequest) (RefreshResult, error)
	Observe           func(context.Context, string) (<-chan repository.WatchEvent, error)
	Now               func() time.Time
	PID               func() int
	MkdirAll          func(string, fs.FileMode) error
	ReadFile          func(string) ([]byte, error)
	WriteFile         func(string, []byte, fs.FileMode) error
	Rename            func(string, string) error
	Remove            func(string) error
	ProcessRunning    func(int) bool
	ReportRefresh     func(repository.WatchRefreshReport)
	StatusFilename    string
	HeartbeatInterval time.Duration
	PollInterval      time.Duration
	DefaultStaleAfter time.Duration
	DefaultDebounce   time.Duration
}

func NewWatchService() WatchService {
	service := WatchService{
		Locator:       repository.NewLocator(),
		ResolveLayout: state.ResolveLayout,
		Refresh: func(ctx context.Context, request RefreshRequest) (RefreshResult, error) {
			return NewRefreshService().Refresh(ctx, request)
		},
		Now:               time.Now,
		PID:               os.Getpid,
		MkdirAll:          os.MkdirAll,
		ReadFile:          os.ReadFile,
		WriteFile:         os.WriteFile,
		Rename:            os.Rename,
		Remove:            os.Remove,
		StatusFilename:    repository.DefaultWatchStatusFilename,
		HeartbeatInterval: defaultWatchHeartbeatInterval,
		PollInterval:      defaultWatchPollInterval,
		DefaultStaleAfter: defaultWatchStaleAfter,
		DefaultDebounce:   defaultWatchDebounceWindow,
	}
	service.ProcessRunning = processRunning
	service.Observe = service.pollingObserver
	return service
}

func (s WatchService) Run(ctx context.Context, request repository.WatchRequest) (repository.WatchRunResult, error) {
	root, layout, debounceWindow, heartbeatInterval, statusPath, err := s.resolveRunDependencies(request)
	if err != nil {
		return repository.WatchRunResult{}, err
	}

	nowFn := s.now
	record := repository.NewWatchStatusRecord(s.pid(), root.RootPath, buildinfo.Version, nowFn().UTC())
	if err := s.writeStatus(statusPath, record); err != nil {
		return repository.WatchRunResult{}, fmt.Errorf("write watch status: %w", err)
	}
	defer func() {
		_ = s.Remove(statusPath)
	}()

	events, err := s.observe(ctx, root.RootPath)
	if err != nil {
		return repository.WatchRunResult{}, fmt.Errorf("start watch observer: %w", err)
	}

	heartbeatTicker := time.NewTicker(heartbeatInterval)
	defer heartbeatTicker.Stop()

	result := repository.WatchRunResult{
		RepositoryRoot: root.RootPath,
		StatePath:      layout.StateDir,
		StatusPath:     statusPath,
	}

	var refreshTimer *time.Timer
	var refreshPending bool
	pendingHints := make(map[string]struct{})
	pendingForceFull := false
	stopRefreshTimer := func() {
		if refreshTimer == nil {
			return
		}
		if !refreshTimer.Stop() {
			select {
			case <-refreshTimer.C:
			default:
			}
		}
		refreshTimer = nil
	}
	startRefreshTimer := func() {
		if debounceWindow <= 0 {
			debounceWindow = s.defaultDebounce()
		}
		if refreshTimer == nil {
			refreshTimer = time.NewTimer(debounceWindow)
			return
		}
		if !refreshTimer.Stop() {
			select {
			case <-refreshTimer.C:
			default:
			}
		}
		refreshTimer.Reset(debounceWindow)
	}
	recordHint := func(path string) {
		normalized, ok := normalizeWatchHint(root.RootPath, path)
		if !ok {
			pendingForceFull = true
			clear(pendingHints)
			return
		}
		if pendingForceFull {
			return
		}
		pendingHints[normalized] = struct{}{}
	}

	for {
		var refreshC <-chan time.Time
		if refreshTimer != nil {
			refreshC = refreshTimer.C
		}

		select {
		case <-ctx.Done():
			stopRefreshTimer()
			return result, nil
		case <-heartbeatTicker.C:
			record = record.WithHeartbeat(repository.WatchHeartbeat{At: nowFn().UTC()})
			if err := s.writeStatus(statusPath, record); err != nil {
				return repository.WatchRunResult{}, fmt.Errorf("update watch heartbeat: %w", err)
			}
		case event, ok := <-events:
			if !ok {
				stopRefreshTimer()
				return result, nil
			}
			eventAt := event.At
			if eventAt.IsZero() {
				eventAt = nowFn().UTC()
			}
			record = record.WithHeartbeat(repository.WatchHeartbeat{
				At:          nowFn().UTC(),
				LastEventAt: eventAt,
			})
			if event.Op == repository.WatchEventOpOverflow {
				record.LastError = "watch observer overflowed; falling back to full refresh"
				pendingForceFull = true
				clear(pendingHints)
			} else if event.Uncertain {
				record.LastError = "watch observer lost path certainty; falling back to full refresh"
				pendingForceFull = true
				clear(pendingHints)
			} else {
				recordHint(event.Path)
			}
			if err := s.writeStatus(statusPath, record); err != nil {
				return repository.WatchRunResult{}, fmt.Errorf("update watch event status: %w", err)
			}
			refreshPending = true
			startRefreshTimer()
		case <-refreshC:
			stopRefreshTimer()
			if !refreshPending {
				continue
			}
			refreshPending = false
			refreshHints := sortedWatchHints(pendingHints)
			refreshForceFull := pendingForceFull
			clear(pendingHints)
			pendingForceFull = false
			startedAt := nowFn().UTC()
			record = record.WithHeartbeat(repository.WatchHeartbeat{
				At:                   startedAt,
				LastRefreshStartedAt: startedAt,
			})
			if err := s.writeStatus(statusPath, record); err != nil {
				return repository.WatchRunResult{}, fmt.Errorf("persist watch refresh start: %w", err)
			}

			refreshResult, refreshErr := s.refresh(ctx, root.RootPath, refreshForceFull, refreshHints)
			doneAt := nowFn().UTC()
			update := repository.WatchHeartbeat{
				At:                doneAt,
				LastRefreshDoneAt: doneAt,
			}
			if refreshErr == nil {
				update.LastRefreshGeneration = refreshResult.Generation
				update.LastError = ""
			} else {
				update.LastError = refreshErr.Error()
			}
			record = record.WithHeartbeat(update)
			if err := s.writeStatus(statusPath, record); err != nil {
				return repository.WatchRunResult{}, fmt.Errorf("persist watch refresh result: %w", err)
			}
			s.reportRefresh(repository.WatchRefreshReport{
				Reason:              repository.RefreshReasonWatch,
				Generation:          refreshResult.Generation,
				FreshnessStatus:     refreshResult.FreshnessStatus,
				ChangedFiles:        refreshResult.ChangedFiles,
				UnchangedFiles:      refreshResult.UnchangedFiles,
				AffectedDirectories: refreshResult.AffectedDirectories,
				ForceFull:           refreshForceFull,
				ChangedHint:         append([]string(nil), refreshHints...),
				Error:               update.LastError,
			})
			if refreshErr != nil && errors.Is(refreshErr, context.Canceled) {
				return result, nil
			}
		}
	}
}

func (s WatchService) Status(ctx context.Context, startPath string, staleAfter time.Duration) (repository.WatchStatusResult, error) {
	_ = ctx
	root, err := s.Locator.Resolve(startPath)
	if err != nil {
		return repository.WatchStatusResult{}, fmt.Errorf("resolve repository root: %w", err)
	}

	layout, err := s.resolveLayout(root.RootPath)
	if err != nil {
		return repository.WatchStatusResult{}, fmt.Errorf("resolve state layout: %w", err)
	}

	statusPath := s.statusPath(layout)
	result := repository.WatchStatusResult{
		RepositoryRoot: root.RootPath,
		StatePath:      layout.StateDir,
		StatusPath:     statusPath,
		Status:         repository.WatchStatusKindAbsent,
		Reason:         "watch status file not found",
	}

	data, err := s.ReadFile(statusPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return result, nil
		}
		return repository.WatchStatusResult{}, fmt.Errorf("read watch status: %w", err)
	}

	var record repository.WatchStatusRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return repository.WatchStatusResult{}, fmt.Errorf("decode watch status: %w", err)
	}
	result.Record = record

	if staleAfter <= 0 {
		staleAfter = s.defaultStaleAfter()
	}
	lastHeartbeat := record.LastHeartbeatAtTime()
	if lastHeartbeat.IsZero() {
		result.Status = repository.WatchStatusKindStale
		result.Reason = "watch heartbeat missing"
		return result, nil
	}

	if s.now().UTC().Sub(lastHeartbeat) > staleAfter {
		result.Status = repository.WatchStatusKindStale
		result.Reason = "watch heartbeat is stale"
		return result, nil
	}

	if record.PID <= 0 || !s.processRunning(record.PID) {
		result.Status = repository.WatchStatusKindStale
		result.Reason = "watch process is not running"
		return result, nil
	}

	result.Status = repository.WatchStatusKindRunning
	result.Reason = "watch process heartbeat is current"
	return result, nil
}

func (s WatchService) resolveRunDependencies(request repository.WatchRequest) (repository.RepositoryRoot, state.Layout, time.Duration, time.Duration, string, error) {
	root, err := s.Locator.Resolve(request.StartPath)
	if err != nil {
		return repository.RepositoryRoot{}, state.Layout{}, 0, 0, "", fmt.Errorf("resolve repository root: %w", err)
	}

	layout, err := s.resolveLayout(root.RootPath)
	if err != nil {
		return repository.RepositoryRoot{}, state.Layout{}, 0, 0, "", fmt.Errorf("resolve state layout: %w", err)
	}
	if err := s.MkdirAll(layout.TmpDir, 0o755); err != nil {
		return repository.RepositoryRoot{}, state.Layout{}, 0, 0, "", fmt.Errorf("create tmp directory: %w", err)
	}

	debounceWindow := request.DebounceWindow
	if debounceWindow <= 0 {
		debounceWindow = s.defaultDebounce()
	}

	heartbeatInterval := request.HeartbeatInterval
	if heartbeatInterval <= 0 {
		heartbeatInterval = s.defaultHeartbeat()
	}

	return root, layout, debounceWindow, heartbeatInterval, s.statusPath(layout), nil
}

func (s WatchService) pollingObserver(ctx context.Context, root string) (<-chan repository.WatchEvent, error) {
	events := make(chan repository.WatchEvent, 1)
	interval := s.PollInterval
	if interval <= 0 {
		interval = defaultWatchPollInterval
	}

	previous, err := snapshotWatchTree(root)
	if err != nil {
		return nil, fmt.Errorf("snapshot repository: %w", err)
	}

	go func() {
		defer close(events)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				current, err := snapshotWatchTree(root)
				if err != nil {
					select {
					case events <- repository.WatchEvent{Path: ".", Op: repository.WatchEventOpOverflow, At: s.now().UTC()}:
					case <-ctx.Done():
					}
					continue
				}
				changed := diffWatchSnapshots(previous, current)
				if len(changed) > 0 {
					previous = current
					for _, path := range changed {
						select {
						case events <- repository.WatchEvent{Path: path, Op: repository.WatchEventOpChange, At: s.now().UTC()}:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}
	}()

	return events, nil
}

func snapshotWatchTree(root string) (map[string]string, error) {
	snapshot := make(map[string]string)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if rel == ".git" || rel == state.DirectoryName {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(rel, ".git/") || strings.HasPrefix(rel, state.DirectoryName+"/") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		mode := "f"
		if entry.IsDir() {
			mode = "d"
		}
		snapshot[rel] = fmt.Sprintf("%s:%d:%d", mode, info.Size(), info.ModTime().UnixNano())
		return nil
	})
	if err != nil {
		return nil, err
	}
	return snapshot, nil
}

func diffWatchSnapshots(previous map[string]string, current map[string]string) []string {
	seen := make(map[string]struct{}, len(previous)+len(current))
	changed := make([]string, 0)
	for path, leftValue := range previous {
		seen[path] = struct{}{}
		if current[path] != leftValue {
			changed = append(changed, path)
		}
	}
	for path, rightValue := range current {
		if _, ok := seen[path]; ok {
			continue
		}
		if previous[path] != rightValue {
			changed = append(changed, path)
		}
	}
	sort.Strings(changed)
	return changed
}

func (s WatchService) resolveLayout(root string) (state.Layout, error) {
	resolver := s.ResolveLayout
	if resolver == nil {
		resolver = state.ResolveLayout
	}
	return resolver(root)
}

func (s WatchService) statusPath(layout state.Layout) string {
	name := s.StatusFilename
	if strings.TrimSpace(name) == "" {
		name = repository.DefaultWatchStatusFilename
	}
	return filepath.Join(layout.TmpDir, name)
}

func (s WatchService) writeStatus(path string, record repository.WatchStatusRecord) error {
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("encode watch status: %w", err)
	}
	data = append(data, '\n')

	tempPath := fmt.Sprintf("%s.%d.tmp", path, s.now().UnixNano())
	if err := s.WriteFile(tempPath, data, 0o644); err != nil {
		return err
	}
	return s.Rename(tempPath, path)
}

func (s WatchService) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func (s WatchService) pid() int {
	if s.PID != nil {
		return s.PID()
	}
	return os.Getpid()
}

func (s WatchService) observe(ctx context.Context, root string) (<-chan repository.WatchEvent, error) {
	if s.Observe != nil {
		return s.Observe(ctx, root)
	}
	return s.pollingObserver(ctx, root)
}

func normalizeWatchHint(root string, hint string) (string, bool) {
	hint = filepath.ToSlash(strings.TrimSpace(hint))
	if hint == "" || hint == "." {
		return "", false
	}
	if filepath.IsAbs(hint) {
		rel, err := filepath.Rel(root, hint)
		if err != nil {
			return "", false
		}
		hint = filepath.ToSlash(rel)
	}
	cleaned := pathClean(hint)
	if cleaned == "." || strings.HasPrefix(cleaned, "../") || cleaned == ".." {
		return "", false
	}
	return cleaned, true
}

func pathClean(path string) string {
	return strings.TrimPrefix(filepath.ToSlash(filepath.Clean(path)), "./")
}

func sortedWatchHints(hints map[string]struct{}) []string {
	if len(hints) == 0 {
		return nil
	}
	paths := make([]string, 0, len(hints))
	for path := range hints {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func (s WatchService) refresh(ctx context.Context, root string, forceFull bool, changedHint []string) (RefreshResult, error) {
	if s.Refresh != nil {
		return s.Refresh(ctx, RefreshRequest{
			StartPath:   root,
			Reason:      repository.RefreshReasonWatch,
			ForceFull:   forceFull,
			ChangedHint: changedHint,
		})
	}
	return NewRefreshService().Refresh(ctx, RefreshRequest{
		StartPath:   root,
		Reason:      repository.RefreshReasonWatch,
		ForceFull:   forceFull,
		ChangedHint: changedHint,
	})
}

func (s WatchService) processRunning(pid int) bool {
	if s.ProcessRunning != nil {
		return s.ProcessRunning(pid)
	}
	return processRunning(pid)
}

func (s WatchService) reportRefresh(report repository.WatchRefreshReport) {
	if s.ReportRefresh != nil {
		s.ReportRefresh(report)
	}
}

func (s WatchService) defaultHeartbeat() time.Duration {
	if s.HeartbeatInterval > 0 {
		return s.HeartbeatInterval
	}
	return defaultWatchHeartbeatInterval
}

func (s WatchService) defaultStaleAfter() time.Duration {
	if s.DefaultStaleAfter > 0 {
		return s.DefaultStaleAfter
	}
	return defaultWatchStaleAfter
}

func (s WatchService) defaultDebounce() time.Duration {
	if s.DefaultDebounce > 0 {
		return s.DefaultDebounce
	}
	return defaultWatchDebounceWindow
}
