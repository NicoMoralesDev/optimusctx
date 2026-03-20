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
			for _, arg := range args {
				switch arg {
				case "-h", "--help":
					_, err := io.WriteString(stdout, "Usage:\n  optimusctx status\n\nShow the canonical OptimusCtx runtime status, host registration evidence, and MCP usage evidence without mutating state.\n")
					return err
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

			_, err = io.WriteString(stdout, formatStatusReport(report))
			return err
		},
	}
}

func formatStatusReport(report repository.DoctorReport) string {
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
	writeDoctorLine(&b, "coverage gaps", formatDoctorInt64(report.Structural.Summary.FilesWithCoverageGap))
	writeDoctorLine(&b, "failed files", formatDoctorInt64(report.Structural.Summary.FailedCount))
	writeDoctorLine(&b, "partial files", formatDoctorInt64(report.Structural.Summary.PartialCount))
	for _, example := range report.Structural.Examples {
		writeDoctorLine(&b, "gap", fmt.Sprintf("%s (%s, reason=%s, symbols=%d)", example.Path, example.CoverageState, renderDoctorValue(string(example.CoverageReason)), example.SymbolCount))
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

	b.WriteString("\nIssues\n")
	if len(report.Summary.Issues) == 0 {
		writeDoctorLine(&b, "item", "none")
	} else {
		for _, issue := range report.Summary.Issues {
			writeDoctorLine(&b, "item", fmt.Sprintf("%s: %s; next action: %s", issue.Section, issue.Summary, issue.Action))
		}
	}

	b.WriteString("\nRecommended Next Steps\n")
	if len(report.RecommendedFix) == 0 {
		writeDoctorLine(&b, "step", "none")
	} else {
		for _, action := range report.RecommendedFix {
			writeDoctorLine(&b, "step", action)
		}
	}

	return b.String()
}
