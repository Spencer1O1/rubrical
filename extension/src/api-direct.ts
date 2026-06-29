import { RUBRICAL_API_BASE } from "./api-bases";
import type { RubricalFetchRequest, RubricalFetchResult } from "./api-fetch-types";
import type { MultipartFetchResult } from "./api-multipart-types";
import { base64ToArrayBuffer } from "./staged-files/file-bytes";

const REQUEST_TIMEOUT_MS = 2500;

export { REQUEST_TIMEOUT_MS };

function isRetryableFetchError(message: string): boolean {
  const lower = message.toLowerCase();
  return (
    lower.includes("failed to fetch") ||
    lower.includes("networkerror") ||
    lower.includes("network error") ||
    lower.includes("timed out") ||
    lower.includes("abort")
  );
}

function authRequiredStatus(status: number): boolean {
  return status === 401 || status === 403;
}

export async function fetchWithTimeout(
  url: string,
  init: RequestInit,
): Promise<Response> {
  return fetch(url, {
    ...init,
    signal: AbortSignal.timeout(REQUEST_TIMEOUT_MS),
  });
}

/** Direct fetch — used from the service worker (not the Canvas content script). */
export async function executeRubricalFetchDirect(
  request: RubricalFetchRequest,
  maxAttempts = 3,
): Promise<RubricalFetchResult> {
  const base = RUBRICAL_API_BASE;
  let lastError = "Failed to fetch";

  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      const response = await fetchWithTimeout(`${base}${request.path}`, {
        cache: "no-store",
        method: request.method ?? "GET",
        headers: request.headers,
        body: request.body,
        credentials: "include",
      });

      const text = await response.text();
      if (!response.ok) {
        if (authRequiredStatus(response.status)) {
          return {
            ok: false,
            error: `HTTP ${response.status}: ${text.slice(0, 200)}`,
            authRequired: true,
            base,
          };
        }
        lastError = `HTTP ${response.status}: ${text.slice(0, 200)}`;
      } else {
        let data: unknown;
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

export type MultipartUploadRequest = {
  path: string;
  fileName: string;
  mimeType: string;
  bytesBase64: string;
  canvasFileId?: string;
};

export async function executeRubricalMultipartDirect(
  request: MultipartUploadRequest,
): Promise<MultipartFetchResult> {
  const base = RUBRICAL_API_BASE;
  const blob = new Blob([base64ToArrayBuffer(request.bytesBase64)], {
    type: request.mimeType || "application/octet-stream",
  });

  try {
    const formData = new FormData();
    formData.append("draft_file", blob, request.fileName);
    if (request.canvasFileId) {
      formData.append("canvas_file_id", request.canvasFileId);
    }

    const response = await fetchWithTimeout(`${base}${request.path}`, {
      method: "POST",
      body: formData,
      cache: "no-store",
      credentials: "include",
    });

    if (!response.ok) {
      const detail = await response.text();
      const error = `HTTP ${response.status}: ${detail.slice(0, 200)}`;
      if (authRequiredStatus(response.status)) {
        return { ok: false, error, authRequired: true, base };
      }
      return { ok: false, error, base };
    }

    return { ok: true, base };
  } catch (err) {
    return {
      ok: false,
      error: err instanceof Error ? err.message : "Failed to fetch",
      base,
    };
  }
}
