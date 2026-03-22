package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/niccrow/optimusctx/internal/buildinfo"
	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"

	_ "modernc.org/sqlite"
)

const defaultDoctorBudgetLimit = 5

type DoctorService struct {
	HealthService   HealthService
	WatchService    WatchService
	BudgetService   BudgetAnalysisService
	MCPActivity     MCPActivityStore
	ResolveLayout   func(string) (state.Layout, error)
	OpenDB          func(string) (*sql.DB, error)
	Getwd           func() (string, error)
	SnippetRenderer func() string
	ReadFile        func(string) ([]byte, error)
	RunCommand      func(context.Context, string, ...string) ([]byte, error)
}

func NewDoctorService() DoctorService {
	health := NewHealthService()
	return DoctorService{
		HealthService: health,
		WatchService:  NewWatchService(),
		BudgetService: NewBudgetAnalysisService(),
		MCPActivity:   NewMCPActivityStore(),
		ResolveLayout: state.ResolveLayout,
		OpenDB:        health.OpenDB,
		Getwd:         os.Getwd,
		ReadFile:      os.ReadFile,
		RunCommand: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return exec.CommandContext(ctx, name, args...).CombinedOutput()
		},
		SnippetRenderer: func() string {
			return NewSnippetGenerator().Render()
		},
	}
}

func (s DoctorService) Doctor(ctx context.Context, startPath string, request repository.DoctorRequest) (repository.DoctorReport, error) {
	workingDir, err := s.getwd()
	if err != nil {
		return repository.DoctorReport{}, fmt.Errorf("resolve working directory: %w", err)
	}

	report := repository.DoctorReport{
		Request: request,
		Install: repository.DoctorInstallSection{
			Status:        repository.DoctorStatusHealthy,
			BinaryVersion: buildinfo.Version,
			WorkingDir:    workingDir,
		},
		MCPReadiness: buildDoctorMCPReadiness(s.SnippetRenderer),
		Structural: repository.DoctorStructuralSection{
			Status: repository.DoctorStatusMissing,
		},
		Budget: repository.DoctorBudgetSection{
			Status: repository.DoctorStatusMissing,
		},
	}

	health, err := s.HealthService.Health(ctx, startPath, repository.HealthRequest{})
	if err != nil {
		return repository.DoctorReport{}, err
	}

	report.Repository = health.Repository
	report.Identity = health.Identity
	report.State = repository.DoctorStateSection{
		Layout:          health.State,
		Metadata:        health.Metadata,
		RepositoryMatch: !health.Metadata.Present || health.Metadata.RepoRoot == "" || health.Metadata.RepoRoot == health.Identity.RootPath,
	}
	report.Refresh = repository.DoctorRefreshSection{
		Health:  health.Refresh,
		Status:  doctorRefreshStatus(health),
		LastRun: repository.DoctorRefreshRun{},
	}
	report.State.Status = doctorStateStatus(health, report.State.RepositoryMatch)

	watch, err := s.WatchService.Status(ctx, startPath, request.WatchStaleAfter)
	if err != nil {
		return repository.DoctorReport{}, fmt.Errorf("read watch status: %w", err)
	}
	report.Watch = repository.DoctorWatchSection{
		Status:   doctorWatchStatus(watch),
		Optional: watch.Status == repository.WatchStatusKindAbsent,
		Summary:  doctorWatchSummary(watch),
		Health:   watch,
	}
	report.MCPActivity, err = s.MCPActivity.Load(health.Identity.RootPath)
	if err != nil {
		return repository.DoctorReport{}, fmt.Errorf("read mcp activity: %w", err)
	}
	report.HostMCP, err = s.hostRegistrations(ctx, health.Identity.RootPath)
	if err != nil {
		return repository.DoctorReport{}, fmt.Errorf("read host mcp registrations: %w", err)
	}

	if health.State.DatabaseFile.Exists {
		layout, err := s.resolveLayout(health.Identity.RootPath)
		if err != nil {
			return repository.DoctorReport{}, fmt.Errorf("resolve state layout: %w", err)
		}
		db, err := s.openDB(layout.DatabasePath)
		if err != nil {
			return repository.DoctorReport{}, fmt.Errorf("open state database: %w", err)
		}
		defer db.Close()

		if err := db.PingContext(ctx); err != nil {
			return repository.DoctorReport{}, fmt.Errorf("ping state database: %w", err)
		}

		if health.Refresh.Present {
			lastRun, err := readDoctorRefreshRun(ctx, db, health.Refresh.RepositoryID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return repository.DoctorReport{}, fmt.Errorf("read latest refresh run: %w", err)
			}
			if err == nil {
				report.Refresh.LastRun = lastRun
				if lastRun.Status == repository.RefreshRunStatusFailed {
					report.Refresh.Status = repository.DoctorStatusDegraded
				}
			}

			structural, err := readDoctorStructuralCoverage(ctx, db, health.Refresh.RepositoryID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return repository.DoctorReport{}, fmt.Errorf("read structural coverage: %w", err)
			}
			if err == nil {
				report.Structural = structural
			}
		}
	}

	if health.Summary.Initialized {
		budgetLimit := request.BudgetLimit
		if budgetLimit <= 0 {
			budgetLimit = defaultDoctorBudgetLimit
		}
		budget, err := s.BudgetService.Analyze(ctx, startPath, repository.BudgetAnalysisRequest{
			GroupBy: repository.BudgetGroupByFile,
			Limit:   budgetLimit,
		})
		if err == nil {
			report.Budget = repository.DoctorBudgetSection{
				Status:   doctorBudgetStatus(budget),
				Summary:  budget.Summary,
				Policy:   budget.Policy,
				Hotspots: append([]repository.BudgetHotspot(nil), budget.Hotspots...),
			}
		}
	}

	report.Summary = doctorSummary(report)
	report.RecommendedFix = doctorRecommendedFixes(report.Summary.Issues)
	return report, nil
}

func buildDoctorMCPReadiness(renderer func() string) repository.DoctorMCPReadinessSection {
	section := repository.DoctorMCPReadinessSection{
		Status:           repository.DoctorStatusHealthy,
		ServerName:       repository.DefaultMCPServerName,
		ServeCommand:     repository.NewServeCommand(""),
		SnippetAvailable: true,
	}
	if renderer == nil {
		return section
	}
	preview := renderer()
	section.SnippetPreview = preview
	document, err := repository.MergeClientConfig(nil, repository.DefaultMCPServerName, repository.NewServeCommand(""))
	if err != nil {
		section.Status = repository.DoctorStatusDegraded
		section.SnippetAvailable = false
		section.SnippetParseFailure = err.Error()
		return section
	}
	section.SnippetDocument = document
	return section
}

func readDoctorRefreshRun(ctx context.Context, db *sql.DB, repositoryID int64) (repository.DoctorRefreshRun, error) {
	var run repository.DoctorRefreshRun
	var reason, status string
	var failureReason sql.NullString
	var startedAt string
	var completedAt sql.NullString

	err := db.QueryRowContext(ctx, `
		SELECT
			generation,
			reason,
			status,
			failure_reason,
			started_at,
			completed_at
		FROM refresh_runs
		WHERE repository_id = ?
		ORDER BY generation DESC, id DESC
		LIMIT 1
	`, repositoryID).Scan(
		&run.Generation,
		&reason,
		&status,
		&failureReason,
		&startedAt,
		&completedAt,
	)
	if err != nil {
		return repository.DoctorRefreshRun{}, err
	}

	run.Present = true
	run.Reason = repository.RefreshReason(reason)
	run.Status = repository.RefreshRunStatus(status)
	run.FailureReason = failureReason.String
	run.StartedAt = parseRFC3339OrZero(startedAt)
	run.CompletedAt = parseNullableRFC3339(completedAt)
	return run, nil
}

func readDoctorStructuralCoverage(ctx context.Context, db *sql.DB, repositoryID int64) (repository.DoctorStructuralSection, error) {
	var summary repository.RepositoryStructuralCoverageSummary
	summary.RepositoryID = repositoryID
	err := db.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN f.ignore_status = ? THEN 1 ELSE 0 END), 0),
			COUNT(fe.id),
			COALESCE(SUM(CASE WHEN fe.coverage_state = 'supported' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN fe.coverage_state = 'partial' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN fe.coverage_state = 'unsupported' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN fe.coverage_state = 'failed' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN fe.coverage_state = 'skipped' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN fe.coverage_state IN ('partial', 'unsupported', 'failed', 'skipped') THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(fe.symbol_count), 0)
		FROM files f
		LEFT JOIN file_extractions fe ON fe.file_id = f.id
		WHERE f.repository_id = ?
	`, string(repository.IgnoreStatusIncluded), repositoryID).Scan(
		&summary.IncludedFileCount,
		&summary.ExtractionCount,
		&summary.SupportedCount,
		&summary.PartialCount,
		&summary.UnsupportedCount,
		&summary.FailedCount,
		&summary.SkippedCount,
		&summary.FilesWithCoverageGap,
		&summary.TotalSymbolCount,
	)
	if err != nil {
		return repository.DoctorStructuralSection{}, err
	}

	rows, err := db.QueryContext(ctx, `
		SELECT
			f.path,
			f.language,
			fe.coverage_state,
			COALESCE(fe.coverage_reason, ''),
			fe.symbol_count
		FROM files f
		JOIN file_extractions fe ON fe.file_id = f.id
		WHERE f.repository_id = ?
			AND f.ignore_status = ?
			AND fe.coverage_state IN ('partial', 'failed', 'unsupported', 'skipped')
		ORDER BY
			CASE fe.coverage_state
				WHEN 'failed' THEN 0
				WHEN 'partial' THEN 1
				WHEN 'unsupported' THEN 2
				ELSE 3
			END,
			f.path
		LIMIT 5
	`, repositoryID, string(repository.IgnoreStatusIncluded))
	if err != nil {
		return repository.DoctorStructuralSection{}, fmt.Errorf("query structural coverage examples: %w", err)
	}
	defer rows.Close()

	examples := make([]repository.DoctorStructuralCoverageExample, 0)
	for rows.Next() {
		var example repository.DoctorStructuralCoverageExample
		var language sql.NullString
		var stateValue string
		var reason string
		if err := rows.Scan(&example.Path, &language, &stateValue, &reason, &example.SymbolCount); err != nil {
			return repository.DoctorStructuralSection{}, fmt.Errorf("scan structural coverage example: %w", err)
		}
		example.Language = language.String
		example.CoverageState = repository.ExtractionCoverageState(stateValue)
		example.CoverageReason = repository.ExtractionCoverageReason(reason)
		examples = append(examples, example)
	}
	if err := rows.Err(); err != nil {
		return repository.DoctorStructuralSection{}, fmt.Errorf("iterate structural coverage examples: %w", err)
	}

	return repository.DoctorStructuralSection{
		Status:   doctorStructuralStatus(summary),
		Summary:  summary,
		Examples: examples,
	}, nil
}

func doctorStateStatus(health repository.HealthResult, repositoryMatch bool) repository.DoctorStatus {
	if health.Summary.StateStatus == repository.HealthStateStatusMissing {
		return repository.DoctorStatusMissing
	}
	if health.Summary.StateStatus == repository.HealthStateStatusPartial || !repositoryMatch {
		return repository.DoctorStatusDegraded
	}
	return repository.DoctorStatusHealthy
}

func doctorRefreshStatus(health repository.HealthResult) repository.DoctorStatus {
	switch {
	case !health.Refresh.Present:
		return repository.DoctorStatusMissing
	case health.Refresh.LastRefreshStatus == repository.RefreshRunStatusFailed:
		return repository.DoctorStatusDegraded
	case health.Refresh.Freshness == repository.FreshnessStatusPartiallyDegraded:
		return repository.DoctorStatusDegraded
	case health.Refresh.Freshness == repository.FreshnessStatusStale:
		return repository.DoctorStatusDegraded
	default:
		return repository.DoctorStatusHealthy
	}
}

func doctorWatchStatus(result repository.WatchStatusResult) repository.DoctorStatus {
	switch result.Status {
	case repository.WatchStatusKindRunning:
		return repository.DoctorStatusHealthy
	case repository.WatchStatusKindAbsent:
		return repository.DoctorStatusHealthy
	default:
		return repository.DoctorStatusDegraded
	}
}

func doctorWatchSummary(result repository.WatchStatusResult) string {
	switch result.Status {
	case repository.WatchStatusKindAbsent:
		return "runtime watch loop is not running"
	case repository.WatchStatusKindRunning:
		return "runtime watch loop is running"
	default:
		return result.Reason
	}
}

func doctorStructuralStatus(summary repository.RepositoryStructuralCoverageSummary) repository.DoctorStatus {
	switch {
	case summary.IncludedFileCount == 0:
		return repository.DoctorStatusMissing
	case summary.FailedCount > 0 || summary.PartialCount > 0:
		return repository.DoctorStatusDegraded
	default:
		return repository.DoctorStatusHealthy
	}
}

func doctorBudgetStatus(result repository.BudgetAnalysisResult) repository.DoctorStatus {
	if result.Summary.ReturnedCount == 0 {
		return repository.DoctorStatusMissing
	}
	return repository.DoctorStatusHealthy
}

func doctorSummary(report repository.DoctorReport) repository.DoctorSummary {
	issues := make([]repository.DoctorIssue, 0)
	addIssue := func(section string, status repository.DoctorStatus, summary string, action string) {
		if status == repository.DoctorStatusHealthy {
			return
		}
		issues = append(issues, repository.DoctorIssue{
			Section: section,
			Status:  status,
			Summary: summary,
			Action:  action,
		})
	}

	addIssue("state", report.State.Status, doctorStateIssue(report), doctorStateAction(report))
	addIssue("refresh", report.Refresh.Status, doctorRefreshIssue(report), doctorRefreshAction(report))
	addIssue("watch", report.Watch.Status, doctorWatchIssue(report), doctorWatchAction(report))
	addIssue("structural", report.Structural.Status, doctorStructuralIssue(report), doctorStructuralAction(report))
	addIssue("budget", report.Budget.Status, "no persisted token-cost hotspots available", "run `optimusctx run` so runtime refresh can persist budget analysis inputs")
	addIssue("mcp", report.MCPReadiness.Status, doctorMCPIssue(report), "use `optimusctx init --client <client> [--write]` for claude-desktop, claude-cli, codex-app, codex-cli, gemini-cli, or cursor-cli to preview or register the MCP contract")
	addIssue("mcp-registration", report.HostMCP.Status, doctorMCPRegistrationIssue(report), doctorMCPRegistrationAction(report))
	if doctorShouldReportMCPActivityIssue(report) {
		addIssue("mcp-usage", report.MCPActivity.Status, doctorMCPActivityIssue(report), doctorMCPActivityAction(report))
	}

	status := repository.DoctorStatusHealthy
	for _, issue := range issues {
		if issue.Status == repository.DoctorStatusDegraded {
			status = repository.DoctorStatusDegraded
			break
		}
		if issue.Status == repository.DoctorStatusMissing && status != repository.DoctorStatusDegraded {
			status = repository.DoctorStatusMissing
		}
	}

	return repository.DoctorSummary{
		Status:             status,
		RepositoryDetected: report.Identity.RootPath != "",
		Initialized:        report.State.Metadata.Present && report.Refresh.Health.Present,
		Issues:             issues,
	}
}

func doctorStateIssue(report repository.DoctorReport) string {
	switch report.State.Status {
	case repository.DoctorStatusMissing:
		return "repository state directory is not initialized"
	case repository.DoctorStatusDegraded:
		if !report.State.RepositoryMatch {
			return "state metadata points at a different repository root"
		}
		return "state layout is only partially present"
	default:
		return ""
	}
}

func doctorStateAction(report repository.DoctorReport) string {
	if report.State.Status == repository.DoctorStatusMissing {
		return "run `optimusctx init` from the repository root to create `.optimusctx/`"
	}
	if !report.State.RepositoryMatch {
		return "remove the stale `.optimusctx/` directory or re-run `optimusctx init` in the correct repository"
	}
	return "restore the missing files under `.optimusctx/` or re-run `optimusctx init` to repair state"
}

func doctorRefreshIssue(report repository.DoctorReport) string {
	switch report.Refresh.Status {
	case repository.DoctorStatusMissing:
		return "no persisted refresh history is available yet"
	case repository.DoctorStatusDegraded:
		if report.Refresh.LastRun.Status == repository.RefreshRunStatusFailed && report.Refresh.LastRun.FailureReason != "" {
			return fmt.Sprintf("last refresh failed: %s", report.Refresh.LastRun.FailureReason)
		}
		if report.Refresh.Health.Freshness == repository.FreshnessStatusStale {
			return "repository freshness is stale"
		}
		if report.Refresh.Health.Freshness == repository.FreshnessStatusPartiallyDegraded {
			return "repository freshness is partially degraded"
		}
		return "refresh state is degraded"
	default:
		return ""
	}
}

func doctorRefreshAction(report repository.DoctorReport) string {
	if report.Refresh.Status == repository.DoctorStatusMissing {
		return "run `optimusctx refresh` to register the repository and persist the first snapshot"
	}
	return "run `optimusctx refresh` and inspect `.optimusctx/logs/` if refresh stays degraded"
}

func doctorWatchIssue(report repository.DoctorReport) string {
	if report.Watch.Status == repository.DoctorStatusDegraded {
		return report.Watch.Health.Reason
	}
	return ""
}

func doctorWatchAction(report repository.DoctorReport) string {
	if report.Watch.Health.Status == repository.WatchStatusKindAbsent {
		return "start `optimusctx watch run` only if you want continuous refreshes"
	}
	return "restart `optimusctx watch run` so the heartbeat and refresh generation recover"
}

func doctorStructuralIssue(report repository.DoctorReport) string {
	switch report.Structural.Status {
	case repository.DoctorStatusMissing:
		return "no structural coverage summary is available"
	case repository.DoctorStatusDegraded:
		return fmt.Sprintf("%d files have failed or partial structural coverage", report.Structural.Summary.FailedCount+report.Structural.Summary.PartialCount)
	default:
		return ""
	}
}

func doctorStructuralAction(report repository.DoctorReport) string {
	if report.Structural.Status == repository.DoctorStatusMissing {
		return "run `optimusctx run` so structural extraction artifacts are persisted"
	}
	return "inspect the flagged files and re-run `optimusctx run` after fixing parser or extraction issues"
}

func doctorMCPIssue(report repository.DoctorReport) string {
	if report.MCPReadiness.Status == repository.DoctorStatusHealthy {
		return ""
	}
	if report.MCPReadiness.SnippetParseFailure != "" {
		return report.MCPReadiness.SnippetParseFailure
	}
	return "MCP readiness preview could not be rendered"
}

func doctorMCPRegistrationIssue(report repository.DoctorReport) string {
	if report.HostMCP.Status == repository.DoctorStatusHealthy {
		return ""
	}
	if report.HostMCP.Status == repository.DoctorStatusMissing {
		return "no supported MCP host registration was detected"
	}
	return "some supported MCP hosts could not be verified or are missing OptimusCtx guidance"
}

func doctorMCPRegistrationAction(report repository.DoctorReport) string {
	return "run `optimusctx init --client <client> [--write]` for the host you expect to use, then rerun `optimusctx status`"
}

func doctorMCPActivityIssue(report repository.DoctorReport) string {
	if report.MCPActivity.Status == repository.DoctorStatusHealthy {
		return ""
	}
	if report.MCPActivity.LastInitializeAt.IsZero() {
		return "no host session has initialized the OptimusCtx MCP server yet"
	}
	if report.MCPActivity.LastToolsListAt.IsZero() {
		return "a host connected, but no tools discovery has been recorded yet"
	}
	return "OptimusCtx tools have been discovered but no tool calls were recorded yet"
}

func doctorMCPActivityAction(report repository.DoctorReport) string {
	return "start a session from your registered host and verify it exposes and calls `optimusctx.*` tools; `optimusctx status` will record the evidence"
}

func doctorShouldReportMCPActivityIssue(report repository.DoctorReport) bool {
	if report.HostMCP.Status != repository.DoctorStatusHealthy {
		return false
	}
	if report.MCPActivity.LastSessionStartAt.IsZero() {
		return false
	}
	return report.MCPActivity.Status != repository.DoctorStatusHealthy
}

func doctorRecommendedFixes(issues []repository.DoctorIssue) []string {
	if len(issues) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(issues))
	fixes := make([]string, 0, len(issues))
	for _, issue := range issues {
		action := strings.TrimSpace(issue.Action)
		if action == "" {
			continue
		}
		if _, ok := seen[action]; ok {
			continue
		}
		seen[action] = struct{}{}
		fixes = append(fixes, action)
	}
	return fixes
}

func (s DoctorService) resolveLayout(root string) (state.Layout, error) {
	if s.ResolveLayout != nil {
		return s.ResolveLayout(root)
	}
	return state.ResolveLayout(root)
}

func (s DoctorService) openDB(path string) (*sql.DB, error) {
	if s.OpenDB != nil {
		return s.OpenDB(path)
	}
	return sql.Open("sqlite", "file:"+path+"?mode=ro")
}

func (s DoctorService) getwd() (string, error) {
	if s.Getwd != nil {
		return s.Getwd()
	}
	return os.Getwd()
}

func (s DoctorService) hostRegistrations(ctx context.Context, repoRoot string) (repository.DoctorHostRegistrationSection, error) {
	clients := repository.SupportedClients()
	hosts := make([]repository.DoctorHostRegistration, 0, len(clients)-1)

	for _, client := range clients {
		if !client.IsFirstClass() {
			continue
		}
		switch client.ID {
		case repository.ClientClaudeDesktop:
			claudeDesktopPath, err := resolveClaudeDesktopConfigPath("")
			if err == nil {
				hosts = append(hosts, s.detectClaudeDesktop(client, claudeDesktopPath))
				continue
			}
			hosts = append(hosts, repository.DoctorHostRegistration{
				Client:               client,
				RegistrationState:    repository.HostRegistrationUnverified,
				RegistrationPath:     "",
				RegistrationEvidence: err.Error(),
				GuidanceState:        repository.GuidanceStateUnsupported,
				GuidanceEvidence:     "Claude Desktop registration is supported, but durable agent guidance is not managed through Claude Desktop config.",
				CapabilityEvidence:   client.CapabilitySummary(),
			})
		case repository.ClientClaudeCLI:
			hosts = append(hosts, s.detectClaudeCLI(ctx, client, repoRoot))
		case repository.ClientCodexApp:
			codexAppSharedPath, _ := resolveCodexAppConfigPath("")
			hosts = append(hosts, s.detectCodexClient(client, repoRoot, codexAppSharedPath))
		case repository.ClientCodexCLI:
			codexCLISharedPath, _ := resolveCodexCLIConfigPath("")
			hosts = append(hosts, s.detectCodexClient(client, repoRoot, codexCLISharedPath))
		case repository.ClientGeminiCLI:
			geminiSharedPath, _ := resolveGeminiCLIConfigPath("")
			hosts = append(hosts, s.detectGeminiCLI(client, repoRoot, geminiSharedPath))
		case repository.ClientCursorCLI:
			cursorSharedPath, _ := resolveCursorCLIConfigPath("")
			hosts = append(hosts, s.detectCursorCLI(client, repoRoot, cursorSharedPath))
		}
	}

	status := repository.DoctorStatusMissing
	hasDetected := false
	hasDetectedWithoutGuidance := false
	for _, host := range hosts {
		if host.RegistrationState == repository.HostRegistrationDetected {
			hasDetected = true
			if host.GuidanceState == repository.GuidanceStateNotConfigured {
				hasDetectedWithoutGuidance = true
			}
		}
	}
	switch {
	case hasDetected && !hasDetectedWithoutGuidance:
		status = repository.DoctorStatusHealthy
	case hasDetected || hasDetectedWithoutGuidance:
		status = repository.DoctorStatusDegraded
	}

	return repository.DoctorHostRegistrationSection{Status: status, Hosts: hosts}, nil
}

func (s DoctorService) detectClaudeDesktop(client repository.SupportedClient, configPath string) repository.DoctorHostRegistration {
	host := repository.DoctorHostRegistration{
		Client:             client,
		RegistrationState:  repository.HostRegistrationNotDetected,
		RegistrationPath:   configPath,
		GuidanceState:      repository.GuidanceStateUnsupported,
		GuidanceEvidence:   "Claude Desktop registration is supported, but durable agent guidance is not managed through Claude Desktop config.",
		CapabilityEvidence: client.CapabilitySummary(),
	}

	existing, err := readExistingClientConfig(s.readFileFn(), configPath)
	if err != nil {
		host.RegistrationState = repository.HostRegistrationUnverified
		host.RegistrationEvidence = err.Error()
		return host
	}
	if len(existing) == 0 {
		host.RegistrationEvidence = "Claude Desktop config not found at the default path"
		return host
	}

	document, err := repository.ParseClientConfig(existing)
	if err != nil {
		host.RegistrationState = repository.HostRegistrationUnverified
		host.RegistrationEvidence = err.Error()
		return host
	}
	if _, ok := document.MCPServers[repository.DefaultMCPServerName]; ok {
		host.RegistrationState = repository.HostRegistrationDetected
		host.RegistrationEvidence = "found OptimusCtx in Claude Desktop config"
		return host
	}
	host.RegistrationEvidence = "Claude Desktop config does not contain an OptimusCtx MCP entry"
	return host
}

func (s DoctorService) detectClaudeCLI(ctx context.Context, client repository.SupportedClient, repoRoot string) repository.DoctorHostRegistration {
	host := repository.DoctorHostRegistration{
		Client:             client,
		RegistrationState:  repository.HostRegistrationUnverified,
		GuidanceState:      repository.GuidanceStateUnverified,
		CapabilityEvidence: client.CapabilitySummary(),
	}

	projectPath := filepath.Join(repoRoot, ".mcp.json")
	if existing, err := readExistingClientConfig(s.readFileFn(), projectPath); err == nil && len(existing) > 0 {
		document, err := repository.ParseClientConfig(existing)
		if err == nil {
			if _, ok := document.MCPServers[repository.DefaultMCPServerName]; ok {
				host.RegistrationState = repository.HostRegistrationDetected
				host.RegistrationPath = projectPath
				host.RegistrationEvidence = "found OptimusCtx in project `.mcp.json`"
				host.GuidancePath = filepath.Join(repoRoot, ".claude", "rules", repository.ClaudeRulesFilename)
				if s.fileExists(host.GuidancePath) {
					host.GuidanceState = repository.GuidanceStateConfigured
					host.GuidanceEvidence = "found project Claude rule for OptimusCtx"
				} else {
					host.GuidanceState = repository.GuidanceStateNotConfigured
					host.GuidanceEvidence = "project Claude rule for OptimusCtx is missing"
				}
				return host
			}
		}
	}

	output, err := s.runCommandFn()(ctx, "claude", "mcp", "list")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			host.RegistrationEvidence = "claude binary not found; could not verify local/user MCP registrations"
			return host
		}
		host.RegistrationEvidence = strings.TrimSpace(string(output))
		if host.RegistrationEvidence == "" {
			host.RegistrationEvidence = err.Error()
		}
		return host
	}

	trimmed := strings.TrimSpace(string(output))
	if strings.Contains(trimmed, repository.DefaultMCPServerName) {
		host.RegistrationState = repository.HostRegistrationDetected
		host.RegistrationEvidence = "Claude CLI reports an OptimusCtx MCP registration"
		homeDir, homeErr := os.UserHomeDir()
		if homeErr == nil {
			host.GuidancePath = filepath.Join(homeDir, ".claude", "rules", repository.ClaudeRulesFilename)
			if s.fileExists(host.GuidancePath) {
				host.GuidanceState = repository.GuidanceStateConfigured
				host.GuidanceEvidence = "found user Claude rule for OptimusCtx"
			} else {
				host.GuidanceState = repository.GuidanceStateNotConfigured
				host.GuidanceEvidence = "user Claude rule for OptimusCtx is missing"
			}
		}
		return host
	}

	host.RegistrationState = repository.HostRegistrationNotDetected
	host.RegistrationEvidence = "Claude CLI did not report an OptimusCtx MCP registration"
	if strings.TrimSpace(repoRoot) != "" {
		host.GuidancePath = filepath.Join(repoRoot, ".claude", "rules", repository.ClaudeRulesFilename)
	}
	return host
}

func (s DoctorService) detectCodexClient(client repository.SupportedClient, repoRoot string, sharedConfigPath string) repository.DoctorHostRegistration {
	host := repository.DoctorHostRegistration{
		Client:             client,
		RegistrationState:  repository.HostRegistrationNotDetected,
		GuidanceState:      repository.GuidanceStateNotConfigured,
		CapabilityEvidence: client.CapabilitySummary(),
	}

	repoConfigPath := filepath.Join(repoRoot, ".codex", "config.toml")
	paths := []string{repoConfigPath}
	if strings.TrimSpace(sharedConfigPath) != "" {
		paths = append(paths, sharedConfigPath)
	}
	for _, path := range paths {
		existing, err := readExistingClientConfig(s.readFileFn(), path)
		if err != nil {
			host.RegistrationState = repository.HostRegistrationUnverified
			host.RegistrationEvidence = err.Error()
			return host
		}
		if len(existing) == 0 {
			continue
		}
		document, err := repository.CodexConfigServerNames(existing)
		if err != nil {
			host.RegistrationState = repository.HostRegistrationUnverified
			host.RegistrationEvidence = err.Error()
			host.RegistrationPath = path
			return host
		}
		if document[repository.DefaultMCPServerName] {
			host.RegistrationState = repository.HostRegistrationDetected
			host.RegistrationPath = path
			if path == repoConfigPath {
				host.RegistrationEvidence = "found OptimusCtx in repo `.codex/config.toml`"
				host.GuidancePath = s.codexGuidancePath(repoRoot)
			} else {
				host.RegistrationEvidence = "found OptimusCtx in shared Codex config"
				host.GuidancePath = s.codexGuidancePath(filepath.Dir(sharedConfigPath))
			}
			if s.fileContains(host.GuidancePath, "OptimusCtx MCP guidance") {
				host.GuidanceState = repository.GuidanceStateConfigured
				host.GuidanceEvidence = "found managed Codex guidance block"
			} else {
				host.GuidanceState = repository.GuidanceStateNotConfigured
				host.GuidanceEvidence = "Codex guidance block is missing"
			}
			return host
		}
	}

	host.RegistrationEvidence = "no Codex config currently references OptimusCtx"
	host.GuidancePath = s.codexGuidancePath(repoRoot)
	return host
}

func (s DoctorService) detectGeminiCLI(client repository.SupportedClient, repoRoot string, sharedConfigPath string) repository.DoctorHostRegistration {
	host := repository.DoctorHostRegistration{
		Client:             client,
		RegistrationState:  repository.HostRegistrationNotDetected,
		GuidanceState:      repository.GuidanceStateNotConfigured,
		CapabilityEvidence: client.CapabilitySummary(),
	}

	repoConfigPath := filepath.Join(repoRoot, ".gemini", "settings.json")
	paths := []string{repoConfigPath}
	if strings.TrimSpace(sharedConfigPath) != "" {
		paths = append(paths, sharedConfigPath)
	}
	for _, path := range paths {
		existing, err := readExistingClientConfig(s.readFileFn(), path)
		if err != nil {
			host.RegistrationState = repository.HostRegistrationUnverified
			host.RegistrationEvidence = err.Error()
			return host
		}
		if len(existing) == 0 {
			continue
		}
		document, err := repository.GeminiConfigServerNames(existing)
		if err != nil {
			host.RegistrationState = repository.HostRegistrationUnverified
			host.RegistrationEvidence = err.Error()
			host.RegistrationPath = path
			return host
		}
		if document[repository.DefaultMCPServerName] {
			host.RegistrationState = repository.HostRegistrationDetected
			host.RegistrationPath = path
			if path == repoConfigPath {
				host.RegistrationEvidence = "found OptimusCtx in repo `.gemini/settings.json`"
				host.GuidancePath = filepath.Join(repoRoot, repository.GeminiGuidanceFilename)
			} else {
				host.RegistrationEvidence = "found OptimusCtx in shared Gemini config"
				host.GuidancePath = filepath.Join(filepath.Dir(sharedConfigPath), repository.GeminiGuidanceFilename)
			}
			if s.fileContains(host.GuidancePath, "OptimusCtx MCP guidance") {
				host.GuidanceState = repository.GuidanceStateConfigured
				host.GuidanceEvidence = "found managed Gemini guidance block"
			} else {
				host.GuidanceState = repository.GuidanceStateNotConfigured
				host.GuidanceEvidence = "Gemini guidance block is missing"
			}
			return host
		}
	}

	host.RegistrationEvidence = "no Gemini config currently references OptimusCtx"
	host.GuidancePath = filepath.Join(repoRoot, repository.GeminiGuidanceFilename)
	return host
}

func (s DoctorService) detectCursorCLI(client repository.SupportedClient, repoRoot string, sharedConfigPath string) repository.DoctorHostRegistration {
	host := repository.DoctorHostRegistration{
		Client:             client,
		RegistrationState:  repository.HostRegistrationNotDetected,
		GuidanceState:      repository.GuidanceStateUnsupported,
		GuidanceEvidence:   "Cursor CLI registration is supported, but durable agent guidance is not managed through Cursor config.",
		CapabilityEvidence: client.CapabilitySummary(),
	}

	repoConfigPath := filepath.Join(repoRoot, ".cursor", "mcp.json")
	paths := []string{repoConfigPath}
	if strings.TrimSpace(sharedConfigPath) != "" {
		paths = append(paths, sharedConfigPath)
	}
	for _, path := range paths {
		existing, err := readExistingClientConfig(s.readFileFn(), path)
		if err != nil {
			host.RegistrationState = repository.HostRegistrationUnverified
			host.RegistrationEvidence = err.Error()
			return host
		}
		if len(existing) == 0 {
			continue
		}
		document, err := repository.ParseClientConfig(existing)
		if err != nil {
			host.RegistrationState = repository.HostRegistrationUnverified
			host.RegistrationEvidence = err.Error()
			host.RegistrationPath = path
			return host
		}
		if _, ok := document.MCPServers[repository.DefaultMCPServerName]; ok {
			host.RegistrationState = repository.HostRegistrationDetected
			host.RegistrationPath = path
			if path == repoConfigPath {
				host.RegistrationEvidence = "found OptimusCtx in repo `.cursor/mcp.json`"
			} else {
				host.RegistrationEvidence = "found OptimusCtx in shared Cursor config"
			}
			return host
		}
	}

	host.RegistrationEvidence = "no Cursor config currently references OptimusCtx"
	return host
}

func (s DoctorService) codexGuidancePath(root string) string {
	overridePath := filepath.Join(root, "AGENTS.override.md")
	overrideContent, err := readExistingClientConfig(s.readFileFn(), overridePath)
	if err == nil && strings.TrimSpace(string(overrideContent)) != "" {
		return overridePath
	}
	return filepath.Join(root, "AGENTS.md")
}

func (s DoctorService) readFileFn() func(string) ([]byte, error) {
	if s.ReadFile != nil {
		return s.ReadFile
	}
	return os.ReadFile
}

func (s DoctorService) runCommandFn() func(context.Context, string, ...string) ([]byte, error) {
	if s.RunCommand != nil {
		return s.RunCommand
	}
	return func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return exec.CommandContext(ctx, name, args...).CombinedOutput()
	}
}

func (s DoctorService) fileExists(path string) bool {
	content, err := readExistingClientConfig(s.readFileFn(), path)
	return err == nil && len(content) > 0
}

func (s DoctorService) fileContains(path string, needle string) bool {
	content, err := readExistingClientConfig(s.readFileFn(), path)
	return err == nil && strings.Contains(string(content), needle)
}
