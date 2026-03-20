package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/niccrow/optimusctx/internal/repository"
)

func TestDoctorCommandWarnsAndDelegatesToStatus(t *testing.T) {
	previous := statusCommandService
	t.Cleanup(func() { statusCommandService = previous })

	statusCommandService = func(ctx context.Context, workingDir string) (repository.DoctorReport, error) {
		return repository.DoctorReport{
			Identity: repository.LayeredContextRepositoryIdentity{
				RootPath:      "/repo",
				DetectionMode: "git",
			},
			Install: repository.DoctorInstallSection{
				BinaryVersion: "dev",
				WorkingDir:    "/repo",
			},
			MCPReadiness: repository.DoctorMCPReadinessSection{
				Status:           repository.DoctorStatusHealthy,
				ServerName:       repository.DefaultMCPServerName,
				ServeCommand:     repository.NewServeCommand(""),
				SnippetAvailable: true,
			},
			Summary: repository.DoctorSummary{
				Status: repository.DoctorStatusHealthy,
			},
		}, nil
	}

	var stdout bytes.Buffer
	if err := newDoctorCommand().Run(&stdout, nil); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := stdout.String()
	for _, want := range []string{
		"warning: `optimusctx doctor` is deprecated; use `optimusctx status`",
		"overall status: healthy",
		"serve command: optimusctx run",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("doctor output missing %q:\n%s", want, output)
		}
	}
}
