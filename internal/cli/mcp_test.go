package cli

import (
	"bytes"
	"context"
	"io"
	"testing"
)

func TestMCPServeCommand(t *testing.T) {
	previousServe := mcpServeServer
	previousInput := mcpServeInput
	previousStderr := mcpServeStderr
	t.Cleanup(func() {
		mcpServeServer = previousServe
		mcpServeInput = previousInput
		mcpServeStderr = previousStderr
	})

	var called bool
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	mcpServeInput = bytes.NewBufferString("")
	mcpServeStderr = &stderr
	mcpServeServer = func(ctx context.Context, stdin io.Reader, serveStdout io.Writer, serveStderr io.Writer) error {
		called = true
		if stdin != mcpServeInput {
			t.Fatal("mcp serve stdin did not use configured input")
		}
		if serveStdout != &stdout {
			t.Fatal("mcp serve stdout did not use command stdout")
		}
		if serveStderr != &stderr {
			t.Fatal("mcp serve stderr did not use configured stderr")
		}
		if _, err := io.WriteString(serveStderr, "optimusctx mcp: ready for stdio requests\n"); err != nil {
			t.Fatalf("WriteString(serve stderr) error = %v", err)
		}
		return nil
	}

	if err := NewRootCommand().Execute([]string{"mcp", "serve"}, &stdout); err != nil {
		t.Fatalf("Execute(mcp serve) error = %v", err)
	}
	if !called {
		t.Fatal("mcp serve server was not called")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.String() != "optimusctx mcp: ready for stdio requests\n" {
		t.Fatalf("stderr = %q, want readiness signal", stderr.String())
	}
}

func TestMCPServeCommandRejectsUnsupportedArguments(t *testing.T) {
	var stdout bytes.Buffer

	err := NewRootCommand().Execute([]string{"mcp", "serve", "--verbose"}, &stdout)
	if err == nil || err.Error() != "mcp serve does not accept flags; got \"--verbose\"" {
		t.Fatalf("Execute(mcp serve --verbose) error = %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestMCPServeReadinessSignalUsesStderr(t *testing.T) {
	previousServe := mcpServeServer
	previousInput := mcpServeInput
	previousStderr := mcpServeStderr
	t.Cleanup(func() {
		mcpServeServer = previousServe
		mcpServeInput = previousInput
		mcpServeStderr = previousStderr
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	mcpServeInput = bytes.NewBufferString("")
	mcpServeStderr = &stderr
	mcpServeServer = func(ctx context.Context, stdin io.Reader, serveStdout io.Writer, serveStderr io.Writer) error {
		if _, err := io.WriteString(serveStderr, "optimusctx mcp: ready for stdio requests\n"); err != nil {
			t.Fatalf("WriteString(serve stderr) error = %v", err)
		}
		return nil
	}

	if err := NewRootCommand().Execute([]string{"mcp", "serve"}, &stdout); err != nil {
		t.Fatalf("Execute(mcp serve) error = %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.String() != "optimusctx mcp: ready for stdio requests\n" {
		t.Fatalf("stderr = %q, want readiness signal", stderr.String())
	}
}
