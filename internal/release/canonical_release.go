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

func NewCanonicalRelease(version string) (CanonicalRelease, error) {
	normalizedVersion, err := NormalizeReleaseVersion(version)
	if err != nil {
		return CanonicalRelease{}, err
	}

	release := CanonicalRelease{
		Version:     normalizedVersion,
		Tag:         "v" + normalizedVersion,
		ProjectName: canonicalProjectName,
		Repository: repositoryRef{
			Owner: canonicalReleaseOwner,
			Name:  canonicalReleaseRepo,
		},
	}
	release.ReleaseURL = fmt.Sprintf("%s/releases/tag/%s", release.RepositoryURL(), release.Tag)
	release.ChecksumManifest = CanonicalChecksumManifest{
		FileName: checksumManifestName(normalizedVersion),
		URL:      release.DownloadURL(checksumManifestName(normalizedVersion)),
	}

	for _, target := range []struct {
		goos   string
		goarch string
	}{
		{goos: "darwin", goarch: "amd64"},
		{goos: "darwin", goarch: "arm64"},
		{goos: "linux", goarch: "amd64"},
		{goos: "linux", goarch: "arm64"},
		{goos: "windows", goarch: "amd64"},
		{goos: "windows", goarch: "arm64"},
	} {
		release.Assets = append(release.Assets, release.archiveAsset(target.goos, target.goarch))
	}

	return release, nil
}

func (r CanonicalRelease) Asset(goos, goarch string) (CanonicalReleaseAsset, error) {
	for _, asset := range r.Assets {
		if asset.GOOS == goos && asset.GOARCH == goarch {
			return asset, nil
		}
	}

	return CanonicalReleaseAsset{}, fmt.Errorf("canonical release asset %s/%s not found", goos, goarch)
}

func (r CanonicalRelease) RepositoryURL() string {
	return fmt.Sprintf("https://github.com/%s/%s", r.Repository.Owner, r.Repository.Name)
}

func (r CanonicalRelease) DownloadURL(fileName string) string {
	return fmt.Sprintf("%s/releases/download/%s/%s", r.RepositoryURL(), r.Tag, fileName)
}

func (r CanonicalRelease) archiveAsset(goos, goarch string) CanonicalReleaseAsset {
	fileName := archiveName(r.Version, goos, goarch)

	return CanonicalReleaseAsset{
		GOOS:          goos,
		GOARCH:        goarch,
		FileName:      fileName,
		DownloadURL:   r.DownloadURL(fileName),
		ArchiveFormat: archiveFormat(goos),
	}
}
