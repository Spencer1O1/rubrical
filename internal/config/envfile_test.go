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
		[]byte("RUBRICAL_STRICT_EXTRACTION=1\nPOSTGRES_HOST=127.0.0.1\nPOSTGRES_PORT=5432\nPOSTGRES_USER=rubrical\nPOSTGRES_PASSWORD=rubrical\nPOSTGRES_DB=rubrical\nPOSTGRES_SSLMODE=disable\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	_ = os.Unsetenv("RUBRICAL_STRICT_EXTRACTION")
	for _, key := range []string{
		"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_USER",
		"POSTGRES_PASSWORD", "POSTGRES_DB", "POSTGRES_SSLMODE",
	} {
		_ = os.Unsetenv(key)
	}

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
