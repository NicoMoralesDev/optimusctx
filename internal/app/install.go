package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/niccrow/optimusctx/internal/repository"
)

type InstallRequest struct {
	ClientID   string
	BinaryPath string
	ConfigPath string
	ServerName string
	Scope      string
	Write      bool
	RepoRoot   string
}

type InstallResult struct {
	Rendered repository.RenderedClientConfig
	Guidance *repository.RenderedGuidance
	Wrote    bool
}

type InstallService struct {
	adapters  map[repository.ClientID]clientRegistrationAdapter
	readFile  func(string) ([]byte, error)
	writeFile func(string, []byte, os.FileMode) error
	mkdirAll  func(string, os.FileMode) error
}

type clientRegistrationAdapter interface {
	Preview(InstallRequest) (repository.RenderedClientConfig, error)
	Write(context.Context, InstallRequest) (repository.RenderedClientConfig, error)
}

type jsonFileClientAdapter struct {
	client      repository.SupportedClient
	resolvePath func(string) (string, error)
	readFile    func(string) ([]byte, error)
	writeFile   func(string, []byte, os.FileMode) error
	mkdirAll    func(string, os.FileMode) error
	notes       []string
}

type previewOnlyClientAdapter struct {
	client     repository.SupportedClient
	configPath string
	notes      []string
}

type claudeCLIClientAdapter struct {
	client     repository.SupportedClient
	notes      []string
	runCommand func(context.Context, string, ...string) ([]byte, error)
}

type codexConfigClientAdapter struct {
	client      repository.SupportedClient
	resolvePath func(string) (string, error)
	readFile    func(string) ([]byte, error)
	writeFile   func(string, []byte, os.FileMode) error
	mkdirAll    func(string, os.FileMode) error
	notes       []string
}

type genericClientAdapter struct {
	client repository.SupportedClient
	notes  []string
}

func NewInstallService() InstallService {
	clients := repository.SupportedClients()
	claudeDesktop := clients[0]
	claudeCLI := clients[1]
	codexApp := clients[2]
	codexCLI := clients[3]
	generic := clients[4]
	jsonAdapter := jsonFileClientAdapter{
		client:      claudeDesktop,
		resolvePath: resolveClaudeDesktopConfigPath,
		readFile:    os.ReadFile,
		writeFile:   os.WriteFile,
		mkdirAll:    os.MkdirAll,
		notes:       claudeDesktopNotes(),
	}

	return InstallService{
		adapters: map[repository.ClientID]clientRegistrationAdapter{
			claudeDesktop.ID: jsonAdapter,
			claudeCLI.ID: claudeCLIClientAdapter{
				client:     claudeCLI,
				notes:      claudeCLINotes(),
				runCommand: runCommand,
			},
			codexApp.ID: codexConfigClientAdapter{
				client:      codexApp,
				resolvePath: resolveCodexConfigPath,
				readFile:    os.ReadFile,
				writeFile:   os.WriteFile,
				mkdirAll:    os.MkdirAll,
				notes:       codexAppNotes(),
			},
			codexCLI.ID: codexConfigClientAdapter{
				client:      codexCLI,
				resolvePath: resolveCodexConfigPath,
				readFile:    os.ReadFile,
				writeFile:   os.WriteFile,
				mkdirAll:    os.MkdirAll,
				notes:       codexCLINotes(),
			},
			generic.ID: genericClientAdapter{client: generic, notes: genericMCPNotes()},
		},
		readFile:  os.ReadFile,
		writeFile: os.WriteFile,
		mkdirAll:  os.MkdirAll,
	}
}

func (s InstallService) Register(ctx context.Context, request InstallRequest) (InstallResult, error) {
	if request.ClientID == "" {
		return InstallResult{}, errors.New("client is required")
	}

	adapter, ok := s.adapters[repository.ClientID(request.ClientID)]
	if !ok {
		return InstallResult{}, fmt.Errorf("unsupported client %q", request.ClientID)
	}

	if request.Write {
		rendered, err := adapter.Write(ctx, request)
		if err != nil {
			return InstallResult{}, err
		}
		guidance, err := s.renderGuidance(request, repository.RenderModeWrite)
		if err != nil {
			return InstallResult{}, err
		}
		if guidance != nil {
			if err := s.writeGuidance(*guidance); err != nil {
				return InstallResult{}, err
			}
		}
		return InstallResult{Rendered: rendered, Guidance: guidance, Wrote: true}, nil
	}

	rendered, err := adapter.Preview(request)
	if err != nil {
		return InstallResult{}, err
	}
	guidance, err := s.renderGuidance(request, repository.RenderModePreview)
	if err != nil {
		return InstallResult{}, err
	}
	return InstallResult{Rendered: rendered, Guidance: guidance, Wrote: false}, nil
}

func (a jsonFileClientAdapter) Preview(request InstallRequest) (repository.RenderedClientConfig, error) {
	configPath, err := a.resolvePath(request.ConfigPath)
	if err != nil {
		return repository.RenderedClientConfig{}, err
	}

	existing, err := a.readExisting(configPath)
	if err != nil {
		return repository.RenderedClientConfig{}, err
	}

	serverName := request.ServerName
	if serverName == "" {
		serverName = repository.DefaultMCPServerName
	}

	document, err := repository.MergeClientConfig(existing, serverName, repository.NewServeCommand(request.BinaryPath))
	if err != nil {
		return repository.RenderedClientConfig{}, err
	}
	content, err := repository.RenderClientConfig(document)
	if err != nil {
		return repository.RenderedClientConfig{}, err
	}
	preview, err := repository.RenderClientConfigSnippet(serverName, repository.NewServeCommand(request.BinaryPath))
	if err != nil {
		return repository.RenderedClientConfig{}, err
	}

	mode := repository.RenderModePreview
	if request.Write {
		mode = repository.RenderModeWrite
	}

	return repository.RenderedClientConfig{
		Client:         a.client,
		ConfigPath:     configPath,
		Mode:           mode,
		Content:        preview,
		AppliedContent: content,
		Notes:          append([]string(nil), a.notes...),
	}, nil
}

func (a jsonFileClientAdapter) Write(_ context.Context, request InstallRequest) (repository.RenderedClientConfig, error) {
	request.Write = true
	rendered, err := a.Preview(request)
	if err != nil {
		return repository.RenderedClientConfig{}, err
	}

	if err := a.mkdirAll(filepath.Dir(rendered.ConfigPath), 0o755); err != nil {
		return repository.RenderedClientConfig{}, fmt.Errorf("prepare client config directory: %w", err)
	}
	if err := a.writeFile(rendered.ConfigPath, []byte(rendered.ContentForWrite()), 0o644); err != nil {
		return repository.RenderedClientConfig{}, fmt.Errorf("write client config: %w", err)
	}

	return rendered, nil
}

func (a jsonFileClientAdapter) readExisting(configPath string) ([]byte, error) {
	content, err := a.readFile(configPath)
	if err == nil {
		return content, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return nil, fmt.Errorf("read client config: %w", err)
}

func (a previewOnlyClientAdapter) Preview(request InstallRequest) (repository.RenderedClientConfig, error) {
	content, err := renderGenericPreviewContent(request)
	if err != nil {
		return repository.RenderedClientConfig{}, err
	}

	return repository.RenderedClientConfig{
		Client:         a.client,
		ConfigPath:     a.configPath,
		Mode:           repository.RenderModePreview,
		Content:        content,
		AppliedContent: content,
		Notes:          append([]string(nil), a.notes...),
	}, nil
}

func (a previewOnlyClientAdapter) Write(_ context.Context, _ InstallRequest) (repository.RenderedClientConfig, error) {
	return repository.RenderedClientConfig{}, errors.New("this client does not support --write yet; preview the host-specific contract until native config writes land")
}

func (a claudeCLIClientAdapter) Preview(request InstallRequest) (repository.RenderedClientConfig, error) {
	return a.render(request, repository.RenderModePreview)
}

func (a claudeCLIClientAdapter) Write(ctx context.Context, request InstallRequest) (repository.RenderedClientConfig, error) {
	rendered, args, err := a.renderCommand(request, repository.RenderModeWrite)
	if err != nil {
		return repository.RenderedClientConfig{}, err
	}

	output, err := a.runCommand(ctx, "claude", args...)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return repository.RenderedClientConfig{}, errors.New("run Claude CLI registration: claude command not found; install Claude Code or rerun without --write to preview the command")
		}
		if trimmed := strings.TrimSpace(string(output)); trimmed != "" {
			return repository.RenderedClientConfig{}, fmt.Errorf("run Claude CLI registration: %s", trimmed)
		}
		return repository.RenderedClientConfig{}, fmt.Errorf("run Claude CLI registration: %w", err)
	}

	return rendered, nil
}

func (a claudeCLIClientAdapter) render(request InstallRequest, mode repository.RenderMode) (repository.RenderedClientConfig, error) {
	rendered, _, err := a.renderCommand(request, mode)
	if err != nil {
		return repository.RenderedClientConfig{}, err
	}

	return rendered, nil
}

func (a claudeCLIClientAdapter) renderCommand(request InstallRequest, mode repository.RenderMode) (repository.RenderedClientConfig, []string, error) {
	serverName := request.ServerName
	if serverName == "" {
		serverName = repository.DefaultMCPServerName
	}

	scope, err := repository.NormalizeClaudeCLIScope(request.Scope)
	if err != nil {
		return repository.RenderedClientConfig{}, nil, err
	}

	command := repository.NewServeCommand(request.BinaryPath)
	content := repository.RenderClaudeCLIAddCommand(serverName, scope, command)
	rendered := repository.RenderedClientConfig{
		Client:         a.client,
		ConfigPath:     "command",
		Mode:           mode,
		Content:        content,
		AppliedContent: content,
		Notes:          append([]string(nil), a.notes...),
	}

	args := []string{"mcp", "add", "--transport", "stdio", "--scope", scope, serverName, "--", command.Command}
	args = append(args, command.Args...)
	return rendered, args, nil
}

func (a codexConfigClientAdapter) Preview(request InstallRequest) (repository.RenderedClientConfig, error) {
	configPath, err := a.resolvePath(request.ConfigPath)
	if err != nil {
		return repository.RenderedClientConfig{}, err
	}

	existing, err := readExistingClientConfig(a.readFile, configPath)
	if err != nil {
		return repository.RenderedClientConfig{}, err
	}

	content, err := repository.MergeCodexConfig(existing, request.ServerName, repository.NewServeCommand(request.BinaryPath))
	if err != nil {
		return repository.RenderedClientConfig{}, err
	}
	preview, err := repository.RenderCodexConfig(request.ServerName, repository.NewServeCommand(request.BinaryPath))
	if err != nil {
		return repository.RenderedClientConfig{}, err
	}

	mode := repository.RenderModePreview
	if request.Write {
		mode = repository.RenderModeWrite
	}

	return repository.RenderedClientConfig{
		Client:         a.client,
		ConfigPath:     configPath,
		Mode:           mode,
		Content:        preview,
		AppliedContent: content,
		Notes:          append([]string(nil), a.notes...),
	}, nil
}

func (a codexConfigClientAdapter) Write(_ context.Context, request InstallRequest) (repository.RenderedClientConfig, error) {
	request.Write = true
	rendered, err := a.Preview(request)
	if err != nil {
		return repository.RenderedClientConfig{}, err
	}

	if err := a.mkdirAll(filepath.Dir(rendered.ConfigPath), 0o755); err != nil {
		return repository.RenderedClientConfig{}, fmt.Errorf("prepare client config directory: %w", err)
	}
	if err := a.writeFile(rendered.ConfigPath, []byte(rendered.ContentForWrite()), 0o644); err != nil {
		return repository.RenderedClientConfig{}, fmt.Errorf("write client config: %w", err)
	}

	return rendered, nil
}

func (a genericClientAdapter) Preview(request InstallRequest) (repository.RenderedClientConfig, error) {
	content, err := renderGenericPreviewContent(request)
	if err != nil {
		return repository.RenderedClientConfig{}, err
	}
	return repository.RenderedClientConfig{
		Client:         a.client,
		ConfigPath:     "manual",
		Mode:           repository.RenderModePreview,
		Content:        content,
		AppliedContent: content,
		Notes:          append([]string(nil), a.notes...),
	}, nil
}

func (a genericClientAdapter) Write(_ context.Context, request InstallRequest) (repository.RenderedClientConfig, error) {
	return repository.RenderedClientConfig{}, errors.New("this client does not support --write yet; preview the config and add it to your MCP host manually")
}

func genericMCPNotes() []string {
	return []string{
		"Place this JSON into your MCP host configuration.",
		"Use command `optimusctx` with args `[\"run\"]`.",
	}
}

func claudeDesktopNotes() []string {
	return []string{
		"Claude Desktop writes the native `claude_desktop_config.json` MCP contract.",
		"The preview shows only the `optimusctx` MCP entry; unrelated desktop config stays preserved during merges.",
		"Use --config to target a non-default Claude Desktop config path.",
	}
}

func claudeCLINotes() []string {
	return []string{
		"This previews the native Claude CLI stdio registration command.",
		"Use --scope local, --scope project, or --scope user to select Claude CLI's registration target.",
		"This command keeps optimusctx run as the canonical runtime handoff.",
	}
}

func codexAppNotes() []string {
	return []string{
		"Codex App writes the native shared Codex config.toml contract.",
		"The preview target defaults to `~/.codex/config.toml` and preserves unrelated Codex settings during merges.",
		"Use --config to target ~/.codex/config.toml or a repo-local .codex/config.toml path.",
	}
}

func codexCLINotes() []string {
	return []string{
		"Codex CLI writes the native shared Codex config.toml contract.",
		"The preview target defaults to `~/.codex/config.toml` and preserves unrelated Codex settings during merges.",
		"Use --config to target ~/.codex/config.toml or a repo-local .codex/config.toml path.",
	}
}

func renderGenericPreviewContent(request InstallRequest) (string, error) {
	serverName := request.ServerName
	if serverName == "" {
		serverName = repository.DefaultMCPServerName
	}

	document, err := repository.MergeClientConfig(nil, serverName, repository.NewServeCommand(request.BinaryPath))
	if err != nil {
		return "", err
	}
	content, err := repository.RenderClientConfig(document)
	if err != nil {
		return "", err
	}

	return content, nil
}

func readExistingClientConfig(readFile func(string) ([]byte, error), configPath string) ([]byte, error) {
	content, err := readFile(configPath)
	if err == nil {
		return content, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return nil, fmt.Errorf("read client config: %w", err)
}

func resolveCodexConfigPath(explicitPath string) (string, error) {
	if explicitPath != "" {
		return explicitPath, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}

	return filepath.Join(homeDir, ".codex", "config.toml"), nil
}

func DefaultCodexConfigPath() (string, error) {
	return resolveCodexConfigPath("")
}

func resolveClaudeDesktopConfigPath(explicitPath string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}

	return resolveClaudeDesktopConfigPathForPlatform(runtime.GOOS, homeDir, os.Getenv("AppData"), explicitPath)
}

func DefaultClaudeDesktopConfigPath() (string, error) {
	return resolveClaudeDesktopConfigPath("")
}

func (s InstallService) renderGuidance(request InstallRequest, mode repository.RenderMode) (*repository.RenderedGuidance, error) {
	switch repository.ClientID(request.ClientID) {
	case repository.ClientCodexApp, repository.ClientCodexCLI:
		return s.renderCodexGuidance(request, mode)
	case repository.ClientClaudeCLI:
		return s.renderClaudeGuidance(request, mode)
	default:
		return nil, nil
	}
}

func (s InstallService) renderCodexGuidance(request InstallRequest, mode repository.RenderMode) (*repository.RenderedGuidance, error) {
	configPath, err := resolveCodexConfigPath(request.ConfigPath)
	if err != nil {
		return nil, err
	}

	targetPath, err := s.resolveCodexGuidancePath(request.RepoRoot, configPath)
	if err != nil {
		return nil, err
	}

	existing, err := readExistingClientConfig(s.readFileOrDefault(), targetPath)
	if err != nil {
		return nil, err
	}
	block := repository.RenderCodexGuidanceBlock()
	applied, err := repository.MergeManagedGuidance(existing, block)
	if err != nil {
		return nil, err
	}

	return &repository.RenderedGuidance{
		Label:          "Codex agent guidance",
		Path:           targetPath,
		Mode:           mode,
		Content:        block,
		AppliedContent: applied,
		Notes: []string{
			"Codex reads AGENTS guidance from the active AGENTS file in the selected scope.",
			"This managed block tells Codex when to use OptimusCtx lookups, maps, bounded context, and refresh/health tools.",
		},
	}, nil
}

func (s InstallService) renderClaudeGuidance(request InstallRequest, mode repository.RenderMode) (*repository.RenderedGuidance, error) {
	targetPath, note, err := s.resolveClaudeGuidancePath(request)
	if err != nil {
		return nil, err
	}
	if targetPath == "" {
		return nil, nil
	}

	content := repository.RenderClaudeGuidanceDocument()
	return &repository.RenderedGuidance{
		Label:          "Claude agent guidance",
		Path:           targetPath,
		Mode:           mode,
		Content:        content,
		AppliedContent: content,
		Notes: []string{
			note,
			"This dedicated rule teaches Claude when to prefer OptimusCtx exact lookup, bounded context, repository maps, and health/refresh recovery.",
		},
	}, nil
}

func (s InstallService) writeGuidance(guidance repository.RenderedGuidance) error {
	if err := s.mkdirAllOrDefault()(filepath.Dir(guidance.Path), 0o755); err != nil {
		return fmt.Errorf("prepare agent guidance directory: %w", err)
	}
	if err := s.writeFileOrDefault()(guidance.Path, []byte(guidance.ContentForWrite()), 0o644); err != nil {
		return fmt.Errorf("write agent guidance: %w", err)
	}
	return nil
}

func (s InstallService) resolveCodexGuidancePath(repoRoot string, configPath string) (string, error) {
	configPath = filepath.Clean(configPath)
	repoRoot = filepath.Clean(strings.TrimSpace(repoRoot))
	if repoRoot != "." && repoRoot != "" && configPath == filepath.Join(repoRoot, ".codex", "config.toml") {
		return s.resolveActiveCodexGuidanceFile(repoRoot)
	}
	return s.resolveActiveCodexGuidanceFile(filepath.Dir(configPath))
}

func (s InstallService) resolveActiveCodexGuidanceFile(root string) (string, error) {
	overridePath := filepath.Join(root, "AGENTS.override.md")
	overrideContent, err := readExistingClientConfig(s.readFileOrDefault(), overridePath)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(string(overrideContent)) != "" {
		return overridePath, nil
	}
	return filepath.Join(root, "AGENTS.md"), nil
}

func (s InstallService) resolveClaudeGuidancePath(request InstallRequest) (string, string, error) {
	scope, err := repository.NormalizeClaudeCLIScope(request.Scope)
	if err != nil {
		return "", "", err
	}
	switch scope {
	case repository.ClaudeCLIScopeProject:
		if strings.TrimSpace(request.RepoRoot) == "" {
			return "", "", nil
		}
		return filepath.Join(request.RepoRoot, ".claude", "rules", repository.ClaudeRulesFilename), "Claude Code loads project rules from `.claude/rules/` when they exist.", nil
	case repository.ClaudeCLIScopeLocal, repository.ClaudeCLIScopeUser:
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", "", fmt.Errorf("resolve home directory: %w", err)
		}
		return filepath.Join(homeDir, ".claude", "rules", repository.ClaudeRulesFilename), "Claude Code loads user rules from `~/.claude/rules/` before project rules.", nil
	default:
		return "", "", nil
	}
}

func (s InstallService) readFileOrDefault() func(string) ([]byte, error) {
	if s.readFile != nil {
		return s.readFile
	}
	return os.ReadFile
}

func (s InstallService) writeFileOrDefault() func(string, []byte, os.FileMode) error {
	if s.writeFile != nil {
		return s.writeFile
	}
	return os.WriteFile
}

func (s InstallService) mkdirAllOrDefault() func(string, os.FileMode) error {
	if s.mkdirAll != nil {
		return s.mkdirAll
	}
	return os.MkdirAll
}

func resolveClaudeDesktopConfigPathForPlatform(goos string, homeDir string, appData string, explicitPath string) (string, error) {
	if explicitPath != "" {
		return explicitPath, nil
	}

	switch goos {
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil
	case "linux":
		return filepath.Join(homeDir, ".config", "Claude", "claude_desktop_config.json"), nil
	case "windows":
		if appData == "" {
			return "", errors.New("resolve Claude Desktop config path: %AppData% is not set; pass --config")
		}
		return filepath.Join(appData, "Claude", "claude_desktop_config.json"), nil
	default:
		return "", fmt.Errorf("resolve Claude Desktop config path: unsupported platform %q; pass --config", goos)
	}
}

func runCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}
