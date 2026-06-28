package handlers

import (
	"context"
	"net/http"
	"time"

	"rubrical/internal/web/pages"
)

func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	assignments, err := h.listAssignments(r.Context())
	if err != nil {
		http.Error(w, "failed to load assignments", http.StatusInternalServerError)
		return
	}

	pages.Dashboard(assignments).Render(r.Context(), w)
}

func (h *Handlers) listAssignments(ctx context.Context) ([]pages.AssignmentListItem, error) {
	rows, err := h.db.Pool.Query(ctx, `
		SELECT
			a.id,
			COALESCE(a.assignment_title, 'Untitled assignment'),
			COALESCE(a.course_name, ''),
			a.imported_at,
			COALESCE(a.submission_type, ''),
			CASE
				WHEN ar.status = 'completed' THEN 'analyzed'
				WHEN sd.id IS NOT NULL AND (
					COALESCE(sd.body, '') <> '' OR
					COALESCE(sd.submission_url, '') <> '' OR
					COALESCE(sd.word_count, 0) > 0 OR
					EXISTS (
						SELECT 1
						FROM submission_draft_files sdf
						WHERE sdf.submission_draft_id = sd.id
					)
				) THEN 'draft_added'
				ELSE 'imported'
			END AS status
		FROM assignment_snapshots a
		LEFT JOIN LATERAL (
			SELECT id, body, submission_url, word_count
			FROM submission_drafts
			WHERE assignment_snapshot_id = a.id AND user_id = $1
			ORDER BY updated_at DESC, id DESC
			LIMIT 1
		) sd ON true
		LEFT JOIN LATERAL (
			SELECT status
			FROM analysis_runs
			WHERE assignment_snapshot_id = a.id
			ORDER BY completed_at DESC NULLS LAST, id DESC
			LIMIT 1
		) ar ON true
		WHERE a.user_id = $1
		ORDER BY a.imported_at DESC
		LIMIT 50
	`, h.userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []pages.AssignmentListItem
	for rows.Next() {
		var item pages.AssignmentListItem
		var importedAt time.Time
		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.CourseName,
			&importedAt,
			&item.SubmissionType,
			&item.Status,
		); err != nil {
			return nil, err
		}
		item.ImportedAtLabel = pages.ImportedAtLabel(importedAt)
		item.URL = pages.AssignmentURL(item.ID)
		items = append(items, item)
	}

	return items, rows.Err()
}
