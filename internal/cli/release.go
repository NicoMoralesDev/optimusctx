package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/niccrow/optimusctx/internal/release"
	"github.com/niccrow/optimusctx/internal/repository"
)

var (
	releasePrepareGetwd           = os.Getwd
	releasePrepareResolveRepoRoot = repository.ResolveRepositoryRoot
	releasePrepareLoadMilestone   = loadReleaseMilestone
	releasePrepareCommandService  = defaultReleasePrepareCommandService
	errReleasePlanHasBlockers     = errors.New("release plan has blockers")
	errReleaseRequiresSubcommand  = errors.New("release requires a subcommand")
)

type releasePrepareRequest struct {
	RepositoryRoot   string
	Milestone        string
	Version          string
	SelectedChannels []string
}

type releasePrepareOptions struct {
	Version  string
	Channels []string
	JSON     bool
	NoPrompt bool
	Confirm  bool
}

type releasePrepareJSONOutput struct {
	Status        string                       `json:"status"`
	Version       string                       `json:"version"`
	Tag           string                       `json:"tag"`
	Channels      []release.ReleaseChannelPlan `json:"channels"`
	Checks        []release.ReleaseCheck       `json:"checks"`
	Warnings      []release.ReleaseIssue       `json:"warnings"`
	Blockers      []release.ReleaseIssue       `json:"blockers"`
	Confirmed     bool                         `json:"confirmed"`
	NextStep      string                       `json:"nextStep"`
	PhaseBoundary string                       `json:"phaseBoundary"`
}

func newReleaseCommand() *Command {
	return &Command{
		Name:    "release",
		Summary: "Prepare and validate a release plan",
		Run: func(stdout io.Writer, args []string) error {
			if len(args) == 0 {
				writeReleaseHelp(stdout)
				return errReleaseRequiresSubcommand
			}

			switch args[0] {
			case "-h", "--help", "help":
				writeReleaseHelp(stdout)
				return nil
			case "prepare":
				return runReleasePrepareCommand(stdout, args[1:])
			default:
				writeReleaseHelp(stdout)
				return fmt.Errorf("unknown release subcommand %q", args[0])
			}
		},
	}
}

func runReleasePrepareCommand(stdout io.Writer, args []string) error {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return writeReleasePrepareHelp(stdout)
		}
	}

	options, err := parseReleasePrepareArgs(args)
	if err != nil {
		return err
	}

	workingDir, err := releasePrepareGetwd()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}

	root, err := releasePrepareResolveRepoRoot(workingDir)
	if err != nil {
		return fmt.Errorf("resolve repository root: %w", err)
	}

	milestone, err := releasePrepareLoadMilestone(root.RootPath)
	if err != nil {
		return fmt.Errorf("load release milestone: %w", err)
	}

	preparation, err := releasePrepareCommandService(context.Background(), releasePrepareRequest{
		RepositoryRoot:   root.RootPath,
		Milestone:        milestone,
		Version:          options.Version,
		SelectedChannels: options.Channels,
	})
	if err != nil {
		return err
	}

	if options.JSON {
		if err := writeReleasePrepareJSON(stdout, preparation, options.Confirm); err != nil {
			return err
		}
	} else {
		if _, err := io.WriteString(stdout, formatReleasePreparation(preparation, options)); err != nil {
			return err
		}
	}

	if len(preparation.Blockers) > 0 {
		return errReleasePlanHasBlockers
	}

	return nil
}

func parseReleasePrepareArgs(args []string) (releasePrepareOptions, error) {
	options := releasePrepareOptions{}

	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "--version":
			value, next, err := requireReleasePrepareValue(args, index, arg)
			if err != nil {
				return releasePrepareOptions{}, err
			}
			options.Version = value
			index = next
		case "--channel":
			value, next, err := requireReleasePrepareValue(args, index, arg)
			if err != nil {
				return releasePrepareOptions{}, err
			}
			options.Channels = append(options.Channels, value)
			index = next
		case "--json":
			options.JSON = true
		case "--no-prompt":
			options.NoPrompt = true
		case "--confirm":
			options.Confirm = true
		default:
			if len(arg) > 0 && arg[0] == '-' {
				return releasePrepareOptions{}, fmt.Errorf("unknown release prepare flag %q", arg)
			}
			return releasePrepareOptions{}, fmt.Errorf("release prepare does not accept arguments; got %q", arg)
		}
	}

	channels, err := normalizeReleaseChannels(options.Channels)
	if err != nil {
		return releasePrepareOptions{}, err
	}
	options.Channels = channels
	return options, nil
}

func requireReleasePrepareValue(args []string, index int, flag string) (string, int, error) {
	next := index + 1
	if next >= len(args) {
		return "", index, fmt.Errorf("%s requires a value", flag)
	}
	value := strings.TrimSpace(args[next])
	if value == "" || strings.HasPrefix(value, "-") {
		return "", index, fmt.Errorf("%s requires a value", flag)
	}
	return value, next, nil
}

func normalizeReleaseChannels(input []string) ([]string, error) {
	if len(input) == 0 {
		return nil, nil
	}

	valid := map[string]struct{}{}
	for _, channel := range release.CurrentDistributionPolicy().SupportedChannels {
		valid[channel.ID] = struct{}{}
	}

	normalized := make([]string, 0, len(input))
	for _, channel := range input {
		if _, ok := valid[channel]; !ok {
			return nil, fmt.Errorf("unknown release channel %q", channel)
		}
		if slices.Contains(normalized, channel) {
			continue
		}
		normalized = append(normalized, channel)
	}

	return normalized, nil
}

func defaultReleasePrepareCommandService(ctx context.Context, request releasePrepareRequest) (release.ReleasePreparation, error) {
	return release.PrepareRelease(ctx, request.Version, request.Milestone, release.ReleasePreparationOptions{
		Git:              releaseCLIGitProbe{dir: request.RepositoryRoot},
		Files:            os.DirFS(request.RepositoryRoot),
		SelectedChannels: request.SelectedChannels,
	})
}

type releaseCLIGitProbe struct {
	dir string
}

func (p releaseCLIGitProbe) WorktreeStatus(ctx context.Context) (string, error) {
	return runGitInDir(ctx, p.dir, "status", "--porcelain")
}

func (p releaseCLIGitProbe) LocalTags(ctx context.Context) ([]string, error) {
	output, err := runGitInDir(ctx, p.dir, "tag", "--list")
	if err != nil {
		return nil, err
	}
	return nonEmptyReleaseLines(output), nil
}

func (p releaseCLIGitProbe) RemoteTags(ctx context.Context, remote string) ([]string, error) {
	output, err := runGitInDir(ctx, p.dir, "ls-remote", "--tags", remote)
	if err != nil {
		return nil, err
	}

	seen := map[string]struct{}{}
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		tag := strings.TrimPrefix(fields[1], "refs/tags/")
		tag = strings.TrimSuffix(tag, "^{}")
		if tag == "" {
			continue
		}
		seen[tag] = struct{}{}
	}

	tags := make([]string, 0, len(seen))
	for tag := range seen {
		tags = append(tags, tag)
	}
	slices.Sort(tags)
	return tags, nil
}

func runGitInDir(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v: %w: %s", args, err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

func loadReleaseMilestone(repoRoot string) (string, error) {
	statePath := filepath.Join(repoRoot, ".planning", "STATE.md")
	data, err := os.ReadFile(statePath)
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "milestone:") {
			continue
		}

		value := strings.TrimSpace(strings.TrimPrefix(line, "milestone:"))
		if value == "" {
			break
		}
		return value, nil
	}

	return "", fmt.Errorf("milestone missing from %s", statePath)
}

func writeReleaseHelp(stdout io.Writer) {
	_, _ = io.WriteString(stdout, "Usage:\n  optimusctx release <command>\n\nAvailable Commands:\n  prepare   Prepare and validate a release plan\n")
}

func writeReleasePrepareHelp(stdout io.Writer) error {
	_, err := io.WriteString(stdout, "Usage:\n  optimusctx release prepare [--version SEMVER] [--channel ID] [--json] [--no-prompt] [--confirm]\n\nReview the proposed release version, canonical tag, selected channels, blockers, and warnings without creating a tag or starting publication.\n")
	return err
}

func formatReleasePreparation(preparation release.ReleasePreparation, options releasePrepareOptions) string {
	var b strings.Builder

	_, _ = fmt.Fprintf(&b, "Version: %s\nTag: %s\n\nSelected Channels:\n", preparation.Version, preparation.Tag)
	for _, channel := range preparation.Channels {
		if !channel.Selected {
			continue
		}
		_, _ = fmt.Fprintf(&b, "  - %s (%s) [%s] -> %s\n", channel.ID, channel.Name, channel.Readiness, channel.PublicationTarget)
	}

	writeReleaseIssues(&b, "Blockers", preparation.Blockers)
	writeReleaseIssues(&b, "Warnings", preparation.Warnings)

	_, _ = fmt.Fprintf(&b, "\nNext Step: %s\n", releasePrepareNextStep(preparation, options.Confirm))
	if options.Confirm && len(preparation.Blockers) == 0 {
		_, _ = fmt.Fprintf(&b, "release plan confirmed\nconfirmed tag: %s\nconfirmed channels: %s\nPhase 16 review only: no tag created; publication not started.\n", preparation.Tag, strings.Join(preparation.SelectedChannelIDs(), ", "))
		return b.String()
	}

	if !options.NoPrompt && len(preparation.Blockers) == 0 {
		_, _ = io.WriteString(&b, "Confirmation pending: rerun with --confirm after reviewing the plan.\n")
	}

	return b.String()
}

func writeReleaseIssues(b *strings.Builder, title string, issues []release.ReleaseIssue) {
	_, _ = fmt.Fprintf(b, "\n%s:\n", title)
	if len(issues) == 0 {
		_, _ = io.WriteString(b, "  - none\n")
		return
	}

	for _, issue := range issues {
		_, _ = fmt.Fprintf(b, "  - %s: %s\n", issue.Code, issue.Message)
		for _, detail := range issue.Details {
			_, _ = fmt.Fprintf(b, "    %s\n", detail)
		}
	}
}

func writeReleasePrepareJSON(stdout io.Writer, preparation release.ReleasePreparation, confirm bool) error {
	payload := releasePrepareJSONOutput{
		Status:        releasePrepareStatus(preparation, confirm),
		Version:       preparation.Version,
		Tag:           preparation.Tag,
		Channels:      nonNilReleaseChannels(preparation.Channels),
		Checks:        nonNilReleaseChecks(preparation.Checks),
		Warnings:      nonNilReleaseIssues(preparation.Warnings),
		Blockers:      nonNilReleaseIssues(preparation.Blockers),
		Confirmed:     confirm && len(preparation.Blockers) == 0,
		NextStep:      releasePrepareNextStep(preparation, confirm),
		PhaseBoundary: "Phase 16 stops before tag creation and publication.",
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(payload)
}

func releasePrepareStatus(preparation release.ReleasePreparation, confirm bool) string {
	switch {
	case len(preparation.Blockers) > 0:
		return "blocked"
	case confirm:
		return "confirmed"
	default:
		return "ready"
	}
}

func releasePrepareNextStep(preparation release.ReleasePreparation, confirm bool) string {
	switch {
	case len(preparation.Blockers) > 0:
		return "Resolve the blockers, then rerun optimusctx release prepare."
	case confirm:
		return "Phase 16 review gate is complete; no tag created and publication not started."
	default:
		return "Review the plan, then rerun optimusctx release prepare --confirm when ready."
	}
}

func nonNilReleaseChannels(channels []release.ReleaseChannelPlan) []release.ReleaseChannelPlan {
	if channels == nil {
		return []release.ReleaseChannelPlan{}
	}
	return channels
}

func nonNilReleaseChecks(checks []release.ReleaseCheck) []release.ReleaseCheck {
	if checks == nil {
		return []release.ReleaseCheck{}
	}
	return checks
}

func nonNilReleaseIssues(issues []release.ReleaseIssue) []release.ReleaseIssue {
	if issues == nil {
		return []release.ReleaseIssue{}
	}
	return issues
}

func nonEmptyReleaseLines(input string) []string {
	lines := strings.Split(input, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		filtered = append(filtered, line)
	}
	return filtered
}

func readReleaseFile(files fs.FS, path string) (string, error) {
	data, err := fs.ReadFile(files, path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
