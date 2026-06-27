package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(dir, ".env.local"),
		[]byte("RUBRICAL_STRICT_EXTRACTION=1\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	_ = os.Unsetenv("RUBRICAL_STRICT_EXTRACTION")

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	loadEnvFiles()

	if got := os.Getenv("RUBRICAL_STRICT_EXTRACTION"); got != "1" {
		t.Fatalf("expected RUBRICAL_STRICT_EXTRACTION=1, got %q", got)
	}
}

func TestLoadEnvFiles_doesNotOverrideExisting(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(dir, ".env.local"),
		[]byte("RUBRICAL_STRICT_EXTRACTION=1\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	t.Setenv("RUBRICAL_STRICT_EXTRACTION", "0")

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	loadEnvFiles()

	if got := os.Getenv("RUBRICAL_STRICT_EXTRACTION"); got != "0" {
		t.Fatalf("expected existing env to win, got %q", got)
	}
}

func TestLoad_readsStrictFromEnvFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(dir, ".env.local"),
		[]byte("RUBRICAL_STRICT_EXTRACTION=1\nDATABASE_URL=postgres://example\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	_ = os.Unsetenv("RUBRICAL_STRICT_EXTRACTION")
	_ = os.Unsetenv("DATABASE_URL")

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.StrictExtraction {
		t.Fatal("expected strict extraction from .env.local")
	}
}
