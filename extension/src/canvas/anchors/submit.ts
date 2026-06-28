/**
 * Submit / reply button anchors for inline Rubrical button placement.
 */
import type { CanvasAnchor } from "./types";
import { discussionIds } from "./discussion";
import { testId } from "../query";

export const submitIds = {
  submitButton: "submit-button",
  discussionTopicReply: "discussion-topic-reply",
  discussionEditSubmit: "DiscussionEdit-submit",
} as const;

export const submit = {
  submitButton: {
    a2: [testId(submitIds.submitButton), "#submit-button"],
    classic: [],
  },
  /** Opens the main topic reply composer — not used for Rubrical placement. */
  discussionReply: {
    a2: [testId(submitIds.discussionTopicReply), "#discussion-reply-btn"],
    classic: [],
  },
  /** Reply submit control when the composer is open (fixtures/3-discussion-reply-open.html). */
  discussionEditSubmit: {
    a2: [testId(submitIds.discussionEditSubmit)],
    classic: [],
  },
} as const satisfies Record<string, CanvasAnchor>;

/** Submit anchors tried in order for findSubmitAnchor. */
export const submitAnchorOrder: CanvasAnchor[] = [
  submit.submitButton,
  submit.discussionEditSubmit,
];
