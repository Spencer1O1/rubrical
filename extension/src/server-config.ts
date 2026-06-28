import { getRubricalJson } from "./api";
import { setStrictExtraction } from "./strict";

type HealthResponse = {
  strictExtraction?: boolean;
};

export async function syncStrictExtractionFromServer(): Promise<boolean> {
  const data = await getRubricalJson<HealthResponse>("/health");
  if (!data) {
    setStrictExtraction(false);
    return false;
  }

  const strict = data.strictExtraction === true;
  setStrictExtraction(strict);
  return strict;
}
