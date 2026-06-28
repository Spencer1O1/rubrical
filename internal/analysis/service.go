package analysis

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"rubrical/internal/aisettings"
	"rubrical/internal/analysis/filecontent"
	"rubrical/internal/analysis/provider"
	"rubrical/internal/draftfiles"
	"rubrical/internal/draftmode"
)

var (
	ErrNotConfigured    = errors.New("ai analysis is not configured")
	ErrNothingToAnalyze = errors.New("add draft text, a file, or a submission url before analyzing")
)

type Service struct {
	pool          *pgxpool.Pool
	files         *draftfiles.Store
	settings      *aisettings.Store
	limiter       *Limiter
	urlFetch      *URLFetcher
	enforceLimits bool
	opts          Options
}

func NewService(
	pool *pgxpool.Pool,
	files *draftfiles.Store,
	settings *aisettings.Store,
	limiter *Limiter,
	enforceLimits bool,
	opts Options,
) *Service {
	return &Service{
		pool:          pool,
		files:         files,
		settings:      settings,
		limiter:       limiter,
		urlFetch:      NewURLFetcher(),
		enforceLimits: enforceLimits,
		opts:          opts.withDefaults(),
	}
}

func (s *Service) resolveProvider(ctx context.Context, userID int64) (provider.Provider, error) {
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
	return provider.NewFromUser(provider.UserCredentials{
		Provider: stored.Provider,
		Model:    stored.Model,
		APIKey:   apiKey,
	})
}

func (s *Service) Run(ctx context.Context, assignmentID, userID int64) (Result, error) {
	provider, err := s.resolveProvider(ctx, userID)
	if err != nil {
		return Result{}, err
	}
	if s.enforceLimits {
		if err := s.limiter.Check(ctx, userID, assignmentID); err != nil {
			return Result{}, err
		}
	}

	input, draftID, err := s.loadInput(ctx, assignmentID, userID)
	if err != nil {
		return Result{}, err
	}
	if err := validateAnalyzable(input); err != nil {
		return Result{}, err
	}

	req := BuildProviderRequest(input, s.opts.PromptMaxDraftChars)
	inputLog, err := EncodePromptLog(req)
	if err != nil {
		return Result{}, err
	}

	runID, err := s.createRun(ctx, assignmentID, draftID, provider.Name(), provider.Model(), inputLog)
	if err != nil {
		return Result{}, err
	}

	if err := s.markRunStatus(ctx, runID, "running", nil, nil); err != nil {
		return Result{}, err
	}

	modelOut, err := provider.Analyze(ctx, req)
	if err != nil {
		_ = s.markRunFailed(ctx, runID, err)
		return Result{}, err
	}

	result, err := s.persistSuccess(ctx, runID, assignmentID, provider, modelOut)
	if err != nil {
		_ = s.markRunFailed(ctx, runID, err)
		return Result{}, err
	}
	return result, nil
}

func (s *Service) LoadLatestResult(ctx context.Context, assignmentID int64) (*Result, error) {
	return loadLatestResult(ctx, s.pool, assignmentID)
}

func validateAnalyzable(input Input) error {
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
		if strings.TrimSpace(input.DraftText) != "" || len(input.Files) > 0 {
			return nil
		}
	}
	return ErrNothingToAnalyze
}

func (s *Service) loadInput(ctx context.Context, assignmentID, userID int64) (Input, int64, error) {
	var input Input
	var draftID int64
	var pointsPossible *float64

	err := s.pool.QueryRow(ctx, `
		SELECT
			id,
			COALESCE(page_type, ''),
			COALESCE(assignment_title, ''),
			COALESCE(course_name, ''),
			COALESCE(instructions_text, ''),
			points_possible
		FROM assignment_snapshots
		WHERE id = $1 AND user_id = $2
	`, assignmentID, userID).Scan(
		&input.AssignmentID,
		&input.PageType,
		&input.Title,
		&input.CourseName,
		&input.Instructions,
		&pointsPossible,
	)
	if err != nil {
		return Input{}, 0, fmt.Errorf("assignment not found")
	}
	input.UserID = userID
	input.PointsPossible = pointsPossible

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

	if draftID > 0 {
		files, err := s.loadDraftFiles(ctx, draftID)
		if err != nil {
			return Input{}, 0, err
		}
		input.Files = files
	}

	if draftmode.Normalize(input.DraftMode) == draftmode.URL && input.DraftURL != "" {
		fetched, err := s.urlFetch.Fetch(ctx, input.DraftURL)
		if err == nil && strings.TrimSpace(fetched) != "" {
			input.DraftText = fetched
		}
	}

	mergedText, attachments, err := s.mergeFileContent(input)
	if err != nil {
		return Input{}, 0, err
	}
	if mergedText != "" {
		if strings.TrimSpace(input.DraftText) != "" {
			input.DraftText = strings.TrimSpace(input.DraftText) + "\n\n" + mergedText
		} else {
			input.DraftText = mergedText
		}
	}
	input.Files = attachments

	return input, draftID, nil
}

func (s *Service) mergeFileContent(input Input) (extractedText string, attachments []SubmissionFile, err error) {
	var textParts []string
	for _, file := range input.Files {
		prepared, prepErr := filecontent.Prepare(file.FileName, file.MimeType, file.Data, s.opts.MaxFileBytes)
		if prepErr != nil {
			return "", nil, prepErr
		}
		switch prepared.Kind {
		case "text":
			if strings.TrimSpace(prepared.Text) != "" {
				textParts = append(textParts, fmt.Sprintf("--- %s ---\n%s", prepared.FileName, prepared.Text))
			}
		case "pdf", "image":
			attachments = append(attachments, SubmissionFile{
				FileName: prepared.FileName,
				MimeType: prepared.MimeType,
				Data:     prepared.Data,
			})
		}
	}
	return strings.Join(textParts, "\n\n"), attachments, nil
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

	var files []SubmissionFile
	for rows.Next() {
		var name, mime, storageKey string
		if err := rows.Scan(&name, &mime, &storageKey); err != nil {
			return nil, err
		}
		data, err := s.files.Read(storageKey)
		if err != nil {
			return nil, fmt.Errorf("read draft file %q: %w", name, err)
		}
		files = append(files, SubmissionFile{
			FileName: name,
			MimeType: mime,
			Data:     data,
		})
	}
	return files, rows.Err()
}

type URLFetcher struct {
	client *http.Client
}

func NewURLFetcher() *URLFetcher {
	return &URLFetcher{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (f *URLFetcher) Fetch(ctx context.Context, rawURL string) (string, error) {
	if f == nil || f.client == nil {
		return "", fmt.Errorf("url fetcher unavailable")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("url fetch status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 512<<10))
	if err != nil {
		return "", err
	}
	return stripHTML(string(body)), nil
}

func stripHTML(html string) string {
	var b strings.Builder
	inTag := false
	for _, r := range html {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
