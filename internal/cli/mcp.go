package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/niccrow/optimusctx/internal/mcp"
)

var (
	mcpServeInput  io.Reader = os.Stdin
	mcpServeStderr io.Writer = os.Stderr
	mcpServeServer           = func(ctx context.Context, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
		return mcp.ServeStdio(ctx, stdin, stdout, stderr)
	}
)

func newMCPCommand() *Command {
	return &Command{
		Name:    "mcp",
		Summary: "Serve the OptimusCtx MCP interface over STDIO",
		Run: func(stdout io.Writer, args []string) error {
			if len(args) == 0 {
				writeMCPHelp(stdout)
				return errors.New("mcp requires a subcommand")
			}

			switch args[0] {
			case "-h", "--help", "help":
				writeMCPHelp(stdout)
				return nil
			case "serve":
				return runMCPServeCommand(stdout, args[1:])
			default:
				writeMCPHelp(stdout)
				return fmt.Errorf("unknown mcp subcommand %q", args[0])
			}
		},
	}
}

func runMCPServeCommand(stdout io.Writer, args []string) error {
	for _, arg := range args {
		switch arg {
		case "-h", "--help":
			_, err := io.WriteString(stdout, "Usage:\n  optimusctx mcp serve\n\nServe the OptimusCtx MCP transport over STDIO.\n")
			return err
		default:
			if len(arg) > 0 && arg[0] == '-' {
				return fmt.Errorf("mcp serve does not accept flags; got %q", arg)
			}
			return fmt.Errorf("mcp serve does not accept arguments; got %q", arg)
		}
	}

	_, _ = io.WriteString(mcpServeStderr, "warning: `optimusctx mcp serve` is deprecated; use `optimusctx run` instead\n")
	return runCommandServer(context.Background(), mcpServeInput, stdout, mcpServeStderr)
}

func writeMCPHelp(stdout io.Writer) {
	_, _ = io.WriteString(stdout, "Usage:\n  optimusctx mcp <command>\n\nAvailable Commands:\n  serve   Serve the OptimusCtx MCP transport over STDIO\n")
}
