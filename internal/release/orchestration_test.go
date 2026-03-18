package release

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestPlanReleaseOrchestrationCreate(t *testing.T) {
	preparation := mustPrepareOrchestrationRelease(t)

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
	if got, want := selectedChannelPlanIDs(plan.SelectedChannels), []string{ReleaseChannelGitHubArchive, ReleaseChannelNPM}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SelectedChannels IDs = %v, want %v", got, want)
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
	if got, want := plan.CanonicalRelease.ReleaseURL, "https://github.com/NicoMoralesDev/optimusctx/releases/tag/v1.2.3"; got != want {
		t.Fatalf("CanonicalRelease.ReleaseURL = %q, want %q", got, want)
	}
	if got, want := plan.GitHubRelease.ReleaseTag, preparation.Tag; got != want {
		t.Fatalf("GitHubRelease.ReleaseTag = %q, want %q", got, want)
	}
	if got, want := plan.GitHubRelease.CanonicalReleaseURL, plan.CanonicalRelease.ReleaseURL; got != want {
		t.Fatalf("GitHubRelease.CanonicalReleaseURL = %q, want %q", got, want)
	}
	if got, want := plan.GitHubRelease.Source, ReleaseAssetSourcePreparedTag; got != want {
		t.Fatalf("GitHubRelease.Source = %q, want %q", got, want)
	}
	if !plan.GitHubRelease.Create {
		t.Fatalf("GitHubRelease.Create = false, want true")
	}
	if plan.GitHubRelease.RequestedReleaseTag != "" {
		t.Fatalf("GitHubRelease.RequestedReleaseTag = %q, want empty", plan.GitHubRelease.RequestedReleaseTag)
	}
}

func TestPlanReleaseOrchestrationReuse(t *testing.T) {
	preparation := mustPrepareOrchestrationRelease(t)

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
	if got, want := reusePlan.GitHubRelease.Source, ReleaseAssetSourceExistingTag; got != want {
		t.Fatalf("GitHubRelease.Source = %q, want %q", got, want)
	}
	if reusePlan.GitHubRelease.Create {
		t.Fatalf("GitHubRelease.Create = true, want false")
	}
	if got, want := reusePlan.GitHubRelease.ReleaseTag, preparation.Tag; got != want {
		t.Fatalf("GitHubRelease.ReleaseTag = %q, want %q", got, want)
	}
	if got, want := reusePlan.GitHubRelease.RequestedReleaseTag, preparation.Tag; got != want {
		t.Fatalf("GitHubRelease.RequestedReleaseTag = %q, want %q", got, want)
	}
	if got, want := reusePlan.GitHubRelease.CanonicalReleaseURL, createPlan.GitHubRelease.CanonicalReleaseURL; got != want {
		t.Fatalf("GitHubRelease.CanonicalReleaseURL = %q, want %q", got, want)
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
		Tag:     "v1.2.4",
		Channels: []ReleaseChannelPlan{
			{ID: ReleaseChannelGitHubArchive, Selected: true},
			{ID: ReleaseChannelNPM, Selected: true},
		},
	}, ReleaseOrchestrationRequest{
		Mode: ReleaseOrchestrationModeCreate,
	})
	if err == nil {
		t.Fatal("PlanReleaseOrchestration() error = nil, want tag mismatch rejection")
	}
	if !strings.Contains(err.Error(), `prepared tag "v1.2.4" does not match canonical release tag "v1.2.3"`) {
		t.Fatalf("error = %q, want mismatch message", err.Error())
	}
}

func TestPlanReleaseOrchestrationCarriesSelectedChannelPlans(t *testing.T) {
	preparation := mustPrepareOrchestrationRelease(t)

	plan, err := PlanReleaseOrchestration(preparation, ReleaseOrchestrationRequest{
		Mode: ReleaseOrchestrationModeCreate,
	})
	if err != nil {
		t.Fatalf("PlanReleaseOrchestration() error = %v", err)
	}

	if got, want := len(plan.SelectedChannels), 2; got != want {
		t.Fatalf("SelectedChannels len = %d, want %d", got, want)
	}

	for _, channel := range plan.SelectedChannels {
		if !channel.Selected {
			t.Fatalf("SelectedChannels included unselected channel: %+v", channel)
		}
		if channel.Readiness != releaseChannelReadinessReady {
			t.Fatalf("SelectedChannels readiness = %q for %s, want %q", channel.Readiness, channel.ID, releaseChannelReadinessReady)
		}
	}

	if got, want := selectedChannelPlanIDs(plan.SelectedChannels), []string{ReleaseChannelGitHubArchive, ReleaseChannelNPM}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SelectedChannels IDs = %v, want %v", got, want)
	}
	if got, want := selectedChannelPlanNames(plan.SelectedChannels), []string{"GitHub Release archives", "npm"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SelectedChannels names = %v, want %v", got, want)
	}
	if got, want := selectedChannelPlanTargets(plan.SelectedChannels), []string{"github.com/NicoMoralesDev/optimusctx releases", "@niccrow/optimusctx"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SelectedChannels publication targets = %v, want %v", got, want)
	}
}

func TestPlanReleaseOrchestrationNormalizesReuseTag(t *testing.T) {
	preparation := mustPrepareOrchestrationRelease(t)

	plan, err := PlanReleaseOrchestration(preparation, ReleaseOrchestrationRequest{
		Mode:       ReleaseOrchestrationModeReuse,
		ReleaseTag: "1.2.3",
	})
	if err != nil {
		t.Fatalf("PlanReleaseOrchestration() error = %v", err)
	}

	if got, want := plan.GitHubRelease.RequestedReleaseTag, "v1.2.3"; got != want {
		t.Fatalf("GitHubRelease.RequestedReleaseTag = %q, want %q", got, want)
	}
	if got, want := plan.GitHubRelease.ReleaseTag, "v1.2.3"; got != want {
		t.Fatalf("GitHubRelease.ReleaseTag = %q, want %q", got, want)
	}
	if got, want := plan.Tag, "v1.2.3"; got != want {
		t.Fatalf("Tag = %q, want %q", got, want)
	}
}

func mustPrepareOrchestrationRelease(t *testing.T) ReleasePreparation {
	t.Helper()

	return mustPrepareOrchestrationReleaseWithSelectedChannels(t, []string{
		ReleaseChannelGitHubArchive,
		ReleaseChannelNPM,
	})
}

func mustPrepareOrchestrationReleaseWithAllPublicationChannels(t *testing.T) ReleasePreparation {
	t.Helper()

	return mustPrepareOrchestrationReleaseWithSelectedChannels(t, []string{
		ReleaseChannelGitHubArchive,
		ReleaseChannelNPM,
		ReleaseChannelHomebrew,
		ReleaseChannelScoop,
	})
}

func mustPrepareOrchestrationReleaseWithSelectedChannels(t *testing.T, selectedChannels []string) ReleasePreparation {
	t.Helper()

	preparation, err := PrepareRelease(context.Background(), "1.2.3", "v1.2", ReleasePreparationOptions{
		Git:              fakeGitProbe{},
		Files:            releaseRepoFiles(),
		SelectedChannels: selectedChannels,
	})
	if err != nil {
		t.Fatalf("PrepareRelease() error = %v", err)
	}

	return preparation
}

func selectedChannelPlanIDs(channels []ReleaseChannelPlan) []string {
	ids := make([]string, 0, len(channels))
	for _, channel := range channels {
		ids = append(ids, channel.ID)
	}

	return ids
}

func selectedChannelPlanNames(channels []ReleaseChannelPlan) []string {
	names := make([]string, 0, len(channels))
	for _, channel := range channels {
		names = append(names, channel.Name)
	}

	return names
}

func selectedChannelPlanTargets(channels []ReleaseChannelPlan) []string {
	targets := make([]string, 0, len(channels))
	for _, channel := range channels {
		targets = append(targets, channel.PublicationTarget)
	}

	return targets
}
