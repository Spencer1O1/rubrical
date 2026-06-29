/** Set at build time via extension/build.mjs (`--define:__RUBRICAL_API_BASE__=...`). */
declare const __RUBRICAL_API_BASE__: string;

export const RUBRICAL_API_BASE = __RUBRICAL_API_BASE__;

export function rubricalLoginURL(base: string = RUBRICAL_API_BASE): string {
  return `${base.replace(/\/$/, "")}/login`;
}

export function rubricalWebURL(): string {
  return RUBRICAL_API_BASE;
}
