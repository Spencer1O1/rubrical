import { RUBRICAL_API_BASES } from "./api";

export const RUBRICAL_DRAFT_FILES_CHANGED = "rubrical:draft-files-changed";

export function isRubricalDraftFilesChangedMessage(event: MessageEvent): boolean {
  if (event.data?.type !== RUBRICAL_DRAFT_FILES_CHANGED) {
    return false;
  }

  try {
    return RUBRICAL_API_BASES.some((base) => new URL(base).origin === event.origin);
  } catch {
    return false;
  }
}
