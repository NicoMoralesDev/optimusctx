package repository

import "time"

type DoctorRequest struct {
	BudgetLimit     int
	WatchStaleAfter time.Duration
}

type DoctorStatus string

const (
	DoctorStatusHealthy  DoctorStatus = "healthy"
	DoctorStatusDegraded DoctorStatus = "degraded"
	DoctorStatusMissing  DoctorStatus = "missing"
)

type DoctorReport struct {
	Repository     LayeredContextEnvelope
	Identity       LayeredContextRepositoryIdentity
	Request        DoctorRequest
	Summary        DoctorSummary
	Install        DoctorInstallSection
	State          DoctorStateSection
	Refresh        DoctorRefreshSection
	Watch          DoctorWatchSection
	Structural     DoctorStructuralSection
	Budget         DoctorBudgetSection
	MCPReadiness   DoctorMCPReadinessSection
	MCPActivity    DoctorMCPActivitySection
	HostMCP        DoctorHostRegistrationSection
	RecommendedFix []string
}

type DoctorSummary struct {
	Status             DoctorStatus
	RepositoryDetected bool
	Initialized        bool
	Issues             []DoctorIssue
}

type DoctorIssue struct {
	Section string
	Status  DoctorStatus
	Summary string
	Action  string
}

type DoctorInstallSection struct {
	Status        DoctorStatus
	BinaryVersion string
	WorkingDir    string
}

type DoctorStateSection struct {
	Status          DoctorStatus
	Layout          HealthStateLayout
	Metadata        HealthStateMetadata
	RepositoryMatch bool
}

type DoctorRefreshSection struct {
	Status  DoctorStatus
	Health  HealthRefreshDiagnostics
	LastRun DoctorRefreshRun
}

type DoctorRefreshRun struct {
	Present       bool
	Generation    int64
	Reason        RefreshReason
	Status        RefreshRunStatus
	FailureReason string
	StartedAt     time.Time
	CompletedAt   time.Time
}

type DoctorWatchSection struct {
	Status   DoctorStatus
	Optional bool
	Summary  string
	Health   WatchStatusResult
}

type DoctorStructuralSection struct {
	Status   DoctorStatus
	Summary  RepositoryStructuralCoverageSummary
	Examples []DoctorStructuralCoverageExample
}

type DoctorStructuralCoverageExample struct {
	Path           string
	Language       string
	CoverageState  ExtractionCoverageState
	CoverageReason ExtractionCoverageReason
	SymbolCount    int64
}

type DoctorBudgetSection struct {
	Status   DoctorStatus
	Summary  BudgetAnalysisSummary
	Policy   BudgetEstimatePolicy
	Hotspots []BudgetHotspot
}

type DoctorMCPReadinessSection struct {
	Status              DoctorStatus
	ServerName          string
	ServeCommand        ServeCommand
	SnippetDocument     ClientConfigDocument
	SnippetPreview      string
	SnippetAvailable    bool
	SnippetParseFailure string
}
