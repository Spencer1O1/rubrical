import {
  canvasStorageKey,
  provisionalStorageKey,
} from "./staging-key";
import { base64ToArrayBuffer } from "./file-bytes";
import type { StagedFileRecord, StagedFilesMessage, StagedFilesResponse } from "./types";

const DB_NAME = "rubrical-staged-files";
const DB_VERSION = 1;
const STORE = "files";

type StoredRow = StagedFileRecord & {
  id: string;
  blob: Blob;
};

function openDb(): Promise<IDBDatabase> {
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

function rowId(
  assignmentKey: string,
  canvasFileId?: string,
  normalizedFileName?: string,
  stagedAt?: string,
): string {
  if (canvasFileId) {
    return `${assignmentKey}:${canvasStorageKey(canvasFileId)}`;
  }
  return `${assignmentKey}:${provisionalStorageKey(normalizedFileName ?? "", stagedAt ?? "")}`;
}

function listFilesInDb(db: IDBDatabase, assignmentKey: string): Promise<StoredRow[]> {
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE, "readonly");
    const request = tx.objectStore(STORE).index("assignmentKey").getAll(assignmentKey);
    request.onsuccess = () => resolve((request.result as StoredRow[]) ?? []);
    request.onerror = () => reject(request.error ?? new Error("idb list failed"));
  });
}

function getRowInDb(db: IDBDatabase, id: string): Promise<StoredRow | undefined> {
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE, "readonly");
    const request = tx.objectStore(STORE).get(id);
    request.onsuccess = () => resolve(request.result as StoredRow | undefined);
    request.onerror = () => reject(request.error ?? new Error("idb get failed"));
  });
}

async function putFile(message: Extract<StagedFilesMessage, { type: "staged-files:put" }>): Promise<void> {
  const encoded = message.blobBase64.trim();
  if (encoded === "") {
    throw new Error("staged file payload is empty");
  }

  const db = await openDb();
  const id = rowId(
    message.assignmentKey,
    message.canvasFileId,
    message.normalizedFileName,
    message.stagedAt,
  );
  const row: StoredRow = {
    id,
    assignmentKey: message.assignmentKey,
    canvasFileId: message.canvasFileId,
    fileName: message.fileName,
    normalizedFileName: message.normalizedFileName,
    stagedAt: message.stagedAt,
    mimeType: message.mimeType,
    blob: new Blob([base64ToArrayBuffer(message.blobBase64)], {
      type: message.mimeType || "application/octet-stream",
    }),
  };

  await new Promise<void>((resolve, reject) => {
    const tx = db.transaction(STORE, "readwrite");
    tx.objectStore(STORE).put(row);
    tx.oncomplete = () => resolve();
    tx.onerror = () => reject(tx.error ?? new Error("idb put failed"));
  });
  db.close();
}

async function deleteFile(message: Extract<StagedFilesMessage, { type: "staged-files:delete" }>): Promise<void> {
  const db = await openDb();
  const id = rowId(
    message.assignmentKey,
    message.canvasFileId,
    message.normalizedFileName,
    message.stagedAt,
  );

  await new Promise<void>((resolve, reject) => {
    const tx = db.transaction(STORE, "readwrite");
    tx.objectStore(STORE).delete(id);
    tx.oncomplete = () => resolve();
    tx.onerror = () => reject(tx.error ?? new Error("idb delete failed"));
  });
  db.close();
}

async function listFiles(assignmentKey: string): Promise<StagedFileRecord[]> {
  const db = await openDb();
  const rows = await listFilesInDb(db, assignmentKey);
  db.close();
  return rows.map(({ id: _id, blob: _blob, ...record }) => record);
}

async function clearAssignment(assignmentKey: string): Promise<void> {
  const db = await openDb();
  const rows = await listFilesInDb(db, assignmentKey);

  await new Promise<void>((resolve, reject) => {
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

async function reconcile(
  message: Extract<StagedFilesMessage, { type: "staged-files:reconcile" }>,
): Promise<void> {
  const db = await openDb();

  for (const promotion of message.promotions) {
    const provisionalId = rowId(
      message.assignmentKey,
      undefined,
      promotion.normalizedFileName,
      promotion.stagedAt,
    );
    const provisional = await getRowInDb(db, provisionalId);
    if (!provisional?.blob) {
      continue;
    }

    const promoted: StoredRow = {
      ...provisional,
      canvasFileId: promotion.canvasFileId,
      id: rowId(message.assignmentKey, promotion.canvasFileId),
    };

    await new Promise<void>((resolve, reject) => {
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

function blobToBase64(blob: Blob): Promise<string> {
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

async function getBlob(message: Extract<StagedFilesMessage, { type: "staged-files:get-blob" }>): Promise<{
  blobBase64: string;
  mimeType: string;
} | null> {
  const db = await openDb();
  const id = rowId(
    message.assignmentKey,
    message.canvasFileId,
    message.normalizedFileName,
    message.stagedAt,
  );
  const row = await getRowInDb(db, id);
  db.close();

  if (!row?.blob) {
    return null;
  }

  return {
    blobBase64: await blobToBase64(row.blob),
    mimeType: row.mimeType || row.blob.type || "application/octet-stream",
  };
}

export async function handleStagedFilesMessage(message: StagedFilesMessage): Promise<StagedFilesResponse> {
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
