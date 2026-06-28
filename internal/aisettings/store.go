package aisettings

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"rubrical/internal/config"
)

type Settings struct {
	Provider        string `json:"provider"`
	Model           string `json:"model"`
	OpenAIAPIKey    string `json:"openaiApiKey"`
	AnthropicAPIKey string `json:"anthropicApiKey"`
}

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Get(ctx context.Context, userID int64) (Settings, error) {
	var settings Settings
	err := s.pool.QueryRow(ctx, `
		SELECT ai_provider, ai_model, openai_api_key, anthropic_api_key
		FROM user_ai_settings
		WHERE user_id = $1
	`, userID).Scan(
		&settings.Provider,
		&settings.Model,
		&settings.OpenAIAPIKey,
		&settings.AnthropicAPIKey,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return DefaultSettings(), nil
	}
	if err != nil {
		return Settings{}, err
	}
	return Normalize(settings), nil
}

func (s *Store) Save(ctx context.Context, userID int64, incoming Settings) (Settings, error) {
	current, err := s.Get(ctx, userID)
	if err != nil {
		return Settings{}, err
	}

	next := Merge(current, incoming)
	if err := validateSave(next); err != nil {
		return Settings{}, err
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO user_ai_settings (
			user_id, ai_provider, ai_model, openai_api_key, anthropic_api_key
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			ai_provider = EXCLUDED.ai_provider,
			ai_model = EXCLUDED.ai_model,
			openai_api_key = EXCLUDED.openai_api_key,
			anthropic_api_key = EXCLUDED.anthropic_api_key,
			updated_at = NOW()
	`, userID, next.Provider, next.Model, next.OpenAIAPIKey, next.AnthropicAPIKey)
	if err != nil {
		return Settings{}, err
	}
	return next, nil
}

func (s *Store) ActiveAPIKey(settings Settings) (string, error) {
	settings = Normalize(settings)
	switch settings.Provider {
	case "openai":
		if settings.OpenAIAPIKey == "" {
			return "", fmt.Errorf("openai api key is not configured")
		}
		return settings.OpenAIAPIKey, nil
	case "anthropic":
		if settings.AnthropicAPIKey == "" {
			return "", fmt.Errorf("anthropic api key is not configured")
		}
		return settings.AnthropicAPIKey, nil
	default:
		return "", fmt.Errorf("unsupported ai provider %q", settings.Provider)
	}
}

func DefaultSettings() Settings {
	return Settings{
		Provider: config.DefaultAIProvider,
		Model:    config.DefaultOpenAIModel,
	}
}

func Normalize(settings Settings) Settings {
	provider := strings.ToLower(strings.TrimSpace(settings.Provider))
	if provider != "anthropic" {
		provider = "openai"
	}

	model := strings.TrimSpace(settings.Model)
	if model == "" {
		if provider == "anthropic" {
			model = config.DefaultAnthropicModel
		} else {
			model = config.DefaultOpenAIModel
		}
	}

	return Settings{
		Provider:        provider,
		Model:           model,
		OpenAIAPIKey:    strings.TrimSpace(settings.OpenAIAPIKey),
		AnthropicAPIKey: strings.TrimSpace(settings.AnthropicAPIKey),
	}
}

func Merge(current, incoming Settings) Settings {
	next := Normalize(incoming)
	if next.OpenAIAPIKey == "" {
		next.OpenAIAPIKey = current.OpenAIAPIKey
	}
	if next.AnthropicAPIKey == "" {
		next.AnthropicAPIKey = current.AnthropicAPIKey
	}
	return next
}

func validateSave(settings Settings) error {
	settings = Normalize(settings)
	switch settings.Provider {
	case "openai":
		if settings.OpenAIAPIKey == "" {
			return fmt.Errorf("openai api key is required for the selected provider")
		}
	case "anthropic":
		if settings.AnthropicAPIKey == "" {
			return fmt.Errorf("anthropic api key is required for the selected provider")
		}
	}
	return nil
}

func IsConfigured(settings Settings) bool {
	settings = Normalize(settings)
	switch settings.Provider {
	case "anthropic":
		return settings.AnthropicAPIKey != ""
	default:
		return settings.OpenAIAPIKey != ""
	}
}
