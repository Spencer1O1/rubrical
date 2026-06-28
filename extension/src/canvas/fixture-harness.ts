import { readFileSync, readdirSync, existsSync } from "node:fs";
import { join } from "node:path";
import { Window } from "happy-dom";
import type { CanvasPageEnv } from "./assignment-env";
import { composeFixtureHtml, type FixtureCompose } from "./fixture-compose";
import {
  discussion,
  discussionIds,
  draftUrl,
  instructions,
  instructionsIds,
  metadataAnchors,
  metadataIds,
  rubric,
  rubricIds,
  submission,
  submissionIds,
  submit,
  submitIds,
  upload,
  uploadIds,
} from "./anchors";
import { firstMatch, queryAnchor, testId } from "./query";

export const FIXTURES_ROOT = join(process.cwd(), "../fixtures");
export const EXPECTATIONS_DIR = join(FIXTURES_ROOT, "expectations");

export type FixtureExpectations = {
  fixture: string;
  compose?: FixtureCompose;
  anchors: Record<string, boolean>;
  extraction: {
    dueAt?: string;
    dueDateText?: string;
    pointsPossibleText?: string;
    submissionTypeText?: string;
    allowedSubmissionTypes?: string[];
  };
};

/** Every literal test id declared in anchor modules — must appear in at least one fixture. */
export const DECLARED_TEST_IDS: ReadonlyArray<{ module: string; id: string }> = [
  ...Object.entries(uploadIds).map(([key, id]) => ({ module: `uploadIds.${key}`, id })),
  ...Object.entries(submissionIds).map(([key, id]) => ({ module: `submissionIds.${key}`, id })),
  ...Object.entries(instructionsIds).map(([key, id]) => ({ module: `instructionsIds.${key}`, id })),
  ...Object.entries(metadataIds).map(([key, id]) => ({ module: `metadataIds.${key}`, id })),
  ...Object.entries(rubricIds).map(([key, id]) => ({ module: `rubricIds.${key}`, id })),
  ...Object.entries(submitIds).map(([key, id]) => ({ module: `submitIds.${key}`, id })),
  ...Object.entries(discussionIds).map(([key, id]) => ({ module: `discussionIds.${key}`, id })),
];

const ANCHOR_CHECKS: Record<string, () => boolean> = {
  dueDate: () => queryAnchor(metadataAnchors.dueDate) !== null,
  gradeDisplay: () => queryAnchor(metadataAnchors.points) !== null,
  submitButton: () => queryAnchor(submit.submitButton) !== null,
  discussionReply: () => queryAnchor(submit.discussionReply) !== null,
  discussionEditSubmit: () => queryAnchor(submit.discussionEditSubmit) !== null,
  discussionPrompt: () => firstMatch(discussion.prompt.a2) !== null,
  discussionDraftTextarea: () => {
    const editor = queryAnchor(discussion.editContainer);
    if (!editor) {
      return false;
    }
    return firstMatch(discussion.textareas.a2, editor) !== null;
  },
  discussionDraftEditorIframe: () => {
    const editor = queryAnchor(discussion.editContainer);
    if (!editor) {
      return false;
    }
    return firstMatch(discussion.editorIframes.a2, editor) !== null;
  },
  uploadTable: () => queryAnchor(upload.table) !== null,
  uploadPane: () => queryAnchor(upload.uploadPane) !== null,
  fileInputDrop: () => firstMatch(upload.fileInput.a2) !== null,
  urlInput: () => queryAnchor(draftUrl.urlInput) !== null,
  submissionTypeSelector: () => queryAnchor(submission.typeSelector) !== null,
  onlineUpload: () => queryAnchor(submission.typeBlocks.onlineUpload) !== null,
  onlineUrl: () => queryAnchor(submission.typeBlocks.onlineUrl) !== null,
  onlineTextEntry: () => queryAnchor(submission.typeBlocks.onlineTextEntry) !== null,
  textEditor: () => queryAnchor(submission.textEditor) !== null,
  textEntryTextarea: () => {
    const editor = queryAnchor(submission.textEditor);
    if (!editor) {
      return false;
    }
    return firstMatch(submission.textareas.a2, editor) !== null;
  },
  textEntryEditorIframe: () => {
    const editor = queryAnchor(submission.textEditor);
    if (!editor) {
      return false;
    }
    return firstMatch(submission.editorIframes.a2, editor) !== null;
  },
  criterionRatings: () => queryAnchor(rubric.criterionRatings) !== null,
  rubricAssessment: () => queryAnchor(rubric.rubricRoot) !== null,
  studentView: () => queryAnchor(metadataAnchors.studentView) !== null,
  assignmentDescription: () => queryAnchor(instructions.description) !== null,
  discussionRubricModal: () => queryAnchor(discussion.assignmentRubricModal) !== null,
  discussionPreviewRubric: () => queryAnchor(discussion.previewRubricButton) !== null,
  discussionRubricTray: () => queryAnchor(discussion.rubricAssessmentTray) !== null,
  discussionGradedRubricInfo: () => queryAnchor(discussion.gradedDiscussionInfo) !== null,
  discussionPostMenuTrigger: () => queryAnchor(discussion.postMenuTrigger) !== null,
  discussionShowRubricMenuItem: () => queryAnchor(discussion.showRubricMenuItem) !== null,
  discussionAttachButton: () => queryAnchor(discussion.attachButton) !== null,
  discussionAttachmentInput: () => {
    const editRoot = queryAnchor(discussion.editContainer);
    return Boolean(editRoot && queryAnchor(discussion.attachmentInput, editRoot));
  },
  discussionAttachment: () => {
    const editRoot = queryAnchor(discussion.editContainer);
    if (!editRoot) {
      return false;
    }
    return firstMatch(discussion.composerAttachmentDownloadLink.a2, editRoot) !== null;
  },
};

function readDiscussionEnvFromFixture(html: string): CanvasPageEnv | undefined {
  const match = html.match(/"can_attach_entries"\s*:\s*(true|false)/);
  if (!match) {
    return undefined;
  }
  return { can_attach_entries: match[1] === "true" };
}

let dom: Window | null = null;

function fixtureStem(name: string): string {
  return name.replace(/\.html$/, "");
}

export function loadFixtureHtml(name: string): string {
  const stem = fixtureStem(name);
  const expectations = loadExpectations(stem);
  if (expectations.compose) {
    return composeFixtureHtml(FIXTURES_ROOT, expectations.compose);
  }
  const path = join(FIXTURES_ROOT, `${stem}.html`);
  if (!existsSync(path)) {
    throw new Error(`fixture html not found: ${stem}.html`);
  }
  return readFileSync(path, "utf8");
}

export function loadExpectations(stem: string): FixtureExpectations {
  return JSON.parse(readFileSync(join(EXPECTATIONS_DIR, `${stem}.json`), "utf8")) as FixtureExpectations;
}

export function listFixtureCases(): Array<{ stem: string; expectations: FixtureExpectations }> {
  return readdirSync(EXPECTATIONS_DIR)
    .filter((name) => name.endsWith(".json"))
    .sort()
    .map((name) => {
      const stem = name.replace(/\.json$/, "");
      const expectations = loadExpectations(stem);
      return { stem, expectations };
    });
}

/** Parse a full Canvas HTML snapshot into happy-dom and attach globals. */
export function installFixture(html: string): void {
  dom = new Window();
  dom.document.open();
  dom.document.write(html);
  dom.document.close();

  const win = dom.window;
  const global = globalThis as typeof globalThis & {
    document: Document;
    window: Window & { ENV?: CanvasPageEnv };
  };

  global.document = win.document as unknown as Document;
  global.window = win as unknown as typeof global.window;
  global.window.ENV = readDiscussionEnvFromFixture(html);

  const assignmentMatch = html.match(/\/courses\/\d+\/assignments\/\d+/);
  const discussionMatch = html.match(/\/courses\/\d+\/discussion_topics\/\d+/);
  const pathname = discussionMatch?.[0] ?? assignmentMatch?.[0] ?? "/";
  Object.defineProperty(win, "location", {
    configurable: true,
    value: {
      ...win.location,
      pathname,
      href: `https://canvas.instructure.com${pathname}`,
    },
  });

  // happy-dom element constructors for `instanceof` checks in extraction code
  const winRecord = win as unknown as Record<string, unknown>;
  for (const name of [
    "HTMLElement",
    "HTMLAnchorElement",
    "HTMLTimeElement",
    "HTMLButtonElement",
    "HTMLInputElement",
    "HTMLTextAreaElement",
    "HTMLIFrameElement",
  ] as const) {
    if (winRecord[name]) {
      (globalThis as Record<string, unknown>)[name] = winRecord[name];
    }
  }
}

export function anchorPresent(key: string): boolean {
  const check = ANCHOR_CHECKS[key];
  if (!check) {
    throw new Error(`Unknown anchor expectation key: ${key}`);
  }
  return check();
}

export function fixtureContainsTestId(html: string, id: string): boolean {
  return html.includes(`data-testid="${id}"`);
}

export { testId };
