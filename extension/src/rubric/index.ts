import {
  beginLongDescriptionScrape,
} from "../activate-control-without-scroll";
import { runTiers, queryAnchor } from "../canvas/query";
import { rubric } from "../canvas/anchors";
import { withLongDescriptionScrapeSession } from "../scrape-session";
import { extractA2TraditionalRubric } from "./a2";
import { extractClassicRubric } from "./classic";
import {
  closeDiscussionRubricUI,
  ensureDiscussionRubricVisible,
  isDiscussionPage,
} from "./discussion-rubric";
import {
  scrapeCriterionLongDescriptions,
  scrapeCriterionLongDescriptionsCore,
  pageHasCriterionLongDescriptionButtons,
} from "./long-descriptions";
import type { RubricTable } from "./types";

export type { RubricRating, RubricTable, RubricTableRow } from "./types";

/** A2 traditional view → classic table (strict: A2 only). */
export async function extractRubricTable(): Promise<RubricTable | null> {
  return withLongDescriptionScrapeSession(async () => {
    const endScrape = beginLongDescriptionScrape({ lockScroll: !isDiscussionPage() });
    let openedDiscussionRubricUI = false;

    try {
      if (isDiscussionPage() && !queryAnchor(rubric.criterionRatings)) {
        const openResult = await ensureDiscussionRubricVisible();
        openedDiscussionRubricUI = openResult.openedByUs;
      }

      const longDescriptions = await scrapeCriterionLongDescriptionsCore();
      return runTiers([
        () => extractA2TraditionalRubric(longDescriptions),
        () => extractClassicRubric(longDescriptions),
      ]);
    } finally {
      if (openedDiscussionRubricUI) {
        await closeDiscussionRubricUI();
      }
      endScrape();
    }
  });
}

export { pageHasCriterionLongDescriptionButtons, scrapeCriterionLongDescriptions };
