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
		`owner: niccrow`,
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

	if got, want := release.Repository.Owner, "niccrow"; got != want {
		t.Fatalf("CanonicalRelease repository owner = %q, want %q", got, want)
	}
	if got, want := release.Repository.Name, "optimusctx"; got != want {
		t.Fatalf("CanonicalRelease repository name = %q, want %q", got, want)
	}
	if got, want := release.ReleaseURL, "https://github.com/niccrow/optimusctx/releases/tag/v1.2.3"; got != want {
		t.Fatalf("CanonicalRelease ReleaseURL = %q, want %q", got, want)
	}

	for _, asset := range release.Assets {
		wantURLPrefix := "https://github.com/niccrow/optimusctx/releases/download/v1.2.3/"
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
		`Manual workflow_dispatch release_tag reruns must point at an existing`,
		`tag so downstream publication reuses the same canonical GitHub Release`,
		`archives, checksums, and release metadata contract.`,
		`name: Resolve canonical release ref`,
		`name: Verify canonical release contract`,
		`name: Publish canonical GitHub Release assets`,
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
		"uses: actions/setup-node@v4",
		"registry-url: https://registry.npmjs.org",
		"bash scripts/render-npm-package.sh",
		"npm publish --access public",
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
		"steps.release_ref.outputs.tag",
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

func TestGitHubReleaseDocsStayCanonical(t *testing.T) {
	checklist := readRepoFile(t, "docs/release-checklist.md")
	installGuide := readRepoFile(t, "docs/install-and-verify.md")

	for _, want := range []string{
		`GitHub Release is the canonical root for archives, checksums, and downstream release facts.`,
		`Use ` + "`workflow_dispatch`" + ` with ` + "`release_tag`" + ` when you need to reuse an existing tagged release contract for reruns or downstream publication recovery.`,
		`Do not claim Homebrew publication is automated in Phase 17.`,
		`Do not claim Scoop publication is automated in Phase 17.`,
		`canonical tagged GitHub Release binary`,
	} {
		if !strings.Contains(checklist, want) {
			t.Fatalf("docs/release-checklist.md missing %q", want)
		}
	}

	for _, want := range []string{
		`GitHub Release is the canonical root for release archives, checksum manifests, and downstream channel facts.`,
		`The npm package is a wrapper over the canonical tagged GitHub Release binary.`,
		`Download the archive that matches your OS and CPU from the canonical tagged GitHub Release.`,
		`this phase does not yet claim automated Homebrew or Scoop publication fan-out`,
	} {
		if !strings.Contains(installGuide, want) {
			t.Fatalf("docs/install-and-verify.md missing %q", want)
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
