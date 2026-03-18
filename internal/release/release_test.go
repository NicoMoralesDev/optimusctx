package release

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestArchiveMatrix(t *testing.T) {
	config := readRepoFile(t, ".goreleaser.yml")

	if got, want := yamlList(config, "goos"), []string{"darwin", "linux", "windows"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("goos = %v, want %v", got, want)
	}
	if got, want := yamlList(config, "goarch"), []string{"amd64", "arm64"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("goarch = %v, want %v", got, want)
	}

	for _, want := range []string{
		`binary: optimusctx`,
		`name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"`,
		`format: tar.gz`,
		`wrap_in_directory: false`,
		`goos: windows`,
		`format: zip`,
		`-X github.com/niccrow/optimusctx/internal/buildinfo.Version={{ .Version }}`,
		`-X github.com/niccrow/optimusctx/internal/buildinfo.Commit={{ .FullCommit }}`,
		`-X github.com/niccrow/optimusctx/internal/buildinfo.BuildDate={{ .Date }}`,
	} {
		if !strings.Contains(config, want) {
			t.Fatalf(".goreleaser.yml missing %q", want)
		}
	}
}

func TestChecksumManifest(t *testing.T) {
	config := readRepoFile(t, ".goreleaser.yml")

	for _, want := range []string{
		`algorithm: sha256`,
		`name_template: "{{ .ProjectName }}_{{ .Version }}_checksums.txt"`,
		`owner: NicoMoralesDev`,
		`name: optimusctx`,
	} {
		if !strings.Contains(config, want) {
			t.Fatalf(".goreleaser.yml missing %q", want)
		}
	}

	for _, forbidden := range []string{"npx", "npm publish", "postinstall"} {
		if strings.Contains(config, forbidden) {
			t.Fatalf(".goreleaser.yml should stay focused on the shipped Go binary, found %q", forbidden)
		}
	}
}

func TestCanonicalReleaseMatchesGoReleaserContract(t *testing.T) {
	config := readRepoFile(t, ".goreleaser.yml")
	workflow := readRepoFile(t, ".github/workflows/release.yml")

	release, err := NewCanonicalRelease("1.2.3")
	if err != nil {
		t.Fatalf("NewCanonicalRelease() error = %v", err)
	}

	if got, want := uniqueAssetValues(release.Assets, func(asset CanonicalReleaseAsset) string {
		return asset.GOOS
	}), yamlList(config, "goos"); !reflect.DeepEqual(got, want) {
		t.Fatalf("CanonicalRelease goos = %v, want %v", got, want)
	}
	if got, want := uniqueAssetValues(release.Assets, func(asset CanonicalReleaseAsset) string {
		return asset.GOARCH
	}), yamlList(config, "goarch"); !reflect.DeepEqual(got, want) {
		t.Fatalf("CanonicalRelease goarch = %v, want %v", got, want)
	}

	if got, want := release.ChecksumManifest.FileName, "optimusctx_1.2.3_checksums.txt"; got != want {
		t.Fatalf("CanonicalRelease checksum file = %q, want %q", got, want)
	}
	if !strings.Contains(config, `name_template: "{{ .ProjectName }}_{{ .Version }}_checksums.txt"`) {
		t.Fatalf(".goreleaser.yml must keep the canonical checksum manifest template")
	}

	if got, want := release.Repository.Owner, "NicoMoralesDev"; got != want {
		t.Fatalf("CanonicalRelease repository owner = %q, want %q", got, want)
	}
	if got, want := release.Repository.Name, "optimusctx"; got != want {
		t.Fatalf("CanonicalRelease repository name = %q, want %q", got, want)
	}
	if got, want := release.ReleaseURL, "https://github.com/NicoMoralesDev/optimusctx/releases/tag/v1.2.3"; got != want {
		t.Fatalf("CanonicalRelease ReleaseURL = %q, want %q", got, want)
	}

	for _, asset := range release.Assets {
		wantURLPrefix := "https://github.com/NicoMoralesDev/optimusctx/releases/download/v1.2.3/"
		if !strings.HasPrefix(asset.DownloadURL, wantURLPrefix) {
			t.Fatalf("CanonicalRelease asset URL %q must use %q", asset.DownloadURL, wantURLPrefix)
		}
		if got, want := asset.FileName, archiveName(release.Version, asset.GOOS, asset.GOARCH); got != want {
			t.Fatalf("CanonicalRelease asset file = %q, want %q", got, want)
		}
		if got, want := asset.ArchiveFormat, archiveFormat(asset.GOOS); got != want {
			t.Fatalf("CanonicalRelease asset format = %q, want %q", got, want)
		}
	}

	if !strings.Contains(workflow, `ref=refs/tags/$INPUT_TAG`) {
		t.Fatalf(".github/workflows/release.yml must keep existing-tag release reuse")
	}
	if !strings.Contains(workflow, `canonical GitHub Release`) {
		t.Fatalf(".github/workflows/release.yml must describe GitHub Release as canonical")
	}
	if !strings.Contains(workflow, `PlanReleaseOrchestrationCreate|TestPlanReleaseOrchestrationReuse`) {
		t.Fatalf(".github/workflows/release.yml must verify orchestration create and reuse contracts")
	}
	if !strings.Contains(workflow, `uses: goreleaser/goreleaser-action@v6`) {
		t.Fatalf(".github/workflows/release.yml must keep GitHub Release publication rooted in GoReleaser")
	}
}

func TestGitHubReleasePublicationConfig(t *testing.T) {
	workflow := readRepoFile(t, ".github/workflows/release.yml")

	for _, want := range []string{
		`tags:`,
		`- "v*"`,
		`workflow_dispatch:`,
		`release_tag:`,
		`publication_channel:`,
		`default: all`,
		`uses: actions/checkout@v4`,
		`uses: actions/setup-go@v5`,
		`uses: goreleaser/goreleaser-action@v6`,
		`args: release --clean`,
		`GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}`,
		`canonical GitHub Release`,
		`go test ./internal/release ./internal/cli -run 'TestArchiveMatrix|TestChecksumManifest|TestGitHubReleasePublicationConfig|TestGitHubReleaseWorkflowReuseContract|TestCanonicalReleaseMatchesGoReleaserContract|TestPlanReleaseOrchestrationCreate|TestPlanReleaseOrchestrationReuse|TestReleaseMetadataInjection'`,
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf(".github/workflows/release.yml missing %q", want)
		}
	}

	if !strings.Contains(workflow, `ref=refs/tags/$INPUT_TAG`) {
		t.Fatalf("manual release dispatch must resolve an existing tag ref")
	}
}

func TestGitHubReleaseWorkflowReuseContract(t *testing.T) {
	workflow := readRepoFile(t, ".github/workflows/release.yml")

	for _, want := range []string{
		`workflow_dispatch:`,
		`release_tag:`,
		`description: Existing v* tag whose canonical GitHub Release contract should be reused`,
		`publication_channel:`,
		`default: all`,
		`Manual workflow_dispatch release_tag reruns must point at an existing`,
		`tag so downstream publication reuses the same canonical GitHub Release`,
		`archives, checksums, and release metadata contract.`,
		`name: Resolve canonical release ref`,
		`name: Verify canonical release contract`,
		`name: Verify workflow_dispatch reuses existing canonical release tag`,
		`name: Publish canonical GitHub Release assets`,
		`github.event_name != 'workflow_dispatch'`,
		`TestGitHubReleaseWorkflowReuseContract`,
		`TestCanonicalReleaseMatchesGoReleaserContract`,
		`TestPlanReleaseOrchestrationCreate`,
		`TestPlanReleaseOrchestrationReuse`,
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf(".github/workflows/release.yml missing %q", want)
		}
	}

	if strings.Contains(workflow, `publish or republish`) {
		t.Fatalf(".github/workflows/release.yml should describe manual reruns as canonical release reuse, not generic republish wording")
	}
}

func TestNPMPublishWorkflow(t *testing.T) {
	workflow := readRepoFile(t, ".github/workflows/release.yml")

	for _, want := range []string{
		"name: Publish npm wrapper package",
		"needs: release",
		"permissions:",
		"id-token: write",
		"inputs.publication_channel == 'npm'",
		"uses: actions/setup-node@v4",
		"node-version: 24",
		"registry-url: https://registry.npmjs.org",
		"bash scripts/render-npm-package.sh",
		"npm publish --access public",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf(".github/workflows/release.yml missing %q", want)
		}
	}
}

func TestChannelPublicationWorkflowFanout(t *testing.T) {
	workflow := readRepoFile(t, ".github/workflows/release.yml")

	for _, want := range []string{
		"name: Publish npm wrapper package",
		"name: Publish Homebrew formula",
		"name: Publish Scoop manifest",
		"github.event_name != 'workflow_dispatch' || inputs.publication_channel == 'all' || inputs.publication_channel == 'npm'",
		"github.event_name != 'workflow_dispatch' || inputs.publication_channel == 'all' || inputs.publication_channel == 'homebrew'",
		"github.event_name != 'workflow_dispatch' || inputs.publication_channel == 'all' || inputs.publication_channel == 'scoop'",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf(".github/workflows/release.yml missing %q", want)
		}
	}
}

func TestChannelPublicationWorkflowSelectiveRerun(t *testing.T) {
	workflow := readRepoFile(t, ".github/workflows/release.yml")

	for _, want := range []string{
		`publication_channel:`,
		`release_tag:`,
		`default: all`,
		`type: choice`,
		`- all`,
		`- npm`,
		`- homebrew`,
		`- scoop`,
		`gh release view "$RELEASE_TAG" --repo "$GITHUB_REPOSITORY" >/dev/null`,
		`skipping goreleaser release --clean`,
		`inputs.publication_channel == 'npm'`,
		`inputs.publication_channel == 'homebrew'`,
		`inputs.publication_channel == 'scoop'`,
		`publication_channel=npm`,
		`publication_channel=homebrew`,
		`publication_channel=scoop`,
		`ref: ${{ needs.release.outputs.ref }}`,
		`${{ needs.release.outputs.tag }}`,
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf(".github/workflows/release.yml missing %q", want)
		}
	}
}

func TestReleaseWorkflowSummaryShowsChannelStatus(t *testing.T) {
	workflow := readRepoFile(t, ".github/workflows/release.yml")

	for _, want := range []string{
		`### GitHub Release publication`,
		`### npm publication`,
		`### Homebrew publication`,
		`### Scoop publication`,
		`- channel:`,
		`- tag:`,
		`- outcome:`,
		`- failure_reason:`,
		`- next_step:`,
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf(".github/workflows/release.yml missing %q", want)
		}
	}
}

func TestReleaseWorkflowSummaryShowsFailureGuidance(t *testing.T) {
	workflow := readRepoFile(t, ".github/workflows/release.yml")

	for _, want := range []string{
		`GitHub Release archive publication failed`,
		`fix the canonical GitHub Release/archive state before any downstream rerun`,
		`npm publish failed`,
		`Homebrew tap update failed`,
		`Scoop bucket update failed`,
		`confirm GitHub Release remains the canonical root`,
		`workflow_dispatch with release_tag=${RELEASE_TAG} and publication_channel=npm`,
		`workflow_dispatch with release_tag=${RELEASE_TAG} and publication_channel=homebrew`,
		`workflow_dispatch with release_tag=${RELEASE_TAG} and publication_channel=scoop`,
		`publication_channel=npm`,
		`publication_channel=homebrew`,
		`publication_channel=scoop`,
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf(".github/workflows/release.yml missing %q", want)
		}
	}

	if strings.Contains(workflow, `publication_channel=github-release`) {
		t.Fatalf(".github/workflows/release.yml must not invent publication_channel=github-release")
	}
}

func TestHomebrewPublishWorkflow(t *testing.T) {
	workflow := readRepoFile(t, ".github/workflows/release.yml")

	for _, want := range []string{
		"name: Publish Homebrew formula",
		"repository: niccrow/homebrew-tap",
		"HOMEBREW_TAP_GITHUB_TOKEN",
		"bash scripts/render-homebrew-formula.sh",
		"Formula/optimusctx.rb",
		"GITHUB_STEP_SUMMARY",
		"publication_channel=homebrew",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf(".github/workflows/release.yml missing %q", want)
		}
	}
}

func TestScoopPublishWorkflow(t *testing.T) {
	workflow := readRepoFile(t, ".github/workflows/release.yml")

	for _, want := range []string{
		"name: Publish Scoop manifest",
		"repository: niccrow/scoop-bucket",
		"SCOOP_BUCKET_GITHUB_TOKEN",
		"bash scripts/render-scoop-manifest.sh",
		"bucket/optimusctx.json",
		"GITHUB_STEP_SUMMARY",
		"publication_channel=scoop",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf(".github/workflows/release.yml missing %q", want)
		}
	}
}

func TestNPMPublishConfig(t *testing.T) {
	workflow := readRepoFile(t, ".github/workflows/release.yml")
	renderScript := readRepoFile(t, "scripts/render-npm-package.sh")

	for _, want := range []string{
		"NPM_TOKEN",
		"NODE_AUTH_TOKEN",
		"id-token: write",
		"needs.release.outputs.tag",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf(".github/workflows/release.yml missing %q", want)
		}
	}

	for _, want := range []string{
		"go run ./cmd/render-npm-package",
		"--release-tag",
		"--package-json",
		"@niccrow/optimusctx",
	} {
		if !strings.Contains(renderScript, want) {
			t.Fatalf("scripts/render-npm-package.sh missing %q", want)
		}
	}

	for _, forbidden := range []string{
		"retagCanonicalURL",
		"expectedArchive",
		"RELEASE_TAG",
		"checksumManifest.url",
	} {
		if strings.Contains(renderScript, forbidden) {
			t.Fatalf("scripts/render-npm-package.sh should not contain %q", forbidden)
		}
	}
}

func TestCanonicalReleaseFeedsDownstreamConsumers(t *testing.T) {
	canonicalRelease, err := NewCanonicalRelease("1.2.3")
	if err != nil {
		t.Fatalf("NewCanonicalRelease() error = %v", err)
	}

	npmRelease, err := newNPMPackageRelease(canonicalRelease.Version)
	if err != nil {
		t.Fatalf("newNPMPackageRelease() error = %v", err)
	}
	packageManagerRelease, err := newPackageManagerRelease(canonicalRelease.Version, sampleChecksumManifest(canonicalRelease.Version))
	if err != nil {
		t.Fatalf("newPackageManagerRelease() error = %v", err)
	}

	if got, want := npmRelease.ReleaseTag, canonicalRelease.Tag; got != want {
		t.Fatalf("npm ReleaseTag = %q, want %q", got, want)
	}
	if got, want := npmRelease.ChecksumManifest.URL, canonicalRelease.ChecksumManifest.URL; got != want {
		t.Fatalf("npm ChecksumManifest.URL = %q, want %q", got, want)
	}
	if got, want := packageManagerRelease.ChecksumManifest.URL, canonicalRelease.ChecksumManifest.URL; got != want {
		t.Fatalf("package-manager ChecksumManifest.URL = %q, want %q", got, want)
	}

	windowsAMD64Asset, err := canonicalRelease.Asset("windows", "amd64")
	if err != nil {
		t.Fatalf("Asset(windows, amd64) error = %v", err)
	}
	if got, want := npmRelease.Platforms.WindowsAMD64.ArchiveURL, windowsAMD64Asset.DownloadURL; got != want {
		t.Fatalf("npm windows archive URL = %q, want %q", got, want)
	}
	if got, want := packageManagerRelease.Assets.WindowsAMD64.URL, windowsAMD64Asset.DownloadURL; got != want {
		t.Fatalf("package-manager windows archive URL = %q, want %q", got, want)
	}

	outputDir := t.TempDir()
	scriptPath := filepath.Join("scripts", "render-npm-package.sh")
	cmd := exec.Command("bash", scriptPath, canonicalRelease.Tag, outputDir)
	cmd.Dir = filepath.Join("..", "..")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("render-npm-package.sh error = %v\n%s", err, output)
	}

	renderedManifest, err := os.ReadFile(filepath.Join(outputDir, "package.json"))
	if err != nil {
		t.Fatalf("ReadFile(rendered package.json) error = %v", err)
	}

	expectedManifest, err := RenderNPMPackageManifestForTag(canonicalRelease.Tag)
	if err != nil {
		t.Fatalf("RenderNPMPackageManifestForTag() error = %v", err)
	}
	if string(renderedManifest) != expectedManifest {
		t.Fatalf("rendered npm package manifest drifted from canonical release manifest\nwant:\n%s\ngot:\n%s", expectedManifest, renderedManifest)
	}
}

func TestReleasePrerequisiteFiles(t *testing.T) {
	for _, path := range []string{
		".goreleaser.yml",
		".github/workflows/release.yml",
		"scripts/render-npm-package.sh",
		"packaging/homebrew/optimusctx.rb.tmpl",
		"packaging/scoop/optimusctx.json.tmpl",
		"packaging/npm/package.json",
	} {
		if content := readRepoFile(t, path); strings.TrimSpace(content) == "" {
			t.Fatalf("%s must not be empty", path)
		}
	}
}

func TestReleaseChecklistPublicationCredentials(t *testing.T) {
	checklist := readRepoFile(t, "docs/release-checklist.md")

	for _, want := range []string{
		"HOMEBREW_TAP_GITHUB_TOKEN",
		"SCOOP_BUCKET_GITHUB_TOKEN",
		"NPM_TOKEN",
		"niccrow/homebrew-tap",
		"niccrow/scoop-bucket",
		"@niccrow/optimusctx",
	} {
		if !strings.Contains(checklist, want) {
			t.Fatalf("docs/release-checklist.md missing %q", want)
		}
	}
}

func TestMultiChannelPublicationDocsStayCanonical(t *testing.T) {
	checklist := readRepoFile(t, "docs/release-checklist.md")
	installGuide := readRepoFile(t, "docs/install-and-verify.md")

	for _, want := range []string{
		`GitHub Release is the canonical root for archives, checksums, and downstream release facts.`,
		`After GitHub Release assets are available, npm, Homebrew, and Scoop are published from the same canonical tagged release contract.`,
		`Use ` + "`workflow_dispatch`" + ` with ` + "`release_tag`" + ` and ` + "`publication_channel`" + ` to rerun ` + "`npm`" + `, ` + "`homebrew`" + `, or ` + "`scoop`" + ` for an existing tagged release without rebuilding unrelated channels.`,
		`Confirm Homebrew publication is automated from the same canonical tagged release after GitHub Release assets are available.`,
		`Confirm Scoop publication is automated from the same canonical tagged release after GitHub Release assets are available.`,
		`canonical tagged GitHub Release binary`,
		`rollback source`,
	} {
		if !strings.Contains(checklist, want) {
			t.Fatalf("docs/release-checklist.md missing %q", want)
		}
	}

	for _, want := range []string{
		`GitHub Release is the canonical root for release archives, checksum manifests, and downstream channel facts.`,
		`After GitHub Release assets are available, npm, Homebrew, and Scoop are published from the same canonical tagged release contract.`,
		`The npm package is a wrapper over the canonical tagged GitHub Release binary.`,
		`Homebrew installs the formula rendered from the same canonical tagged GitHub Release checksum and archive contract.`,
		`Scoop installs the manifest rendered from the same canonical tagged GitHub Release checksum and archive contract.`,
		`Download the archive that matches your OS and CPU from the canonical tagged GitHub Release.`,
		`GitHub Release is the canonical root and rollback source even when downstream automation republishes one package-manager channel.`,
		`without rebuilding unrelated channels`,
	} {
		if !strings.Contains(installGuide, want) {
			t.Fatalf("docs/install-and-verify.md missing %q", want)
		}
	}
}

func TestOperatorReleaseGuideStaysCanonical(t *testing.T) {
	guide := readRepoFile(t, "docs/operator-release-guide.md")

	for _, want := range []string{
		`GitHub Release remains the canonical root and rollback source.`,
		`Downstream reruns reuse the existing tag via ` + "`workflow_dispatch`" + ` with ` + "`release_tag`" + ` and ` + "`publication_channel`" + `.`,
		`GitHub Release remains the canonical root and rollback source even when npm, Homebrew, or Scoop were the failing channels.`,
		`optimusctx release prepare`,
		`optimusctx release prepare --confirm`,
		`gh release view "$TAG"`,
		`gh release download "$TAG" --dir /tmp/optimusctx-release-check`,
		`@niccrow/optimusctx`,
		`niccrow/homebrew-tap`,
		`niccrow/scoop-bucket`,
		`optimusctx version`,
		`optimusctx doctor`,
		`optimusctx snippet`,
	} {
		if !strings.Contains(guide, want) {
			t.Fatalf("docs/operator-release-guide.md missing %q", want)
		}
	}
}

func yamlList(content, key string) []string {
	lines := strings.Split(content, "\n")
	needle := key + ":"

	for index, line := range lines {
		if strings.TrimSpace(line) != needle {
			continue
		}

		baseIndent := leadingSpaces(line)
		var values []string
		for _, next := range lines[index+1:] {
			trimmed := strings.TrimSpace(next)
			if trimmed == "" {
				continue
			}

			nextIndent := leadingSpaces(next)
			if nextIndent <= baseIndent {
				break
			}

			if strings.HasPrefix(trimmed, "- ") {
				values = append(values, strings.TrimPrefix(trimmed, "- "))
				continue
			}

			if len(values) > 0 {
				break
			}
		}

		return values
	}

	return nil
}

func leadingSpaces(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

func uniqueAssetValues(assets []CanonicalReleaseAsset, project func(CanonicalReleaseAsset) string) []string {
	values := make(map[string]struct{}, len(assets))
	for _, asset := range assets {
		values[project(asset)] = struct{}{}
	}

	ordered := make([]string, 0, len(values))
	for value := range values {
		ordered = append(ordered, value)
	}
	sort.Strings(ordered)

	return ordered
}

func readRepoFile(t *testing.T, relPath string) string {
	t.Helper()

	path := filepath.Join("..", "..", relPath)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	return string(content)
}
