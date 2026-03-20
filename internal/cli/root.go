package cli

import (
	"fmt"
	"io"
	"os"
)

type Command struct {
	Name        string
	Summary     string
	Description string
	Run         func(io.Writer, []string) error
}

func NewRootCommand() *Command {
	return &Command{
		Name:        "optimusctx",
		Summary:     "Local-first persistent repository context runtime",
		Description: "OptimusCtx builds and maintains repository-local context state for coding agents without rewriting project instruction files.",
	}
}

func (c *Command) Execute(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		c.printHelp(stdout)
		return nil
	}

	switch args[0] {
	case "-h", "--help", "help":
		c.printHelp(stdout)
		return nil
	case "doctor":
		return newDoctorCommand().Run(stdout, args[1:])
	case "init":
		return newInitCommand().Run(stdout, args[1:])
	case "release":
		return newReleaseCommand().Run(stdout, args[1:])
	case "run":
		return newRunCommand().Run(stdout, args[1:])
	case "status":
		return newStatusCommand().Run(stdout, args[1:])
	case "version":
		return newVersionCommand().Run(stdout, args[1:])
	default:
		c.printHelp(stdout)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func (c *Command) printHelp(stdout io.Writer) {
	_, _ = fmt.Fprintf(stdout, "%s\n\n%s\n\nUsage:\n  %s <command>\n\nAvailable Commands:\n  doctor    %s\n  init      %s\n  release   %s\n  run       %s\n  status    %s\n  version   %s\n\nFlags:\n  -h, --help   Show help for optimusctx\n",
		c.Name,
		c.Description,
		c.Name,
		newDoctorCommand().Summary,
		newInitCommand().Summary,
		newReleaseCommand().Summary,
		newRunCommand().Summary,
		newStatusCommand().Summary,
		newVersionCommand().Summary,
	)
}

func newSnippetCommand() *Command {
	return newSnippetCommandImpl()
}

func newDoctorCommand() *Command {
	return newDoctorCommandImpl()
}

func Execute() int {
	if err := NewRootCommand().Execute(os.Args[1:], os.Stdout); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}
