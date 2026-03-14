package cli

import (
	"errors"
	"io"

	"github.com/niccrow/optimusctx/internal/app"
)

func newSnippetCommandImpl() *Command {
	return &Command{
		Name:    "snippet",
		Summary: "Print the manual integration snippet",
		Run: func(stdout io.Writer, args []string) error {
			if len(args) > 0 {
				return errors.New("snippet does not accept arguments")
			}

			_, err := io.WriteString(stdout, app.NewSnippetGenerator().Render())
			return err
		},
	}
}
