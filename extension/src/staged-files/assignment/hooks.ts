import { upload, uploadFileInputSelector } from "../../canvas/anchors";
import { queryAnchor } from "../../canvas/query";
import { rubricalDebugLog } from "../debug";
import { deleteStagedFile, listStagedFiles, putStagedFile, reconcileStagedFiles } from "../store";
import { normalizeFileName } from "../staging-key";
import { scanAssignmentUploadedRows } from "./canvas-rows";
import { newCanvasIdAssignments, snapshotFileIds } from "./id-map-diff";
import {
  flushPendingUploads,
  forgetPendingUpload,
  rememberPendingUpload,
  countPendingUploads,
} from "./pending-staging";
import { isStagedFileTooLarge, stagedFileSizeError } from "../size-limits";
import { findReconcilePromotions } from "./reconcile";

let workChain: Promise<void> = Promise.resolve();

function enqueueStagingWork(work: () => Promise<void>): Promise<void> {
  const next = workChain.then(work);
  workChain = next.catch(() => {});
  return next;
}

const FILE_INPUT_SELECTOR = uploadFileInputSelector();

function isSubmissionFileInput(target: EventTarget | null): target is HTMLInputElement {
  return target instanceof HTMLInputElement && target.matches(FILE_INPUT_SELECTOR);
}

type CapturedUpload = {
  file: File;
  stagedAt: string;
  normalizedFileName: string;
};

function captureUploads(files: File[]): CapturedUpload[] {
  return files.map((file) => ({
    file,
    stagedAt: new Date().toISOString(),
    normalizedFileName: normalizeFileName(file.name),
  }));
}

function canvasRowsForHooks() {
  return scanAssignmentUploadedRows().map((row) => ({
    ...row,
    normalizedFileName: normalizeFileName(row.fileName),
  }));
}

function hookRoot(): Element {
  return queryAnchor(upload.attemptRoot) ?? document.body;
}

/** True when the student can select new files (not the graded/submitted read-only view). */
export function isAssignmentFileUploadEditable(): boolean {
  const root = hookRoot();
  return Boolean(root.querySelector(uploadFileInputSelector()));
}

function canAttachFileHooks(): boolean {
  const root = hookRoot();
  return Boolean(
    root.querySelector(FILE_INPUT_SELECTOR) || queryAnchor(upload.table, root),
  );
}

async function stageCapturedFiles(stagingKey: string, uploads: CapturedUpload[]): Promise<void> {
  for (const upload of uploads) {
    if (isStagedFileTooLarge(upload.file.size)) {
      console.warn("[rubrical]", stagedFileSizeError(upload.file.name, upload.file.size));
      continue;
    }

    const record = {
      assignmentKey: stagingKey,
      fileName: upload.file.name,
      normalizedFileName: upload.normalizedFileName,
      stagedAt: upload.stagedAt,
      mimeType: upload.file.type || "application/octet-stream",
      bytes: null as ArrayBuffer | null,
    };

    try {
      await putStagedFile({
        assignmentKey: stagingKey,
        fileName: upload.file.name,
        normalizedFileName: upload.normalizedFileName,
        stagedAt: upload.stagedAt,
        mimeType: upload.file.type || "application/octet-stream",
        blob: upload.file,
      });
      forgetPendingUpload({
        assignmentKey: stagingKey,
        normalizedFileName: upload.normalizedFileName,
        stagedAt: upload.stagedAt,
      });
      rubricalDebugLog("staged file", {
        fileName: upload.file.name,
        stagedAt: upload.stagedAt,
        byteLength: upload.file.size,
      });
    } catch (err) {
      const buffer = await upload.file.arrayBuffer();
      rememberPendingUpload({
        ...record,
        bytes: buffer,
      });
      console.warn("[rubrical] staged file failed", {
        fileName: upload.file.name,
        stagedAt: upload.stagedAt,
        error: err instanceof Error ? err.message : String(err),
      });
      rubricalDebugLog("staged file failed", {
        fileName: upload.file.name,
        stagedAt: upload.stagedAt,
        error: err instanceof Error ? err.message : String(err),
      });
    }
  }
}

async function reconcileFromTable(stagingKey: string, fileIdSnapshot: (string | null)[]): Promise<(string | null)[]> {
  const rows = canvasRowsForHooks();
  const assignments = newCanvasIdAssignments(fileIdSnapshot, rows);
  const nextSnapshot = snapshotFileIds(rows);

  if (assignments.length === 0) {
    return nextSnapshot;
  }

  const staged = await listStagedFiles(stagingKey).catch(() => [] as Awaited<ReturnType<typeof listStagedFiles>>);
  const promotions = findReconcilePromotions(assignments, staged);
  if (promotions.length === 0) {
    return nextSnapshot;
  }

  await reconcileStagedFiles({ assignmentKey: stagingKey, promotions });
  rubricalDebugLog("reconciled staged files", { count: promotions.length });
  return nextSnapshot;
}

export async function reconcileStagingFromTable(stagingKey: string): Promise<void> {
  await reconcileFromTable(stagingKey, []);
}

export async function retryPendingStaging(stagingKey: string): Promise<number> {
  return flushPendingUploads(stagingKey);
}

/** Wait for in-flight staging work and pending retries before import capture. */
export async function awaitStagingIdle(stagingKey: string): Promise<void> {
  await workChain;
  await flushPendingUploads(stagingKey);
  const remaining = countPendingUploads(stagingKey);
  if (remaining > 0) {
    throw new Error(
      `${remaining} uploaded file${remaining === 1 ? "" : "s"} could not be staged — refresh the page and try again`,
    );
  }
}

export type CanvasHooksCallbacks = {
  onStaged: () => void;
  onRemoved: () => void;
  onCanvasIdAssigned: () => void;
};

export type CanvasHooksHandle = {
  disconnect: () => void;
};

export function connectCanvasFileHooks(
  stagingKey: string,
  callbacks: CanvasHooksCallbacks,
): CanvasHooksHandle | null {
  if (!canAttachFileHooks()) {
    return null;
  }

  const root = hookRoot();
  let fileIdSnapshot = snapshotFileIds(canvasRowsForHooks());

  const onInputChange = (event: Event): void => {
    if (!isSubmissionFileInput(event.target)) {
      return;
    }

    const files = Array.from(event.target.files ?? []);
    if (files.length === 0) {
      return;
    }

    const uploads = captureUploads(files);
    void enqueueStagingWork(async () => {
      await stageCapturedFiles(stagingKey, uploads);
    }).then(callbacks.onStaged);
  };

  const onTableClick = (event: Event): void => {
    const target = event.target;
    if (!(target instanceof Element)) {
      return;
    }

    const table = queryAnchor(upload.table, root);
    const button = target.closest("button[id]");
    if (!table || !button || !table.contains(button)) {
      return;
    }

    const canvasFileId = button.id.trim();
    if (!/^\d+$/.test(canvasFileId)) {
      return;
    }

    void enqueueStagingWork(async () => {
      await deleteStagedFile({ assignmentKey: stagingKey, canvasFileId });
      rubricalDebugLog("deleted staged canvas file", { fileId: canvasFileId });
    }).then(callbacks.onRemoved);
  };

  const onTableMutation = (): void => {
    void enqueueStagingWork(async () => {
      await flushPendingUploads(stagingKey);
      fileIdSnapshot = await reconcileFromTable(stagingKey, fileIdSnapshot);
    }).then(callbacks.onCanvasIdAssigned);
  };

  root.addEventListener("change", onInputChange, true);
  root.addEventListener("click", onTableClick, true);

  const table = queryAnchor(upload.table, root);
  const observer = new MutationObserver(onTableMutation);
  if (table) {
    observer.observe(table, { childList: true, subtree: true });
  }

  return {
    disconnect: () => {
      observer.disconnect();
      root.removeEventListener("change", onInputChange, true);
      root.removeEventListener("click", onTableClick, true);
    },
  };
}
