package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/repository"
)

var statusCommandService = func(ctx context.Context, workingDir string) (repository.DoctorReport, error) {
	return app.NewDoctorService().Doctor(ctx, workingDir, repository.DoctorRequest{})
}

func newStatusCommand() *Command {
	return &Command{
		Name:    "status",
		Summary: "Show the canonical OptimusCtx runtime, host registration, and MCP usage status",
		Run: func(stdout io.Writer, args []string) error {
			verbose := false
			for _, arg := range args {
				switch arg {
				case "-h", "--help":
					_, err := io.WriteString(stdout, "Usage:\n  optimusctx status [--verbose]\n\nShow the canonical OptimusCtx runtime status, host registration evidence, and MCP usage evidence without mutating state.\n")
					return err
				case "-v", "--verbose":
					verbose = true
				default:
					if len(arg) > 0 && arg[0] == '-' {
						return fmt.Errorf("status does not accept flag %q", arg)
					}
					return fmt.Errorf("status does not accept argument %q", arg)
				}
			}

			workingDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve working directory: %w", err)
			}

			report, err := statusCommandService(context.Background(), workingDir)
			if err != nil {
				return err
			}

			_, err = io.WriteString(stdout, formatStatusReport(report, verbose))
			return err
		},
	}
}

func formatStatusReport(report repository.DoctorReport, verbose bool) string {
	var b strings.Builder
	writeDoctorLine(&b, "status", statusHeadline(report.Summary.Status))
	writeDoctorLine(&b, "repository", report.Identity.RootPath)
	writeDoctorLine(&b, "freshness", formatStatusFreshness(report))
	writeDoctorLine(&b, "refresh", formatStatusRefresh(report))
	writeDoctorLine(&b, "watch", formatStatusWatch(report))

	b.WriteString("\nMCP\n")
	writeDoctorLine(&b, "setup", formatStatusMCPSetup(report))
	if path := statusPrimaryRegistrationPath(report); path != "" {
		writeDoctorLine(&b, "config", path)
	}
	if path := statusPrimaryGuidancePath(report); path != "" {
		writeDoctorLine(&b, "agent guidance", path)
	}
	writeDoctorLine(&b, "runtime launch", formatStatusMCPLaunch(report))
	writeDoctorLine(&b, "usage", formatStatusMCPUsage(report))
	if evidence := formatStatusMCPEvidence(report); evidence != "" {
		writeDoctorLine(&b, "evidence", evidence)
	}

	if issues := statusRenderableIssues(report); len(issues) > 0 {
		b.WriteString("\nAttention\n")
		for _, issue := range issues {
			writeDoctorLine(&b, "item", issue.Summary)
		}
	}

	if next := statusNextSteps(report); len(next) > 0 {
		b.WriteString("\nNext\n")
		for _, step := range next {
			writeDoctorLine(&b, "step", step)
		}
	}

	if !verbose {
		b.WriteString("\nDetails\n")
		writeDoctorLine(&b, "tip", "rerun `optimusctx status --verbose` for full diagnostics")
		return b.String()
	}

	b.WriteString("\nDiagnostics\n")
	b.WriteString(formatStatusDiagnosticReport(report))
	return b.String()
}

func statusHeadline(status repository.DoctorStatus) string {
	switch status {
	case repository.DoctorStatusHealthy:
		return "ready"
	case repository.DoctorStatusDegraded:
		return "needs attention"
	case repository.DoctorStatusMissing:
		return "setup needed"
	default:
		return string(status)
	}
}

func formatStatusFreshness(report repository.DoctorReport) string {
	return fmt.Sprintf("%s (generation %d)", report.Repository.Freshness, report.Repository.Generation)
}

func formatStatusRefresh(report repository.DoctorReport) string {
	switch {
	case report.Refresh.Status == repository.DoctorStatusMissing:
		return "not initialized yet"
	case report.Refresh.LastRun.Status == repository.RefreshRunStatusFailed && report.Refresh.LastRun.FailureReason != "":
		return fmt.Sprintf("failed: %s", report.Refresh.LastRun.FailureReason)
	case !report.Refresh.Health.LastRefreshCompletedAt.IsZero():
		return fmt.Sprintf("%s at %s", renderDoctorValue(string(report.Refresh.Health.LastRefreshStatus)), formatDoctorTime(report.Refresh.Health.LastRefreshCompletedAt))
	default:
		return renderDoctorValue(string(report.Refresh.Health.LastRefreshStatus))
	}
}

func formatStatusWatch(report repository.DoctorReport) string {
	switch report.Watch.Health.Status {
	case repository.WatchStatusKindAbsent:
		return "not running (optional)"
	case repository.WatchStatusKindRunning:
		return "running"
	default:
		return renderDoctorValue(report.Watch.Summary)
	}
}

func formatStatusMCPSetup(report repository.DoctorReport) string {
	detected := statusDetectedHosts(report)
	if len(detected) == 0 {
		if report.HostMCP.Status == repository.DoctorStatusDegraded {
			return "partially configured or not fully verified"
		}
		return "not configured for a supported host yet"
	}

	names := make([]string, 0, len(detected))
	for _, host := range detected {
		names = append(names, host.Client.DisplayName)
	}
	return fmt.Sprintf("configured for %s", strings.Join(names, ", "))
}

func formatStatusMCPLaunch(report repository.DoctorReport) string {
	if len(statusDetectedHosts(report)) > 0 {
		return "automatic via the registered host (`optimusctx run`)"
	}
	return "will be automatic after host registration (`optimusctx run`)"
}

func formatStatusMCPUsage(report repository.DoctorReport) string {
	switch {
	case !report.MCPActivity.LastToolCallAt.IsZero():
		return "confirmed"
	case !report.MCPActivity.LastToolsListAt.IsZero():
		return "discovered, but no tool calls recorded yet"
	case !report.MCPActivity.LastInitializeAt.IsZero() || !report.MCPActivity.LastSessionStartAt.IsZero():
		return "host connected, but discovery is not recorded yet"
	case len(statusDetectedHosts(report)) > 0:
		return "not observed yet"
	default:
		return "n/a"
	}
}

func formatStatusMCPEvidence(report repository.DoctorReport) string {
	switch {
	case !report.MCPActivity.LastToolCallAt.IsZero() && len(report.MCPActivity.RecentToolCalls) > 0:
		last := report.MCPActivity.RecentToolCalls[len(report.MCPActivity.RecentToolCalls)-1]
		return fmt.Sprintf("last tool call %s at %s", last.Name, last.At)
	case !report.MCPActivity.LastToolsListAt.IsZero():
		return fmt.Sprintf("tools listed at %s", formatDoctorTime(report.MCPActivity.LastToolsListAt))
	case !report.MCPActivity.LastInitializeAt.IsZero():
		return fmt.Sprintf("host initialized the MCP server at %s", formatDoctorTime(report.MCPActivity.LastInitializeAt))
	case !report.MCPActivity.LastSessionStartAt.IsZero():
		return fmt.Sprintf("session started at %s", formatDoctorTime(report.MCPActivity.LastSessionStartAt))
	case len(statusDetectedHosts(report)) > 0:
		return "no host session has been recorded yet"
	default:
		return ""
	}
}

func statusRenderableIssues(report repository.DoctorReport) []repository.DoctorIssue {
	issues := make([]repository.DoctorIssue, 0, len(report.Summary.Issues))
	for _, issue := range report.Summary.Issues {
		if issue.Section == "mcp-usage" && report.MCPActivity.LastSessionStartAt.IsZero() {
			continue
		}
		issues = append(issues, issue)
	}
	return issues
}

func statusNextSteps(report repository.DoctorReport) []string {
	if issues := statusRenderableIssues(report); len(issues) > 0 {
		return statusRecommendedSteps(issues)
	}
	if len(statusDetectedHosts(report)) > 0 && report.MCPActivity.LastSessionStartAt.IsZero() {
		return []string{"open your registered host and use the repo once; rerun `optimusctx status` to confirm MCP discovery and tool calls"}
	}
	return nil
}

func statusRecommendedSteps(issues []repository.DoctorIssue) []string {
	if len(issues) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(issues))
	steps := make([]string, 0, len(issues))
	for _, issue := range issues {
		action := strings.TrimSpace(issue.Action)
		if action == "" {
			continue
		}
		if _, ok := seen[action]; ok {
			continue
		}
		seen[action] = struct{}{}
		steps = append(steps, action)
	}
	return steps
}

func statusDetectedHosts(report repository.DoctorReport) []repository.DoctorHostRegistration {
	hosts := make([]repository.DoctorHostRegistration, 0, len(report.HostMCP.Hosts))
	for _, host := range report.HostMCP.Hosts {
		if host.RegistrationState == repository.HostRegistrationDetected {
			hosts = append(hosts, host)
		}
	}
	return hosts
}

func statusPrimaryRegistrationPath(report repository.DoctorReport) string {
	for _, host := range statusDetectedHosts(report) {
		if strings.TrimSpace(host.RegistrationPath) != "" {
			return host.RegistrationPath
		}
	}
	return ""
}

func statusPrimaryGuidancePath(report repository.DoctorReport) string {
	for _, host := range statusDetectedHosts(report) {
		if host.GuidanceState == repository.GuidanceStateConfigured && strings.TrimSpace(host.GuidancePath) != "" {
			return host.GuidancePath
		}
	}
	return ""
}

func formatStatusDiagnosticReport(report repository.DoctorReport) string {
	var b strings.Builder
	writeDoctorLine(&b, "overall status", string(report.Summary.Status))
	writeDoctorLine(&b, "repository root", report.Identity.RootPath)
	writeDoctorLine(&b, "detection mode", report.Identity.DetectionMode)
	writeDoctorLine(&b, "runtime version", report.Install.BinaryVersion)
	writeDoctorLine(&b, "working directory", report.Install.WorkingDir)
	writeDoctorLine(&b, "freshness", string(report.Repository.Freshness))
	writeDoctorLine(&b, "generation", formatDoctorInt64(report.Repository.Generation))

	b.WriteString("\nState\n")
	writeDoctorLine(&b, "status", string(report.State.Status))
	writeDoctorLine(&b, "state dir", report.State.Layout.StateDir.Path)
	writeDoctorLine(&b, "database file", report.State.Layout.DatabaseFile.Path)
	writeDoctorLine(&b, "metadata present", formatDoctorBool(report.State.Metadata.Present))
	writeDoctorLine(&b, "repository match", formatDoctorBool(report.State.RepositoryMatch))

	b.WriteString("\nRefresh\n")
	writeDoctorLine(&b, "status", string(report.Refresh.Status))
	writeDoctorLine(&b, "repository registered", formatDoctorBool(report.Refresh.Health.Present))
	writeDoctorLine(&b, "last refresh status", renderDoctorValue(string(report.Refresh.Health.LastRefreshStatus)))
	writeDoctorLine(&b, "freshness reason", renderDoctorValue(report.Refresh.Health.FreshnessReason))
	writeDoctorLine(&b, "last refresh completed", formatDoctorTime(report.Refresh.Health.LastRefreshCompletedAt))
	if report.Refresh.LastRun.Present {
		writeDoctorLine(&b, "latest run", fmt.Sprintf("generation=%d reason=%s status=%s", report.Refresh.LastRun.Generation, report.Refresh.LastRun.Reason, report.Refresh.LastRun.Status))
		writeDoctorLine(&b, "latest run failure", renderDoctorValue(report.Refresh.LastRun.FailureReason))
	}

	b.WriteString("\nWatch\n")
	writeDoctorLine(&b, "status", string(report.Watch.Status))
	writeDoctorLine(&b, "watch state", string(report.Watch.Health.Status))
	writeDoctorLine(&b, "summary", report.Watch.Summary)
	writeDoctorLine(&b, "optional", formatDoctorBool(report.Watch.Optional))
	writeDoctorLine(&b, "reason", doctorWatchReason(report))
	writeDoctorLine(&b, "status path", report.Watch.Health.StatusPath)
	writeDoctorLine(&b, "last heartbeat", formatDoctorTime(report.Watch.Health.Record.LastHeartbeatAtTime()))
	writeDoctorLine(&b, "last refresh generation", formatDoctorInt64(report.Watch.Health.Record.LastRefreshGeneration))
	writeDoctorLine(&b, "watch error", renderDoctorValue(report.Watch.Health.Record.LastError))

	b.WriteString("\nStructural Coverage\n")
	writeDoctorLine(&b, "status", string(report.Structural.Status))
	writeDoctorLine(&b, "included files", formatDoctorInt64(report.Structural.Summary.IncludedFileCount))
	writeDoctorLine(&b, "supported files", formatDoctorInt64(report.Structural.Summary.SupportedCount))
	writeDoctorLine(&b, "unsupported files", formatDoctorInt64(report.Structural.Summary.UnsupportedCount))
	writeDoctorLine(&b, "failed files", formatDoctorInt64(report.Structural.Summary.FailedCount))
	writeDoctorLine(&b, "partial files", formatDoctorInt64(report.Structural.Summary.PartialCount))
	for _, example := range report.Structural.Examples {
		writeDoctorLine(&b, "example", fmt.Sprintf("%s (%s, reason=%s, symbols=%d)", example.Path, example.CoverageState, renderDoctorValue(string(example.CoverageReason)), example.SymbolCount))
	}

	b.WriteString("\nBudget\n")
	writeDoctorLine(&b, "status", string(report.Budget.Status))
	writeDoctorLine(&b, "estimate policy", report.Budget.Policy.Name)
	writeDoctorLine(&b, "returned hotspots", formatDoctorInt(report.Budget.Summary.ReturnedCount))
	for _, hotspot := range report.Budget.Hotspots {
		writeDoctorLine(&b, "hotspot", fmt.Sprintf("%s tokens=%d files=%d bytes=%d", hotspot.Path, hotspot.EstimatedTokens, hotspot.IncludedFileCount, hotspot.TotalSizeBytes))
	}

	b.WriteString("\nMCP Host Registration\n")
	writeDoctorLine(&b, "status", string(report.HostMCP.Status))
	for _, host := range report.HostMCP.Hosts {
		writeDoctorLine(&b, "host", fmt.Sprintf("%s registration=%s guidance=%s registration_path=%s guidance_path=%s", host.Client.DisplayName, host.RegistrationState, host.GuidanceState, renderDoctorValue(host.RegistrationPath), renderDoctorValue(host.GuidancePath)))
		writeDoctorLine(&b, "host detail", fmt.Sprintf("%s: %s", host.Client.DisplayName, renderDoctorValue(host.RegistrationEvidence)))
		writeDoctorLine(&b, "guidance detail", fmt.Sprintf("%s: %s", host.Client.DisplayName, renderDoctorValue(host.GuidanceEvidence)))
		writeDoctorLine(&b, "capability detail", fmt.Sprintf("%s: %s", host.Client.DisplayName, renderDoctorValue(host.CapabilityEvidence)))
	}

	b.WriteString("\nMCP Evidence\n")
	writeDoctorLine(&b, "status", string(report.MCPActivity.Status))
	writeDoctorLine(&b, "last session start", formatDoctorTime(report.MCPActivity.LastSessionStartAt))
	writeDoctorLine(&b, "last initialize", formatDoctorTime(report.MCPActivity.LastInitializeAt))
	writeDoctorLine(&b, "last tools list", formatDoctorTime(report.MCPActivity.LastToolsListAt))
	writeDoctorLine(&b, "last tool call", formatDoctorTime(report.MCPActivity.LastToolCallAt))
	if len(report.MCPActivity.RecentToolCalls) == 0 {
		writeDoctorLine(&b, "recent tool call", "none")
	} else {
		for _, call := range report.MCPActivity.RecentToolCalls {
			writeDoctorLine(&b, "recent tool call", fmt.Sprintf("%s at %s", call.Name, call.At))
		}
	}

	b.WriteString("\nMCP Readiness\n")
	writeDoctorLine(&b, "status", string(report.MCPReadiness.Status))
	writeDoctorLine(&b, "server name", report.MCPReadiness.ServerName)
	writeDoctorLine(&b, "serve command", fmt.Sprintf("%s %s", report.MCPReadiness.ServeCommand.Command, strings.Join(report.MCPReadiness.ServeCommand.Args, " ")))
	writeDoctorLine(&b, "snippet available", formatDoctorBool(report.MCPReadiness.SnippetAvailable))

	return b.String()
}
