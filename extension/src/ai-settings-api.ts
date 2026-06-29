import { executeRubricalFetch } from "./api-fetch";
import type { AISettings } from "./ai-settings";
import {
  DEFAULT_AI_SETTINGS,
  normalizeAISettings,
  type AIProviderId,
} from "./ai-settings";
import { fetchSession, RubricalAuthRequiredError } from "./auth-api";

export { RubricalAuthRequiredError, fetchSession } from "./auth-api";

export async function fetchAISettingsFromServer(): Promise<AISettings> {
  const result = await executeRubricalFetch({
    path: "/settings/ai?format=json",
    method: "GET",
    headers: {
      Accept: "application/json",
    },
  });
  if (!result.ok) {
    if (result.authRequired) {
      throw new RubricalAuthRequiredError(result.base);
    }
    return DEFAULT_AI_SETTINGS;
  }
  return normalizeAISettings(result.data as Partial<AISettings>);
}

export async function saveAISettingsToServer(settings: AISettings): Promise<AISettings> {
  const payload = normalizeAISettings(settings);
  const result = await executeRubricalFetch({
    path: "/settings/ai",
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });
  if (!result.ok) {
    if (result.authRequired) {
      throw new RubricalAuthRequiredError(result.base);
    }
    throw new Error(result.error || "Failed to save AI settings");
  }
  return normalizeAISettings(result.data as Partial<AISettings>);
}

export { type AIProviderId };
