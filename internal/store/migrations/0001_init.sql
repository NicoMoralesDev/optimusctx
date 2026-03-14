CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at TEXT NOT NULL
);

CREATE TABLE repositories (
    id INTEGER PRIMARY KEY,
    root_path TEXT NOT NULL UNIQUE,
    root_real_path TEXT NOT NULL,
    detection_mode TEXT NOT NULL,
    git_common_dir TEXT,
    git_head_ref TEXT,
    git_head_commit TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE directories (
    id INTEGER PRIMARY KEY,
    repository_id INTEGER NOT NULL,
    path TEXT NOT NULL,
    parent_path TEXT,
    discovered_at TEXT NOT NULL,
    ignore_status TEXT NOT NULL,
    ignore_reason TEXT,
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE,
    UNIQUE (repository_id, path)
);

CREATE TABLE files (
    id INTEGER PRIMARY KEY,
    repository_id INTEGER NOT NULL,
    path TEXT NOT NULL,
    directory_path TEXT NOT NULL,
    extension TEXT,
    language TEXT,
    size_bytes INTEGER NOT NULL,
    content_hash TEXT,
    last_indexed_at TEXT,
    ignore_status TEXT NOT NULL,
    ignore_reason TEXT,
    fs_mod_time TEXT,
    discovered_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE,
    UNIQUE (repository_id, path)
);

CREATE INDEX idx_directories_repository_path
    ON directories (repository_id, path);

CREATE INDEX idx_directories_repository_ignore_status
    ON directories (repository_id, ignore_status);

CREATE INDEX idx_files_repository_directory_path
    ON files (repository_id, directory_path);

CREATE INDEX idx_files_repository_ignore_status
    ON files (repository_id, ignore_status);

CREATE INDEX idx_files_repository_language
    ON files (repository_id, language);
