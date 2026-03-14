ALTER TABLE repositories ADD COLUMN last_refresh_started_at TEXT;
ALTER TABLE repositories ADD COLUMN last_refresh_completed_at TEXT;
ALTER TABLE repositories ADD COLUMN last_refresh_reason TEXT;
ALTER TABLE repositories ADD COLUMN last_refresh_status TEXT NOT NULL DEFAULT 'pending';
ALTER TABLE repositories ADD COLUMN freshness_status TEXT NOT NULL DEFAULT 'stale';
ALTER TABLE repositories ADD COLUMN freshness_reason TEXT;
ALTER TABLE repositories ADD COLUMN current_refresh_generation INTEGER NOT NULL DEFAULT 0;
ALTER TABLE repositories ADD COLUMN last_refresh_generation INTEGER NOT NULL DEFAULT 0;

ALTER TABLE directories ADD COLUMN subtree_fingerprint TEXT;
ALTER TABLE directories ADD COLUMN included_file_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE directories ADD COLUMN included_directory_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE directories ADD COLUMN total_size_bytes INTEGER NOT NULL DEFAULT 0;
ALTER TABLE directories ADD COLUMN last_refreshed_at TEXT;
ALTER TABLE directories ADD COLUMN last_refresh_generation INTEGER NOT NULL DEFAULT 0;

ALTER TABLE files ADD COLUMN last_seen_generation INTEGER NOT NULL DEFAULT 0;
ALTER TABLE files ADD COLUMN refresh_run_id INTEGER;
ALTER TABLE files ADD COLUMN updated_reason TEXT;

CREATE TABLE refresh_runs (
    id INTEGER PRIMARY KEY,
    repository_id INTEGER NOT NULL,
    generation INTEGER NOT NULL,
    reason TEXT NOT NULL,
    status TEXT NOT NULL,
    failure_reason TEXT,
    started_at TEXT NOT NULL,
    completed_at TEXT,
    metadata_json TEXT,
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE,
    UNIQUE (repository_id, generation)
);

CREATE TABLE refresh_file_events (
    id INTEGER PRIMARY KEY,
    repository_id INTEGER NOT NULL,
    refresh_run_id INTEGER NOT NULL,
    path TEXT NOT NULL,
    previous_path TEXT,
    event_type TEXT NOT NULL,
    content_hash TEXT,
    occurred_at TEXT NOT NULL,
    metadata_json TEXT,
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE,
    FOREIGN KEY (refresh_run_id) REFERENCES refresh_runs(id) ON DELETE CASCADE
);

CREATE INDEX idx_repositories_root_path
    ON repositories (root_path);

CREATE INDEX idx_repositories_freshness_status
    ON repositories (freshness_status);

CREATE INDEX idx_directories_repository_refresh_lookup
    ON directories (repository_id, path);

CREATE INDEX idx_directories_repository_generation
    ON directories (repository_id, last_refresh_generation);

CREATE INDEX idx_files_repository_path
    ON files (repository_id, path);

CREATE INDEX idx_files_repository_generation
    ON files (repository_id, last_seen_generation);

CREATE INDEX idx_refresh_runs_repository_started_at
    ON refresh_runs (repository_id, started_at DESC);

CREATE INDEX idx_refresh_runs_repository_status
    ON refresh_runs (repository_id, status);

CREATE INDEX idx_refresh_file_events_run_path
    ON refresh_file_events (refresh_run_id, path);

CREATE INDEX idx_refresh_file_events_repository_path
    ON refresh_file_events (repository_id, path);
