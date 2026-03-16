CREATE TABLE benchmark_runs (
    id INTEGER PRIMARY KEY,
    repository_id INTEGER NOT NULL,
    suite_id TEXT NOT NULL,
    suite_version TEXT NOT NULL,
    fixture_id TEXT NOT NULL,
    fixture_path TEXT NOT NULL,
    arm_kind TEXT NOT NULL,
    arm_name TEXT NOT NULL,
    attempt INTEGER NOT NULL DEFAULT 1,
    workspace_path TEXT NOT NULL,
    started_at TEXT NOT NULL,
    completed_at TEXT NOT NULL,
    metadata_json TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE,
    UNIQUE (repository_id, suite_id, suite_version, arm_kind, attempt),
    CHECK (arm_kind IN ('baseline', 'optimusctx')),
    CHECK (attempt >= 1)
);

CREATE TABLE benchmark_lane_samples (
    id INTEGER PRIMARY KEY,
    benchmark_run_id INTEGER NOT NULL,
    lane TEXT NOT NULL,
    start_marker TEXT NOT NULL,
    success_marker TEXT NOT NULL,
    stop_marker TEXT NOT NULL,
    started_at TEXT NOT NULL,
    finished_at TEXT NOT NULL,
    elapsed_ms INTEGER NOT NULL DEFAULT 0,
    success INTEGER NOT NULL DEFAULT 0,
    metadata_json TEXT,
    FOREIGN KEY (benchmark_run_id) REFERENCES benchmark_runs(id) ON DELETE CASCADE,
    UNIQUE (benchmark_run_id, lane),
    CHECK (lane IN ('discovery', 'context_assembly', 'refresh_after_change', 'task_completion')),
    CHECK (elapsed_ms >= 0),
    CHECK (success IN (0, 1))
);

CREATE TABLE benchmark_lane_metrics (
    id INTEGER PRIMARY KEY,
    benchmark_lane_sample_id INTEGER NOT NULL,
    metric_name TEXT NOT NULL,
    ordinal INTEGER NOT NULL DEFAULT 0,
    value_int INTEGER NOT NULL DEFAULT 0,
    value_text TEXT,
    metadata_json TEXT,
    FOREIGN KEY (benchmark_lane_sample_id) REFERENCES benchmark_lane_samples(id) ON DELETE CASCADE,
    UNIQUE (benchmark_lane_sample_id, metric_name, ordinal),
    CHECK (metric_name <> ''),
    CHECK (ordinal >= 0),
    CHECK (value_int >= 0)
);

CREATE INDEX idx_benchmark_runs_repository_suite_started_at
    ON benchmark_runs (repository_id, suite_id, started_at DESC);

CREATE INDEX idx_benchmark_runs_repository_arm_started_at
    ON benchmark_runs (repository_id, arm_kind, started_at DESC);

CREATE INDEX idx_benchmark_lane_samples_run_lane
    ON benchmark_lane_samples (benchmark_run_id, lane);

CREATE INDEX idx_benchmark_lane_metrics_sample_metric
    ON benchmark_lane_metrics (benchmark_lane_sample_id, metric_name);
