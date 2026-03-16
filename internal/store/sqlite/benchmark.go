package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
)

const benchmarkMetricActionCount = "action_count"

type BenchmarkRunRecord struct {
	ID            int64
	RepositoryID  int64
	SuiteID       string
	SuiteVersion  string
	FixtureID     string
	FixturePath   string
	ArmKind       repository.BenchmarkArmKind
	ArmName       string
	Attempt       int
	WorkspacePath string
	StartedAt     time.Time
	CompletedAt   time.Time
	MetadataJSON  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type BenchmarkLaneSampleRecord struct {
	ID             int64
	BenchmarkRunID int64
	Lane           repository.BenchmarkLane
	StartMarker    string
	SuccessMarker  string
	StopMarker     string
	StartedAt      time.Time
	FinishedAt     time.Time
	ElapsedMS      int64
	Success        bool
	MetadataJSON   string
}

type BenchmarkLaneMetricRecord struct {
	ID                    int64
	BenchmarkLaneSampleID int64
	MetricName            string
	Ordinal               int
	ValueInt              int64
	ValueText             string
	MetadataJSON          string
}

type BenchmarkLaneSampleBundle struct {
	Sample  BenchmarkLaneSampleRecord
	Metrics []BenchmarkLaneMetricRecord
}

type BenchmarkPersistedArm struct {
	Run     BenchmarkRunRecord
	Samples []BenchmarkLaneSampleBundle
}

type BenchmarkEvidenceBundleRecord struct {
	ID                     int64
	RepositoryID           int64
	SchemaVersion          string
	SuiteID                string
	SuiteVersion           string
	SuiteSchemaVersion     string
	CountedInputPolicy     string
	SystemOutputPolicy     string
	FinalArtifactPolicy    string
	CountedInputCount      int
	MethodologyFingerprint string
	RerunCommand           string
	GeneratedAt            time.Time
	MetadataJSON           string
	BundleJSON             string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type BenchmarkEvidenceLaneSummaryRecord struct {
	ID                         int64
	BenchmarkEvidenceBundleID  int64
	ArmKind                    repository.BenchmarkArmKind
	ArmName                    string
	Lane                       repository.BenchmarkLane
	AttemptCount               int
	SuccessCount               int
	InvalidAttemptCount        int
	ElapsedMSJSON              string
	ActionCountJSON            string
	BroadSearchActionsJSON     string
	TargetedLookupActionsJSON  string
	FileReadActionsJSON        string
	BytesReadJSON              string
	FinalArtifactContractID    string
	FinalArtifactPath          string
	FinalArtifactPassed        bool
	FinalArtifactFailureReason string
	MetadataJSON               string
}

type BenchmarkEvidenceAttributionRecord struct {
	ID                        int64
	BenchmarkEvidenceBundleID int64
	Attempt                   int
	ArmKind                   repository.BenchmarkArmKind
	Lane                      repository.BenchmarkLane
	Ordinal                   int
	StepID                    string
	StepName                  string
	Boundary                  repository.BenchmarkEvidenceBoundary
	CountsTowardTokens        bool
	Surface                   repository.BenchmarkTreatmentSurface
	CommandName               repository.EvalCommandName
	ToolName                  string
	ArtifactType              repository.BenchmarkArtifactType
	ReportLabel               repository.BenchmarkReportArtifactLabel
	SourceKind                repository.BenchmarkTokenEstimateSourceKind
	ArtifactPath              string
	EstimatedBytes            int64
	EstimatedTokens           int64
	MetadataJSON              string
}

type sqlQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func BenchmarkPersistedArmsFromResult(repositoryID int64, attempt int, result repository.BenchmarkRunResult) []BenchmarkPersistedArm {
	persisted := make([]BenchmarkPersistedArm, 0, len(result.Arms))
	for _, arm := range result.Arms {
		run := BenchmarkRunRecord{
			RepositoryID:  repositoryID,
			SuiteID:       result.SuiteID,
			SuiteVersion:  result.SuiteVersion,
			FixtureID:     result.FixtureID,
			FixturePath:   result.FixturePath,
			ArmKind:       arm.Kind,
			ArmName:       arm.Name,
			Attempt:       attempt,
			WorkspacePath: result.WorkspacePath,
			StartedAt:     arm.StartedAt.UTC(),
			CompletedAt:   arm.FinishedAt.UTC(),
			MetadataJSON: mustMarshalBenchmarkMetadata(map[string]any{
				"workspacePath":         arm.Workspace,
				"tokenEstimateContract": repository.DefaultBenchmarkTokenEstimateContract(),
			}),
		}
		samples := make([]BenchmarkLaneSampleBundle, 0, len(arm.LaneResults))
		for _, lane := range arm.LaneResults {
			sample := BenchmarkLaneSampleRecord{
				Lane:          lane.Lane,
				StartMarker:   lane.StartMarker,
				SuccessMarker: lane.SuccessMarker,
				StopMarker:    lane.StopMarker,
				StartedAt:     lane.StartedAt.UTC(),
				FinishedAt:    lane.FinishedAt.UTC(),
				ElapsedMS:     lane.Elapsed.Milliseconds(),
				Success:       lane.Success,
				MetadataJSON: mustMarshalBenchmarkMetadata(map[string]any{
					"setupAppliedAt": lane.SetupAppliedAt.UTC().Format(time.RFC3339Nano),
					"setup":          lane.Setup,
					"assertions":     lane.Assertions,
					"evidencePaths":  lane.EvidencePaths,
					"finalArtifact":  lane.FinalArtifact,
					"attribution":    lane.Attribution,
				}),
			}
			metrics := []BenchmarkLaneMetricRecord{
				{MetricName: benchmarkMetricActionCount, ValueInt: lane.Effort.ActionCount},
				{MetricName: string(repository.BenchmarkMetricBroadSearchActions), ValueInt: lane.Effort.BroadSearchActions},
				{MetricName: string(repository.BenchmarkMetricTargetedLookupActions), ValueInt: lane.Effort.TargetedLookupActions},
				{MetricName: string(repository.BenchmarkMetricFileReadActions), ValueInt: lane.Effort.FileReadActions},
				{MetricName: string(repository.BenchmarkMetricBytesRead), ValueInt: lane.Effort.BytesRead},
			}
			for idx, artifact := range lane.Effort.ConsultedArtifacts {
				metrics = append(metrics, BenchmarkLaneMetricRecord{
					MetricName: string(repository.BenchmarkMetricConsultedArtifacts),
					Ordinal:    idx,
					ValueInt:   1,
					ValueText:  artifact,
				})
			}
			samples = append(samples, BenchmarkLaneSampleBundle{
				Sample:  sample,
				Metrics: metrics,
			})
		}
		persisted = append(persisted, BenchmarkPersistedArm{Run: run, Samples: samples})
	}
	return persisted
}

func mustMarshalBenchmarkMetadata(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}

func (s *Store) SaveBenchmarkEvidenceBundle(ctx context.Context, repositoryID int64, bundle repository.BenchmarkEvidenceBundle) (repository.BenchmarkEvidenceBundle, error) {
	if s == nil || s.db == nil {
		return repository.BenchmarkEvidenceBundle{}, fmt.Errorf("save benchmark evidence bundle: store is not initialized")
	}
	if repositoryID == 0 {
		return repository.BenchmarkEvidenceBundle{}, fmt.Errorf("save benchmark evidence bundle: repository ID is required")
	}
	normalized := repository.NormalizeBenchmarkEvidenceBundle(bundle)
	if err := validateBenchmarkEvidenceBundle(normalized); err != nil {
		return repository.BenchmarkEvidenceBundle{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return repository.BenchmarkEvidenceBundle{}, fmt.Errorf("begin benchmark evidence transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	record, err := saveBenchmarkEvidenceBundleRecord(ctx, tx, repositoryID, normalized)
	if err != nil {
		return repository.BenchmarkEvidenceBundle{}, err
	}
	if err = replaceBenchmarkEvidenceLaneSummaries(ctx, tx, record.ID, normalized.Comparison, normalized.Attempts); err != nil {
		return repository.BenchmarkEvidenceBundle{}, err
	}
	if err = replaceBenchmarkEvidenceAttributions(ctx, tx, record.ID, normalized.Attempts); err != nil {
		return repository.BenchmarkEvidenceBundle{}, err
	}
	if err = tx.Commit(); err != nil {
		return repository.BenchmarkEvidenceBundle{}, fmt.Errorf("commit benchmark evidence transaction: %w", err)
	}
	return normalized, nil
}

func (s *Store) LoadLatestBenchmarkEvidenceBundle(ctx context.Context, repositoryID int64, suiteID string, suiteVersion string) (repository.BenchmarkEvidenceBundle, error) {
	if s == nil || s.db == nil {
		return repository.BenchmarkEvidenceBundle{}, fmt.Errorf("load benchmark evidence bundle: store is not initialized")
	}
	if repositoryID == 0 {
		return repository.BenchmarkEvidenceBundle{}, fmt.Errorf("load benchmark evidence bundle: repository ID is required")
	}
	if suiteID == "" {
		return repository.BenchmarkEvidenceBundle{}, fmt.Errorf("load benchmark evidence bundle: suite ID is required")
	}
	if suiteVersion == "" {
		return repository.BenchmarkEvidenceBundle{}, fmt.Errorf("load benchmark evidence bundle: suite version is required")
	}

	var payload string
	err := s.db.QueryRowContext(ctx, `
		SELECT bundle_json
		FROM benchmark_evidence_bundles
		WHERE repository_id = ? AND suite_id = ? AND suite_version = ?
		ORDER BY generated_at DESC, id DESC
		LIMIT 1
	`, repositoryID, suiteID, suiteVersion).Scan(&payload)
	if err != nil {
		return repository.BenchmarkEvidenceBundle{}, fmt.Errorf("load benchmark evidence bundle for suite %q: %w", suiteID, err)
	}
	var bundle repository.BenchmarkEvidenceBundle
	if err := json.Unmarshal([]byte(payload), &bundle); err != nil {
		return repository.BenchmarkEvidenceBundle{}, fmt.Errorf("decode benchmark evidence bundle for suite %q: %w", suiteID, err)
	}
	return repository.NormalizeBenchmarkEvidenceBundle(bundle), nil
}

func (s *Store) SaveBenchmarkRun(ctx context.Context, run BenchmarkRunRecord, samples []BenchmarkLaneSampleBundle) (BenchmarkRunRecord, []BenchmarkLaneSampleBundle, error) {
	if s == nil || s.db == nil {
		return BenchmarkRunRecord{}, nil, fmt.Errorf("save benchmark run: store is not initialized")
	}
	if err := validateBenchmarkRun(run); err != nil {
		return BenchmarkRunRecord{}, nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return BenchmarkRunRecord{}, nil, fmt.Errorf("begin benchmark run transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	run, err = saveBenchmarkRunRecord(ctx, tx, run)
	if err != nil {
		return BenchmarkRunRecord{}, nil, err
	}
	samples, err = replaceBenchmarkLaneSamples(ctx, tx, run.ID, samples)
	if err != nil {
		return BenchmarkRunRecord{}, nil, err
	}
	if err = tx.Commit(); err != nil {
		return BenchmarkRunRecord{}, nil, fmt.Errorf("commit benchmark run transaction: %w", err)
	}
	return run, samples, nil
}

func (s *Store) LoadBenchmarkRun(ctx context.Context, runID int64) (BenchmarkRunRecord, []BenchmarkLaneSampleBundle, error) {
	if s == nil || s.db == nil {
		return BenchmarkRunRecord{}, nil, fmt.Errorf("load benchmark run: store is not initialized")
	}
	run, err := loadBenchmarkRunRecord(ctx, s.db, runID)
	if err != nil {
		return BenchmarkRunRecord{}, nil, err
	}
	samples, err := listBenchmarkLaneSamples(ctx, s.db, runID)
	if err != nil {
		return BenchmarkRunRecord{}, nil, err
	}
	return run, samples, nil
}

func (s *Store) NextBenchmarkAttempt(ctx context.Context, repositoryID int64, suiteID string, suiteVersion string) (int, error) {
	if s == nil || s.db == nil {
		return 0, fmt.Errorf("next benchmark attempt: store is not initialized")
	}
	var maxAttempt sql.NullInt64
	if err := s.db.QueryRowContext(ctx, `
		SELECT MAX(attempt)
		FROM benchmark_runs
		WHERE repository_id = ? AND suite_id = ? AND suite_version = ?
	`, repositoryID, suiteID, suiteVersion).Scan(&maxAttempt); err != nil {
		return 0, fmt.Errorf("next benchmark attempt for suite %q: %w", suiteID, err)
	}
	if !maxAttempt.Valid {
		return 1, nil
	}
	return int(maxAttempt.Int64) + 1, nil
}

func (s *Store) ListBenchmarkRuns(ctx context.Context, repositoryID int64, suiteID string, suiteVersion string) ([]BenchmarkPersistedArm, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("list benchmark runs: store is not initialized")
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id
		FROM benchmark_runs
		WHERE repository_id = ? AND suite_id = ? AND suite_version = ?
		ORDER BY attempt, CASE arm_kind WHEN 'baseline' THEN 0 ELSE 1 END, id
	`, repositoryID, suiteID, suiteVersion)
	if err != nil {
		return nil, fmt.Errorf("list benchmark runs for suite %q: %w", suiteID, err)
	}
	defer rows.Close()

	var runIDs []int64
	for rows.Next() {
		var runID int64
		if err := rows.Scan(&runID); err != nil {
			return nil, fmt.Errorf("scan benchmark run id for suite %q: %w", suiteID, err)
		}
		runIDs = append(runIDs, runID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate benchmark runs for suite %q: %w", suiteID, err)
	}

	runs := make([]BenchmarkPersistedArm, 0, len(runIDs))
	for _, runID := range runIDs {
		run, samples, err := s.LoadBenchmarkRun(ctx, runID)
		if err != nil {
			return nil, err
		}
		runs = append(runs, BenchmarkPersistedArm{Run: run, Samples: samples})
	}
	return runs, nil
}

func (s *Store) ListLatestBenchmarkRuns(ctx context.Context, repositoryID int64, suiteID string, suiteVersion string, runCount int) ([]BenchmarkPersistedArm, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("list latest benchmark runs: store is not initialized")
	}
	if runCount <= 0 {
		return nil, fmt.Errorf("list latest benchmark runs: run count must be positive")
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id
		FROM benchmark_runs
		WHERE repository_id = ? AND suite_id = ? AND suite_version = ?
		ORDER BY id DESC
		LIMIT ?
	`, repositoryID, suiteID, suiteVersion, runCount)
	if err != nil {
		return nil, fmt.Errorf("list latest benchmark runs for suite %q: %w", suiteID, err)
	}
	defer rows.Close()

	runIDs := make([]int64, 0, runCount)
	for rows.Next() {
		var runID int64
		if err := rows.Scan(&runID); err != nil {
			return nil, fmt.Errorf("scan latest benchmark run for suite %q: %w", suiteID, err)
		}
		runIDs = append(runIDs, runID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate latest benchmark runs for suite %q: %w", suiteID, err)
	}
	if len(runIDs) == 0 {
		return nil, nil
	}

	runIDSet := make(map[int64]struct{}, len(runIDs))
	for _, runID := range runIDs {
		runIDSet[runID] = struct{}{}
	}

	persisted, err := s.ListBenchmarkRuns(ctx, repositoryID, suiteID, suiteVersion)
	if err != nil {
		return nil, err
	}

	filtered := make([]BenchmarkPersistedArm, 0, len(persisted))
	for _, run := range persisted {
		if _, ok := runIDSet[run.Run.ID]; ok {
			filtered = append(filtered, run)
		}
	}
	return filtered, nil
}

func validateBenchmarkRun(run BenchmarkRunRecord) error {
	switch {
	case run.RepositoryID == 0:
		return fmt.Errorf("save benchmark run: repository ID is required")
	case run.SuiteID == "":
		return fmt.Errorf("save benchmark run: suite ID is required")
	case run.SuiteVersion == "":
		return fmt.Errorf("save benchmark run: suite version is required")
	case run.FixtureID == "":
		return fmt.Errorf("save benchmark run: fixture ID is required")
	case run.FixturePath == "":
		return fmt.Errorf("save benchmark run: fixture path is required")
	case run.ArmKind == "":
		return fmt.Errorf("save benchmark run: arm kind is required")
	case run.ArmName == "":
		return fmt.Errorf("save benchmark run: arm name is required")
	case run.Attempt <= 0:
		return fmt.Errorf("save benchmark run: attempt must be positive")
	case run.WorkspacePath == "":
		return fmt.Errorf("save benchmark run: workspace path is required")
	case run.StartedAt.IsZero():
		return fmt.Errorf("save benchmark run: started at is required")
	case run.CompletedAt.IsZero():
		return fmt.Errorf("save benchmark run: completed at is required")
	}
	return nil
}

func validateBenchmarkEvidenceBundle(bundle repository.BenchmarkEvidenceBundle) error {
	switch {
	case bundle.SchemaVersion != repository.BenchmarkEvidenceBundleSchemaV2:
		return fmt.Errorf("save benchmark evidence bundle: schema version must be %q", repository.BenchmarkEvidenceBundleSchemaV2)
	case bundle.RepositoryRoot == "":
		return fmt.Errorf("save benchmark evidence bundle: repository root is required")
	case bundle.SuiteID == "":
		return fmt.Errorf("save benchmark evidence bundle: suite ID is required")
	case bundle.SuiteVersion == "":
		return fmt.Errorf("save benchmark evidence bundle: suite version is required")
	case bundle.FixtureID == "":
		return fmt.Errorf("save benchmark evidence bundle: fixture ID is required")
	case bundle.FixturePath == "":
		return fmt.Errorf("save benchmark evidence bundle: fixture path is required")
	case bundle.MethodologyFingerprint == "":
		return fmt.Errorf("save benchmark evidence bundle: methodology fingerprint is required")
	case bundle.RerunCommand == "":
		return fmt.Errorf("save benchmark evidence bundle: rerun command is required")
	case bundle.GeneratedAt.IsZero():
		return fmt.Errorf("save benchmark evidence bundle: generated time is required")
	}
	if bundle.Methodology.SuiteSchemaVersion != repository.BenchmarkSuiteSchemaV2 {
		return fmt.Errorf("save benchmark evidence bundle: methodology suite schema version must be %q", repository.BenchmarkSuiteSchemaV2)
	}
	if err := bundle.Methodology.Boundary.Validate(); err != nil {
		return fmt.Errorf("save benchmark evidence bundle: methodology boundary: %w", err)
	}
	if len(bundle.Methodology.CountedInputs) == 0 {
		return fmt.Errorf("save benchmark evidence bundle: methodology counted inputs are required")
	}
	for idx, input := range bundle.Methodology.CountedInputs {
		if err := input.Validate(); err != nil {
			return fmt.Errorf("save benchmark evidence bundle: methodology countedInputs[%d]: %w", idx, err)
		}
	}
	if bundle.Methodology.TaskFinalArtifact != nil {
		if err := bundle.Methodology.TaskFinalArtifact.Validate(); err != nil {
			return fmt.Errorf("save benchmark evidence bundle: methodology taskFinalArtifact: %w", err)
		}
	}
	for idx, artifact := range bundle.Methodology.LaneFinalArtifacts {
		if err := artifact.Contract.Validate(); err != nil {
			return fmt.Errorf("save benchmark evidence bundle: methodology laneFinalArtifacts[%d]: %w", idx, err)
		}
	}
	for attemptIdx, attempt := range bundle.Attempts {
		for armIdx, arm := range attempt.Arms {
			for laneIdx, lane := range arm.Lanes {
				if lane.FinalArtifact != nil && strings.TrimSpace(lane.FinalArtifact.ContractID) == "" {
					return fmt.Errorf("save benchmark evidence bundle: attempts[%d].arms[%d].lanes[%d].finalArtifact.contractId is required", attemptIdx, armIdx, laneIdx)
				}
				for attrIdx, attribution := range lane.Attribution {
					switch attribution.Boundary {
					case repository.BenchmarkEvidenceBoundaryAgentInput, repository.BenchmarkEvidenceBoundarySystemProvenance, repository.BenchmarkEvidenceBoundaryFinalArtifactVerified:
					default:
						return fmt.Errorf("save benchmark evidence bundle: attempts[%d].arms[%d].lanes[%d].attribution[%d].boundary is required", attemptIdx, armIdx, laneIdx, attrIdx)
					}
				}
			}
		}
	}
	return nil
}

func saveBenchmarkEvidenceBundleRecord(ctx context.Context, tx *sql.Tx, repositoryID int64, bundle repository.BenchmarkEvidenceBundle) (BenchmarkEvidenceBundleRecord, error) {
	now := time.Now().UTC()
	metadataJSON := mustMarshalBenchmarkMetadata(map[string]any{
		"repositoryRoot":        bundle.RepositoryRoot,
		"fixtureID":             bundle.FixtureID,
		"fixturePath":           bundle.FixturePath,
		"methodology":           bundle.Methodology,
		"tokenEstimateContract": bundle.TokenEstimateContract,
		"verification":          bundle.Verification,
	})
	bundleJSONBytes, err := repository.MarshalBenchmarkEvidenceBundle(bundle)
	if err != nil {
		return BenchmarkEvidenceBundleRecord{}, fmt.Errorf("encode benchmark evidence bundle for suite %q: %w", bundle.SuiteID, err)
	}

	record := BenchmarkEvidenceBundleRecord{
		RepositoryID:           repositoryID,
		SchemaVersion:          bundle.SchemaVersion,
		SuiteID:                bundle.SuiteID,
		SuiteVersion:           bundle.SuiteVersion,
		SuiteSchemaVersion:     bundle.Methodology.SuiteSchemaVersion,
		CountedInputPolicy:     string(bundle.Methodology.Boundary.CountedInputs),
		SystemOutputPolicy:     string(bundle.Methodology.Boundary.SystemOutputs),
		FinalArtifactPolicy:    string(bundle.Methodology.Boundary.FinalArtifacts),
		CountedInputCount:      len(bundle.Methodology.CountedInputs),
		MethodologyFingerprint: bundle.MethodologyFingerprint,
		RerunCommand:           bundle.RerunCommand,
		GeneratedAt:            bundle.GeneratedAt.UTC(),
		MetadataJSON:           metadataJSON,
		BundleJSON:             string(bundleJSONBytes),
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO benchmark_evidence_bundles (
			repository_id,
			schema_version,
			suite_id,
			suite_version,
			suite_schema_version,
			counted_input_policy,
			system_output_policy,
			final_artifact_policy,
			counted_input_count,
			methodology_fingerprint,
			rerun_command,
			generated_at,
			metadata_json,
			bundle_json,
			created_at,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(repository_id, suite_id, suite_version, methodology_fingerprint) DO UPDATE SET
			suite_schema_version = excluded.suite_schema_version,
			counted_input_policy = excluded.counted_input_policy,
			system_output_policy = excluded.system_output_policy,
			final_artifact_policy = excluded.final_artifact_policy,
			counted_input_count = excluded.counted_input_count,
			rerun_command = excluded.rerun_command,
			generated_at = excluded.generated_at,
			metadata_json = excluded.metadata_json,
			bundle_json = excluded.bundle_json,
			updated_at = excluded.updated_at
	`,
		record.RepositoryID,
		record.SchemaVersion,
		record.SuiteID,
		record.SuiteVersion,
		record.SuiteSchemaVersion,
		record.CountedInputPolicy,
		record.SystemOutputPolicy,
		record.FinalArtifactPolicy,
		record.CountedInputCount,
		record.MethodologyFingerprint,
		record.RerunCommand,
		record.GeneratedAt.Format(time.RFC3339Nano),
		emptyToNil(record.MetadataJSON),
		record.BundleJSON,
		record.CreatedAt.Format(time.RFC3339Nano),
		record.UpdatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return BenchmarkEvidenceBundleRecord{}, fmt.Errorf("insert benchmark evidence bundle for suite %q: %w", bundle.SuiteID, err)
	}
	record.ID, err = result.LastInsertId()
	if err != nil || record.ID == 0 {
		var createdAt string
		err = tx.QueryRowContext(ctx, `
			SELECT id, created_at
			FROM benchmark_evidence_bundles
			WHERE repository_id = ? AND suite_id = ? AND suite_version = ? AND methodology_fingerprint = ?
		`, repositoryID, bundle.SuiteID, bundle.SuiteVersion, bundle.MethodologyFingerprint).Scan(&record.ID, &createdAt)
		if err != nil {
			return BenchmarkEvidenceBundleRecord{}, fmt.Errorf("load benchmark evidence bundle ID: %w", err)
		}
		if createdAt != "" {
			record.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
			if err != nil {
				return BenchmarkEvidenceBundleRecord{}, fmt.Errorf("parse benchmark evidence bundle created_at: %w", err)
			}
		}
	}
	return record, nil
}

func replaceBenchmarkEvidenceLaneSummaries(ctx context.Context, tx *sql.Tx, bundleID int64, summaries []repository.BenchmarkEvidenceArmSummary, attempts []repository.BenchmarkEvidenceAttempt) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM benchmark_evidence_lane_summaries WHERE benchmark_evidence_bundle_id = ?`, bundleID); err != nil {
		return fmt.Errorf("delete benchmark evidence lane summaries for bundle %d: %w", bundleID, err)
	}
	laneFinalArtifacts := latestBenchmarkLaneFinalArtifacts(attempts)
	for _, arm := range summaries {
		for _, lane := range arm.Lanes {
			metadataJSON := mustMarshalBenchmarkMetadata(map[string]any{
				"consultedArtifacts":     lane.ConsultedArtifacts,
				"rejectedAttemptReasons": lane.RejectedAttemptReasons,
			})
			finalArtifact := laneFinalArtifacts[lane.Lane]
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO benchmark_evidence_lane_summaries (
					benchmark_evidence_bundle_id,
					arm_kind,
					arm_name,
					lane,
					attempt_count,
					success_count,
					invalid_attempt_count,
					elapsed_ms_json,
					action_count_json,
					broad_search_actions_json,
					targeted_lookup_actions_json,
					file_read_actions_json,
					bytes_read_json,
					final_artifact_contract_id,
					final_artifact_path,
					final_artifact_passed,
					final_artifact_failure_reason,
					metadata_json
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`,
				bundleID,
				string(arm.ArmKind),
				arm.ArmName,
				string(lane.Lane),
				lane.AttemptCount,
				lane.SuccessCount,
				lane.InvalidAttemptCount,
				mustMarshalBenchmarkMetadata(lane.ElapsedMS),
				mustMarshalBenchmarkMetadata(lane.ActionCount),
				mustMarshalBenchmarkMetadata(lane.BroadSearchActions),
				mustMarshalBenchmarkMetadata(lane.TargetedLookupActions),
				mustMarshalBenchmarkMetadata(lane.FileReadActions),
				mustMarshalBenchmarkMetadata(lane.BytesRead),
				emptyToNil(finalArtifact.ContractID),
				emptyToNil(finalArtifact.Path),
				boolToInt(finalArtifact.Passed),
				emptyToNil(finalArtifact.FailureReason),
				emptyToNil(metadataJSON),
			); err != nil {
				return fmt.Errorf("insert benchmark evidence lane summary %q/%q: %w", arm.ArmKind, lane.Lane, err)
			}
		}
	}
	return nil
}

func latestBenchmarkLaneFinalArtifacts(attempts []repository.BenchmarkEvidenceAttempt) map[repository.BenchmarkLane]repository.BenchmarkLaneFinalArtifactVerification {
	finalArtifacts := make(map[repository.BenchmarkLane]repository.BenchmarkLaneFinalArtifactVerification)
	for _, attempt := range attempts {
		for _, arm := range attempt.Arms {
			for _, lane := range arm.Lanes {
				if lane.FinalArtifact == nil {
					continue
				}
				finalArtifacts[lane.Lane] = *lane.FinalArtifact
			}
		}
	}
	return finalArtifacts
}

func replaceBenchmarkEvidenceAttributions(ctx context.Context, tx *sql.Tx, bundleID int64, attempts []repository.BenchmarkEvidenceAttempt) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM benchmark_evidence_attributions WHERE benchmark_evidence_bundle_id = ?`, bundleID); err != nil {
		return fmt.Errorf("delete benchmark evidence attributions for bundle %d: %w", bundleID, err)
	}
	for _, attempt := range attempts {
		for _, arm := range attempt.Arms {
			for _, lane := range arm.Lanes {
				for ordinal, attribution := range lane.Attribution {
					metadataJSON := mustMarshalBenchmarkMetadata(map[string]any{
						"stepName": attribution.StepName,
					})
					if _, err := tx.ExecContext(ctx, `
						INSERT INTO benchmark_evidence_attributions (
							benchmark_evidence_bundle_id,
							attempt,
							arm_kind,
							lane,
							ordinal,
							step_id,
							step_name,
							boundary,
							counts_toward_tokens,
							surface,
							command_name,
							tool_name,
							artifact_type,
							report_label,
							source_kind,
							artifact_path,
							estimated_bytes,
							estimated_tokens,
							metadata_json
						) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
					`,
						bundleID,
						attempt.Attempt,
						string(arm.Kind),
						string(lane.Lane),
						ordinal,
						attribution.StepID,
						emptyToNil(attribution.StepName),
						string(attribution.Boundary),
						boolToInt(attribution.CountsTowardEstimatedTokens()),
						emptyToNil(string(attribution.Surface)),
						emptyToNil(string(attribution.Command)),
						emptyToNil(attribution.Tool),
						emptyToNil(string(attribution.ArtifactType)),
						emptyToNil(string(attribution.ReportLabel)),
						string(attribution.SourceKind),
						emptyToNil(attribution.ArtifactPath),
						attribution.EstimatedBytes,
						attribution.EstimatedTokens,
						emptyToNil(metadataJSON),
					); err != nil {
						return fmt.Errorf("insert benchmark evidence attribution %q/%q/%s: %w", arm.Kind, lane.Lane, attribution.StepID, err)
					}
				}
			}
		}
	}
	return nil
}

func saveBenchmarkRunRecord(ctx context.Context, tx *sql.Tx, run BenchmarkRunRecord) (BenchmarkRunRecord, error) {
	now := time.Now().UTC()
	if run.CreatedAt.IsZero() {
		run.CreatedAt = now
	}
	run.UpdatedAt = now

	result, err := tx.ExecContext(ctx, `
		INSERT INTO benchmark_runs (
			repository_id,
			suite_id,
			suite_version,
			fixture_id,
			fixture_path,
			arm_kind,
			arm_name,
			attempt,
			workspace_path,
			started_at,
			completed_at,
			metadata_json,
			created_at,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		run.RepositoryID,
		run.SuiteID,
		run.SuiteVersion,
		run.FixtureID,
		run.FixturePath,
		string(run.ArmKind),
		run.ArmName,
		run.Attempt,
		run.WorkspacePath,
		run.StartedAt.UTC().Format(time.RFC3339),
		run.CompletedAt.UTC().Format(time.RFC3339),
		emptyToNil(run.MetadataJSON),
		run.CreatedAt.UTC().Format(time.RFC3339),
		run.UpdatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return BenchmarkRunRecord{}, fmt.Errorf("insert benchmark run for suite %q: %w", run.SuiteID, err)
	}
	run.ID, err = result.LastInsertId()
	if err != nil {
		return BenchmarkRunRecord{}, fmt.Errorf("load benchmark run ID: %w", err)
	}
	return run, nil
}

func replaceBenchmarkLaneSamples(ctx context.Context, tx *sql.Tx, runID int64, samples []BenchmarkLaneSampleBundle) ([]BenchmarkLaneSampleBundle, error) {
	if _, err := tx.ExecContext(ctx, `DELETE FROM benchmark_lane_samples WHERE benchmark_run_id = ?`, runID); err != nil {
		return nil, fmt.Errorf("delete benchmark lane samples for run %d: %w", runID, err)
	}
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM benchmark_lane_metrics
		WHERE benchmark_lane_sample_id NOT IN (
			SELECT id FROM benchmark_lane_samples
		)
	`); err != nil {
		return nil, fmt.Errorf("delete orphaned benchmark lane metrics for run %d: %w", runID, err)
	}

	persisted := make([]BenchmarkLaneSampleBundle, 0, len(samples))
	for _, bundle := range samples {
		sample := bundle.Sample
		sample.BenchmarkRunID = runID
		record, err := insertBenchmarkLaneSample(ctx, tx, sample)
		if err != nil {
			return nil, err
		}
		metrics, err := insertBenchmarkLaneMetrics(ctx, tx, record.ID, bundle.Metrics)
		if err != nil {
			return nil, err
		}
		persisted = append(persisted, BenchmarkLaneSampleBundle{Sample: record, Metrics: metrics})
	}
	return persisted, nil
}

func insertBenchmarkLaneSample(ctx context.Context, tx *sql.Tx, sample BenchmarkLaneSampleRecord) (BenchmarkLaneSampleRecord, error) {
	result, err := tx.ExecContext(ctx, `
		INSERT INTO benchmark_lane_samples (
			benchmark_run_id,
			lane,
			start_marker,
			success_marker,
			stop_marker,
			started_at,
			finished_at,
			elapsed_ms,
			success,
			metadata_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		sample.BenchmarkRunID,
		string(sample.Lane),
		sample.StartMarker,
		sample.SuccessMarker,
		sample.StopMarker,
		sample.StartedAt.UTC().Format(time.RFC3339),
		sample.FinishedAt.UTC().Format(time.RFC3339),
		sample.ElapsedMS,
		boolToInt(sample.Success),
		emptyToNil(sample.MetadataJSON),
	)
	if err != nil {
		return BenchmarkLaneSampleRecord{}, fmt.Errorf("insert benchmark lane sample %q: %w", sample.Lane, err)
	}
	sample.ID, err = result.LastInsertId()
	if err != nil {
		return BenchmarkLaneSampleRecord{}, fmt.Errorf("load benchmark lane sample ID: %w", err)
	}
	return sample, nil
}

func insertBenchmarkLaneMetrics(ctx context.Context, tx *sql.Tx, sampleID int64, metrics []BenchmarkLaneMetricRecord) ([]BenchmarkLaneMetricRecord, error) {
	persisted := make([]BenchmarkLaneMetricRecord, 0, len(metrics))
	for _, metric := range metrics {
		if metric.MetricName == "" {
			return nil, fmt.Errorf("insert benchmark lane metric: metric name is required")
		}
		metric.BenchmarkLaneSampleID = sampleID
		result, err := tx.ExecContext(ctx, `
			INSERT INTO benchmark_lane_metrics (
				benchmark_lane_sample_id,
				metric_name,
				ordinal,
				value_int,
				value_text,
				metadata_json
			) VALUES (?, ?, ?, ?, ?, ?)
		`,
			metric.BenchmarkLaneSampleID,
			metric.MetricName,
			metric.Ordinal,
			metric.ValueInt,
			emptyToNil(metric.ValueText),
			emptyToNil(metric.MetadataJSON),
		)
		if err != nil {
			return nil, fmt.Errorf("insert benchmark lane metric %q: %w", metric.MetricName, err)
		}
		metric.ID, err = result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("load benchmark lane metric ID: %w", err)
		}
		persisted = append(persisted, metric)
	}
	return persisted, nil
}

func loadBenchmarkRunRecord(ctx context.Context, queryer sqlQueryer, runID int64) (BenchmarkRunRecord, error) {
	var run BenchmarkRunRecord
	var armKind string
	var startedAt, completedAt, metadataJSON, createdAt, updatedAt sql.NullString
	err := queryer.QueryRowContext(ctx, `
		SELECT
			id,
			repository_id,
			suite_id,
			suite_version,
			fixture_id,
			fixture_path,
			arm_kind,
			arm_name,
			attempt,
			workspace_path,
			started_at,
			completed_at,
			metadata_json,
			created_at,
			updated_at
		FROM benchmark_runs
		WHERE id = ?
	`, runID).Scan(
		&run.ID,
		&run.RepositoryID,
		&run.SuiteID,
		&run.SuiteVersion,
		&run.FixtureID,
		&run.FixturePath,
		&armKind,
		&run.ArmName,
		&run.Attempt,
		&run.WorkspacePath,
		&startedAt,
		&completedAt,
		&metadataJSON,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return BenchmarkRunRecord{}, fmt.Errorf("load benchmark run %d: %w", runID, err)
	}
	run.ArmKind = repository.BenchmarkArmKind(armKind)
	run.StartedAt = parseOptionalRFC3339(startedAt)
	run.CompletedAt = parseOptionalRFC3339(completedAt)
	run.MetadataJSON = metadataJSON.String
	run.CreatedAt = parseOptionalRFC3339(createdAt)
	run.UpdatedAt = parseOptionalRFC3339(updatedAt)
	return run, nil
}

func listBenchmarkLaneSamples(ctx context.Context, queryer sqlQueryer, runID int64) ([]BenchmarkLaneSampleBundle, error) {
	rows, err := queryer.QueryContext(ctx, `
		SELECT
			id,
			benchmark_run_id,
			lane,
			start_marker,
			success_marker,
			stop_marker,
			started_at,
			finished_at,
			elapsed_ms,
			success,
			metadata_json
		FROM benchmark_lane_samples
		WHERE benchmark_run_id = ?
		ORDER BY id
	`, runID)
	if err != nil {
		return nil, fmt.Errorf("list benchmark lane samples for run %d: %w", runID, err)
	}
	defer rows.Close()

	var bundles []BenchmarkLaneSampleBundle
	for rows.Next() {
		var sample BenchmarkLaneSampleRecord
		var lane string
		var startedAt, finishedAt, metadataJSON sql.NullString
		var success int
		if err := rows.Scan(
			&sample.ID,
			&sample.BenchmarkRunID,
			&lane,
			&sample.StartMarker,
			&sample.SuccessMarker,
			&sample.StopMarker,
			&startedAt,
			&finishedAt,
			&sample.ElapsedMS,
			&success,
			&metadataJSON,
		); err != nil {
			return nil, fmt.Errorf("scan benchmark lane sample for run %d: %w", runID, err)
		}
		sample.Lane = repository.BenchmarkLane(lane)
		sample.StartedAt = parseOptionalRFC3339(startedAt)
		sample.FinishedAt = parseOptionalRFC3339(finishedAt)
		sample.Success = success != 0
		sample.MetadataJSON = metadataJSON.String
		metrics, err := listBenchmarkLaneMetrics(ctx, queryer, sample.ID)
		if err != nil {
			return nil, err
		}
		bundles = append(bundles, BenchmarkLaneSampleBundle{Sample: sample, Metrics: metrics})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate benchmark lane samples for run %d: %w", runID, err)
	}
	return bundles, nil
}

func listBenchmarkLaneMetrics(ctx context.Context, queryer sqlQueryer, sampleID int64) ([]BenchmarkLaneMetricRecord, error) {
	rows, err := queryer.QueryContext(ctx, `
		SELECT
			id,
			benchmark_lane_sample_id,
			metric_name,
			ordinal,
			value_int,
			value_text,
			metadata_json
		FROM benchmark_lane_metrics
		WHERE benchmark_lane_sample_id = ?
		ORDER BY metric_name, ordinal, id
	`, sampleID)
	if err != nil {
		return nil, fmt.Errorf("list benchmark lane metrics for sample %d: %w", sampleID, err)
	}
	defer rows.Close()

	var metrics []BenchmarkLaneMetricRecord
	for rows.Next() {
		var metric BenchmarkLaneMetricRecord
		var valueText, metadataJSON sql.NullString
		if err := rows.Scan(
			&metric.ID,
			&metric.BenchmarkLaneSampleID,
			&metric.MetricName,
			&metric.Ordinal,
			&metric.ValueInt,
			&valueText,
			&metadataJSON,
		); err != nil {
			return nil, fmt.Errorf("scan benchmark lane metric for sample %d: %w", sampleID, err)
		}
		metric.ValueText = valueText.String
		metric.MetadataJSON = metadataJSON.String
		metrics = append(metrics, metric)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate benchmark lane metrics for sample %d: %w", sampleID, err)
	}
	return metrics, nil
}
