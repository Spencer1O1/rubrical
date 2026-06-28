import { extractInstructions } from "../extractor";
import { pageHasCriterionLongDescriptionButtons } from "../rubric";
import { normalizeSourceUrl } from "../staged-files/normalize-source-url";
import {
  assignmentContextSignalsPresent,
  extractAssignmentContext,
  rubricPresentInDOM,
} from "./assignment-context";
import type { CachedAssignmentContext } from "./types";

const cache = new Map<string, CachedAssignmentContext>();

let prefetchInFlight: { key: string; promise: Promise<void> } | null = null;

const readyListeners = new Set<() => void>();

function notifyReadyChange(): void {
  for (const listener of readyListeners) {
    listener();
  }
}

function cacheKey(sourceUrl: string): string {
  return normalizeSourceUrl(sourceUrl);
}

function cacheMissingLongDescriptions(cached: CachedAssignmentContext): boolean {
  if (cached.longDescriptionsFetched) {
    return false;
  }
  if (!pageHasCriterionLongDescriptionButtons()) {
    return false;
  }
  return (cached.rubric?.rows.length ?? 0) > 0;
}

function shouldPrefetch(pageType: string): boolean {
  if (pageType === "unknown" || !assignmentContextSignalsPresent()) {
    return false;
  }

  const key = cacheKey(window.location.href);
  const cached = cache.get(key);
  if (!cached) {
    return true;
  }

  if (!cached.instructionsText.trim() && extractInstructions().trim()) {
    return true;
  }

  if (rubricPresentInDOM() && (cached.rubric?.rows.length ?? 0) === 0) {
    return true;
  }

  if (cacheMissingLongDescriptions(cached)) {
    return true;
  }

  return false;
}

/** True when assignment context is cached and no prefetch (e.g. long descriptions) is in progress. */
export function isAssignmentContextReady(pageType: string): boolean {
  if (pageType === "unknown" || !assignmentContextSignalsPresent()) {
    return false;
  }

  const key = cacheKey(window.location.href);
  if (prefetchInFlight?.key === key) {
    return false;
  }

  return !shouldPrefetch(pageType);
}

export function subscribeAssignmentContextReady(callback: () => void): () => void {
  readyListeners.add(callback);
  return () => {
    readyListeners.delete(callback);
  };
}

export function getCachedAssignmentContext(
  sourceUrl = window.location.href,
): CachedAssignmentContext | null {
  return cache.get(cacheKey(sourceUrl)) ?? null;
}

export async function prefetchAssignmentContext(pageType: string): Promise<void> {
  if (!shouldPrefetch(pageType)) {
    return;
  }

  const key = cacheKey(window.location.href);
  if (prefetchInFlight?.key === key) {
    await prefetchInFlight.promise;
    return;
  }

  notifyReadyChange();

  const promise = (async () => {
    const context = await extractAssignmentContext(pageType);
    cache.set(cacheKey(context.sourceUrl), context);
  })();

  prefetchInFlight = { key, promise };
  try {
    await promise;
  } finally {
    if (prefetchInFlight?.key === key) {
      prefetchInFlight = null;
    }
    notifyReadyChange();
  }
}

export async function getOrPrefetchAssignmentContext(
  pageType: string,
): Promise<CachedAssignmentContext> {
  const cached = getCachedAssignmentContext();
  if (cached && !shouldPrefetch(pageType)) {
    return cached;
  }

  await prefetchAssignmentContext(pageType);
  return getCachedAssignmentContext() ?? extractAssignmentContext(pageType);
}

export function clearAssignmentContextCache(sourceUrl?: string): void {
  if (sourceUrl) {
    cache.delete(cacheKey(sourceUrl));
  } else {
    cache.clear();
  }
  notifyReadyChange();
}
