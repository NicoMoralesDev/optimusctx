package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niccrow/optimusctx/internal/release"
)

func TestParseGitHubRepositorySlug(t *testing.T) {
	t.Run("supports ssh and https remotes", func(t *testing.T) {
		cases := map[string]string{
			"git@github.com:NicoMoralesDev/optimusctx.git":     "NicoMoralesDev/optimusctx",
			"https://github.com/NicoMoralesDev/optimusctx.git": "NicoMoralesDev/optimusctx",
			"https://github.com/NicoMoralesDev/optimusctx":     "NicoMoralesDev/optimusctx",
		}

		for remote, want := range cases {
			got, err := parseGitHubRepositorySlug(remote)
			if err != nil {
				t.Fatalf("parseGitHubRepositorySlug(%q) error = %v", remote, err)
			}
			if got != want {
				t.Fatalf("parseGitHubRepositorySlug(%q) = %q, want %q", remote, got, want)
			}
		}
	})

	t.Run("rejects non github remotes", func(t *testing.T) {
		if _, err := parseGitHubRepositorySlug("https://gitlab.com/example/repo.git"); err == nil {
			t.Fatal("parseGitHubRepositorySlug(non-github) error = nil, want rejection")
		}
	})
}

func TestFormatReleasePreparationShowsChannelSummariesAndWarningNextStep(t *testing.T) {
	output := formatReleasePreparation(release.ReleasePreparation{
		Version: "1.3.4",
		Tag:     "v1.3.4",
		Channels: []release.ReleaseChannelPlan{
			{
				ID:                release.ReleaseChannelHomebrew,
				Name:              "Homebrew",
				PublicationTarget: "niccrow/homebrew-tap",
				Selected:          true,
				Readiness:         "review_required",
				Summary:           "Homebrew publication still needs the GitHub Actions secret HOMEBREW_TAP_GITHUB_TOKEN; release prepare could not verify whether it exists",
				Details: []string{
					"required secret: HOMEBREW_TAP_GITHUB_TOKEN",
					"verification source: GitHub Actions repository secrets via gh secret list",
				},
			},
		},
		Warnings: []release.ReleaseIssue{{
			Code:    "channel-homebrew",
			Message: "Homebrew publication still needs the GitHub Actions secret HOMEBREW_TAP_GITHUB_TOKEN; release prepare could not verify whether it exists",
		}},
	}, releasePrepareOptions{})

	for _, want := range []string{
		"summary: Homebrew publication still needs the GitHub Actions secret HOMEBREW_TAP_GITHUB_TOKEN",
		"required secret: HOMEBREW_TAP_GITHUB_TOKEN",
		"verification source: GitHub Actions repository secrets via gh secret list",
		"Review the warnings, especially any downstream credential verification gaps",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("formatReleasePreparation output missing %q:\n%s", want, output)
		}
	}
}

func TestLoadReleaseMilestoneStripsOptionalQuotes(t *testing.T) {
	repoRoot := t.TempDir()
	planningDir := filepath.Join(repoRoot, ".planning")
	if err := os.MkdirAll(planningDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(.planning) error = %v", err)
	}

	state := "---\nmilestone: \"1.3.8\"\n---\n"
	if err := os.WriteFile(filepath.Join(planningDir, "STATE.md"), []byte(state), 0o644); err != nil {
		t.Fatalf("WriteFile(STATE.md) error = %v", err)
	}

	got, err := loadReleaseMilestone(repoRoot)
	if err != nil {
		t.Fatalf("loadReleaseMilestone error = %v", err)
	}
	if got != "1.3.8" {
		t.Fatalf("loadReleaseMilestone = %q, want %q", got, "1.3.8")
	}
}
