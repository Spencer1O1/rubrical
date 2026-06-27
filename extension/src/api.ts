/** Local Rubrical server bases to try. Prefer localhost on WSL — Windows often forwards localhost but not 127.0.0.1. */
export const RUBRICAL_API_BASES = [
  "http://localhost:8787",
  "http://127.0.0.1:8787",
] as const;

export async function postImport(
  payload: unknown,
): Promise<{ data: { redirect?: string }; base: string }> {
  let lastError: unknown;

  for (const base of RUBRICAL_API_BASES) {
    try {
      const response = await fetch(`${base}/imports`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }

      const data = (await response.json()) as { redirect?: string };
      return { data, base };
    } catch (err) {
      lastError = err;
    }
  }

  throw lastError;
}
