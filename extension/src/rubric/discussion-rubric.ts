import { activateControlWithoutScroll } from "../activate-control-without-scroll";
import { discussion, rubric } from "../canvas/anchors";
import { queryAnchor } from "../canvas/query";
import { closeAllLongDescriptionModals } from "./long-descriptions";
import { closeButtonIn, untilDom } from "./modal-scrape";

function isDiscussionPage(): boolean {
  return window.location.pathname.includes("/discussion_topics/");
}

function hasVisibleRubricCriteria(): boolean {
  return queryAnchor(rubric.criterionRatings) !== null;
}

export function findDiscussionPostMenuTrigger(): HTMLButtonElement | null {
  return queryAnchor<HTMLButtonElement>(discussion.postMenuTrigger);
}

export function findDiscussionShowRubricMenuItem(): HTMLElement | null {
  return queryAnchor(discussion.showRubricMenuItem);
}

export function discussionPageHasGradedRubric(): boolean {
  return (
    queryAnchor(discussion.gradedDiscussionInfo) !== null &&
    queryAnchor(discussion.postMenuTrigger) !== null
  );
}

function isDiscussionPostMenuOpen(): boolean {
  return findDiscussionShowRubricMenuItem() !== null;
}

async function closeDiscussionPostMenu(): Promise<void> {
  if (!isDiscussionPostMenuOpen()) {
    return;
  }

  const trigger = findDiscussionPostMenuTrigger();
  if (trigger) {
    activateControlWithoutScroll(trigger);
    await untilDom(() => !isDiscussionPostMenuOpen());
    return;
  }

  document.dispatchEvent(
    new KeyboardEvent("keydown", { key: "Escape", bubbles: true, cancelable: true }),
  );
  await untilDom(() => !isDiscussionPostMenuOpen());
}

export type DiscussionRubricOpenResult = {
  criteriaVisible: boolean;
  openedByUs: boolean;
};

async function dismissDiscussionOverlay(
  root: Element | null,
  isGone: () => boolean,
): Promise<void> {
  if (!root) {
    return;
  }

  const close = closeButtonIn(root);
  if (close) {
    activateControlWithoutScroll(close);
    await untilDom(isGone);
    if (isGone()) {
      return;
    }
  }

  document.dispatchEvent(
    new KeyboardEvent("keydown", { key: "Escape", bubbles: true, cancelable: true }),
  );
  await untilDom(isGone);
}

async function openAssignmentRubricModal(): Promise<boolean> {
  if (queryAnchor(discussion.assignmentRubricModal)) {
    await closeDiscussionPostMenu();
    return true;
  }

  const menuTrigger = findDiscussionPostMenuTrigger();
  if (!menuTrigger) {
    return false;
  }

  activateControlWithoutScroll(menuTrigger);
  await untilDom(() => findDiscussionShowRubricMenuItem() !== null);

  const showRubric = findDiscussionShowRubricMenuItem();
  if (!showRubric) {
    return false;
  }

  activateControlWithoutScroll(showRubric);
  await untilDom(() => queryAnchor(discussion.assignmentRubricModal) !== null);
  await closeDiscussionPostMenu();
  return queryAnchor(discussion.assignmentRubricModal) !== null;
}

async function openRubricAssessmentTray(): Promise<boolean> {
  if (hasVisibleRubricCriteria()) {
    return true;
  }

  const preview = queryAnchor<HTMLButtonElement>(discussion.previewRubricButton);
  if (!preview) {
    return false;
  }

  activateControlWithoutScroll(preview);
  await untilDom(() => hasVisibleRubricCriteria());
  return hasVisibleRubricCriteria();
}

/** Opens discussion rubric tray; caller holds scrape lock for the full extract pass. */
export async function ensureDiscussionRubricVisible(): Promise<DiscussionRubricOpenResult> {
  if (!isDiscussionPage()) {
    return { criteriaVisible: false, openedByUs: false };
  }

  if (hasVisibleRubricCriteria()) {
    return { criteriaVisible: true, openedByUs: false };
  }

  if (!discussionPageHasGradedRubric()) {
    return { criteriaVisible: false, openedByUs: false };
  }

  if (!(await openAssignmentRubricModal())) {
    return { criteriaVisible: false, openedByUs: false };
  }

  const criteriaVisible = await openRubricAssessmentTray();
  return { criteriaVisible, openedByUs: true };
}

/** Tear down every UI layer opened for discussion rubric extraction. */
export async function closeDiscussionRubricUI(): Promise<void> {
  await closeAllLongDescriptionModals();

  await dismissDiscussionOverlay(queryAnchor(discussion.rubricAssessmentTray), () =>
    queryAnchor(discussion.rubricAssessmentTray) === null,
  );

  await dismissDiscussionOverlay(queryAnchor(discussion.assignmentRubricModal), () =>
    queryAnchor(discussion.assignmentRubricModal) === null,
  );

  await closeDiscussionPostMenu();
}

export { hasVisibleRubricCriteria, isDiscussionPage };
