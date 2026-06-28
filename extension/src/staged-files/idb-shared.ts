import { canvasStorageKey, provisionalStorageKey } from "./staging-key";
import type { StagedFileRecord } from "./types";

export const STAGED_FILES_DB_NAME = "rubrical-staged-files";
export const STAGED_FILES_DB_VERSION = 1;
export const STAGED_FILES_STORE = "files";

export type StoredStagedFileRow = StagedFileRecord & {
  id: string;
  blob: Blob;
};

export function stagedFileRowId(
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

export function openStagedFilesDb(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(STAGED_FILES_DB_NAME, STAGED_FILES_DB_VERSION);
    request.onupgradeneeded = () => {
      const db = request.result;
      if (!db.objectStoreNames.contains(STAGED_FILES_STORE)) {
        const store = db.createObjectStore(STAGED_FILES_STORE, { keyPath: "id" });
        store.createIndex("assignmentKey", "assignmentKey", { unique: false });
      }
    };
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error ?? new Error("idb open failed"));
  });
}
