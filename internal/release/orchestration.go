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

type ReleaseOrchestrationPlan struct {
	Mode                 ReleaseOrchestrationMode `json:"mode"`
	Version              string                   `json:"version"`
	Tag                  string                   `json:"tag"`
	CanonicalRelease     CanonicalRelease         `json:"canonicalRelease"`
	SelectedChannelIDs   []string                 `json:"selectedChannelIDs"`
	CreateGitHubRelease  bool                     `json:"createGitHubRelease"`
	ReuseExistingRelease bool                     `json:"reuseExistingRelease"`
}

func PlanReleaseOrchestration(preparation ReleasePreparation, request ReleaseOrchestrationRequest) (ReleaseOrchestrationPlan, error) {
	if request.Mode != ReleaseOrchestrationModeCreate && request.Mode != ReleaseOrchestrationModeReuse {
		return ReleaseOrchestrationPlan{}, fmt.Errorf("release orchestration mode %q is invalid", request.Mode)
	}

	release, err := NewCanonicalRelease(preparation.Version)
	if err != nil {
		return ReleaseOrchestrationPlan{}, fmt.Errorf("build canonical release: %w", err)
	}
	if preparation.Tag != release.Tag {
		return ReleaseOrchestrationPlan{}, fmt.Errorf("prepared tag %q does not match canonical release tag %q", preparation.Tag, release.Tag)
	}

	plan := ReleaseOrchestrationPlan{
		Mode:               request.Mode,
		Version:            preparation.Version,
		Tag:                preparation.Tag,
		CanonicalRelease:   release,
		SelectedChannelIDs: append([]string(nil), preparation.SelectedChannelIDs()...),
	}

	switch request.Mode {
	case ReleaseOrchestrationModeCreate:
		plan.CreateGitHubRelease = true
	case ReleaseOrchestrationModeReuse:
		reuseTag, err := NormalizeReleaseTag(request.ReleaseTag)
		if err != nil {
			return ReleaseOrchestrationPlan{}, fmt.Errorf("reuse release_tag: %w", err)
		}
		if reuseTag != preparation.Tag {
			return ReleaseOrchestrationPlan{}, fmt.Errorf("reuse release_tag %q does not match prepared tag %q", reuseTag, preparation.Tag)
		}
		plan.ReuseExistingRelease = true
	default:
		return ReleaseOrchestrationPlan{}, fmt.Errorf("release orchestration mode %q is invalid", request.Mode)
	}

	return plan, nil
}
