package release

import (
	"reflect"
	"strings"
	"testing"
)

func TestDistributionChannelPolicy(t *testing.T) {
	policy := CurrentDistributionPolicy()

	if !policy.ProductShape.LocalFirst || !policy.ProductShape.SingleBinary {
		t.Fatalf("product shape must stay local-first and single-binary: %+v", policy.ProductShape)
	}
	if policy.ProductShape.ManagedService {
		t.Fatalf("distribution policy must not claim a managed service")
	}
	if policy.ProductShape.AutomaticConfigEditsByDefault {
		t.Fatalf("distribution policy must keep install config edits opt-in")
	}

	if got, want := policyChannelIDs(policy), []string{"github-release-archive", "homebrew", "scoop", "npm"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("channel IDs = %v, want %v", got, want)
	}

	if got, want := policy.Upgrade.VerificationCommands, []string{"optimusctx version", "optimusctx status", "optimusctx doctor"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("verification commands = %v, want %v", got, want)
	}

	if got, want := policy.Support.SupportedCommands, []string{
		"optimusctx version",
		"optimusctx status",
		"optimusctx doctor",
		"optimusctx status --client claude-desktop --write",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("supported commands = %v, want %v", got, want)
	}

	assertChannelPolicy(t, policy.SupportedChannels[0], DistributionChannel{
		ID:                "github-release-archive",
		Name:              "GitHub Release archives",
		PublicationTarget: "github.com/NicoMoralesDev/optimusctx releases",
	})
	assertChannelPolicy(t, policy.SupportedChannels[1], DistributionChannel{
		ID:                "homebrew",
		Name:              "Homebrew",
		PublicationTarget: "niccrow/homebrew-tap",
	})
	assertChannelPolicy(t, policy.SupportedChannels[2], DistributionChannel{
		ID:                "scoop",
		Name:              "Scoop",
		PublicationTarget: "niccrow/scoop-bucket",
	})
	assertChannelPolicy(t, policy.SupportedChannels[3], DistributionChannel{
		ID:                "npm",
		Name:              "npm",
		PublicationTarget: "@niccrow/optimusctx",
	})

	if got, want := deferredScopeNames(policy), []string{
		"native Linux packages",
		"WinGet",
		"Chocolatey",
		"artifact signing",
		"SBOM publication",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("deferred scope = %v, want %v", got, want)
	}
}

func assertChannelPolicy(t *testing.T, got DistributionChannel, want DistributionChannel) {
	t.Helper()

	if got.ID != want.ID {
		t.Fatalf("channel ID = %q, want %q", got.ID, want.ID)
	}
	if got.Name != want.Name {
		t.Fatalf("channel %s name = %q, want %q", got.ID, got.Name, want.Name)
	}
	if got.PublicationTarget != want.PublicationTarget {
		t.Fatalf("channel %s publication target = %q, want %q", got.ID, got.PublicationTarget, want.PublicationTarget)
	}
	if got.UserFacingInstall == "" || got.Audience == "" || got.UpgradePath == "" || got.RollbackPath == "" || got.SupportBoundary == "" {
		t.Fatalf("channel %s must define install, audience, upgrade, rollback, and support details: %+v", got.ID, got)
	}
}

func policyChannelIDs(policy DistributionPolicy) []string {
	values := make([]string, 0, len(policy.SupportedChannels))
	for _, channel := range policy.SupportedChannels {
		values = append(values, channel.ID)
	}
	return values
}

func deferredScopeNames(policy DistributionPolicy) []string {
	values := make([]string, 0, len(policy.DeferredScope))
	for _, item := range policy.DeferredScope {
		values = append(values, item.Name)
	}
	return values
}

func TestRolloutPlanExamples(t *testing.T) {
	policy := CurrentDistributionPolicy()
	strategy := readRepoFile(t, "docs/distribution-strategy.md")
	checklist := readRepoFile(t, "docs/release-checklist.md")

	for _, want := range []string{
		policy.SupportedChannels[0].Name,
		policy.SupportedChannels[0].PublicationTarget,
		policy.SupportedChannels[1].Name,
		policy.SupportedChannels[1].PublicationTarget,
		policy.SupportedChannels[2].Name,
		policy.SupportedChannels[2].PublicationTarget,
		policy.SupportedChannels[3].Name,
		policy.SupportedChannels[3].PublicationTarget,
		"brew install niccrow/tap/optimusctx",
		"scoop bucket add niccrow https://github.com/niccrow/scoop-bucket.git",
		"scoop install niccrow/optimusctx",
		"npm install -g @niccrow/optimusctx",
		"npx @niccrow/optimusctx version",
		"optimusctx version",
		"optimusctx status",
		"optimusctx doctor",
	} {
		if !strings.Contains(strategy, want) {
			t.Fatalf("distribution strategy missing %q", want)
		}
	}

	for _, want := range []string{
		"GitHub Release archives",
		"Homebrew",
		"Scoop",
		"npm",
		"HOMEBREW_TAP_GITHUB_TOKEN",
		"SCOOP_BUCKET_GITHUB_TOKEN",
		"NPM_TOKEN",
		"best-effort and issue-driven",
	} {
		if !strings.Contains(checklist, want) {
			t.Fatalf("release checklist missing %q", want)
		}
	}
}

func TestUpgradePolicy(t *testing.T) {
	policy := CurrentDistributionPolicy()
	strategy := readRepoFile(t, "docs/distribution-strategy.md")

	for _, want := range []string{
		policy.Upgrade.ArchiveExpectation,
		policy.Upgrade.PackageManagerRule,
		"replace the binary manually",
		"prior tagged GitHub Release archive",
		"brew upgrade niccrow/tap/optimusctx",
		"scoop update optimusctx",
		"npm install -g @niccrow/optimusctx@latest",
		"`optimusctx status --client ...` is preview-first",
		"operator opts into `--write`",
		"native Linux packages such as `.deb` and `.rpm`",
		"WinGet",
		"Chocolatey",
		"artifact signing",
		"SBOM publication",
	} {
		if !strings.Contains(strategy, want) {
			t.Fatalf("distribution strategy missing upgrade or scope detail %q", want)
		}
	}
}

func TestOperatorRecoveryGuideStaysCanonical(t *testing.T) {
	guide := readRepoFile(t, "docs/operator-release-guide.md")
	strategy := readRepoFile(t, "docs/distribution-strategy.md")

	for _, want := range []string{
		`GitHub Release remains the canonical root and rollback source.`,
		`gh workflow run release.yml`,
		`-f release_tag=`,
		`-f publication_channel=npm`,
		`Allowed ` + "`publication_channel`" + ` values are ` + "`all`" + `, ` + "`npm`" + `, ` + "`homebrew`" + `, and ` + "`scoop`" + `.`,
	} {
		if !strings.Contains(guide, want) {
			t.Fatalf("docs/operator-release-guide.md missing %q", want)
		}
	}

	for _, want := range []string{
		`docs/operator-release-guide.md`,
		`gh workflow run release.yml`,
		`-f release_tag=`,
		`-f publication_channel=npm`,
		`-f publication_channel=homebrew`,
		`-f publication_channel=scoop`,
		`prior tagged GitHub Release archive`,
		`publish a new fixed version`,
	} {
		if !strings.Contains(strategy, want) {
			t.Fatalf("docs/distribution-strategy.md missing %q", want)
		}
	}

	for path, content := range map[string]string{
		"docs/operator-release-guide.md": guide,
		"docs/distribution-strategy.md":  strategy,
	} {
		for _, forbidden := range []string{
			`npm unpublish`,
			`publication_channel=github-release`,
		} {
			if strings.Contains(content, forbidden) {
				t.Fatalf("%s must not recommend unsupported recovery marker %q", path, forbidden)
			}
		}
	}
}

func TestDistributionDocsStayWithinSupportedScope(t *testing.T) {
	policy := CurrentDistributionPolicy()
	documents := map[string]string{
		"docs/distribution-strategy.md": readRepoFile(t, "docs/distribution-strategy.md"),
		"docs/release-checklist.md":     readRepoFile(t, "docs/release-checklist.md"),
	}
	recoveryDocuments := map[string]string{
		"docs/operator-release-guide.md": readRepoFile(t, "docs/operator-release-guide.md"),
		"docs/distribution-strategy.md":  documents["docs/distribution-strategy.md"],
	}

	for path, content := range documents {
		for _, channel := range policy.SupportedChannels {
			if !strings.Contains(content, channel.Name) {
				t.Fatalf("%s missing supported channel %q", path, channel.Name)
			}
		}
	}

	for _, want := range []string{
		"best-effort",
		"GitHub issues",
		"local-first single binary",
		"GitHub Release archive",
		"brew upgrade niccrow/tap/optimusctx",
		"scoop update optimusctx",
		"npm install -g @niccrow/optimusctx",
		"npx @niccrow/optimusctx version",
	} {
		if !strings.Contains(documents["docs/distribution-strategy.md"], want) && !strings.Contains(documents["docs/release-checklist.md"], want) {
			t.Fatalf("distribution docs must keep %q visible", want)
		}
	}

	for _, forbidden := range []string{
		"npm unpublish",
		"winget install",
		"choco install",
		"apt install",
		"dnf install",
		"yum install",
		"managed rollout service",
		"publication_channel=github-release",
	} {
		for path, content := range documents {
			if strings.Contains(strings.ToLower(content), forbidden) {
				t.Fatalf("%s should not claim unsupported behavior %q", path, forbidden)
			}
		}
		for path, content := range recoveryDocuments {
			if strings.Contains(strings.ToLower(content), forbidden) {
				t.Fatalf("%s should not recommend unsupported recovery behavior %q", path, forbidden)
			}
		}
	}
}

