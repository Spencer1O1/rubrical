package analysis

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"rubrical/internal/analysis/files"
)

// PreviewFiles runs the file pipeline without calling the model (for UI summaries).
func (s *Service) PreviewFiles(ctx context.Context, userID int64, assignmentID int64) (files.ProcessResult, error) {
	providerName := "openai"
	if s.settings != nil {
		stored, err := s.settings.Get(ctx, userID)
		if err != nil {
			return files.ProcessResult{}, err
		}
		providerName = strings.ToLower(strings.TrimSpace(stored.Provider))
		if providerName == "" {
			providerName = "openai"
		}
	}

	submissionFiles, err := s.loadAssignmentDraftFiles(ctx, assignmentID, userID)
	if err != nil {
		return files.ProcessResult{}, err
	}
	inputs := make([]files.SubmissionInput, len(submissionFiles))
	for i, file := range submissionFiles {
		inputs[i] = files.SubmissionInput{
			FileName: file.FileName,
			MimeType: file.MimeType,
			Data:     file.Data,
		}
	}
	return files.Process(processProviderName(providerName), inputs, s.opts.FileLimits())
}

func processProviderName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "anthropic" {
		return "anthropic"
	}
	return "openai"
}

func (s *Service) loadAssignmentDraftFiles(ctx context.Context, assignmentID, userID int64) ([]SubmissionFile, error) {
	var draftID int64
	err := s.pool.QueryRow(ctx, `
		SELECT id
		FROM submission_drafts
		WHERE assignment_snapshot_id = $1 AND user_id = $2
		ORDER BY updated_at DESC, id DESC
		LIMIT 1
	`, assignmentID, userID).Scan(&draftID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s.loadDraftFiles(ctx, draftID)
}
