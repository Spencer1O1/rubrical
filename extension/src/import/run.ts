import { postImport } from "../api";
import {
  afterSuccessfulImportClearStaging,
  pauseStagedFilesSync,
  reloadDraftManifest,
  resumeStagedFilesSync,
  uploadDiscussionAttachmentAfterImport,
} from "../staged-files";
import { syncStrictExtractionFromServer } from "../server-config";
import { extractLiveImportCapture } from "./live-capture";
import { buildImportPayload } from "./payload";
import { getOrPrefetchAssignmentContext } from "./prefetch";
import type { ImportPayload } from "./types";

export type ImportResult = {
  redirect?: string;
  title: string;
  base: string;
};

function assignmentIdFromImportResponse(data: {
  id?: number;
  redirect?: string;
}): number | null {
  if (typeof data.id === "number" && data.id > 0) {
    return data.id;
  }

  const match = data.redirect?.match(/\/assignments\/(\d+)/);
  if (!match?.[1]) {
    return null;
  }

  const parsed = Number.parseInt(match[1], 10);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
}

function importPayloadWithoutFileBytes(payload: ImportPayload): ImportPayload {
  return {
    ...payload,
    draftFiles: [],
  };
}

/** Merge cached assignment context with live submission capture and POST /imports. */
export async function runImportOnClick(pageType: string): Promise<ImportResult> {
  pauseStagedFilesSync();
  try {
    await syncStrictExtractionFromServer();
    if (pageType !== "discussion") {
      await reloadDraftManifest();
    }

    const [assignment, live] = await Promise.all([
      getOrPrefetchAssignmentContext(pageType),
      extractLiveImportCapture(),
    ]);

    const payload = buildImportPayload(assignment, live);
    const discussionAttachment = pageType === "discussion" ? (live.draftFiles[0] ?? null) : null;
    const importBody =
      pageType === "discussion" ? importPayloadWithoutFileBytes(payload) : payload;

    const { data, base } = await postImport(importBody);

    let resolvedBase = base;
    if (discussionAttachment) {
      const assignmentId = assignmentIdFromImportResponse(data);
      if (!assignmentId) {
        throw new Error("import succeeded but assignment id was missing");
      }
      resolvedBase = await uploadDiscussionAttachmentAfterImport(
        assignmentId,
        discussionAttachment,
      );
    }

    await afterSuccessfulImportClearStaging();

    return {
      redirect: data.redirect,
      title: assignment.title,
      base: resolvedBase,
    };
  } finally {
    resumeStagedFilesSync();
  }
}
