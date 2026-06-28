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
import { scanAssignmentUploadedRows } from "../staged-files/assignment/canvas-rows";
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

function lingeringFileWarning(draftKind: string): string[] {
  if (draftKind !== "text" && draftKind !== "url") {
    return [];
  }

  const uploadedRows = scanAssignmentUploadedRows();
  if (uploadedRows.length === 0) {
    return [];
  }

  const names = uploadedRows.map((row) => row.fileName).join(", ");
  return [
    `Canvas shows uploaded file${uploadedRows.length === 1 ? "" : "s"} (${names}) but the active submission type is ${draftKind === "url" ? "URL" : "text"} — those files will not be imported`,
  ];
}

function skippedFileWarnings(skipped: string[]): string[] {
  if (skipped.length === 0) {
    return [];
  }

  return [
    `Could not import ${skipped.length} uploaded file${skipped.length === 1 ? "" : "s"}: ${skipped.join(", ")}`,
  ];
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
      stagedUploads: [],
      fileImportWarnings: [],
      capturedAt: new Date().toISOString(),
    };
  }

  const draftKind = detectActiveSubmissionKind();

  let draftText = "";
  let draftUrl = "";
  let draftFiles: LiveImportCapture["draftFiles"] = [];
  let draftFileRefs: LiveImportCapture["draftFileRefs"] = [];
  let stagedUploads: LiveImportCapture["stagedUploads"] = [];
  const fileImportWarnings: string[] = [];

  switch (draftKind) {
    case "file": {
      const resolved = await resolveAssignmentFilesForImport();
      draftFiles = resolved.draftFiles;
      draftFileRefs = resolved.draftFileRefs;
      stagedUploads = resolved.stagedUploads;
      fileImportWarnings.push(...skippedFileWarnings(resolved.skipped));
      break;
    }
    case "url":
      draftUrl = extractDraftURL();
      break;
    default:
      draftText = extractDraftText();
      break;
  }

  fileImportWarnings.push(...lingeringFileWarning(draftKind));

  if (rubricalDebugEnabled()) {
    console.info("[rubrical] live capture", {
      draftKind,
      draftTextLength: draftText.length,
      draftUrl,
      draftFileCount: draftFiles.length,
      draftFileRefCount: draftFileRefs.length,
      draftFileNames: draftFiles.map((file) => file.fileName),
      fileImportWarnings,
    });
  }

  return {
    visibleText: extractVisibleText(),
    draftText,
    draftUrl,
    draftKind,
    draftFiles,
    draftFileRefs,
    stagedUploads,
    fileImportWarnings,
    capturedAt: new Date().toISOString(),
  };
}
