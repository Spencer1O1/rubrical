export type AIProviderId = "openai" | "anthropic";

export type AISettings = {
  provider: AIProviderId;
  model: string;
  openaiApiKey: string;
  anthropicApiKey: string;
  openaiApiKeyConfigured?: boolean;
  anthropicApiKeyConfigured?: boolean;
};

export const PROVIDER_MODELS: Record<AIProviderId, readonly { id: string; label: string }[]> = {
  openai: [
    { id: "gpt-4o-mini", label: "GPT-4o mini (recommended)" },
    { id: "gpt-4o", label: "GPT-4o" },
    { id: "gpt-4.1-mini", label: "GPT-4.1 mini" },
    { id: "gpt-4.1", label: "GPT-4.1" },
  ],
  anthropic: [
    { id: "claude-sonnet-4-20250514", label: "Claude Sonnet 4" },
    { id: "claude-3-7-sonnet-20250219", label: "Claude 3.7 Sonnet" },
    { id: "claude-3-5-haiku-20241022", label: "Claude 3.5 Haiku" },
  ],
};

export const DEFAULT_AI_SETTINGS: AISettings = {
  provider: "openai",
  model: PROVIDER_MODELS.openai[0].id,
  openaiApiKey: "",
  anthropicApiKey: "",
  openaiApiKeyConfigured: false,
  anthropicApiKeyConfigured: false,
};

export function defaultModelForProvider(provider: AIProviderId): string {
  return PROVIDER_MODELS[provider][0]?.id ?? DEFAULT_AI_SETTINGS.model;
}

export function normalizeAISettings(raw: Partial<AISettings> | null | undefined): AISettings {
  const provider: AIProviderId =
    raw?.provider === "anthropic" ? "anthropic" : "openai";
  const models = PROVIDER_MODELS[provider];
  const modelIds = new Set(models.map((entry) => entry.id));
  const model =
    typeof raw?.model === "string" && modelIds.has(raw.model)
      ? raw.model
      : defaultModelForProvider(provider);

  return {
    provider,
    model,
    openaiApiKey: typeof raw?.openaiApiKey === "string" ? raw.openaiApiKey.trim() : "",
    anthropicApiKey:
      typeof raw?.anthropicApiKey === "string" ? raw.anthropicApiKey.trim() : "",
    openaiApiKeyConfigured: raw?.openaiApiKeyConfigured === true,
    anthropicApiKeyConfigured: raw?.anthropicApiKeyConfigured === true,
  };
}

export function apiKeyForSettings(settings: AISettings): string {
  return settings.provider === "anthropic"
    ? settings.anthropicApiKey
    : settings.openaiApiKey;
}

export function isAISettingsConfigured(settings: AISettings): boolean {
  if (settings.provider === "anthropic") {
    return (
      settings.anthropicApiKey.length > 0 || settings.anthropicApiKeyConfigured === true
    );
  }
  return settings.openaiApiKey.length > 0 || settings.openaiApiKeyConfigured === true;
}
