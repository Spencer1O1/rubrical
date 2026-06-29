import { RUBRICAL_API_BASE } from "./api";

export const RUBRICAL_DRAFT_FILES_CHANGED = "rubrical:draft-files-changed";

export function isRubricalDraftFilesChangedMessage(event: MessageEvent): boolean {
  if (event.data?.type !== RUBRICAL_DRAFT_FILES_CHANGED) {
    return false;
  }

  try {
    return new URL(RUBRICAL_API_BASE).origin === event.origin;
  } catch {
    return false;
  }
}
