package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

func promptInitOnboarding(stdin io.Reader, stdout io.Writer, repoRoot string) (app.InstallRequest, bool, error) {
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

	request := app.InstallRequest{ClientID: clientChoice, RepoRoot: repoRoot}
	switch request.ClientID {
	case string(repository.ClientClaudeDesktop):
		configPath, err := app.DefaultClaudeDesktopConfigPath()
		if err != nil {
			return app.InstallRequest{}, false, err
		}
		if _, err := fmt.Fprintf(stdout, "Where should Claude Desktop load OptimusCtx from?\n  1. Claude Desktop app config\n     %s\nChoose [1, default: 1]: ", configPath); err != nil {
			return app.InstallRequest{}, false, err
		}
		if _, _, err := readPromptChoice(reader, stdout, map[string]string{
			"1": "desktop",
		}, "desktop"); err != nil {
			return app.InstallRequest{}, false, err
		}
	case string(repository.ClientClaudeCLI):
		if _, err := io.WriteString(stdout, "Where should Claude CLI register OptimusCtx?\n  1. This project\n     Native target: claude mcp add --scope project\n  2. Your current Claude setup\n     Native target: claude mcp add --scope local\n  3. Your Claude user profile\n     Native target: claude mcp add --scope user\nChoose [1-3, default: 1]: "); err != nil {
			return app.InstallRequest{}, false, err
		}
		scope, _, err := readPromptChoice(reader, stdout, map[string]string{
			"1": repository.ClaudeCLIScopeProject,
			"2": repository.ClaudeCLIScopeLocal,
			"3": repository.ClaudeCLIScopeUser,
		}, repository.ClaudeCLIScopeProject)
		if err != nil {
			return app.InstallRequest{}, false, err
		}
		request.Scope = scope
	case string(repository.ClientCodexApp), string(repository.ClientCodexCLI):
		sharedConfigPath, err := app.DefaultCodexConfigPath()
		if err != nil {
			return app.InstallRequest{}, false, err
		}
		repoConfigPath := filepath.Join(repoRoot, ".codex", "config.toml")
		clientLabel := "Codex App"
		if request.ClientID == string(repository.ClientCodexCLI) {
			clientLabel = "Codex CLI"
		}
		if _, err := fmt.Fprintf(stdout, "Where should %s use OptimusCtx?\n  1. This repo only\n     %s\n  2. Your shared Codex config\n     %s\nChoose [1-2, default: 1]: ", clientLabel, repoConfigPath, sharedConfigPath); err != nil {
			return app.InstallRequest{}, false, err
		}
		destination, _, err := readPromptChoice(reader, stdout, map[string]string{
			"1": "repo",
			"2": "shared",
		}, "repo")
		if err != nil {
			return app.InstallRequest{}, false, err
		}
		if destination == "repo" {
			request.ConfigPath = repoConfigPath
		}
	}

	if _, err := io.WriteString(stdout, "How should OptimusCtx continue?\n  1. Configure it now\n  2. Review the exact change first\nChoose [1-2, default: 2]: "); err != nil {
		return app.InstallRequest{}, false, err
	}
	mode, _, err := readPromptChoice(reader, stdout, map[string]string{
		"1":         "write",
		"configure": "write",
		"now":       "write",
		"y":         "write",
		"2":         "review",
		"review":    "review",
		"preview":   "review",
		"p":         "review",
	}, "review")
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
