import type { DraftFile, DraftFileRef } from "../../import/types";
import { getStagedFilePayload, listStagedFiles } from "../messages";
import { stagingKeyFromPage, normalizeFileName } from "../staging-key";
import { fetchDraftManifestOnce, getDraftManifest } from "./manifest-client";
import { mergeRowAccessibility } from "./merge";
import { scanAssignmentUploadedRows } from "./canvas-rows";

export async function resolveAssignmentFilesForImport(): Promise<{
  draftFiles: DraftFile[];
  draftFileRefs: DraftFileRef[];
}> {
  const stagingKey = stagingKeyFromPage();
  const canvasRows = scanAssignmentUploadedRows().map((row) => ({
    ...row,
    normalizedFileName: normalizeFileName(row.fileName),
  }));

  if (canvasRows.length === 0) {
    return { draftFiles: [], draftFileRefs: [] };
  }

  await fetchDraftManifestOnce();
  const manifest = getDraftManifest();
  const staged = stagingKey ? await listStagedFiles(stagingKey) : [];

  const accessibility = mergeRowAccessibility(canvasRows, staged, manifest.files);
  const draftFiles: DraftFile[] = [];
  const draftFileRefs: DraftFileRef[] = [];

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

    if (entry.state !== "staged" || !stagingKey || !entry.stagedRecord) {
      continue;
    }

    const stagedMatch = staged.find(
      (file) =>
        (entry.stagedRecord!.canvasFileId &&
          file.canvasFileId === entry.stagedRecord!.canvasFileId) ||
        (file.normalizedFileName === entry.stagedRecord!.normalizedFileName &&
          file.stagedAt === entry.stagedRecord!.stagedAt),
    );
    if (!stagedMatch) {
      continue;
    }

    const payload = await getStagedFilePayload(stagingKey, stagedMatch);
    if (payload) {
      draftFiles.push({
        ...payload,
        fileName: row.fileName,
        canvasFileId: stagedMatch.canvasFileId ?? row.fileId ?? undefined,
        sortOrder: index,
      });
    }
  }

  return { draftFiles, draftFileRefs };
}
