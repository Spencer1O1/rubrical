import { page } from "../canvas/anchors";
import { isDiscussionPage } from "../canvas/anchors/page";
import { queryAnchor } from "../canvas/query";
import { extractDraftText } from "../draft";
import { extractDraftURL } from "../draft-url";
import { detectActiveSubmissionKind } from "../submission-kind";
import {
  resolveAssignmentFilesForImport,
  resolveDiscussionAttachmentForImport,
} from "../staged-files";
import type { LiveImportCapture } from "./types";

function extractVisibleText(): string {
  return queryAnchor(page.visibleTextRoot)?.textContent?.trim() ?? "";
}

function rubricalDebugEnabled(): boolean {
  try {
    return localStorage.getItem("rubrical_debug") === "1";
  } catch {
    return false;
  }
}

/** Extract draft, submission files, and other student-owned state at click time. */
export async function extractLiveImportCapture(): Promise<LiveImportCapture> {
  if (isDiscussionPage()) {
    const draftText = extractDraftText();
    const attachment = await resolveDiscussionAttachmentForImport();
    const draftFiles = attachment ? [attachment] : [];

    if (rubricalDebugEnabled()) {
      console.info("[rubrical] live capture (discussion)", {
        draftKind: "text",
        draftTextLength: draftText.length,
        draftFileCount: draftFiles.length,
        draftFileNames: draftFiles.map((file) => file.fileName),
      });
    }

    return {
      visibleText: extractVisibleText(),
      draftText,
      draftUrl: "",
      draftKind: "text",
      draftFiles,
      draftFileRefs: [],
      capturedAt: new Date().toISOString(),
    };
  }

  const draftKind = detectActiveSubmissionKind();

  let draftText = "";
  let draftUrl = "";
  let draftFiles: LiveImportCapture["draftFiles"] = [];
  let draftFileRefs: LiveImportCapture["draftFileRefs"] = [];

  switch (draftKind) {
    case "file": {
      const resolved = await resolveAssignmentFilesForImport();
      draftFiles = resolved.draftFiles;
      draftFileRefs = resolved.draftFileRefs;
      break;
    }
    case "url":
      draftUrl = extractDraftURL();
      break;
    default:
      draftText = extractDraftText();
      break;
  }

  if (rubricalDebugEnabled()) {
    console.info("[rubrical] live capture", {
      draftKind,
      draftTextLength: draftText.length,
      draftUrl,
      draftFileCount: draftFiles.length,
      draftFileRefCount: draftFileRefs.length,
      draftFileNames: draftFiles.map((file) => file.fileName),
    });
  }

  return {
    visibleText: extractVisibleText(),
    draftText,
    draftUrl,
    draftKind,
    draftFiles,
    draftFileRefs,
    capturedAt: new Date().toISOString(),
  };
}
