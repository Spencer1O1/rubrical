import { RUBRICAL_API_BASES } from "./api-bases";
import type { RubricalFetchRequest, RubricalFetchResult } from "./api-fetch-types";
import type { MultipartFetchResult } from "./api-multipart-types";

function isRetryableFetchError(message: string): boolean {
  const lower = message.toLowerCase();
  return (
    lower.includes("failed to fetch") ||
    lower.includes("networkerror") ||
    lower.includes("network error")
  );
}

/** Direct fetch — used from the service worker (not the Canvas content script). */
export async function executeRubricalFetchDirect(
  request: RubricalFetchRequest,
  maxAttempts = 3,
): Promise<RubricalFetchResult> {
  let lastError = "Failed to fetch";

  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    for (const base of RUBRICAL_API_BASES) {
      try {
        const response = await fetch(`${base}${request.path}`, {
          cache: "no-store",
          method: request.method ?? "GET",
          headers: request.headers,
          body: request.body,
        });

        const text = await response.text();
        if (!response.ok) {
          lastError = `HTTP ${response.status}: ${text.slice(0, 200)}`;
          continue;
        }

        let data: unknown;
        try {
          data = text ? JSON.parse(text) : null;
        } catch {
          lastError = "Invalid JSON response from Rubrical server";
          continue;
        }

        return { ok: true, data, base };
      } catch (err) {
        lastError = err instanceof Error ? err.message : "Failed to fetch";
      }
    }

    if (!isRetryableFetchError(lastError) || attempt === maxAttempts) {
      break;
    }

    await new Promise((resolve) => {
      setTimeout(resolve, 200 * attempt);
    });
  }

  return { ok: false, error: lastError };
}

export type MultipartUploadRequest = {
  path: string;
  fileName: string;
  mimeType: string;
  bytes: ArrayBuffer;
  canvasFileId?: string;
};

export async function executeRubricalMultipartDirect(
  request: MultipartUploadRequest,
): Promise<MultipartFetchResult> {
  let lastError = "Failed to fetch";
  const blob = new Blob([request.bytes], {
    type: request.mimeType || "application/octet-stream",
  });

  for (const base of RUBRICAL_API_BASES) {
    try {
      const formData = new FormData();
      formData.append("draft_file", blob, request.fileName);
      if (request.canvasFileId) {
        formData.append("canvas_file_id", request.canvasFileId);
      }

      const response = await fetch(`${base}${request.path}`, {
        method: "POST",
        body: formData,
        cache: "no-store",
      });

      if (!response.ok) {
        const detail = await response.text();
        lastError = `HTTP ${response.status}: ${detail.slice(0, 200)}`;
        continue;
      }

      return { ok: true, base };
    } catch (err) {
      lastError = err instanceof Error ? err.message : "Failed to fetch";
    }
  }

  return { ok: false, error: lastError };
}
