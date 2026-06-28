import { discussion, discussionAttachmentInputSelector } from "../../canvas/anchors";
import { firstMatch, queryAnchor } from "../../canvas/query";
import { rubricalDebugLog } from "../debug";
import { normalizeFileName } from "../staging-key";
import {
  clearStagedAssignment,
  listStagedFiles,
  pingStagedFilesServiceWorker,
  putStagedFile,
  reconcileStagedFiles,
} from "../messages";
import type { StagedFileRecord } from "../types";
import {
  forgetPendingUpload,
  rememberPendingUpload,
} from "../assignment/pending-staging";
import { readDiscussionComposerAttachment } from "./composer";

let workChain: Promise<void> = Promise.resolve();
let hooksHandle: { disconnect: () => void } | null = null;
let bootedStagingKey: string | null = null;

function enqueueWork(work: () => Promise<void>): Promise<void> {
  const next = workChain.then(work);
  workChain = next.catch(() => {});
  return next;
}

async function stagePickedFile(stagingKey: string, file: File): Promise<void> {
  const buffer = await file.arrayBuffer();
  const stagedAt = new Date().toISOString();
  const record = {
    assignmentKey: stagingKey,
    fileName: file.name,
    normalizedFileName: normalizeFileName(file.name),
    stagedAt,
    mimeType: file.type || "application/octet-stream",
    bytes: buffer,
  };

  try {
    await putStagedFile({ ...record, blobBytes: buffer });
    forgetPendingUpload(record);
    rubricalDebugLog("staged file", { fileName: file.name, stagedAt });
  } catch (err) {
    rememberPendingUpload(record);
    rubricalDebugLog("staged file failed", {
      fileName: file.name,
      stagedAt,
      error: err instanceof Error ? err.message : String(err),
    });
  }
}

function composerHasAttachmentSurface(editRoot: Element): boolean {
  return Boolean(
    editRoot.querySelector(discussionAttachmentInputSelector()) ||
      queryAnchor(discussion.attachButton, editRoot) ||
      queryAnchor(discussion.attachmentItem, editRoot) ||
      firstMatch(discussion.composerAttachmentDownloadLink.a2, editRoot),
  );
}

function provisionalStagedFile(staged: StagedFileRecord[]): StagedFileRecord | null {
  const withoutCanvasId = staged.filter((file) => !file.canvasFileId);
  if (withoutCanvasId.length === 0) {
    return null;
  }
  return [...withoutCanvasId].sort((left, right) => right.stagedAt.localeCompare(left.stagedAt))[0]!;
}

async function promoteStagedCanvasId(stagingKey: string): Promise<void> {
  const attachment = readDiscussionComposerAttachment();
  if (!attachment) {
    return;
  }

  const staged = await listStagedFiles(stagingKey);
  const provisional = provisionalStagedFile(staged);
  if (!provisional) {
    return;
  }

  await reconcileStagedFiles({
    assignmentKey: stagingKey,
    promotions: [
      {
        normalizedFileName: provisional.normalizedFileName,
        stagedAt: provisional.stagedAt,
        canvasFileId: attachment.canvasFileId,
      },
    ],
  });
  rubricalDebugLog("promoted discussion staged file", {
    canvasFileId: attachment.canvasFileId,
    fileName: provisional.fileName,
  });
}

function connectSessionHooks(stagingKey: string): void {
  const editRoot = queryAnchor(discussion.editContainer);
  if (!editRoot || !composerHasAttachmentSurface(editRoot)) {
    return;
  }

  let composerHadAttachment = Boolean(readDiscussionComposerAttachment());

  const onComposerMutation = (): void => {
    const hasAttachment = Boolean(readDiscussionComposerAttachment());
    if (composerHadAttachment && !hasAttachment) {
      void enqueueWork(async () => {
        await clearStagedAssignment(stagingKey);
        rubricalDebugLog("cleared discussion session attachment", { stagingKey });
      });
    } else if (!composerHadAttachment && hasAttachment) {
      void enqueueWork(async () => {
        await promoteStagedCanvasId(stagingKey);
      });
    }
    composerHadAttachment = hasAttachment;
  };

  const onInputChange = (event: Event): void => {
    const target = event.target;
    if (
      !(target instanceof HTMLInputElement) ||
      !target.matches(discussionAttachmentInputSelector()) ||
      !editRoot.contains(target)
    ) {
      return;
    }

    const files = Array.from(target.files ?? []);
    if (files.length === 0) {
      onComposerMutation();
      return;
    }

    void enqueueWork(async () => {
      await clearStagedAssignment(stagingKey);
      await stagePickedFile(stagingKey, files[0]!);
      await promoteStagedCanvasId(stagingKey);
    });
  };

  const observer = new MutationObserver(onComposerMutation);
  observer.observe(editRoot, { childList: true, subtree: true });
  document.addEventListener("change", onInputChange, true);

  hooksHandle = {
    disconnect: () => {
      observer.disconnect();
      document.removeEventListener("change", onInputChange, true);
    },
  };
}

export function discussionSessionActive(): boolean {
  const editRoot = queryAnchor(discussion.editContainer);
  if (!editRoot) {
    return false;
  }

  if (readDiscussionComposerAttachment()) {
    return true;
  }

  if (queryAnchor(discussion.attachmentInput, editRoot)) {
    return true;
  }

  return queryAnchor(discussion.attachButton, editRoot) !== null;
}

export function disconnectDiscussionSession(): void {
  hooksHandle?.disconnect();
  hooksHandle = null;
  bootedStagingKey = null;
}

async function clearSessionIfComposerEmpty(stagingKey: string): Promise<void> {
  if (readDiscussionComposerAttachment()) {
    return;
  }

  const staged = await listStagedFiles(stagingKey);
  if (staged.length === 0) {
    return;
  }

  await clearStagedAssignment(stagingKey);
  rubricalDebugLog("cleared discussion session attachment", { stagingKey });
}

/** Keep one composer attachment mirrored in IDB for the current discussion reply. */
export async function syncDiscussionSession(stagingKey: string): Promise<void> {
  const stagingReady = await pingStagedFilesServiceWorker();
  if (!stagingReady) {
    rubricalDebugLog("service worker unavailable", { stagingKey });
    return;
  }

  if (!hooksHandle) {
    connectSessionHooks(stagingKey);
  }

  if (bootedStagingKey !== stagingKey) {
    bootedStagingKey = stagingKey;
    await clearSessionIfComposerEmpty(stagingKey);
    rubricalDebugLog("discussion session sync started", { stagingKey });
  }
}
