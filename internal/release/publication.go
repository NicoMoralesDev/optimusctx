package release

import "fmt"

type ReleasePublicationMode string

const (
	ReleasePublicationModeFanout ReleasePublicationMode = "fanout"
	ReleasePublicationModeRerun  ReleasePublicationMode = "rerun"
)

type ReleasePublicationRequest struct {
	Mode    ReleasePublicationMode `json:"mode"`
	Channel string                 `json:"channel,omitempty"`
}

type ReleasePublicationChannel struct {
	ID                       string `json:"id"`
	Name                     string `json:"name"`
	PublicationTarget        string `json:"publicationTarget"`
	CanonicalReleaseURL      string `json:"canonicalReleaseURL"`
	ReleaseTag               string `json:"releaseTag"`
	ChecksumManifestURL      string `json:"checksumManifestURL"`
	CredentialEnvVar         string `json:"credentialEnvVar"`
	RenderCommand            string `json:"renderCommand"`
	RetrySafeWithExistingTag bool   `json:"retrySafeWithExistingTag"`
	NeedsExternalRepo        bool   `json:"needsExternalRepo"`
}

type ReleasePublicationPlan struct {
	Mode             ReleasePublicationMode      `json:"mode"`
	ReleaseTag       string                      `json:"releaseTag"`
	CanonicalRelease CanonicalRelease            `json:"canonicalRelease"`
	Channels         []ReleasePublicationChannel `json:"channels"`
	RequestedChannel string                      `json:"requestedChannel,omitempty"`
}

type releasePublicationChannelSpec struct {
	CredentialEnvVar  string
	RenderCommand     string
	NeedsExternalRepo bool
}

func PlanReleasePublication(orchestration ReleaseOrchestrationPlan, request ReleasePublicationRequest) (ReleasePublicationPlan, error) {
	if err := validateReleasePublicationMode(request.Mode); err != nil {
		return ReleasePublicationPlan{}, err
	}
	if err := validateReleaseTagAgreement(orchestration.Tag, orchestration.CanonicalRelease.Tag, "orchestration tag", "canonical release tag"); err != nil {
		return ReleasePublicationPlan{}, err
	}
	if err := validateReleaseTagAgreement(orchestration.GitHubRelease.ReleaseTag, orchestration.CanonicalRelease.Tag, "github release tag", "canonical release tag"); err != nil {
		return ReleasePublicationPlan{}, err
	}

	plan := ReleasePublicationPlan{
		Mode:             request.Mode,
		ReleaseTag:       orchestration.GitHubRelease.ReleaseTag,
		CanonicalRelease: orchestration.CanonicalRelease,
		Channels:         []ReleasePublicationChannel{},
	}

	switch request.Mode {
	case ReleasePublicationModeFanout:
		for _, channel := range orchestration.DownstreamSelectedChannels() {
			publicationChannel, err := releasePublicationChannelFromPlan(channel, orchestration)
			if err != nil {
				return ReleasePublicationPlan{}, err
			}
			plan.Channels = append(plan.Channels, publicationChannel)
		}
		return plan, nil
	case ReleasePublicationModeRerun:
		requestedChannel := request.Channel
		spec, ok := publicationChannelSpec(requestedChannel)
		if !ok {
			return ReleasePublicationPlan{}, unsupportedPublicationChannelError(requestedChannel)
		}
		plan.RequestedChannel = requestedChannel
		channel, ok := orchestration.SelectedChannel(requestedChannel)
		if !ok {
			return ReleasePublicationPlan{}, fmt.Errorf("release publication channel %q was not selected in orchestration", requestedChannel)
		}
		publicationChannel := newReleasePublicationChannel(channel, orchestration, spec)
		plan.Channels = append(plan.Channels, publicationChannel)
		return plan, nil
	default:
		return ReleasePublicationPlan{}, fmt.Errorf("release publication mode %q is invalid", request.Mode)
	}
}

func validateReleasePublicationMode(mode ReleasePublicationMode) error {
	if mode != ReleasePublicationModeFanout && mode != ReleasePublicationModeRerun {
		return fmt.Errorf("release publication mode %q is invalid", mode)
	}

	return nil
}

func isDownstreamPublicationChannel(channelID string) bool {
	switch channelID {
	case ReleaseChannelNPM, ReleaseChannelHomebrew, ReleaseChannelScoop:
		return true
	default:
		return false
	}
}

func publicationChannelSpec(channelID string) (releasePublicationChannelSpec, bool) {
	switch channelID {
	case ReleaseChannelNPM:
		return releasePublicationChannelSpec{
			RenderCommand:     "bash scripts/render-npm-package.sh",
			NeedsExternalRepo: false,
		}, true
	case ReleaseChannelHomebrew:
		return releasePublicationChannelSpec{
			CredentialEnvVar:  "HOMEBREW_TAP_GITHUB_TOKEN",
			RenderCommand:     "bash scripts/render-homebrew-formula.sh",
			NeedsExternalRepo: true,
		}, true
	case ReleaseChannelScoop:
		return releasePublicationChannelSpec{
			CredentialEnvVar:  "SCOOP_BUCKET_GITHUB_TOKEN",
			RenderCommand:     "bash scripts/render-scoop-manifest.sh",
			NeedsExternalRepo: true,
		}, true
	default:
		return releasePublicationChannelSpec{}, false
	}
}

func releasePublicationChannelFromPlan(channel ReleaseChannelPlan, orchestration ReleaseOrchestrationPlan) (ReleasePublicationChannel, error) {
	spec, ok := publicationChannelSpec(channel.ID)
	if !ok {
		return ReleasePublicationChannel{}, unsupportedPublicationChannelError(channel.ID)
	}

	return newReleasePublicationChannel(channel, orchestration, spec), nil
}

func newReleasePublicationChannel(channel ReleaseChannelPlan, orchestration ReleaseOrchestrationPlan, spec releasePublicationChannelSpec) ReleasePublicationChannel {
	return ReleasePublicationChannel{
		ID:                       channel.ID,
		Name:                     channel.Name,
		PublicationTarget:        channel.PublicationTarget,
		CanonicalReleaseURL:      orchestration.CanonicalRelease.ReleaseURL,
		ReleaseTag:               orchestration.GitHubRelease.ReleaseTag,
		ChecksumManifestURL:      orchestration.CanonicalRelease.ChecksumManifest.URL,
		CredentialEnvVar:         spec.CredentialEnvVar,
		RenderCommand:            spec.RenderCommand,
		RetrySafeWithExistingTag: true,
		NeedsExternalRepo:        spec.NeedsExternalRepo,
	}
}

func unsupportedPublicationChannelError(channelID string) error {
	if channelID == ReleaseChannelGitHubArchive {
		return fmt.Errorf("release publication channel %q is not a downstream publication target", channelID)
	}

	return fmt.Errorf("release publication channel %q is unsupported", channelID)
}
