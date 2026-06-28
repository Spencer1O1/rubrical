import type { MultipartFetchResult, RubricalMultipartMessage } from "./api-multipart-types";
import { executeRubricalMultipartDirect } from "./api-direct";
import { arrayBufferToBase64 } from "./staged-files/file-bytes";
import { RUBRICAL_API_BASES } from "./api-bases";

export type { MultipartFetchResult } from "./api-multipart-types";

function canProxyThroughServiceWorker(): boolean {
  return (
    typeof chrome !== "undefined" &&
    typeof chrome.runtime?.sendMessage === "function" &&
    Boolean(chrome.runtime.id)
  );
}

async function sendMultipartMessage(
  message: RubricalMultipartMessage,
): Promise<MultipartFetchResult> {
  return new Promise((resolve) => {
    chrome.runtime.sendMessage(message, (response: MultipartFetchResult | undefined) => {
      if (chrome.runtime.lastError) {
        resolve({
          ok: false,
          error: chrome.runtime.lastError.message ?? "Extension service worker unavailable",
        });
        return;
      }

      if (!response || typeof response !== "object" || !("ok" in response)) {
        resolve({ ok: false, error: "Invalid multipart response from service worker" });
        return;
      }

      resolve(response);
    });
  });
}

export async function postRubricalMultipart(
  path: string,
  formData: FormData,
): Promise<MultipartFetchResult> {
  const file = formData.get("draft_file");
  if (!(file instanceof Blob)) {
    return { ok: false, error: "draft_file is required" };
  }

  const fileName = file instanceof File ? file.name : "attachment";
  const canvasFileId = formData.get("canvas_file_id");
  const bytes = await file.arrayBuffer();
  const message: RubricalMultipartMessage = {
    type: "rubrical-api:multipart",
    path,
    fileName,
    mimeType: file.type || "application/octet-stream",
    bytesBase64: arrayBufferToBase64(bytes),
    canvasFileId: typeof canvasFileId === "string" && canvasFileId.trim() !== "" ? canvasFileId : undefined,
  };

  if (canProxyThroughServiceWorker()) {
    return sendMultipartMessage(message);
  }

  return executeRubricalMultipartDirect(message);
}

/** Upload from a Blob in the content script — no base64 over sendMessage. */
export async function postRubricalMultipartBlob(
  path: string,
  blob: Blob,
  fileName: string,
  canvasFileId?: string,
): Promise<MultipartFetchResult> {
  let lastError = "Failed to fetch";

  for (const base of RUBRICAL_API_BASES) {
    try {
      const formData = new FormData();
      formData.append("draft_file", blob, fileName);
      if (canvasFileId?.trim()) {
        formData.append("canvas_file_id", canvasFileId.trim());
      }

      const response = await fetch(`${base}${path}`, {
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
