import type { CanvasIdAssignment, ReconcilePromotion, StagedFileRecord } from "../types";

function provisionalKey(entry: StagedFileRecord): string {
  return `${entry.normalizedFileName}:${entry.stagedAt}`;
}

/** Match provisional staged entries to canvas rows that gained a file id. */
export function findReconcilePromotions(
  assignments: CanvasIdAssignment[],
  staged: StagedFileRecord[],
): ReconcilePromotion[] {
  const promotions: ReconcilePromotion[] = [];
  const matchedProvisional = new Set<string>();

  const provisional = staged
    .filter((entry) => !entry.canvasFileId)
    .sort(
      (left, right) => new Date(left.stagedAt).getTime() - new Date(right.stagedAt).getTime(),
    );

  for (const assignment of assignments) {
    if (staged.some((entry) => entry.canvasFileId === assignment.fileId)) {
      continue;
    }

    const nameCandidates = provisional.filter(
      (entry) =>
        entry.normalizedFileName === assignment.normalizedFileName &&
        !matchedProvisional.has(provisionalKey(entry)),
    );

    const chosen =
      nameCandidates.length === 1
        ? nameCandidates[0]!
        : provisional[assignment.rowIndex];

    if (!chosen || chosen.canvasFileId || matchedProvisional.has(provisionalKey(chosen))) {
      continue;
    }

    matchedProvisional.add(provisionalKey(chosen));
    promotions.push({
      normalizedFileName: chosen.normalizedFileName,
      stagedAt: chosen.stagedAt,
      canvasFileId: assignment.fileId,
    });
  }

  return promotions;
}
