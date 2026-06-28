import { discussion, discussionAttachmentInputSelector } from "../../canvas/anchors";
import { queryAnchor } from "../../canvas/query";
import { postRubricalMultipart } from "../../api-multipart";
import type { DraftFile } from "../../import/types";
import { arrayBufferToBase64, draftFileToBlob, mimeTypeForFileName } from "../file-bytes";
import { getStagedFilePayload, listStagedFiles, putStagedFileBytes } from "../store";
import { stagingKeyFromPage, normalizeFileName } from "../staging-key";
import {
  readDiscussionComposerAttachment,
  type DiscussionComposerAttachment,
} from "./composer";

async function readSessionStagedFile(
  composerAttachment: DiscussionComposerAttachment,
): Promise<DraftFile | null> {
  const stagingKey = stagingKeyFromPage();
  if (!stagingKey) {
    return null;
  }

  const staged = await listStagedFiles(stagingKey).catch(() => [] as Awaited<ReturnType<typeof listStagedFiles>>);
  if (staged.length === 0) {
    return null;
  }

  const record =
    staged.find((file) => file.canvasFileId === composerAttachment.canvasFileId) ??
    [...staged].sort((left, right) => right.stagedAt.localeCompare(left.stagedAt))[0]!;

  const payload = await getStagedFilePayload(stagingKey, record);
  if (!payload) {
    return null;
  }

  return {
    ...payload,
    canvasFileId: payload.canvasFileId ?? composerAttachment.canvasFileId,
    sortOrder: 0,
  };
}

async function readComposerInputFile(
  composerAttachment: DiscussionComposerAttachment,
): Promise<DraftFile | null> {
  const editRoot = queryAnchor(discussion.editContainer);
  const input = editRoot?.querySelector<HTMLInputElement>(discussionAttachmentInputSelector());
  const file = input?.files?.[0];
  if (!file) {
    return null;
  }

  const buffer = await file.arrayBuffer();
  return {
    fileName: file.name,
    mimeType: file.type || mimeTypeForFileName(file.name),
    contentBase64: arrayBufferToBase64(buffer),
    canvasFileId: composerAttachment.canvasFileId,
    sortOrder: 0,
  };
}

async function downloadAndStageComposerAttachment(
  attachment: DiscussionComposerAttachment,
): Promise<DraftFile> {
  let response: Response;
  try {
    response = await fetch(attachment.downloadUrl, { credentials: "include" });
  } catch (err) {
    const detail = err instanceof Error ? err.message : String(err);
    throw new Error(`Canvas attachment download failed: ${detail}`);
  }
  if (!response.ok) {
    throw new Error(`Canvas attachment download failed (HTTP ${response.status})`);
  }

  const buffer = await response.arrayBuffer();
  const mimeType =
    response.headers.get("content-type")?.split(";")[0]?.trim() ||
    mimeTypeForFileName(attachment.fileName);

  const stagingKey = stagingKeyFromPage();
  if (stagingKey) {
    const stagedAt = new Date().toISOString();
    await putStagedFileBytes({
      assignmentKey: stagingKey,
      fileName: attachment.fileName,
      normalizedFileName: normalizeFileName(attachment.fileName),
      stagedAt,
      mimeType,
      blobBytes: buffer,
      canvasFileId: attachment.canvasFileId,
    });
  }

  return {
    fileName: attachment.fileName,
    mimeType,
    contentBase64: arrayBufferToBase64(buffer),
    canvasFileId: attachment.canvasFileId,
    sortOrder: 0,
  };
}

/** Read the composer attachment from session IDB, input, or Canvas download. */
export async function resolveDiscussionAttachmentForImport(): Promise<DraftFile | null> {
  const composerAttachment = readDiscussionComposerAttachment();
  if (!composerAttachment) {
    return null;
  }

  return (
    (await readSessionStagedFile(composerAttachment)) ??
    (await readComposerInputFile(composerAttachment)) ??
    downloadAndStageComposerAttachment(composerAttachment)
  );
}

/** Multipart upload after JSON import — discussions strip file bytes from POST /imports. */
export async function uploadDiscussionAttachmentAfterImport(
  assignmentId: number,
  file: DraftFile,
): Promise<string> {
  const formData = new FormData();
  formData.append("draft_file", draftFileToBlob(file), file.fileName);
  if (file.canvasFileId) {
    formData.append("canvas_file_id", file.canvasFileId);
  }

  const result = await postRubricalMultipart(
    `/assignments/${assignmentId}/draft/discussion-attachment`,
    formData,
  );
  if (!result.ok) {
    throw new Error(result.error);
  }

  return result.base;
}
