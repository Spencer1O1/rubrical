import { arrayBufferToBase64 } from "./file-bytes";
import {
  openStagedFilesDb,
  stagedFileRowId,
  STAGED_FILES_STORE,
  type StoredStagedFileRow,
} from "./idb-shared";
import type { ReconcilePromotion, StagedFileRecord } from "./types";

type Row = StoredStagedFileRow;
type PutInput = StagedFileRecord & { blob: Blob };

const rowId = (k: string, canvasFileId?: string, name?: string, at?: string) =>
  stagedFileRowId(k, canvasFileId, name, at);

async function withDb<T>(mode: IDBTransactionMode, fn: (db: IDBDatabase) => Promise<T>): Promise<T> {
  const db = await openStagedFilesDb();
  try {
    return await fn(db);
  } finally {
    db.close();
  }
}

function tx<T>(db: IDBDatabase, mode: IDBTransactionMode, run: (store: IDBObjectStore) => IDBRequest<T>): Promise<T> {
  return new Promise((resolve, reject) => {
    const t = db.transaction(STAGED_FILES_STORE, mode);
    const req = run(t.objectStore(STAGED_FILES_STORE));
    req.onsuccess = () => resolve(req.result as T);
    t.onerror = () => reject(t.error ?? req.error ?? new Error("idb failed"));
  });
}

const toRecord = ({ id: _i, blob: _b, ...r }: Row): StagedFileRecord => r;

export async function putStagedFile(input: PutInput): Promise<void> {
  if (input.blob.size === 0) throw new Error("staged file payload is empty");
  await withDb("readwrite", async (db) => {
    const row: Row = { ...input, id: rowId(input.assignmentKey, input.canvasFileId, input.normalizedFileName, input.stagedAt) };
    await tx(db, "readwrite", (s) => s.put(row));
  });
}

export async function putStagedFileBytes(
  input: Omit<PutInput, "blob"> & { blobBytes: ArrayBuffer },
): Promise<void> {
  await putStagedFile({
    ...input,
    blob: new Blob([input.blobBytes], { type: input.mimeType || "application/octet-stream" }),
  });
}

export async function listStagedFiles(assignmentKey: string): Promise<StagedFileRecord[]> {
  return withDb("readonly", async (db) =>
    (await tx(db, "readonly", (s) => s.index("assignmentKey").getAll(assignmentKey)) as Row[]).map(toRecord),
  );
}

export async function deleteStagedFile(input: {
  assignmentKey: string;
  canvasFileId?: string;
  normalizedFileName?: string;
  stagedAt?: string;
}): Promise<void> {
  await withDb("readwrite", async (db) => {
    await tx(db, "readwrite", (s) => s.delete(rowId(input.assignmentKey, input.canvasFileId, input.normalizedFileName, input.stagedAt)));
  });
}

export async function clearStagedAssignment(assignmentKey: string): Promise<void> {
  await withDb("readwrite", async (db) => {
    const rows = await tx(db, "readonly", (s) => s.index("assignmentKey").getAll(assignmentKey)) as Row[];
    const t = db.transaction(STAGED_FILES_STORE, "readwrite");
    const store = t.objectStore(STAGED_FILES_STORE);
    for (const row of rows) store.delete(row.id);
    await new Promise<void>((ok, no) => { t.oncomplete = () => ok(); t.onerror = () => no(t.error); });
  });
}

export async function reconcileStagedFiles(input: {
  assignmentKey: string;
  promotions: ReconcilePromotion[];
}): Promise<void> {
  await withDb("readwrite", async (db) => {
    for (const p of input.promotions) {
      const from = rowId(input.assignmentKey, undefined, p.normalizedFileName, p.stagedAt);
      const row = await tx(db, "readonly", (s) => s.get(from)) as Row | undefined;
      if (!row?.blob) continue;
      const promoted: Row = { ...row, canvasFileId: p.canvasFileId, id: rowId(input.assignmentKey, p.canvasFileId) };
      const t = db.transaction(STAGED_FILES_STORE, "readwrite");
      const store = t.objectStore(STAGED_FILES_STORE);
      store.delete(from);
      store.put(promoted);
      await new Promise<void>((ok, no) => { t.oncomplete = () => ok(); t.onerror = () => no(t.error); });
    }
  });
}

export async function getStagedFileBlob(record: StagedFileRecord): Promise<Blob | null> {
  return withDb("readonly", async (db) => {
    const row = await tx(db, "readonly", (s) =>
      s.get(rowId(record.assignmentKey, record.canvasFileId, record.normalizedFileName, record.stagedAt)),
    ) as Row | undefined;
    return row?.blob ?? null;
  });
}

export async function getStagedFilePayload(assignmentKey: string, record: StagedFileRecord) {
  const blob = await getStagedFileBlob({ ...record, assignmentKey });
  if (!blob) return null;
  return {
    fileName: record.fileName,
    mimeType: record.mimeType || blob.type || "application/octet-stream",
    contentBase64: arrayBufferToBase64(await blob.arrayBuffer()),
    canvasFileId: record.canvasFileId,
  };
}
