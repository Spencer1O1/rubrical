package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Addr              string
	DatabaseURL       string
	AIProvider        string
	OpenAIKey         string
	AnthropicKey      string
	StrictExtraction  bool
}

func Load() (Config, error) {
	loadEnvFiles()

	cfg := Config{
		Addr:             envOrDefault("RUBRICAL_ADDR", ":8787"),
		DatabaseURL:      envOrDefault("DATABASE_URL", "postgres://rubrical:rubrical@localhost:5432/rubrical?sslmode=disable"),
		AIProvider:       envOrDefault("AI_PROVIDER", ""),
		OpenAIKey:        os.Getenv("OPENAI_API_KEY"),
		AnthropicKey:     os.Getenv("ANTHROPIC_API_KEY"),
		StrictExtraction: envBool("RUBRICAL_STRICT_EXTRACTION"),
	}

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
