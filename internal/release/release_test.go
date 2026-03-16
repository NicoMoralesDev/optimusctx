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

	if strings.Contains(config, "npx") {
		t.Fatalf(".goreleaser.yml should stay focused on the shipped Go binary")
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
