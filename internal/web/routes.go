package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"rubrical/internal/aisettings"
	"rubrical/internal/config"
	"rubrical/internal/db"
	"rubrical/internal/draftfiles"
	"rubrical/internal/analysis"
	"rubrical/internal/web/handlers"
)

func NewRouter(database *db.DB, fileStore *draftfiles.Store, userID int64, cfg config.Config, analysisSvc *analysis.Service) http.Handler {
	r := chi.NewRouter()
	r.Use(cors)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	h := handlers.New(database, fileStore, userID, cfg, analysisSvc, aisettings.NewStore(database.Pool))

	r.Get("/health", h.Health)
	r.Get("/", h.Dashboard)
	r.Get("/settings", h.SettingsPage)
	r.Get("/settings/ai", h.GetAISettingsAPI)
	r.Post("/settings/ai", h.SaveAISettings)
	r.Get("/assignments/draft-manifest", h.DraftManifest)

	r.Route("/assignments", func(r chi.Router) {
		r.Get("/{id}", h.AssignmentDetail)
		r.Post("/{id}/draft", h.SaveDraft)
		r.Post("/{id}/draft/upload", h.UploadDraft)
		r.Post("/{id}/draft/discussion-attachment", h.UploadDiscussionAttachment)
		r.Post("/{id}/draft/url", h.SaveDraftURL)
		r.Post("/{id}/draft/mode", h.SetDraftMode)
		r.Post("/{id}/draft/files/{fileId}/remove", h.RemoveDraftFile)
		r.Post("/{id}/analyze", h.AnalyzeDraft)
		r.Get("/{id}/results", h.AnalysisResults)
	})

	r.Post("/imports", h.ImportAssignment)

	fileServer := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	return r
}
