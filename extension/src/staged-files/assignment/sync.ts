import { rubricalDebugLog } from "../debug";
import { clearStagedAssignment, listStagedFiles, pingStagedFilesServiceWorker } from "../messages";
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
import { clearPendingUploads } from "./pending-staging";

let hooksHandle: { disconnect: () => void } | null = null;
let hooksStagingKey: string | null = null;
let bootedStagingKey: string | null = null;
let lastPaintedSignature = "";

function canvasRowsForMerge() {
  return scanAssignmentUploadedRows().map((row) => ({
    ...row,
    normalizedFileName: normalizeFileName(row.fileName),
  }));
}

async function paintIndicators(): Promise<void> {
  const stagingKey = stagingKeyFromPage();
  if (!stagingKey) {
    return;
  }

  const staged = await listStagedFiles(stagingKey);
  const manifest = getDraftManifest();
  const merged = mergeRowAccessibility(canvasRowsForMerge(), staged, manifest.files);
  const signature = accessibilitySignature(merged);

  if (signature === lastPaintedSignature) {
    return;
  }
  lastPaintedSignature = signature;

  if (merged.every((row) => row.state !== "inaccessible")) {
    clearUploadedFileIndicators();
    return;
  }

  decorateUploadedFileIndicators(merged);
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
  hooksStagingKey = stagingKey;
}

export function disconnectAssignmentStaging(): void {
  disconnectHooks();
  bootedStagingKey = null;
  lastPaintedSignature = "";
}

export function clearAssignmentIndicators(): void {
  clearUploadedFileIndicators();
  lastPaintedSignature = "";
}

export async function syncAssignmentStaging(stagingKey: string): Promise<void> {
  const stagingReady = await pingStagedFilesServiceWorker();
  if (!stagingReady) {
    rubricalDebugLog("service worker unavailable", { stagingKey });
    return;
  }

  ensureHooksConnected(stagingKey);
  await retryPendingStaging(stagingKey);

  if (bootedStagingKey !== stagingKey) {
    bootedStagingKey = stagingKey;
    await fetchDraftManifestOnce();
    await reconcileStagingFromTable(stagingKey);
    rubricalDebugLog("staged files sync started", { stagingKey });
  }

  await paintIndicators();
}

export async function refreshAssignmentIndicators(): Promise<void> {
  if (!stagingKeyFromPage()) {
    return;
  }

  const manifest = await reloadDraftManifest();
  const stagingKey = stagingKeyFromPage();
  if (stagingKey && manifest.files.length === 0) {
    await clearStagedAssignment(stagingKey);
    clearPendingUploads(stagingKey);
  }
  lastPaintedSignature = "";
  await paintIndicators();
  rubricalDebugLog("repainted draft file indicators", {
    manifestFileCount: manifest.files.length,
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
