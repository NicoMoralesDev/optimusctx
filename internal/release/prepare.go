package release

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	ReleaseChannelGitHubArchive = "github-release-archive"
	ReleaseChannelHomebrew      = "homebrew"
	ReleaseChannelScoop         = "scoop"
	ReleaseChannelNPM           = "npm"

	releaseChannelReadinessPending = "pending"
	releaseChannelReadinessReady   = "ready"
	releaseChannelReadinessBlocked = "blocked"

	releaseCheckStatusReady   = "ready"
	releaseCheckStatusBlocked = "blocked"

	defaultGitRemote = "origin"

	releaseWorkflowPath     = ".github/workflows/release.yml"
	releaseChecklistPath    = "docs/release-checklist.md"
	goReleaserConfigPath    = ".goreleaser.yml"
	npmRenderScriptPath     = "scripts/render-npm-package.sh"
	homebrewTemplatePath    = "packaging/homebrew/optimusctx.rb.tmpl"
	scoopTemplatePath       = "packaging/scoop/optimusctx.json.tmpl"
	npmPackageConfigPath    = "packaging/npm/package.json"
	channelCheckGitHub      = "channel-github-release"
	channelCheckNPM         = "channel-npm"
	channelCheckHomebrew    = "channel-homebrew"
	channelCheckScoop       = "channel-scoop"
	prerequisiteCheckPrefix = "prerequisite:"
)

var (
	canonicalReleaseVersionPattern = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`)
	milestoneSeriesPattern         = regexp.MustCompile(`^v?(0|[1-9]\d*)\.(0|[1-9]\d*)$`)
	legacyReleaseTagPattern        = regexp.MustCompile(`^v(0|[1-9]\d*)\.(0|[1-9]\d*)(?:\.(0|[1-9]\d*))?$`)
)

type releaseVersion struct {
	Major int
	Minor int
	Patch int
}

type ReleaseIssue struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

type ReleaseCheck struct {
	Code    string   `json:"code"`
	Target  string   `json:"target,omitempty"`
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

type ReleaseChannelPlan struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	PublicationTarget string `json:"publicationTarget"`
	Selected          bool   `json:"selected"`
	Readiness         string `json:"readiness"`
}

type ReleasePreparation struct {
	Version  string               `json:"version"`
	Tag      string               `json:"tag"`
	Channels []ReleaseChannelPlan `json:"channels"`
	Checks   []ReleaseCheck       `json:"checks"`
	Warnings []ReleaseIssue       `json:"warnings"`
	Blockers []ReleaseIssue       `json:"blockers"`
}

type ReleaseOrchestrationHandoff struct {
	Version            string               `json:"version"`
	Tag                string               `json:"tag"`
	CanonicalRelease   CanonicalRelease     `json:"canonicalRelease"`
	SelectedChannelIDs []string             `json:"selectedChannelIDs"`
	SelectedChannels   []ReleaseChannelPlan `json:"selectedChannels"`
}

type GitProbe interface {
	WorktreeStatus(ctx context.Context) (string, error)
	LocalTags(ctx context.Context) ([]string, error)
	RemoteTags(ctx context.Context, remote string) ([]string, error)
}

type ReleasePreparationOptions struct {
	Git                GitProbe
	Files              fs.FS
	SelectedChannels   []string
	RequireRemoteCheck bool
	RemoteName         string
}

type systemGitProbe struct{}

type jsonReleasePreparation struct {
	Version  string               `json:"version"`
	Tag      string               `json:"tag"`
	Channels []ReleaseChannelPlan `json:"channels"`
	Checks   []ReleaseCheck       `json:"checks"`
	Warnings []ReleaseIssue       `json:"warnings"`
	Blockers []ReleaseIssue       `json:"blockers"`
}

func NormalizeReleaseVersion(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("release version is required")
	}
	if strings.TrimSpace(input) != input {
		return "", fmt.Errorf("release version must not contain leading or trailing whitespace")
	}

	matches := canonicalReleaseVersionPattern.FindStringSubmatch(input)
	if matches == nil {
		return "", fmt.Errorf("release version %q must use MAJOR.MINOR.PATCH", input)
	}

	version, err := parseCanonicalReleaseVersion(input)
	if err != nil {
		return "", err
	}

	return version.String(), nil
}

func NormalizeReleaseTag(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("release tag is required")
	}
	if strings.TrimSpace(input) != input {
		return "", fmt.Errorf("release tag must not contain leading or trailing whitespace")
	}

	candidate := strings.TrimPrefix(input, "v")
	version, err := NormalizeReleaseVersion(candidate)
	if err != nil {
		return "", fmt.Errorf("release tag %q is invalid: %w", input, err)
	}

	return "v" + version, nil
}

func CanonicalizeExistingTag(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("existing tag is required")
	}
	if strings.TrimSpace(input) != input {
		return "", fmt.Errorf("existing tag must not contain leading or trailing whitespace")
	}

	matches := legacyReleaseTagPattern.FindStringSubmatch(input)
	if matches == nil {
		return "", fmt.Errorf("existing tag %q must use vMAJOR.MINOR or vMAJOR.MINOR.PATCH", input)
	}

	patch := matches[3]
	if patch == "" {
		patch = "0"
	}

	version, err := NormalizeReleaseVersion(fmt.Sprintf("%s.%s.%s", matches[1], matches[2], patch))
	if err != nil {
		return "", err
	}

	return "v" + version, nil
}

func ProposeReleaseVersion(milestone string, existingTags []string) (string, error) {
	series, err := parseMilestoneSeries(milestone)
	if err != nil {
		return "", err
	}

	highestPatch := -1
	for _, tag := range existingTags {
		canonicalTag, err := CanonicalizeExistingTag(tag)
		if err != nil {
			continue
		}

		version, err := parseCanonicalReleaseVersion(strings.TrimPrefix(canonicalTag, "v"))
		if err != nil {
			continue
		}
		if version.Major != series.Major || version.Minor != series.Minor {
			continue
		}
		if version.Patch > highestPatch {
			highestPatch = version.Patch
		}
	}

	return releaseVersion{
		Major: series.Major,
		Minor: series.Minor,
		Patch: highestPatch + 1,
	}.String(), nil
}

func BuildReleasePreparation(versionInput string, milestone string, existingTags []string) (ReleasePreparation, error) {
	version := versionInput
	var err error
	if version == "" {
		version, err = ProposeReleaseVersion(milestone, existingTags)
		if err != nil {
			return ReleasePreparation{}, err
		}
	}

	version, err = NormalizeReleaseVersion(version)
	if err != nil {
		return ReleasePreparation{}, err
	}

	tag, err := NormalizeReleaseTag(version)
	if err != nil {
		return ReleasePreparation{}, err
	}

	preparation := ReleasePreparation{
		Version:  version,
		Tag:      tag,
		Channels: defaultReleaseChannels(nil),
		Checks:   []ReleaseCheck{},
		Warnings: []ReleaseIssue{},
		Blockers: []ReleaseIssue{},
	}

	if exactMatches := exactTagConflicts(tag, existingTags); len(exactMatches) > 0 {
		preparation.Blockers = append(preparation.Blockers, ReleaseIssue{
			Code:    "local-tag-conflict",
			Message: fmt.Sprintf("release tag %s already exists locally", tag),
			Details: exactMatches,
		})
	}

	if semanticAliases := semanticTagAliases(tag, existingTags); len(semanticAliases) > 0 {
		preparation.Blockers = append(preparation.Blockers, ReleaseIssue{
			Code:    "semantic-tag-conflict",
			Message: fmt.Sprintf("existing legacy tags %s conflict with requested tag %s", strings.Join(semanticAliases, ", "), tag),
			Details: semanticAliases,
		})
	}

	return preparation, nil
}

func PrepareRelease(ctx context.Context, versionInput string, milestone string, options ReleasePreparationOptions) (ReleasePreparation, error) {
	options = options.withDefaults()

	localTags, err := options.Git.LocalTags(ctx)
	if err != nil {
		return ReleasePreparation{}, fmt.Errorf("list local tags: %w", err)
	}

	preparation, err := BuildReleasePreparation(versionInput, milestone, localTags)
	if err != nil {
		return ReleasePreparation{}, err
	}
	preparation.Channels = defaultReleaseChannels(options.SelectedChannels)
	preparation.Checks = []ReleaseCheck{}

	if err := applyGitPreflight(ctx, &preparation, options, localTags); err != nil {
		return ReleasePreparation{}, err
	}
	if err := applyPrerequisiteChecks(&preparation, options); err != nil {
		return ReleasePreparation{}, err
	}

	return preparation, nil
}

func (p ReleasePreparation) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonReleasePreparation{
		Version:  p.Version,
		Tag:      p.Tag,
		Channels: nonNilChannels(p.Channels),
		Checks:   nonNilChecks(p.Checks),
		Warnings: nonNilIssues(p.Warnings),
		Blockers: nonNilIssues(p.Blockers),
	})
}

func (p ReleasePreparation) SelectedChannelIDs() []string {
	ids := make([]string, 0, len(p.Channels))
	for _, channel := range p.Channels {
		if channel.Selected {
			ids = append(ids, channel.ID)
		}
	}
	return ids
}

func (p ReleasePreparation) CanonicalRelease() (CanonicalRelease, error) {
	release, err := NewCanonicalRelease(p.Version)
	if err != nil {
		return CanonicalRelease{}, err
	}
	if p.Tag != release.Tag {
		return CanonicalRelease{}, fmt.Errorf("prepared tag %q does not match canonical release tag %q", p.Tag, release.Tag)
	}
	return release, nil
}

func (p ReleasePreparation) OrchestrationHandoff() (ReleaseOrchestrationHandoff, error) {
	release, err := p.CanonicalRelease()
	if err != nil {
		return ReleaseOrchestrationHandoff{}, err
	}

	handoff := ReleaseOrchestrationHandoff{
		Version:            p.Version,
		Tag:                p.Tag,
		CanonicalRelease:   release,
		SelectedChannelIDs: append([]string(nil), p.SelectedChannelIDs()...),
		SelectedChannels:   cloneReleaseChannelPlans(selectedReleaseChannels(p.Channels)),
	}

	if err := handoff.Validate(); err != nil {
		return ReleaseOrchestrationHandoff{}, err
	}

	return handoff, nil
}

func (h ReleaseOrchestrationHandoff) Validate() error {
	if h.Version == "" {
		return fmt.Errorf("release orchestration handoff version is required")
	}
	if h.Tag == "" {
		return fmt.Errorf("release orchestration handoff tag is required")
	}
	if h.CanonicalRelease.Tag == "" {
		return fmt.Errorf("release orchestration handoff canonical release tag is required")
	}
	if err := validateReleaseTagAgreement(h.Tag, h.CanonicalRelease.Tag, "prepared tag", "canonical release tag"); err != nil {
		return err
	}
	if ids := selectedChannelIDsFromPlans(h.SelectedChannels); !equalStringSlices(ids, h.SelectedChannelIDs) {
		return fmt.Errorf("release orchestration handoff selected channels %v do not match selected channel ids %v", ids, h.SelectedChannelIDs)
	}

	return nil
}

func defaultReleaseChannels(selected []string) []ReleaseChannelPlan {
	policy := CurrentDistributionPolicy()
	channels := make([]ReleaseChannelPlan, 0, len(policy.SupportedChannels))
	selectedSet := make(map[string]struct{}, len(selected))
	for _, id := range selected {
		selectedSet[id] = struct{}{}
	}

	selectAll := len(selectedSet) == 0
	for _, channel := range policy.SupportedChannels {
		_, explicitlySelected := selectedSet[channel.ID]
		channels = append(channels, ReleaseChannelPlan{
			ID:                channel.ID,
			Name:              channel.Name,
			PublicationTarget: channel.PublicationTarget,
			Selected:          selectAll || explicitlySelected,
			Readiness:         releaseChannelReadinessPending,
		})
	}

	return channels
}

func exactTagConflicts(targetTag string, existingTags []string) []string {
	target, err := NormalizeReleaseTag(targetTag)
	if err != nil {
		return nil
	}

	conflicts := make([]string, 0, len(existingTags))
	for _, tag := range existingTags {
		if tag == target {
			conflicts = append(conflicts, tag)
		}
	}

	sort.Strings(conflicts)
	return conflicts
}

func semanticTagAliases(targetTag string, existingTags []string) []string {
	target, err := NormalizeReleaseTag(targetTag)
	if err != nil {
		return nil
	}

	conflicts := make([]string, 0, len(existingTags))
	for _, tag := range existingTags {
		canonicalTag, err := CanonicalizeExistingTag(tag)
		if err != nil {
			continue
		}
		if canonicalTag == target && tag != target {
			conflicts = append(conflicts, tag)
		}
	}

	sort.Strings(conflicts)
	return conflicts
}

func applyGitPreflight(ctx context.Context, preparation *ReleasePreparation, options ReleasePreparationOptions, localTags []string) error {
	worktreeStatus, err := options.Git.WorktreeStatus(ctx)
	if err != nil {
		return fmt.Errorf("check worktree status: %w", err)
	}
	worktreeLines := nonEmptyLines(worktreeStatus)
	if len(worktreeLines) > 0 {
		preparation.Blockers = append(preparation.Blockers, ReleaseIssue{
			Code:    "dirty-worktree",
			Message: "release preparation requires a clean git worktree",
			Details: worktreeLines,
		})
		preparation.Checks = append(preparation.Checks, ReleaseCheck{
			Code:    "dirty-worktree",
			Target:  "git-worktree",
			Status:  releaseCheckStatusBlocked,
			Message: "git status --porcelain reported uncommitted changes",
			Details: worktreeLines,
		})
	} else {
		preparation.Checks = append(preparation.Checks, ReleaseCheck{
			Code:    "clean-worktree",
			Target:  "git-worktree",
			Status:  releaseCheckStatusReady,
			Message: "git worktree is clean",
		})
	}

	if exactMatches := exactTagConflicts(preparation.Tag, localTags); len(exactMatches) > 0 {
		preparation.Checks = append(preparation.Checks, ReleaseCheck{
			Code:    "local-tag-conflict",
			Target:  "git-local-tags",
			Status:  releaseCheckStatusBlocked,
			Message: fmt.Sprintf("local tags already contain %s", preparation.Tag),
			Details: exactMatches,
		})
	}
	if semanticAliases := semanticTagAliases(preparation.Tag, localTags); len(semanticAliases) > 0 {
		preparation.Checks = append(preparation.Checks, ReleaseCheck{
			Code:    "semantic-tag-conflict",
			Target:  "git-local-tags",
			Status:  releaseCheckStatusBlocked,
			Message: fmt.Sprintf("legacy local tags conflict with %s", preparation.Tag),
			Details: semanticAliases,
		})
	}

	if !options.RequireRemoteCheck && len(preparation.SelectedChannelIDs()) == 0 {
		return nil
	}

	remoteTags, err := options.Git.RemoteTags(ctx, options.RemoteName)
	if err != nil {
		details := []string{
			fmt.Sprintf("remote=%s", options.RemoteName),
			err.Error(),
		}
		preparation.Blockers = append(preparation.Blockers, ReleaseIssue{
			Code:    "remote-tag-check-unavailable",
			Message: fmt.Sprintf("unable to verify remote tags on %s", options.RemoteName),
			Details: details,
		})
		preparation.Checks = append(preparation.Checks, ReleaseCheck{
			Code:    "remote-tag-check-unavailable",
			Target:  "git-remote-tags",
			Status:  releaseCheckStatusBlocked,
			Message: fmt.Sprintf("git ls-remote --tags %s failed", options.RemoteName),
			Details: details,
		})
		return nil
	}

	if exactMatches := exactTagConflicts(preparation.Tag, remoteTags); len(exactMatches) > 0 {
		preparation.Blockers = append(preparation.Blockers, ReleaseIssue{
			Code:    "remote-tag-conflict",
			Message: fmt.Sprintf("release tag %s already exists on remote %s", preparation.Tag, options.RemoteName),
			Details: exactMatches,
		})
		preparation.Checks = append(preparation.Checks, ReleaseCheck{
			Code:    "remote-tag-conflict",
			Target:  "git-remote-tags",
			Status:  releaseCheckStatusBlocked,
			Message: fmt.Sprintf("remote tags already contain %s", preparation.Tag),
			Details: exactMatches,
		})
	} else {
		preparation.Checks = append(preparation.Checks, ReleaseCheck{
			Code:    "remote-tag-clear",
			Target:  "git-remote-tags",
			Status:  releaseCheckStatusReady,
			Message: fmt.Sprintf("remote %s does not contain %s", options.RemoteName, preparation.Tag),
		})
	}

	if semanticAliases := semanticTagAliases(preparation.Tag, remoteTags); len(semanticAliases) > 0 {
		preparation.Blockers = append(preparation.Blockers, ReleaseIssue{
			Code:    "semantic-tag-conflict",
			Message: fmt.Sprintf("existing remote legacy tags %s conflict with requested tag %s", strings.Join(semanticAliases, ", "), preparation.Tag),
			Details: semanticAliases,
		})
		preparation.Checks = append(preparation.Checks, ReleaseCheck{
			Code:    "semantic-tag-conflict",
			Target:  "git-remote-tags",
			Status:  releaseCheckStatusBlocked,
			Message: fmt.Sprintf("legacy remote tags conflict with %s", preparation.Tag),
			Details: semanticAliases,
		})
	}

	return nil
}

func applyPrerequisiteChecks(preparation *ReleasePreparation, options ReleasePreparationOptions) error {
	requiredFiles := []string{
		goReleaserConfigPath,
		releaseWorkflowPath,
		releaseChecklistPath,
		npmRenderScriptPath,
		homebrewTemplatePath,
		scoopTemplatePath,
		npmPackageConfigPath,
	}

	fileContents := map[string]string{}
	missingFiles := map[string]bool{}
	for _, path := range requiredFiles {
		content, err := readRequiredFile(options.Files, path)
		if err != nil {
			if errorsIs(err, fs.ErrNotExist) {
				missingFiles[path] = true
				preparation.Blockers = append(preparation.Blockers, ReleaseIssue{
					Code:    "missing-release-prerequisite",
					Message: fmt.Sprintf("required release prerequisite %s is missing", path),
					Details: []string{path},
				})
				preparation.Checks = append(preparation.Checks, ReleaseCheck{
					Code:    prerequisiteCheckPrefix + path,
					Target:  path,
					Status:  releaseCheckStatusBlocked,
					Message: fmt.Sprintf("required release prerequisite %s is missing", path),
					Details: []string{path},
				})
				continue
			}
			return fmt.Errorf("read prerequisite %s: %w", path, err)
		}

		fileContents[path] = content
		preparation.Checks = append(preparation.Checks, ReleaseCheck{
			Code:    prerequisiteCheckPrefix + path,
			Target:  path,
			Status:  releaseCheckStatusReady,
			Message: fmt.Sprintf("required release prerequisite %s is present", path),
		})
	}

	checklist := fileContents[releaseChecklistPath]
	workflow := fileContents[releaseWorkflowPath]

	setChannelReadiness(preparation, ReleaseChannelGitHubArchive, evaluateGitHubChannel(workflow, missingFiles))
	setChannelReadiness(preparation, ReleaseChannelNPM, evaluateNPMChannel(workflow, missingFiles))
	setChannelReadiness(preparation, ReleaseChannelHomebrew, evaluateHomebrewChannel(workflow, checklist, missingFiles))
	setChannelReadiness(preparation, ReleaseChannelScoop, evaluateScoopChannel(workflow, checklist, missingFiles))

	return nil
}

func evaluateGitHubChannel(workflow string, missingFiles map[string]bool) channelEvaluation {
	if missingFiles[goReleaserConfigPath] || missingFiles[releaseWorkflowPath] {
		return channelEvaluation{
			ID:        ReleaseChannelGitHubArchive,
			CheckCode: channelCheckGitHub,
			Readiness: releaseChannelReadinessBlocked,
			Message:   "GitHub Release publication is blocked because required release files are missing",
			Details:   missingDetailList(missingFiles, goReleaserConfigPath, releaseWorkflowPath),
			Blocker:   true,
		}
	}

	requiredMarkers := []string{
		"workflow_dispatch:",
		"release_tag:",
		"goreleaser/goreleaser-action@v6",
		"args: release --clean",
	}
	if missing := missingMarkers(workflow, requiredMarkers...); len(missing) > 0 {
		return channelEvaluation{
			ID:        ReleaseChannelGitHubArchive,
			CheckCode: channelCheckGitHub,
			Readiness: releaseChannelReadinessBlocked,
			Message:   "GitHub Release publication workflow is missing required release contract markers",
			Details:   missing,
			Blocker:   true,
		}
	}

	return channelEvaluation{
		ID:        ReleaseChannelGitHubArchive,
		CheckCode: channelCheckGitHub,
		Readiness: releaseChannelReadinessReady,
		Message:   "GitHub Release publication contract is wired",
	}
}

func evaluateNPMChannel(workflow string, missingFiles map[string]bool) channelEvaluation {
	if missingFiles[releaseWorkflowPath] || missingFiles[npmRenderScriptPath] || missingFiles[npmPackageConfigPath] {
		return channelEvaluation{
			ID:        ReleaseChannelNPM,
			CheckCode: channelCheckNPM,
			Readiness: releaseChannelReadinessBlocked,
			Message:   "npm publication is blocked because required npm release files are missing",
			Details:   missingDetailList(missingFiles, releaseWorkflowPath, npmRenderScriptPath, npmPackageConfigPath),
			Blocker:   true,
		}
	}

	requiredMarkers := []string{
		"name: Publish npm wrapper package",
		"needs: release",
		"bash scripts/render-npm-package.sh",
		"npm publish --access public",
		"NPM_TOKEN",
	}
	if missing := missingMarkers(workflow, requiredMarkers...); len(missing) > 0 {
		return channelEvaluation{
			ID:        ReleaseChannelNPM,
			CheckCode: channelCheckNPM,
			Readiness: releaseChannelReadinessBlocked,
			Message:   "npm publication workflow is missing required release contract markers",
			Details:   missing,
			Blocker:   true,
		}
	}

	return channelEvaluation{
		ID:        ReleaseChannelNPM,
		CheckCode: channelCheckNPM,
		Readiness: releaseChannelReadinessReady,
		Message:   "npm publication contract is wired",
	}
}

func evaluateHomebrewChannel(workflow string, checklist string, missingFiles map[string]bool) channelEvaluation {
	details := missingDetailList(missingFiles, homebrewTemplatePath)
	if !strings.Contains(checklist, homebrewTapTokenEnv) {
		details = append(details, homebrewTapTokenEnv)
	}
	if !strings.Contains(workflow, homebrewTapTokenEnv) {
		details = append(details, "release workflow does not yet wire "+homebrewTapTokenEnv)
	}

	return channelEvaluation{
		ID:        ReleaseChannelHomebrew,
		CheckCode: channelCheckHomebrew,
		Readiness: releaseChannelReadinessBlocked,
		Message:   "Homebrew publication remains blocked until release automation wires the tap publication path",
		Details:   details,
	}
}

func evaluateScoopChannel(workflow string, checklist string, missingFiles map[string]bool) channelEvaluation {
	details := missingDetailList(missingFiles, scoopTemplatePath)
	if !strings.Contains(checklist, scoopBucketTokenEnv) {
		details = append(details, scoopBucketTokenEnv)
	}
	if !strings.Contains(workflow, scoopBucketTokenEnv) {
		details = append(details, "release workflow does not yet wire "+scoopBucketTokenEnv)
	}

	return channelEvaluation{
		ID:        ReleaseChannelScoop,
		CheckCode: channelCheckScoop,
		Readiness: releaseChannelReadinessBlocked,
		Message:   "Scoop publication remains blocked until release automation wires the bucket publication path",
		Details:   details,
	}
}

type channelEvaluation struct {
	ID        string
	CheckCode string
	Readiness string
	Message   string
	Details   []string
	Blocker   bool
}

func setChannelReadiness(preparation *ReleasePreparation, channelID string, evaluation channelEvaluation) {
	selected := false
	for index := range preparation.Channels {
		if preparation.Channels[index].ID != channelID {
			continue
		}
		selected = preparation.Channels[index].Selected
		preparation.Channels[index].Readiness = evaluation.Readiness
		break
	}

	status := releaseCheckStatusReady
	if evaluation.Readiness == releaseChannelReadinessBlocked {
		status = releaseCheckStatusBlocked
	}
	preparation.Checks = append(preparation.Checks, ReleaseCheck{
		Code:    evaluation.CheckCode,
		Target:  channelID,
		Status:  status,
		Message: evaluation.Message,
		Details: evaluation.Details,
	})

	if evaluation.Readiness == releaseChannelReadinessBlocked && selected {
		preparation.Blockers = append(preparation.Blockers, ReleaseIssue{
			Code:    evaluation.CheckCode,
			Message: evaluation.Message,
			Details: evaluation.Details,
		})
	}
}

func readRequiredFile(files fs.FS, path string) (string, error) {
	content, err := fs.ReadFile(files, path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func missingMarkers(content string, markers ...string) []string {
	var missing []string
	for _, marker := range markers {
		if !strings.Contains(content, marker) {
			missing = append(missing, marker)
		}
	}
	return missing
}

func missingDetailList(missingFiles map[string]bool, paths ...string) []string {
	details := make([]string, 0, len(paths))
	for _, path := range paths {
		if missingFiles[path] {
			details = append(details, path)
		}
	}
	return details
}

func nonEmptyLines(content string) []string {
	lines := strings.Split(content, "\n")
	values := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		values = append(values, line)
	}
	return values
}

func nonNilChannels(channels []ReleaseChannelPlan) []ReleaseChannelPlan {
	if channels == nil {
		return []ReleaseChannelPlan{}
	}
	return channels
}

func nonNilChecks(checks []ReleaseCheck) []ReleaseCheck {
	if checks == nil {
		return []ReleaseCheck{}
	}
	return checks
}

func nonNilIssues(issues []ReleaseIssue) []ReleaseIssue {
	if issues == nil {
		return []ReleaseIssue{}
	}
	return issues
}

func parseMilestoneSeries(input string) (releaseVersion, error) {
	if input == "" {
		return releaseVersion{}, fmt.Errorf("milestone series is required")
	}
	if strings.TrimSpace(input) != input {
		return releaseVersion{}, fmt.Errorf("milestone series must not contain leading or trailing whitespace")
	}

	matches := milestoneSeriesPattern.FindStringSubmatch(input)
	if matches == nil {
		return releaseVersion{}, fmt.Errorf("milestone series %q must use vMAJOR.MINOR", input)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return releaseVersion{}, fmt.Errorf("parse milestone major: %w", err)
	}
	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return releaseVersion{}, fmt.Errorf("parse milestone minor: %w", err)
	}

	return releaseVersion{Major: major, Minor: minor}, nil
}

func parseCanonicalReleaseVersion(input string) (releaseVersion, error) {
	matches := canonicalReleaseVersionPattern.FindStringSubmatch(input)
	if matches == nil {
		return releaseVersion{}, fmt.Errorf("release version %q must use MAJOR.MINOR.PATCH", input)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return releaseVersion{}, fmt.Errorf("parse release major: %w", err)
	}
	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return releaseVersion{}, fmt.Errorf("parse release minor: %w", err)
	}
	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return releaseVersion{}, fmt.Errorf("parse release patch: %w", err)
	}

	return releaseVersion{Major: major, Minor: minor, Patch: patch}, nil
}

func (v releaseVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (o ReleasePreparationOptions) withDefaults() ReleasePreparationOptions {
	if o.Git == nil {
		o.Git = systemGitProbe{}
	}
	if o.Files == nil {
		o.Files = os.DirFS(".")
	}
	if o.RemoteName == "" {
		o.RemoteName = defaultGitRemote
	}
	if !o.RequireRemoteCheck {
		o.RequireRemoteCheck = true
	}
	return o
}

func (systemGitProbe) WorktreeStatus(ctx context.Context) (string, error) {
	return runGitCommand(ctx, "status", "--porcelain")
}

func (systemGitProbe) LocalTags(ctx context.Context) ([]string, error) {
	output, err := runGitCommand(ctx, "tag", "--list")
	if err != nil {
		return nil, err
	}
	return nonEmptyLines(output), nil
}

func (systemGitProbe) RemoteTags(ctx context.Context, remote string) ([]string, error) {
	output, err := runGitCommand(ctx, "ls-remote", "--tags", remote)
	if err != nil {
		return nil, err
	}
	return parseRemoteTags(output), nil
}

func runGitCommand(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v: %w: %s", args, err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

func parseRemoteTags(output string) []string {
	seen := map[string]struct{}{}
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		ref := fields[1]
		ref = strings.TrimPrefix(ref, "refs/tags/")
		ref = strings.TrimSuffix(ref, "^{}")
		if ref == "" {
			continue
		}
		seen[ref] = struct{}{}
	}

	tags := make([]string, 0, len(seen))
	for tag := range seen {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

func errorsIs(err error, target error) bool {
	return err != nil && target != nil && (err == target || strings.Contains(err.Error(), target.Error()))
}
