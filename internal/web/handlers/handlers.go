package handlers

import (
	"rubrical/internal/aisettings"
	"rubrical/internal/analysis"
	"rubrical/internal/config"
	"rubrical/internal/db"
	"rubrical/internal/draftfiles"
	"rubrical/internal/importpayload"
)

type Handlers struct {
	db               *db.DB
	files            *draftfiles.Store
	userID           int64
	strictExtraction bool
	analysis         *analysis.Service
	aiSettings       *aisettings.Store
	importLimits     importpayload.Limits
}

func New(
	database *db.DB,
	files *draftfiles.Store,
	userID int64,
	cfg config.Config,
	analysisSvc *analysis.Service,
	aiSettings *aisettings.Store,
) *Handlers {
	return &Handlers{
		db:               database,
		files:            files,
		userID:           userID,
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
