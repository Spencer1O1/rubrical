package main

import (
	"context"
	"log"

	"rubrical/internal/config"
	"rubrical/internal/db"
	"rubrical/internal/draftfiles"
	"rubrical/internal/purge"
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

	fileStore, err := draftfiles.NewStore(cfg.DataDir)
	if err != nil {
		log.Fatalf("draft files: %v", err)
	}

	policy := purge.Policy{
		PostDueDateRetention: cfg.PostDueDateRetention,
		PostUploadRetention:  cfg.PostUploadRetention,
	}
	n, err := purge.PurgeDraftFiles(ctx, database.Pool, fileStore, policy)
	if err != nil {
		log.Fatalf("purge: %v", err)
	}

	log.Printf("purge complete: removed %d draft file(s)", n)
}
