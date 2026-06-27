import {
  debounce,
  detectPageType,
  ensureInlineButton,
  isSupportedCanvasPage,
  needsPlacement,
} from "./injector";
import { postImport, RUBRICAL_API_BASES } from "./api";
import { RUBRICAL_BUTTON_LABEL } from "./labels";
import { openAssignmentModal } from "./modal";
import {
  extractCourseName,
  extractInstructions,
  extractTitle,
  findLabeledText,
} from "./extractor";
import { extractDraftText } from "./draft";
import { extractRubricTable } from "./rubric";
import { syncStrictExtractionFromServer } from "./server-config";

const PLACE_DEBOUNCE_MS = 300;

async function handleRubricalClick(pageType: string): Promise<void> {
  await syncStrictExtractionFromServer();

  const draftText = extractDraftText();
  const title = extractTitle();
  const payload = {
    sourceUrl: window.location.href,
    pageType,
    title,
    visibleText: document.querySelector("#assignment_show, main")?.textContent?.trim() ?? "",
    instructionsText: extractInstructions(),
    draftText,
    rubric: extractRubricTable(),
    metadata: {
      dueDateText: findLabeledText("Due"),
      pointsPossibleText: findLabeledText("Points"),
      submissionTypeText: findLabeledText("Submitting"),
      courseName: extractCourseName(),
    },
    captureMode: "check_with_rubrical",
    capturedAt: new Date().toISOString(),
  };

  try {
    const { data, base } = await postImport(payload);
    if (data.redirect) {
      openAssignmentModal(base, data.redirect, title);
    }
  } catch {
    alert(
      `Rubrical could not reach the local server.\n\nTried: ${RUBRICAL_API_BASES.join(", ")}\n\nFrom WSL run: make server\nFrom Windows test: curl http://localhost:8787/health -UseBasicParsing`,
    );
  }
}

function placeButton(): void {
  if (!isSupportedCanvasPage()) {
    return;
  }

  const pageType = detectPageType();
  ensureInlineButton(RUBRICAL_BUTTON_LABEL, () => {
    void handleRubricalClick(pageType);
  });
}

const debouncedPlaceButton = debounce(placeButton, PLACE_DEBOUNCE_MS);

function boot(): void {
  debouncedPlaceButton();

  const observer = new MutationObserver(() => {
    if (needsPlacement()) {
      debouncedPlaceButton();
    }
  });

  observer.observe(document.body, { childList: true, subtree: true });
}

boot();
