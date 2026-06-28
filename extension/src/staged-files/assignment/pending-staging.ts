import { rubricalDebugLog } from "../debug";
import { putStagedFile } from "../store";
import { isStagedFileTooLarge, stagedFileSizeError } from "../size-limits";

export type PendingUpload = {
  assignmentKey: string;
  fileName: string;
  normalizedFileName: string;
  stagedAt: string;
  mimeType: string;
  bytes: ArrayBuffer;
};

const pendingByKey = new Map<string, PendingUpload>();

function pendingKey(upload: Pick<PendingUpload, "assignmentKey" | "normalizedFileName" | "stagedAt">): string {
  return `${upload.assignmentKey}:${upload.normalizedFileName}:${upload.stagedAt}`;
}

export function rememberPendingUpload(upload: PendingUpload): void {
  pendingByKey.set(pendingKey(upload), upload);
}

export function forgetPendingUpload(
  upload: Pick<PendingUpload, "assignmentKey" | "normalizedFileName" | "stagedAt">,
): void {
  pendingByKey.delete(pendingKey(upload));
}

export function clearPendingUploads(assignmentKey: string): void {
  for (const key of pendingByKey.keys()) {
    if (key.startsWith(`${assignmentKey}:`)) {
      pendingByKey.delete(key);
    }
  }
}

export function countPendingUploads(assignmentKey: string): number {
  return [...pendingByKey.values()].filter((entry) => entry.assignmentKey === assignmentKey).length;
}

export function getPendingUploads(assignmentKey: string): PendingUpload[] {
  return [...pendingByKey.values()].filter((entry) => entry.assignmentKey === assignmentKey);
}

export async function flushPendingUploads(assignmentKey: string): Promise<number> {
  let stored = 0;

  for (const upload of [...pendingByKey.values()].filter((entry) => entry.assignmentKey === assignmentKey)) {
    if (isStagedFileTooLarge(upload.bytes.byteLength)) {
      forgetPendingUpload(upload);
      console.warn("[rubrical]", stagedFileSizeError(upload.fileName, upload.bytes.byteLength));
      continue;
    }

    try {
      await putStagedFile({
        assignmentKey: upload.assignmentKey,
        fileName: upload.fileName,
        normalizedFileName: upload.normalizedFileName,
        stagedAt: upload.stagedAt,
        mimeType: upload.mimeType,
        blob: new Blob([upload.bytes], { type: upload.mimeType || "application/octet-stream" }),
      });
      forgetPendingUpload(upload);
      stored++;
      rubricalDebugLog("staged file", {
        fileName: upload.fileName,
        stagedAt: upload.stagedAt,
        retried: true,
      });
    } catch (err) {
      rubricalDebugLog("staged file failed", {
        fileName: upload.fileName,
        stagedAt: upload.stagedAt,
        error: err instanceof Error ? err.message : String(err),
      });
    }
  }

  return stored;
}
