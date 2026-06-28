import { upload } from "../../canvas/anchors";
import { queryAnchor } from "../../canvas/query";
import { getSubmissionRoot } from "../../draft";

export type AssignmentUploadedRow = {
  fileName: string;
  fileId: string | null;
};

const fileNameSelectors = [
  ...upload.fileRow.fileName.a2,
  ...upload.fileRow.fileName.classic,
] as const;

const trashButtonSelectors = [
  ...upload.fileRow.trashButton.a2,
  ...upload.fileRow.trashButton.classic,
] as const;

function getFileSearchRoot(): Element {
  return queryAnchor(upload.attemptRoot) ?? getSubmissionRoot();
}

function fileNameFromRow(row: Element): string {
  for (const selector of fileNameSelectors) {
    const title = row.querySelector(selector)?.getAttribute("title")?.trim();
    if (title) {
      return title;
    }
  }
  return "";
}

function fileIdFromUploadedRow(row: Element): string | null {
  for (const selector of trashButtonSelectors) {
    for (const button of Array.from(row.querySelectorAll(selector))) {
      const id = button.id?.trim() ?? "";
      if (/^\d+$/.test(id)) {
        return id;
      }
    }
  }
  return null;
}

function uploadedTableRows(root: Element): AssignmentUploadedRow[] {
  const table = queryAnchor(upload.table, root);
  if (!table) {
    return [];
  }

  const rows: AssignmentUploadedRow[] = [];
  for (const row of Array.from(table.querySelectorAll("tbody tr"))) {
    const fileName = fileNameFromRow(row);
    if (!fileName) {
      continue;
    }

    rows.push({ fileName, fileId: fileIdFromUploadedRow(row) });
  }

  return rows;
}

/** Scan Canvas uploaded_files_table rows for staging indicators and import merge. */
export function scanAssignmentUploadedRows(
  root: Element = getFileSearchRoot(),
): AssignmentUploadedRow[] {
  return uploadedTableRows(root);
}
