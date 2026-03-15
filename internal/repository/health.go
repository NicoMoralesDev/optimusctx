package repository

import "time"

type HealthRequest struct{}

type HealthStateStatus string

const (
	HealthStateStatusMissing HealthStateStatus = "missing"
	HealthStateStatusPartial HealthStateStatus = "partial"
	HealthStateStatusReady   HealthStateStatus = "ready"
)

type HealthPathStatus struct {
	Path   string
	Exists bool
}

type HealthStateLayout struct {
	StateDir     HealthPathStatus
	MetadataFile HealthPathStatus
	DatabaseFile HealthPathStatus
	LogsDir      HealthPathStatus
	TmpDir       HealthPathStatus
}

type HealthStateMetadata struct {
	Present           bool
	FormatVersion     int
	RepoRoot          string
	RepoDetectionMode string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	RuntimeVersion    string
	SchemaVersion     int
}

type HealthRefreshDiagnostics struct {
	Present                bool
	RepositoryID           int64
	DetectionMode          string
	GitHeadRef             string
	GitHeadCommit          string
	CurrentGeneration      int64
	LastRefreshGeneration  int64
	LastRefreshReason      RefreshReason
	LastRefreshStatus      RefreshRunStatus
	LastRefreshStartedAt   time.Time
	LastRefreshCompletedAt time.Time
	Freshness              FreshnessStatus
	FreshnessReason        string
}

type HealthSummary struct {
	Initialized          bool
	StateStatus          HealthStateStatus
	RepositoryRegistered bool
}

type HealthResult struct {
	Repository LayeredContextEnvelope
	Identity   LayeredContextRepositoryIdentity
	Request    HealthRequest
	Summary    HealthSummary
	State      HealthStateLayout
	Metadata   HealthStateMetadata
	Refresh    HealthRefreshDiagnostics
}
