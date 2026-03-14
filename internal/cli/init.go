package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/repository"
)

func newInitCommand() *Command {
	return &Command{
		Name:    "init",
		Summary: "Initialize repository-local OptimusCtx state",
		Run: func(stdout io.Writer, args []string) error {
			if len(args) > 0 {
				return errors.New("init does not accept arguments")
			}

			workingDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve working directory: %w", err)
			}

			result, err := app.NewInitService().Init(context.Background(), workingDir)
			if err != nil {
				if errors.Is(err, repository.ErrRepositoryNotFound) {
					return fmt.Errorf("no supported repository root found from %s; run `optimusctx init` inside a Git repository or an existing .optimusctx state directory", workingDir)
				}
				return err
			}

			_, err = fmt.Fprintf(
				stdout,
				"repository root: %s\nstate directory: %s\nschema version: %d\ndiscovered files: %d\n",
				result.RepositoryRoot,
				result.StatePath,
				result.SchemaVersion,
				result.FileCount,
			)
			return err
		},
	}
}
