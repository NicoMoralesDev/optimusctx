package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/niccrow/optimusctx/internal/release"
	"github.com/niccrow/optimusctx/internal/repository"
)

func TestReleasePrepareCommand(t *testing.T) {
	t.Run("root help lists release", func(t *testing.T) {
		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"--help"}, &stdout); err != nil {
			t.Fatalf("Execute(--help) error = %v", err)
		}

		output := stdout.String()
		for _, want := range []string{
			"release",
			"Prepare and validate a release plan",
		} {
			if !strings.Contains(output, want) {
				t.Fatalf("root help missing %q in %q", want, output)
			}
		}
	})

	t.Run("proposes a default release plan from the active milestone", func(t *testing.T) {
		deps := stubReleasePrepareDeps(t)
		deps.milestone = "v1.2"
		deps.preparation = release.ReleasePreparation{
			Version: "1.2.0",
			Tag:     "v1.2.0",
			Channels: []release.ReleaseChannelPlan{
				{
					ID:                release.ReleaseChannelGitHubArchive,
					Name:              "GitHub Release archives",
					PublicationTarget: "github.com/niccrow/optimusctx releases",
					Selected:          true,
					Readiness:         "ready",
				},
				{
					ID:                release.ReleaseChannelNPM,
					Name:              "npm",
					PublicationTarget: "@niccrow/optimusctx",
					Selected:          true,
					Readiness:         "ready",
				},
			},
		}

		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"release", "prepare"}, &stdout); err != nil {
			t.Fatalf("Execute(release prepare) error = %v", err)
		}

		if deps.request.Version != "" {
			t.Fatalf("request version = %q, want empty default proposal", deps.request.Version)
		}
		if deps.request.Milestone != "v1.2" {
			t.Fatalf("request milestone = %q, want v1.2", deps.request.Milestone)
		}
		if deps.request.RepositoryRoot != "/repo" {
			t.Fatalf("request repository root = %q, want /repo", deps.request.RepositoryRoot)
		}

		output := stdout.String()
		for _, want := range []string{
			"Version: 1.2.0",
			"Tag: v1.2.0",
			"Selected Channels:",
			"github-release-archive",
			"npm",
			"Next Step:",
		} {
			if !strings.Contains(output, want) {
				t.Fatalf("release prepare output missing %q in %q", want, output)
			}
		}
	})

	t.Run("honors version override and repeated channel selection", func(t *testing.T) {
		deps := stubReleasePrepareDeps(t)
		deps.preparation = release.ReleasePreparation{
			Version: "1.2.3",
			Tag:     "v1.2.3",
			Channels: []release.ReleaseChannelPlan{
				{
					ID:                release.ReleaseChannelNPM,
					Name:              "npm",
					PublicationTarget: "@niccrow/optimusctx",
					Selected:          true,
					Readiness:         "ready",
				},
			},
		}

		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{
			"release", "prepare",
			"--version", "1.2.3",
			"--channel", release.ReleaseChannelNPM,
			"--channel", release.ReleaseChannelNPM,
			"--no-prompt",
		}, &stdout); err != nil {
			t.Fatalf("Execute(release prepare override) error = %v", err)
		}

		if deps.request.Version != "1.2.3" {
			t.Fatalf("request version = %q, want 1.2.3", deps.request.Version)
		}
		if got, want := deps.request.SelectedChannels, []string{release.ReleaseChannelNPM}; strings.Join(got, ",") != strings.Join(want, ",") {
			t.Fatalf("selected channels = %v, want %v", got, want)
		}
		if strings.Contains(stdout.String(), "Confirmation pending") {
			t.Fatalf("output should suppress confirmation guidance when --no-prompt is set: %q", stdout.String())
		}
	})

	t.Run("emits machine-readable release plan output", func(t *testing.T) {
		deps := stubReleasePrepareDeps(t)
		deps.preparation = release.ReleasePreparation{
			Version: "1.2.0",
			Tag:     "v1.2.0",
			Channels: []release.ReleaseChannelPlan{
				{
					ID:                release.ReleaseChannelGitHubArchive,
					Name:              "GitHub Release archives",
					PublicationTarget: "github.com/niccrow/optimusctx releases",
					Selected:          true,
					Readiness:         "ready",
				},
			},
			Warnings: []release.ReleaseIssue{
				{Code: "warning", Message: "watch this"},
			},
		}

		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"release", "prepare", "--json"}, &stdout); err != nil {
			t.Fatalf("Execute(release prepare --json) error = %v", err)
		}

		var payload map[string]any
		if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
			t.Fatalf("json.Unmarshal(stdout) error = %v; output=%s", err, stdout.String())
		}

		if got, want := payload["version"], "1.2.0"; got != want {
			t.Fatalf("json version = %v, want %v", got, want)
		}
		if got, want := payload["tag"], "v1.2.0"; got != want {
			t.Fatalf("json tag = %v, want %v", got, want)
		}
		for _, key := range []string{"channels", "warnings", "blockers", "checks"} {
			value, ok := payload[key]
			if !ok {
				t.Fatalf("json payload missing %q: %s", key, stdout.String())
			}
			if _, ok := value.([]any); !ok {
				t.Fatalf("json %q should be an array, got %T", key, value)
			}
		}
	})

	t.Run("rejects unknown release channels", func(t *testing.T) {
		var stdout bytes.Buffer
		err := NewRootCommand().Execute([]string{"release", "prepare", "--channel", "unknown"}, &stdout)
		if err == nil {
			t.Fatal("Execute(release prepare --channel unknown) error = nil, want rejection")
		}
		if !strings.Contains(err.Error(), `unknown release channel "unknown"`) {
			t.Fatalf("error = %v, want unknown channel rejection", err)
		}
	})
}

func TestReleasePrepareHelp(t *testing.T) {
	var stdout bytes.Buffer
	if err := NewRootCommand().Execute([]string{"release", "prepare", "--help"}, &stdout); err != nil {
		t.Fatalf("Execute(release prepare --help) error = %v", err)
	}

	output := stdout.String()
	for _, want := range []string{
		"optimusctx release prepare",
		"--version",
		"--channel",
		"--json",
		"--confirm",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("help output missing %q in %q", want, output)
		}
	}
}

func TestReleasePrepareConfirmGate(t *testing.T) {
	t.Run("prints a review-only confirmation without mutation", func(t *testing.T) {
		deps := stubReleasePrepareDeps(t)
		deps.preparation = release.ReleasePreparation{
			Version: "1.2.0",
			Tag:     "v1.2.0",
			Channels: []release.ReleaseChannelPlan{
				{
					ID:                release.ReleaseChannelGitHubArchive,
					Name:              "GitHub Release archives",
					PublicationTarget: "github.com/niccrow/optimusctx releases",
					Selected:          true,
					Readiness:         "ready",
				},
				{
					ID:                release.ReleaseChannelNPM,
					Name:              "npm",
					PublicationTarget: "@niccrow/optimusctx",
					Selected:          true,
					Readiness:         "ready",
				},
			},
		}

		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{"release", "prepare", "--confirm"}, &stdout); err != nil {
			t.Fatalf("Execute(release prepare --confirm) error = %v", err)
		}

		output := stdout.String()
		for _, want := range []string{
			"release plan confirmed",
			"confirmed tag: v1.2.0",
			"confirmed channels: github-release-archive, npm",
			"no tag created",
			"publication not started",
		} {
			if !strings.Contains(output, want) {
				t.Fatalf("confirm output missing %q in %q", want, output)
			}
		}
	})

	t.Run("emits machine-readable blocker output and returns non-zero", func(t *testing.T) {
		deps := stubReleasePrepareDeps(t)
		deps.preparation = release.ReleasePreparation{
			Version: "1.2.0",
			Tag:     "v1.2.0",
			Channels: []release.ReleaseChannelPlan{
				{
					ID:                release.ReleaseChannelGitHubArchive,
					Name:              "GitHub Release archives",
					PublicationTarget: "github.com/niccrow/optimusctx releases",
					Selected:          true,
					Readiness:         "blocked",
				},
			},
			Warnings: []release.ReleaseIssue{
				{Code: "warning", Message: "watch this"},
			},
			Blockers: []release.ReleaseIssue{
				{Code: "blocked", Message: "stop here"},
			},
		}

		var stdout bytes.Buffer
		err := NewRootCommand().Execute([]string{"release", "prepare", "--json"}, &stdout)
		if !errors.Is(err, errReleasePlanHasBlockers) {
			t.Fatalf("Execute(release prepare --json) error = %v, want %v", err, errReleasePlanHasBlockers)
		}

		var payload map[string]any
		if unmarshalErr := json.Unmarshal(stdout.Bytes(), &payload); unmarshalErr != nil {
			t.Fatalf("json.Unmarshal(stdout) error = %v; output=%s", unmarshalErr, stdout.String())
		}

		if got, want := payload["version"], "1.2.0"; got != want {
			t.Fatalf("json version = %v, want %v", got, want)
		}
		if got, want := payload["tag"], "v1.2.0"; got != want {
			t.Fatalf("json tag = %v, want %v", got, want)
		}
		for _, key := range []string{"channels", "warnings", "blockers", "checks"} {
			value, ok := payload[key]
			if !ok {
				t.Fatalf("json payload missing %q: %s", key, stdout.String())
			}
			if _, ok := value.([]any); !ok {
				t.Fatalf("json %q should be an array, got %T", key, value)
			}
		}
		if got, want := payload["status"], "blocked"; got != want {
			t.Fatalf("json status = %v, want %v", got, want)
		}
	})
}

func TestReleasePrepareSelectedChannelsReady(t *testing.T) {
	t.Run("json output stays ready when only selected channels are ready", func(t *testing.T) {
		deps := stubReleasePrepareDeps(t)
		deps.preparation = release.ReleasePreparation{
			Version: "1.2.0",
			Tag:     "v1.2.0",
			Channels: []release.ReleaseChannelPlan{
				{
					ID:                release.ReleaseChannelGitHubArchive,
					Name:              "GitHub Release archives",
					PublicationTarget: "github.com/niccrow/optimusctx releases",
					Selected:          true,
					Readiness:         "ready",
				},
				// Selected:          false keeps blocked unselected channels informational in the shared model.
				{
					ID:                release.ReleaseChannelHomebrew,
					Name:              "Homebrew",
					PublicationTarget: "niccrow/tap",
					Selected:          false,
					Readiness:         "blocked",
				},
				{
					ID:                release.ReleaseChannelScoop,
					Name:              "Scoop",
					PublicationTarget: "niccrow/scoop-bucket",
					Selected:          false,
					Readiness:         "blocked",
				},
				{
					ID:                release.ReleaseChannelNPM,
					Name:              "npm",
					PublicationTarget: "@niccrow/optimusctx",
					Selected:          true,
					Readiness:         "ready",
				},
			},
			Checks: []release.ReleaseCheck{
				{Code: "channel-github-release", Target: release.ReleaseChannelGitHubArchive, Status: "ready", Message: "GitHub ready"},
				{Code: "channel-homebrew", Target: release.ReleaseChannelHomebrew, Status: "blocked", Message: "Homebrew still unwired"},
				{Code: "channel-scoop", Target: release.ReleaseChannelScoop, Status: "blocked", Message: "Scoop still unwired"},
				{Code: "channel-npm", Target: release.ReleaseChannelNPM, Status: "ready", Message: "npm ready"},
			},
			Blockers: []release.ReleaseIssue{},
		}

		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{
			"release", "prepare",
			"--channel", release.ReleaseChannelGitHubArchive,
			"--channel", release.ReleaseChannelNPM,
			"--json",
		}, &stdout); err != nil {
			t.Fatalf("Execute(release prepare selected channels --json) error = %v", err)
		}

		if got, want := deps.request.SelectedChannels, []string{release.ReleaseChannelGitHubArchive, release.ReleaseChannelNPM}; strings.Join(got, ",") != strings.Join(want, ",") {
			t.Fatalf("selected channels = %v, want %v", got, want)
		}

		var payload releasePrepareJSONOutput
		if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
			t.Fatalf("json.Unmarshal(stdout) error = %v; output=%s", err, stdout.String())
		}

		if got, want := payload.Status, "ready"; got != want {
			t.Fatalf("\"status\" = %q, want %q", got, want)
		}
		if got := len(payload.Blockers); got != 0 {
			t.Fatalf("blockers len = %d, want 0", got)
		}
		if got := len(payload.Channels); got != 4 {
			t.Fatalf("channels len = %d, want 4", got)
		}
		if got := payload.Channels[0].Selected; !got {
			t.Fatalf("github selected = %t, want true", got)
		}
		if got := payload.Channels[1].Selected; got {
			t.Fatalf("homebrew selected = %t, want false", got)
		}
		if got := payload.Channels[1].Readiness; got != "blocked" {
			t.Fatalf("homebrew readiness = %q, want %q", got, "blocked")
		}
		if got := payload.Channels[2].Selected; got {
			t.Fatalf("scoop selected = %t, want false", got)
		}
		if got := payload.Channels[2].Readiness; got != "blocked" {
			t.Fatalf("scoop readiness = %q, want %q", got, "blocked")
		}
		if got := payload.Channels[3].Selected; !got {
			t.Fatalf("npm selected = %t, want true", got)
		}
	})

	t.Run("confirm output stays review-only for the selected ready subset", func(t *testing.T) {
		deps := stubReleasePrepareDeps(t)
		deps.preparation = release.ReleasePreparation{
			Version: "1.2.0",
			Tag:     "v1.2.0",
			Channels: []release.ReleaseChannelPlan{
				{
					ID:                release.ReleaseChannelGitHubArchive,
					Name:              "GitHub Release archives",
					PublicationTarget: "github.com/niccrow/optimusctx releases",
					Selected:          true,
					Readiness:         "ready",
				},
				// Selected:          false keeps blocked unselected channels informational in the shared model.
				{
					ID:                release.ReleaseChannelHomebrew,
					Name:              "Homebrew",
					PublicationTarget: "niccrow/tap",
					Selected:          false,
					Readiness:         "blocked",
				},
				{
					ID:                release.ReleaseChannelScoop,
					Name:              "Scoop",
					PublicationTarget: "niccrow/scoop-bucket",
					Selected:          false,
					Readiness:         "blocked",
				},
				{
					ID:                release.ReleaseChannelNPM,
					Name:              "npm",
					PublicationTarget: "@niccrow/optimusctx",
					Selected:          true,
					Readiness:         "ready",
				},
			},
		}

		var stdout bytes.Buffer
		if err := NewRootCommand().Execute([]string{
			"release", "prepare",
			"--channel", release.ReleaseChannelGitHubArchive,
			"--channel", release.ReleaseChannelNPM,
			"--confirm",
		}, &stdout); err != nil {
			t.Fatalf("Execute(release prepare selected channels --confirm) error = %v", err)
		}

		output := stdout.String()
		for _, want := range []string{
			"release plan confirmed",
			"confirmed channels: github-release-archive, npm",
			"no tag created",
			"publication not started",
		} {
			if !strings.Contains(output, want) {
				t.Fatalf("confirm output missing %q in %q", want, output)
			}
		}
	})
}

type releasePrepareTestDeps struct {
	milestone   string
	preparation release.ReleasePreparation
	request     releasePrepareRequest
	err         error
}

func stubReleasePrepareDeps(t *testing.T) *releasePrepareTestDeps {
	t.Helper()

	deps := &releasePrepareTestDeps{
		milestone: "v1.2",
	}

	origGetwd := releasePrepareGetwd
	origResolveRepoRoot := releasePrepareResolveRepoRoot
	origLoadMilestone := releasePrepareLoadMilestone
	origService := releasePrepareCommandService

	releasePrepareGetwd = func() (string, error) {
		return "/repo/subdir", nil
	}
	releasePrepareResolveRepoRoot = func(string) (repository.RepositoryRoot, error) {
		return repository.RepositoryRoot{RootPath: "/repo"}, nil
	}
	releasePrepareLoadMilestone = func(string) (string, error) {
		return deps.milestone, nil
	}
	releasePrepareCommandService = func(_ context.Context, request releasePrepareRequest) (release.ReleasePreparation, error) {
		deps.request = request
		return deps.preparation, deps.err
	}

	t.Cleanup(func() {
		releasePrepareGetwd = origGetwd
		releasePrepareResolveRepoRoot = origResolveRepoRoot
		releasePrepareLoadMilestone = origLoadMilestone
		releasePrepareCommandService = origService
	})

	return deps
}
