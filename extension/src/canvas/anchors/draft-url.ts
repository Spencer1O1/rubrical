/**
 * Web URL draft input anchors.
 */
import { submissionIds } from "./submission";
import type { CanvasAnchor } from "./types";
import { testId } from "../query";

export const draftUrl = {
  urlInput: {
    a2: [testId(submissionIds.urlInput)],
    classic: ['input[type="url"]', 'input[name*="url"]', "#submission_url"],
  },
} as const satisfies Record<string, CanvasAnchor>;
