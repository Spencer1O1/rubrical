package analysispipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"rubrical/internal/aisettings"
	"rubrical/internal/analysispipeline/analysis"
	analysisschema "rubrical/internal/analysispipeline/analysis/schema"
	"rubrical/internal/analysispipeline/checkability"
	"rubrical/internal/analysispipeline/files"
	"rubrical/internal/draftfiles"
	"rubrical/internal/draftmode"
	"rubrical/internal/llm"
	"rubrical/internal/submissiontypes"
	"rubrical/internal/urlfetch"
)

var (
	ErrNotConfigured       = errors.New("ai analysis is not configured")
	ErrNothingToAnalyze    = errors.New("add draft text, a file, or a submission url before analyzing")
	ErrNoCheckableContent = files.ErrNoCheckableContent
	ErrURLFetchFailed      = errors.New("could not fetch submission url content")
)

type Service struct {
	pool          *pgxpool.Pool
	files         *draftfiles.Store
	settings      *aisettings.Store
	limiter       *Limiter
	urlFetch      *urlfetch.SafeFetcher
	enforceLimits bool
	opts          Options
}

func NewService(
	pool *pgxpool.Pool,
	files *draftfiles.Store,
	settings *aisettings.Store,
	limiter *Limiter,
	enforceLimits bool,
	allowLocalURLFetch bool,
	opts Options,
) *Service {
	return &Service{
		pool:          pool,
		files:         files,
		settings:      settings,
		limiter:       limiter,
		urlFetch:      urlfetch.NewSafeFetcher(allowLocalURLFetch),
		enforceLimits: enforceLimits,
		opts:          opts.withDefaults(),
	}
}

func (s *Service) resolveProvider(ctx context.Context, userID int64) (llm.Provider, error) {
	if s == nil || s.settings == nil {
		return nil, ErrNotConfigured
	}
	stored, err := s.settings.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !aisettings.IsConfigured(stored) {
		return nil, ErrNotConfigured
	}
	apiKey, err := s.settings.ActiveAPIKey(stored)
	if err != nil {
		return nil, ErrNotConfigured
	}
	return llm.New(stored.Provider, stored.Model, apiKey)
}

func (s *Service) Run(ctx context.Context, assignmentID, userID int64) (Result, error) {
	provider, err := s.resolveProvider(ctx, userID)
	if err != nil {
		return Result{}, err
	}

	input, draftID, err := s.loadInput(ctx, assignmentID, userID)
	if err != nil {
		return Result{}, err
	}
	if err := validateCheckable(input); err != nil {
		return Result{}, err
	}

	fileResult, err := s.processFiles(provider.Name(), input.Files)
	if err != nil {
		return Result{}, err
	}
	if err := validateProcessedContent(input, fileResult); err != nil {
		return Result{}, err
	}

	refs := input.Rubric.AssignCriterionIDs()
	pass1Req := checkability.BuildRequest(checkability.Input{
		PageType:        input.PageType,
		Instructions:    input.Instructions,
		AllowedChannels: input.AllowedChannels,
		Criteria:        refs,
	}, provider.Name())
	if err := ValidateLLMRequest(pass1Req); err != nil {
		return Result{}, err
	}

	inputLog, err := EncodePipelinePromptLog(pass1Req, nil)
	if err != nil {
		return Result{}, err
	}

	runHandle, err := s.beginRun(ctx, userID, assignmentID, draftID, provider.Name(), provider.Model(), inputLog)
	if err != nil {
		return Result{}, err
	}

	pass1Raw, err := provider.CompleteJSON(ctx, pass1Req)
	if err != nil {
		_ = s.markRunFailed(ctx, runHandle, err)
		return Result{}, err
	}
	class, err := checkability.ParseResponse(pass1Raw, refs)
	if err != nil {
		_ = s.markRunFailed(ctx, runHandle, err)
		return Result{}, fmt.Errorf("checkability: %w", err)
	}

	checkableRubric := filterRubric(input.Rubric, class)
	var scored *analysisschema.ScoredAnalysis

	if len(checkableRubric.Rows) > 0 {
		pass2Req := analysis.BuildAnalysisRequest(toDraftInput(input), fileResult, s.opts.MaxSubmissionTextChars, provider.Name(), checkableRubric)
		if err := ValidateLLMRequest(pass2Req); err != nil {
			_ = s.markRunFailed(ctx, runHandle, err)
			return Result{}, err
		}
		combinedLog, err := EncodePipelinePromptLog(pass1Req, &pass2Req)
		if err != nil {
			_ = s.markRunFailed(ctx, runHandle, err)
			return Result{}, err
		}
		if err := s.updateRunPromptLog(ctx, runHandle.RunID, combinedLog); err != nil {
			_ = s.markRunFailed(ctx, runHandle, err)
			return Result{}, err
		}

		pass2Raw, err := provider.CompleteJSON(ctx, pass2Req)
		if err != nil {
			_ = s.markRunFailed(ctx, runHandle, err)
			return Result{}, err
		}
		var providerResp analysisschema.ProviderResponse
		if err := json.Unmarshal(pass2Raw, &providerResp); err != nil {
			_ = s.markRunFailed(ctx, runHandle, err)
			return Result{}, err
		}
		if err := analysisschema.ValidateProviderResponse(&providerResp); err != nil {
			_ = s.markRunFailed(ctx, runHandle, err)
			return Result{}, err
		}

		scored, err = analysis.ApplyRubricScoring(&providerResp, checkableRubric)
		if err != nil {
			_ = s.markRunFailed(ctx, runHandle, err)
			return Result{}, fmt.Errorf("rubric scoring: %w", err)
		}
	}

	merged, err := MergeAnalysis(class, scored, input.Rubric)
	if err != nil {
		_ = s.markRunFailed(ctx, runHandle, err)
		return Result{}, fmt.Errorf("merge analysis: %w", err)
	}
	if err := analysisschema.ValidateScoredAnalysis(merged); err != nil {
		_ = s.markRunFailed(ctx, runHandle, err)
		return Result{}, fmt.Errorf("analysis output: %w", err)
	}

	if err := s.saveScoredAnalysis(ctx, runHandle.RunID, merged); err != nil {
		_ = s.markRunFailed(ctx, runHandle, err)
		return Result{}, err
	}

	result, err := s.persistSuccess(ctx, runHandle, assignmentID, provider, merged)
	if err != nil {
		return result, err
	}
	return result, nil
}

func toDraftInput(input Input) analysis.DraftInput {
	return analysis.DraftInput{
		PageType:       input.PageType,
		Title:          input.Title,
		CourseName:     input.CourseName,
		Instructions:   input.Instructions,
		PointsPossible: input.PointsPossible,
		DraftMode:      input.DraftMode,
		DraftText:      input.DraftText,
		DraftURL:       input.DraftURL,
		Rubric:         input.Rubric,
	}
}

func (s *Service) processFiles(providerName string, submissionFiles []SubmissionFile) (files.ProcessResult, error) {
	inputs := make([]files.SubmissionInput, len(submissionFiles))
	for i, file := range submissionFiles {
		inputs[i] = files.SubmissionInput{
			FileName: file.FileName,
			MimeType: file.MimeType,
			Data:     file.Data,
		}
	}
	return files.Process(providerName, inputs, s.opts.FileLimits())
}

func validateProcessedContent(input Input, result files.ProcessResult) error {
	if result.HasContent() {
		return nil
	}
	if strings.TrimSpace(input.DraftText) != "" {
		return nil
	}
	if draftmode.Normalize(input.DraftMode) == draftmode.URL && strings.TrimSpace(input.DraftURL) != "" {
		return nil
	}
	return files.ErrNoCheckableContent
}

func (s *Service) LoadLatestResult(ctx context.Context, assignmentID int64) (*Result, error) {
	return loadLatestResult(ctx, s.pool, assignmentID)
}

func (s *Service) LoadRubricContext(ctx context.Context, assignmentID int64) (analysis.RubricContext, error) {
	return loadRubricContext(ctx, s.pool, assignmentID)
}

func validateCheckable(input Input) error {
	switch draftmode.Normalize(input.DraftMode) {
	case draftmode.URL:
		if strings.TrimSpace(input.DraftURL) != "" || strings.TrimSpace(input.DraftText) != "" {
			return nil
		}
	case draftmode.File:
		if len(input.Files) > 0 || strings.TrimSpace(input.DraftText) != "" {
			return nil
		}
	default:
		if strings.TrimSpace(input.DraftText) != "" {
			return nil
		}
	}
	return ErrNothingToAnalyze
}

func (s *Service) loadInput(ctx context.Context, assignmentID, userID int64) (Input, int64, error) {
	var input Input
	var draftID int64
	var pointsPossible *float64
	var allowedJSON []byte

	err := s.pool.QueryRow(ctx, `
		SELECT
			id,
			COALESCE(page_type, ''),
			COALESCE(assignment_title, ''),
			COALESCE(course_name, ''),
			COALESCE(instructions_text, ''),
			points_possible,
			allowed_submission_types
		FROM assignment_snapshots
		WHERE id = $1 AND user_id = $2
	`, assignmentID, userID).Scan(
		&input.AssignmentID,
		&input.PageType,
		&input.Title,
		&input.CourseName,
		&input.Instructions,
		&pointsPossible,
		&allowedJSON,
	)
	if err != nil {
		return Input{}, 0, fmt.Errorf("assignment not found")
	}
	input.UserID = userID
	input.PointsPossible = pointsPossible
	input.AllowedChannels = submissiontypes.AllowedDraftModes(decodeAllowedSubmissionTypes(allowedJSON))

	input.Rubric, err = loadRubricContext(ctx, s.pool, assignmentID)
	if err != nil {
		return Input{}, 0, err
	}

	var draftURL *string
	err = s.pool.QueryRow(ctx, `
		SELECT id, COALESCE(draft_mode, 'text'), COALESCE(body, ''), submission_url
		FROM submission_drafts
		WHERE assignment_snapshot_id = $1 AND user_id = $2
		ORDER BY updated_at DESC, id DESC
		LIMIT 1
	`, assignmentID, userID).Scan(&draftID, &input.DraftMode, &input.DraftText, &draftURL)
	if errors.Is(err, pgx.ErrNoRows) {
		input.DraftMode = draftmode.Text
	} else if err != nil {
		return Input{}, 0, err
	} else if draftURL != nil {
		input.DraftURL = strings.TrimSpace(*draftURL)
	}

	if draftID > 0 && draftmode.Normalize(input.DraftMode) == draftmode.File {
		draftFiles, err := s.loadDraftFiles(ctx, draftID)
		if err != nil {
			return Input{}, 0, err
		}
		input.Files = draftFiles
	}

	if draftmode.Normalize(input.DraftMode) == draftmode.URL && input.DraftURL != "" {
		fetched, err := s.urlFetch.Fetch(ctx, input.DraftURL)
		if errors.Is(err, urlfetch.ErrNonHTMLContent) {
			return Input{}, 0, err
		}
		if err != nil || strings.TrimSpace(fetched) == "" {
			return Input{}, 0, ErrURLFetchFailed
		}
		input.DraftText = fetched
	}

	return input, draftID, nil
}

func (s *Service) loadDraftFiles(ctx context.Context, draftID int64) ([]SubmissionFile, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT source_file_name, COALESCE(file_mime_type, ''), file_storage_key
		FROM submission_draft_files
		WHERE submission_draft_id = $1
		ORDER BY sort_order ASC, id ASC
	`, draftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []SubmissionFile
	for rows.Next() {
		var name, mime, storageKey string
		if err := rows.Scan(&name, &mime, &storageKey); err != nil {
			return nil, err
		}
		data, err := s.files.Read(storageKey)
		if err != nil {
			if errors.Is(err, draftfiles.ErrNotFound) {
				return nil, fmt.Errorf("draft file %q is missing from storage — re-upload it", name)
			}
			return nil, fmt.Errorf("read draft file %q: %w", name, err)
		}
		out = append(out, SubmissionFile{
			FileName: name,
			MimeType: mime,
			Data:     data,
		})
	}
	return out, rows.Err()
}

func decodeAllowedSubmissionTypes(raw []byte) []string {
	if len(raw) == 0 {
		return nil
	}
	var types []string
	if err := json.Unmarshal(raw, &types); err != nil {
		return nil
	}
	return types
}
