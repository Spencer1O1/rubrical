package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"rubrical/internal/config"
	"rubrical/internal/db"
	"rubrical/internal/web/handlers"
)

func NewRouter(database *db.DB, userID int64, cfg config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(cors)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	h := handlers.New(database, userID, cfg.StrictExtraction)

	r.Get("/health", h.Health)
	r.Get("/", h.Dashboard)

	r.Route("/assignments", func(r chi.Router) {
		r.Get("/{id}", h.AssignmentDetail)
		r.Post("/{id}/draft", h.SaveDraft)
		r.Post("/{id}/analyze", h.AnalyzeDraft)
		r.Get("/{id}/results", h.AnalysisResults)
	})

	r.Post("/imports", h.ImportAssignment)
	r.Post("/feedback/{id}/resolve", h.ResolveFeedback)

	fileServer := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	return r
}
