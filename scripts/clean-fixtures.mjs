#!/usr/bin/env node
/**
 * Prune Canvas page snapshots down to DOM contract tests need, then sanitize PII/URLs.
 *
 * Workflow:
 *   1. Save a live Canvas HTML snapshot to fixtures/my-case.html
 *   2. Add fixtures/expectations/my-case.json (copy a similar case and edit)
 *   3. pnpm clean:fixtures
 *   4. pnpm test:fixtures — fix expectations until green
 *
 * Without expectations, keeps every [data-testid] subtree (good first pass on new captures).
 *
 * Usage:
 *   pnpm clean:fixtures [--name ...] [--email ...] [--inst-host ...] [fixtures-dir] [file.html ...]
 *
 * Pass redaction values via flags, env vars (RUBRICAL_FIXTURE_*), or prompts when run interactively.
 */

import { readFileSync, writeFileSync, readdirSync, existsSync } from "node:fs";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";
import { Window } from "../extension/node_modules/happy-dom/lib/index.js";
import {
  parseCleanFixtureArgs,
  printCleanFixtureHelp,
  resolveRedactionOptions,
} from "./fixture-clean-cli.mjs";
import {
  assertClean,
  buildMinimalHtml,
  extractCanAttachEntries,
  extractCourseAssignmentPath,
  extractHtmlOpenTag,
  extractTitle,
  sanitizeBody,
} from "./fixture-sanitize-lib.mjs";
import { minifyFixtureDocument } from "./fixture-minify.mjs";

const __dirname = dirname(fileURLToPath(import.meta.url));
const cli = parseCleanFixtureArgs(process.argv);

if (cli.help) {
  printCleanFixtureHelp();
  process.exit(0);
}

const redaction = await resolveRedactionOptions(cli);
const FIXTURES_DIR = redaction.fixturesDir;
const ONLY_FILES = redaction.files;

/** Mirrors extension/src/canvas/fixture-harness.ts ANCHOR_CHECKS selectors. */
const ANCHOR_SELECTORS = {
  dueDate: '[data-testid="due-date"]',
  gradeDisplay: '[data-testid="grade-display"]',
  submitButton: '[data-testid="submit-button"], #submit-button',
  discussionReply: '[data-testid="discussion-topic-reply"], #discussion-reply-btn',
  discussionEditSubmit: '[data-testid="DiscussionEdit-submit"]',
  discussionPrompt: '[data-testid="discussion-root-entry-container"] [data-resource-type="discussion_topic.body"], [data-resource-type="discussion_topic.body"]',
  discussionDraftTextarea:
    '[data-testid="DiscussionEdit-container"] textarea, [data-testid="DiscussionEdit-container"] [data-testid="message-body"]',
  discussionDraftEditorIframe:
    '[data-testid="DiscussionEdit-container"] .tox-edit-area iframe, [data-testid="DiscussionEdit-container"] iframe.tox-edit-area__iframe',
  uploadTable: '[data-testid="uploaded_files_table"]',
  uploadPane: '[data-testid="upload-pane"]',
  fileInputDrop:
    'input[type="file"][data-testid="input-file-drop"], input[type="file"][data-testid*="file"]',
  urlInput: '[data-testid="url-input"]',
  submissionTypeSelector: '[data-testid="submission-type-selector"]',
  onlineUpload: '[data-testid="online_upload"]',
  onlineUrl: '[data-testid="online_url"]',
  onlineTextEntry: '[data-testid="online_text_entry"]',
  textEditor: '[data-testid="text-editor"]',
  textEntryTextarea: '[data-testid="text-editor"] textarea',
  textEntryEditorIframe:
    '[data-testid="text-editor"] .tox-edit-area iframe, [data-testid="text-editor"] iframe.tox-edit-area__iframe',
  criterionRatings:
    '[data-testid="traditional-view-criterion-ratings"], [data-testid="graded-discussion-info"], [data-testid="enhanced-rubric-assessment-tray"]',
  rubricAssessment:
    '[data-testid="rubric-assessment-traditional-view"], [data-testid="rubric-preview"]',
  studentView: '[data-testid="assignments-2-student-view"]',
  assignmentDescription: '[data-testid="assignments-2-assignment-description"]',
  discussionRubricModal: '[data-testid="assignment-rubric-modal"]',
  discussionPreviewRubric: '[data-testid="preview-assignment-rubric-button"]',
  discussionRubricTray: '[data-testid="enhanced-rubric-assessment-tray"]',
  discussionGradedRubricInfo: '[data-testid="graded-discussion-info"]',
  discussionPostMenuTrigger: '[data-testid="discussion-post-menu-trigger"]',
  discussionShowRubricMenuItem: '[data-testid="discussion-thread-menuitem-rubric"]',
  discussionAttachButton: '[data-testid="attach-btn"]',
  discussionAttachmentInput:
    '[data-testid="DiscussionEdit-container"] [data-testid="attachment-input"], [data-testid="attachment-input"]',
  discussionAttachment:
    '[data-testid="DiscussionEdit-container"] [data-testid="removable-item"], [data-testid="DiscussionEdit-container"] a[href*="/files/"][href*="download"], [data-testid="removable-item"]',
};

/** Expand a matched node to a larger subtree tests walk (tables, rubrics, composers). */
const EXPAND_TO_ANCESTOR = [
  {
    match: '[data-testid="traditional-view-criterion-ratings"]',
    ancestor: '[data-testid="rubric-assessment-traditional-view"], [data-testid="rubric-preview"]',
  },
  {
    match: '[data-testid="uploaded_files_table"] tbody tr',
    ancestor: '[data-testid="uploaded_files_table"]',
  },
  {
    match: '[data-testid="DiscussionEdit-container"]',
    ancestor: '[data-testid="DiscussionEdit-container"]',
  },
  {
    match: '[data-testid="enhanced-rubric-assessment-tray"]',
    ancestor: '[data-testid="enhanced-rubric-assessment-tray"]',
  },
  {
    match: '[data-testid="assignment-rubric-modal"]',
    ancestor: '[data-testid="assignment-rubric-modal"]',
  },
  {
    match: '[data-testid="discussion-thread-menuitem-rubric"]',
    ancestor: '[role="menu"], [data-testid="thread-actions-menu"]',
  },
  {
    match: '[data-testid="DiscussionEdit-submit"]',
    ancestor: '[data-testid="DiscussionEdit-container"]',
  },
];

/** Always keep when present — modal / upload chrome tests touch these indirectly. */
const KEEP_IF_PRESENT = [
  '[role="dialog"][aria-label="Criterion Long Description"]',
  '[data-testid="long-description-close-button"]',
  '[data-testid="attempt-tab"]',
  '[data-testid="assignment-2-student-content-tabs"]',
  '[data-testid="upload-box"]',
  '[data-testid="similarity-pledge"]',
  '[data-testid="student-footer"]',
  '[data-testid="discussion-root-entry-container"]',
  '[data-testid="discussion-topic-container"]',
  '[data-testid="assignments-2-assignment-toggle-details"]',
];

function loadExpectations(stem) {
  const path = join(FIXTURES_DIR, "expectations", `${stem}.json`);
  if (!existsSync(path)) {
    return null;
  }
  return JSON.parse(readFileSync(path, "utf8"));
}

function selectorsForExpectations(expectations) {
  if (!expectations?.anchors) {
    return null;
  }

  const entries = [];
  for (const [key, expected] of Object.entries(expectations.anchors)) {
    if (!expected) {
      continue;
    }
    const selector = ANCHOR_SELECTORS[key];
    if (!selector) {
      console.warn(`  warn: no selector for anchor key ${key}`);
      continue;
    }
    for (const part of selector.split(",").map((s) => s.trim())) {
      entries.push({ key, selector: part });
    }
  }
  return entries;
}

/** Anchors that need the full DOM subtree (tables, composers, rubrics). */
const SUBTREE_ANCHORS = new Set([
  "dueDate",
  "gradeDisplay",
  "submissionTypeSelector",
  "assignmentDescription",
  "discussionPrompt",
  "uploadTable",
  "uploadPane",
  "criterionRatings",
  "rubricAssessment",
  "textEditor",
  "textEntryTextarea",
  "textEntryEditorIframe",
  "discussionDraftTextarea",
  "discussionDraftEditorIframe",
  "discussionRubricModal",
  "discussionRubricTray",
  "discussionAttachment",
  "discussionAttachmentInput",
  "discussionPostMenuTrigger",
  "discussionShowRubricMenuItem",
  "discussionGradedRubricInfo",
  "discussionEditSubmit",
]);

function markAnchor(node, kept, fullSubtree) {
  let current = node;
  while (current) {
    kept.add(current);
    current = current.parentElement;
  }
  if (!fullSubtree) {
    return;
  }
  for (const descendant of node.querySelectorAll("*")) {
    kept.add(descendant);
  }
}

function expandMarkedRoots(document, kept) {
  for (const rule of EXPAND_TO_ANCESTOR) {
    for (const node of document.querySelectorAll(rule.match)) {
      const root =
        rule.ancestor === rule.match
          ? node
          : node.closest(rule.ancestor.split(",").map((s) => s.trim()).join(", "));
      if (root) {
        markAnchor(root, kept, true);
      }
    }
  }
}

function collapseInstructionBlocks(document) {
  for (const el of document.querySelectorAll(
    '[data-testid="assignments-2-assignment-description"]',
  )) {
    el.innerHTML = "<p>Assignment instructions (fixture).</p>";
  }
}

function pruneDocument(document, anchorEntries) {
  const kept = new Set();

  if (anchorEntries === null) {
    for (const el of document.querySelectorAll("[data-testid]")) {
      markAnchor(el, kept, true);
    }
  } else {
    for (const { key, selector } of anchorEntries) {
      for (const el of document.querySelectorAll(selector)) {
        markAnchor(el, kept, SUBTREE_ANCHORS.has(key));
      }
    }
  }

  for (const selector of KEEP_IF_PRESENT) {
    for (const el of document.querySelectorAll(selector)) {
      markAnchor(el, kept, true);
    }
  }

  expandMarkedRoots(document, kept);

  const body = document.body;
  if (!body) {
    throw new Error("missing body");
  }

  for (const child of [...body.querySelectorAll("*")]) {
    if (kept.has(child)) {
      continue;
    }
    if ([...kept].some((node) => node !== child && child.contains(node))) {
      continue;
    }
    child.remove();
  }

  for (const child of [...body.children]) {
    if (!kept.has(child) && ![...kept].some((node) => child.contains(node))) {
      child.remove();
    }
  }

  collapseInstructionBlocks(document);
  minifyFixtureDocument(document);
}

function extractBodyInnerHtml(html) {
  const window = new Window();
  window.document.open();
  window.document.write(html);
  window.document.close();
  return window.document.body;
}

function cleanFixture(html, filename, sanitizeOptions) {
  const stem = filename.replace(/\.html$/, "");
  const expectations = loadExpectations(stem);
  const anchorEntries = selectorsForExpectations(expectations);

  const title = extractTitle(html);
  const htmlOpenTag = extractHtmlOpenTag(html);
  const coursePath = extractCourseAssignmentPath(html);
  const canAttach = html.includes('"can_attach_entries"')
    ? extractCanAttachEntries(html)
    : undefined;

  const bodyNode = extractBodyInnerHtml(html);
  pruneDocument(bodyNode.ownerDocument, anchorEntries);

  const body = sanitizeBody(`<body>${bodyNode.innerHTML}</body>`, sanitizeOptions);
  return buildMinimalHtml({ htmlOpenTag, title, body, canAttachEntries: canAttach, coursePath });
}

const sanitizeOptions = {
  displayName: redaction.displayName,
  email: redaction.email,
  instHosts: redaction.instHosts,
};

function listTargets() {
  if (ONLY_FILES.length > 0) {
    return ONLY_FILES.map((name) => (name.includes("/") ? name : join(FIXTURES_DIR, name)));
  }
  return readdirSync(FIXTURES_DIR)
    .filter((name) => name.endsWith(".html"))
    .sort()
    .map((name) => join(FIXTURES_DIR, name));
}

const targets = listTargets();
let beforeBytes = 0;
let afterBytes = 0;

for (const path of targets) {
  const name = path.split("/").pop();
  const original = readFileSync(path, "utf8");
  beforeBytes += original.length;

  const cleaned = cleanFixture(original, name, sanitizeOptions);
  assertClean(cleaned, name, sanitizeOptions);
  writeFileSync(path, cleaned, "utf8");
  afterBytes += cleaned.length;

  console.log(
    `${name}: ${(original.length / 1024).toFixed(1)}KB → ${(cleaned.length / 1024).toFixed(1)}KB`,
  );
}

console.log(
  `\nCleaned ${targets.length} fixtures: ${(beforeBytes / 1024 / 1024).toFixed(2)}MB → ${(afterBytes / 1024 / 1024).toFixed(2)}MB`,
);
