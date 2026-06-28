/**
 * Page-level anchors (visible text capture, etc.).
 */
import type { CanvasAnchor } from "./types";

export const page = {
  visibleTextRoot: {
    a2: ["main"],
    classic: ["#assignment_show", "main"],
  },
} as const satisfies Record<string, CanvasAnchor>;

export function isSupportedCanvasPath(pathname: string): boolean {
  return pathname.includes("/assignments/") || pathname.includes("/discussion_topics/");
}

export function isDiscussionPage(pathname = window.location.pathname): boolean {
  return pathname.includes("/discussion_topics/");
}
