import type { StagedFileRecord, StagedFilesMessage, StagedFilesResponse } from "./types";

const MAX_ATTEMPTS = 8;
const RETRY_DELAY_MS = 250;

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => {
    window.setTimeout(resolve, ms);
  });
}

function extensionContextError(): string | null {
  if (!chrome.runtime?.id) {
    return "extension context invalidated — refresh the Canvas tab after reloading the extension";
  }
  return null;
}

function isRetryableSendError(error: string): boolean {
  const lower = error.toLowerCase();
  return (
    lower.includes("receiving end does not exist") ||
    lower.includes("message port closed") ||
    lower.includes("no response from service worker")
  );
}

function sendMessageOnce(message: StagedFilesMessage): Promise<StagedFilesResponse> {
  const contextError = extensionContextError();
  if (contextError) {
    return Promise.resolve({ ok: false, error: contextError });
  }

  return new Promise((resolve) => {
    try {
      chrome.runtime.sendMessage(message, (response: StagedFilesResponse | undefined) => {
        const runtimeError = chrome.runtime.lastError?.message;
        if (runtimeError) {
          resolve({ ok: false, error: runtimeError });
          return;
        }

        if (response && typeof response === "object" && "ok" in response) {
          resolve(response);
          return;
        }

        resolve({ ok: false, error: "no response from service worker" });
      });
    } catch (err) {
      resolve({
        ok: false,
        error: err instanceof Error ? err.message : "sendMessage failed",
      });
    }
  });
}

async function sendMessage(message: StagedFilesMessage): Promise<StagedFilesResponse> {
  let lastError = "service worker unavailable";

  for (let attempt = 1; attempt <= MAX_ATTEMPTS; attempt++) {
    const response = await sendMessageOnce(message);
    if (response.ok) {
      return response;
    }

    lastError = response.error ?? lastError;
    if (!isRetryableSendError(lastError) || attempt === MAX_ATTEMPTS) {
      return response;
    }

    await sleep(RETRY_DELAY_MS * attempt);
  }

  return { ok: false, error: lastError };
}

export async function pingStagedFilesServiceWorker(): Promise<boolean> {
  const response = await sendMessage({ type: "staged-files:ping" });
  return response.ok;
}

export async function putStagedFile(input: {
  assignmentKey: string;
  fileName: string;
  normalizedFileName: string;
  stagedAt: string;
  mimeType: string;
  canvasFileId?: string;
  blobBytes: ArrayBuffer;
}): Promise<void> {
  const response = await sendMessage({ type: "staged-files:put", ...input });
  if (!response.ok) {
    throw new Error(response.error ?? "staged file write failed");
  }
}

export async function deleteStagedFile(input: {
  assignmentKey: string;
  canvasFileId?: string;
  normalizedFileName?: string;
  stagedAt?: string;
}): Promise<void> {
  const response = await sendMessage({ type: "staged-files:delete", ...input });
  if (!response.ok) {
    throw new Error(response.error ?? "staged file delete failed");
  }
}

export async function listStagedFiles(assignmentKey: string): Promise<StagedFileRecord[]> {
  const response = await sendMessage({ type: "staged-files:list", assignmentKey });
  if (!response.ok) {
    return [];
  }
  return response.files ?? [];
}

export async function clearStagedAssignment(assignmentKey: string): Promise<void> {
  const response = await sendMessage({ type: "staged-files:clear-assignment", assignmentKey });
  if (!response.ok) {
    throw new Error(response.error);
  }
}

export async function reconcileStagedFiles(input: {
  assignmentKey: string;
  promotions: Array<{
    normalizedFileName: string;
    stagedAt: string;
    canvasFileId: string;
  }>;
}): Promise<void> {
  const response = await sendMessage({ type: "staged-files:reconcile", ...input });
  if (!response.ok) {
    throw new Error(response.error ?? "staged file reconcile failed");
  }
}

export async function getStagedFilePayload(
  assignmentKey: string,
  record: StagedFileRecord,
): Promise<{
  fileName: string;
  mimeType: string;
  contentBase64: string;
  canvasFileId?: string;
} | null> {
  const response = await sendMessage({
    type: "staged-files:get-blob",
    assignmentKey,
    canvasFileId: record.canvasFileId,
    normalizedFileName: record.normalizedFileName,
    stagedAt: record.stagedAt,
  });
  if (!response.ok || !response.blobBase64) {
    return null;
  }

  return {
    fileName: record.fileName,
    mimeType: response.mimeType ?? record.mimeType,
    contentBase64: response.blobBase64,
    canvasFileId: record.canvasFileId,
  };
}
