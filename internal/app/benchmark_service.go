package app

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
	"github.com/niccrow/optimusctx/internal/store/sqlite"
)

type BenchmarkService struct {
	Locator       repository.Locator
	Runner        BenchmarkRunner
	OpenStore     func(context.Context, state.Layout, string) (*sqlite.Store, error)
	ResolveLayout func(string) (state.Layout, error)
	Now           func() time.Time
}

type BenchmarkRunServiceRequest struct {
	StartPath     string
	SuiteID       string
	SuitePath     string
	SuitesDir     string
	FixturesRoot  string
	WorkspaceRoot string
}

type BenchmarkRepeatedRunRequest struct {
	StartPath     string
	SuiteID       string
	SuitePath     string
	SuitesDir     string
	FixturesRoot  string
	WorkspaceRoot string
	Attempts      int
}

type BenchmarkEvidenceBundleRequest struct {
	StartPath     string
	SuiteID       string
	SuitePath     string
	SuitesDir     string
	FixturesRoot  string
	WorkspaceRoot string
	Attempts      int
}

type BenchmarkHumanReportRequest struct {
	StartPath     string
	SuiteID       string
	SuitePath     string
	SuitesDir     string
	FixturesRoot  string
	WorkspaceRoot string
	Attempts      int
}

type BenchmarkRepeatedRunResult struct {
	RepositoryRoot string
	SuiteID        string
	SuiteVersion   string
	Attempts       []BenchmarkAttemptResult
	Summary        BenchmarkComparisonSummary
	EvidenceBundle repository.BenchmarkEvidenceBundle
}

type BenchmarkAttemptResult struct {
	Attempt int
	Result  repository.BenchmarkRunResult
}

type BenchmarkComparisonSummary struct {
	SuiteID                string
	SuiteVersion           string
	AttemptCount           int
	Arms                   []BenchmarkArmComparisonSummary
	Verification           BenchmarkVerificationResult
	InvalidRunReasons      []string
	RerunCommand           string
	MethodologyFingerprint string
}

type BenchmarkArmComparisonSummary struct {
	ArmKind repository.BenchmarkArmKind
	ArmName string
	Lanes   []BenchmarkLaneComparisonSummary
}

type BenchmarkLaneComparisonSummary struct {
	Lane                   repository.BenchmarkLane
	AttemptCount           int
	SuccessCount           int
	InvalidAttemptCount    int
	ElapsedMS              BenchmarkInt64Stats
	ActionCount            BenchmarkInt64Stats
	BroadSearchActions     BenchmarkInt64Stats
	TargetedLookupActions  BenchmarkInt64Stats
	FileReadActions        BenchmarkInt64Stats
	BytesRead              BenchmarkInt64Stats
	ConsultedArtifacts     []string
	RejectedAttemptReasons []string
	elapsedValues          []int64
	actionValues           []int64
	broadSearchValues      []int64
	targetedLookupValues   []int64
	fileReadValues         []int64
	bytesReadValues        []int64
}

type BenchmarkInt64Stats struct {
	Min    int64
	Max    int64
	Median int64
	Mean   int64
}

type BenchmarkVerificationResult struct {
	Passed        bool
	FailureReason string
	DriftReasons  []string
}

type BenchmarkHumanSummary struct {
	SuiteID                string
	SuiteVersion           string
	FixtureID              string
	FixturePath            string
	AttemptCount           int
	MethodologyFingerprint string
	RerunCommand           string
	EstimatorPolicy        string
	UsageClaim             string
	BillingDisambiguator   string
	Verification           repository.BenchmarkEvidenceVerification
	LaneComparisons        []BenchmarkHumanLaneComparison
	AttributionRows        []BenchmarkHumanAttributionRow
}

type BenchmarkHumanLaneComparison struct {
	Lane                     repository.BenchmarkLane
	AttemptCount             int
	BaselineElapsedMS        BenchmarkInt64Stats
	TreatmentElapsedMS       BenchmarkInt64Stats
	ElapsedDeltaMS           int64
	BaselineEstimatedTokens  BenchmarkInt64Stats
	TreatmentEstimatedTokens BenchmarkInt64Stats
	EstimatedTokenDelta      int64
	InvalidAttemptReasons    []string
}

type BenchmarkHumanAttributionRow struct {
	Lane                  repository.BenchmarkLane
	ReportLabel           repository.BenchmarkReportArtifactLabel
	DisplayLabel          string
	AttemptCount          int
	MedianEstimatedTokens int64
	TotalEstimatedTokens  int64
	EvidenceRefs          []BenchmarkHumanEvidenceRef
}

type BenchmarkHumanEvidenceRef struct {
	Attempt         int
	StepID          string
	ArtifactPath    string
	EstimatedTokens int64
}

type benchmarkHumanLaneAccumulator struct {
	attemptCount     int
	baselineElapsed  []int64
	treatmentElapsed []int64
	baselineTokens   []int64
	treatmentTokens  []int64
	invalidReasons   []string
}

func NewBenchmarkService() BenchmarkService {
	return BenchmarkService{
		Locator: repository.NewLocator(),
		Runner:  NewBenchmarkRunner(),
		OpenStore: func(ctx context.Context, layout state.Layout, detectionMode string) (*sqlite.Store, error) {
			return sqlite.OpenOrCreateStore(ctx, layout, detectionMode)
		},
		ResolveLayout: state.ResolveLayout,
		Now:           time.Now,
	}
}

func (s BenchmarkService) Run(ctx context.Context, request BenchmarkRunServiceRequest) (repository.BenchmarkRunResult, error) {
	result, _, _, err := s.runAndPersist(ctx, request, 0)
	return result, err
}

func (s BenchmarkService) RunRepeated(ctx context.Context, request BenchmarkRepeatedRunRequest) (BenchmarkRepeatedRunResult, error) {
	if request.Attempts <= 0 {
		return BenchmarkRepeatedRunResult{}, fmt.Errorf("benchmark attempts must be positive")
	}

	summaryResult := BenchmarkRepeatedRunResult{Attempts: make([]BenchmarkAttemptResult, 0, request.Attempts)}
	var resolvedRoot repository.RepositoryRoot
	var resolvedSuite repository.BenchmarkSuiteDefinition
	for attempt := 1; attempt <= request.Attempts; attempt++ {
		result, root, suite, err := s.runAndPersist(ctx, BenchmarkRunServiceRequest{
			StartPath:     request.StartPath,
			SuiteID:       request.SuiteID,
			SuitePath:     request.SuitePath,
			SuitesDir:     request.SuitesDir,
			FixturesRoot:  request.FixturesRoot,
			WorkspaceRoot: request.WorkspaceRoot,
		}, attempt)
		if err != nil {
			return BenchmarkRepeatedRunResult{}, err
		}
		resolvedRoot = root
		resolvedSuite = suite
		summaryResult.RepositoryRoot = root.RootPath
		summaryResult.SuiteID = suite.ID
		summaryResult.SuiteVersion = suite.Version
		summaryResult.Attempts = append(summaryResult.Attempts, BenchmarkAttemptResult{
			Attempt: attempt,
			Result:  result,
		})
	}

	summaryResult.Summary = summarizeBenchmarkAttempts(summaryResult.Attempts, summaryResult.SuiteID, summaryResult.SuiteVersion)
	evidenceBundle, err := s.exportEvidenceBundleFromStore(ctx, BenchmarkEvidenceBundleRequest{
		StartPath:     request.StartPath,
		SuiteID:       request.SuiteID,
		SuitePath:     request.SuitePath,
		SuitesDir:     request.SuitesDir,
		FixturesRoot:  request.FixturesRoot,
		WorkspaceRoot: request.WorkspaceRoot,
		Attempts:      request.Attempts,
	}, resolvedRoot, resolvedSuite)
	if err != nil {
		return BenchmarkRepeatedRunResult{}, err
	}
	summaryResult.EvidenceBundle = evidenceBundle
	summaryResult.Summary.RerunCommand = evidenceBundle.RerunCommand
	summaryResult.Summary.MethodologyFingerprint = evidenceBundle.MethodologyFingerprint
	return summaryResult, nil
}

func (s BenchmarkService) ExportEvidenceBundle(ctx context.Context, request BenchmarkEvidenceBundleRequest) (repository.BenchmarkEvidenceBundle, error) {
	if request.Attempts > 0 {
		result, err := s.RunRepeated(ctx, BenchmarkRepeatedRunRequest{
			StartPath:     request.StartPath,
			SuiteID:       request.SuiteID,
			SuitePath:     request.SuitePath,
			SuitesDir:     request.SuitesDir,
			FixturesRoot:  request.FixturesRoot,
			WorkspaceRoot: request.WorkspaceRoot,
			Attempts:      request.Attempts,
		})
		if err != nil {
			return repository.BenchmarkEvidenceBundle{}, err
		}
		return result.EvidenceBundle, nil
	}

	root, suite, store, repoID, err := s.resolveBenchmarkEvidenceContext(ctx, request)
	if err != nil {
		return repository.BenchmarkEvidenceBundle{}, err
	}
	defer store.Close()
	return s.buildAndPersistEvidenceBundle(ctx, store, repoID, root.RootPath, suite, request)
}

func (s BenchmarkService) RenderHumanReport(ctx context.Context, request BenchmarkHumanReportRequest) (string, error) {
	bundle, err := s.ExportEvidenceBundle(ctx, BenchmarkEvidenceBundleRequest{
		StartPath:     request.StartPath,
		SuiteID:       request.SuiteID,
		SuitePath:     request.SuitePath,
		SuitesDir:     request.SuitesDir,
		FixturesRoot:  request.FixturesRoot,
		WorkspaceRoot: request.WorkspaceRoot,
		Attempts:      request.Attempts,
	})
	if err != nil {
		return "", err
	}
	return RenderBenchmarkComparisonReport(BuildBenchmarkHumanSummary(bundle)), nil
}

func (s BenchmarkService) VerifyMethodology(ctx context.Context, request BenchmarkRepeatedRunRequest) (BenchmarkVerificationResult, error) {
	result, err := s.RunRepeated(ctx, request)
	if err != nil {
		return BenchmarkVerificationResult{}, err
	}
	return result.Summary.Verification, nil
}

func (s BenchmarkService) runAndPersist(ctx context.Context, request BenchmarkRunServiceRequest, forcedAttempt int) (repository.BenchmarkRunResult, repository.RepositoryRoot, repository.BenchmarkSuiteDefinition, error) {
	root, err := s.Locator.Resolve(request.StartPath)
	if err != nil {
		return repository.BenchmarkRunResult{}, repository.RepositoryRoot{}, repository.BenchmarkSuiteDefinition{}, fmt.Errorf("resolve repository root: %w", err)
	}

	layoutResolver := s.ResolveLayout
	if layoutResolver == nil {
		layoutResolver = state.ResolveLayout
	}
	layout, err := layoutResolver(root.RootPath)
	if err != nil {
		return repository.BenchmarkRunResult{}, repository.RepositoryRoot{}, repository.BenchmarkSuiteDefinition{}, fmt.Errorf("resolve state layout: %w", err)
	}

	openStore := s.OpenStore
	if openStore == nil {
		openStore = func(ctx context.Context, layout state.Layout, detectionMode string) (*sqlite.Store, error) {
			return sqlite.OpenOrCreateStore(ctx, layout, detectionMode)
		}
	}
	store, err := openStore(ctx, layout, root.DetectionMode)
	if err != nil {
		return repository.BenchmarkRunResult{}, repository.RepositoryRoot{}, repository.BenchmarkSuiteDefinition{}, fmt.Errorf("open benchmark store: %w", err)
	}
	defer store.Close()

	repoRecord, err := store.UpsertRepository(ctx, root, s.nowUTC())
	if err != nil {
		return repository.BenchmarkRunResult{}, repository.RepositoryRoot{}, repository.BenchmarkSuiteDefinition{}, fmt.Errorf("persist repository metadata: %w", err)
	}

	runner := s.Runner.withDefaults()
	suite, err := runner.LoadSuite(BenchmarkSuiteRequest{
		SuiteID:      request.SuiteID,
		SuitePath:    request.SuitePath,
		SuitesDir:    request.SuitesDir,
		FixturesRoot: request.FixturesRoot,
	})
	if err != nil {
		return repository.BenchmarkRunResult{}, repository.RepositoryRoot{}, repository.BenchmarkSuiteDefinition{}, err
	}

	result, runErr := runner.Run(ctx, BenchmarkRunRequest{
		SuiteID:       request.SuiteID,
		SuitePath:     request.SuitePath,
		SuitesDir:     request.SuitesDir,
		FixturesRoot:  request.FixturesRoot,
		WorkspaceRoot: request.WorkspaceRoot,
	})
	if result.SuiteID == "" {
		return result, root, suite, runErr
	}

	attempt := forcedAttempt
	if attempt == 0 {
		attempt, err = store.NextBenchmarkAttempt(ctx, repoRecord.ID, result.SuiteID, result.SuiteVersion)
		if err != nil {
			return result, root, suite, combineBenchmarkErrors(runErr, fmt.Errorf("next benchmark attempt: %w", err))
		}
	}

	for _, arm := range sqlite.BenchmarkPersistedArmsFromResult(repoRecord.ID, attempt, result) {
		_, _, err := store.SaveBenchmarkRun(ctx, arm.Run, arm.Samples)
		if err != nil {
			return result, root, suite, combineBenchmarkErrors(runErr, fmt.Errorf("persist benchmark attempt %d: %w", attempt, err))
		}
	}

	return result, root, suite, runErr
}

func summarizeBenchmarkAttempts(attempts []BenchmarkAttemptResult, suiteID string, suiteVersion string) BenchmarkComparisonSummary {
	summary := BenchmarkComparisonSummary{
		SuiteID:      suiteID,
		SuiteVersion: suiteVersion,
		AttemptCount: len(attempts),
	}
	if len(attempts) == 0 {
		summary.Verification = BenchmarkVerificationResult{
			Passed:        false,
			FailureReason: "no benchmark attempts recorded",
			DriftReasons:  []string{"no benchmark attempts recorded"},
		}
		return summary
	}

	methodologyFingerprints := make([]string, 0, len(attempts))
	armOrder := map[repository.BenchmarkArmKind]int{}
	laneOrder := map[repository.BenchmarkLane]int{}
	arms := map[repository.BenchmarkArmKind]*BenchmarkArmComparisonSummary{}
	rejectionSet := map[string]struct{}{}
	driftReasons := make([]string, 0)
	baselineFingerprint := benchmarkAttemptFingerprint(attempts[0].Result)

	for _, attempt := range attempts {
		currentFingerprint := benchmarkAttemptFingerprint(attempt.Result)
		methodologyFingerprints = append(methodologyFingerprints, currentFingerprint)
		if currentFingerprint != baselineFingerprint {
			reason := fmt.Sprintf("attempt %d drifted from frozen methodology", attempt.Attempt)
			driftReasons = appendIfMissing(driftReasons, reason)
			rejectionSet[reason] = struct{}{}
		}
		for armIndex, arm := range attempt.Result.Arms {
			if _, ok := armOrder[arm.Kind]; !ok {
				armOrder[arm.Kind] = armIndex
			}
			armSummary, ok := arms[arm.Kind]
			if !ok {
				armSummary = &BenchmarkArmComparisonSummary{ArmKind: arm.Kind, ArmName: arm.Name}
				arms[arm.Kind] = armSummary
			}
			for laneIndex, lane := range arm.LaneResults {
				if _, ok := laneOrder[lane.Lane]; !ok {
					laneOrder[lane.Lane] = laneIndex
				}
				laneSummary := upsertLaneSummary(armSummary, lane.Lane)
				laneSummary.AttemptCount++
				if lane.Success {
					laneSummary.SuccessCount++
				}
				if !lane.Success {
					reason := fmt.Sprintf("attempt %d %s/%s did not satisfy stop condition", attempt.Attempt, arm.Kind, lane.Lane)
					laneSummary.InvalidAttemptCount++
					laneSummary.RejectedAttemptReasons = appendIfMissing(laneSummary.RejectedAttemptReasons, reason)
					rejectionSet[reason] = struct{}{}
				}
				if lane.StopMarker != lane.SuccessMarker {
					reason := fmt.Sprintf("attempt %d %s/%s changed stop marker from %q to %q", attempt.Attempt, arm.Kind, lane.Lane, lane.SuccessMarker, lane.StopMarker)
					laneSummary.InvalidAttemptCount++
					laneSummary.RejectedAttemptReasons = appendIfMissing(laneSummary.RejectedAttemptReasons, reason)
					rejectionSet[reason] = struct{}{}
				}
				accumulateLaneMetrics(laneSummary, lane)
			}
		}
	}

	fingerprintSet := uniqueSorted(methodologyFingerprints)
	if len(fingerprintSet) == 1 {
		summary.MethodologyFingerprint = fingerprintSet[0]
	}

	armKeys := make([]repository.BenchmarkArmKind, 0, len(arms))
	for kind := range arms {
		armKeys = append(armKeys, kind)
	}
	sort.SliceStable(armKeys, func(i, j int) bool { return armOrder[armKeys[i]] < armOrder[armKeys[j]] })
	for _, armKind := range armKeys {
		armSummary := arms[armKind]
		sort.SliceStable(armSummary.Lanes, func(i, j int) bool { return laneOrder[armSummary.Lanes[i].Lane] < laneOrder[armSummary.Lanes[j].Lane] })
		for index := range armSummary.Lanes {
			finalizeLaneStats(&armSummary.Lanes[index])
		}
		summary.Arms = append(summary.Arms, *armSummary)
	}

	summary.InvalidRunReasons = sortedSetKeys(rejectionSet)
	summary.Verification = BenchmarkVerificationResult{
		Passed:       len(summary.InvalidRunReasons) == 0 && len(driftReasons) == 0,
		DriftReasons: uniqueSorted(driftReasons),
	}
	if !summary.Verification.Passed {
		reasons := append([]string(nil), summary.Verification.DriftReasons...)
		reasons = append(reasons, summary.InvalidRunReasons...)
		summary.Verification.FailureReason = strings.Join(uniqueSorted(reasons), "; ")
	}
	return summary
}

func upsertLaneSummary(arm *BenchmarkArmComparisonSummary, lane repository.BenchmarkLane) *BenchmarkLaneComparisonSummary {
	for index := range arm.Lanes {
		if arm.Lanes[index].Lane == lane {
			return &arm.Lanes[index]
		}
	}
	arm.Lanes = append(arm.Lanes, BenchmarkLaneComparisonSummary{Lane: lane})
	return &arm.Lanes[len(arm.Lanes)-1]
}

func accumulateLaneMetrics(summary *BenchmarkLaneComparisonSummary, lane repository.BenchmarkLaneRunResult) {
	summary.ConsultedArtifacts = unionStrings(summary.ConsultedArtifacts, lane.EvidencePaths)
	summary.elapsedValues = append(summary.elapsedValues, lane.Elapsed.Milliseconds())
	summary.actionValues = append(summary.actionValues, lane.Effort.ActionCount)
	summary.broadSearchValues = append(summary.broadSearchValues, lane.Effort.BroadSearchActions)
	summary.targetedLookupValues = append(summary.targetedLookupValues, lane.Effort.TargetedLookupActions)
	summary.fileReadValues = append(summary.fileReadValues, lane.Effort.FileReadActions)
	summary.bytesReadValues = append(summary.bytesReadValues, lane.Effort.BytesRead)
}

func finalizeLaneStats(summary *BenchmarkLaneComparisonSummary) {
	summary.ConsultedArtifacts = uniqueSorted(summary.ConsultedArtifacts)
	summary.ElapsedMS = summarizeInt64s(summary.elapsedValues)
	summary.ActionCount = summarizeInt64s(summary.actionValues)
	summary.BroadSearchActions = summarizeInt64s(summary.broadSearchValues)
	summary.TargetedLookupActions = summarizeInt64s(summary.targetedLookupValues)
	summary.FileReadActions = summarizeInt64s(summary.fileReadValues)
	summary.BytesRead = summarizeInt64s(summary.bytesReadValues)
}

func benchmarkAttemptFingerprint(result repository.BenchmarkRunResult) string {
	var b strings.Builder
	b.WriteString(result.SuiteID)
	b.WriteString("|")
	b.WriteString(result.SuiteVersion)
	for _, arm := range result.Arms {
		b.WriteString("|arm=")
		b.WriteString(string(arm.Kind))
		b.WriteString(":")
		b.WriteString(arm.Name)
		for _, lane := range arm.LaneResults {
			b.WriteString("|lane=")
			b.WriteString(string(lane.Lane))
			b.WriteString(":")
			b.WriteString(lane.StartMarker)
			b.WriteString(":")
			b.WriteString(lane.SuccessMarker)
			b.WriteString(":")
			b.WriteString(lane.StopMarker)
			b.WriteString(":setup=")
			b.WriteString(fmt.Sprint(len(lane.Setup)))
			b.WriteString(":assert=")
			b.WriteString(fmt.Sprint(len(lane.Assertions)))
			b.WriteString(":evidence=")
			b.WriteString(strings.Join(uniqueSorted(lane.EvidencePaths), ","))
		}
	}
	return b.String()
}

func benchmarkRerunCommand(request BenchmarkRepeatedRunRequest) string {
	return benchmarkEvidenceRerunCommand(request.SuiteID, request.SuitePath, request.Attempts)
}

func summarizeInt64s(values []int64) BenchmarkInt64Stats {
	if len(values) == 0 {
		return BenchmarkInt64Stats{}
	}
	ordered := append([]int64(nil), values...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	sum := int64(0)
	for _, item := range ordered {
		sum += item
	}
	return BenchmarkInt64Stats{
		Min:    ordered[0],
		Max:    ordered[len(ordered)-1],
		Median: ordered[len(ordered)/2],
		Mean:   sum / int64(len(ordered)),
	}
}

func BuildBenchmarkHumanSummary(bundle repository.BenchmarkEvidenceBundle) BenchmarkHumanSummary {
	bundle = repository.NormalizeBenchmarkEvidenceBundle(bundle)
	summary := BenchmarkHumanSummary{
		SuiteID:                bundle.SuiteID,
		SuiteVersion:           bundle.SuiteVersion,
		FixtureID:              bundle.FixtureID,
		FixturePath:            bundle.FixturePath,
		AttemptCount:           len(bundle.Attempts),
		MethodologyFingerprint: bundle.MethodologyFingerprint,
		RerunCommand:           bundle.RerunCommand,
		EstimatorPolicy:        bundle.TokenEstimateContract.Policy.Name,
		UsageClaim:             bundle.TokenEstimateContract.UsageClaim,
		BillingDisambiguator:   bundle.TokenEstimateContract.BillingDisambiguator,
		Verification:           bundle.Verification,
	}

	type attributionAccumulator struct {
		attempts []int64
		total    int64
		refs     []BenchmarkHumanEvidenceRef
	}

	laneOrder := map[repository.BenchmarkLane]int{}
	laneAccumulators := map[repository.BenchmarkLane]*benchmarkHumanLaneAccumulator{}
	attributionOrder := []string{}
	attributionAccumulators := map[string]*attributionAccumulator{}

	for _, attempt := range bundle.Attempts {
		var perAttemptBaseline = map[repository.BenchmarkLane]repository.BenchmarkEvidenceLane{}
		var perAttemptTreatment = map[repository.BenchmarkLane]repository.BenchmarkEvidenceLane{}
		for _, arm := range attempt.Arms {
			for laneIndex, lane := range arm.Lanes {
				if _, ok := laneOrder[lane.Lane]; !ok {
					laneOrder[lane.Lane] = laneIndex
				}
				switch arm.Kind {
				case repository.BenchmarkArmKindBaseline:
					perAttemptBaseline[lane.Lane] = lane
				case repository.BenchmarkArmKindOptimusCtx:
					perAttemptTreatment[lane.Lane] = lane
				}
			}
		}

		for lane, baseline := range perAttemptBaseline {
			acc := ensureHumanLaneAccumulator(laneAccumulators, lane)
			acc.attemptCount++
			acc.baselineElapsed = append(acc.baselineElapsed, baseline.ElapsedMS)
			acc.baselineTokens = append(acc.baselineTokens, repository.EstimateBenchmarkTokensFromBytes(baseline.Effort.BytesRead))
			if !baseline.Success {
				acc.invalidReasons = appendIfMissing(acc.invalidReasons, fmt.Sprintf("attempt %d baseline/%s did not satisfy stop condition", attempt.Attempt, lane))
			}
		}
		for lane, treatment := range perAttemptTreatment {
			acc := ensureHumanLaneAccumulator(laneAccumulators, lane)
			if _, ok := perAttemptBaseline[lane]; !ok {
				acc.attemptCount++
			}
			acc.treatmentElapsed = append(acc.treatmentElapsed, treatment.ElapsedMS)
			laneTreatmentTokens := int64(0)
			for _, attribution := range treatment.Attribution {
				laneTreatmentTokens += attribution.EstimatedTokens
				key := string(lane) + "|" + string(attribution.ReportLabel)
				group, ok := attributionAccumulators[key]
				if !ok {
					group = &attributionAccumulator{}
					attributionAccumulators[key] = group
					attributionOrder = append(attributionOrder, key)
				}
				group.total += attribution.EstimatedTokens
				group.attempts = append(group.attempts, attribution.EstimatedTokens)
				group.refs = append(group.refs, BenchmarkHumanEvidenceRef{
					Attempt:         attempt.Attempt,
					StepID:          attribution.StepID,
					ArtifactPath:    attribution.ArtifactPath,
					EstimatedTokens: attribution.EstimatedTokens,
				})
			}
			acc.treatmentTokens = append(acc.treatmentTokens, laneTreatmentTokens)
			if !treatment.Success {
				acc.invalidReasons = appendIfMissing(acc.invalidReasons, fmt.Sprintf("attempt %d optimusctx/%s did not satisfy stop condition", attempt.Attempt, lane))
			}
			for _, reason := range bundle.Verification.InvalidRunReasons {
				if strings.Contains(reason, string(lane)) {
					acc.invalidReasons = appendIfMissing(acc.invalidReasons, reason)
				}
			}
		}
	}

	lanes := make([]repository.BenchmarkLane, 0, len(laneAccumulators))
	for lane := range laneAccumulators {
		lanes = append(lanes, lane)
	}
	sort.SliceStable(lanes, func(i, j int) bool {
		return benchmarkEvidenceLaneSortKey(lanes[i]) < benchmarkEvidenceLaneSortKey(lanes[j])
	})
	for _, lane := range lanes {
		acc := laneAccumulators[lane]
		baselineTokens := summarizeInt64s(acc.baselineTokens)
		treatmentTokens := summarizeInt64s(acc.treatmentTokens)
		baselineElapsed := summarizeInt64s(acc.baselineElapsed)
		treatmentElapsed := summarizeInt64s(acc.treatmentElapsed)
		summary.LaneComparisons = append(summary.LaneComparisons, BenchmarkHumanLaneComparison{
			Lane:                     lane,
			AttemptCount:             acc.attemptCount,
			BaselineElapsedMS:        baselineElapsed,
			TreatmentElapsedMS:       treatmentElapsed,
			ElapsedDeltaMS:           baselineElapsed.Median - treatmentElapsed.Median,
			BaselineEstimatedTokens:  baselineTokens,
			TreatmentEstimatedTokens: treatmentTokens,
			EstimatedTokenDelta:      baselineTokens.Median - treatmentTokens.Median,
			InvalidAttemptReasons:    uniqueSorted(acc.invalidReasons),
		})
	}

	for _, key := range attributionOrder {
		parts := strings.SplitN(key, "|", 2)
		lane := repository.BenchmarkLane(parts[0])
		label := repository.BenchmarkReportArtifactLabel(parts[1])
		acc := attributionAccumulators[key]
		sort.SliceStable(acc.refs, func(i, j int) bool {
			if acc.refs[i].Attempt == acc.refs[j].Attempt {
				if acc.refs[i].StepID == acc.refs[j].StepID {
					return acc.refs[i].ArtifactPath < acc.refs[j].ArtifactPath
				}
				return acc.refs[i].StepID < acc.refs[j].StepID
			}
			return acc.refs[i].Attempt < acc.refs[j].Attempt
		})
		summary.AttributionRows = append(summary.AttributionRows, BenchmarkHumanAttributionRow{
			Lane:                  lane,
			ReportLabel:           label,
			DisplayLabel:          benchmarkReportLabelDisplay(label),
			AttemptCount:          len(acc.attempts),
			MedianEstimatedTokens: summarizeInt64s(acc.attempts).Median,
			TotalEstimatedTokens:  acc.total,
			EvidenceRefs:          acc.refs,
		})
	}

	return summary
}

func RenderBenchmarkComparisonReport(summary BenchmarkHumanSummary) string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "benchmark report\nsuite: %s@%s\nfixture: %s (%s)\nattempts: %d\nmethodology fingerprint: %s\n",
		summary.SuiteID,
		summary.SuiteVersion,
		summary.FixtureID,
		summary.FixturePath,
		summary.AttemptCount,
		summary.MethodologyFingerprint,
	)
	if summary.Verification.Passed {
		_, _ = fmt.Fprintf(&b, "verification: passed\n")
	} else {
		_, _ = fmt.Fprintf(&b, "verification: failed\nfailure reason: %s\n", summary.Verification.FailureReason)
	}
	_, _ = fmt.Fprintf(&b, "\nlane comparison\n")
	for _, lane := range summary.LaneComparisons {
		_, _ = fmt.Fprintf(&b, "- %s: timing median baseline=%dms optimusctx=%dms delta=%dms; estimated tokens median baseline=%d optimusctx=%d delta=%d\n",
			renderBenchmarkLaneLabel(lane.Lane),
			lane.BaselineElapsedMS.Median,
			lane.TreatmentElapsedMS.Median,
			lane.ElapsedDeltaMS,
			lane.BaselineEstimatedTokens.Median,
			lane.TreatmentEstimatedTokens.Median,
			lane.EstimatedTokenDelta,
		)
		for _, reason := range lane.InvalidAttemptReasons {
			_, _ = fmt.Fprintf(&b, "  caveat: %s\n", reason)
		}
	}
	_, _ = fmt.Fprintf(&b, "\ntreatment artifact attribution\n")
	for _, row := range summary.AttributionRows {
		_, _ = fmt.Fprintf(&b, "- %s / %s: median estimated tokens=%d total estimated tokens=%d\n",
			renderBenchmarkLaneLabel(row.Lane),
			row.DisplayLabel,
			row.MedianEstimatedTokens,
			row.TotalEstimatedTokens,
		)
		for _, ref := range row.EvidenceRefs {
			_, _ = fmt.Fprintf(&b, "  evidence: attempt=%d step=%s path=%s estimated_tokens=%d\n", ref.Attempt, ref.StepID, ref.ArtifactPath, ref.EstimatedTokens)
		}
	}
	if len(summary.Verification.DriftReasons) > 0 || len(summary.Verification.InvalidRunReasons) > 0 {
		_, _ = fmt.Fprintf(&b, "\ninvalid attempts and drift\n")
		for _, reason := range uniqueSorted(append(append([]string(nil), summary.Verification.DriftReasons...), summary.Verification.InvalidRunReasons...)) {
			_, _ = fmt.Fprintf(&b, "- %s\n", reason)
		}
	}
	_, _ = fmt.Fprintf(&b, "\ncaveats\n")
	_, _ = fmt.Fprintf(&b, "- estimated tokens use %s\n", summary.EstimatorPolicy)
	_, _ = fmt.Fprintf(&b, "- %s are %s\n", summary.UsageClaim, summary.BillingDisambiguator)
	_, _ = fmt.Fprintf(&b, "- results describe only the recorded frozen-suite attempts and explicit estimator output\n")
	_, _ = fmt.Fprintf(&b, "\nrerun\n%s\n", summary.RerunCommand)
	return b.String()
}

func ensureHumanLaneAccumulator(items map[repository.BenchmarkLane]*benchmarkHumanLaneAccumulator, lane repository.BenchmarkLane) *benchmarkHumanLaneAccumulator {
	if item, ok := items[lane]; ok {
		return item
	}
	item := &benchmarkHumanLaneAccumulator{}
	items[lane] = item
	return item
}

func benchmarkReportLabelDisplay(label repository.BenchmarkReportArtifactLabel) string {
	switch label {
	case repository.BenchmarkReportArtifactLabelRepositoryMap:
		return "Repository Map"
	case repository.BenchmarkReportArtifactLabelExactLookup:
		return "Exact Lookup"
	case repository.BenchmarkReportArtifactLabelL2Context:
		return "L2 Context"
	case repository.BenchmarkReportArtifactLabelPackExport:
		return "Pack Export"
	case repository.BenchmarkReportArtifactLabelOperational:
		return "Operational"
	default:
		return "Unlabeled"
	}
}

func renderBenchmarkLaneLabel(lane repository.BenchmarkLane) string {
	switch lane {
	case repository.BenchmarkLaneDiscovery:
		return "Discovery"
	case repository.BenchmarkLaneContextAssembly:
		return "Context Assembly"
	case repository.BenchmarkLaneRefreshReady:
		return "Refresh After Change"
	case repository.BenchmarkLaneTaskCompletion:
		return "Task Completion"
	default:
		return string(lane)
	}
}

func unionStrings(existing []string, incoming []string) []string {
	out := append([]string(nil), existing...)
	for _, item := range incoming {
		if item == "" || slices.Contains(out, item) {
			continue
		}
		out = append(out, item)
	}
	return out
}

func appendIfMissing(items []string, value string) []string {
	if value == "" || slices.Contains(items, value) {
		return items
	}
	return append(items, value)
}

func uniqueSorted(items []string) []string {
	set := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item == "" {
			continue
		}
		set[item] = struct{}{}
	}
	return sortedSetKeys(set)
}

func sortedSetKeys(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func combineBenchmarkErrors(primary error, secondary error) error {
	switch {
	case primary == nil:
		return secondary
	case secondary == nil:
		return primary
	default:
		return fmt.Errorf("%v; %w", primary, secondary)
	}
}

func (s BenchmarkService) nowUTC() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func (s BenchmarkService) resolveBenchmarkEvidenceContext(ctx context.Context, request BenchmarkEvidenceBundleRequest) (repository.RepositoryRoot, repository.BenchmarkSuiteDefinition, *sqlite.Store, int64, error) {
	root, err := s.Locator.Resolve(request.StartPath)
	if err != nil {
		return repository.RepositoryRoot{}, repository.BenchmarkSuiteDefinition{}, nil, 0, fmt.Errorf("resolve repository root: %w", err)
	}

	layoutResolver := s.ResolveLayout
	if layoutResolver == nil {
		layoutResolver = state.ResolveLayout
	}
	layout, err := layoutResolver(root.RootPath)
	if err != nil {
		return repository.RepositoryRoot{}, repository.BenchmarkSuiteDefinition{}, nil, 0, fmt.Errorf("resolve state layout: %w", err)
	}

	openStore := s.OpenStore
	if openStore == nil {
		openStore = func(ctx context.Context, layout state.Layout, detectionMode string) (*sqlite.Store, error) {
			return sqlite.OpenOrCreateStore(ctx, layout, detectionMode)
		}
	}
	store, err := openStore(ctx, layout, root.DetectionMode)
	if err != nil {
		return repository.RepositoryRoot{}, repository.BenchmarkSuiteDefinition{}, nil, 0, fmt.Errorf("open benchmark store: %w", err)
	}

	repoID, err := store.LookupRepositoryID(ctx, root.RootPath)
	if err != nil {
		store.Close()
		return repository.RepositoryRoot{}, repository.BenchmarkSuiteDefinition{}, nil, 0, fmt.Errorf("lookup benchmark repository: %w", err)
	}

	runner := s.Runner.withDefaults()
	suite, err := runner.LoadSuite(BenchmarkSuiteRequest{
		SuiteID:      request.SuiteID,
		SuitePath:    request.SuitePath,
		SuitesDir:    request.SuitesDir,
		FixturesRoot: request.FixturesRoot,
	})
	if err != nil {
		store.Close()
		return repository.RepositoryRoot{}, repository.BenchmarkSuiteDefinition{}, nil, 0, err
	}
	return root, suite, store, repoID, nil
}

func (s BenchmarkService) exportEvidenceBundleFromStore(ctx context.Context, request BenchmarkEvidenceBundleRequest, root repository.RepositoryRoot, suite repository.BenchmarkSuiteDefinition) (repository.BenchmarkEvidenceBundle, error) {
	layoutResolver := s.ResolveLayout
	if layoutResolver == nil {
		layoutResolver = state.ResolveLayout
	}
	layout, err := layoutResolver(root.RootPath)
	if err != nil {
		return repository.BenchmarkEvidenceBundle{}, fmt.Errorf("resolve state layout: %w", err)
	}
	openStore := s.OpenStore
	if openStore == nil {
		openStore = func(ctx context.Context, layout state.Layout, detectionMode string) (*sqlite.Store, error) {
			return sqlite.OpenOrCreateStore(ctx, layout, detectionMode)
		}
	}
	store, err := openStore(ctx, layout, root.DetectionMode)
	if err != nil {
		return repository.BenchmarkEvidenceBundle{}, fmt.Errorf("open benchmark store: %w", err)
	}
	defer store.Close()

	repoID, err := store.LookupRepositoryID(ctx, root.RootPath)
	if err != nil {
		return repository.BenchmarkEvidenceBundle{}, fmt.Errorf("lookup benchmark repository: %w", err)
	}
	return s.buildAndPersistEvidenceBundle(ctx, store, repoID, root.RootPath, suite, request)
}

func (s BenchmarkService) buildAndPersistEvidenceBundle(ctx context.Context, store *sqlite.Store, repositoryID int64, repositoryRoot string, suite repository.BenchmarkSuiteDefinition, request BenchmarkEvidenceBundleRequest) (repository.BenchmarkEvidenceBundle, error) {
	persistedArms, err := store.ListBenchmarkRuns(ctx, repositoryID, suite.ID, suite.Version)
	if err != nil {
		return repository.BenchmarkEvidenceBundle{}, err
	}
	if len(persistedArms) == 0 {
		return repository.BenchmarkEvidenceBundle{}, fmt.Errorf("no benchmark runs recorded for suite %q", suite.ID)
	}

	attempts, tokenContract, err := benchmarkAttemptsFromPersistedArms(persistedArms)
	if err != nil {
		return repository.BenchmarkEvidenceBundle{}, err
	}
	summary := summarizeBenchmarkAttempts(attempts, suite.ID, suite.Version)
	bundle := buildBenchmarkEvidenceBundle(repositoryRoot, tokenContract, summary, attempts, benchmarkEvidenceRerunCommand(request.SuiteID, request.SuitePath, len(attempts)))
	return store.SaveBenchmarkEvidenceBundle(ctx, repositoryID, bundle)
}

func buildBenchmarkEvidenceBundle(repositoryRoot string, tokenContract repository.BenchmarkTokenEstimateContract, summary BenchmarkComparisonSummary, attempts []BenchmarkAttemptResult, rerunCommand string) repository.BenchmarkEvidenceBundle {
	bundle := repository.BenchmarkEvidenceBundle{
		SchemaVersion:          repository.BenchmarkEvidenceBundleSchemaV1,
		GeneratedAt:            time.Now().UTC(),
		RepositoryRoot:         repositoryRoot,
		SuiteID:                summary.SuiteID,
		SuiteVersion:           summary.SuiteVersion,
		TokenEstimateContract:  tokenContract,
		MethodologyFingerprint: summary.MethodologyFingerprint,
		RerunCommand:           rerunCommand,
		Verification: repository.BenchmarkEvidenceVerification{
			Passed:            summary.Verification.Passed,
			FailureReason:     summary.Verification.FailureReason,
			DriftReasons:      append([]string(nil), summary.Verification.DriftReasons...),
			InvalidRunReasons: append([]string(nil), summary.InvalidRunReasons...),
		},
		Comparison: make([]repository.BenchmarkEvidenceArmSummary, 0, len(summary.Arms)),
		Attempts:   make([]repository.BenchmarkEvidenceAttempt, 0, len(attempts)),
	}
	if len(attempts) > 0 {
		bundle.FixtureID = attempts[0].Result.FixtureID
		bundle.FixturePath = attempts[0].Result.FixturePath
	}
	for _, arm := range summary.Arms {
		armSummary := repository.BenchmarkEvidenceArmSummary{
			ArmKind: arm.ArmKind,
			ArmName: arm.ArmName,
			Lanes:   make([]repository.BenchmarkEvidenceLaneSummary, 0, len(arm.Lanes)),
		}
		for _, lane := range arm.Lanes {
			armSummary.Lanes = append(armSummary.Lanes, repository.BenchmarkEvidenceLaneSummary{
				Lane:                   lane.Lane,
				AttemptCount:           lane.AttemptCount,
				SuccessCount:           lane.SuccessCount,
				InvalidAttemptCount:    lane.InvalidAttemptCount,
				ElapsedMS:              toRepositoryEvidenceStats(lane.ElapsedMS),
				ActionCount:            toRepositoryEvidenceStats(lane.ActionCount),
				BroadSearchActions:     toRepositoryEvidenceStats(lane.BroadSearchActions),
				TargetedLookupActions:  toRepositoryEvidenceStats(lane.TargetedLookupActions),
				FileReadActions:        toRepositoryEvidenceStats(lane.FileReadActions),
				BytesRead:              toRepositoryEvidenceStats(lane.BytesRead),
				ConsultedArtifacts:     append([]string(nil), lane.ConsultedArtifacts...),
				RejectedAttemptReasons: append([]string(nil), lane.RejectedAttemptReasons...),
			})
		}
		bundle.Comparison = append(bundle.Comparison, armSummary)
	}
	for _, attempt := range attempts {
		evidenceAttempt := repository.BenchmarkEvidenceAttempt{
			Attempt: attempt.Attempt,
			Arms:    make([]repository.BenchmarkEvidenceArmAttempt, 0, len(attempt.Result.Arms)),
		}
		for _, arm := range attempt.Result.Arms {
			evidenceArm := repository.BenchmarkEvidenceArmAttempt{
				Kind:          arm.Kind,
				Name:          arm.Name,
				WorkspacePath: arm.Workspace,
				StartedAt:     arm.StartedAt,
				FinishedAt:    arm.FinishedAt,
				Lanes:         make([]repository.BenchmarkEvidenceLane, 0, len(arm.LaneResults)),
			}
			for _, lane := range arm.LaneResults {
				evidenceArm.Lanes = append(evidenceArm.Lanes, repository.BenchmarkEvidenceLane{
					Lane:           lane.Lane,
					StartMarker:    lane.StartMarker,
					SuccessMarker:  lane.SuccessMarker,
					StopMarker:     lane.StopMarker,
					SetupAppliedAt: lane.SetupAppliedAt,
					StartedAt:      lane.StartedAt,
					FinishedAt:     lane.FinishedAt,
					ElapsedMS:      lane.Elapsed.Milliseconds(),
					Success:        lane.Success,
					EvidencePaths:  append([]string(nil), lane.EvidencePaths...),
					Effort:         lane.Effort,
					Attribution:    append([]repository.BenchmarkArtifactConsumption(nil), lane.Attribution...),
				})
			}
			evidenceAttempt.Arms = append(evidenceAttempt.Arms, evidenceArm)
		}
		bundle.Attempts = append(bundle.Attempts, evidenceAttempt)
	}
	return repository.NormalizeBenchmarkEvidenceBundle(bundle)
}

func benchmarkAttemptsFromPersistedArms(persistedArms []sqlite.BenchmarkPersistedArm) ([]BenchmarkAttemptResult, repository.BenchmarkTokenEstimateContract, error) {
	type laneMetadata struct {
		SetupAppliedAt string                                    `json:"setupAppliedAt"`
		EvidencePaths  []string                                  `json:"evidencePaths"`
		Attribution    []repository.BenchmarkArtifactConsumption `json:"attribution"`
	}
	type runMetadata struct {
		WorkspacePath         string                                    `json:"workspacePath"`
		TokenEstimateContract repository.BenchmarkTokenEstimateContract `json:"tokenEstimateContract"`
	}

	grouped := make(map[int]*BenchmarkAttemptResult)
	attemptOrder := make([]int, 0)
	tokenContract := repository.DefaultBenchmarkTokenEstimateContract()
	for _, persisted := range persistedArms {
		result, ok := grouped[persisted.Run.Attempt]
		if !ok {
			grouped[persisted.Run.Attempt] = &BenchmarkAttemptResult{
				Attempt: persisted.Run.Attempt,
				Result: repository.BenchmarkRunResult{
					SchemaVersion: repository.BenchmarkSuiteSchemaV1,
					SuiteID:       persisted.Run.SuiteID,
					SuiteVersion:  persisted.Run.SuiteVersion,
					FixtureID:     persisted.Run.FixtureID,
					FixturePath:   persisted.Run.FixturePath,
					WorkspacePath: persisted.Run.WorkspacePath,
				},
			}
			result = grouped[persisted.Run.Attempt]
			attemptOrder = append(attemptOrder, persisted.Run.Attempt)
		}
		var metadata runMetadata
		if persisted.Run.MetadataJSON != "" {
			if err := json.Unmarshal([]byte(persisted.Run.MetadataJSON), &metadata); err != nil {
				return nil, repository.BenchmarkTokenEstimateContract{}, fmt.Errorf("decode benchmark run metadata for attempt %d: %w", persisted.Run.Attempt, err)
			}
			if metadata.TokenEstimateContract.Policy.Name != "" {
				tokenContract = metadata.TokenEstimateContract
			}
		}
		arm := repository.BenchmarkArmRunResult{
			Kind:        persisted.Run.ArmKind,
			Name:        persisted.Run.ArmName,
			Workspace:   firstNonEmpty(metadata.WorkspacePath, persisted.Run.WorkspacePath),
			StartedAt:   persisted.Run.StartedAt.UTC(),
			FinishedAt:  persisted.Run.CompletedAt.UTC(),
			LaneResults: make([]repository.BenchmarkLaneRunResult, 0, len(persisted.Samples)),
		}
		for _, sample := range persisted.Samples {
			var metadata laneMetadata
			if sample.Sample.MetadataJSON != "" {
				if err := json.Unmarshal([]byte(sample.Sample.MetadataJSON), &metadata); err != nil {
					return nil, repository.BenchmarkTokenEstimateContract{}, fmt.Errorf("decode benchmark lane metadata for attempt %d lane %q: %w", persisted.Run.Attempt, sample.Sample.Lane, err)
				}
			}
			lane := repository.BenchmarkLaneRunResult{
				Lane:          sample.Sample.Lane,
				StartMarker:   sample.Sample.StartMarker,
				SuccessMarker: sample.Sample.SuccessMarker,
				StopMarker:    sample.Sample.StopMarker,
				StartedAt:     sample.Sample.StartedAt.UTC(),
				FinishedAt:    sample.Sample.FinishedAt.UTC(),
				Elapsed:       time.Duration(sample.Sample.ElapsedMS) * time.Millisecond,
				Success:       sample.Sample.Success,
				EvidencePaths: append([]string(nil), metadata.EvidencePaths...),
				Attribution:   append([]repository.BenchmarkArtifactConsumption(nil), metadata.Attribution...),
			}
			if metadata.SetupAppliedAt != "" {
				parsed, err := time.Parse(time.RFC3339Nano, metadata.SetupAppliedAt)
				if err != nil {
					return nil, repository.BenchmarkTokenEstimateContract{}, fmt.Errorf("parse setupAppliedAt for attempt %d lane %q: %w", persisted.Run.Attempt, sample.Sample.Lane, err)
				}
				lane.SetupAppliedAt = parsed.UTC()
			}
			for _, metric := range sample.Metrics {
				switch metric.MetricName {
				case benchmarkEvidenceMetricActionCount():
					lane.Effort.ActionCount = metric.ValueInt
				case string(repository.BenchmarkMetricBroadSearchActions):
					lane.Effort.BroadSearchActions = metric.ValueInt
				case string(repository.BenchmarkMetricTargetedLookupActions):
					lane.Effort.TargetedLookupActions = metric.ValueInt
				case string(repository.BenchmarkMetricFileReadActions):
					lane.Effort.FileReadActions = metric.ValueInt
				case string(repository.BenchmarkMetricBytesRead):
					lane.Effort.BytesRead = metric.ValueInt
				case string(repository.BenchmarkMetricConsultedArtifacts):
					if metric.ValueText != "" {
						lane.Effort.ConsultedArtifacts = append(lane.Effort.ConsultedArtifacts, metric.ValueText)
					}
				}
			}
			arm.LaneResults = append(arm.LaneResults, lane)
		}
		sort.SliceStable(arm.LaneResults, func(i, j int) bool {
			return benchmarkEvidenceLaneSortKey(arm.LaneResults[i].Lane) < benchmarkEvidenceLaneSortKey(arm.LaneResults[j].Lane)
		})
		result.Result.Arms = append(result.Result.Arms, arm)
	}
	sort.Ints(attemptOrder)
	attempts := make([]BenchmarkAttemptResult, 0, len(attemptOrder))
	for _, attempt := range attemptOrder {
		current := grouped[attempt]
		sort.SliceStable(current.Result.Arms, func(i, j int) bool {
			return benchmarkEvidenceArmSortKey(current.Result.Arms[i].Kind) < benchmarkEvidenceArmSortKey(current.Result.Arms[j].Kind)
		})
		attempts = append(attempts, *current)
	}
	return attempts, tokenContract, nil
}

func benchmarkEvidenceRerunCommand(suiteID string, suitePath string, attempts int) string {
	base := []string{"go run ./cmd/optimusctx eval benchmark export"}
	if strings.TrimSpace(suiteID) != "" {
		base = append(base, "--suite "+suiteID)
	} else if strings.TrimSpace(suitePath) != "" {
		base = append(base, "--suite-file "+suitePath)
	}
	if attempts > 0 {
		base = append(base, fmt.Sprintf("--attempts %d", attempts))
	}
	return strings.Join(base, " ")
}

func toRepositoryEvidenceStats(stats BenchmarkInt64Stats) repository.BenchmarkEvidenceInt64Stats {
	return repository.BenchmarkEvidenceInt64Stats{
		Min:    stats.Min,
		Max:    stats.Max,
		Median: stats.Median,
		Mean:   stats.Mean,
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func benchmarkEvidenceMetricActionCount() string {
	return "action_count"
}

func benchmarkEvidenceLaneSortKey(lane repository.BenchmarkLane) int {
	switch lane {
	case repository.BenchmarkLaneDiscovery:
		return 0
	case repository.BenchmarkLaneContextAssembly:
		return 1
	case repository.BenchmarkLaneRefreshReady:
		return 2
	case repository.BenchmarkLaneTaskCompletion:
		return 3
	default:
		return 4
	}
}

func benchmarkEvidenceArmSortKey(kind repository.BenchmarkArmKind) int {
	switch kind {
	case repository.BenchmarkArmKindBaseline:
		return 0
	case repository.BenchmarkArmKindOptimusCtx:
		return 1
	default:
		return 2
	}
}
