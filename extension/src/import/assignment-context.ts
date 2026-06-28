import { instructions, rubric } from "../canvas/anchors";
import { anyAnchorPresent } from "../canvas/query";
import { extractCourseName, extractInstructions, extractTitle } from "../extractor";
import { extractPrefetchAssignmentMetadata } from "../metadata";
import { extractRubricTable, pageHasCriterionLongDescriptionButtons } from "../rubric";
import { syncStrictExtractionFromServer } from "../server-config";
import { normalizeSourceUrl } from "../staged-files/normalize-source-url";
import type { CachedAssignmentContext } from "./types";

export function assignmentContextSignalsPresent(): boolean {
  return (
    anyAnchorPresent(instructions.contextSignals) || anyAnchorPresent(rubric.present)
  );
}

export function rubricPresentInDOM(): boolean {
  return anyAnchorPresent(rubric.present);
}

/** Instructions, rubric, metadata — not the student's draft submission. */
export async function extractAssignmentContext(pageType: string): Promise<CachedAssignmentContext> {
  await syncStrictExtractionFromServer();

  const hadLongDescriptionButtons = pageHasCriterionLongDescriptionButtons();
  const rubricTable = await extractRubricTable();

  return {
    sourceUrl: normalizeSourceUrl(window.location.href) || window.location.href,
    pageType,
    title: extractTitle(),
    instructionsText: extractInstructions(),
    rubric: rubricTable,
    metadata: extractPrefetchAssignmentMetadata(extractCourseName()),
    cachedAt: new Date().toISOString(),
    longDescriptionsFetched: hadLongDescriptionButtons,
  };
}
