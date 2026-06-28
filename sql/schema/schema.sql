-- sqlc schema mirror of migrations for code generation

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    display_name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_ai_settings (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    ai_provider TEXT NOT NULL DEFAULT 'openai',
    ai_model TEXT NOT NULL DEFAULT 'gpt-4o-mini',
    openai_api_key TEXT NOT NULL DEFAULT '',
    anthropic_api_key TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE assignment_snapshots (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    source_url TEXT NOT NULL,
    source_platform TEXT NOT NULL DEFAULT 'canvas',
    page_type TEXT,
    course_name TEXT,
    assignment_title TEXT,
    raw_text TEXT,
    instructions_text TEXT,
    due_at TIMESTAMPTZ,
    points_possible NUMERIC(10, 2),
    submission_type TEXT,
    allowed_submission_types JSONB,
    imported_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_assignment_snapshots_imported_at
    ON assignment_snapshots (imported_at DESC);

CREATE INDEX idx_assignment_snapshots_due_at
    ON assignment_snapshots (due_at)
    WHERE due_at IS NOT NULL;

CREATE UNIQUE INDEX idx_assignment_snapshots_user_source
    ON assignment_snapshots (user_id, source_url);

CREATE TABLE rubric_criteria (
    id BIGSERIAL PRIMARY KEY,
    assignment_snapshot_id BIGINT NOT NULL REFERENCES assignment_snapshots(id) ON DELETE CASCADE,
    name TEXT,
    description TEXT,
    points_possible NUMERIC(10, 2),
    ratings_json JSONB,
    raw_text TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rubric_criteria_assignment
    ON rubric_criteria (assignment_snapshot_id, sort_order);

CREATE TABLE submission_drafts (
    id BIGSERIAL PRIMARY KEY,
    assignment_snapshot_id BIGINT NOT NULL REFERENCES assignment_snapshots(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    body TEXT NOT NULL,
    word_count INT NOT NULL DEFAULT 0,
    source_type TEXT NOT NULL DEFAULT 'manual_paste',
    draft_mode TEXT NOT NULL DEFAULT 'text',
    submission_url TEXT,
    captured_from_canvas BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_submission_drafts_assignment
    ON submission_drafts (assignment_snapshot_id, updated_at DESC);

CREATE TABLE submission_draft_files (
    id BIGSERIAL PRIMARY KEY,
    submission_draft_id BIGINT NOT NULL REFERENCES submission_drafts(id) ON DELETE CASCADE,
    source_file_name TEXT NOT NULL,
    file_storage_key TEXT NOT NULL,
    file_mime_type TEXT,
    file_byte_size BIGINT,
    canvas_file_id TEXT,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_submission_draft_files_draft
    ON submission_draft_files (submission_draft_id, sort_order);

CREATE INDEX idx_submission_draft_files_canvas_file_id
    ON submission_draft_files (canvas_file_id)
    WHERE canvas_file_id IS NOT NULL;

CREATE TABLE analysis_runs (
    id BIGSERIAL PRIMARY KEY,
    assignment_snapshot_id BIGINT NOT NULL REFERENCES assignment_snapshots(id) ON DELETE CASCADE,
    submission_draft_id BIGINT REFERENCES submission_drafts(id) ON DELETE SET NULL,
    provider TEXT,
    model TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    overall_summary TEXT,
    predicted_score NUMERIC(10, 2),
    predicted_score_max NUMERIC(10, 2),
    confidence TEXT,
    raw_model_input JSONB,
    raw_model_output JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE TABLE analysis_attempts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assignment_snapshot_id BIGINT NOT NULL REFERENCES assignment_snapshots(id) ON DELETE CASCADE,
    analysis_run_id BIGINT REFERENCES analysis_runs(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'started',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_analysis_attempts_user_created
    ON analysis_attempts (user_id, created_at DESC);

CREATE INDEX idx_analysis_attempts_assignment_created
    ON analysis_attempts (assignment_snapshot_id, created_at DESC);

CREATE TABLE feedback_items (
    id BIGSERIAL PRIMARY KEY,
    analysis_run_id BIGINT NOT NULL REFERENCES analysis_runs(id) ON DELETE CASCADE,
    rubric_criterion_id BIGINT REFERENCES rubric_criteria(id) ON DELETE SET NULL,
    category TEXT NOT NULL,
    severity TEXT NOT NULL DEFAULT 'info',
    title TEXT NOT NULL,
    explanation TEXT,
    score_rationale TEXT,
    fulfilled_requirements JSONB NOT NULL DEFAULT '[]'::jsonb,
    unfulfilled_requirements JSONB NOT NULL DEFAULT '[]'::jsonb,
    criterion_status TEXT,
    criterion_score NUMERIC,
    predicted_points NUMERIC,
    max_points NUMERIC,
    selected_rating TEXT,
    status TEXT NOT NULL DEFAULT 'open',
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE extracted_sources (
    id BIGSERIAL PRIMARY KEY,
    assignment_snapshot_id BIGINT NOT NULL REFERENCES assignment_snapshots(id) ON DELETE CASCADE,
    source_kind TEXT NOT NULL,
    raw_content TEXT,
    normalized_content TEXT,
    extraction_method TEXT,
    confidence TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
