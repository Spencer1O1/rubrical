-- sqlc schema mirror of migrations for code generation

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    display_name TEXT,
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
    imported_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

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

CREATE TABLE submission_drafts (
    id BIGSERIAL PRIMARY KEY,
    assignment_snapshot_id BIGINT NOT NULL REFERENCES assignment_snapshots(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    body TEXT NOT NULL,
    word_count INT NOT NULL DEFAULT 0,
    source_type TEXT NOT NULL DEFAULT 'manual_paste',
    captured_from_canvas BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE analysis_runs (
    id BIGSERIAL PRIMARY KEY,
    assignment_snapshot_id BIGINT NOT NULL REFERENCES assignment_snapshots(id) ON DELETE CASCADE,
    submission_draft_id BIGINT REFERENCES submission_drafts(id) ON DELETE SET NULL,
    provider TEXT,
    model TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    overall_summary TEXT,
    estimated_score NUMERIC(10, 2),
    estimated_score_max NUMERIC(10, 2),
    confidence TEXT,
    raw_model_input JSONB,
    raw_model_output JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE TABLE feedback_items (
    id BIGSERIAL PRIMARY KEY,
    analysis_run_id BIGINT NOT NULL REFERENCES analysis_runs(id) ON DELETE CASCADE,
    rubric_criterion_id BIGINT REFERENCES rubric_criteria(id) ON DELETE SET NULL,
    category TEXT NOT NULL,
    severity TEXT NOT NULL DEFAULT 'info',
    title TEXT NOT NULL,
    explanation TEXT,
    evidence TEXT,
    suggestion TEXT,
    status TEXT NOT NULL DEFAULT 'open',
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
