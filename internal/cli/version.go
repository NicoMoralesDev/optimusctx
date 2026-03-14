package cli

import (
	"fmt"
	"io"

	"github.com/niccrow/optimusctx/internal/buildinfo"
)

func newVersionCommand() *Command {
	return &Command{
		Name:    "version",
		Summary: "Print build and version metadata",
		Run: func(stdout io.Writer, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("version does not accept arguments")
			}

			_, err := fmt.Fprintln(stdout, buildinfo.Summary())
			return err
		},
	}
}
