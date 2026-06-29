import { executeRubricalFetch } from "./api-fetch";
import { RUBRICAL_API_BASE, rubricalLoginURL, rubricalWebURL } from "./api-bases";

export type RubricalSession = {
  email: string;
  displayName: string;
};

export type AuthConfig = {
  googleEnabled: boolean;
  strictExtraction: boolean;
};

export class RubricalAuthRequiredError extends Error {
  readonly loginURL: string;

  constructor(base?: string) {
    super("Sign in to Rubrical to continue.");
    this.name = "RubricalAuthRequiredError";
    this.loginURL = rubricalLoginURL(base ?? RUBRICAL_API_BASE);
  }
}

function authErrorMessage(result: { ok: false; error: string }): string {
  const match = result.error.match(/^HTTP \d+: (.+)$/);
  return match?.[1]?.trim() || result.error;
}

export async function fetchAuthConfig(): Promise<AuthConfig> {
  const result = await executeRubricalFetch({
    path: "/auth/config",
    method: "GET",
    headers: { Accept: "application/json" },
  });
  if (!result.ok) {
    return { googleEnabled: false, strictExtraction: false };
  }
  const data = result.data as Partial<AuthConfig>;
  return {
    googleEnabled: Boolean(data.googleEnabled),
    strictExtraction: Boolean(data.strictExtraction),
  };
}

export async function fetchSession(): Promise<RubricalSession | null> {
  const result = await executeRubricalFetch({
    path: "/auth/session",
    method: "GET",
    headers: { Accept: "application/json" },
  });
  if (!result.ok) {
    return null;
  }
  const data = result.data as Partial<RubricalSession>;
  if (!data.email) {
    return null;
  }
  return {
    email: data.email,
    displayName: data.displayName ?? data.email,
  };
}

export async function loginWithPassword(email: string, password: string): Promise<RubricalSession> {
  const result = await executeRubricalFetch({
    path: "/auth/login",
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ email, password }),
  });
  if (!result.ok) {
    if (result.authRequired) {
      throw new RubricalAuthRequiredError(result.base);
    }
    throw new Error(authErrorMessage(result));
  }
  const data = result.data as Partial<RubricalSession>;
  if (!data.email) {
    throw new Error("Invalid email or password.");
  }
  return {
    email: data.email,
    displayName: data.displayName ?? data.email,
  };
}

export async function signupWithPassword(
  email: string,
  password: string,
  displayName: string,
): Promise<RubricalSession> {
  const result = await executeRubricalFetch({
    path: "/auth/signup",
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ email, password, displayName }),
  });
  if (!result.ok) {
    throw new Error(authErrorMessage(result));
  }
  const data = result.data as Partial<RubricalSession>;
  if (!data.email) {
    throw new Error("Could not create account.");
  }
  return {
    email: data.email,
    displayName: data.displayName ?? data.email,
  };
}

export async function requestPasswordReset(email: string): Promise<string> {
  const result = await executeRubricalFetch({
    path: "/auth/forgot-password",
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ email }),
  });
  if (!result.ok) {
    throw new Error(authErrorMessage(result));
  }
  const data = result.data as { message?: string };
  return data.message ?? "If an account exists for that email, a reset link has been sent.";
}

export function googleAuthURL(base?: string): string {
  const resolved = (base ?? RUBRICAL_API_BASE).replace(/\/$/, "");
  return `${resolved}/auth/google`;
}

export function webLoginURL(): string {
  return `${rubricalWebURL()}/login`;
}
