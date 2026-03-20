package cli

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/niccrow/optimusctx/internal/app"
	"github.com/niccrow/optimusctx/internal/repository"
)

var initPromptInput io.Reader = os.Stdin

var initShouldPrompt = func(stdout io.Writer) bool {
	return fileIsInteractive(os.Stdin) && writerIsInteractive(stdout)
}

func fileIsInteractive(file *os.File) bool {
	if file == nil {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func writerIsInteractive(writer io.Writer) bool {
	file, ok := writer.(*os.File)
	if !ok {
		return false
	}
	return fileIsInteractive(file)
}

func promptInitOnboarding(stdin io.Reader, stdout io.Writer) (app.InstallRequest, bool, error) {
	reader := bufio.NewReader(stdin)

	if _, err := io.WriteString(stdout, "\nset up a supported MCP client now?\n  1. Claude Desktop\n  2. Claude CLI\n  3. Codex App\n  4. Codex CLI\nChoose [1-4, Enter to skip]: "); err != nil {
		return app.InstallRequest{}, false, err
	}

	clientChoice, skipped, err := readPromptChoice(reader, stdout, map[string]string{
		"1":              string(repository.ClientClaudeDesktop),
		"claude-desktop": string(repository.ClientClaudeDesktop),
		"claude desktop": string(repository.ClientClaudeDesktop),
		"2":              string(repository.ClientClaudeCLI),
		"claude-cli":     string(repository.ClientClaudeCLI),
		"claude cli":     string(repository.ClientClaudeCLI),
		"3":              string(repository.ClientCodexApp),
		"codex-app":      string(repository.ClientCodexApp),
		"codex app":      string(repository.ClientCodexApp),
		"4":              string(repository.ClientCodexCLI),
		"codex-cli":      string(repository.ClientCodexCLI),
		"codex cli":      string(repository.ClientCodexCLI),
	}, "")
	if err != nil {
		return app.InstallRequest{}, false, err
	}
	if skipped {
		return app.InstallRequest{}, true, nil
	}

	request := app.InstallRequest{ClientID: clientChoice}
	if request.ClientID == string(repository.ClientClaudeCLI) {
		if _, err := io.WriteString(stdout, "Claude CLI scope:\n  1. local\n  2. project\n  3. user\nChoose [1-3, default: 1]: "); err != nil {
			return app.InstallRequest{}, false, err
		}
		scope, _, err := readPromptChoice(reader, stdout, map[string]string{
			"1":       repository.ClaudeCLIScopeLocal,
			"local":   repository.ClaudeCLIScopeLocal,
			"2":       repository.ClaudeCLIScopeProject,
			"project": repository.ClaudeCLIScopeProject,
			"3":       repository.ClaudeCLIScopeUser,
			"user":    repository.ClaudeCLIScopeUser,
		}, repository.ClaudeCLIScopeLocal)
		if err != nil {
			return app.InstallRequest{}, false, err
		}
		request.Scope = scope
	}

	if _, err := io.WriteString(stdout, "How should OptimusCtx continue?\n  1. Preview only\n  2. Write config / register now\nChoose [1-2, default: 1]: "); err != nil {
		return app.InstallRequest{}, false, err
	}
	mode, _, err := readPromptChoice(reader, stdout, map[string]string{
		"1":       "preview",
		"preview": "preview",
		"p":       "preview",
		"2":       "write",
		"write":   "write",
		"w":       "write",
		"y":       "write",
	}, "preview")
	if err != nil {
		return app.InstallRequest{}, false, err
	}
	request.Write = mode == "write"

	return request, false, nil
}

func readPromptChoice(reader *bufio.Reader, stdout io.Writer, choices map[string]string, defaultValue string) (string, bool, error) {
	for {
		line, err := reader.ReadString('\n')
		if err != nil && len(line) == 0 {
			return "", false, err
		}
		value := strings.TrimSpace(strings.ToLower(line))
		if value == "" {
			if defaultValue != "" {
				return defaultValue, false, nil
			}
			return "", true, nil
		}
		if value == "skip" || value == "s" {
			return "", true, nil
		}
		if resolved, ok := choices[value]; ok {
			return resolved, false, nil
		}
		if _, err := io.WriteString(stdout, "Please choose one of the listed options.\n> "); err != nil {
			return "", false, err
		}
	}
}
