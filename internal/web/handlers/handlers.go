package handlers

import (
	"context"

	"rubrical/internal/aisettings"
	"rubrical/internal/analysis"
	"rubrical/internal/auth"
	"rubrical/internal/config"
	"rubrical/internal/db"
	"rubrical/internal/draftfiles"
	"rubrical/internal/email"
	"rubrical/internal/importpayload"
	"rubrical/internal/web/pages"
)

type Handlers struct {
	db               *db.DB
	files            *draftfiles.Store
	auth             *auth.Service
	google           auth.GoogleConfig
	mailer           email.Sender
	publicURL        string
	authSecure       bool
	strictExtraction bool
	analysis         *analysis.Service
	aiSettings       *aisettings.Store
	importLimits     importpayload.Limits
}

func New(
	database *db.DB,
	files *draftfiles.Store,
	authSvc *auth.Service,
	cfg config.Config,
	analysisSvc *analysis.Service,
	aiSettings *aisettings.Store,
	mailer email.Sender,
) *Handlers {
	return &Handlers{
		db:               database,
		files:            files,
		auth:             authSvc,
		google: auth.GoogleConfig{
			ClientID:     cfg.GoogleOAuthClientID,
			ClientSecret: cfg.GoogleOAuthClientSecret,
			RedirectURL:  cfg.GoogleOAuthRedirectURL(),
		},
		mailer:           mailer,
		publicURL:        cfg.PublicURL,
		authSecure:       cfg.CookieSecure(),
		strictExtraction: cfg.StrictExtraction,
		analysis:         analysisSvc,
		aiSettings:       aiSettings,
		importLimits:     importpayload.LimitsFromConfig(cfg.DraftMaxUploadBytes, cfg.DraftMaxUploadSlots),
	}
}

func (h *Handlers) maxDraftUploadBytes() int {
	if h == nil {
		return importpayload.DefaultLimits().MaxUploadBytes
	}
	return h.importLimits.MaxUploadBytes
}

func (h *Handlers) analysisResultsView(ctx context.Context, assignmentID int64, result *analysis.Result) pages.AnalysisResultsView {
	rubric := analysis.RubricContext{}
	if h.analysis != nil {
		if loaded, err := h.analysis.LoadRubricContext(ctx, assignmentID); err == nil {
			rubric = loaded
		}
	}
	return pages.AnalysisResultsFromResult(result, rubric)
}
