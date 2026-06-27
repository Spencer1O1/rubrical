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
			id,
			COALESCE(assignment_title, 'Untitled assignment'),
			COALESCE(course_name, ''),
			imported_at,
			COALESCE(submission_type, '')
		FROM assignment_snapshots
		WHERE user_id = $1
		ORDER BY imported_at DESC
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
		); err != nil {
			return nil, err
		}
		item.ImportedAtLabel = pages.ImportedAtLabel(importedAt)
		item.URL = pages.AssignmentURL(item.ID)
		item.Status = "imported"
		items = append(items, item)
	}

	return items, rows.Err()
}
