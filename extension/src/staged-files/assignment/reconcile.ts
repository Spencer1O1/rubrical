import type { CanvasIdAssignment, ReconcilePromotion, StagedFileRecord } from "../types";

/** Match provisional staged entries to canvas rows that gained a file id. */
export function findReconcilePromotions(
  assignments: CanvasIdAssignment[],
  staged: StagedFileRecord[],
): ReconcilePromotion[] {
  const promotions: ReconcilePromotion[] = [];
  const matchedProvisional = new Set<string>();

  for (const assignment of assignments) {
    if (staged.some((entry) => entry.canvasFileId === assignment.fileId)) {
      continue;
    }

    const candidates = staged.filter(
      (entry) =>
        !entry.canvasFileId &&
        entry.normalizedFileName === assignment.normalizedFileName &&
        !matchedProvisional.has(`${entry.normalizedFileName}:${entry.stagedAt}`),
    );

    if (candidates.length === 0) {
      continue;
    }

    const chosen = candidates.sort(
      (a, b) => new Date(a.stagedAt).getTime() - new Date(b.stagedAt).getTime(),
    )[0]!;

    matchedProvisional.add(`${chosen.normalizedFileName}:${chosen.stagedAt}`);
    promotions.push({
      normalizedFileName: chosen.normalizedFileName,
      stagedAt: chosen.stagedAt,
      canvasFileId: assignment.fileId,
    });
  }

  return promotions;
}
