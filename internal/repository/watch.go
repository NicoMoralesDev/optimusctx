package repository

import "time"

const DefaultWatchStatusFilename = "watch-status.json"

type WatchEventOp string

const (
	WatchEventOpChange   WatchEventOp = "change"
	WatchEventOpOverflow WatchEventOp = "overflow"
)

type WatchStatusKind string

const (
	WatchStatusKindAbsent  WatchStatusKind = "absent"
	WatchStatusKindRunning WatchStatusKind = "running"
	WatchStatusKindStale   WatchStatusKind = "stale"
)

type WatchRequest struct {
	StartPath         string
	HeartbeatInterval time.Duration
	DebounceWindow    time.Duration
	StaleAfter        time.Duration
}

type WatchEvent struct {
	Path      string
	Op        WatchEventOp
	At        time.Time
	Uncertain bool
}

type WatchHeartbeat struct {
	At                    time.Time
	LastEventAt           time.Time
	LastRefreshStartedAt  time.Time
	LastRefreshDoneAt     time.Time
	LastRefreshGeneration int64
	LastError             string
}

type WatchStatusRecord struct {
	PID                   int    `json:"pid"`
	RepoRoot              string `json:"repo_root"`
	BinaryVersion         string `json:"binary_version"`
	StartedAt             string `json:"started_at"`
	LastHeartbeatAt       string `json:"last_heartbeat_at"`
	LastEventAt           string `json:"last_event_at,omitempty"`
	LastRefreshStartedAt  string `json:"last_refresh_started_at,omitempty"`
	LastRefreshDoneAt     string `json:"last_refresh_completed_at,omitempty"`
	LastRefreshGeneration int64  `json:"last_refresh_generation,omitempty"`
	LastError             string `json:"last_error,omitempty"`
}

type WatchRunResult struct {
	RepositoryRoot string
	StatePath      string
	StatusPath     string
}

type WatchRefreshReport struct {
	Reason              RefreshReason
	Generation          int64
	FreshnessStatus     FreshnessStatus
	ChangedFiles        int
	UnchangedFiles      int
	AffectedDirectories int
	ForceFull           bool
	ChangedHint         []string
	Error               string
}

type WatchStatusResult struct {
	RepositoryRoot string
	StatePath      string
	StatusPath     string
	Status         WatchStatusKind
	Reason         string
	Record         WatchStatusRecord
}

func (r WatchStatusRecord) StartedAtTime() time.Time {
	return parseWatchTime(r.StartedAt)
}

func (r WatchStatusRecord) LastHeartbeatAtTime() time.Time {
	return parseWatchTime(r.LastHeartbeatAt)
}

func (r WatchStatusRecord) LastEventAtTime() time.Time {
	return parseWatchTime(r.LastEventAt)
}

func (r WatchStatusRecord) LastRefreshStartedAtTime() time.Time {
	return parseWatchTime(r.LastRefreshStartedAt)
}

func (r WatchStatusRecord) LastRefreshDoneAtTime() time.Time {
	return parseWatchTime(r.LastRefreshDoneAt)
}

func NewWatchStatusRecord(pid int, repoRoot string, version string, now time.Time) WatchStatusRecord {
	now = now.UTC()
	ts := formatWatchTime(now)
	return WatchStatusRecord{
		PID:             pid,
		RepoRoot:        repoRoot,
		BinaryVersion:   version,
		StartedAt:       ts,
		LastHeartbeatAt: ts,
	}
}

func (r WatchStatusRecord) WithHeartbeat(update WatchHeartbeat) WatchStatusRecord {
	if !update.At.IsZero() {
		r.LastHeartbeatAt = formatWatchTime(update.At.UTC())
	}
	if !update.LastEventAt.IsZero() {
		r.LastEventAt = formatWatchTime(update.LastEventAt.UTC())
	}
	if !update.LastRefreshStartedAt.IsZero() {
		r.LastRefreshStartedAt = formatWatchTime(update.LastRefreshStartedAt.UTC())
	}
	if !update.LastRefreshDoneAt.IsZero() {
		r.LastRefreshDoneAt = formatWatchTime(update.LastRefreshDoneAt.UTC())
	}
	if update.LastRefreshGeneration > 0 {
		r.LastRefreshGeneration = update.LastRefreshGeneration
	}
	r.LastError = update.LastError
	return r
}

func formatWatchTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func parseWatchTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}
