import { executeRubricalFetchDirect } from "./api-direct";
import type {
  RubricalFetchRequest,
  RubricalFetchResult,
} from "./api-fetch-types";
import type { RubricalApiFetchMessage } from "./api-messages";

export type {
  RubricalFetchRequest,
  RubricalFetchResult,
  RubricalFetchSuccess,
  RubricalFetchFailure,
} from "./api-fetch-types";

function canProxyThroughServiceWorker(): boolean {
  return (
    typeof chrome !== "undefined" &&
    typeof chrome.runtime?.sendMessage === "function" &&
    Boolean(chrome.runtime.id)
  );
}

async function sendRubricalApiMessage(message: RubricalApiFetchMessage): Promise<RubricalFetchResult> {
  return new Promise((resolve) => {
    chrome.runtime.sendMessage(message, (response: RubricalFetchResult | undefined) => {
      if (chrome.runtime.lastError) {
        resolve({
          ok: false,
          error: chrome.runtime.lastError.message ?? "Extension service worker unavailable",
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

/** Fetch Rubrical API via the service worker when on Canvas (avoids PNA blocking localhost). */
export async function executeRubricalFetch(
  request: RubricalFetchRequest,
  maxAttempts = 3,
): Promise<RubricalFetchResult> {
  if (canProxyThroughServiceWorker()) {
    return sendRubricalApiMessage({ type: "rubrical-api:fetch", request, maxAttempts });
  }

  return executeRubricalFetchDirect(request, maxAttempts);
}
