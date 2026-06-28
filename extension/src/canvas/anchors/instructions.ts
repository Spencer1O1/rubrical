/**
 * Assignment instructions and context signal anchors.
 */
import { readInstructionEnv } from "../instruction-html";
import type { CanvasAnchor } from "./types";
import { discussion, discussionIds } from "./discussion";
import { testId } from "../query";

export const instructionsIds = {
  description: "assignments-2-assignment-description",
  toggleDetails: "assignments-2-assignment-toggle-details",
  studentView: "assignments-2-student-view",
} as const;

export const instructions = {
  description: {
    a2: [testId(instructionsIds.description), ...discussion.prompt.a2],
    classic: [
      "#assignment_show .description .user_content",
      "#assignment_show .user_content",
      ...discussion.prompt.classic,
    ],
    env: readInstructionEnv,
  },
  contextSignals: {
    a2: [
      testId(instructionsIds.description),
      testId(instructionsIds.studentView),
      testId(discussionIds.rootEntryContainer),
      '[data-resource-type="discussion_topic.body"]',
      "h1",
    ],
    classic: [
      "#assignment_show .user_content",
      "#assignment_show table.rubric",
      '[data-resource-type="discussion_topic.body"]',
    ],
  },
} as const satisfies Record<string, CanvasAnchor>;
