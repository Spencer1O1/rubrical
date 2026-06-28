import type { CanvasIdAssignment } from "../types";

export type FileIdRow = {
  normalizedFileName: string;
  fileId: string | null;
};

/** Per-row canvas file ids in upload table order. */
export function snapshotFileIds(rows: FileIdRow[]): (string | null)[] {
  return rows.map((row) => row.fileId);
}

/** Rows whose canvas file id changed or appeared since the last snapshot. */
export function newCanvasIdAssignments(
  previous: readonly (string | null)[],
  rows: FileIdRow[],
): CanvasIdAssignment[] {
  const assignments: CanvasIdAssignment[] = [];

  for (let rowIndex = 0; rowIndex < rows.length; rowIndex++) {
    const row = rows[rowIndex]!;
    if (!row.fileId) {
      continue;
    }

    const prev = rowIndex < previous.length ? previous[rowIndex] : null;
    if (prev !== row.fileId) {
      assignments.push({
        rowIndex,
        normalizedFileName: row.normalizedFileName,
        fileId: row.fileId,
      });
    }
  }

  return assignments;
}
