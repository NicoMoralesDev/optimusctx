package release

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestPlanReleaseOrchestrationCreate(t *testing.T) {
	preparation := ReleasePreparation{
		Version: "1.2.3",
		Tag:     "v1.2.3",
		Channels: []ReleaseChannelPlan{
			{ID: ReleaseChannelGitHubArchive, Selected: true},
			{ID: ReleaseChannelHomebrew, Selected: false},
			{ID: ReleaseChannelScoop, Selected: false},
			{ID: ReleaseChannelNPM, Selected: true},
		},
	}

	plan, err := PlanReleaseOrchestration(preparation, ReleaseOrchestrationRequest{
		Mode: ReleaseOrchestrationModeCreate,
	})
	if err != nil {
		t.Fatalf("PlanReleaseOrchestration() error = %v", err)
	}

	if got, want := plan.Mode, ReleaseOrchestrationModeCreate; got != want {
		t.Fatalf("Mode = %q, want %q", got, want)
	}
	if got, want := plan.Version, preparation.Version; got != want {
		t.Fatalf("Version = %q, want %q", got, want)
	}
	if got, want := plan.Tag, preparation.Tag; got != want {
		t.Fatalf("Tag = %q, want %q", got, want)
	}
	if got, want := plan.SelectedChannelIDs, []string{ReleaseChannelGitHubArchive, ReleaseChannelNPM}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SelectedChannelIDs = %v, want %v", got, want)
	}
	if !plan.CreateGitHubRelease {
		t.Fatalf("CreateGitHubRelease = false, want true")
	}
	if plan.ReuseExistingRelease {
		t.Fatalf("ReuseExistingRelease = true, want false")
	}
	if got, want := plan.CanonicalRelease.Tag, preparation.Tag; got != want {
		t.Fatalf("CanonicalRelease.Tag = %q, want %q", got, want)
	}
	if got, want := plan.CanonicalRelease.ReleaseURL, "https://github.com/niccrow/optimusctx/releases/tag/v1.2.3"; got != want {
		t.Fatalf("CanonicalRelease.ReleaseURL = %q, want %q", got, want)
	}
}

func TestPlanReleaseOrchestrationReuse(t *testing.T) {
	preparation, err := PrepareRelease(context.Background(), "1.2.3", "v1.2", ReleasePreparationOptions{
		Git:              fakeGitProbe{},
		Files:            releaseRepoFiles(),
		SelectedChannels: []string{ReleaseChannelGitHubArchive, ReleaseChannelNPM},
	})
	if err != nil {
		t.Fatalf("PrepareRelease() error = %v", err)
	}

	createPlan, err := PlanReleaseOrchestration(preparation, ReleaseOrchestrationRequest{
		Mode: ReleaseOrchestrationModeCreate,
	})
	if err != nil {
		t.Fatalf("PlanReleaseOrchestration(create) error = %v", err)
	}

	reusePlan, err := PlanReleaseOrchestration(preparation, ReleaseOrchestrationRequest{
		Mode:       ReleaseOrchestrationModeReuse,
		ReleaseTag: "v1.2.3",
	})
	if err != nil {
		t.Fatalf("PlanReleaseOrchestration(reuse) error = %v", err)
	}

	if got, want := reusePlan.Mode, ReleaseOrchestrationModeReuse; got != want {
		t.Fatalf("Mode = %q, want %q", got, want)
	}
	if reusePlan.CreateGitHubRelease {
		t.Fatalf("CreateGitHubRelease = true, want false")
	}
	if !reusePlan.ReuseExistingRelease {
		t.Fatalf("ReuseExistingRelease = false, want true")
	}
	if got, want := reusePlan.Tag, preparation.Tag; got != want {
		t.Fatalf("Tag = %q, want %q", got, want)
	}
	if got, want := reusePlan.SelectedChannelIDs, []string{ReleaseChannelGitHubArchive, ReleaseChannelNPM}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SelectedChannelIDs = %v, want %v", got, want)
	}
	if !reflect.DeepEqual(reusePlan.CanonicalRelease, createPlan.CanonicalRelease) {
		t.Fatalf("CanonicalRelease = %+v, want %+v", reusePlan.CanonicalRelease, createPlan.CanonicalRelease)
	}
	if got := reusePlan.CanonicalRelease.Tag; got != preparation.Tag {
		t.Fatalf("CanonicalRelease.Tag = %q, want %q", got, preparation.Tag)
	}
}

func TestPlanReleaseOrchestrationRejectsInvalidMode(t *testing.T) {
	_, err := PlanReleaseOrchestration(ReleasePreparation{
		Version: "1.2.3",
		Tag:     "v1.2.3",
	}, ReleaseOrchestrationRequest{
		Mode: ReleaseOrchestrationMode("rerun"),
	})
	if err == nil {
		t.Fatal("PlanReleaseOrchestration() error = nil, want invalid mode rejection")
	}
	if got, want := err.Error(), `release orchestration mode "rerun" is invalid`; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestPlanReleaseOrchestrationRejectsTagMismatch(t *testing.T) {
	_, err := PlanReleaseOrchestration(ReleasePreparation{
		Version: "1.2.3",
		Tag:     "v1.2.3",
	}, ReleaseOrchestrationRequest{
		Mode:       ReleaseOrchestrationModeReuse,
		ReleaseTag: "v1.2.4",
	})
	if err == nil {
		t.Fatal("PlanReleaseOrchestration() error = nil, want tag mismatch rejection")
	}
	if !strings.Contains(err.Error(), `reuse release_tag "v1.2.4" does not match prepared tag "v1.2.3"`) {
		t.Fatalf("error = %q, want mismatch message", err.Error())
	}
}
