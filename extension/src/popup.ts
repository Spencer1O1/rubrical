import {
  fetchAISettingsFromServer,
  saveAISettingsToServer,
} from "./ai-settings-api";
import { fetchSession, type RubricalSession } from "./auth-api";
import { mountAuthCard } from "./auth-ui";
import {
  defaultModelForProvider,
  isAISettingsConfigured,
  normalizeAISettings,
  PROVIDER_MODELS,
  type AIProviderId,
  type AISettings,
} from "./ai-settings";

function byId<T extends HTMLElement>(id: string): T {
  const element = document.getElementById(id);
  if (!element) {
    throw new Error(`missing #${id}`);
  }
  return element as T;
}

function renderModelOptions(provider: AIProviderId, selectedModel: string): void {
  const modelSelect = byId<HTMLSelectElement>("model");
  const models = PROVIDER_MODELS[provider];
  modelSelect.replaceChildren(
    ...models.map((entry) => {
      const option = document.createElement("option");
      option.value = entry.id;
      option.textContent = entry.label;
      option.selected = entry.id === selectedModel;
      return option;
    }),
  );
  if (!models.some((entry) => entry.id === modelSelect.value)) {
    modelSelect.value = defaultModelForProvider(provider);
  }
}

function readForm(): AISettings {
  const provider = byId<HTMLSelectElement>("provider").value as AIProviderId;
  return normalizeAISettings({
    provider,
    model: byId<HTMLSelectElement>("model").value,
    openaiApiKey: byId<HTMLInputElement>("openai-key").value,
    anthropicApiKey: byId<HTMLInputElement>("anthropic-key").value,
  });
}

function setStatus(message: string, kind: "info" | "success" | "error" = "info"): void {
  const status = byId<HTMLParagraphElement>("status");
  status.textContent = message;
  status.dataset.kind = kind;
}

function updateKeyConfiguredHints(settings: AISettings): void {
  byId<HTMLSpanElement>("openai-key-status").classList.toggle("hidden", !settings.openaiApiKeyConfigured);
  byId<HTMLSpanElement>("anthropic-key-status").classList.toggle("hidden", !settings.anthropicApiKeyConfigured);
}

function applySettingsToForm(settings: AISettings): void {
  byId<HTMLSelectElement>("provider").value = settings.provider;
  renderModelOptions(settings.provider, settings.model);
  byId<HTMLSelectElement>("model").value = settings.model;
  byId<HTMLInputElement>("openai-key").value = "";
  byId<HTMLInputElement>("anthropic-key").value = "";
  updateKeyConfiguredHints(settings);
}

function showAuthView(): void {
  byId<HTMLDivElement>("auth-view").classList.remove("hidden");
  byId<HTMLDivElement>("settings-view").classList.add("hidden");
}

function showSettingsView(session: RubricalSession): void {
  byId<HTMLSpanElement>("signed-in-email").textContent = session.email;
  byId<HTMLDivElement>("auth-view").classList.add("hidden");
  byId<HTMLDivElement>("settings-view").classList.remove("hidden");
}

async function initSettings(session: RubricalSession): Promise<void> {
  showSettingsView(session);

  let settings: AISettings;
  try {
    settings = await fetchAISettingsFromServer();
  } catch {
    settings = normalizeAISettings(undefined);
    setStatus("Could not reach Rubrical. Start the server, then reopen this popup.", "error");
    return;
  }

  applySettingsToForm(settings);

  if (isAISettingsConfigured(settings)) {
    setStatus(`${settings.provider} · ${settings.model}`, "success");
  } else {
    setStatus("Add an API key for your chosen provider, then save.", "info");
  }

  byId<HTMLSelectElement>("provider").addEventListener("change", () => {
    const provider = byId<HTMLSelectElement>("provider").value as AIProviderId;
    renderModelOptions(provider, defaultModelForProvider(provider));
  });

  byId<HTMLFormElement>("settings-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const next = readForm();
    try {
      const saved = await saveAISettingsToServer(next);
      applySettingsToForm(saved);
      setStatus(`Saved ${saved.provider} · ${saved.model}`, "success");
    } catch (err) {
      setStatus(err instanceof Error ? err.message : "Failed to save settings", "error");
    }
  });
}

async function initPopup(): Promise<void> {
  const session = await fetchSession();
  if (session) {
    await initSettings(session);
    return;
  }

  showAuthView();
  await mountAuthCard(byId<HTMLDivElement>("auth-root"), {
    onSignedIn: (signedIn) => {
      void initSettings(signedIn);
    },
  });
}

void initPopup();

export { readForm, renderModelOptions };
