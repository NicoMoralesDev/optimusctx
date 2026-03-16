CREATE TABLE eval_runs (
    id INTEGER PRIMARY KEY,
    repository_id INTEGER NOT NULL,
    scenario_id TEXT NOT NULL,
    scenario_version TEXT NOT NULL,
    fixture_id TEXT NOT NULL,
    fixture_version TEXT NOT NULL,
    status TEXT NOT NULL,
    passed INTEGER NOT NULL DEFAULT 0,
    workspace_path TEXT NOT NULL,
    artifact_root TEXT NOT NULL,
    started_at TEXT NOT NULL,
    completed_at TEXT,
    metadata_json TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE,
    CHECK (status IN ('running', 'passed', 'failed', 'error')),
    CHECK (passed IN (0, 1))
);

CREATE TABLE eval_steps (
    id INTEGER PRIMARY KEY,
    eval_run_id INTEGER NOT NULL,
    step_id TEXT NOT NULL,
    ordinal INTEGER NOT NULL,
    name TEXT NOT NULL,
    surface TEXT NOT NULL,
    command TEXT NOT NULL,
    args_json TEXT,
    exit_code INTEGER NOT NULL,
    passed INTEGER NOT NULL DEFAULT 0,
    started_at TEXT NOT NULL,
    finished_at TEXT NOT NULL,
    stdout_path TEXT,
    stderr_path TEXT,
    metadata_json TEXT,
    FOREIGN KEY (eval_run_id) REFERENCES eval_runs(id) ON DELETE CASCADE,
    UNIQUE (eval_run_id, step_id),
    UNIQUE (eval_run_id, ordinal),
    CHECK (ordinal >= 0),
    CHECK (passed IN (0, 1))
);

CREATE TABLE eval_artifacts (
    id INTEGER PRIMARY KEY,
    eval_run_id INTEGER NOT NULL,
    step_id TEXT,
    artifact_id TEXT NOT NULL,
    kind TEXT NOT NULL,
    logical_path TEXT,
    stored_path TEXT NOT NULL,
    required INTEGER NOT NULL DEFAULT 0,
    present INTEGER NOT NULL DEFAULT 0,
    size_bytes INTEGER NOT NULL DEFAULT 0,
    metadata_json TEXT,
    FOREIGN KEY (eval_run_id) REFERENCES eval_runs(id) ON DELETE CASCADE,
    UNIQUE (eval_run_id, step_id, artifact_id),
    CHECK (kind IN ('stdout', 'stderr', 'file')),
    CHECK (required IN (0, 1)),
    CHECK (present IN (0, 1)),
    CHECK (size_bytes >= 0)
);

CREATE INDEX idx_eval_runs_repository_started_at
    ON eval_runs (repository_id, started_at DESC);

CREATE INDEX idx_eval_runs_repository_scenario
    ON eval_runs (repository_id, scenario_id, started_at DESC);

CREATE INDEX idx_eval_steps_run_ordinal
    ON eval_steps (eval_run_id, ordinal);

CREATE INDEX idx_eval_artifacts_run_step
    ON eval_artifacts (eval_run_id, step_id);
