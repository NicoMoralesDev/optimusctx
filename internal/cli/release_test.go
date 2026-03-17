package cli

import (
	"bytes"
	"context"
	"encoding/json"
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
			Blockers: []release.ReleaseIssue{
				{Code: "blocked", Message: "stop here"},
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
