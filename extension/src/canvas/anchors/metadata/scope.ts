import { instructionsIds } from "../instructions";
import { submissionIds } from "../submission";
import type { CanvasAnchor } from "../types";
import { testId } from "../../query";
import { metadataClassic } from "./ids";

export const studentViewAnchor = {
  a2: [testId(instructionsIds.studentView)],
  classic: [metadataClassic.assignmentShow],
} as const satisfies CanvasAnchor;

export const screenReaderContentAnchor = {
  a2: [metadataClassic.screenReaderContent],
  classic: [metadataClassic.screenReaderContent],
} as const satisfies CanvasAnchor;

export const submissionTypeScopeAnchor = {
  a2: [testId(submissionIds.typeSelector)],
  classic: [metadataClassic.assignmentShow, metadataClassic.submissionForm],
} as const satisfies CanvasAnchor;
