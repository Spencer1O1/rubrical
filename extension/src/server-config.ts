import { fetchAuthConfig } from "./auth-api";
import { setStrictExtraction } from "./strict";

export async function syncStrictExtractionFromServer(): Promise<boolean> {
  const config = await fetchAuthConfig();
  setStrictExtraction(config.strictExtraction);
  return config.strictExtraction;
}
