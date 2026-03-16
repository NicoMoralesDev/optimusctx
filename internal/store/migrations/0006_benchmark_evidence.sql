CREATE TABLE benchmark_evidence_bundles (
    id INTEGER PRIMARY KEY,
    repository_id INTEGER NOT NULL,
    schema_version TEXT NOT NULL,
    suite_id TEXT NOT NULL,
    suite_version TEXT NOT NULL,
    methodology_fingerprint TEXT NOT NULL,
    rerun_command TEXT NOT NULL,
    generated_at TEXT NOT NULL,
    metadata_json TEXT,
    bundle_json TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE,
    UNIQUE (repository_id, suite_id, suite_version, methodology_fingerprint),
    CHECK (schema_version <> ''),
    CHECK (suite_id <> ''),
    CHECK (suite_version <> ''),
    CHECK (methodology_fingerprint <> ''),
    CHECK (rerun_command <> ''),
    CHECK (bundle_json <> '')
);

CREATE TABLE benchmark_evidence_lane_summaries (
    id INTEGER PRIMARY KEY,
    benchmark_evidence_bundle_id INTEGER NOT NULL,
    arm_kind TEXT NOT NULL,
    arm_name TEXT NOT NULL,
    lane TEXT NOT NULL,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    success_count INTEGER NOT NULL DEFAULT 0,
    invalid_attempt_count INTEGER NOT NULL DEFAULT 0,
    elapsed_ms_json TEXT NOT NULL,
    action_count_json TEXT NOT NULL,
    broad_search_actions_json TEXT NOT NULL,
    targeted_lookup_actions_json TEXT NOT NULL,
    file_read_actions_json TEXT NOT NULL,
    bytes_read_json TEXT NOT NULL,
    metadata_json TEXT,
    FOREIGN KEY (benchmark_evidence_bundle_id) REFERENCES benchmark_evidence_bundles(id) ON DELETE CASCADE,
    UNIQUE (benchmark_evidence_bundle_id, arm_kind, lane),
    CHECK (arm_kind IN ('baseline', 'optimusctx')),
    CHECK (lane IN ('discovery', 'context_assembly', 'refresh_after_change', 'task_completion')),
    CHECK (attempt_count >= 0),
    CHECK (success_count >= 0),
    CHECK (invalid_attempt_count >= 0)
);

CREATE TABLE benchmark_evidence_attributions (
    id INTEGER PRIMARY KEY,
    benchmark_evidence_bundle_id INTEGER NOT NULL,
    attempt INTEGER NOT NULL,
    arm_kind TEXT NOT NULL,
    lane TEXT NOT NULL,
    ordinal INTEGER NOT NULL DEFAULT 0,
    step_id TEXT NOT NULL,
    step_name TEXT,
    surface TEXT,
    command_name TEXT,
    tool_name TEXT,
    artifact_type TEXT,
    report_label TEXT,
    source_kind TEXT NOT NULL,
    artifact_path TEXT,
    estimated_bytes INTEGER NOT NULL DEFAULT 0,
    estimated_tokens INTEGER NOT NULL DEFAULT 0,
    metadata_json TEXT,
    FOREIGN KEY (benchmark_evidence_bundle_id) REFERENCES benchmark_evidence_bundles(id) ON DELETE CASCADE,
    UNIQUE (benchmark_evidence_bundle_id, attempt, arm_kind, lane, ordinal),
    CHECK (attempt >= 1),
    CHECK (arm_kind IN ('baseline', 'optimusctx')),
    CHECK (lane IN ('discovery', 'context_assembly', 'refresh_after_change', 'task_completion')),
    CHECK (ordinal >= 0),
    CHECK (step_id <> ''),
    CHECK (source_kind <> ''),
    CHECK (estimated_bytes >= 0),
    CHECK (estimated_tokens >= 0)
);

CREATE INDEX idx_benchmark_evidence_bundles_repository_suite_generated_at
    ON benchmark_evidence_bundles (repository_id, suite_id, suite_version, generated_at DESC);

CREATE INDEX idx_benchmark_evidence_lane_summaries_bundle_arm_lane
    ON benchmark_evidence_lane_summaries (benchmark_evidence_bundle_id, arm_kind, lane);

CREATE INDEX idx_benchmark_evidence_attributions_bundle_attempt_lane
    ON benchmark_evidence_attributions (benchmark_evidence_bundle_id, attempt, arm_kind, lane);
