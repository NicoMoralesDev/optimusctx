package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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
			MetadataJSON:  mustMarshalBenchmarkMetadata(map[string]any{"workspacePath": arm.Workspace}),
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
