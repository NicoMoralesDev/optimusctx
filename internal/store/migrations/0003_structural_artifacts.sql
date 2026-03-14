CREATE TABLE file_extractions (
    id INTEGER PRIMARY KEY,
    repository_id INTEGER NOT NULL,
    file_id INTEGER NOT NULL UNIQUE,
    path TEXT NOT NULL,
    language TEXT NOT NULL,
    adapter_name TEXT NOT NULL,
    grammar_version TEXT NOT NULL,
    source_content_hash TEXT NOT NULL,
    source_generation INTEGER NOT NULL,
    coverage_state TEXT NOT NULL,
    coverage_reason TEXT,
    parser_error_count INTEGER NOT NULL DEFAULT 0,
    has_error_nodes INTEGER NOT NULL DEFAULT 0,
    symbol_count INTEGER NOT NULL DEFAULT 0,
    top_level_symbol_count INTEGER NOT NULL DEFAULT 0,
    max_symbol_depth INTEGER NOT NULL DEFAULT 0,
    extracted_at TEXT NOT NULL,
    refresh_run_id INTEGER,
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
    FOREIGN KEY (refresh_run_id) REFERENCES refresh_runs(id) ON DELETE SET NULL,
    CHECK (coverage_state IN ('supported', 'partial', 'unsupported', 'failed', 'skipped')),
    CHECK (parser_error_count >= 0),
    CHECK (has_error_nodes IN (0, 1)),
    CHECK (symbol_count >= 0),
    CHECK (top_level_symbol_count >= 0),
    CHECK (max_symbol_depth >= 0)
);

CREATE TABLE symbols (
    id INTEGER PRIMARY KEY,
    repository_id INTEGER NOT NULL,
    file_id INTEGER NOT NULL,
    file_extraction_id INTEGER NOT NULL,
    stable_key TEXT NOT NULL,
    parent_symbol_id INTEGER,
    path TEXT NOT NULL,
    language TEXT NOT NULL,
    kind TEXT NOT NULL,
    name TEXT NOT NULL,
    qualified_name TEXT,
    ordinal INTEGER NOT NULL,
    depth INTEGER NOT NULL,
    start_byte INTEGER NOT NULL,
    end_byte INTEGER NOT NULL,
    start_row INTEGER NOT NULL,
    start_column INTEGER NOT NULL,
    end_row INTEGER NOT NULL,
    end_column INTEGER NOT NULL,
    name_start_byte INTEGER,
    name_end_byte INTEGER,
    signature_start_byte INTEGER,
    signature_end_byte INTEGER,
    is_exported INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
    FOREIGN KEY (file_extraction_id) REFERENCES file_extractions(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_symbol_id) REFERENCES symbols(id) ON DELETE CASCADE,
    UNIQUE (file_extraction_id, stable_key),
    CHECK (ordinal >= 0),
    CHECK (depth >= 0),
    CHECK (start_byte >= 0),
    CHECK (end_byte >= start_byte),
    CHECK (start_row >= 0),
    CHECK (start_column >= 0),
    CHECK (end_row >= start_row),
    CHECK (end_column >= 0),
    CHECK (is_exported IN (0, 1))
);

CREATE INDEX idx_file_extractions_repository_path
    ON file_extractions (repository_id, path);

CREATE INDEX idx_file_extractions_repository_coverage_state
    ON file_extractions (repository_id, coverage_state);

CREATE INDEX idx_file_extractions_repository_language
    ON file_extractions (repository_id, language);

CREATE INDEX idx_file_extractions_repository_generation
    ON file_extractions (repository_id, source_generation);

CREATE INDEX idx_symbols_repository_name
    ON symbols (repository_id, name);

CREATE INDEX idx_symbols_repository_kind
    ON symbols (repository_id, kind);

CREATE INDEX idx_symbols_file_ordinal
    ON symbols (file_id, ordinal);

CREATE INDEX idx_symbols_repository_path_ordinal
    ON symbols (repository_id, path, ordinal);

CREATE INDEX idx_symbols_repository_qualified_name
    ON symbols (repository_id, qualified_name);
