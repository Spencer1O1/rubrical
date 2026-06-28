import { discussion, submission, submissionTypeTestIds, upload } from "./canvas/anchors";
import { firstMatch, queryAnchor, queryAnchorAll } from "./canvas/query";
import { getSubmissionRoot } from "./draft";
import { isStrictExtraction } from "./strict";

function readA2SelectedKind(root: Element): "text" | "file" | "url" | null {
  const selector = queryAnchor(submission.typeSelector, root) ?? root;
  const selected: Array<"text" | "file" | "url"> = [];

  for (const { kind } of submissionTypeTestIds) {
    const block = firstMatch(
      submission.typeBlocks[
        kind === "file" ? "onlineUpload" : kind === "url" ? "onlineUrl" : "onlineTextEntry"
      ].a2,
      selector,
    );
    const label = block?.textContent?.toLowerCase() ?? "";
    if (label.includes("currently selected")) {
      selected.push(kind);
    }
  }

  if (selected.length === 0) {
    return null;
  }

  for (const kind of ["text", "url", "file"] as const) {
    if (selected.includes(kind)) {
      return kind;
    }
  }

  return selected[0]!;
}

function hasUploadedFileRows(root: Element): boolean {
  const table = queryAnchor(upload.table, root);
  return Boolean(table?.querySelector("tbody tr"));
}

/** Which Canvas submission tab appears selected (text, file, or url). */
export function detectActiveSubmissionKind(): "text" | "file" | "url" {
  if (queryAnchor(discussion.editContainer)) {
    return "text";
  }

  const root = getSubmissionRoot();

  const a2Kind = readA2SelectedKind(root);
  if (a2Kind) {
    return a2Kind;
  }

  if (hasUploadedFileRows(root)) {
    return "file";
  }

  for (const button of queryAnchorAll(submission.pressedTabButtons, root)) {
    const label = button.textContent?.toLowerCase() ?? "";
    if (label.includes("upload") || label.includes("file")) {
      return "file";
    }
    if (label.includes("url") || label.includes("website") || label.includes("web")) {
      return "url";
    }
    if (label.includes("text")) {
      return "text";
    }
  }

  if (!isStrictExtraction()) {
    if (firstMatch(submission.fileInputLoose.classic, root)) {
      return "file";
    }
    if (firstMatch(submission.urlInputLoose.classic, root)) {
      return "url";
    }
  }

  return "text";
}
