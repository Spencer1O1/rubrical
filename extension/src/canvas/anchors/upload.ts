/**
 * Canvas upload UI anchors.
 * a2: verified data-testid values from fixtures/README.md
 * classic: stable classic Canvas ids/classes
 * extra: best-effort fallbacks only
 */
import type { CanvasAnchor } from "./types";
import { testId, testIdContains } from "../query";

export const uploadIds = {
  table: "uploaded_files_table",
  fileInputDrop: "input-file-drop",
  uploadPane: "upload-pane",
  attemptTab: "attempt-tab",
  studentTabs: "assignment-2-student-content-tabs",
} as const;

export const upload = {
  table: {
    a2: [testId(uploadIds.table)],
    classic: [],
  },
  fileInput: {
    a2: [
      `input[type="file"]${testId(uploadIds.fileInputDrop)}`,
      `input[type="file"]${testIdContains("file")}`,
    ],
    classic: ['input[type="file"]'],
  },
  uploadPane: {
    a2: [testId(uploadIds.uploadPane)],
    classic: [],
  },
  attemptRoot: {
    a2: [testId(uploadIds.attemptTab), testId(uploadIds.studentTabs)],
    classic: ["#assignment_show", ".submission_form", "form.submit_assignment"],
  },
  fileRow: {
    fileName: {
      a2: ["span[title]"],
      classic: ["span[title]", "[title]"],
    },
    trashButton: {
      a2: ["button[id]"],
      classic: ["button[id]"],
    },
  },
} as const satisfies Record<string, CanvasAnchor | Record<string, CanvasAnchor>>;

/** Comma-joined selector for file input change events. */
export function uploadFileInputSelector(): string {
  return [...upload.fileInput.a2, ...upload.fileInput.classic].join(", ");
}
