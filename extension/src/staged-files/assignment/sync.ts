import { rubricalDebugLog } from "../debug";
import { clearStagedAssignment, listStagedFiles } from "../store";
import { normalizeFileName, stagingKeyFromPage } from "../staging-key";
import { accessibilitySignature } from "./accessibility-signature";
import { scanAssignmentUploadedRows } from "./canvas-rows";
import {
  connectCanvasFileHooks,
  reconcileStagingFromTable,
  retryPendingStaging,
} from "./hooks";
import {
  clearUploadedFileIndicators,
  decorateUploadedFileIndicators,
} from "./indicators";
import {
  fetchDraftManifestOnce,
  getDraftManifest,
  reloadDraftManifest,
} from "./manifest-client";
import { mergeRowAccessibility } from "./merge";
import { clearPendingUploads, getPendingUploads } from "./pending-staging";

let hooksHandle: { disconnect: () => void } | null = null;
let hooksStagingKey: string | null = null;
let bootedStagingKey: string | null = null;
let lastPaintedSignature = "";
let assignmentFileHooksUnavailable = false;

function canvasRowsForMerge() {
  return scanAssignmentUploadedRows().map((row) => ({
    ...row,
    normalizedFileName: normalizeFileName(row.fileName),
  }));
}

async function listStagedFilesSafe(stagingKey: string) {
  try {
    return await listStagedFiles(stagingKey);
  } catch {
    return [];
  }
}

async function maybeClearOrphanStaging(stagingKey: string): Promise<void> {
  const canvasRows = canvasRowsForMerge();
  const manifest = getDraftManifest();
  if (canvasRows.length === 0 && manifest.files.length === 0) {
    await clearStagedAssignment(stagingKey);
    clearPendingUploads(stagingKey);
    rubricalDebugLog("cleared orphan staged files", { stagingKey });
  }
}

async function paintIndicators(): Promise<void> {
  const stagingKey = stagingKeyFromPage();
  if (!stagingKey) {
    return;
  }

  const staged = await listStagedFilesSafe(stagingKey);
  const manifest = getDraftManifest();
  const pending = getPendingUploads(stagingKey);
  const merged = mergeRowAccessibility(canvasRowsForMerge(), staged, manifest.files, pending);
  const signature = accessibilitySignature(merged);

  if (signature === lastPaintedSignature) {
    return;
  }
  lastPaintedSignature = signature;

  if (merged.every((row) => row.state !== "inaccessible" && row.state !== "staging_failed")) {
    clearUploadedFileIndicators();
    return;
  }

  decorateUploadedFileIndicators(merged, { fileHooksUnavailable: assignmentFileHooksUnavailable });
}

function disconnectHooks(): void {
  hooksHandle?.disconnect();
  hooksHandle = null;
  hooksStagingKey = null;
}

function ensureHooksConnected(stagingKey: string): void {
  if (hooksStagingKey === stagingKey && hooksHandle) {
    return;
  }

  disconnectHooks();

  hooksHandle = connectCanvasFileHooks(stagingKey, {
    onStaged: () => {
      void paintIndicators();
    },
    onRemoved: () => {
      void paintIndicators();
    },
    onCanvasIdAssigned: () => {
      void paintIndicators();
    },
  });
  assignmentFileHooksUnavailable = hooksHandle === null;
  hooksStagingKey = stagingKey;
}

export function disconnectAssignmentStaging(): void {
  disconnectHooks();
  bootedStagingKey = null;
  lastPaintedSignature = "";
  assignmentFileHooksUnavailable = false;
}

export function clearAssignmentIndicators(): void {
  clearUploadedFileIndicators();
  lastPaintedSignature = "";
}

export async function syncAssignmentStaging(stagingKey: string): Promise<void> {
  ensureHooksConnected(stagingKey);
  await retryPendingStaging(stagingKey);

  if (bootedStagingKey !== stagingKey) {
    bootedStagingKey = stagingKey;
    await fetchDraftManifestOnce();
    await reconcileStagingFromTable(stagingKey);
    rubricalDebugLog("staged files sync started", { stagingKey });
  }

  await maybeClearOrphanStaging(stagingKey);
  await paintIndicators();
}

export async function refreshAssignmentIndicators(): Promise<void> {
  if (!stagingKeyFromPage()) {
    return;
  }

  await reloadDraftManifest();
  lastPaintedSignature = "";
  await paintIndicators();
  rubricalDebugLog("repainted draft file indicators", {
    manifestFileCount: getDraftManifest().files.length,
  });
}

export async function afterSuccessfulAssignmentImportClearStaging(): Promise<void> {
  const stagingKey = stagingKeyFromPage();
  if (!stagingKey) {
    return;
  }

  await clearStagedAssignment(stagingKey);
  clearPendingUploads(stagingKey);
  await refreshAssignmentIndicators();
}
