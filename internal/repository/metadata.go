package repository

import "time"

type DiscoveryResult struct {
	Repository  RepositoryRecord
	Directories []DirectoryRecord
	Files       []FileRecord
}

type FreshnessStatus string

const (
	FreshnessStatusFresh             FreshnessStatus = "fresh"
	FreshnessStatusStale             FreshnessStatus = "stale"
	FreshnessStatusPartiallyDegraded FreshnessStatus = "partially_degraded"
)

type RefreshRunStatus string

const (
	RefreshRunStatusPending RefreshRunStatus = "pending"
	RefreshRunStatusRunning RefreshRunStatus = "running"
	RefreshRunStatusSuccess RefreshRunStatus = "success"
	RefreshRunStatusFailed  RefreshRunStatus = "failed"
)

type RefreshReason string

const (
	RefreshReasonInit   RefreshReason = "init"
	RefreshReasonManual RefreshReason = "manual"
	RefreshReasonWatch  RefreshReason = "watch"
	RefreshReasonRepair RefreshReason = "repair"
)

type RepositoryRecord struct {
	RootPath      string
	DetectionMode string
	GitCommonDir  string
	GitHeadRef    string
	GitHeadCommit string
}

type RepositoryFreshness struct {
	RepositoryID           int64
	RootPath               string
	DetectionMode          string
	GitCommonDir           string
	GitHeadRef             string
	GitHeadCommit          string
	LastRefreshStartedAt   time.Time
	LastRefreshCompletedAt time.Time
	LastRefreshReason      RefreshReason
	LastRefreshStatus      RefreshRunStatus
	FreshnessStatus        FreshnessStatus
	FreshnessReason        string
	CurrentGeneration      int64
	LastRefreshGeneration  int64
}

type RepositorySnapshot struct {
	Repository  RepositoryFreshness
	Directories []DirectorySnapshotRecord
	Files       []PersistedFileSnapshotRecord
}

type DirectoryRecord struct {
	Path         string
	ParentPath   string
	IgnoreStatus IgnoreStatus
	IgnoreReason IgnoreReason
	DiscoveredAt time.Time
}

type DirectorySnapshotRecord struct {
	Path                   string
	ParentPath             string
	IgnoreStatus           IgnoreStatus
	IgnoreReason           IgnoreReason
	DiscoveredAt           time.Time
	SubtreeFingerprint     string
	IncludedFileCount      int64
	IncludedDirectoryCount int64
	TotalSizeBytes         int64
	LastRefreshedAt        time.Time
	LastRefreshGeneration  int64
}

type FileRecord struct {
	Path              string
	DirectoryPath     string
	Extension         string
	LanguageHint      string
	SizeBytes         int64
	ContentHash       string
	LastIndexedAt     time.Time
	FilesystemModTime time.Time
	IgnoreStatus      IgnoreStatus
	IgnoreReason      IgnoreReason
	DiscoveredAt      time.Time
}

type PersistedFileSnapshotRecord struct {
	Path               string
	DirectoryPath      string
	Extension          string
	LanguageHint       string
	SizeBytes          int64
	ContentHash        string
	LastIndexedAt      time.Time
	FilesystemModTime  time.Time
	IgnoreStatus       IgnoreStatus
	IgnoreReason       IgnoreReason
	DiscoveredAt       time.Time
	UpdatedAt          time.Time
	LastSeenGeneration int64
	RefreshRunID       int64
	UpdatedReason      string
}

type RefreshRunRecord struct {
	ID            int64
	RepositoryID  int64
	Generation    int64
	Reason        RefreshReason
	Status        RefreshRunStatus
	FailureReason string
	StartedAt     time.Time
	CompletedAt   time.Time
	MetadataJSON  string
}
