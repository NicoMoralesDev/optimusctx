ALTER TABLE benchmark_evidence_bundles
    ADD COLUMN suite_schema_version TEXT NOT NULL DEFAULT '';

ALTER TABLE benchmark_evidence_bundles
    ADD COLUMN counted_input_policy TEXT NOT NULL DEFAULT '';

ALTER TABLE benchmark_evidence_bundles
    ADD COLUMN system_output_policy TEXT NOT NULL DEFAULT '';

ALTER TABLE benchmark_evidence_bundles
    ADD COLUMN final_artifact_policy TEXT NOT NULL DEFAULT '';

ALTER TABLE benchmark_evidence_bundles
    ADD COLUMN counted_input_count INTEGER NOT NULL DEFAULT 0;

ALTER TABLE benchmark_evidence_lane_summaries
    ADD COLUMN final_artifact_contract_id TEXT;

ALTER TABLE benchmark_evidence_lane_summaries
    ADD COLUMN final_artifact_path TEXT;

ALTER TABLE benchmark_evidence_lane_summaries
    ADD COLUMN final_artifact_passed INTEGER NOT NULL DEFAULT 0;

ALTER TABLE benchmark_evidence_lane_summaries
    ADD COLUMN final_artifact_failure_reason TEXT;

ALTER TABLE benchmark_evidence_attributions
    ADD COLUMN boundary TEXT NOT NULL DEFAULT '';

ALTER TABLE benchmark_evidence_attributions
    ADD COLUMN counts_toward_tokens INTEGER NOT NULL DEFAULT 0;

UPDATE benchmark_evidence_bundles
SET
    suite_schema_version = 'optimusctx/benchmark-suite@v2',
    counted_input_policy = 'declared_agent_inputs_only',
    system_output_policy = 'persist_system_outputs',
    final_artifact_policy = 'required_per_lane_or_task',
    counted_input_count = 0
WHERE suite_schema_version = '';

UPDATE benchmark_evidence_attributions
SET
    boundary = 'system_provenance',
    counts_toward_tokens = 0
WHERE boundary = '';

CREATE INDEX idx_benchmark_evidence_bundles_repository_suite_methodology_v2
    ON benchmark_evidence_bundles (
        repository_id,
        suite_id,
        suite_version,
        suite_schema_version,
        counted_input_policy,
        system_output_policy,
        final_artifact_policy
    );

CREATE INDEX idx_benchmark_evidence_lane_summaries_bundle_lane_final_artifact
    ON benchmark_evidence_lane_summaries (
        benchmark_evidence_bundle_id,
        lane,
        final_artifact_contract_id
    );

CREATE INDEX idx_benchmark_evidence_attributions_bundle_boundary
    ON benchmark_evidence_attributions (
        benchmark_evidence_bundle_id,
        boundary,
        counts_toward_tokens
    );
