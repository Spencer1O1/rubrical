package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rubrical/internal/aisettings"
	"rubrical/internal/analysis"
	"rubrical/internal/auth"
	"rubrical/internal/config"
	"rubrical/internal/db"
	"rubrical/internal/draftfiles"
	"rubrical/internal/email"
	"rubrical/internal/purge"
	"rubrical/internal/secrets"
	"rubrical/internal/web"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	bootstrap := context.Background()
	database, err := db.Connect(bootstrap, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer database.Close()

	authSvc := auth.NewService(database.Pool, cfg.SessionTTL)
	if err := authSvc.PurgeExpiredSessions(bootstrap); err != nil {
		log.Printf("session cleanup: %v", err)
	}

	if err := analysis.FailAllStaleRuns(bootstrap, database.Pool, analysis.DefaultStaleRunTTL); err != nil {
		log.Printf("stale analysis run cleanup: %v", err)
	}
	if cfg.StrictExtraction {
		log.Printf("RUBRICAL_STRICT_EXTRACTION=1 — extraction/display fallbacks disabled")
	}

	fileStore, err := draftfiles.NewStore(cfg.DataDir)
	if err != nil {
		log.Fatalf("draft files: %v", err)
	}

	secretsCipher, err := secrets.NewCipherFromEnv(cfg.SecretsEncryptionKey)
	if err != nil {
		log.Fatalf("secrets encryption: %v", err)
	}

	aiSettingsStore := aisettings.NewStore(database.Pool, secretsCipher)
	analysisSvc := analysis.NewService(
		database.Pool,
		fileStore,
		aiSettingsStore,
		analysis.NewLimiter(database.Pool, analysis.NewRateLimits(
			cfg.AIMaxRunsPerHour,
			cfg.AIMaxRunsPerDay,
			cfg.AIMinSecondsBetweenRuns,
		)),
		cfg.AIEnforceRateLimits,
		cfg.AllowLocalURLFetch,
		analysis.OptionsFromConfig(cfg),
	)
	if cfg.AIEnforceRateLimits {
		log.Printf("ai analysis rate limits enabled: %d/hr %d/day min_gap=%ds",
			cfg.AIMaxRunsPerHour,
			cfg.AIMaxRunsPerDay,
			cfg.AIMinSecondsBetweenRuns,
		)
	} else {
		log.Printf("ai analysis: per-user keys from database; rate limits disabled (BYOK beta)")
	}

	mailer := email.NewSender(email.Config{
		From:         cfg.EmailFrom,
		ResendAPIKey: cfg.ResendAPIKey,
		SMTPHost:     cfg.SMTPHost,
		SMTPPort:     cfg.SMTPPort,
		SMTPUsername: cfg.SMTPUsername,
		SMTPPassword: cfg.SMTPPassword,
		DevLog:       cfg.EmailDevLog,
	})

	purgeCtx, stopPurge := context.WithCancel(context.Background())
	defer stopPurge()
	policy := purge.Policy{
		PostDueDateRetention: cfg.PostDueDateRetention,
		PostUploadRetention:  cfg.PostUploadRetention,
	}
	log.Printf(
		"draft file purge: %s after due date (POST_DUE_DATE_RETENTION_TIME); %s after upload when no due date (POST_UPLOAD_RETENTION_TIME)",
		cfg.PostDueDateRetention,
		cfg.PostUploadRetention,
	)
	purge.RunBackground(purgeCtx, database.Pool, fileStore, policy, time.Hour)

	listenAddr := cfg.ListenAddr()
	server := &http.Server{
		Addr:         listenAddr,
		Handler:      web.NewRouter(database, fileStore, cfg, analysisSvc, aiSettingsStore, authSvc, mailer),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 180 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if cfg.Host == "" {
			log.Printf("rubrical listening on http://localhost:%d", cfg.Port)
		} else {
			log.Printf("rubrical listening on http://%s", listenAddr)
		}
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	stopPurge()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
}
