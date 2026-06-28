/**
 * Assignment title and course name anchors.
 */
import { envReaders } from "../assignment-env";
import { instructionsIds } from "./instructions";
import type { CanvasAnchor } from "./types";
import { testId } from "../query";
import { metadataClassic } from "./metadata/ids";

export const title = {
  heading: {
    a2: [
      `${testId(instructionsIds.studentView)} h1`,
      `${metadataClassic.assignmentShow} h1`,
    ],
    classic: [".assignment-title"],
    env: envReaders.name,
  },
  breadcrumb: {
    a2: [],
    classic: ["#breadcrumb li:last-child", ".ellipsible"],
  },
} as const satisfies Record<string, CanvasAnchor>;
