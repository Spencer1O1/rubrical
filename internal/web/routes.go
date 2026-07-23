package web

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"rubrical/internal/aisettings"
	"rubrical/internal/analysispipeline"
	"rubrical/internal/auth"
	"rubrical/internal/config"
	"rubrical/internal/db"
	"rubrical/internal/draftfiles"
	"rubrical/internal/email"
	"rubrical/internal/web/handlers"
)

func NewRouter(
	database *db.DB,
	fileStore *draftfiles.Store,
	cfg config.Config,
	analysisSvc *analysispipeline.Service,
	aiSettings *aisettings.Store,
	authSvc *auth.Service,
	mailer email.Sender,
) http.Handler {
	r := chi.NewRouter()
	authMW := auth.NewMiddleware(authSvc, cfg)

	r.Use(cors(cfg))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(authMW.OptionalUser)

	h := handlers.New(database, fileStore, authSvc, cfg, analysisSvc, aiSettings, mailer)

	r.Group(func(r chi.Router) {
		r.Get("/login", h.AuthPage)
		r.Post("/login", h.Login)
		r.Post("/signup", h.Signup)
		r.Post("/forgot-password", h.ForgotPassword)
		r.Get("/reset-password", h.ResetPasswordPage)
		r.Post("/reset-password", h.ResetPassword)
		r.Post("/auth/login", h.LoginAPI)
		r.Post("/auth/signup", h.SignupAPI)
		r.Post("/auth/forgot-password", h.ForgotPasswordAPI)
		r.Get("/auth/session", h.SessionAPI)
		r.Get("/auth/config", h.AuthConfigAPI)
		r.Get("/auth/google", h.GoogleAuthStart)
		r.Get("/auth/google/callback", h.GoogleAuthCallback)
		r.Get("/auth/embed", h.EmbedHandoff)
		r.Get("/install", h.Install)
		r.Get("/install/rubrical-extension.zip", h.ExtensionZip)
		// Legacy path — same no-store handler (Cloudflare was caching FileServer zips).
		r.Get("/static/downloads/rubrical-extension.zip", h.ExtensionZip)
		r.Get("/onboarding", h.Onboarding)
		r.Get("/", h.Home)
	})

	r.Group(func(r chi.Router) {
		r.Use(authMW.RequireUser)

		r.Get("/auth/embed-url", h.EmbedURL)
		r.Get("/settings", h.SettingsPage)
		r.Get("/settings/ai", h.GetAISettingsAPI)
		r.Post("/settings/ai", h.SaveAISettings)
		r.Get("/assignments/draft-manifest", h.DraftManifest)
		r.Post("/logout", h.Logout)
		r.Post("/imports", h.ImportAssignment)

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
	})

	fileServer := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	return r
}

func cors(cfg config.Config) func(http.Handler) http.Handler {
	allowed := cfg.AllowedOrigins()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin != "" && originAllowed(origin, allowed) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Vary", "Origin")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
			w.Header().Set("Access-Control-Allow-Private-Network", "true")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func originAllowed(origin string, allowed []string) bool {
	origin = strings.TrimRight(origin, "/")
	for _, candidate := range allowed {
		if origin == strings.TrimRight(candidate, "/") {
			return true
		}
	}
	return false
}
