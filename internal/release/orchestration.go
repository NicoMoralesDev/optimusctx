package release

import "fmt"

type ReleaseOrchestrationMode string

const (
	ReleaseOrchestrationModeCreate ReleaseOrchestrationMode = "create"
	ReleaseOrchestrationModeReuse  ReleaseOrchestrationMode = "reuse"
)

type ReleaseOrchestrationRequest struct {
	Mode       ReleaseOrchestrationMode `json:"mode"`
	ReleaseTag string                   `json:"release_tag,omitempty"`
}

type ReleaseAssetSource string

const (
	ReleaseAssetSourcePreparedTag ReleaseAssetSource = "prepared-tag"
	ReleaseAssetSourceExistingTag ReleaseAssetSource = "existing-tag"
)

type GitHubReleaseAction struct {
	ReleaseTag          string             `json:"releaseTag"`
	RequestedReleaseTag string             `json:"requestedReleaseTag,omitempty"`
	CanonicalReleaseURL string             `json:"canonicalReleaseURL"`
	Source              ReleaseAssetSource `json:"source"`
	Create              bool               `json:"create"`
}

type ReleaseOrchestrationPlan struct {
	Mode                 ReleaseOrchestrationMode `json:"mode"`
	Version              string                   `json:"version"`
	Tag                  string                   `json:"tag"`
	CanonicalRelease     CanonicalRelease         `json:"canonicalRelease"`
	SelectedChannelIDs   []string                 `json:"selectedChannelIDs"`
	SelectedChannels     []ReleaseChannelPlan     `json:"selectedChannels"`
	GitHubRelease        GitHubReleaseAction      `json:"gitHubRelease"`
	CreateGitHubRelease  bool                     `json:"createGitHubRelease"`
	ReuseExistingRelease bool                     `json:"reuseExistingRelease"`
}

func PlanReleaseOrchestration(preparation ReleasePreparation, request ReleaseOrchestrationRequest) (ReleaseOrchestrationPlan, error) {
	handoff, err := preparation.OrchestrationHandoff()
	if err != nil {
		return ReleaseOrchestrationPlan{}, fmt.Errorf("build release orchestration handoff: %w", err)
	}

	return planReleaseOrchestrationFromHandoff(handoff, request)
}

func planReleaseOrchestrationFromHandoff(handoff ReleaseOrchestrationHandoff, request ReleaseOrchestrationRequest) (ReleaseOrchestrationPlan, error) {
	if err := validateReleaseOrchestrationMode(request.Mode); err != nil {
		return ReleaseOrchestrationPlan{}, err
	}
	if err := handoff.Validate(); err != nil {
		return ReleaseOrchestrationPlan{}, err
	}

	plan := ReleaseOrchestrationPlan{
		Mode:               request.Mode,
		Version:            handoff.Version,
		Tag:                handoff.Tag,
		CanonicalRelease:   handoff.CanonicalRelease,
		SelectedChannelIDs: append([]string(nil), handoff.SelectedChannelIDs...),
		SelectedChannels:   cloneReleaseChannelPlans(handoff.SelectedChannels),
	}

	action, err := planGitHubReleaseAction(plan.CanonicalRelease, plan.Tag, request)
	if err != nil {
		return ReleaseOrchestrationPlan{}, err
	}
	plan.GitHubRelease = action
	plan.CreateGitHubRelease = action.Create
	plan.ReuseExistingRelease = action.Source == ReleaseAssetSourceExistingTag

	return plan, nil
}

func validateReleaseOrchestrationMode(mode ReleaseOrchestrationMode) error {
	if mode != ReleaseOrchestrationModeCreate && mode != ReleaseOrchestrationModeReuse {
		return fmt.Errorf("release orchestration mode %q is invalid", mode)
	}

	return nil
}

func (p ReleaseOrchestrationPlan) SelectedChannel(channelID string) (ReleaseChannelPlan, bool) {
	for _, channel := range p.SelectedChannels {
		if channel.ID == channelID {
			return channel, true
		}
	}

	return ReleaseChannelPlan{}, false
}

func (p ReleaseOrchestrationPlan) DownstreamSelectedChannels() []ReleaseChannelPlan {
	downstream := make([]ReleaseChannelPlan, 0, len(p.SelectedChannels))
	for _, channel := range p.SelectedChannels {
		if !isDownstreamPublicationChannel(channel.ID) {
			continue
		}
		downstream = append(downstream, channel)
	}

	return downstream
}

func selectedReleaseChannels(channels []ReleaseChannelPlan) []ReleaseChannelPlan {
	selected := make([]ReleaseChannelPlan, 0, len(channels))
	for _, channel := range channels {
		if !channel.Selected {
			continue
		}
		selected = append(selected, channel)
	}

	return selected
}

func selectedChannelIDsFromPlans(channels []ReleaseChannelPlan) []string {
	ids := make([]string, 0, len(channels))
	for _, channel := range channels {
		ids = append(ids, channel.ID)
	}

	return ids
}

func cloneReleaseChannelPlans(channels []ReleaseChannelPlan) []ReleaseChannelPlan {
	if len(channels) == 0 {
		return nil
	}

	cloned := make([]ReleaseChannelPlan, 0, len(channels))
	for _, channel := range channels {
		cloned = append(cloned, channel)
	}

	return cloned
}

func equalStringSlices(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}

	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}

	return true
}

func planGitHubReleaseAction(canonicalRelease CanonicalRelease, preparedTag string, request ReleaseOrchestrationRequest) (GitHubReleaseAction, error) {
	action := GitHubReleaseAction{
		ReleaseTag:          canonicalRelease.Tag,
		CanonicalReleaseURL: canonicalRelease.ReleaseURL,
	}

	switch request.Mode {
	case ReleaseOrchestrationModeCreate:
		if err := validateReleaseTagAgreement(preparedTag, canonicalRelease.Tag, "prepared tag", "canonical release tag"); err != nil {
			return GitHubReleaseAction{}, err
		}
		action.Source = ReleaseAssetSourcePreparedTag
		action.Create = true
		return action, nil
	case ReleaseOrchestrationModeReuse:
		reuseTag, err := resolveRequestedReuseTag(request.ReleaseTag)
		if err != nil {
			return GitHubReleaseAction{}, err
		}
		if err := validateReleaseTagAgreement(reuseTag, preparedTag, "reuse release_tag", "prepared tag"); err != nil {
			return GitHubReleaseAction{}, err
		}
		if err := validateReleaseTagAgreement(reuseTag, canonicalRelease.Tag, "reuse release_tag", "canonical release tag"); err != nil {
			return GitHubReleaseAction{}, err
		}
		action.ReleaseTag = reuseTag
		action.RequestedReleaseTag = reuseTag
		action.Source = ReleaseAssetSourceExistingTag
		action.Create = false
		return action, nil
	default:
		return GitHubReleaseAction{}, fmt.Errorf("release orchestration mode %q is invalid", request.Mode)
	}
}

func resolveRequestedReuseTag(releaseTag string) (string, error) {
	reuseTag, err := NormalizeReleaseTag(releaseTag)
	if err != nil {
		return "", fmt.Errorf("reuse release_tag: %w", err)
	}

	return reuseTag, nil
}

func validateReleaseTagAgreement(left string, right string, leftName string, rightName string) error {
	if left != right {
		return fmt.Errorf("%s %q does not match %s %q", leftName, left, rightName, right)
	}

	return nil
}
