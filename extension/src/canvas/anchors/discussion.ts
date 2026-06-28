/**
 * Canvas discussion page anchors (prompt, open reply composer).
 *
 * Verified in fixtures/discussion-prompt (prompt) and
 * fixtures/discussion-reply-open (composer + draft editor).
 *
 * Graded discussion rubrics: fixtures/discussion-menu-open (post menu),
 * fixtures/discussion-rubric-modal (details modal),
 * fixtures/discussion-rubric-tray (assessment tray + traditional view),
 * fixtures/discussion-attachment (optional reply attachment in composer).
 */
import { readCanAttachDiscussionEntries } from "../assignment-env";
import type { CanvasAnchor } from "./types";
import { firstMatch, queryAnchor, testId } from "../query";

export const discussionIds = {
  rootEntryContainer: "discussion-root-entry-container",
  topicContainer: "discussion-topic-container",
  editContainer: "DiscussionEdit-container",
  editSubmit: "DiscussionEdit-submit",
  messageBody: "message-body",
  gradedDiscussionInfo: "graded-discussion-info",
  postMenuTrigger: "discussion-post-menu-trigger",
  showRubricMenuItem: "discussion-thread-menuitem-rubric",
  assignmentRubricModal: "assignment-rubric-modal",
  previewRubricButton: "preview-assignment-rubric-button",
  rubricAssessmentTray: "enhanced-rubric-assessment-tray",
  attachButton: "attach-btn",
  attachmentInput: "attachment-input",
  attachmentItem: "removable-item",
} as const;

const DISCUSSION_TEXT_SUBMISSION_TYPE = "online_text_entry";
const DISCUSSION_UPLOAD_SUBMISSION_TYPE = "online_upload";

/** Whether reply attachments are allowed (attach control, staged file, or ENV fallback). */
export function discussionAttachmentsAllowed(): boolean {
  const editRoot = queryAnchor(discussion.editContainer);
  if (editRoot) {
    if (queryAnchor(discussion.attachButton, editRoot) !== null) {
      return true;
    }
    if (firstMatch(discussion.composerAttachmentDownloadLink.a2, editRoot) !== null) {
      return true;
    }
    if (queryAnchor(discussion.attachmentItem, editRoot) !== null) {
      return true;
    }
  }

  const fromEnv = readCanAttachDiscussionEntries();
  if (fromEnv !== undefined) {
    return fromEnv;
  }

  return false;
}

/** Canvas type ids for graded discussion replies (no submission-type tabs on page). */
export function discussionAllowedSubmissionTypes(): string[] {
  const types = [DISCUSSION_TEXT_SUBMISSION_TYPE];
  if (discussionAttachmentsAllowed()) {
    types.push(DISCUSSION_UPLOAD_SUBMISSION_TYPE);
  }
  return types;
}

export const discussion = {
  prompt: {
    a2: [
      '[data-resource-type="discussion_topic.body"]',
      `${testId(discussionIds.topicContainer)} [data-resource-type="discussion_topic.body"]`,
    ],
    classic: [".discussion_topic .user_content"],
  },
  editContainer: {
    a2: [testId(discussionIds.editContainer)],
    classic: [],
  },
  textareas: {
    a2: [
      `${testId(discussionIds.editContainer)} textarea`,
      `${testId(discussionIds.editContainer)} ${testId(discussionIds.messageBody)}`,
    ],
    classic: [],
  },
  editorIframes: {
    a2: [
      `${testId(discussionIds.editContainer)} .tox-edit-area iframe`,
      `${testId(discussionIds.editContainer)} iframe.tox-edit-area__iframe`,
    ],
    classic: [],
  },
  gradedDiscussionInfo: {
    a2: [testId(discussionIds.gradedDiscussionInfo)],
    classic: [],
  },
  postMenuTrigger: {
    a2: [testId(discussionIds.postMenuTrigger)],
    classic: [],
  },
  showRubricMenuItem: {
    a2: [testId(discussionIds.showRubricMenuItem)],
    classic: [],
  },
  assignmentRubricModal: {
    a2: [testId(discussionIds.assignmentRubricModal)],
    classic: ['[role="dialog"][aria-label="Assignment Rubric Details"]'],
  },
  previewRubricButton: {
    a2: [testId(discussionIds.previewRubricButton)],
    classic: [],
  },
  rubricAssessmentTray: {
    a2: [testId(discussionIds.rubricAssessmentTray)],
    classic: ['[role="dialog"][aria-label="Rubric Assessment Tray"]'],
  },
  attachButton: {
    a2: [
      `${testId(discussionIds.editContainer)} ${testId(discussionIds.attachButton)}`,
      testId(discussionIds.attachButton),
    ],
    classic: [],
  },
  attachmentInput: {
    a2: [
      `${testId(discussionIds.editContainer)} ${testId(discussionIds.attachmentInput)}`,
      testId(discussionIds.attachmentInput),
    ],
    classic: [],
  },
  composerAttachmentDownloadLink: {
    a2: [
      `${testId(discussionIds.editContainer)} a[href*="/files/"]`,
    ],
    classic: [],
  },
  attachmentItem: {
    a2: [
      `${testId(discussionIds.editContainer)} ${testId(discussionIds.attachmentItem)}`,
      testId(discussionIds.attachmentItem),
    ],
    classic: [],
  },
  attachmentDownloadLink: {
    a2: [
      `${testId(discussionIds.editContainer)} a[href*="/files/"][href*="download"]`,
      `${testId(discussionIds.editContainer)} ${testId(discussionIds.attachmentItem)} a[href*="/files/"][href*="download"]`,
      `${testId(discussionIds.attachmentItem)} a[href*="/files/"][href*="download"]`,
    ],
    classic: [],
  },
  attachmentFileName: {
    a2: [
      `${testId(discussionIds.editContainer)} a[href*="/files/"] [aria-hidden="true"] span[wrap="normal"]`,
      `${testId(discussionIds.attachmentItem)} a[href*="/files/"] [aria-hidden="true"] span[wrap="normal"]`,
      `a[href*="/files/"] [aria-hidden="true"] span[wrap="normal"]`,
    ],
    classic: [],
  },
} as const satisfies Record<string, CanvasAnchor>;

/** Hidden file picker in the reply composer (fixtures/discussion-reply-open). */
export function discussionAttachmentInputSelector(): string {
  return testId(discussionIds.attachmentInput);
}
