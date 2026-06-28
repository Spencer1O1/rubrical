import { envReaders } from "../../assignment-env";
import { firstMatch, queryAnchor, testId } from "../../query";
import type { CanvasAnchor } from "../types";
import { normalizePointsLabel } from "./helpers";
import { metadataClassic, metadataIds } from "./ids";
import { studentViewAnchor } from "./scope";

const pointsA2 = [testId(metadataIds.gradeDisplay)] as const;

const pointsClassic = [
  metadataClassic.pointsPossible,
  metadataClassic.assignmentPoints,
] as const;

/** Visible A2 markup: `<span data-testid="grade-display"><span class="points-value">50</span> Points Possible</span>` */
function readGradeDisplayElement(el: Element): string {
  const pointsValue = el.querySelector(metadataClassic.pointsValue)?.textContent?.trim();
  if (pointsValue) {
    return normalizePointsLabel(`${pointsValue} Points Possible`);
  }
  return normalizePointsLabel(el.textContent?.trim() ?? "");
}

function pointsFromScreenReaderText(text: string): string {
  const match = text.match(/(\d+(?:\.\d+)?)\s+possible\s+points/i);
  if (match) {
    return normalizePointsLabel(`${match[1]} Points Possible`);
  }
  return normalizePointsLabel(text);
}

function screenReaderPointsText(): string {
  const root = queryAnchor(studentViewAnchor) ?? document.body;

  for (const el of Array.from(
    root.querySelectorAll(metadataClassic.screenReaderContent),
  )) {
    const text = el.textContent?.trim() ?? "";
    if (/possible\s+points/i.test(text)) {
      return pointsFromScreenReaderText(text);
    }
  }

  return "";
}

function readA2PointsText(): string {
  const gradeDisplay = firstMatch(pointsA2);
  if (gradeDisplay) {
    const text = readGradeDisplayElement(gradeDisplay);
    if (text) {
      return text;
    }
  }

  return screenReaderPointsText();
}

function readClassicPointsText(): string {
  const text = firstMatch(pointsClassic)?.textContent?.trim() ?? "";
  return normalizePointsLabel(text);
}

function envPointsText(): string {
  const points = envReaders.pointsPossible();
  if (points === null || points === undefined || Number.isNaN(points)) {
    return "";
  }
  return normalizePointsLabel(String(points));
}

export const pointsAnchor = {
  a2: pointsA2,
  classic: pointsClassic,
  readA2: readA2PointsText,
  readClassic: readClassicPointsText,
  env: envPointsText,
} as const satisfies CanvasAnchor;
