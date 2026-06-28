import { envReaders } from "../../assignment-env";
import { firstMatch, testId } from "../../query";
import type { CanvasAnchor } from "../types";
import { normalizeDueLabel } from "./helpers";
import { metadataClassic, metadataIds } from "./ids";

const dueDateA2 = [testId(metadataIds.dueDate)] as const;

const dueDateClassic = [
  metadataClassic.assignmentDatesDueDate,
  metadataClassic.assignmentShowDateDue,
] as const;

/** Read ISO due date and label from `<time data-testid="due-date" datetime="…">`. */
export function readDueFromTimeElement(): { dueAt: string; dueDateText: string } {
  const dueRoot = document.querySelector(dueDateA2[0]);
  if (!dueRoot) {
    return { dueAt: "", dueDateText: "" };
  }

  const timeEl =
    dueRoot instanceof HTMLTimeElement ? dueRoot : dueRoot.querySelector("time");
  const dueAt = timeEl instanceof HTMLTimeElement ? (timeEl.dateTime?.trim() ?? "") : "";

  const visibleDesktop = dueRoot.querySelector(".visible-desktop")?.textContent?.trim();
  const label =
    visibleDesktop ||
    timeEl?.textContent?.trim() ||
    dueRoot.textContent?.trim() ||
    "";
  const dueDateText = normalizeDueLabel(label);
  return { dueAt, dueDateText };
}

function readA2DueDateText(): string {
  return readDueFromTimeElement().dueDateText;
}

function readClassicDueDateText(): string {
  const text = firstMatch(dueDateClassic)?.textContent?.trim();
  return text ? normalizeDueLabel(text) : "";
}

function envDueDateText(): string {
  const iso = envReaders.dueAt();
  if (!iso) {
    return "";
  }
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  const formatted = date.toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
    hour12: true,
  });
  return normalizeDueLabel(`Due ${formatted}`);
}

export const dueDateAnchor = {
  a2: dueDateA2,
  classic: dueDateClassic,
  readA2: readA2DueDateText,
  readClassic: readClassicDueDateText,
  env: envDueDateText,
} as const satisfies CanvasAnchor;
