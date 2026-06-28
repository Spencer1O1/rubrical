import type { DraftFile, DraftFileRef } from "../../import/types";
import { listStagedFiles } from "../store";
import { stagingKeyFromPage, normalizeFileName } from "../staging-key";
import { fetchDraftManifestOnce, getDraftManifest } from "./manifest-client";
import { mergeRowAccessibility } from "./merge";
import { scanAssignmentUploadedRows } from "./canvas-rows";
import type { StagedUploadRecord } from "../../import/types";

export async function resolveAssignmentFilesForImport(): Promise<{
  draftFiles: DraftFile[];
  draftFileRefs: DraftFileRef[];
  stagedUploads: StagedUploadRecord[];
  skipped: string[];
}> {
  const stagingKey = stagingKeyFromPage();
  const canvasRows = scanAssignmentUploadedRows().map((row) => ({
    ...row,
    normalizedFileName: normalizeFileName(row.fileName),
  }));

  if (canvasRows.length === 0) {
    return { draftFiles: [], draftFileRefs: [], stagedUploads: [], skipped: [] };
  }

  await fetchDraftManifestOnce();
  const manifest = getDraftManifest();
  let staged: Awaited<ReturnType<typeof listStagedFiles>> = [];
  if (stagingKey) {
    try {
      staged = await listStagedFiles(stagingKey);
    } catch {
      staged = [];
    }
  }

  const accessibility = mergeRowAccessibility(canvasRows, staged, manifest.files);
  const draftFiles: DraftFile[] = [];
  const draftFileRefs: DraftFileRef[] = [];
  const stagedUploads: StagedUploadRecord[] = [];
  const skipped: string[] = [];

  for (let index = 0; index < accessibility.length; index++) {
    const row = canvasRows[index]!;
    const entry = accessibility[index]!;

    if (entry.state === "saved" && entry.serverFileId) {
      draftFileRefs.push({
        serverFileId: entry.serverFileId,
        fileName: row.fileName,
        canvasFileId: row.fileId ?? undefined,
        sortOrder: index,
      });
      continue;
    }

    if (entry.state === "staged" && stagingKey && entry.stagedRecord) {
      const stagedMatch = staged.find(
        (file) =>
          (entry.stagedRecord!.canvasFileId &&
            file.canvasFileId === entry.stagedRecord!.canvasFileId) ||
          (file.normalizedFileName === entry.stagedRecord!.normalizedFileName &&
            file.stagedAt === entry.stagedRecord!.stagedAt),
      );
      if (stagedMatch) {
        stagedUploads.push({
          ...stagedMatch,
          sortOrder: index,
        });
        continue;
      }
    }

    skipped.push(row.fileName);
  }

  return { draftFiles, draftFileRefs, stagedUploads, skipped };
}
