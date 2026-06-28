import type { DraftManifestFile, RowAccessibility, StagedFileRecord } from "../types";

export type CanvasRowForMerge = {
  fileName: string;
  normalizedFileName: string;
  fileId: string | null;
};

function normalizeManifestFileName(fileName: string): string {
  return fileName.trim().toLowerCase().replace(/\s+/g, " ");
}

function sameNameOccurrenceIndex(rows: CanvasRowForMerge[], rowIndex: number): number {
  const name = rows[rowIndex]!.normalizedFileName;
  let count = 0;
  for (let i = 0; i < rowIndex; i++) {
    if (rows[i]!.normalizedFileName === name) {
      count++;
    }
  }
  return count;
}

function provisionalKey(entry: StagedFileRecord): string {
  return `${entry.normalizedFileName}:${entry.stagedAt}`;
}

function matchManifestRow(
  row: CanvasRowForMerge,
  rowIndex: number,
  canvasRows: CanvasRowForMerge[],
  manifest: DraftManifestFile[],
  usedServerIds: Set<number>,
): DraftManifestFile | null {
  if (row.fileId) {
    const byCanvasId = manifest.find(
      (file) =>
        !usedServerIds.has(file.serverFileId) && file.canvasFileId === row.fileId,
    );
    if (byCanvasId) {
      return byCanvasId;
    }
  }

  const sortedManifest = [...manifest].sort(
    (left, right) =>
      left.uploadedAt.localeCompare(right.uploadedAt) ||
      left.serverFileId - right.serverFileId,
  );

  const unused = sortedManifest.filter((file) => !usedServerIds.has(file.serverFileId));

  const nameCandidates = unused.filter(
    (file) => normalizeManifestFileName(file.fileName) === row.normalizedFileName,
  );

  if (nameCandidates.length > 0) {
    const occurrence = sameNameOccurrenceIndex(canvasRows, rowIndex);
    const byName = nameCandidates[occurrence];
    if (byName) {
      return byName;
    }
  }

  const byRowIndex = sortedManifest[rowIndex];
  if (byRowIndex && !usedServerIds.has(byRowIndex.serverFileId)) {
    return byRowIndex;
  }

  return null;
}

function matchStagedRow(
  row: CanvasRowForMerge,
  rowIndex: number,
  canvasRows: CanvasRowForMerge[],
  staged: StagedFileRecord[],
  usedProvisional: Set<string>,
): StagedFileRecord | null {
  if (row.fileId) {
    const byId = staged.find((entry) => entry.canvasFileId === row.fileId);
    if (byId) {
      return byId;
    }
  }

  const sortedProvisional = staged
    .filter((entry) => !entry.canvasFileId)
    .sort(
      (left, right) => new Date(left.stagedAt).getTime() - new Date(right.stagedAt).getTime(),
    );

  const provisional = sortedProvisional.filter(
    (entry) => !usedProvisional.has(provisionalKey(entry)),
  );

  const nameCandidates = provisional.filter(
    (entry) => entry.normalizedFileName === row.normalizedFileName,
  );

  let chosen: StagedFileRecord | undefined;
  if (nameCandidates.length > 0) {
    chosen = nameCandidates[sameNameOccurrenceIndex(canvasRows, rowIndex)];
  }
  if (!chosen) {
    const byRowIndex = sortedProvisional[rowIndex];
    if (byRowIndex && !usedProvisional.has(provisionalKey(byRowIndex))) {
      chosen = byRowIndex;
    }
  }

  if (!chosen) {
    return null;
  }

  usedProvisional.add(provisionalKey(chosen));
  return chosen;
}

export function mergeRowAccessibility(
  canvasRows: CanvasRowForMerge[],
  staged: StagedFileRecord[],
  manifest: DraftManifestFile[],
): RowAccessibility[] {
  const usedServerIds = new Set<number>();
  const usedProvisional = new Set<string>();

  return canvasRows.map((row, rowIndex) => {
    const stagedMatch = matchStagedRow(row, rowIndex, canvasRows, staged, usedProvisional);
    if (stagedMatch) {
      return {
        fileName: row.fileName,
        fileId: row.fileId,
        state: "staged" as const,
        stagedRecord: {
          normalizedFileName: stagedMatch.normalizedFileName,
          stagedAt: stagedMatch.stagedAt,
          canvasFileId: stagedMatch.canvasFileId,
        },
      };
    }

    const manifestMatch = matchManifestRow(row, rowIndex, canvasRows, manifest, usedServerIds);
    if (manifestMatch) {
      usedServerIds.add(manifestMatch.serverFileId);
      return {
        fileName: row.fileName,
        fileId: row.fileId,
        state: "saved" as const,
        serverFileId: manifestMatch.serverFileId,
      };
    }

    return { fileName: row.fileName, fileId: row.fileId, state: "inaccessible" as const };
  });
}
