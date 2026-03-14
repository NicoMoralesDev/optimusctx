package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/repository"
)

var refreshCommandService = func(ctx context.Context, workingDir string) (app.RefreshResult, error) {
	return app.NewRefreshService().Refresh(ctx, app.RefreshRequest{
		StartPath: workingDir,
		Reason:    repository.RefreshReasonManual,
	})
}

func newRefreshCommand() *Command {
	return &Command{
		Name:    "refresh",
		Summary: "Refresh repository-local OptimusCtx state",
		Run: func(stdout io.Writer, args []string) error {
			if len(args) > 0 {
				return errors.New("refresh does not accept arguments")
			}

			workingDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve working directory: %w", err)
			}

			result, err := refreshCommandService(context.Background(), workingDir)
			if err != nil {
				if errors.Is(err, repository.ErrRepositoryNotFound) || strings.Contains(err.Error(), repository.ErrRepositoryNotFound.Error()) {
					return fmt.Errorf("no supported repository root found from %s; run `optimusctx refresh` inside an initialized Git repository or existing .optimusctx state directory", workingDir)
				}
				return err
			}

			_, err = io.WriteString(stdout, formatRefreshSummary(result))
			return err
		},
	}
}

func formatRefreshSummary(result app.RefreshResult) string {
	return fmt.Sprintf(
		"repository root: %s\nrefresh generation: %d\nfreshness: %s\nadded files: %d\nchanged files: %d\ndeleted files: %d\nmoved files: %d\nnewly ignored files: %d\nunchanged files: %d\n",
		result.RepositoryRoot,
		result.Generation,
		renderFreshnessStatus(result.FreshnessStatus),
		result.AddedFiles+result.ReincludedFiles,
		result.ChangedContentFiles,
		result.DeletedFiles,
		result.MovedFiles,
		result.NewlyIgnoredFiles,
		result.UnchangedFiles,
	)
}

func renderFreshnessStatus(status repository.FreshnessStatus) string {
	if status == repository.FreshnessStatusPartiallyDegraded {
		return "partially degraded"
	}
	return string(status)
}
