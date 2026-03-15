package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/buildinfo"
	"github.com/niccrow/optimusctx/internal/repository"
)

var packExportCommandService = func(ctx context.Context, workingDir string, stdout io.Writer, request repository.PackExportRequest) (repository.PackExportResult, error) {
	return app.NewPackExportService().Write(ctx, workingDir, request, stdout)
}

func newPackCommand() *Command {
	return &Command{
		Name:    "pack",
		Summary: "Export deterministic repository packs for offline use",
		Run: func(stdout io.Writer, args []string) error {
			if len(args) == 0 {
				writePackHelp(stdout)
				return errors.New("pack requires a subcommand")
			}

			switch args[0] {
			case "-h", "--help", "help":
				writePackHelp(stdout)
				return nil
			case "export":
				return runPackExportCommand(stdout, args[1:])
			default:
				writePackHelp(stdout)
				return fmt.Errorf("unknown pack subcommand %q", args[0])
			}
		},
	}
}

func runPackExportCommand(stdout io.Writer, args []string) error {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return writePackExportHelp(stdout)
		}
	}

	request, err := parsePackExportArgs(args)
	if err != nil {
		return err
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}

	result, err := packExportCommandService(context.Background(), workingDir, stdout, request)
	if err != nil {
		return err
	}
	if result.Request.OutputPath == "" {
		return nil
	}

	_, err = io.WriteString(stdout, formatPackExportSummary(result))
	return err
}

func parsePackExportArgs(args []string) (repository.PackExportRequest, error) {
	request := repository.PackExportRequest{
		PackRequest: repository.PackRequest{
			IncludeRepositoryContext: true,
			IncludeStructuralContext: true,
		},
		Format:      repository.PackExportFormatJSON,
		Compression: repository.PackExportCompressionNone,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Generator:   buildinfo.Summary(),
	}

	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "-o", "--output":
			index++
			if index >= len(args) {
				return repository.PackExportRequest{}, fmt.Errorf("%s requires a value", arg)
			}
			request.OutputPath = args[index]
		case "--format":
			index++
			if index >= len(args) {
				return repository.PackExportRequest{}, fmt.Errorf("%s requires a value", arg)
			}
			request.Format = repository.PackExportFormat(args[index])
		case "--gzip":
			request.Compression = repository.PackExportCompressionGzip
		default:
			if len(arg) > 0 && arg[0] == '-' {
				return repository.PackExportRequest{}, fmt.Errorf("unknown pack export flag %q", arg)
			}
			return repository.PackExportRequest{}, fmt.Errorf("pack export does not accept arguments; got %q", arg)
		}
	}

	if request.Format != repository.PackExportFormatJSON {
		return repository.PackExportRequest{}, fmt.Errorf("unsupported pack export format %q", request.Format)
	}

	return request, nil
}

func writePackHelp(stdout io.Writer) {
	_, _ = io.WriteString(stdout, "Usage:\n  optimusctx pack <command>\n\nAvailable Commands:\n  export   Export a deterministic repository pack\n")
}

func writePackExportHelp(stdout io.Writer) error {
	_, err := io.WriteString(stdout, "Usage:\n  optimusctx pack export [--output PATH] [--format json] [--gzip]\n\nExport a deterministic repository pack to stdout or a file.\n")
	return err
}

func formatPackExportSummary(result repository.PackExportResult) string {
	return fmt.Sprintf(
		"pack export path: %s\nformat: %s\ncompression: %s\nbytes written: %d\nincluded sections: %d\nomitted sections: %d\n",
		result.Output.Path,
		result.Output.Format,
		result.Output.Compression,
		result.Output.BytesWritten,
		result.Artifact.Manifest.ExportSummary.IncludedSectionCount,
		result.Artifact.Manifest.ExportSummary.OmittedSectionCount,
	)
}
