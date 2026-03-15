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

type ExtractionCoverageState string

const (
	ExtractionCoverageStateSupported   ExtractionCoverageState = "supported"
	ExtractionCoverageStatePartial     ExtractionCoverageState = "partial"
	ExtractionCoverageStateUnsupported ExtractionCoverageState = "unsupported"
	ExtractionCoverageStateFailed      ExtractionCoverageState = "failed"
	ExtractionCoverageStateSkipped     ExtractionCoverageState = "skipped"
)

type ExtractionCoverageReason string

const (
	ExtractionCoverageReasonNone                ExtractionCoverageReason = ""
	ExtractionCoverageReasonUnsupportedLanguage ExtractionCoverageReason = "unsupported_language"
	ExtractionCoverageReasonParseError          ExtractionCoverageReason = "parse_error"
	ExtractionCoverageReasonAdapterError        ExtractionCoverageReason = "adapter_error"
	ExtractionCoverageReasonQueryError          ExtractionCoverageReason = "query_error"
)

type ExtractionCandidate struct {
	RepositoryID     int64
	FileID           int64
	Path             string
	Language         string
	ContentHash      string
	SourceGeneration int64
	RefreshRunID     int64
}

type FileExtractionRecord struct {
	ID                  int64
	RepositoryID        int64
	FileID              int64
	Path                string
	Language            string
	AdapterName         string
	GrammarVersion      string
	SourceContentHash   string
	SourceGeneration    int64
	CoverageState       ExtractionCoverageState
	CoverageReason      ExtractionCoverageReason
	ParserErrorCount    int64
	HasErrorNodes       bool
	SymbolCount         int64
	TopLevelSymbolCount int64
	MaxSymbolDepth      int64
	ExtractedAt         time.Time
	RefreshRunID        int64
}

type SymbolRecord struct {
	ID                 int64
	RepositoryID       int64
	FileID             int64
	FileExtractionID   int64
	StableKey          string
	ParentSymbolID     int64
	ParentStableKey    string
	Path               string
	Language           string
	Kind               string
	Name               string
	QualifiedName      string
	Ordinal            int64
	Depth              int64
	StartByte          int64
	EndByte            int64
	StartRow           int64
	StartColumn        int64
	EndRow             int64
	EndColumn          int64
	NameStartByte      int64
	NameEndByte        int64
	SignatureStartByte int64
	SignatureEndByte   int64
	IsExported         bool
}

type FileStructuralArtifacts struct {
	Extraction FileExtractionRecord
	Symbols    []SymbolRecord
}

type RepositoryStructuralCoverageSummary struct {
	RepositoryID         int64
	IncludedFileCount    int64
	ExtractionCount      int64
	SupportedCount       int64
	PartialCount         int64
	UnsupportedCount     int64
	FailedCount          int64
	SkippedCount         int64
	FilesWithCoverageGap int64
	TotalSymbolCount     int64
}

type RepositoryMapFileRecord struct {
	FileID              int64
	Path                string
	DirectoryPath       string
	Language            string
	IgnoreStatus        IgnoreStatus
	CoverageState       ExtractionCoverageState
	CoverageReason      ExtractionCoverageReason
	SymbolCount         int64
	TopLevelSymbolCount int64
	MaxSymbolDepth      int64
	SourceGeneration    int64
	Symbols             []SymbolRecord
}

type RepositoryMapDirectoryRecord struct {
	Path                   string
	ParentPath             string
	IncludedFileCount      int64
	IncludedDirectoryCount int64
	TotalSizeBytes         int64
	LastRefreshGeneration  int64
}

type RepositoryMap struct {
	RepositoryRoot string
	Generation     int64
	Freshness      FreshnessStatus
	Directories    []RepositoryMapDirectory
}

type RepositoryMapDirectory struct {
	Path                   string
	ParentPath             string
	IncludedFileCount      int64
	IncludedDirectoryCount int64
	TotalSizeBytes         int64
	LastRefreshGeneration  int64
	Files                  []RepositoryMapFile
}

type RepositoryMapFile struct {
	Path                string
	DirectoryPath       string
	Language            string
	CoverageState       ExtractionCoverageState
	CoverageReason      ExtractionCoverageReason
	HasCoverageGap      bool
	SymbolCount         int64
	TopLevelSymbolCount int64
	MaxSymbolDepth      int64
	SourceGeneration    int64
	Symbols             []RepositoryMapSymbol
}

type RepositoryMapSymbol struct {
	Kind          string
	Name          string
	QualifiedName string
	Ordinal       int64
}

type LayeredContextEnvelope struct {
	RepositoryRoot string
	Generation     int64
	Freshness      FreshnessStatus
}

type LayeredContextL0 struct {
	Repository LayeredContextEnvelope
	Identity   LayeredContextRepositoryIdentity
	Languages  []LayeredContextLanguageSummary
	MajorAreas []LayeredContextMajorAreaSummary
}

type LayeredContextL1 struct {
	Repository  LayeredContextEnvelope
	Identity    LayeredContextRepositoryIdentity
	Limits      LayeredContextL1LimitMetadata
	Candidates  []LayeredContextL1CandidateFile
	Directories []LayeredContextL1DirectorySummary
}

type LayeredContextRepositoryIdentity struct {
	RootPath      string
	DetectionMode string
	GitHeadRef    string
	GitHeadCommit string
}

type LayeredContextLanguageSummary struct {
	Language       string
	FileCount      int64
	TotalSizeBytes int64
}

type MajorAreaKind string

const (
	MajorAreaKindDirectory MajorAreaKind = "directory"
	MajorAreaKindRootFiles MajorAreaKind = "root_files"
)

type LayeredContextMajorAreaSummary struct {
	Path              string
	Kind              MajorAreaKind
	IncludedFileCount int64
	TotalSizeBytes    int64
}

type LayeredContextL1LimitMetadata struct {
	FileLimit           int
	ReturnedFileCount   int
	TotalCandidateCount int64
	FileTruncated       bool
	PerFileSymbolLimit  int
}

type LayeredContextL1CandidateFile struct {
	Path                string
	DirectoryPath       string
	Language            string
	CoverageState       ExtractionCoverageState
	CoverageReason      ExtractionCoverageReason
	HasCoverageGap      bool
	SymbolCount         int64
	TopLevelSymbolCount int64
	MaxSymbolDepth      int64
	SourceGeneration    int64
	Summary             string
	DirectorySummary    LayeredContextL1DirectorySummary
	SymbolWindow        LayeredContextL1SymbolWindow
	Symbols             []LayeredContextL1Symbol
}

type LayeredContextL1DirectorySummary struct {
	Path                   string
	IncludedFileCount      int64
	IncludedDirectoryCount int64
	TotalSizeBytes         int64
	LastRefreshGeneration  int64
}

type LayeredContextL1SymbolWindow struct {
	ReturnedCount int
	TotalCount    int64
	Truncated     bool
}

type LayeredContextL1Symbol struct {
	Kind          string
	Name          string
	QualifiedName string
	Ordinal       int64
}

type SymbolLookupRequest struct {
	Name       string
	PathPrefix string
	Language   string
	Kind       string
	Limit      int
}

type SymbolLookupResult struct {
	Repository LayeredContextEnvelope
	Identity   LayeredContextRepositoryIdentity
	Request    SymbolLookupRequest
	Limit      int
	Matches    []SymbolLookupMatch
}

type SymbolLookupMatch struct {
	StableKey     string
	Path          string
	Language      string
	Kind          string
	Name          string
	QualifiedName string
	Ordinal       int64
	StartRow      int64
	StartColumn   int64
	EndRow        int64
	EndColumn     int64
}
