import { extractAllowedSubmissionTypes, extractSubmissionTypeText } from "../metadata";
import type { CachedAssignmentContext, ImportPayload, LiveImportCapture } from "./types";

export function buildImportPayload(
  assignment: CachedAssignmentContext,
  live: LiveImportCapture,
): ImportPayload {
  const metadata =
    assignment.pageType === "discussion"
      ? {
          ...assignment.metadata,
          allowedSubmissionTypes: extractAllowedSubmissionTypes(),
          submissionTypeText: extractSubmissionTypeText(),
        }
      : assignment.metadata;

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
    rubric: assignment.rubric,
    metadata,
    capturedAt: live.capturedAt,
  };
}
