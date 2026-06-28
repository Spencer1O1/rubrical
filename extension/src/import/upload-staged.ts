import { postRubricalMultipartBlob } from "../api-multipart";
import { getStagedFileBlob } from "../staged-files/store";
import type { StagedUploadRecord } from "./types";

/** Multipart upload after JSON import — assignment files are never inlined in POST /imports. */
export async function uploadAssignmentStagedFilesAfterImport(
  assignmentId: number,
  records: StagedUploadRecord[],
): Promise<string> {
  let resolvedBase = "";

  for (const record of records) {
    const blob = await getStagedFileBlob(record);
    if (!blob) {
      throw new Error(`staged file missing from Canvas page storage: ${record.fileName}`);
    }

    const result = await postRubricalMultipartBlob(
      `/assignments/${assignmentId}/draft/upload`,
      blob,
      record.fileName,
      record.canvasFileId,
    );
    if (!result.ok) {
      throw new Error(result.error ?? `failed to upload ${record.fileName}`);
    }
    resolvedBase = result.base;
  }

  return resolvedBase;
}
