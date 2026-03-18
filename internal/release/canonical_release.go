package release

import "fmt"

type CanonicalRelease struct {
	Version          string
	Tag              string
	ProjectName      string
	Repository       repositoryRef
	ReleaseURL       string
	ChecksumManifest CanonicalChecksumManifest
	Assets           []CanonicalReleaseAsset
}

type CanonicalReleaseTarget struct {
	GOOS          string
	GOARCH        string
	ArchiveFormat string
}

type CanonicalReleaseAsset struct {
	GOOS          string
	GOARCH        string
	FileName      string
	DownloadURL   string
	ArchiveFormat string
}

type CanonicalChecksumManifest struct {
	FileName string
	URL      string
}

var canonicalReleaseTargetInventory = []CanonicalReleaseTarget{
	{
		GOOS:          "darwin",
		GOARCH:        "amd64",
		ArchiveFormat: "tar.gz",
	},
	{
		GOOS:          "darwin",
		GOARCH:        "arm64",
		ArchiveFormat: "tar.gz",
	},
	{
		GOOS:          "linux",
		GOARCH:        "amd64",
		ArchiveFormat: "tar.gz",
	},
	{
		GOOS:          "linux",
		GOARCH:        "arm64",
		ArchiveFormat: "tar.gz",
	},
	{
		GOOS:          "windows",
		GOARCH:        "amd64",
		ArchiveFormat: "zip",
	},
	{
		GOOS:          "windows",
		GOARCH:        "arm64",
		ArchiveFormat: "zip",
	},
}

func NewCanonicalRelease(version string) (CanonicalRelease, error) {
	normalizedVersion, err := NormalizeReleaseVersion(version)
	if err != nil {
		return CanonicalRelease{}, err
	}

	return newCanonicalRelease(normalizedVersion), nil
}

func newCanonicalRelease(version string) CanonicalRelease {
	release := CanonicalRelease{
		Version:     version,
		Tag:         canonicalReleaseTag(version),
		ProjectName: canonicalProjectName,
		Repository: repositoryRef{
			Owner: canonicalReleaseRepoOwner,
			Name:  canonicalReleaseRepo,
		},
	}

	release.ReleaseURL = release.releaseTagURL()
	release.ChecksumManifest = release.checksumManifestContract()
	release.Assets = release.assetsFromSupportedTargets()

	return release
}

func (r CanonicalRelease) Targets() []CanonicalReleaseTarget {
	targets := make([]CanonicalReleaseTarget, len(canonicalReleaseTargetInventory))
	copy(targets, canonicalReleaseTargetInventory)
	return targets
}

func (r CanonicalRelease) AssetKey(goos, goarch string) string {
	return canonicalReleaseAssetKey(goos, goarch)
}

func (r CanonicalRelease) ChecksumManifestURL() string {
	return r.ChecksumManifest.URL
}

func (r CanonicalRelease) ArchiveFileNames() []string {
	fileNames := make([]string, 0, len(r.Targets()))
	for _, target := range r.Targets() {
		fileNames = append(fileNames, archiveName(r.Version, target.GOOS, target.GOARCH))
	}

	return fileNames
}

func (r CanonicalRelease) Asset(goos, goarch string) (CanonicalReleaseAsset, error) {
	key := r.AssetKey(goos, goarch)
	for _, asset := range r.Assets {
		if r.AssetKey(asset.GOOS, asset.GOARCH) == key {
			return asset, nil
		}
	}

	return CanonicalReleaseAsset{}, fmt.Errorf("canonical release asset %s not found", key)
}

func (r CanonicalRelease) RepositoryURL() string {
	return fmt.Sprintf("https://github.com/%s/%s", r.Repository.Owner, r.Repository.Name)
}

func (r CanonicalRelease) DownloadURL(fileName string) string {
	return fmt.Sprintf("%s/releases/download/%s/%s", r.RepositoryURL(), r.Tag, fileName)
}

func (r CanonicalRelease) releaseTagURL() string {
	return fmt.Sprintf("%s/releases/tag/%s", r.RepositoryURL(), r.Tag)
}

func (r CanonicalRelease) checksumManifestContract() CanonicalChecksumManifest {
	fileName := checksumManifestName(r.Version)
	return CanonicalChecksumManifest{
		FileName: fileName,
		URL:      r.DownloadURL(fileName),
	}
}

func (r CanonicalRelease) assetsFromSupportedTargets() []CanonicalReleaseAsset {
	assets := make([]CanonicalReleaseAsset, 0, len(canonicalReleaseTargetInventory))
	for _, target := range r.Targets() {
		assets = append(assets, r.archiveAssetForTarget(target))
	}

	return assets
}

func (r CanonicalRelease) archiveAsset(goos, goarch string) CanonicalReleaseAsset {
	return r.archiveAssetForTarget(CanonicalReleaseTarget{
		GOOS:          goos,
		GOARCH:        goarch,
		ArchiveFormat: archiveFormat(goos),
	})
}

func (r CanonicalRelease) archiveAssetForTarget(target CanonicalReleaseTarget) CanonicalReleaseAsset {
	fileName := archiveName(r.Version, target.GOOS, target.GOARCH)

	return CanonicalReleaseAsset{
		GOOS:          target.GOOS,
		GOARCH:        target.GOARCH,
		FileName:      fileName,
		DownloadURL:   r.DownloadURL(fileName),
		ArchiveFormat: target.ArchiveFormat,
	}
}

func canonicalReleaseTag(version string) string {
	return "v" + version
}

func canonicalReleaseAssetKey(goos, goarch string) string {
	return fmt.Sprintf("%s/%s", goos, goarch)
}
