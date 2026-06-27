import { RUBRICAL_API_BASES } from "./api";
import { setStrictExtraction } from "./strict";

type HealthResponse = {
  strictExtraction?: boolean;
};

export async function syncStrictExtractionFromServer(): Promise<boolean> {
  for (const base of RUBRICAL_API_BASES) {
    try {
      const response = await fetch(`${base}/health`);
      if (!response.ok) {
        continue;
      }

      const data = (await response.json()) as HealthResponse;
      const strict = data.strictExtraction === true;
      setStrictExtraction(strict);
      return strict;
    } catch {
      // try next base
    }
  }

  setStrictExtraction(false);
  return false;
}
