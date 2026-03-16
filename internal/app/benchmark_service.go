package app

import (
	"context"
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

type BenchmarkRepeatedRunResult struct {
	RepositoryRoot string
	SuiteID        string
	SuiteVersion   string
	Attempts       []BenchmarkAttemptResult
	Summary        BenchmarkComparisonSummary
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
		summaryResult.RepositoryRoot = root.RootPath
		summaryResult.SuiteID = suite.ID
		summaryResult.SuiteVersion = suite.Version
		summaryResult.Attempts = append(summaryResult.Attempts, BenchmarkAttemptResult{
			Attempt: attempt,
			Result:  result,
		})
	}

	summaryResult.Summary = summarizeBenchmarkAttempts(summaryResult.Attempts, summaryResult.SuiteID, summaryResult.SuiteVersion)
	summaryResult.Summary.RerunCommand = benchmarkRerunCommand(request)
	return summaryResult, nil
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
	if strings.TrimSpace(request.SuiteID) != "" {
		return fmt.Sprintf("go test ./internal/app ./internal/cli ./internal/mcp -run 'TestBenchmarkVerificationWorkflow|TestBenchmarkRerunsDeterministic' && go run ./cmd/optimusctx eval --scenario %s", request.SuiteID)
	}
	if strings.TrimSpace(request.SuitePath) != "" {
		return fmt.Sprintf("go test ./internal/app ./internal/cli ./internal/mcp -run 'TestBenchmarkVerificationWorkflow|TestBenchmarkRerunsDeterministic' && benchmark suite file %s", request.SuitePath)
	}
	return "go test ./internal/app ./internal/cli ./internal/mcp -run 'TestBenchmarkVerificationWorkflow|TestBenchmarkRerunsDeterministic'"
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
