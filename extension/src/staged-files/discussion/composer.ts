import { discussion } from "../../canvas/anchors";
import { firstMatch, queryAnchor } from "../../canvas/query";

export type DiscussionComposerAttachment = {
  fileName: string;
  downloadUrl: string;
  canvasFileId: string;
};

function canvasFileIdFromUrl(url: string): string | null {
  const match = url.match(/\/files\/(\d+)/i);
  return match?.[1] ?? null;
}

function normalizeCanvasDownloadUrl(href: string, canvasFileId: string): string {
  if (href.includes("/download")) {
    return href;
  }

  try {
    const origin = new URL(href, window.location.origin).origin;
    return `${origin}/files/${canvasFileId}/download?download_frd=1`;
  } catch {
    return href;
  }
}

function normalizeText(value: string | null | undefined): string {
  return (value ?? "").replace(/\s+/g, " ").trim();
}

function fileNameFromAttachmentItemAriaLabel(editRoot: Element): string | null {
  const item = queryAnchor(discussion.attachmentItem, editRoot);
  const label = item?.getAttribute("aria-label")?.trim() ?? "";
  const match = label.match(/^Replace\s+(.+?)\s+button$/i);
  return match?.[1]?.includes(".") ? match[1] : null;
}

function fileNameFromComposerAttachmentLink(
  link: HTMLAnchorElement,
  editRoot: Element,
): string | null {
  const fromHidden = normalizeText(
    firstMatch(discussion.attachmentFileName.a2, editRoot)?.textContent,
  );
  if (fromHidden.includes(".")) {
    return fromHidden;
  }

  const fromLinkHidden = normalizeText(
    link.querySelector('[aria-hidden="true"] span[wrap="normal"]')?.textContent,
  );
  if (fromLinkHidden.includes(".")) {
    return fromLinkHidden;
  }

  const label = normalizeText(link.textContent);
  const downloadMatch = label.match(/Download\s+(\S+\.\S+)\s*$/i);
  if (downloadMatch?.[1]) {
    return downloadMatch[1];
  }

  const trailingName = label.match(/(\S+\.\S+)\s*$/);
  if (trailingName?.[1] && !trailingName[1].startsWith("http")) {
    return trailingName[1];
  }

  const fromAria = fileNameFromAttachmentItemAriaLabel(editRoot);
  if (fromAria) {
    return fromAria;
  }

  return null;
}

/** Visible attachment filename in the reply composer, if any. */
export function extractDiscussionAttachmentFileName(): string | null {
  return readDiscussionComposerAttachment()?.fileName ?? null;
}

/** Reply composer attachment metadata (fixtures/discussion-attachment). */
export function readDiscussionComposerAttachment(): DiscussionComposerAttachment | null {
  const editRoot = queryAnchor(discussion.editContainer);
  if (!editRoot) {
    return null;
  }

  const link = firstMatch(discussion.composerAttachmentDownloadLink.a2, editRoot);
  if (!(link instanceof HTMLAnchorElement)) {
    return null;
  }

  const href = link.href.trim();
  const canvasFileId = canvasFileIdFromUrl(href);
  if (!canvasFileId) {
    return null;
  }

  const fileName =
    fileNameFromComposerAttachmentLink(link, editRoot) ?? `canvas-file-${canvasFileId}`;

  return {
    fileName,
    downloadUrl: normalizeCanvasDownloadUrl(href, canvasFileId),
    canvasFileId,
  };
}
