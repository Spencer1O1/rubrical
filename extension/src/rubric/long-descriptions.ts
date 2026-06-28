import {
  activateControlWithoutScroll,
  beginLongDescriptionScrape,
} from "../activate-control-without-scroll";
import { rubric } from "../canvas/anchors";
import { combinedSelector, firstMatch, queryAnchor, queryAnchorAll } from "../canvas/query";
import { createLongDescriptionScrollDebugSession } from "../rubric-scroll-debug";
import { withLongDescriptionScrapeSession } from "../scrape-session";
import { untilDom } from "./modal-scrape";
import { isLongDescriptionControl, findA2Row, a2NonRatingCells } from "./shared";

const LONG_DESCRIPTION_MODAL_SELECTOR = combinedSelector(rubric.longDescriptionModal);

function findAllLongDescriptionButtonsOnPage(): HTMLButtonElement[] {
  return Array.from(document.querySelectorAll("button")).filter(
    (button): button is HTMLButtonElement =>
      button instanceof HTMLButtonElement && isLongDescriptionControl(button),
  );
}

function lockLongDescriptionButtons(): () => void {
  const previous = new Map<
    HTMLButtonElement,
    { pointerEvents: string; ariaDisabled: string | null }
  >();

  for (const button of findAllLongDescriptionButtonsOnPage()) {
    previous.set(button, {
      pointerEvents: button.style.pointerEvents,
      ariaDisabled: button.getAttribute("aria-disabled"),
    });
    button.dataset.rubricalLongDescriptionLocked = "true";
    button.setAttribute("aria-disabled", "true");
    button.style.pointerEvents = "none";
  }

  return () => {
    for (const [button, state] of previous) {
      if (!button.isConnected) {
        continue;
      }
      delete button.dataset.rubricalLongDescriptionLocked;
      button.style.pointerEvents = state.pointerEvents;
      if (state.ariaDisabled === null) {
        button.removeAttribute("aria-disabled");
      } else {
        button.setAttribute("aria-disabled", state.ariaDisabled);
      }
    }
  };
}

function unlockLongDescriptionButtonForClick(button: HTMLButtonElement): () => void {
  if (button.dataset.rubricalLongDescriptionLocked !== "true") {
    return () => {};
  }

  button.style.pointerEvents = "";
  button.removeAttribute("aria-disabled");
  return () => {
    if (!button.isConnected) {
      return;
    }
    button.dataset.rubricalLongDescriptionLocked = "true";
    button.setAttribute("aria-disabled", "true");
    button.style.pointerEvents = "none";
  };
}

function findCriterionLongDescriptionButtons(): HTMLButtonElement[] {
  const buttons: HTMLButtonElement[] = [];
  const seen = new Set<HTMLButtonElement>();

  for (const container of queryAnchorAll(rubric.criterionRatings)) {
    const row = findA2Row(container);
    if (!row) {
      continue;
    }
    const criterionCell = a2NonRatingCells(row)[0];
    if (!criterionCell) {
      continue;
    }
    const button = criterionCell.querySelector("button");
    if (
      button instanceof HTMLButtonElement &&
      isLongDescriptionControl(button) &&
      !seen.has(button)
    ) {
      seen.add(button);
      buttons.push(button);
    }
  }

  if (buttons.length > 0) {
    return buttons;
  }

  const rubricRoot = queryAnchor(rubric.rubricRoot);
  if (!rubricRoot) {
    return [];
  }

  for (const button of Array.from(rubricRoot.querySelectorAll("button"))) {
    if (
      button instanceof HTMLButtonElement &&
      isLongDescriptionControl(button) &&
      !seen.has(button)
    ) {
      seen.add(button);
      buttons.push(button);
    }
  }

  return buttons;
}

function readLongDescriptionModalText(modal: Element): string {
  const body = firstMatch([...rubric.longDescriptionModalBody.a2, ...rubric.longDescriptionModalBody.classic], modal);
  return (body?.textContent ?? "").replace(/\s+/g, " ").trim();
}

function closeLongDescriptionModal(modal: Element): void {
  const close = firstMatch([...rubric.longDescriptionClose.a2, ...rubric.longDescriptionClose.classic], modal);
  if (close instanceof HTMLElement) {
    activateControlWithoutScroll(close);
  }
}

function modalCount(): number {
  return document.querySelectorAll(LONG_DESCRIPTION_MODAL_SELECTOR).length;
}

export async function closeAllLongDescriptionModals(): Promise<void> {
  if (modalCount() === 0) {
    return;
  }
  for (const modal of Array.from(document.querySelectorAll(LONG_DESCRIPTION_MODAL_SELECTOR))) {
    closeLongDescriptionModal(modal);
  }
  await untilDom(() => modalCount() === 0);
}

function openAllLongDescriptionModals(
  buttons: HTMLButtonElement[],
  scrollDebug: ReturnType<typeof createLongDescriptionScrollDebugSession>,
): void {
  for (let i = 0; i < buttons.length; i++) {
    const button = buttons[i]!;
    const relockButton = unlockLongDescriptionButtonForClick(button);
    scrollDebug.logPhase(i + 1, "before-click");
    activateControlWithoutScroll(button);
    scrollDebug.logPhase(i + 1, "after-click");
    relockButton();
  }
}

/** Long-description batch scrape; caller must hold `beginLongDescriptionScrape()` when nested. */
export async function scrapeCriterionLongDescriptionsCore(): Promise<string[]> {
  const unlockButtons = lockLongDescriptionButtons();
  const scrollDebug = createLongDescriptionScrollDebugSession();
  try {
    await closeAllLongDescriptionModals();

    const buttons = findCriterionLongDescriptionButtons();
    if (buttons.length === 0) {
      return [];
    }

    openAllLongDescriptionModals(buttons, scrollDebug);
    await untilDom(() => modalCount() >= buttons.length);
    scrollDebug.logPhase(0, "modal-open");

    const modals = Array.from(document.querySelectorAll(LONG_DESCRIPTION_MODAL_SELECTOR));
    const descriptions = modals.slice(0, buttons.length).map(readLongDescriptionModalText);
    while (descriptions.length < buttons.length) {
      descriptions.push("");
    }

    await closeAllLongDescriptionModals();
    scrollDebug.logPhase(0, "after-modal-removed");

    return descriptions;
  } finally {
    await closeAllLongDescriptionModals();
    scrollDebug.stop();
    unlockButtons();
  }
}

export async function scrapeCriterionLongDescriptions(): Promise<string[]> {
  return withLongDescriptionScrapeSession(async () => {
    const endScrape = beginLongDescriptionScrape();
    try {
      return await scrapeCriterionLongDescriptionsCore();
    } finally {
      endScrape();
    }
  });
}

export function pageHasCriterionLongDescriptionButtons(): boolean {
  return findAllLongDescriptionButtonsOnPage().length > 0;
}

export function countCriterionLongDescriptionButtons(): number {
  return findCriterionLongDescriptionButtons().length;
}
