import { draftUrl } from "./canvas/anchors";
import { queryAnchorAll } from "./canvas/query";
import { getSubmissionRoot } from "./draft";

/** Read the website URL from Canvas when the Web URL submission tab is active. */
export function extractDraftURL(): string {
  const root = getSubmissionRoot();

  for (const input of queryAnchorAll<HTMLInputElement>(draftUrl.urlInput, root)) {
    const value = input.value?.trim();
    if (value) {
      return value;
    }
  }

  return "";
}
