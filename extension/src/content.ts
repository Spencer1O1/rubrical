import {
  debounce,
  detectPageType,
  ensureInlineButton,
  isSupportedCanvasPage,
  needsPlacement,
  setRubricalButtonEnabled,
} from "./injector";
import { RUBRICAL_API_BASE } from "./api";
import { fetchSession, RubricalConnectionError } from "./auth-api";
import { showAuthModal } from "./auth-modal";
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

function isDevApiBase(base: string): boolean {
  return /localhost|127\.0\.0\.1/.test(base);
}

function isNetworkishDetail(detail: string): boolean {
  const lower = detail.toLowerCase();
  return (
    lower.includes("failed to fetch") ||
    lower.includes("networkerror") ||
    lower.includes("network error") ||
    lower.includes("timed out") ||
    lower.includes("timeout") ||
    lower.includes("abort") ||
    lower.includes("couldn't reach rubrical") ||
    lower.includes("service worker unavailable")
  );
}

/** User-facing import failure copy — no WSL/dev instructions in production builds. */
function formatImportFailureAlert(err: unknown): string {
  const detail =
    err instanceof Error && err.message.trim() !== ""
      ? err.message.trim()
      : "Something went wrong.";

  if (detail.toLowerCase().includes("canvas attachment")) {
    return `Rubrical couldn't import this assignment.\n\n${detail}`;
  }

  if (err instanceof RubricalConnectionError || isNetworkishDetail(detail)) {
    if (isDevApiBase(RUBRICAL_API_BASE)) {
      return (
        `Rubrical couldn't reach the local server.\n\n` +
        `Start it with make server, then try again.\n` +
        `(${RUBRICAL_API_BASE})`
      );
    }
    return (
      `Rubrical couldn't reach the server.\n\n` +
      `1. Open ${RUBRICAL_API_BASE} and sign in in this browser\n` +
      `2. Try Check with Rubrical again\n\n` +
      `If it still fails, reinstall from ${RUBRICAL_API_BASE}/install`
    );
  }

  return `Rubrical couldn't import this assignment.\n\n${detail}`;
}

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
    const session = await fetchSession();
    if (!session) {
      showAuthModal({
        onSignedIn: () => {
          void handleRubricalClick(pageType);
        },
      });
      return;
    }

    const { redirect, title, base, draftWarning } = await runImportOnClick(pageType);
    if (draftWarning) {
      alert(`Rubrical imported your work, but:\n\n${draftWarning}`);
    }
    if (redirect) {
      openAssignmentModal(base, redirect, title);
    }
  } catch (err) {
    alert(formatImportFailureAlert(err));
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
