package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/niccrow/optimusctx/internal/repository"
)

func TestPackExportCommand(t *testing.T) {
	repoRoot := initCLIRepo(t)

	t.Run("root help lists pack", func(t *testing.T) {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"--help"}, &stdout); err != nil {
			t.Fatalf("Execute(--help) error = %v", err)
		}
		assertContains(t, stdout.String(), "pack      Export deterministic repository packs for offline use")
	})

	t.Run("export delegates to service and writes file summary", func(t *testing.T) {
		previous := packExportCommandService
		t.Cleanup(func() {
			packExportCommandService = previous
		})

		called := false
		packExportCommandService = func(ctx context.Context, workingDir string, stdout io.Writer, request repository.PackExportRequest) (repository.PackExportResult, error) {
			called = true
			if workingDir != repoRoot {
				t.Fatalf("workingDir = %q, want %q", workingDir, repoRoot)
			}
			if request.OutputPath != "artifact.json.gz" {
				t.Fatalf("OutputPath = %q, want artifact.json.gz", request.OutputPath)
			}
			if request.Compression != repository.PackExportCompressionGzip {
				t.Fatalf("Compression = %q, want gzip", request.Compression)
			}
			if request.Format != repository.PackExportFormatJSON {
				t.Fatalf("Format = %q, want json", request.Format)
			}
			return repository.PackExportResult{
				Request: request,
				Artifact: repository.PackExportArtifact{
					Manifest: repository.PackExportManifest{
						ExportSummary: repository.PackExportSummary{
							IncludedSectionCount: 2,
						},
					},
				},
				Output: repository.PackExportOutput{
					Path:         request.OutputPath,
					Format:       request.Format,
					Compression:  request.Compression,
					BytesWritten: 512,
				},
			}, nil
		}

		withWorkingDirectory(t, repoRoot, func() {
			var stdout bytes.Buffer
			if err := NewRootCommand().Execute([]string{"pack", "export", "--output", "artifact.json.gz", "--gzip"}, &stdout); err != nil {
				t.Fatalf("Execute(pack export) error = %v", err)
			}
			if !called {
				t.Fatal("pack export service was not called")
			}
			output := stdout.String()
			assertContains(t, output, "pack export path: artifact.json.gz")
			assertContains(t, output, "compression: gzip")
			assertContains(t, output, "bytes written: 512")
		})
	})

	t.Run("export leaves stdout for artifact streaming", func(t *testing.T) {
		previous := packExportCommandService
		t.Cleanup(func() {
			packExportCommandService = previous
		})

		packExportCommandService = func(ctx context.Context, workingDir string, stdout io.Writer, request repository.PackExportRequest) (repository.PackExportResult, error) {
			_, _ = io.WriteString(stdout, "{\"manifest\":{}}\n")
			return repository.PackExportResult{Request: request}, nil
		}

		withWorkingDirectory(t, repoRoot, func() {
			var stdout bytes.Buffer
			if err := NewRootCommand().Execute([]string{"pack", "export"}, &stdout); err != nil {
				t.Fatalf("Execute(pack export) error = %v", err)
			}
			if got := stdout.String(); got != "{\"manifest\":{}}\n" {
				t.Fatalf("stdout = %q, want streamed artifact only", got)
			}
		})
	})
}

func TestPackExportCommandBudgetFlags(t *testing.T) {
	repoRoot := initCLIRepo(t)

	previous := packExportCommandService
	t.Cleanup(func() {
		packExportCommandService = previous
	})

	packExportCommandService = func(ctx context.Context, workingDir string, stdout io.Writer, request repository.PackExportRequest) (repository.PackExportResult, error) {
		if workingDir != repoRoot {
			t.Fatalf("workingDir = %q, want %q", workingDir, repoRoot)
		}
		if request.Policy.TargetTokenBudget != 256 {
			t.Fatalf("TargetTokenBudget = %d, want 256", request.Policy.TargetTokenBudget)
		}
		if len(request.Policy.IncludePaths) != 2 || request.Policy.IncludePaths[0] != "pkg" || request.Policy.IncludePaths[1] != "docs" {
			t.Fatalf("IncludePaths = %+v, want [pkg docs]", request.Policy.IncludePaths)
		}
		if len(request.Policy.ExcludePaths) != 1 || request.Policy.ExcludePaths[0] != "docs/generated" {
			t.Fatalf("ExcludePaths = %+v, want [docs/generated]", request.Policy.ExcludePaths)
		}
		return repository.PackExportResult{
			Request: request,
			Artifact: repository.PackExportArtifact{
				Manifest: repository.PackExportManifest{
					ExportSummary: repository.PackExportSummary{
						IncludedSectionCount: 1,
						EstimatedTokens:      200,
						TargetTokenBudget:    256,
						FitsTargetBudget:     true,
					},
				},
			},
			Output: repository.PackExportOutput{
				Path:         "artifact.json",
				Format:       repository.PackExportFormatJSON,
				Compression:  repository.PackExportCompressionNone,
				BytesWritten: 200,
			},
		}, nil
	}

	withWorkingDirectory(t, repoRoot, func() {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{
			"pack", "export",
			"--output", "artifact.json",
			"--include", "pkg",
			"--include", "docs",
			"--exclude", "docs/generated",
			"--target-budget", "256",
		}, &stdout); err != nil {
			t.Fatalf("Execute(pack export with budget flags) error = %v", err)
		}
		assertContains(t, stdout.String(), "pack export path: artifact.json")
	})
}

func TestPackExportCommandErrors(t *testing.T) {
	repoRoot := initCLIRepo(t)

	t.Run("pack requires subcommand", func(t *testing.T) {
		var stdout bytes.Buffer
		err := NewRootCommand().Execute([]string{"pack"}, &stdout)
		if err == nil || err.Error() != "pack requires a subcommand" {
			t.Fatalf("Execute(pack) error = %v", err)
		}
		assertContains(t, stdout.String(), "optimusctx pack <command>")
	})

	t.Run("export help renders usage", func(t *testing.T) {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"pack", "export", "--help"}, &stdout); err != nil {
			t.Fatalf("Execute(pack export --help) error = %v", err)
		}
		assertContains(t, stdout.String(), "optimusctx pack export [--output PATH] [--format json] [--gzip]")
	})

	t.Run("rejects unsupported format", func(t *testing.T) {
		var stdout bytes.Buffer
		err := NewRootCommand().Execute([]string{"pack", "export", "--format", "yaml"}, &stdout)
		if err == nil || err.Error() != "unsupported pack export format \"yaml\"" {
			t.Fatalf("Execute(pack export --format yaml) error = %v", err)
		}
	})

	t.Run("rejects missing output value", func(t *testing.T) {
		var stdout bytes.Buffer
		err := NewRootCommand().Execute([]string{"pack", "export", "--output"}, &stdout)
		if err == nil || err.Error() != "--output requires a value" {
			t.Fatalf("Execute(pack export --output) error = %v", err)
		}
	})

	t.Run("rejects unknown flag", func(t *testing.T) {
		var stdout bytes.Buffer
		err := NewRootCommand().Execute([]string{"pack", "export", "--verbose"}, &stdout)
		if err == nil || err.Error() != "unknown pack export flag \"--verbose\"" {
			t.Fatalf("Execute(pack export --verbose) error = %v", err)
		}
	})

	t.Run("returns service error", func(t *testing.T) {
		previous := packExportCommandService
		t.Cleanup(func() {
			packExportCommandService = previous
		})
		packExportCommandService = func(ctx context.Context, workingDir string, stdout io.Writer, request repository.PackExportRequest) (repository.PackExportResult, error) {
			return repository.PackExportResult{}, errors.New("boom")
		}

		withWorkingDirectory(t, repoRoot, func() {
			var stdout bytes.Buffer
			err := NewRootCommand().Execute([]string{"pack", "export"}, &stdout)
			if err == nil || err.Error() != "boom" {
				t.Fatalf("Execute(pack export) error = %v, want boom", err)
			}
		})
	})

	t.Run("rejects invalid target budget", func(t *testing.T) {
		var stdout bytes.Buffer
		err := NewRootCommand().Execute([]string{"pack", "export", "--target-budget", "zero"}, &stdout)
		if err == nil || err.Error() != "--target-budget requires a positive integer token budget" {
			t.Fatalf("Execute(pack export --target-budget zero) error = %v", err)
		}
	})
}

func TestFormatPackExportSummary(t *testing.T) {
	output := formatPackExportSummary(repository.PackExportResult{
		Artifact: repository.PackExportArtifact{
			Manifest: repository.PackExportManifest{
				ExportSummary: repository.PackExportSummary{
					IncludedSectionCount: 5,
					OmittedSectionCount:  1,
				},
			},
		},
		Output: repository.PackExportOutput{
			Path:         "artifact.json",
			Format:       repository.PackExportFormatJSON,
			Compression:  repository.PackExportCompressionNone,
			BytesWritten: 1234,
		},
	})
	for _, fragment := range []string{
		"pack export path: artifact.json",
		"format: json",
		"compression: none",
		"bytes written: 1234",
		"included sections: 5",
		"omitted sections: 1",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("output = %q, want fragment %q", output, fragment)
		}
	}
}
