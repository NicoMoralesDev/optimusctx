package release

import (
	"os"
	"path/filepath"
	"reflect"
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
		`go test ./internal/release ./internal/cli -run 'TestArchiveMatrix|TestChecksumManifest|TestGitHubReleasePublicationConfig|TestReleaseMetadataInjection'`,
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf(".github/workflows/release.yml missing %q", want)
		}
	}

	if !strings.Contains(workflow, `ref=refs/tags/$INPUT_TAG`) {
		t.Fatalf("manual release dispatch must resolve an existing tag ref")
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
		"package.json",
		"optimusctx_${versionNoV}_${goos}_${goarch}",
		"npm package metadata must stay release-derived",
		"@niccrow/optimusctx",
	} {
		if !strings.Contains(renderScript, want) {
			t.Fatalf("scripts/render-npm-package.sh missing %q", want)
		}
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

func readRepoFile(t *testing.T, relPath string) string {
	t.Helper()

	path := filepath.Join("..", "..", relPath)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	return string(content)
}
