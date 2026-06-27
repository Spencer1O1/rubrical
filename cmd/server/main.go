package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rubrical/internal/auth"
	"rubrical/internal/config"
	"rubrical/internal/db"
	"rubrical/internal/web"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	database, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer database.Close()

	userID, err := auth.EnsureLocalUser(ctx, database.Pool)
	if err != nil {
		log.Fatalf("local user: %v", err)
	}
	log.Printf("using local dev user id=%d (%s)", userID, auth.LocalDevEmail)
	if cfg.StrictExtraction {
		log.Printf("RUBRICAL_STRICT_EXTRACTION=1 — extraction/display fallbacks disabled")
	}

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      web.NewRouter(database, userID, cfg),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("rubrical listening on http://localhost%s", cfg.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
}
