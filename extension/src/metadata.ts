import { metadataAnchors, readDueFromTimeElement, readSubmissionTypesInScope } from "./canvas/anchors";
import { discussionAllowedSubmissionTypes } from "./canvas/anchors/discussion";
import { isDiscussionPage } from "./canvas/anchors/page";
import { humanizeSubmissionType } from "./canvas/anchors/metadata/helpers";
import { envReaders } from "./canvas/assignment-env";
import { extractAnchor, queryAnchor } from "./canvas/query";

export type AssignmentMetadata = {
  dueDateText: string;
  dueAt: string;
  pointsPossibleText: string;
  submissionTypeText: string;
  allowedSubmissionTypes: string[];
  courseName: string;
};

export function extractDueDateText(): string {
  return extractAnchor(metadataAnchors.dueDate);
}

export function extractDueAtISO(): string {
  const fromDOM = readDueFromTimeElement().dueAt;
  if (fromDOM) {
    return fromDOM;
  }
  return envReaders.dueAt();
}

export function extractPointsPossibleText(): string {
  return extractAnchor(metadataAnchors.points);
}

export function extractSubmissionTypeText(): string {
  if (isDiscussionPage()) {
    return discussionAllowedSubmissionTypes()
      .map(humanizeSubmissionType)
      .join(", ");
  }
  return extractAnchor(metadataAnchors.submissionType);
}

const DRAFT_CAPABLE_CANVAS_TYPE_ORDER = [
  "online_text_entry",
  "online_upload",
  "online_url",
] as const;

function submissionTypeScope(): Element {
  return queryAnchor(metadataAnchors.submissionTypeScope) ?? document.body;
}

function readAllowedSubmissionTypesFromDOM(): string[] {
  const scope = submissionTypeScope();
  return readSubmissionTypesInScope(scope);
}

function readAllowedSubmissionTypesFromEnv(): string[] {
  return envReaders
    .submissionTypes()
    .map((value) => value.trim().toLowerCase())
    .filter((value) =>
      (DRAFT_CAPABLE_CANVAS_TYPE_ORDER as readonly string[]).includes(value),
    );
}

export function extractAllowedSubmissionTypes(): string[] {
  if (isDiscussionPage()) {
    return discussionAllowedSubmissionTypes();
  }

  const fromDOM = readAllowedSubmissionTypesFromDOM();
  if (fromDOM.length > 0) {
    return fromDOM;
  }

  return readAllowedSubmissionTypesFromEnv();
}

export function extractAssignmentMetadata(courseName: string): AssignmentMetadata {
  return {
    dueDateText: extractDueDateText(),
    dueAt: extractDueAtISO(),
    pointsPossibleText: extractPointsPossibleText(),
    submissionTypeText: extractSubmissionTypeText(),
    allowedSubmissionTypes: extractAllowedSubmissionTypes(),
    courseName,
  };
}
