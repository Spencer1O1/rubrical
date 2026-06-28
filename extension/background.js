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

  // src/staged-files/staging-key.ts
  function provisionalStorageKey(normalizedFileName, stagedAt) {
    return `provisional:${normalizedFileName}:${stagedAt}`;
  }
  function canvasStorageKey(canvasFileId) {
    return `canvas:${canvasFileId}`;
  }

  // src/staged-files/idb.ts
  var DB_NAME = "rubrical-staged-files";
  var DB_VERSION = 1;
  var STORE = "files";
  function openDb() {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open(DB_NAME, DB_VERSION);
      request.onupgradeneeded = () => {
        const db = request.result;
        if (!db.objectStoreNames.contains(STORE)) {
          const store = db.createObjectStore(STORE, { keyPath: "id" });
          store.createIndex("assignmentKey", "assignmentKey", { unique: false });
        }
      };
      request.onsuccess = () => resolve(request.result);
      request.onerror = () => reject(request.error ?? new Error("idb open failed"));
    });
  }
  function rowId(assignmentKey, canvasFileId, normalizedFileName, stagedAt) {
    if (canvasFileId) {
      return `${assignmentKey}:${canvasStorageKey(canvasFileId)}`;
    }
    return `${assignmentKey}:${provisionalStorageKey(normalizedFileName ?? "", stagedAt ?? "")}`;
  }
  function listFilesInDb(db, assignmentKey) {
    return new Promise((resolve, reject) => {
      const tx = db.transaction(STORE, "readonly");
      const request = tx.objectStore(STORE).index("assignmentKey").getAll(assignmentKey);
      request.onsuccess = () => resolve(request.result ?? []);
      request.onerror = () => reject(request.error ?? new Error("idb list failed"));
    });
  }
  function getRowInDb(db, id) {
    return new Promise((resolve, reject) => {
      const tx = db.transaction(STORE, "readonly");
      const request = tx.objectStore(STORE).get(id);
      request.onsuccess = () => resolve(request.result);
      request.onerror = () => reject(request.error ?? new Error("idb get failed"));
    });
  }
  async function putFile(message) {
    const encoded = message.blobBase64.trim();
    if (encoded === "") {
      throw new Error("staged file payload is empty");
    }
    const db = await openDb();
    const id = rowId(
      message.assignmentKey,
      message.canvasFileId,
      message.normalizedFileName,
      message.stagedAt
    );
    const row = {
      id,
      assignmentKey: message.assignmentKey,
      canvasFileId: message.canvasFileId,
      fileName: message.fileName,
      normalizedFileName: message.normalizedFileName,
      stagedAt: message.stagedAt,
      mimeType: message.mimeType,
      blob: new Blob([base64ToArrayBuffer(message.blobBase64)], {
        type: message.mimeType || "application/octet-stream"
      })
    };
    await new Promise((resolve, reject) => {
      const tx = db.transaction(STORE, "readwrite");
      tx.objectStore(STORE).put(row);
      tx.oncomplete = () => resolve();
      tx.onerror = () => reject(tx.error ?? new Error("idb put failed"));
    });
    db.close();
  }
  async function deleteFile(message) {
    const db = await openDb();
    const id = rowId(
      message.assignmentKey,
      message.canvasFileId,
      message.normalizedFileName,
      message.stagedAt
    );
    await new Promise((resolve, reject) => {
      const tx = db.transaction(STORE, "readwrite");
      tx.objectStore(STORE).delete(id);
      tx.oncomplete = () => resolve();
      tx.onerror = () => reject(tx.error ?? new Error("idb delete failed"));
    });
    db.close();
  }
  async function listFiles(assignmentKey) {
    const db = await openDb();
    const rows = await listFilesInDb(db, assignmentKey);
    db.close();
    return rows.map(({ id: _id, blob: _blob, ...record }) => record);
  }
  async function clearAssignment(assignmentKey) {
    const db = await openDb();
    const rows = await listFilesInDb(db, assignmentKey);
    await new Promise((resolve, reject) => {
      const tx = db.transaction(STORE, "readwrite");
      const store = tx.objectStore(STORE);
      for (const row of rows) {
        store.delete(row.id);
      }
      tx.oncomplete = () => resolve();
      tx.onerror = () => reject(tx.error ?? new Error("idb clear failed"));
    });
    db.close();
  }
  async function reconcile(message) {
    const db = await openDb();
    for (const promotion of message.promotions) {
      const provisionalId = rowId(
        message.assignmentKey,
        void 0,
        promotion.normalizedFileName,
        promotion.stagedAt
      );
      const provisional = await getRowInDb(db, provisionalId);
      if (!provisional?.blob) {
        continue;
      }
      const promoted = {
        ...provisional,
        canvasFileId: promotion.canvasFileId,
        id: rowId(message.assignmentKey, promotion.canvasFileId)
      };
      await new Promise((resolve, reject) => {
        const tx = db.transaction(STORE, "readwrite");
        const store = tx.objectStore(STORE);
        store.delete(provisionalId);
        store.put(promoted);
        tx.oncomplete = () => resolve();
        tx.onerror = () => reject(tx.error ?? new Error("idb reconcile failed"));
      });
    }
    db.close();
  }
  function blobToBase64(blob) {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => {
        const result = reader.result;
        if (typeof result !== "string") {
          reject(new Error("read failed"));
          return;
        }
        resolve(result.split(",", 2)[1] ?? "");
      };
      reader.onerror = () => reject(reader.error ?? new Error("read failed"));
      reader.readAsDataURL(blob);
    });
  }
  async function getBlob(message) {
    const db = await openDb();
    const id = rowId(
      message.assignmentKey,
      message.canvasFileId,
      message.normalizedFileName,
      message.stagedAt
    );
    const row = await getRowInDb(db, id);
    db.close();
    if (!row?.blob) {
      return null;
    }
    return {
      blobBase64: await blobToBase64(row.blob),
      mimeType: row.mimeType || row.blob.type || "application/octet-stream"
    };
  }
  async function handleStagedFilesMessage(message) {
    try {
      switch (message.type) {
        case "staged-files:ping":
          return { ok: true };
        case "staged-files:put":
          await putFile(message);
          return { ok: true };
        case "staged-files:delete":
          await deleteFile(message);
          return { ok: true };
        case "staged-files:list": {
          const files = await listFiles(message.assignmentKey);
          return { ok: true, files };
        }
        case "staged-files:clear-assignment":
          await clearAssignment(message.assignmentKey);
          return { ok: true };
        case "staged-files:reconcile":
          await reconcile(message);
          return { ok: true };
        case "staged-files:get-blob": {
          const blob = await getBlob(message);
          if (!blob) {
            return { ok: false, error: "staged file not found" };
          }
          return { ok: true, blobBase64: blob.blobBase64, mimeType: blob.mimeType };
        }
        default:
          return { ok: false, error: "unknown message" };
      }
    } catch (err) {
      return { ok: false, error: err instanceof Error ? err.message : "idb error" };
    }
  }

  // src/background.ts
  function isStagedFilesMessage(message) {
    return typeof message === "object" && message !== null && "type" in message && typeof message.type === "string" && message.type.startsWith("staged-files:");
  }
  chrome.runtime.onMessage.addListener((message, _sender, sendResponse) => {
    if (isStagedFilesMessage(message)) {
      void handleStagedFilesMessage(message).then(sendResponse);
      return true;
    }
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
