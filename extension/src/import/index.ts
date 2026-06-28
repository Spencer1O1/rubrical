/**
 * Import pipeline:
 *
 * - **Assignment context** — instructions, rubric, metadata (cached until click).
 * - **Live capture** — draft text/files/URL read from Canvas at click time.
 * - **Click (`run`)** — merges context + capture → POST /imports.
 *
 * Draft file bytes live under `staged-files/store.ts` (Canvas page-origin IDB; multipart after import).
 */

export type {
  CachedAssignmentContext,
  DraftFile,
  DraftFileRef,
  ImportPayload,
  LiveImportCapture,
} from "./types";

export { buildImportPayload } from "./payload";
export { extractLiveImportCapture } from "./live-capture";
export {
  assignmentContextSignalsPresent,
  extractAssignmentContext,
  rubricPresentInDOM,
} from "./assignment-context";
export {
  clearAssignmentContextCache,
  getCachedAssignmentContext,
  getOrPrefetchAssignmentContext,
  isAssignmentContextReady,
  prefetchAssignmentContext,
  subscribeAssignmentContextReady,
} from "./prefetch";
export { runImportOnClick, type ImportResult } from "./run";
