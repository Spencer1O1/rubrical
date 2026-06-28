import {
  debounce,
  detectPageType,
  ensureInlineButton,
  isSupportedCanvasPage,
  needsPlacement,
  setRubricalButtonEnabled,
} from "./injector";
import { RUBRICAL_API_BASES } from "./api";
import {
  isAssignmentContextReady,
  prefetchAssignmentContext,
  runImportOnClick,
  subscribeAssignmentContextReady,
} from "./import";
import {
  pauseStagedFilesSync,
  refreshStagedFileIndicators,
  resumeStagedFilesSync,
  startStagedFilesSync,
} from "./staged-files";
import { RUBRICAL_BUTTON_LABEL } from "./labels";
import { openAssignmentModal } from "./modal";
import { isRubricalDraftFilesChangedMessage } from "./panel-bridge";
import { onLongDescriptionScrapeSession } from "./scrape-session";

const PLACE_DEBOUNCE_MS = 300;
const PREFETCH_DEBOUNCE_MS = 500;

let importInFlight = false;

function syncButtonReadyState(): void {
  if (!isSupportedCanvasPage()) {
    return;
  }
  setRubricalButtonEnabled(isAssignmentContextReady(detectPageType()));
}

async function handleRubricalClick(pageType: string): Promise<void> {
  if (!isAssignmentContextReady(pageType) || importInFlight) {
    return;
  }

  importInFlight = true;
  try {
    const { redirect, title, base, draftWarning } = await runImportOnClick(pageType);
    if (draftWarning) {
      alert(`Rubrical imported your work, but:\n\n${draftWarning}`);
    }
    if (redirect) {
      openAssignmentModal(base, redirect, title);
    }
  } catch (err) {
    const detail =
      err instanceof Error && err.message.trim() !== ""
        ? err.message
        : "Unknown error";
    const serverHint = detail.toLowerCase().includes("canvas attachment")
      ? ""
      : `\n\nIf this is a connection problem, the extension tried:\n${RUBRICAL_API_BASES.join("\n")}\n\nFrom WSL run: make server\nFrom Windows test: curl http://localhost:8787/health -UseBasicParsing`;
    alert(`Rubrical import failed.\n\n${detail}${serverHint}`);
  } finally {
    importInFlight = false;
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
  syncButtonReadyState();
}

const debouncedPlaceButton = debounce(placeButton, PLACE_DEBOUNCE_MS);

const debouncedPrefetch = debounce(() => {
  if (!isSupportedCanvasPage()) {
    return;
  }
  void prefetchAssignmentContext(detectPageType());
}, PREFETCH_DEBOUNCE_MS);

function syncButtonReady(): void {
  syncButtonReadyState();
}

const debouncedStagedFilesSync = debounce(() => {
  if (!isSupportedCanvasPage()) {
    return;
  }
  void startStagedFilesSync();
}, PLACE_DEBOUNCE_MS);

function onDomMutation(): void {
  if (needsPlacement()) {
    debouncedPlaceButton();
  }
  debouncedPrefetch();
  if (isSupportedCanvasPage()) {
    syncButtonReady();
    debouncedStagedFilesSync();
  }
}

let domSyncConnected = false;
const domSyncObserver = new MutationObserver(onDomMutation);

function connectDomSync(): void {
  if (domSyncConnected || !document.body) {
    return;
  }
  domSyncObserver.observe(document.body, { childList: true, subtree: true });
  domSyncConnected = true;
}

function disconnectDomSync(): void {
  if (!domSyncConnected) {
    return;
  }
  domSyncObserver.disconnect();
  domSyncConnected = false;
}

function onRubricalPanelMessage(event: MessageEvent): void {
  if (!isRubricalDraftFilesChangedMessage(event)) {
    return;
  }
  void refreshStagedFileIndicators();
}

function boot(): void {
  window.addEventListener("message", onRubricalPanelMessage);

  subscribeAssignmentContextReady(() => {
    syncButtonReadyState();
    if (!isSupportedCanvasPage() || !isAssignmentContextReady(detectPageType())) {
      return;
    }
    void startStagedFilesSync();
  });

  // Long rubric description scrape pauses file staging (hooks disconnect) so DOM
  // mutations from the scraper do not race with upload capture; staging resumes after.
  onLongDescriptionScrapeSession((active) => {
    if (active) {
      pauseStagedFilesSync();
      disconnectDomSync();
      return;
    }

    connectDomSync();
    syncButtonReadyState();
    if (needsPlacement()) {
      debouncedPlaceButton();
    }
    resumeStagedFilesSync();
  });

  debouncedPlaceButton();
  debouncedPrefetch();
  connectDomSync();
}

boot();
