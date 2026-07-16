// src/api-bases.ts
var RUBRICAL_API_BASE = "https://rubrical.spencerls.dev";
function rubricalLoginURL(base = RUBRICAL_API_BASE) {
  return `${base.replace(/\/$/, "")}/login`;
}
function rubricalWebURL() {
  return RUBRICAL_API_BASE;
}

// src/api-direct.ts
var REQUEST_TIMEOUT_MS = 2500;
function isRetryableFetchError(message) {
  const lower = message.toLowerCase();
  return lower.includes("failed to fetch") || lower.includes("networkerror") || lower.includes("network error") || lower.includes("timed out") || lower.includes("abort");
}
function authRequiredStatus(status) {
  return status === 401 || status === 403;
}
async function fetchWithTimeout(url, init) {
  return fetch(url, {
    ...init,
    signal: AbortSignal.timeout(REQUEST_TIMEOUT_MS)
  });
}
async function executeRubricalFetchDirect(request, maxAttempts = 3) {
  const base = RUBRICAL_API_BASE;
  let lastError = "Failed to fetch";
  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      const response = await fetchWithTimeout(`${base}${request.path}`, {
        cache: "no-store",
        method: request.method ?? "GET",
        headers: request.headers,
        body: request.body,
        credentials: "include"
      });
      const text = await response.text();
      if (!response.ok) {
        if (authRequiredStatus(response.status)) {
          return {
            ok: false,
            error: `HTTP ${response.status}: ${text.slice(0, 200)}`,
            authRequired: true,
            base
          };
        }
        lastError = `HTTP ${response.status}: ${text.slice(0, 200)}`;
      } else {
        let data;
        try {
          data = text ? JSON.parse(text) : null;
        } catch {
          lastError = "Invalid JSON response from Rubrical server";
          if (attempt === maxAttempts) {
            break;
          }
          await new Promise((resolve) => setTimeout(resolve, 200 * attempt));
          continue;
        }
        return { ok: true, data, base };
      }
    } catch (err) {
      lastError = err instanceof Error ? err.message : "Failed to fetch";
    }
    if (!isRetryableFetchError(lastError) || attempt === maxAttempts) {
      break;
    }
    await new Promise((resolve) => {
      setTimeout(resolve, 200 * attempt);
    });
  }
  return { ok: false, error: lastError, base };
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
  anthropicApiKey: "",
  openaiApiKeyConfigured: false,
  anthropicApiKeyConfigured: false
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
    anthropicApiKey: typeof raw?.anthropicApiKey === "string" ? raw.anthropicApiKey.trim() : "",
    openaiApiKeyConfigured: raw?.openaiApiKeyConfigured === true,
    anthropicApiKeyConfigured: raw?.anthropicApiKeyConfigured === true
  };
}
function isAISettingsConfigured(settings) {
  if (settings.provider === "anthropic") {
    return settings.anthropicApiKey.length > 0 || settings.anthropicApiKeyConfigured === true;
  }
  return settings.openaiApiKey.length > 0 || settings.openaiApiKeyConfigured === true;
}

// src/auth-api.ts
var RubricalAuthRequiredError = class extends Error {
  loginURL;
  constructor(base) {
    super("Sign in to Rubrical to continue.");
    this.name = "RubricalAuthRequiredError";
    this.loginURL = rubricalLoginURL(base ?? RUBRICAL_API_BASE);
  }
};
function authErrorMessage(result) {
  const match = result.error.match(/^HTTP \d+: (.+)$/);
  return match?.[1]?.trim() || result.error;
}
async function fetchAuthConfig() {
  const result = await executeRubricalFetch({
    path: "/auth/config",
    method: "GET",
    headers: { Accept: "application/json" }
  });
  if (!result.ok) {
    return { googleEnabled: false, strictExtraction: false };
  }
  const data = result.data;
  return {
    googleEnabled: Boolean(data.googleEnabled),
    strictExtraction: Boolean(data.strictExtraction)
  };
}
async function fetchSession() {
  const result = await executeRubricalFetch({
    path: "/auth/session",
    method: "GET",
    headers: { Accept: "application/json" }
  });
  if (!result.ok) {
    return null;
  }
  const data = result.data;
  if (!data.email) {
    return null;
  }
  return {
    email: data.email,
    displayName: data.displayName ?? data.email
  };
}
async function loginWithPassword(email, password) {
  const result = await executeRubricalFetch({
    path: "/auth/login",
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json"
    },
    body: JSON.stringify({ email, password })
  });
  if (!result.ok) {
    if (result.authRequired) {
      throw new RubricalAuthRequiredError(result.base);
    }
    throw new Error(authErrorMessage(result));
  }
  const data = result.data;
  if (!data.email) {
    throw new Error("Invalid email or password.");
  }
  return {
    email: data.email,
    displayName: data.displayName ?? data.email
  };
}
async function signupWithPassword(email, password, displayName) {
  const result = await executeRubricalFetch({
    path: "/auth/signup",
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json"
    },
    body: JSON.stringify({ email, password, displayName })
  });
  if (!result.ok) {
    throw new Error(authErrorMessage(result));
  }
  const data = result.data;
  if (!data.email) {
    throw new Error("Could not create account.");
  }
  return {
    email: data.email,
    displayName: data.displayName ?? data.email
  };
}
async function requestPasswordReset(email) {
  const result = await executeRubricalFetch({
    path: "/auth/forgot-password",
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json"
    },
    body: JSON.stringify({ email })
  });
  if (!result.ok) {
    throw new Error(authErrorMessage(result));
  }
  const data = result.data;
  return data.message ?? "If an account exists for that email, a reset link has been sent.";
}
function googleAuthURL(base) {
  const resolved = (base ?? RUBRICAL_API_BASE).replace(/\/$/, "");
  return `${resolved}/auth/google`;
}
function webLoginURL() {
  return `${rubricalWebURL()}/login`;
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
    if (result.authRequired) {
      throw new RubricalAuthRequiredError(result.base);
    }
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
    if (result.authRequired) {
      throw new RubricalAuthRequiredError(result.base);
    }
    throw new Error(result.error || "Failed to save AI settings");
  }
  return normalizeAISettings(result.data);
}

// src/auth-ui.ts
var S = {
  card: [
    "font-family:system-ui,sans-serif",
    "color:#292524",
    "background:#fff",
    "border:1px solid #e7e5e4",
    "border-radius:12px",
    "padding:16px",
    "box-shadow:0 10px 40px rgba(0,0,0,0.12)",
    "width:100%",
    "box-sizing:border-box"
  ].join(";"),
  topRow: "display:flex;align-items:center;justify-content:space-between;margin-bottom:10px;",
  brand: "margin:0;font-size:14px;font-weight:600;color:#312e81;",
  close: "width:28px;height:28px;border:none;border-radius:8px;background:#f5f5f4;color:#44403c;font-size:20px;line-height:1;cursor:pointer;",
  title: "margin:0 0 2px;font-size:18px;font-weight:600;color:#1c1917;",
  subtitle: "margin:0 0 12px;font-size:13px;line-height:1.4;color:#57534e;",
  field: "display:grid;gap:4px;margin-bottom:10px;",
  labelRow: "display:flex;align-items:baseline;justify-content:space-between;gap:8px;",
  label: "font-size:13px;font-weight:600;color:#44403c;",
  input: "display:block;font-family:system-ui,-apple-system,sans-serif;font-size:15px;line-height:1.25;padding:9px 11px;min-height:40px;border:1px solid #d6d3d1;border-radius:8px;background:#fff;color:#1c1917;width:100%;box-sizing:border-box;margin:0;appearance:none;-webkit-appearance:none;",
  hint: "margin:2px 0 0;font-size:11px;color:#78716c;",
  button: "width:100%;font-family:system-ui,-apple-system,sans-serif;font-size:15px;line-height:1.25;padding:9px 11px;min-height:40px;border:none;border-radius:8px;background:#4f46e5;color:#fff;font-weight:600;cursor:pointer;margin-top:2px;",
  buttonSecondary: "width:100%;font-family:system-ui,-apple-system,sans-serif;font-size:15px;line-height:1.25;padding:9px 11px;min-height:40px;border:1px solid #d6d3d1;border-radius:8px;background:#fff;color:#44403c;font-weight:600;cursor:pointer;margin-top:6px;",
  divider: "display:flex;align-items:center;gap:10px;margin:10px 0;color:#78716c;font-size:11px;",
  dividerLine: "flex:1;height:1px;background:#e7e5e4;",
  switch: "margin:10px 0 0;font-size:13px;line-height:1.4;color:#57534e;text-align:center;",
  alertError: "margin-bottom:10px;padding:8px 10px;border-radius:8px;border:1px solid #fecaca;background:#fef2f2;color:#b91c1c;font-size:13px;",
  alertSuccess: "margin-bottom:10px;padding:8px 10px;border-radius:8px;border:1px solid #bbf7d0;background:#ecfdf5;color:#047857;font-size:13px;",
  footer: "margin-top:10px;font-size:12px;color:#78716c;text-align:center;line-height:1.4;"
};
var SHADOW_STYLES = `
  :host {
    display: block;
    width: 100%;
    font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    font-size: 16px;
    line-height: 1.5;
    color: #292524;
  }
  *, *::before, *::after {
    box-sizing: border-box;
  }
  a.text-link {
    display: inline;
    vertical-align: baseline;
    line-height: inherit;
    font-size: inherit;
    font-family: inherit;
    color: #4338ca;
    font-weight: 600;
    text-decoration: none;
    cursor: pointer;
  }
  a.text-link:hover {
    text-decoration: underline;
  }
  a.field-link {
    font-size: 13px;
    line-height: 1.25;
  }
`;
async function mountAuthCard(root, options) {
  const googleEnabled = options.googleEnabled ?? (await fetchAuthConfig()).googleEnabled;
  const host = document.createElement("div");
  host.style.cssText = "display:block;width:100%;";
  root.append(host);
  const shadow = host.attachShadow({ mode: "open" });
  const style = document.createElement("style");
  style.textContent = SHADOW_STYLES;
  shadow.append(style);
  let mode = options.initialMode ?? "login";
  let errorMessage = "";
  let successMessage = "";
  const showBrand = options.showBrand ?? Boolean(options.onClose);
  const card = document.createElement("div");
  card.style.cssText = S.card;
  shadow.append(card);
  function render() {
    card.replaceChildren();
    const top = document.createElement("div");
    top.style.cssText = S.topRow;
    if (showBrand) {
      const brand = document.createElement("p");
      brand.style.cssText = S.brand;
      brand.textContent = "Rubrical";
      top.append(brand);
    }
    if (options.onClose) {
      if (!showBrand) {
        top.style.justifyContent = "flex-end";
      }
      const close = document.createElement("button");
      close.type = "button";
      close.setAttribute("aria-label", "Close");
      close.textContent = "\xD7";
      close.style.cssText = S.close;
      close.addEventListener("click", options.onClose);
      top.append(close);
    }
    if (top.childElementCount > 0) {
      card.append(top);
    }
    const title = document.createElement("h2");
    title.style.cssText = S.title;
    const subtitle = document.createElement("p");
    subtitle.style.cssText = S.subtitle;
    if (mode === "signup") {
      title.textContent = "Create your account";
      subtitle.textContent = "One account for Canvas imports and AI settings.";
    } else if (mode === "forgot") {
      title.textContent = "Reset your password";
      subtitle.textContent = "Enter your email and we will send a reset link.";
    } else {
      title.textContent = "Sign in";
      subtitle.textContent = "Import assignments and save AI settings.";
    }
    card.append(title, subtitle);
    if (errorMessage) card.append(alert(errorMessage, "error"));
    if (successMessage) card.append(alert(successMessage, "success"));
    if (mode === "signup") {
      card.append(buildSignupForm());
      card.append(switchLink("Already have an account? ", "Sign in", "login"));
    } else if (mode === "forgot") {
      if (!successMessage) card.append(buildForgotForm());
      card.append(switchLink("", "Back to sign in", "login"));
    } else {
      card.append(buildLoginForm());
      card.append(switchLink("Don't have an account? ", "Sign up", "signup"));
    }
    const footer = document.createElement("p");
    footer.style.cssText = S.footer;
    footer.innerHTML = `Or use the <a class="text-link" href="${webLoginURL()}" target="_blank" rel="noopener noreferrer">Rubrical website</a>.`;
    card.append(footer);
  }
  function alert(message, kind) {
    const el = document.createElement("div");
    el.style.cssText = kind === "error" ? S.alertError : S.alertSuccess;
    el.textContent = message;
    return el;
  }
  function textLink(label, onClick, extraClass = "") {
    const link = document.createElement("a");
    link.href = "#";
    link.className = extraClass ? `text-link ${extraClass}` : "text-link";
    link.textContent = label;
    link.addEventListener("click", (event) => {
      event.preventDefault();
      onClick();
    });
    return link;
  }
  function switchLink(prefix, linkLabel, nextMode) {
    const wrap = document.createElement("p");
    wrap.style.cssText = S.switch;
    if (prefix) {
      wrap.append(document.createTextNode(prefix));
    }
    wrap.append(
      textLink(linkLabel, () => {
        mode = nextMode;
        errorMessage = "";
        successMessage = "";
        render();
      })
    );
    return wrap;
  }
  function field(label, input) {
    const wrap = document.createElement("label");
    wrap.style.cssText = S.field;
    const text = document.createElement("span");
    text.style.cssText = S.label;
    text.textContent = label;
    input.style.cssText = S.input;
    wrap.append(text, input);
    return wrap;
  }
  function passwordField(label, input, forgotLink) {
    const wrap = document.createElement("div");
    wrap.style.cssText = S.field;
    const row = document.createElement("div");
    row.style.cssText = S.labelRow;
    const text = document.createElement("span");
    text.style.cssText = S.label;
    text.textContent = label;
    row.append(text);
    if (forgotLink) {
      row.append(
        textLink("Forgot password?", () => {
          mode = "forgot";
          errorMessage = "";
          render();
        }, "field-link")
      );
    }
    input.style.cssText = S.input;
    wrap.append(row, input);
    return wrap;
  }
  function submitButton(label) {
    const button = document.createElement("button");
    button.type = "submit";
    button.style.cssText = S.button;
    button.textContent = label;
    return button;
  }
  function googleButton() {
    const button = document.createElement("button");
    button.type = "button";
    button.style.cssText = S.buttonSecondary;
    button.textContent = "Continue with Google";
    button.addEventListener("click", () => {
      void chrome.tabs.create({ url: googleAuthURL(RUBRICAL_API_BASE) });
    });
    return button;
  }
  function orDivider() {
    const wrap = document.createElement("div");
    wrap.style.cssText = S.divider;
    const left = document.createElement("div");
    left.style.cssText = S.dividerLine;
    const text = document.createElement("span");
    text.textContent = "or";
    const right = document.createElement("div");
    right.style.cssText = S.dividerLine;
    wrap.append(left, text, right);
    return wrap;
  }
  function buildLoginForm() {
    const form = document.createElement("form");
    const email = document.createElement("input");
    email.type = "email";
    email.required = true;
    email.autocomplete = "email";
    const password = document.createElement("input");
    password.type = "password";
    password.required = true;
    password.autocomplete = "current-password";
    form.append(field("Email", email), passwordField("Password", password, true), submitButton("Sign in"));
    if (googleEnabled) {
      form.append(orDivider(), googleButton());
    }
    form.addEventListener("submit", (event) => {
      event.preventDefault();
      void (async () => {
        errorMessage = "";
        try {
          options.onSignedIn(await loginWithPassword(email.value, password.value));
        } catch (err) {
          errorMessage = err instanceof Error ? err.message : "Sign in failed.";
          render();
        }
      })();
    });
    return form;
  }
  function buildSignupForm() {
    const form = document.createElement("form");
    form.style.cssText = "margin:0;padding:0;";
    const displayName = document.createElement("input");
    displayName.type = "text";
    displayName.autocomplete = "name";
    const email = document.createElement("input");
    email.type = "email";
    email.required = true;
    email.autocomplete = "email";
    const password = document.createElement("input");
    password.type = "password";
    password.required = true;
    password.minLength = 8;
    password.autocomplete = "new-password";
    form.append(field("Display name", displayName), field("Email", email));
    const passwordWrap = passwordField("Password", password);
    const hint = document.createElement("p");
    hint.style.cssText = S.hint;
    hint.textContent = "At least 8 characters.";
    passwordWrap.append(hint);
    form.append(passwordWrap, submitButton("Create account"));
    if (googleEnabled) {
      form.append(orDivider(), googleButton());
    }
    form.addEventListener("submit", (event) => {
      event.preventDefault();
      void (async () => {
        errorMessage = "";
        try {
          options.onSignedIn(
            await signupWithPassword(email.value, password.value, displayName.value)
          );
        } catch (err) {
          errorMessage = err instanceof Error ? err.message : "Sign up failed.";
          render();
        }
      })();
    });
    return form;
  }
  function buildForgotForm() {
    const form = document.createElement("form");
    const email = document.createElement("input");
    email.type = "email";
    email.required = true;
    email.autocomplete = "email";
    form.append(field("Email", email), submitButton("Send reset link"));
    form.addEventListener("submit", (event) => {
      event.preventDefault();
      void (async () => {
        errorMessage = "";
        try {
          successMessage = await requestPasswordReset(email.value);
          render();
        } catch (err) {
          errorMessage = err instanceof Error ? err.message : "Request failed.";
          render();
        }
      })();
    });
    return form;
  }
  render();
  return { destroy: () => host.remove() };
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
function updateKeyConfiguredHints(settings) {
  byId("openai-key-status").classList.toggle("hidden", !settings.openaiApiKeyConfigured);
  byId("anthropic-key-status").classList.toggle("hidden", !settings.anthropicApiKeyConfigured);
}
function applySettingsToForm(settings) {
  byId("provider").value = settings.provider;
  renderModelOptions(settings.provider, settings.model);
  byId("model").value = settings.model;
  byId("openai-key").value = "";
  byId("anthropic-key").value = "";
  updateKeyConfiguredHints(settings);
}
function showAuthView() {
  byId("auth-view").classList.remove("hidden");
  byId("settings-view").classList.add("hidden");
}
function showSettingsView(session) {
  byId("signed-in-email").textContent = session.email;
  byId("auth-view").classList.add("hidden");
  byId("settings-view").classList.remove("hidden");
}
async function initSettings(session) {
  showSettingsView(session);
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
    setStatus(`${settings.provider} \xB7 ${settings.model}`, "success");
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
async function initPopup() {
  const session = await fetchSession();
  if (session) {
    await initSettings(session);
    return;
  }
  showAuthView();
  await mountAuthCard(byId("auth-root"), {
    onSignedIn: (signedIn) => {
      void initSettings(signedIn);
    }
  });
}
void initPopup();
export {
  readForm,
  renderModelOptions
};
