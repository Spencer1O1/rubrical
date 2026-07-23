import { describe, expect, it } from "vitest";
import {
  DEFAULT_AI_SETTINGS,
  apiKeyForSettings,
  defaultModelForProvider,
  isAISettingsConfigured,
  normalizeAISettings,
} from "./ai-settings";

describe("normalizeAISettings", () => {
  it("defaults to openai gpt-4o-mini", () => {
    expect(normalizeAISettings(undefined)).toEqual(DEFAULT_AI_SETTINGS);
  });

  it("keeps anthropic provider and valid model", () => {
    expect(
      normalizeAISettings({
        provider: "anthropic",
        model: "claude-haiku-4-5",
        anthropicApiKey: "sk-ant-test",
      }),
    ).toMatchObject({
      provider: "anthropic",
      model: "claude-haiku-4-5",
      anthropicApiKey: "sk-ant-test",
    });
  });

  it("falls back when model does not match provider", () => {
    expect(
      normalizeAISettings({
        provider: "anthropic",
        model: "gpt-4o-mini",
      }).model,
    ).toBe(defaultModelForProvider("anthropic"));
  });
});

describe("isAISettingsConfigured", () => {
  it("requires the active provider key", () => {
    expect(
      isAISettingsConfigured(
        normalizeAISettings({
          provider: "openai",
          openaiApiKey: "sk-test",
        }),
      ),
    ).toBe(true);
    expect(
      isAISettingsConfigured(
        normalizeAISettings({
          provider: "anthropic",
          openaiApiKey: "sk-test",
        }),
      ),
    ).toBe(false);
    expect(
      apiKeyForSettings(
        normalizeAISettings({
          provider: "anthropic",
          anthropicApiKey: "sk-ant-test",
        }),
      ),
    ).toBe("sk-ant-test");
  });

  it("accepts configured flags from the server without returning raw keys", () => {
    expect(
      isAISettingsConfigured(
        normalizeAISettings({
          provider: "openai",
          openaiApiKeyConfigured: true,
        }),
      ),
    ).toBe(true);
  });
});
