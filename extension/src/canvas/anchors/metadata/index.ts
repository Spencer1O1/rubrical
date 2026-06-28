import type { CanvasAnchor } from "../types";
import { dueDateAnchor } from "./due-date";
export { metadataClassic, metadataIds } from "./ids";
import { pointsAnchor } from "./points";
import {
  screenReaderContentAnchor,
  studentViewAnchor,
  submissionTypeScopeAnchor,
} from "./scope";
import { submissionTypeAnchor } from "./submission-type";

export { readDueFromTimeElement } from "./due-date";
export { readSubmissionTypesInScope } from "./submission-type";

export const metadataAnchors = {
  dueDate: dueDateAnchor,
  points: pointsAnchor,
  submissionType: submissionTypeAnchor,
  studentView: studentViewAnchor,
  screenReaderContent: screenReaderContentAnchor,
  submissionTypeScope: submissionTypeScopeAnchor,
} as const satisfies Record<string, CanvasAnchor>;
