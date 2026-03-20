package cli

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
)

func newDoctorCommandImpl() *Command {
	return &Command{
		Name:    "doctor",
		Summary: "Deprecated alias for `optimusctx status`",
		Run: func(stdout io.Writer, args []string) error {
			_, _ = io.WriteString(stdout, "warning: `optimusctx doctor` is deprecated; use `optimusctx status`\n\n")
			return newStatusCommand().Run(stdout, args)
		},
	}
}

func writeDoctorLine(b *strings.Builder, label string, value string) {
	_, _ = fmt.Fprintf(b, "%s: %s\n", label, renderDoctorValue(value))
}

func renderDoctorValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "n/a"
	}
	return value
}

func formatDoctorBool(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func formatDoctorTime(value time.Time) string {
	if value.IsZero() {
		return "n/a"
	}
	return value.UTC().Format(time.RFC3339)
}

func formatDoctorInt(value int) string {
	return fmt.Sprintf("%d", value)
}

func formatDoctorInt64(value int64) string {
	return fmt.Sprintf("%d", value)
}

func doctorWatchReason(report repository.DoctorReport) string {
	if report.Watch.Optional && report.Watch.Health.Status == repository.WatchStatusKindAbsent {
		return "watch status file not found because `optimusctx run` is not active"
	}
	return report.Watch.Health.Reason
}
