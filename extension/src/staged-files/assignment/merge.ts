import type { DraftManifestFile, RowAccessibility, StagedFileRecord } from "../types";

export type CanvasRowForMerge = {
  fileName: string;
  normalizedFileName: string;
  fileId: string | null;
};

function matchManifestRow(
  row: CanvasRowForMerge,
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

  const candidates = manifest.filter(
    (file) =>
      !usedServerIds.has(file.serverFileId) &&
      file.fileName.trim().toLowerCase().replace(/\s+/g, " ") === row.normalizedFileName,
  );

  if (candidates.length === 0) {
    return null;
  }
  if (candidates.length === 1) {
    return candidates[0]!;
  }

  return null;
}

function matchStagedRow(
  row: CanvasRowForMerge,
  staged: StagedFileRecord[],
  usedProvisional: Set<string>,
): StagedFileRecord | null {
  if (row.fileId) {
    const byId = staged.find((entry) => entry.canvasFileId === row.fileId);
    if (byId) {
      return byId;
    }
  }

  const candidates = staged.filter((entry) => {
    if (entry.normalizedFileName !== row.normalizedFileName) {
      return false;
    }
    const key = `${entry.normalizedFileName}:${entry.stagedAt}`;
    return !usedProvisional.has(key);
  });

  if (candidates.length === 0) {
    return null;
  }

  const chosen =
    candidates.length === 1
      ? candidates[0]!
      : candidates.sort(
          (a, b) => new Date(a.stagedAt).getTime() - new Date(b.stagedAt).getTime(),
        )[0]!;

  usedProvisional.add(`${chosen.normalizedFileName}:${chosen.stagedAt}`);
  return chosen;
}

export function mergeRowAccessibility(
  canvasRows: CanvasRowForMerge[],
  staged: StagedFileRecord[],
  manifest: DraftManifestFile[],
): RowAccessibility[] {
  const usedServerIds = new Set<number>();
  const usedProvisional = new Set<string>();

  return canvasRows.map((row) => {
    const stagedMatch = matchStagedRow(row, staged, usedProvisional);
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

    const manifestMatch = matchManifestRow(row, manifest, usedServerIds);
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
