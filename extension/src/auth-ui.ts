import {
  fetchAuthConfig,
  fetchSession,
  googleAuthURL,
  loginWithPassword,
  requestPasswordReset,
  signupWithPassword,
  webLoginURL,
  type RubricalSession,
} from "./auth-api";
import { RUBRICAL_API_BASE } from "./api-bases";

export type AuthMode = "login" | "signup" | "forgot";

export type AuthCardOptions = {
  initialMode?: AuthMode;
  onSignedIn: (session: RubricalSession) => void;
  onClose?: () => void;
  /** Override server `GET /auth/config` when set explicitly (e.g. tests). */
  googleEnabled?: boolean;
  /** Show the small “Rubrical” brand row. Default: true when `onClose` is set (modal). */
  showBrand?: boolean;
};

const S = {
  card: [
    "font-family:system-ui,sans-serif",
    "color:#292524",
    "background:#fff",
    "border:1px solid #e7e5e4",
    "border-radius:12px",
    "padding:16px",
    "box-shadow:0 10px 40px rgba(0,0,0,0.12)",
    "width:100%",
    "box-sizing:border-box",
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
  footer: "margin-top:10px;font-size:12px;color:#78716c;text-align:center;line-height:1.4;",
};

const SHADOW_STYLES = `
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

export async function mountAuthCard(
  root: HTMLElement,
  options: AuthCardOptions,
): Promise<{ destroy: () => void }> {
  const googleEnabled =
    options.googleEnabled ?? (await fetchAuthConfig()).googleEnabled;

  const host = document.createElement("div");
  host.style.cssText = "display:block;width:100%;";
  root.append(host);

  const shadow = host.attachShadow({ mode: "open" });
  const style = document.createElement("style");
  style.textContent = SHADOW_STYLES;
  shadow.append(style);

  let mode: AuthMode = options.initialMode ?? "login";
  let errorMessage = "";
  let successMessage = "";
  const showBrand = options.showBrand ?? Boolean(options.onClose);

  const card = document.createElement("div");
  card.style.cssText = S.card;
  shadow.append(card);

  function render(): void {
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
      close.textContent = "×";
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

  function alert(message: string, kind: "error" | "success"): HTMLElement {
    const el = document.createElement("div");
    el.style.cssText = kind === "error" ? S.alertError : S.alertSuccess;
    el.textContent = message;
    return el;
  }

  function textLink(label: string, onClick: () => void, extraClass = ""): HTMLAnchorElement {
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

  function switchLink(prefix: string, linkLabel: string, nextMode: AuthMode): HTMLElement {
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
      }),
    );
    return wrap;
  }

  function field(label: string, input: HTMLInputElement): HTMLElement {
    const wrap = document.createElement("label");
    wrap.style.cssText = S.field;
    const text = document.createElement("span");
    text.style.cssText = S.label;
    text.textContent = label;
    input.style.cssText = S.input;
    wrap.append(text, input);
    return wrap;
  }

  function passwordField(
    label: string,
    input: HTMLInputElement,
    forgotLink?: boolean,
  ): HTMLElement {
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
        }, "field-link"),
      );
    }
    input.style.cssText = S.input;
    wrap.append(row, input);
    return wrap;
  }

  function submitButton(label: string): HTMLButtonElement {
    const button = document.createElement("button");
    button.type = "submit";
    button.style.cssText = S.button;
    button.textContent = label;
    return button;
  }

  function googleButton(): HTMLButtonElement {
    const button = document.createElement("button");
    button.type = "button";
    button.style.cssText = S.buttonSecondary;
    button.textContent = "Continue with Google";
    button.addEventListener("click", () => {
      void chrome.tabs.create({ url: googleAuthURL(RUBRICAL_API_BASE) });
    });
    return button;
  }

  function orDivider(): HTMLElement {
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

  function buildLoginForm(): HTMLFormElement {
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

  function buildSignupForm(): HTMLFormElement {
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
            await signupWithPassword(email.value, password.value, displayName.value),
          );
        } catch (err) {
          errorMessage = err instanceof Error ? err.message : "Sign up failed.";
          render();
        }
      })();
    });
    return form;
  }

  function buildForgotForm(): HTMLFormElement {
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

export async function isSignedIn(): Promise<boolean> {
  return (await fetchSession()) !== null;
}
