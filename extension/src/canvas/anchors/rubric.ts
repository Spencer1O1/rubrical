/**
 * Canvas rubric anchors (A2 traditional + classic table + long-description modal).
 */
import type { CanvasAnchor } from "./types";
import { discussionIds } from "./discussion";
import { testId, testIdContains, testIdStartsWith } from "../query";

export const rubricIds = {
  criterionRatings: "traditional-view-criterion-ratings",
  rubricAssessment: "rubric-assessment-traditional-view",
  longDescriptionClose: "long-description-close-button",
} as const;

/**
 * Cell-level selectors used inside rubric row walks.
 * Dynamic criterion ids use prefix/contains wildcards (`testIdStartsWith`, `testIdContains`).
 */
export const rubricCellSelectors = {
  ratingCell: `${testIdStartsWith("traditional-criterion-")}${testIdContains("-ratings-")}`,
  ratingPoints: `${testIdStartsWith("traditional-criterion-")}${testIdContains("-ratings-")}${testIdContains("-points")}`,
  criterionMaxPointsLabel: `span[wrap="normal"][letter-spacing="normal"]:not(${testIdStartsWith("criterion-score")})`,
  flexItem: '[class*="flexItem"]',
  flexColumn: '[class*="flex-flex"]',
  classicRatingTitle: ".rating-description-title, .header, strong",
  classicRatingDescription: ".description, .details, p",
  classicRatingPoints: ".points",
  classicRatingBox: ":scope > .rating",
  classicNestedRating: ":scope > td.rating, :scope > td",
} as const;

export const rubric = {
  present: {
    a2: [
      testId(rubricIds.criterionRatings),
      testId(discussionIds.gradedDiscussionInfo),
      testId(discussionIds.rubricAssessmentTray),
    ],
    classic: [
      "#assignment_show table.rubric",
      ".rubric_container table.rubric",
    ],
  },
  criterionRatings: {
    a2: [testId(rubricIds.criterionRatings)],
    classic: [],
  },
  rubricRoot: {
    a2: [testId(rubricIds.rubricAssessment)],
    classic: ["#assignment_show"],
  },
  classicTable: {
    a2: [],
    classic: [
      "#assignment_show table.rubrics",
      "#assignment_show table.rubric",
      ".rubric_container table.rubrics",
      ".rubric_container table.rubric",
    ],
  },
  longDescriptionModal: {
    a2: ['[role="dialog"][aria-label="Criterion Long Description"]'],
    classic: ['[role="dialog"][aria-label="Criterion Long Description"]'],
  },
  longDescriptionModalBody: {
    a2: ['[data-cid="ModalBody"]', '[class*="modalBody"]'],
    classic: ['[data-cid="ModalBody"]', '[class*="modalBody"]'],
  },
  longDescriptionClose: {
    a2: [
      `${testId(rubricIds.longDescriptionClose)} button`,
      testId(rubricIds.longDescriptionClose),
    ],
    classic: [],
  },
} as const satisfies Record<string, CanvasAnchor>;
