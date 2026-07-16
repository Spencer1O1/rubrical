import type { MultipartFetchResult, RubricalMultipartMessage } from "./api-multipart-types";
import { executeRubricalMultipartDirect } from "./api-direct";
import { arrayBufferToBase64 } from "./staged-files/file-bytes";

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

/** Upload a Blob via the service worker (never fetch from the Canvas page origin). */
export async function postRubricalMultipartBlob(
  path: string,
  blob: Blob,
  fileName: string,
  canvasFileId?: string,
): Promise<MultipartFetchResult> {
  const formData = new FormData();
  formData.append("draft_file", blob, fileName);
  if (canvasFileId?.trim()) {
    formData.append("canvas_file_id", canvasFileId.trim());
  }
  return postRubricalMultipart(path, formData);
}
