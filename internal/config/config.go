package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const DefaultDraftMaxFileBytes = 32 << 20

type Config struct {
	Addr                    string
	DatabaseURL             string
	DataDir                 string
	DraftMaxFileBytes       int
	DraftMaxFilesPerDraft   int
	AIPromptMaxDraftChars   int
	AIMaxRunsPerHour        int
	AIMaxRunsPerDay         int
	AIMinSecondsBetweenRuns int
	AIEnforceRateLimits     bool
	StrictExtraction        bool
	PostDueDateRetention    time.Duration
	PostUploadRetention     time.Duration
}

func Load() (Config, error) {
	loadEnvFiles()

	cfg := Config{
		Addr:                    envOrDefault("RUBRICAL_ADDR", ":8787"),
		DatabaseURL:             envOrDefault("DATABASE_URL", "postgres://rubrical:rubrical@localhost:5432/rubrical?sslmode=disable"),
		DataDir:                 envOrDefault("RUBRICAL_DATA_DIR", "./data"),
		DraftMaxFileBytes:       envInt("DRAFT_MAX_FILE_BYTES", DefaultDraftMaxFileBytes),
		DraftMaxFilesPerDraft:   envInt("DRAFT_MAX_FILES_PER_DRAFT", 20),
		AIPromptMaxDraftChars:   envInt("AI_PROMPT_MAX_DRAFT_CHARS", 120_000),
		AIMaxRunsPerHour:        envInt("AI_MAX_RUNS_PER_HOUR", 0),
		AIMaxRunsPerDay:         envInt("AI_MAX_RUNS_PER_DAY", 0),
		AIMinSecondsBetweenRuns: envInt("AI_MIN_SECONDS_BETWEEN_RUNS", 0),
		AIEnforceRateLimits:     envBool("AI_ENFORCE_RATE_LIMITS"),
		StrictExtraction:        envBool("RUBRICAL_STRICT_EXTRACTION"),
	}

	retention, err := envDuration("POST_DUE_DATE_RETENTION_TIME", 7*24*time.Hour)
	if err != nil {
		return Config{}, fmt.Errorf("POST_DUE_DATE_RETENTION_TIME: %w", err)
	}
	cfg.PostDueDateRetention = retention

	uploadRetention, err := envDuration("POST_UPLOAD_RETENTION_TIME", 30*24*time.Hour)
	if err != nil {
		return Config{}, fmt.Errorf("POST_UPLOAD_RETENTION_TIME: %w", err)
	}
	cfg.PostUploadRetention = uploadRetention

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func (c Config) Port() int {
	if len(c.Addr) > 0 && c.Addr[0] == ':' {
		port, err := strconv.Atoi(c.Addr[1:])
		if err == nil {
			return port
		}
	}
	return 8787
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return fallback
	}
	return n
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, err
	}
	if d < 0 {
		return 0, fmt.Errorf("must be >= 0")
	}
	return d, nil
}
