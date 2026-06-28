/** Local Rubrical server bases to try. Prefer localhost on WSL — Windows often forwards localhost but not 127.0.0.1. */
export const RUBRICAL_API_BASES = [
  "http://localhost:8787",
  "http://127.0.0.1:8787",
] as const;
