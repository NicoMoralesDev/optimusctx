package release

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	canonicalNPMPackageName        = "@niccrow/optimusctx"
	canonicalNPMBinCommand         = canonicalBinaryName
	canonicalNPMBinPath            = "bin/optimusctx.js"
	canonicalNPMPostInstallScript  = "node ./lib/install.js"
	canonicalNPMMinimumNodeVersion = ">=18"
	canonicalNPMDevelopmentVersion = "0.0.0-development"
	canonicalNPMRepositoryURL      = "git+https://github.com/niccrow/optimusctx.git"
	canonicalNPMIssuesURL          = "https://github.com/niccrow/optimusctx/issues"
	canonicalNPMLauncherModule     = "./bin/optimusctx.js"
)

type npmPackageRelease struct {
	PackageName      string
	Version          string
	Description      string
	License          string
	Homepage         string
	BinCommand       string
	BinPath          string
	PostInstall      string
	MinimumNode      string
	Repository       repositoryRef
	ReleaseTag       string
	RepositoryURL    string
	IssuesURL        string
	ChecksumManifest npmChecksumManifest
	Platforms        npmPlatformAssets
}

type npmChecksumManifest struct {
	FileName string `json:"file"`
	URL      string `json:"url"`
}

type npmPlatformAsset struct {
	OS               string `json:"os"`
	Arch             string `json:"arch"`
	ArchiveFileName  string `json:"archive"`
	ArchiveURL       string `json:"archiveUrl"`
	ArchiveFormat    string `json:"archiveFormat"`
	RuntimeBinary    string `json:"binary"`
	RuntimeDirectory string `json:"runtimeDirectory"`
}

type npmPlatformAssets struct {
	DarwinAMD64  npmPlatformAsset `json:"darwin-amd64"`
	DarwinARM64  npmPlatformAsset `json:"darwin-arm64"`
	LinuxAMD64   npmPlatformAsset `json:"linux-amd64"`
	LinuxARM64   npmPlatformAsset `json:"linux-arm64"`
	WindowsAMD64 npmPlatformAsset `json:"windows-amd64"`
	WindowsARM64 npmPlatformAsset `json:"windows-arm64"`
}

type npmPackageManifest struct {
	Name        string                    `json:"name"`
	Version     string                    `json:"version"`
	Description string                    `json:"description"`
	License     string                    `json:"license"`
	Homepage    string                    `json:"homepage"`
	Repository  npmRepositoryMetadata     `json:"repository"`
	Bugs        npmBugsMetadata           `json:"bugs"`
	Engines     npmEngineMetadata         `json:"engines"`
	Bin         map[string]string         `json:"bin"`
	Scripts     npmScriptMetadata         `json:"scripts"`
	Files       []string                  `json:"files"`
	OptimusCtx  npmPackageRuntimeMetadata `json:"optimusctx"`
}

type npmRepositoryMetadata struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type npmBugsMetadata struct {
	URL string `json:"url"`
}

type npmEngineMetadata struct {
	Node string `json:"node"`
}

type npmScriptMetadata struct {
	PostInstall string `json:"postinstall"`
}

type npmPackageRuntimeMetadata struct {
	Command          string              `json:"command"`
	Launcher         string              `json:"launcher"`
	ProjectName      string              `json:"projectName"`
	Repository       npmRepositoryTarget `json:"repository"`
	ReleaseTag       string              `json:"releaseTag"`
	Version          string              `json:"version"`
	ChecksumManifest npmChecksumManifest `json:"checksumManifest"`
	Platforms        npmPlatformAssets   `json:"platforms"`
}

type npmRepositoryTarget struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

func newNPMPackageRelease(version string) (npmPackageRelease, error) {
	version = strings.TrimSpace(version)
	if version == "" {
		return npmPackageRelease{}, fmt.Errorf("version is required")
	}

	release := npmPackageRelease{
		PackageName: canonicalNPMPackageName,
		Version:     version,
		Description: canonicalDescription,
		License:     canonicalLicense,
		Homepage:    canonicalHomepage,
		BinCommand:  canonicalNPMBinCommand,
		BinPath:     canonicalNPMBinPath,
		PostInstall: canonicalNPMPostInstallScript,
		MinimumNode: canonicalNPMMinimumNodeVersion,
		Repository: repositoryRef{
			Owner: canonicalReleaseOwner,
			Name:  canonicalReleaseRepo,
		},
		ReleaseTag:    "v" + version,
		RepositoryURL: canonicalNPMRepositoryURL,
		IssuesURL:     canonicalNPMIssuesURL,
		ChecksumManifest: npmChecksumManifest{
			FileName: checksumManifestName(version),
		},
	}
	release.ChecksumManifest.URL = release.releaseDownloadURL(release.ChecksumManifest.FileName)

	for _, target := range []struct {
		assign *npmPlatformAsset
		goos   string
		goarch string
	}{
		{assign: &release.Platforms.DarwinAMD64, goos: "darwin", goarch: "amd64"},
		{assign: &release.Platforms.DarwinARM64, goos: "darwin", goarch: "arm64"},
		{assign: &release.Platforms.LinuxAMD64, goos: "linux", goarch: "amd64"},
		{assign: &release.Platforms.LinuxARM64, goos: "linux", goarch: "arm64"},
		{assign: &release.Platforms.WindowsAMD64, goos: "windows", goarch: "amd64"},
		{assign: &release.Platforms.WindowsARM64, goos: "windows", goarch: "arm64"},
	} {
		*target.assign = release.platformAsset(target.goos, target.goarch)
	}

	return release, nil
}

func renderNPMPackageManifest(release npmPackageRelease) (string, error) {
	manifest := npmPackageManifest{
		Name:        release.PackageName,
		Version:     release.Version,
		Description: release.Description,
		License:     release.License,
		Homepage:    release.Homepage,
		Repository: npmRepositoryMetadata{
			Type: "git",
			URL:  release.RepositoryURL,
		},
		Bugs: npmBugsMetadata{
			URL: release.IssuesURL,
		},
		Engines: npmEngineMetadata{
			Node: release.MinimumNode,
		},
		Bin: map[string]string{
			release.BinCommand: release.BinPath,
		},
		Scripts: npmScriptMetadata{
			PostInstall: release.PostInstall,
		},
		Files: []string{
			"bin",
			"lib",
		},
		OptimusCtx: npmPackageRuntimeMetadata{
			Command:     release.BinCommand,
			Launcher:    canonicalNPMLauncherModule,
			ProjectName: canonicalProjectName,
			Repository: npmRepositoryTarget{
				Owner: release.Repository.Owner,
				Name:  release.Repository.Name,
			},
			ReleaseTag:       release.ReleaseTag,
			Version:          release.Version,
			ChecksumManifest: release.ChecksumManifest,
			Platforms:        release.Platforms,
		},
	}

	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(manifest); err != nil {
		return "", fmt.Errorf("marshal npm package manifest: %w", err)
	}

	return output.String(), nil
}

func renderCommittedNPMPackageManifest() (string, error) {
	release, err := newNPMPackageRelease(canonicalNPMDevelopmentVersion)
	if err != nil {
		return "", err
	}

	return renderNPMPackageManifest(release)
}

func checksumManifestName(version string) string {
	return fmt.Sprintf("%s_%s_checksums.txt", canonicalProjectName, version)
}

func (r npmPackageRelease) releaseDownloadURL(fileName string) string {
	return fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", r.Repository.Owner, r.Repository.Name, r.ReleaseTag, fileName)
}

func (r npmPackageRelease) platformAsset(goos, goarch string) npmPlatformAsset {
	fileName := archiveName(r.Version, goos, goarch)

	return npmPlatformAsset{
		OS:               goos,
		Arch:             goarch,
		ArchiveFileName:  fileName,
		ArchiveURL:       r.releaseDownloadURL(fileName),
		ArchiveFormat:    archiveFormat(goos),
		RuntimeBinary:    runtimeBinaryName(goos),
		RuntimeDirectory: runtimeDirectoryName(goos, goarch),
	}
}

func archiveFormat(goos string) string {
	if goos == "windows" {
		return "zip"
	}

	return "tar.gz"
}

func runtimeBinaryName(goos string) string {
	if goos == "windows" {
		return canonicalWindowsBinaryName
	}

	return canonicalBinaryName
}

func runtimeDirectoryName(goos, goarch string) string {
	return fmt.Sprintf("%s-%s", goos, goarch)
}
