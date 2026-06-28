// src/api-bases.ts
var RUBRICAL_API_BASES = [
  "http://localhost:8787",
  "http://127.0.0.1:8787"
];

// src/api-direct.ts
function isRetryableFetchError(message) {
  const lower = message.toLowerCase();
  return lower.includes("failed to fetch") || lower.includes("networkerror") || lower.includes("network error");
}
async function executeRubricalFetchDirect(request, maxAttempts = 3) {
  let lastError = "Failed to fetch";
  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    for (const base of RUBRICAL_API_BASES) {
      try {
        const response = await fetch(`${base}${request.path}`, {
          cache: "no-store",
          method: request.method ?? "GET",
          headers: request.headers,
          body: request.body
        });
        const text = await response.text();
        if (!response.ok) {
          lastError = `HTTP ${response.status}: ${text.slice(0, 200)}`;
          continue;
        }
        let data;
        try {
          data = text ? JSON.parse(text) : null;
        } catch {
          lastError = "Invalid JSON response from Rubrical server";
          continue;
        }
        return { ok: true, data, base };
      } catch (err) {
        lastError = err instanceof Error ? err.message : "Failed to fetch";
      }
    }
    if (!isRetryableFetchError(lastError) || attempt === maxAttempts) {
      break;
    }
    await new Promise((resolve) => {
      setTimeout(resolve, 200 * attempt);
    });
  }
  return { ok: false, error: lastError };
}

// src/api-fetch.ts
function canProxyThroughServiceWorker() {
  return typeof chrome !== "undefined" && typeof chrome.runtime?.sendMessage === "function" && Boolean(chrome.runtime.id);
}
async function sendRubricalApiMessage(message) {
  return new Promise((resolve) => {
    chrome.runtime.sendMessage(message, (response) => {
      if (chrome.runtime.lastError) {
        resolve({
          ok: false,
          error: chrome.runtime.lastError.message ?? "Extension service worker unavailable"
        });
        return;
      }
      if (!response || typeof response !== "object" || !("ok" in response)) {
        resolve({ ok: false, error: "Invalid response from Rubrical service worker" });
        return;
      }
      resolve(response);
    });
  });
}
async function executeRubricalFetch(request, maxAttempts = 3) {
  if (canProxyThroughServiceWorker()) {
    return sendRubricalApiMessage({ type: "rubrical-api:fetch", request, maxAttempts });
  }
  return executeRubricalFetchDirect(request, maxAttempts);
}

// src/ai-settings.ts
var PROVIDER_MODELS = {
  openai: [
    { id: "gpt-4o-mini", label: "GPT-4o mini (recommended)" },
    { id: "gpt-4o", label: "GPT-4o" },
    { id: "gpt-4.1-mini", label: "GPT-4.1 mini" },
    { id: "gpt-4.1", label: "GPT-4.1" }
  ],
  anthropic: [
    { id: "claude-sonnet-4-20250514", label: "Claude Sonnet 4" },
    { id: "claude-3-7-sonnet-20250219", label: "Claude 3.7 Sonnet" },
    { id: "claude-3-5-haiku-20241022", label: "Claude 3.5 Haiku" }
  ]
};
var DEFAULT_AI_SETTINGS = {
  provider: "openai",
  model: PROVIDER_MODELS.openai[0].id,
  openaiApiKey: "",
  anthropicApiKey: ""
};
function defaultModelForProvider(provider) {
  return PROVIDER_MODELS[provider][0]?.id ?? DEFAULT_AI_SETTINGS.model;
}
function normalizeAISettings(raw) {
  const provider = raw?.provider === "anthropic" ? "anthropic" : "openai";
  const models = PROVIDER_MODELS[provider];
  const modelIds = new Set(models.map((entry) => entry.id));
  const model = typeof raw?.model === "string" && modelIds.has(raw.model) ? raw.model : defaultModelForProvider(provider);
  return {
    provider,
    model,
    openaiApiKey: typeof raw?.openaiApiKey === "string" ? raw.openaiApiKey.trim() : "",
    anthropicApiKey: typeof raw?.anthropicApiKey === "string" ? raw.anthropicApiKey.trim() : ""
  };
}
function apiKeyForSettings(settings) {
  return settings.provider === "anthropic" ? settings.anthropicApiKey : settings.openaiApiKey;
}
function isAISettingsConfigured(settings) {
  return apiKeyForSettings(settings).length > 0;
}

// src/ai-settings-api.ts
async function fetchAISettingsFromServer() {
  const result = await executeRubricalFetch({
    path: "/settings/ai?format=json",
    method: "GET",
    headers: {
      Accept: "application/json"
    }
  });
  if (!result.ok) {
    return DEFAULT_AI_SETTINGS;
  }
  return normalizeAISettings(result.data);
}
async function saveAISettingsToServer(settings) {
  const payload = normalizeAISettings(settings);
  const result = await executeRubricalFetch({
    path: "/settings/ai",
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json"
    },
    body: JSON.stringify(payload)
  });
  if (!result.ok) {
    throw new Error(result.error || "Failed to save AI settings");
  }
  return normalizeAISettings(result.data);
}

// src/popup.ts
function byId(id) {
  const element = document.getElementById(id);
  if (!element) {
    throw new Error(`missing #${id}`);
  }
  return element;
}
function renderModelOptions(provider, selectedModel) {
  const modelSelect = byId("model");
  const models = PROVIDER_MODELS[provider];
  modelSelect.replaceChildren(
    ...models.map((entry) => {
      const option = document.createElement("option");
      option.value = entry.id;
      option.textContent = entry.label;
      option.selected = entry.id === selectedModel;
      return option;
    })
  );
  if (!models.some((entry) => entry.id === modelSelect.value)) {
    modelSelect.value = defaultModelForProvider(provider);
  }
}
function readForm() {
  const provider = byId("provider").value;
  return normalizeAISettings({
    provider,
    model: byId("model").value,
    openaiApiKey: byId("openai-key").value,
    anthropicApiKey: byId("anthropic-key").value
  });
}
function setStatus(message, kind = "info") {
  const status = byId("status");
  status.textContent = message;
  status.dataset.kind = kind;
}
function applySettingsToForm(settings) {
  byId("provider").value = settings.provider;
  renderModelOptions(settings.provider, settings.model);
  byId("model").value = settings.model;
  byId("openai-key").value = settings.openaiApiKey;
  byId("anthropic-key").value = settings.anthropicApiKey;
}
async function initPopup() {
  let settings;
  try {
    settings = await fetchAISettingsFromServer();
  } catch {
    settings = normalizeAISettings(void 0);
    setStatus("Could not reach Rubrical. Start the server, then reopen this popup.", "error");
    return;
  }
  applySettingsToForm(settings);
  if (isAISettingsConfigured(settings)) {
    setStatus(`Saved: ${settings.provider} \xB7 ${settings.model}`, "success");
  } else {
    setStatus("Add an API key for your chosen provider, then save.", "info");
  }
  byId("provider").addEventListener("change", () => {
    const provider = byId("provider").value;
    renderModelOptions(provider, defaultModelForProvider(provider));
  });
  byId("settings-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const next = readForm();
    try {
      const saved = await saveAISettingsToServer(next);
      applySettingsToForm(saved);
      setStatus(`Saved ${saved.provider} \xB7 ${saved.model}`, "success");
    } catch (err) {
      setStatus(err instanceof Error ? err.message : "Failed to save settings", "error");
    }
  });
}
void initPopup();
export {
  readForm,
  renderModelOptions
};
