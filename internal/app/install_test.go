package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niccrow/optimusctx/internal/repository"
)

func TestResolveClaudeDesktopConfigPathExplicitOverride(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	got, err := resolveClaudeDesktopConfigPathForPlatform("windows", "/ignored/home", "", "/tmp/custom/claude.json")
	if err != nil {
		t.Fatalf("resolveClaudeDesktopConfigPathForPlatform() error = %v", err)
	}
	if got != "/tmp/custom/claude.json" {
		t.Fatalf("config path = %q", got)
	}
}

func TestResolveClaudeDesktopConfigPathLinux(t *testing.T) {
	got, err := resolveClaudeDesktopConfigPathForPlatform("linux", "/home/tester", "", "")
	if err != nil {
		t.Fatalf("resolveClaudeDesktopConfigPathForPlatform() error = %v", err)
	}

	want := filepath.Join("/home/tester", ".config", "Claude", "claude_desktop_config.json")
	if got != want {
		t.Fatalf("config path = %q, want %q", got, want)
	}
}

func TestResolveClaudeDesktopConfigPathDarwin(t *testing.T) {
	got, err := resolveClaudeDesktopConfigPathForPlatform("darwin", "/Users/tester", "", "")
	if err != nil {
		t.Fatalf("resolveClaudeDesktopConfigPathForPlatform() error = %v", err)
	}

	want := filepath.Join("/Users/tester", "Library", "Application Support", "Claude", "claude_desktop_config.json")
	if got != want {
		t.Fatalf("config path = %q, want %q", got, want)
	}
}

func TestResolveClaudeDesktopConfigPathWindowsRequiresAppData(t *testing.T) {
	_, err := resolveClaudeDesktopConfigPathForPlatform("windows", `C:\Users\tester`, "", "")
	if err == nil || !strings.Contains(err.Error(), "%AppData% is not set") {
		t.Fatalf("resolveClaudeDesktopConfigPathForPlatform() error = %v", err)
	}
}

func TestResolveClaudeDesktopConfigPathWSLUsesWindowsAppData(t *testing.T) {
	previousGetenv := codexConfigGetenv
	previousReadFile := codexConfigReadFile
	previousGOOS := codexConfigGOOS
	t.Cleanup(func() {
		codexConfigGetenv = previousGetenv
		codexConfigReadFile = previousReadFile
		codexConfigGOOS = previousGOOS
	})
	codexConfigGOOS = "linux"
	codexConfigGetenv = func(key string) string {
		switch key {
		case "WSL_DISTRO_NAME":
			return "Ubuntu-22.04"
		case "APPDATA":
			return `C:\Users\nicle\AppData\Roaming`
		default:
			return ""
		}
	}
	codexConfigReadFile = func(string) ([]byte, error) { return nil, os.ErrNotExist }

	got, err := resolveClaudeDesktopConfigPath("")
	if err != nil {
		t.Fatalf("resolveClaudeDesktopConfigPath() error = %v", err)
	}
	if want := "/mnt/c/Users/nicle/AppData/Roaming/Claude/claude_desktop_config.json"; got != want {
		t.Fatalf("config path = %q, want %q", got, want)
	}
}

func TestResolveClaudeDesktopConfigPathWSLRequiresExplicitPathWhenWindowsProfileUnknown(t *testing.T) {
	previousGetenv := codexConfigGetenv
	previousReadFile := codexConfigReadFile
	previousGOOS := codexConfigGOOS
	t.Cleanup(func() {
		codexConfigGetenv = previousGetenv
		codexConfigReadFile = previousReadFile
		codexConfigGOOS = previousGOOS
	})
	codexConfigGOOS = "linux"
	codexConfigGetenv = func(key string) string {
		if key == "WSL_DISTRO_NAME" {
			return "Ubuntu-22.04"
		}
		return ""
	}
	codexConfigReadFile = func(string) ([]byte, error) { return nil, os.ErrNotExist }

	_, err := resolveClaudeDesktopConfigPath("")
	if err == nil || !strings.Contains(err.Error(), "pass --config /mnt/c/Users/<user>/AppData/Roaming/Claude/claude_desktop_config.json") {
		t.Fatalf("resolveClaudeDesktopConfigPath() error = %v", err)
	}
}

func TestResolveCodexCLIConfigPathUsesHomeDir(t *testing.T) {
	previousHome := codexConfigUserHomeDir
	previousGOOS := codexConfigGOOS
	t.Cleanup(func() {
		codexConfigUserHomeDir = previousHome
		codexConfigGOOS = previousGOOS
	})
	codexConfigUserHomeDir = func() (string, error) { return "/home/tester", nil }
	codexConfigGOOS = "linux"

	got, err := resolveCodexCLIConfigPath("")
	if err != nil {
		t.Fatalf("resolveCodexCLIConfigPath() error = %v", err)
	}
	if want := "/home/tester/.codex/config.toml"; got != want {
		t.Fatalf("config path = %q, want %q", got, want)
	}
}

func TestResolveCodexAppConfigPathWSLUsesWindowsUserProfile(t *testing.T) {
	previousGetenv := codexConfigGetenv
	previousReadFile := codexConfigReadFile
	previousGOOS := codexConfigGOOS
	t.Cleanup(func() {
		codexConfigGetenv = previousGetenv
		codexConfigReadFile = previousReadFile
		codexConfigGOOS = previousGOOS
	})
	codexConfigGOOS = "linux"
	codexConfigGetenv = func(key string) string {
		switch key {
		case "WSL_DISTRO_NAME":
			return "Ubuntu-22.04"
		case "USERPROFILE":
			return `C:\Users\nicle`
		default:
			return ""
		}
	}
	codexConfigReadFile = func(string) ([]byte, error) { return nil, os.ErrNotExist }

	got, err := resolveCodexAppConfigPath("")
	if err != nil {
		t.Fatalf("resolveCodexAppConfigPath() error = %v", err)
	}
	if want := "/mnt/c/Users/nicle/.codex/config.toml"; got != want {
		t.Fatalf("config path = %q, want %q", got, want)
	}
}

func TestResolveCodexAppConfigPathWSLRequiresExplicitPathWhenWindowsProfileUnknown(t *testing.T) {
	previousGetenv := codexConfigGetenv
	previousReadFile := codexConfigReadFile
	previousGOOS := codexConfigGOOS
	t.Cleanup(func() {
		codexConfigGetenv = previousGetenv
		codexConfigReadFile = previousReadFile
		codexConfigGOOS = previousGOOS
	})
	codexConfigGOOS = "linux"
	codexConfigGetenv = func(key string) string {
		if key == "WSL_DISTRO_NAME" {
			return "Ubuntu-22.04"
		}
		return ""
	}
	codexConfigReadFile = func(string) ([]byte, error) { return nil, os.ErrNotExist }

	_, err := resolveCodexAppConfigPath("")
	if err == nil || !strings.Contains(err.Error(), "pass --config /mnt/c/Users/<user>/.codex/config.toml") {
		t.Fatalf("resolveCodexAppConfigPath() error = %v", err)
	}
}

func TestInstallServiceSupportsGenericPreview(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("OPTIMUSCTX_CODEX_APP_CONFIG", filepath.Join(homeDir, ".codex", "config.toml"))

	service := NewInstallService()
	result, err := service.Register(context.Background(), InstallRequest{ClientID: "generic"})
	if err != nil {
		t.Fatalf("Register(generic) error = %v", err)
	}
	if result.Wrote {
		t.Fatal("generic preview should not write")
	}
	if result.Rendered.Client.ID != "generic" {
		t.Fatalf("client id = %q", result.Rendered.Client.ID)
	}
	if result.Rendered.ConfigPath != "manual" {
		t.Fatalf("config path = %q", result.Rendered.ConfigPath)
	}
	if !strings.Contains(result.Rendered.Content, "\"command\": \"optimusctx\"") || !strings.Contains(result.Rendered.Content, "\"run\"") {
		t.Fatalf("content missing run command: %s", result.Rendered.Content)
	}
	if len(result.Rendered.Notes) == 0 {
		t.Fatal("generic preview should include notes")
	}

	for _, clientID := range []string{"claude-cli", "codex-app", "codex-cli", "gemini-cli", "cursor-cli"} {
		namedResult, err := service.Register(context.Background(), InstallRequest{ClientID: clientID})
		if err != nil {
			t.Fatalf("Register(%s) error = %v", clientID, err)
		}
		if namedResult.Rendered.ConfigPath == "manual" {
			t.Fatalf("%s should not reuse the manual generic fallback", clientID)
		}
	}
}

func TestInstallServiceSupportsClaudeCLIPreview(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	service := NewInstallService()
	result, err := service.Register(context.Background(), InstallRequest{ClientID: "claude-cli"})
	if err != nil {
		t.Fatalf("Register(claude-cli) error = %v", err)
	}
	if result.Wrote {
		t.Fatal("claude-cli preview should not write")
	}
	if result.Rendered.Client.ID != "claude-cli" {
		t.Fatalf("client id = %q", result.Rendered.Client.ID)
	}
	if result.Rendered.ConfigPath != "command" {
		t.Fatalf("config path = %q", result.Rendered.ConfigPath)
	}
	if result.Rendered.Mode != "preview" {
		t.Fatalf("mode = %q", result.Rendered.Mode)
	}
	if result.Rendered.Content != "claude mcp add --transport stdio --scope local optimusctx -- optimusctx run" {
		t.Fatalf("content = %q", result.Rendered.Content)
	}
	if len(result.Rendered.Notes) == 0 {
		t.Fatal("claude-cli preview should include notes")
	}
}

func TestInstallServiceClaudeCLIPreviewUsesRequestedScope(t *testing.T) {
	service := NewInstallService()

	result, err := service.Register(context.Background(), InstallRequest{
		ClientID:   "claude-cli",
		BinaryPath: "/tmp/optimusctx",
		Scope:      "project",
	})
	if err != nil {
		t.Fatalf("Register(claude-cli preview) error = %v", err)
	}

	if result.Rendered.Content != "claude mcp add --transport stdio --scope project optimusctx -- /tmp/optimusctx run" {
		t.Fatalf("content = %q", result.Rendered.Content)
	}
	if result.Rendered.ConfigPath != "command" {
		t.Fatalf("config path = %q", result.Rendered.ConfigPath)
	}
}

func TestInstallServiceClaudeCLIWriteExecutesCommand(t *testing.T) {
	var gotName string
	var gotArgs []string
	var wroteGuidance bool

	service := InstallService{
		adapters: map[repository.ClientID]clientRegistrationAdapter{
			repository.ClientClaudeCLI: claudeCLIClientAdapter{
				client: repository.SupportedClient{ID: repository.ClientClaudeCLI, DisplayName: "Claude CLI"},
				notes:  claudeCLINotes(),
				runCommand: func(_ context.Context, name string, args ...string) ([]byte, error) {
					gotName = name
					gotArgs = append([]string(nil), args...)
					return []byte("registered"), nil
				},
			},
		},
		readFile: func(string) ([]byte, error) {
			return nil, os.ErrNotExist
		},
		writeFile: func(path string, _ []byte, _ os.FileMode) error {
			if strings.Contains(path, repository.ClaudeRulesFilename) {
				wroteGuidance = true
			}
			return nil
		},
		mkdirAll: func(string, os.FileMode) error { return nil },
	}

	result, err := service.Register(context.Background(), InstallRequest{ClientID: "claude-cli", Write: true})
	if err != nil {
		t.Fatalf("Register(claude-cli write) error = %v", err)
	}
	if !result.Wrote {
		t.Fatal("write should report wrote=true")
	}
	if gotName != "claude" {
		t.Fatalf("command name = %q", gotName)
	}
	wantArgs := []string{"mcp", "add", "--transport", "stdio", "--scope", "local", "optimusctx", "--", "optimusctx", "run"}
	if len(gotArgs) != len(wantArgs) {
		t.Fatalf("args length = %d, want %d (%v)", len(gotArgs), len(wantArgs), gotArgs)
	}
	for i, want := range wantArgs {
		if gotArgs[i] != want {
			t.Fatalf("arg %d = %q, want %q (args=%v)", i, gotArgs[i], want, gotArgs)
		}
	}
	if result.Rendered.ConfigPath != "command" {
		t.Fatalf("config path = %q", result.Rendered.ConfigPath)
	}
	if result.Rendered.Mode != repository.RenderModeWrite {
		t.Fatalf("mode = %q", result.Rendered.Mode)
	}
	if result.Rendered.Content != "claude mcp add --transport stdio --scope local optimusctx -- optimusctx run" {
		t.Fatalf("content = %q", result.Rendered.Content)
	}
	if !wroteGuidance {
		t.Fatal("expected Claude CLI write to persist guidance")
	}
}

func TestInstallServiceClaudeCLIRejectsUnsupportedScope(t *testing.T) {
	called := false
	service := InstallService{
		adapters: map[repository.ClientID]clientRegistrationAdapter{
			repository.ClientClaudeCLI: claudeCLIClientAdapter{
				client: repository.SupportedClient{ID: repository.ClientClaudeCLI, DisplayName: "Claude CLI"},
				runCommand: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
					called = true
					return nil, nil
				},
			},
		},
	}

	_, err := service.Register(context.Background(), InstallRequest{ClientID: "claude-cli", Scope: "workspace", Write: true})
	if err == nil || err.Error() != `unsupported Claude CLI scope "workspace"; expected local, project, or user` {
		t.Fatalf("Register(claude-cli invalid scope) error = %v", err)
	}
	if called {
		t.Fatal("runCommand should not execute for unsupported scope")
	}
}

func TestInstallServiceClaudeCLIWriteReportsMissingClaudeBinary(t *testing.T) {
	service := InstallService{
		adapters: map[repository.ClientID]clientRegistrationAdapter{
			repository.ClientClaudeCLI: claudeCLIClientAdapter{
				client: repository.SupportedClient{ID: repository.ClientClaudeCLI, DisplayName: "Claude CLI"},
				runCommand: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
					return nil, exec.ErrNotFound
				},
			},
		},
	}

	_, err := service.Register(context.Background(), InstallRequest{ClientID: "claude-cli", Write: true})
	const want = "run Claude CLI registration: claude command not found; install Claude Code or rerun without --write to preview the command"
	if err == nil || err.Error() != want {
		t.Fatalf("Register(claude-cli missing binary) error = %v, want %q", err, want)
	}
}

func TestInstallServiceClaudeCLIWriteReportsCommandFailure(t *testing.T) {
	service := InstallService{
		adapters: map[repository.ClientID]clientRegistrationAdapter{
			repository.ClientClaudeCLI: claudeCLIClientAdapter{
				client: repository.SupportedClient{ID: repository.ClientClaudeCLI, DisplayName: "Claude CLI"},
				runCommand: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
					return []byte("permission denied\n"), errors.New("exit status 1")
				},
			},
		},
	}

	_, err := service.Register(context.Background(), InstallRequest{ClientID: "claude-cli", Write: true})
	if err == nil || err.Error() != "run Claude CLI registration: permission denied" {
		t.Fatalf("Register(claude-cli command failure) error = %v", err)
	}
}

func TestInstallServiceSupportsCodexAppPreview(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("OPTIMUSCTX_CODEX_APP_CONFIG", filepath.Join(homeDir, ".codex", "config.toml"))

	service := NewInstallService()
	result, err := service.Register(context.Background(), InstallRequest{ClientID: "codex-app"})
	if err != nil {
		t.Fatalf("Register(codex-app) error = %v", err)
	}
	if result.Wrote {
		t.Fatal("codex-app preview should not write")
	}
	if result.Rendered.Client.ID != "codex-app" {
		t.Fatalf("client id = %q", result.Rendered.Client.ID)
	}
	if result.Rendered.Client.DisplayName != "Codex App" {
		t.Fatalf("display name = %q", result.Rendered.Client.DisplayName)
	}
	if got, want := result.Rendered.ConfigPath, filepath.Join(homeDir, ".codex", "config.toml"); got != want {
		t.Fatalf("config path = %q, want %q", got, want)
	}
	if result.Rendered.Mode != "preview" {
		t.Fatalf("mode = %q", result.Rendered.Mode)
	}
	if !strings.Contains(result.Rendered.Content, "[mcp_servers.optimusctx]") {
		t.Fatalf("content missing optimusctx table: %s", result.Rendered.Content)
	}
	if !strings.Contains(result.Rendered.Content, `command = "optimusctx"`) {
		t.Fatalf("content missing command line: %s", result.Rendered.Content)
	}
	if strings.TrimSpace(result.Rendered.AppliedContent) == "" {
		t.Fatal("codex-app preview should keep applied content for real writes")
	}
	if len(result.Rendered.Notes) == 0 {
		t.Fatal("codex-app preview should include notes")
	}
}

func TestInstallServiceSupportsCodexCLIPreview(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	service := NewInstallService()
	result, err := service.Register(context.Background(), InstallRequest{ClientID: "codex-cli"})
	if err != nil {
		t.Fatalf("Register(codex-cli) error = %v", err)
	}
	if result.Wrote {
		t.Fatal("codex-cli preview should not write")
	}
	if result.Rendered.Client.ID != "codex-cli" {
		t.Fatalf("client id = %q", result.Rendered.Client.ID)
	}
	if result.Rendered.Client.DisplayName != "Codex CLI" {
		t.Fatalf("display name = %q", result.Rendered.Client.DisplayName)
	}
	if got, want := result.Rendered.ConfigPath, filepath.Join(homeDir, ".codex", "config.toml"); got != want {
		t.Fatalf("config path = %q, want %q", got, want)
	}
	if result.Rendered.Mode != "preview" {
		t.Fatalf("mode = %q", result.Rendered.Mode)
	}
	if !strings.Contains(result.Rendered.Content, "[mcp_servers.optimusctx]") {
		t.Fatalf("content missing optimusctx table: %s", result.Rendered.Content)
	}
	if !strings.Contains(result.Rendered.Content, `command = "optimusctx"`) {
		t.Fatalf("content missing command line: %s", result.Rendered.Content)
	}
	if strings.TrimSpace(result.Rendered.AppliedContent) == "" {
		t.Fatal("codex-cli preview should keep applied content for real writes")
	}
	if len(result.Rendered.Notes) == 0 {
		t.Fatal("codex-cli preview should include notes")
	}
}

func TestInstallServiceSupportsGeminiCLIPreview(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	service := NewInstallService()
	result, err := service.Register(context.Background(), InstallRequest{ClientID: "gemini-cli"})
	if err != nil {
		t.Fatalf("Register(gemini-cli) error = %v", err)
	}
	if result.Wrote {
		t.Fatal("gemini-cli preview should not write")
	}
	if result.Rendered.Client.ID != "gemini-cli" {
		t.Fatalf("client id = %q", result.Rendered.Client.ID)
	}
	if result.Rendered.Client.DisplayName != "Gemini CLI" {
		t.Fatalf("display name = %q", result.Rendered.Client.DisplayName)
	}
	if got, want := result.Rendered.ConfigPath, filepath.Join(homeDir, ".gemini", "settings.json"); got != want {
		t.Fatalf("config path = %q, want %q", got, want)
	}
	if result.Rendered.Mode != repository.RenderModePreview {
		t.Fatalf("mode = %q", result.Rendered.Mode)
	}
	if !strings.Contains(result.Rendered.Content, "\"mcpServers\"") {
		t.Fatalf("content missing mcpServers: %s", result.Rendered.Content)
	}
	if !strings.Contains(result.Rendered.Content, "\"optimusctx\"") {
		t.Fatalf("content missing optimusctx entry: %s", result.Rendered.Content)
	}
	if strings.TrimSpace(result.Rendered.AppliedContent) == "" {
		t.Fatal("gemini-cli preview should keep applied content for real writes")
	}
	if result.Guidance == nil {
		t.Fatal("gemini-cli preview should render guidance")
	}
	if got, want := result.Guidance.Path, filepath.Join(homeDir, ".gemini", repository.GeminiGuidanceFilename); got != want {
		t.Fatalf("guidance path = %q, want %q", got, want)
	}
}

func TestInstallServiceGeminiCLIWritePreservesExistingContent(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	configPath := filepath.Join(homeDir, ".gemini", "settings.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	const existing = "{\n  \"theme\": \"dark\",\n  \"mcpServers\": {\n    \"other\": {\n      \"command\": \"other\",\n      \"args\": [\"serve\"]\n    }\n  }\n}\n"
	if err := os.WriteFile(configPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	existingGuidance := "# user guidance\n"
	guidancePath := filepath.Join(homeDir, ".gemini", repository.GeminiGuidanceFilename)
	if err := os.WriteFile(guidancePath, []byte(existingGuidance), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", guidancePath, err)
	}

	service := NewInstallService()
	result, err := service.Register(context.Background(), InstallRequest{
		ClientID: "gemini-cli",
		Write:    true,
	})
	if err != nil {
		t.Fatalf("Register(gemini-cli write) error = %v", err)
	}
	if !result.Wrote {
		t.Fatal("write should report wrote=true")
	}
	if result.Guidance == nil {
		t.Fatal("gemini-cli write should persist guidance")
	}

	contentBytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", configPath, err)
	}
	content := string(contentBytes)
	for _, want := range []string{
		"\"theme\": \"dark\"",
		"\"other\": {",
		"\"optimusctx\": {",
		"\"command\": \"optimusctx\"",
		"\"run\"",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("written Gemini config missing %q:\n%s", want, content)
		}
	}

	guidanceBytes, err := os.ReadFile(guidancePath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", guidancePath, err)
	}
	guidance := string(guidanceBytes)
	if !strings.Contains(guidance, "# user guidance") {
		t.Fatalf("guidance should preserve existing user content:\n%s", guidance)
	}
	if !strings.Contains(guidance, "OptimusCtx MCP guidance") {
		t.Fatalf("guidance missing managed block:\n%s", guidance)
	}
}

func TestInstallServiceSupportsCursorCLIPreview(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	service := NewInstallService()
	result, err := service.Register(context.Background(), InstallRequest{ClientID: "cursor-cli"})
	if err != nil {
		t.Fatalf("Register(cursor-cli) error = %v", err)
	}
	if result.Wrote {
		t.Fatal("cursor-cli preview should not write")
	}
	if result.Rendered.Client.ID != "cursor-cli" {
		t.Fatalf("client id = %q", result.Rendered.Client.ID)
	}
	if result.Rendered.Client.DisplayName != "Cursor CLI" {
		t.Fatalf("display name = %q", result.Rendered.Client.DisplayName)
	}
	if got, want := result.Rendered.ConfigPath, filepath.Join(homeDir, ".cursor", "mcp.json"); got != want {
		t.Fatalf("config path = %q, want %q", got, want)
	}
	if result.Rendered.Mode != repository.RenderModePreview {
		t.Fatalf("mode = %q", result.Rendered.Mode)
	}
	if !strings.Contains(result.Rendered.Content, "\"mcpServers\"") {
		t.Fatalf("content missing mcpServers: %s", result.Rendered.Content)
	}
	if !strings.Contains(result.Rendered.Content, "\"optimusctx\"") {
		t.Fatalf("content missing optimusctx entry: %s", result.Rendered.Content)
	}
	if strings.TrimSpace(result.Rendered.AppliedContent) == "" {
		t.Fatal("cursor-cli preview should keep applied content for real writes")
	}
	if result.Guidance != nil {
		t.Fatal("cursor-cli preview should not claim managed guidance")
	}
}

func TestInstallServiceCursorCLIWritePreservesExistingContent(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	configPath := filepath.Join(homeDir, ".cursor", "mcp.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	writeClaudeDesktopConfig(t, configPath, repository.ClientConfigDocument{
		MCPServers: map[string]repository.ServeCommand{
			"other": {Command: "other", Args: []string{"serve"}},
		},
	})

	service := NewInstallService()
	result, err := service.Register(context.Background(), InstallRequest{
		ClientID: "cursor-cli",
		Write:    true,
	})
	if err != nil {
		t.Fatalf("Register(cursor-cli write) error = %v", err)
	}
	if !result.Wrote {
		t.Fatal("write should report wrote=true")
	}
	if result.Guidance != nil {
		t.Fatal("cursor-cli write should not persist guidance")
	}

	document := readClaudeDesktopConfig(t, configPath)
	if _, ok := document.MCPServers["other"]; !ok {
		t.Fatalf("written Cursor config missing existing server: %+v", document.MCPServers)
	}
	command, ok := document.MCPServers["optimusctx"]
	if !ok {
		t.Fatalf("written Cursor config missing optimusctx: %+v", document.MCPServers)
	}
	if command.Command != "optimusctx" || len(command.Args) != 1 || command.Args[0] != "run" {
		t.Fatalf("optimusctx server command = %+v", command)
	}
}

func TestInstallServicePreservesExistingCodexConfigPreview(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("OPTIMUSCTX_CODEX_APP_CONFIG", filepath.Join(homeDir, ".codex", "config.toml"))

	configPath := filepath.Join(homeDir, ".codex", "config.toml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	const existing = `model = "gpt-5"

[mcp_servers.other]
command = "other"
args = ["serve"]

[profiles.default]
approval_policy = "on-request"
`
	if err := os.WriteFile(configPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	service := NewInstallService()
	result, err := service.Register(context.Background(), InstallRequest{ClientID: "codex-app"})
	if err != nil {
		t.Fatalf("Register(codex-app) error = %v", err)
	}
	if strings.Contains(result.Rendered.Content, `model = "gpt-5"`) {
		t.Fatalf("preview should not dump existing model settings: %s", result.Rendered.Content)
	}
	if strings.Contains(result.Rendered.Content, "[mcp_servers.other]") {
		t.Fatalf("preview should not dump unrelated MCP servers: %s", result.Rendered.Content)
	}
	if strings.Contains(result.Rendered.Content, "[profiles.default]") {
		t.Fatalf("preview should not dump unrelated profile tables: %s", result.Rendered.Content)
	}
	if !strings.Contains(result.Rendered.Content, "[mcp_servers.optimusctx]") {
		t.Fatalf("content missing optimusctx table: %s", result.Rendered.Content)
	}
	if strings.Count(result.Rendered.Content, "[mcp_servers.optimusctx]") != 1 {
		t.Fatalf("optimusctx table duplicated: %s", result.Rendered.Content)
	}
	if !strings.Contains(result.Rendered.AppliedContent, `model = "gpt-5"`) {
		t.Fatalf("applied content missing existing model: %s", result.Rendered.AppliedContent)
	}
	if !strings.Contains(result.Rendered.AppliedContent, "[mcp_servers.other]") {
		t.Fatalf("applied content missing existing server: %s", result.Rendered.AppliedContent)
	}
	if !strings.Contains(result.Rendered.AppliedContent, "[profiles.default]") {
		t.Fatalf("applied content missing profile table: %s", result.Rendered.AppliedContent)
	}
	if !strings.Contains(result.Rendered.AppliedContent, "[mcp_servers.optimusctx]") {
		t.Fatalf("applied content missing optimusctx table: %s", result.Rendered.AppliedContent)
	}
	if strings.Count(result.Rendered.AppliedContent, "[mcp_servers.optimusctx]") != 1 {
		t.Fatalf("applied optimusctx table duplicated: %s", result.Rendered.AppliedContent)
	}
}

func TestInstallServiceCodexAppWritePersistsNativeConfig(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("OPTIMUSCTX_CODEX_APP_CONFIG", filepath.Join(homeDir, ".codex", "config.toml"))

	service := NewInstallService()
	result, err := service.Register(context.Background(), InstallRequest{
		ClientID: "codex-app",
		Write:    true,
	})
	if err != nil {
		t.Fatalf("Register(codex-app write) error = %v", err)
	}
	if !result.Wrote {
		t.Fatal("write should report wrote=true")
	}
	if result.Rendered.Mode != repository.RenderModeWrite {
		t.Fatalf("mode = %q", result.Rendered.Mode)
	}

	content, err := os.ReadFile(result.Rendered.ConfigPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(content), "[mcp_servers.optimusctx]") {
		t.Fatalf("content missing optimusctx table: %s", string(content))
	}
	if !strings.Contains(string(content), `command = "optimusctx"`) {
		t.Fatalf("content missing command line: %s", string(content))
	}
	if !strings.Contains(string(content), `args = ["run"]`) {
		t.Fatalf("content missing args line: %s", string(content))
	}
	if strings.Count(string(content), "[mcp_servers.optimusctx]") != 1 {
		t.Fatalf("optimusctx table duplicated: %s", string(content))
	}
}

func TestInstallServiceCodexCLIWriteUsesExplicitConfigPath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	configPath := filepath.Join(t.TempDir(), ".codex", "config.toml")
	service := NewInstallService()
	result, err := service.Register(context.Background(), InstallRequest{
		ClientID:   "codex-cli",
		ConfigPath: configPath,
		Write:      true,
	})
	if err != nil {
		t.Fatalf("Register(codex-cli write) error = %v", err)
	}
	if !result.Wrote {
		t.Fatal("write should report wrote=true")
	}
	if result.Rendered.ConfigPath != configPath {
		t.Fatalf("config path = %q, want %q", result.Rendered.ConfigPath, configPath)
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
}

func TestInstallServiceCodexAppWritePreservesExistingContent(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("OPTIMUSCTX_CODEX_APP_CONFIG", filepath.Join(homeDir, ".codex", "config.toml"))

	configPath := filepath.Join(homeDir, ".codex", "config.toml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	const existing = `model = "gpt-5"

[mcp_servers.other]
command = "other"
args = ["serve"]

[profiles.default]
approval_policy = "on-request"
`
	if err := os.WriteFile(configPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	service := NewInstallService()
	result, err := service.Register(context.Background(), InstallRequest{
		ClientID: "codex-app",
		Write:    true,
	})
	if err != nil {
		t.Fatalf("Register(codex-app write) error = %v", err)
	}
	contentBytes, err := os.ReadFile(result.Rendered.ConfigPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	content := string(contentBytes)
	if !strings.Contains(content, `model = "gpt-5"`) {
		t.Fatalf("content missing existing model: %s", content)
	}
	if !strings.Contains(content, "[mcp_servers.other]") {
		t.Fatalf("content missing existing server: %s", content)
	}
	if !strings.Contains(content, "[profiles.default]") {
		t.Fatalf("content missing profile table: %s", content)
	}
	if !strings.Contains(content, "[mcp_servers.optimusctx]") {
		t.Fatalf("content missing optimusctx table: %s", content)
	}
	if !strings.Contains(content, `command = "optimusctx"`) {
		t.Fatalf("content missing command line: %s", content)
	}
	if !strings.Contains(content, `args = ["run"]`) {
		t.Fatalf("content missing args line: %s", content)
	}
	if strings.Count(content, "[mcp_servers.optimusctx]") != 1 {
		t.Fatalf("optimusctx table duplicated: %s", content)
	}
}

func TestInstallServiceCodexWriteIsIdempotent(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("OPTIMUSCTX_CODEX_APP_CONFIG", filepath.Join(homeDir, ".codex", "config.toml"))

	service := NewInstallService()
	for i := 0; i < 2; i++ {
		result, err := service.Register(context.Background(), InstallRequest{
			ClientID: "codex-app",
			Write:    true,
		})
		if err != nil {
			t.Fatalf("Register(codex-app write %d) error = %v", i+1, err)
		}
		if !result.Wrote {
			t.Fatalf("write %d should report wrote=true", i+1)
		}
	}

	contentBytes, err := os.ReadFile(filepath.Join(homeDir, ".codex", "config.toml"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	content := string(contentBytes)
	if strings.Count(content, "[mcp_servers.optimusctx]") != 1 {
		t.Fatalf("optimusctx table duplicated: %s", content)
	}
}

func TestInstallServiceRejectsGenericWrite(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	service := NewInstallService()
	_, err := service.Register(context.Background(), InstallRequest{ClientID: "generic", Write: true})
	if err == nil || !strings.Contains(err.Error(), "does not support --write") {
		t.Fatalf("Register(generic write) error = %v", err)
	}
}

func TestInstallServiceClaudeDesktopPreviewUsesResolvedPath(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("OPTIMUSCTX_CLAUDE_DESKTOP_CONFIG", filepath.Join(homeDir, "AppData", "Roaming", "Claude", "claude_desktop_config.json"))

	configPath, err := resolveClaudeDesktopConfigPath("")
	if err != nil {
		t.Fatalf("resolveClaudeDesktopConfigPath() error = %v", err)
	}
	writeClaudeDesktopConfig(t, configPath, repository.ClientConfigDocument{
		MCPServers: map[string]repository.ServeCommand{
			"existing": {Command: "existing-mcp", Args: []string{"serve"}},
		},
	})

	service := NewInstallService()
	result, err := service.Register(context.Background(), InstallRequest{ClientID: "claude-desktop"})
	if err != nil {
		t.Fatalf("Register(claude-desktop) error = %v", err)
	}
	if result.Wrote {
		t.Fatal("claude-desktop preview should not write")
	}
	if result.Rendered.ConfigPath != configPath {
		t.Fatalf("config path = %q, want %q", result.Rendered.ConfigPath, configPath)
	}

	document := mustDecodeClientConfig(t, result.Rendered.Content)
	if _, ok := document.MCPServers["existing"]; ok {
		t.Fatalf("preview should not dump existing server: %s", result.Rendered.Content)
	}
	command, ok := document.MCPServers["optimusctx"]
	if !ok {
		t.Fatalf("preview missing optimusctx server: %s", result.Rendered.Content)
	}
	if command.Command != "optimusctx" || len(command.Args) != 1 || command.Args[0] != "run" {
		t.Fatalf("optimusctx server command = %+v", command)
	}
	applied := mustDecodeClientConfig(t, result.Rendered.AppliedContent)
	if _, ok := applied.MCPServers["existing"]; !ok {
		t.Fatalf("applied content missing existing server: %s", result.Rendered.AppliedContent)
	}
}

func TestInstallServiceClaudeDesktopWritePreservesExistingServers(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "Claude", "claude_desktop_config.json")
	writeClaudeDesktopConfig(t, configPath, repository.ClientConfigDocument{
		MCPServers: map[string]repository.ServeCommand{
			"existing": {Command: "existing-mcp", Args: []string{"serve"}},
		},
	})

	service := NewInstallService()
	result, err := service.Register(context.Background(), InstallRequest{
		ClientID:   "claude-desktop",
		ConfigPath: configPath,
		Write:      true,
	})
	if err != nil {
		t.Fatalf("Register(claude-desktop write) error = %v", err)
	}
	if !result.Wrote {
		t.Fatal("claude-desktop write should report wrote=true")
	}

	document := readClaudeDesktopConfig(t, configPath)
	if _, ok := document.MCPServers["existing"]; !ok {
		t.Fatalf("written config missing existing server: %+v", document.MCPServers)
	}
	command, ok := document.MCPServers["optimusctx"]
	if !ok {
		t.Fatalf("written config missing optimusctx server: %+v", document.MCPServers)
	}
	if command.Command != "optimusctx" || len(command.Args) != 1 || command.Args[0] != "run" {
		t.Fatalf("optimusctx server command = %+v", command)
	}
	if len(document.MCPServers) != 2 {
		t.Fatalf("server count = %d, want 2", len(document.MCPServers))
	}
}

func TestInstallServiceClaudeDesktopWriteIsIdempotent(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "Claude", "claude_desktop_config.json")

	service := NewInstallService()
	request := InstallRequest{
		ClientID:   "claude-desktop",
		ConfigPath: configPath,
		Write:      true,
	}
	if _, err := service.Register(context.Background(), request); err != nil {
		t.Fatalf("Register(first write) error = %v", err)
	}
	if _, err := service.Register(context.Background(), request); err != nil {
		t.Fatalf("Register(second write) error = %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Count(string(content), "\"optimusctx\":") != 1 {
		t.Fatalf("optimusctx entry duplicated: %s", string(content))
	}

	document := mustDecodeClientConfig(t, string(content))
	command, ok := document.MCPServers["optimusctx"]
	if !ok {
		t.Fatalf("written config missing optimusctx server: %+v", document.MCPServers)
	}
	if command.Command != "optimusctx" || len(command.Args) != 1 || command.Args[0] != "run" {
		t.Fatalf("optimusctx server command = %+v", command)
	}
	if len(document.MCPServers) != 1 {
		t.Fatalf("server count = %d, want 1", len(document.MCPServers))
	}
}

func writeClaudeDesktopConfig(t *testing.T, configPath string, document repository.ClientConfigDocument) {
	t.Helper()

	content, err := repository.RenderClientConfig(document)
	if err != nil {
		t.Fatalf("RenderClientConfig() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func readClaudeDesktopConfig(t *testing.T, configPath string) repository.ClientConfigDocument {
	t.Helper()

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	return mustDecodeClientConfig(t, string(content))
}

func mustDecodeClientConfig(t *testing.T, content string) repository.ClientConfigDocument {
	t.Helper()

	var document repository.ClientConfigDocument
	if err := json.Unmarshal([]byte(content), &document); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return document
}
