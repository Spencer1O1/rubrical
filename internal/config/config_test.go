package config

import (
	"testing"
	"time"
)

func TestLoad_postDueDateRetentionDefault(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
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
	t.Setenv("DATABASE_URL", "postgres://example")
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
	t.Setenv("DATABASE_URL", "postgres://example")
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
	t.Setenv("DATABASE_URL", "postgres://example")
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
	t.Setenv("DATABASE_URL", "postgres://example")
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
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("POST_DUE_DATE_RETENTION_TIME", "not-a-duration")

	if _, err := Load(); err == nil {
		t.Fatal("expected error for invalid duration")
	}
}
