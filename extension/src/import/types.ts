import type { AssignmentMetadata } from "../metadata";
import type { RubricTable } from "../rubric";
import type { StagedFileRecord } from "../staged-files/types";

/**
 * Teacher-published assignment context.
 * Cached in extension memory when the page is ready — not sent to the server until click.
 */
export type CachedAssignmentContext = {
  sourceUrl: string;
  pageType: string;
  title: string;
  instructionsText: string;
  rubric: RubricTable | null;
  metadata: AssignmentMetadata;
  cachedAt: string;
  /** True once a prefetch ran while long-description buttons were on the page. */
  longDescriptionsFetched: boolean;
};

export type DraftFile = {
  fileName: string;
  mimeType: string;
  contentBase64: string;
  canvasFileId?: string;
  sortOrder?: number;
};

export type DraftFileRef = {
  serverFileId: number;
  fileName: string;
  canvasFileId?: string;
  sortOrder?: number;
};

export type StagedUploadRecord = StagedFileRecord & {
  sortOrder: number;
};

/**
 * Student-owned submission state.
 * Captured only when the user clicks Check with Rubrical.
 */
export type LiveImportCapture = {
  visibleText: string;
  draftText: string;
  draftUrl: string;
  draftKind: string;
  draftFiles: DraftFile[];
  draftFileRefs: DraftFileRef[];
  stagedUploads: StagedUploadRecord[];
  fileImportWarnings: string[];
  capturedAt: string;
};

/** POST /imports body — prefetched fields + live capture merged at click time. */
export type ImportPayload = {
  sourceUrl: string;
  pageType: string;
  title: string;
  visibleText: string;
  instructionsText: string;
  draftText: string;
  draftUrl: string;
  draftKind: string;
  draftFiles: DraftFile[];
  draftFileRefs: DraftFileRef[];
  fileImportWarnings: string[];
  rubric: RubricTable | null;
  metadata: AssignmentMetadata;
  capturedAt: string;
};
