import type { RowAccessibility } from "../types";

/** Stable string for skipping indicator DOM updates when merge output is unchanged. */
export function accessibilitySignature(rows: RowAccessibility[]): string {
  return JSON.stringify(
    rows.map((row) => ({
      fileName: row.fileName,
      fileId: row.fileId,
      state: row.state,
      serverFileId: row.serverFileId ?? null,
      stagedAt: row.stagedRecord?.stagedAt ?? null,
    })),
  );
}
