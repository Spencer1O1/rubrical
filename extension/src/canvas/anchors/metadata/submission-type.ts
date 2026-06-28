import { envReaders } from "../../assignment-env";
import { firstMatch, queryAnchor } from "../../query";
import { submission } from "../submission";
import type { CanvasAnchor } from "../types";
import { findLabeledText, humanizeSubmissionType } from "./helpers";
import { submissionTypeScopeAnchor } from "./scope";

const SUBMISSION_TYPE_BLOCK_ORDER = [
  { type: "online_text_entry", block: "onlineTextEntry" },
  { type: "online_upload", block: "onlineUpload" },
  { type: "online_url", block: "onlineUrl" },
] as const;

/** Canvas type ids visible inside the submission-type selector (no metadata test id on A2). */
export function readSubmissionTypesInScope(scope: ParentNode = document): string[] {
  return SUBMISSION_TYPE_BLOCK_ORDER.filter(
    ({ block }) => firstMatch(submission.typeBlocks[block].a2, scope) !== null,
  ).map(({ type }) => type);
}

function readA2SubmissionTypeText(): string {
  const scope = queryAnchor(submissionTypeScopeAnchor) ?? document.body;
  const types = readSubmissionTypesInScope(scope);
  if (types.length === 0) {
    return "";
  }
  return types.map(humanizeSubmissionType).join(", ");
}

function envSubmissionTypeText(): string {
  const types = envReaders.submissionTypes();
  if (types.length === 0) {
    return "";
  }
  return types.map(humanizeSubmissionType).join(", ");
}

export const submissionTypeAnchor = {
  a2: [],
  classic: [],
  readA2: readA2SubmissionTypeText,
  readClassic: () => findLabeledText("Submitting"),
  env: envSubmissionTypeText,
} as const satisfies CanvasAnchor;
