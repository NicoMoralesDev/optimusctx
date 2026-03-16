package sqlite

import (
	"context"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/niccrow/optimusctx/internal/repository"
	"github.com/niccrow/optimusctx/internal/state"
)

func TestApplyMigrationsCreatesBenchmarkTables(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	assertTablesExist(t, store.DB(), "benchmark_runs", "benchmark_lane_samples", "benchmark_lane_metrics")
	assertIndexColumns(t, store.DB(), "benchmark_runs", []string{"repository_id", "suite_id", "started_at"})
	assertIndexColumns(t, store.DB(), "benchmark_runs", []string{"repository_id", "arm_kind", "started_at"})
	assertIndexColumns(t, store.DB(), "benchmark_lane_samples", []string{"benchmark_run_id", "lane"})
	assertIndexColumns(t, store.DB(), "benchmark_lane_metrics", []string{"benchmark_lane_sample_id", "metric_name"})

	if _, _, err := store.SaveBenchmarkRun(ctx, BenchmarkRunRecord{
		RepositoryID:  repoID,
		SuiteID:       "go-benchmark-discovery-v1",
		SuiteVersion:  "v1",
		FixtureID:     "go-benchmark",
		FixturePath:   "go-benchmark/v1/repository",
		ArmKind:       repository.BenchmarkArmKindBaseline,
		ArmName:       "Baseline",
		Attempt:       1,
		WorkspacePath: layout.RepoRoot,
		StartedAt:     time.Date(2026, 3, 16, 15, 0, 0, 0, time.UTC),
		CompletedAt:   time.Date(2026, 3, 16, 15, 0, 1, 0, time.UTC),
	}, nil); err != nil {
		t.Fatalf("SaveBenchmarkRun() error = %v", err)
	}
}

func TestBenchmarkLaneMetricsPersist(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	layout, err := state.ResolveLayout(t.TempDir())
	if err != nil {
		t.Fatalf("ResolveLayout() error = %v", err)
	}

	store, repoID := openStoreWithRepository(t, ctx, layout)
	defer store.Close()

	runStarted := time.Date(2026, 3, 16, 16, 0, 0, 0, time.UTC)
	runCompleted := runStarted.Add(2 * time.Second)
	run, samples, err := store.SaveBenchmarkRun(ctx, BenchmarkRunRecord{
		RepositoryID:  repoID,
		SuiteID:       "go-benchmark-discovery-v1",
		SuiteVersion:  "v1",
		FixtureID:     "go-benchmark",
		FixturePath:   "go-benchmark/v1/repository",
		ArmKind:       repository.BenchmarkArmKindOptimusCtx,
		ArmName:       "OptimusCtx MCP workflow",
		Attempt:       2,
		WorkspacePath: layout.RepoRoot,
		StartedAt:     runStarted,
		CompletedAt:   runCompleted,
		MetadataJSON:  `{"source":"runner"}`,
	}, []BenchmarkLaneSampleBundle{
		{
			Sample: BenchmarkLaneSampleRecord{
				Lane:          repository.BenchmarkLaneDiscovery,
				StartMarker:   "discovery_started",
				SuccessMarker: "target_identified",
				StopMarker:    "target_identified",
				StartedAt:     runStarted,
				FinishedAt:    runStarted.Add(500 * time.Millisecond),
				ElapsedMS:     500,
				Success:       true,
			},
			Metrics: []BenchmarkLaneMetricRecord{
				{MetricName: benchmarkMetricActionCount, ValueInt: 2},
				{MetricName: string(repository.BenchmarkMetricBroadSearchActions), ValueInt: 1},
				{MetricName: string(repository.BenchmarkMetricTargetedLookupActions), ValueInt: 1},
				{MetricName: string(repository.BenchmarkMetricConsultedArtifacts), Ordinal: 0, ValueInt: 1, ValueText: "internal/http/handler/rollout.go"},
			},
		},
		{
			Sample: BenchmarkLaneSampleRecord{
				Lane:          repository.BenchmarkLaneContextAssembly,
				StartMarker:   "context_started",
				SuccessMarker: "context_ready",
				StopMarker:    "context_ready",
				StartedAt:     runStarted.Add(750 * time.Millisecond),
				FinishedAt:    runStarted.Add(1400 * time.Millisecond),
				ElapsedMS:     650,
				Success:       true,
			},
			Metrics: []BenchmarkLaneMetricRecord{
				{MetricName: benchmarkMetricActionCount, ValueInt: 1},
				{MetricName: string(repository.BenchmarkMetricFileReadActions), ValueInt: 1},
				{MetricName: string(repository.BenchmarkMetricBytesRead), ValueInt: 144},
				{MetricName: string(repository.BenchmarkMetricConsultedArtifacts), Ordinal: 0, ValueInt: 1, ValueText: "internal/config/loader.go"},
			},
		},
	})
	if err != nil {
		t.Fatalf("SaveBenchmarkRun() error = %v", err)
	}

	gotRun, gotSamples, err := store.LoadBenchmarkRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("LoadBenchmarkRun() error = %v", err)
	}

	run.CreatedAt = gotRun.CreatedAt
	run.UpdatedAt = gotRun.UpdatedAt
	if !reflect.DeepEqual(gotRun, run) {
		t.Fatalf("run mismatch\n got=%+v\nwant=%+v", gotRun, run)
	}
	if len(gotSamples) != 2 {
		t.Fatalf("len(gotSamples) = %d, want 2", len(gotSamples))
	}

	for idx := range samples {
		samples[idx].Sample.ID = gotSamples[idx].Sample.ID
		samples[idx].Sample.BenchmarkRunID = run.ID
		samples[idx].Sample.StartedAt = gotSamples[idx].Sample.StartedAt
		samples[idx].Sample.FinishedAt = gotSamples[idx].Sample.FinishedAt
		samples[idx].Metrics = reorderBenchmarkMetricsForLoad(samples[idx].Metrics)
		for metricIdx := range samples[idx].Metrics {
			samples[idx].Metrics[metricIdx].ID = gotSamples[idx].Metrics[metricIdx].ID
			samples[idx].Metrics[metricIdx].BenchmarkLaneSampleID = gotSamples[idx].Sample.ID
		}
	}
	if !reflect.DeepEqual(gotSamples, samples) {
		t.Fatalf("samples mismatch\n got=%+v\nwant=%+v", gotSamples, samples)
	}
}

func reorderBenchmarkMetricsForLoad(metrics []BenchmarkLaneMetricRecord) []BenchmarkLaneMetricRecord {
	ordered := append([]BenchmarkLaneMetricRecord(nil), metrics...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].MetricName == ordered[j].MetricName {
			if ordered[i].Ordinal == ordered[j].Ordinal {
				return ordered[i].ID < ordered[j].ID
			}
			return ordered[i].Ordinal < ordered[j].Ordinal
		}
		return ordered[i].MetricName < ordered[j].MetricName
	})
	return ordered
}
