-- name: ListAssignments :many
SELECT
    id,
    COALESCE(assignment_title, 'Untitled assignment') AS assignment_title,
    COALESCE(course_name, '') AS course_name,
    imported_at,
    due_at,
    COALESCE(submission_type, '') AS submission_type
FROM assignment_snapshots
ORDER BY imported_at DESC
LIMIT $1;

-- name: GetAssignment :one
SELECT
    id,
    COALESCE(assignment_title, 'Untitled assignment') AS assignment_title,
    COALESCE(course_name, '') AS course_name,
    COALESCE(instructions_text, '') AS instructions_text,
    COALESCE(raw_text, '') AS raw_text,
    imported_at,
    due_at
FROM assignment_snapshots
WHERE id = $1;

-- name: CreateAssignmentSnapshot :one
INSERT INTO assignment_snapshots (
    source_url,
    source_platform,
    page_type,
    course_name,
    assignment_title,
    raw_text,
    instructions_text,
    submission_type,
    imported_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, NOW()
)
RETURNING id;

-- name: ListRubricCriteria :many
SELECT
    id,
    COALESCE(name, '') AS name,
    COALESCE(description, '') AS description,
    COALESCE(raw_text, '') AS raw_text,
    COALESCE(points_possible, 0)::float8 AS points_possible,
    sort_order
FROM rubric_criteria
WHERE assignment_snapshot_id = $1
ORDER BY sort_order;

-- name: CreateRubricCriterion :exec
INSERT INTO rubric_criteria (
    assignment_snapshot_id,
    name,
    raw_text,
    sort_order
) VALUES ($1, $2, $3, $4);

-- name: CreateSubmissionDraft :exec
INSERT INTO submission_drafts (
    assignment_snapshot_id,
    body,
    word_count,
    source_type,
    captured_from_canvas
) VALUES ($1, $2, $3, $4, $5);

-- name: GetLatestDraft :one
SELECT
    id,
    body,
    word_count,
    source_type,
    captured_from_canvas,
    updated_at
FROM submission_drafts
WHERE assignment_snapshot_id = $1
ORDER BY updated_at DESC
LIMIT 1;
