import { mergeSubmissionTypesAtClick } from "../metadata";
import type { CachedAssignmentContext, ImportPayload, LiveImportCapture } from "./types";

export function buildImportPayload(
  assignment: CachedAssignmentContext,
  live: LiveImportCapture,
): ImportPayload {
  return {
    sourceUrl: assignment.sourceUrl,
    pageType: assignment.pageType,
    title: assignment.title,
    visibleText: live.visibleText,
    instructionsText: assignment.instructionsText,
    draftText: live.draftText,
    draftUrl: live.draftUrl,
    draftKind: live.draftKind,
    draftFiles: live.draftFiles,
    draftFileRefs: live.draftFileRefs,
    fileImportWarnings: live.fileImportWarnings,
    rubric: assignment.rubric,
    metadata: mergeSubmissionTypesAtClick(assignment.metadata),
    capturedAt: live.capturedAt,
  };
}
