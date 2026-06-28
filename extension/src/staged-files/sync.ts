import { detectActiveSubmissionKind } from "../submission-kind";
import { isDiscussionPage } from "../canvas/anchors/page";
import {
  afterSuccessfulAssignmentImportClearStaging,
  clearAssignmentIndicators,
  disconnectAssignmentStaging,
  refreshAssignmentIndicators,
  syncAssignmentStaging,
} from "./assignment/sync";
import { scanAssignmentUploadedRows } from "./assignment/canvas-rows";
import { clearPendingUploads } from "./assignment/pending-staging";
import {
  disconnectDiscussionSession,
  discussionSessionActive,
  syncDiscussionSession,
} from "./discussion/session";
import { clearStagedAssignment } from "./store";
import { stagingKeyFromPage } from "./staging-key";

let syncPaused = false;
let activeStagingKey: string | null = null;

function shouldRunStagedFilesSync(): boolean {
  if (isDiscussionPage() && discussionSessionActive()) {
    return true;
  }

  if (detectActiveSubmissionKind() === "file") {
    return true;
  }

  return scanAssignmentUploadedRows().length > 0;
}

function disconnectAll(): void {
  disconnectAssignmentStaging();
  disconnectDiscussionSession();
}

export function pauseStagedFilesSync(): void {
  syncPaused = true;
  disconnectAll();
}

export function resumeStagedFilesSync(): void {
  syncPaused = false;
  void startStagedFilesSync();
}

export async function startStagedFilesSync(): Promise<void> {
  if (syncPaused || !shouldRunStagedFilesSync()) {
    if (!isDiscussionPage()) {
      clearAssignmentIndicators();
    }
    disconnectAll();
    return;
  }

  const stagingKey = stagingKeyFromPage();
  if (!stagingKey) {
    return;
  }

  if (activeStagingKey !== stagingKey) {
    disconnectAll();
    activeStagingKey = stagingKey;
  }

  if (isDiscussionPage()) {
    await syncDiscussionSession(stagingKey);
    return;
  }

  await syncAssignmentStaging(stagingKey);
}

export async function refreshStagedFileIndicators(): Promise<void> {
  await refreshAssignmentIndicators();
}

export async function afterSuccessfulImportClearStaging(): Promise<void> {
  const stagingKey = stagingKeyFromPage();
  if (isDiscussionPage()) {
    if (stagingKey) {
      await clearStagedAssignment(stagingKey);
      clearPendingUploads(stagingKey);
      disconnectDiscussionSession();
    }
    return;
  }

  await afterSuccessfulAssignmentImportClearStaging();
}
