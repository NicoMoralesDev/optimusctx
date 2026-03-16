package release

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

const (
	canonicalReleaseOwner      = "niccrow"
	canonicalReleaseRepo       = "optimusctx"
	canonicalProjectName       = "optimusctx"
	canonicalHomepage          = "https://github.com/niccrow/optimusctx"
	canonicalDescription       = "Local-first runtime that builds and maintains persistent repository context for coding agents."
	canonicalLicense           = "MIT"
	homebrewTapRepo            = "homebrew-tap"
	homebrewTapDirectory       = "Formula"
	homebrewTapBranch          = "main"
	homebrewTapTokenEnv        = "HOMEBREW_TAP_GITHUB_TOKEN"
	canonicalBinaryName        = "optimusctx"
	canonicalWindowsBinaryName = "optimusctx.exe"
	canonicalChecksumDelimiter = "  "
)

type repositoryRef struct {
	Owner string
	Name  string
}

type releaseAsset struct {
	FileName string
	URL      string
	SHA256   string
}

type packageManagerRelease struct {
	Version     string
	Tag         string
	ProjectName string
	Homepage    string
	Description string
	License     string
	Repository  repositoryRef
	Assets      packageManagerAssets
}

type packageManagerAssets struct {
	DarwinAMD64  releaseAsset
	DarwinARM64  releaseAsset
	LinuxAMD64   releaseAsset
	LinuxARM64   releaseAsset
	WindowsAMD64 releaseAsset
	WindowsARM64 releaseAsset
}

type homebrewTapTarget struct {
	Repository   repositoryRef
	Branch       string
	Directory    string
	FormulaName  string
	TokenEnvVar  string
	InstallTap   string
	InstallValue string
}

func newPackageManagerRelease(version string, checksumManifest string) (packageManagerRelease, error) {
	version = strings.TrimSpace(version)
	if version == "" {
		return packageManagerRelease{}, fmt.Errorf("version is required")
	}

	checksums, err := parseChecksumManifest(checksumManifest)
	if err != nil {
		return packageManagerRelease{}, err
	}

	release := packageManagerRelease{
		Version:     version,
		Tag:         "v" + version,
		ProjectName: canonicalProjectName,
		Homepage:    canonicalHomepage,
		Description: canonicalDescription,
		License:     canonicalLicense,
		Repository: repositoryRef{
			Owner: canonicalReleaseOwner,
			Name:  canonicalReleaseRepo,
		},
		Assets: packageManagerAssets{},
	}

	for _, target := range []struct {
		assign *releaseAsset
		goos   string
		goarch string
	}{
		{assign: &release.Assets.DarwinAMD64, goos: "darwin", goarch: "amd64"},
		{assign: &release.Assets.DarwinARM64, goos: "darwin", goarch: "arm64"},
		{assign: &release.Assets.LinuxAMD64, goos: "linux", goarch: "amd64"},
		{assign: &release.Assets.LinuxARM64, goos: "linux", goarch: "arm64"},
		{assign: &release.Assets.WindowsAMD64, goos: "windows", goarch: "amd64"},
		{assign: &release.Assets.WindowsARM64, goos: "windows", goarch: "arm64"},
	} {
		asset, err := release.archiveFor(target.goos, target.goarch, checksums)
		if err != nil {
			return packageManagerRelease{}, err
		}
		*target.assign = asset
	}

	return release, nil
}

func defaultHomebrewTapTarget() homebrewTapTarget {
	return homebrewTapTarget{
		Repository: repositoryRef{
			Owner: canonicalReleaseOwner,
			Name:  homebrewTapRepo,
		},
		Branch:       homebrewTapBranch,
		Directory:    homebrewTapDirectory,
		FormulaName:  canonicalBinaryName,
		TokenEnvVar:  homebrewTapTokenEnv,
		InstallTap:   fmt.Sprintf("%s/tap", canonicalReleaseOwner),
		InstallValue: fmt.Sprintf("%s/tap/%s", canonicalReleaseOwner, canonicalBinaryName),
	}
}

func renderHomebrewFormula(templateText string, release packageManagerRelease, target homebrewTapTarget) (string, error) {
	renderer, err := template.New("homebrew").Parse(templateText)
	if err != nil {
		return "", fmt.Errorf("parse homebrew template: %w", err)
	}

	var output bytes.Buffer
	data := struct {
		Release     packageManagerRelease
		Publication homebrewTapTarget
	}{
		Release:     release,
		Publication: target,
	}

	if err := renderer.Execute(&output, data); err != nil {
		return "", fmt.Errorf("render homebrew template: %w", err)
	}

	return output.String(), nil
}

func (r packageManagerRelease) archiveFor(goos, goarch string, checksums map[string]string) (releaseAsset, error) {
	fileName := archiveName(r.Version, goos, goarch)
	sha256, ok := checksums[fileName]
	if !ok {
		return releaseAsset{}, fmt.Errorf("checksum missing for %s", fileName)
	}

	return releaseAsset{
		FileName: fileName,
		URL:      fmt.Sprintf("%s/releases/download/%s/%s", r.githubRepositoryURL(), r.Tag, fileName),
		SHA256:   sha256,
	}, nil
}

func (r packageManagerRelease) githubRepositoryURL() string {
	return fmt.Sprintf("https://github.com/%s/%s", r.Repository.Owner, r.Repository.Name)
}

func archiveName(version, goos, goarch string) string {
	format := "tar.gz"
	if goos == "windows" {
		format = "zip"
	}
	return fmt.Sprintf("%s_%s_%s_%s.%s", canonicalProjectName, version, goos, goarch, format)
}

func parseChecksumManifest(content string) (map[string]string, error) {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) == 0 || (len(lines) == 1 && strings.TrimSpace(lines[0]) == "") {
		return nil, fmt.Errorf("checksum manifest is empty")
	}

	checksums := make(map[string]string, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.SplitN(line, canonicalChecksumDelimiter, 2)
		if len(fields) != 2 || strings.TrimSpace(fields[0]) == "" || strings.TrimSpace(fields[1]) == "" {
			return nil, fmt.Errorf("invalid checksum manifest line %q", line)
		}
		checksums[strings.TrimSpace(fields[1])] = strings.TrimSpace(fields[0])
	}

	return checksums, nil
}
