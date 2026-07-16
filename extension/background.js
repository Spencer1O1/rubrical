"use strict";
(() => {
  // src/api-bases.ts
  var RUBRICAL_API_BASE = "https://rubrical.spencerls.dev";

  // src/staged-files/file-bytes.ts
  function base64ToArrayBuffer(base64) {
    const binary = atob(base64);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i);
    }
    return bytes.buffer;
  }

  // src/api-direct.ts
  var REQUEST_TIMEOUT_MS = 3e4;
  var MULTIPART_TIMEOUT_MS = 12e4;
  function isRetryableFetchError(message) {
    const lower = message.toLowerCase();
    return lower.includes("failed to fetch") || lower.includes("networkerror") || lower.includes("network error") || lower.includes("timed out") || lower.includes("abort");
  }
  function authRequiredStatus(status) {
    return status === 401 || status === 403;
  }
  async function fetchWithTimeout(url, init, timeoutMs = REQUEST_TIMEOUT_MS) {
    return fetch(url, {
      ...init,
      signal: AbortSignal.timeout(timeoutMs)
    });
  }
  async function executeRubricalFetchDirect(request, maxAttempts = 3) {
    const base = RUBRICAL_API_BASE;
    let lastError = "Failed to fetch";
    for (let attempt = 1; attempt <= maxAttempts; attempt++) {
      try {
        const response = await fetchWithTimeout(`${base}${request.path}`, {
          cache: "no-store",
          method: request.method ?? "GET",
          headers: request.headers,
          body: request.body,
          credentials: "include"
        });
        const text = await response.text();
        if (!response.ok) {
          if (authRequiredStatus(response.status)) {
            return {
              ok: false,
              error: `HTTP ${response.status}: ${text.slice(0, 200)}`,
              authRequired: true,
              base
            };
          }
          lastError = `HTTP ${response.status}: ${text.slice(0, 200)}`;
        } else {
          let data;
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
  async function executeRubricalMultipartDirect(request) {
    const base = RUBRICAL_API_BASE;
    const blob = new Blob([base64ToArrayBuffer(request.bytesBase64)], {
      type: request.mimeType || "application/octet-stream"
    });
    try {
      const formData = new FormData();
      formData.append("draft_file", blob, request.fileName);
      if (request.canvasFileId) {
        formData.append("canvas_file_id", request.canvasFileId);
      }
      const response = await fetchWithTimeout(
        `${base}${request.path}`,
        {
          method: "POST",
          body: formData,
          cache: "no-store",
          credentials: "include"
        },
        MULTIPART_TIMEOUT_MS
      );
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
        base
      };
    }
  }

  // src/api-messages.ts
  function isRubricalApiMessage(message) {
    if (typeof message !== "object" || message === null || !("type" in message)) {
      return false;
    }
    const type = message.type;
    return type === "rubrical-api:fetch" || type === "rubrical-api:multipart";
  }

  // src/background.ts
  chrome.runtime.onMessage.addListener((message, _sender, sendResponse) => {
    if (!isRubricalApiMessage(message)) {
      return false;
    }
    if (message.type === "rubrical-api:fetch") {
      void executeRubricalFetchDirect(message.request, message.maxAttempts).then(sendResponse);
      return true;
    }
    void executeRubricalMultipartDirect(message).then(sendResponse);
    return true;
  });
})();
