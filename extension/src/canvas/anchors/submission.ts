/**
 * Canvas submission area anchors (draft text, type tabs, editor).
 *
 * `online_text_entry` is the type-tab control only. The mounted RCE lives under
 * `text-editor` (verified in fixtures/assignment-text-tab).
 */
import { testId } from "../query";
import { uploadIds } from "./upload";

export const submissionIds = {
  typeSelector: "submission-type-selector",
  onlineUpload: "online_upload",
  onlineUrl: "online_url",
  onlineTextEntry: "online_text_entry",
  textEditor: "text-editor",
  urlInput: "url-input",
} as const;

export const submission = {
  root: {
    a2: [testId(uploadIds.studentTabs), testId(uploadIds.attemptTab)],
    classic: ["#assignment_show", ".submission_form", "form.submit_assignment"],
  },
  textEditor: {
    a2: [testId(submissionIds.textEditor)],
    classic: [],
  },
  textareas: {
    a2: [`${testId(submissionIds.textEditor)} textarea`],
    classic: [
      "textarea#submission_body",
      "textarea[name='submission[body]']",
      "textarea[name*='submission']",
      "textarea[id*='submission']",
    ],
  },
  editorIframes: {
    a2: [
      `${testId(submissionIds.textEditor)} .tox-edit-area iframe`,
      `${testId(submissionIds.textEditor)} iframe.tox-edit-area__iframe`,
    ],
    classic: [],
    extra: [
      ".tox-edit-area iframe",
      "iframe.tox-edit-area__iframe",
      "iframe[id*='tinymce']",
    ],
  },
  typeSelector: {
    a2: [testId(submissionIds.typeSelector)],
    classic: [],
  },
  typeBlocks: {
    onlineUpload: {
      a2: [testId(submissionIds.onlineUpload)],
      classic: [],
    },
    onlineUrl: {
      a2: [testId(submissionIds.onlineUrl)],
      classic: [],
    },
    onlineTextEntry: {
      a2: [testId(submissionIds.onlineTextEntry)],
      classic: [],
    },
  },
  pressedTabButtons: {
    a2: [],
    classic: [],
    extra: ["button[aria-pressed='true']"],
  },
  fileInputLoose: {
    a2: [],
    classic: ['input[type="file"]'],
  },
  urlInputLoose: {
    a2: [],
    classic: ['input[type="url"]:not([disabled])'],
  },
} as const;

export const submissionTypeTestIds = [
  { testId: submissionIds.onlineUpload, kind: "file" as const },
  { testId: submissionIds.onlineUrl, kind: "url" as const },
  { testId: submissionIds.onlineTextEntry, kind: "text" as const },
];
