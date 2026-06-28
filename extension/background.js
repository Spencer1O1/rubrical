"use strict";
(() => {
  // src/api-bases.ts
  var RUBRICAL_API_BASES = [
    "http://localhost:8787",
    "http://127.0.0.1:8787"
  ];

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
  function isRetryableFetchError(message) {
    const lower = message.toLowerCase();
    return lower.includes("failed to fetch") || lower.includes("networkerror") || lower.includes("network error");
  }
  async function executeRubricalFetchDirect(request, maxAttempts = 3) {
    let lastError = "Failed to fetch";
    for (let attempt = 1; attempt <= maxAttempts; attempt++) {
      for (const base of RUBRICAL_API_BASES) {
        try {
          const response = await fetch(`${base}${request.path}`, {
            cache: "no-store",
            method: request.method ?? "GET",
            headers: request.headers,
            body: request.body
          });
          const text = await response.text();
          if (!response.ok) {
            lastError = `HTTP ${response.status}: ${text.slice(0, 200)}`;
            continue;
          }
          let data;
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
  async function executeRubricalMultipartDirect(request) {
    let lastError = "Failed to fetch";
    const blob = new Blob([base64ToArrayBuffer(request.bytesBase64)], {
      type: request.mimeType || "application/octet-stream"
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
          cache: "no-store"
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
