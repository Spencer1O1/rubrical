package config

import (
	"strings"
	"testing"
	"time"
)

func setTestPostgres(t *testing.T) {
	t.Helper()
	t.Setenv("POSTGRES_HOST", "127.0.0.1")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_USER", "rubrical")
	t.Setenv("POSTGRES_PASSWORD", "s3cret")
	t.Setenv("POSTGRES_DB", "rubrical")
	t.Setenv("POSTGRES_SSLMODE", "disable")
}

func TestLoadDatabaseURLFromPieces(t *testing.T) {
	setTestPostgres(t)

	got, err := loadDatabaseURL()
	if err != nil {
		t.Fatal(err)
	}
	want := "postgres://rubrical:s3cret@127.0.0.1:5432/rubrical?sslmode=disable"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestLoadDatabaseURLRequiresEveryPiece(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "")
	t.Setenv("POSTGRES_PORT", "")
	t.Setenv("POSTGRES_USER", "")
	t.Setenv("POSTGRES_PASSWORD", "")
	t.Setenv("POSTGRES_DB", "")
	t.Setenv("POSTGRES_SSLMODE", "")

	_, err := loadDatabaseURL()
	if err == nil {
		t.Fatal("expected error when postgres env is missing")
	}
	for _, key := range []string{
		"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_USER",
		"POSTGRES_PASSWORD", "POSTGRES_DB", "POSTGRES_SSLMODE",
	} {
		if !strings.Contains(err.Error(), key) {
			t.Fatalf("error %q missing %s", err, key)
		}
	}
}

func TestLoad_postDueDateRetentionDefault(t *testing.T) {
	setTestPostgres(t)
	t.Setenv("POST_DUE_DATE_RETENTION_TIME", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PostDueDateRetention != 7*24*time.Hour {
		t.Fatalf("default retention = %s", cfg.PostDueDateRetention)
	}
}

func TestLoad_postDueDateRetentionCustom(t *testing.T) {
	setTestPostgres(t)
	t.Setenv("POST_DUE_DATE_RETENTION_TIME", "48h")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PostDueDateRetention != 48*time.Hour {
		t.Fatalf("retention = %s", cfg.PostDueDateRetention)
	}
}

func TestLoad_postDueDateRetentionDisabled(t *testing.T) {
	setTestPostgres(t)
	t.Setenv("POST_DUE_DATE_RETENTION_TIME", "0")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PostDueDateRetention != 0 {
		t.Fatalf("retention = %s", cfg.PostDueDateRetention)
	}
}

func TestLoad_postUploadRetentionDefault(t *testing.T) {
	setTestPostgres(t)
	t.Setenv("POST_UPLOAD_RETENTION_TIME", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PostUploadRetention != 30*24*time.Hour {
		t.Fatalf("default upload retention = %s", cfg.PostUploadRetention)
	}
}

func TestLoad_postUploadRetentionCustom(t *testing.T) {
	setTestPostgres(t)
	t.Setenv("POST_UPLOAD_RETENTION_TIME", "336h")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PostUploadRetention != 336*time.Hour {
		t.Fatalf("upload retention = %s", cfg.PostUploadRetention)
	}
}

func TestLoad_postDueDateRetentionInvalid(t *testing.T) {
	setTestPostgres(t)
	t.Setenv("POST_DUE_DATE_RETENTION_TIME", "not-a-duration")

	if _, err := Load(); err == nil {
		t.Fatal("expected error for invalid duration")
	}
}
