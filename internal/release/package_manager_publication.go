package release

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func RenderHomebrewFormulaForTag(releaseTag string, checksumManifest string) (string, error) {
	packageRelease, err := newPackageManagerReleaseForTag(releaseTag, checksumManifest)
	if err != nil {
		return "", err
	}

	templateText, err := readPublicationTemplate(filepath.Join("packaging", "homebrew", "optimusctx.rb.tmpl"))
	if err != nil {
		return "", err
	}
	return renderHomebrewFormula(templateText, packageRelease, defaultHomebrewTapTarget())
}

func RenderScoopManifestForTag(releaseTag string, checksumManifest string) (string, error) {
	packageRelease, err := newPackageManagerReleaseForTag(releaseTag, checksumManifest)
	if err != nil {
		return "", err
	}

	templateText, err := readPublicationTemplate(filepath.Join("packaging", "scoop", "optimusctx.json.tmpl"))
	if err != nil {
		return "", err
	}
	return renderScoopManifest(templateText, packageRelease, defaultScoopBucketTarget())
}

func newPackageManagerReleaseForTag(releaseTag string, checksumManifest string) (packageManagerRelease, error) {
	canonicalRelease, err := canonicalReleaseForPackageManagerTag(releaseTag)
	if err != nil {
		return packageManagerRelease{}, err
	}

	checksums, err := parseChecksumManifest(checksumManifest)
	if err != nil {
		return packageManagerRelease{}, err
	}

	return newPackageManagerReleaseFromCanonical(canonicalRelease, checksums)
}

func canonicalReleaseForPackageManagerTag(releaseTag string) (CanonicalRelease, error) {
	normalizedTag, err := NormalizeReleaseTag(releaseTag)
	if err != nil {
		return CanonicalRelease{}, err
	}

	return NewCanonicalRelease(strings.TrimPrefix(normalizedTag, "v"))
}

func readPublicationTemplate(path string) (string, error) {
	root, err := repositoryRootForRender()
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(filepath.Join(root, path))
	if err != nil {
		return "", fmt.Errorf("read publication template %s: %w", path, err)
	}
	return string(content), nil
}

func repositoryRootForRender() (string, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("resolve release package path")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..")), nil
}
