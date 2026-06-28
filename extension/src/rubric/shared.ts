import { rubricCellSelectors } from "../canvas/anchors";
import type { RubricRating } from "./types";

export function normalizeText(value: string | null | undefined): string {
  return (value ?? "").replace(/\s+/g, " ").trim();
}

export function findA2Row(el: Element): Element | null {
  return el.closest("tr") ?? el.closest('[role="row"]');
}

export function a2RowCells(row: Element): Element[] {
  return Array.from(
    row.querySelectorAll(
      ':scope > td, :scope > th, :scope > [role="cell"], :scope > [role="columnheader"]',
    ),
  );
}

export function a2CellHasRatings(cell: Element): boolean {
  return (
    cell.matches(rubricCellSelectors.ratingCell) ||
    cell.querySelector(rubricCellSelectors.ratingCell) !== null
  );
}

export function a2NonRatingCells(row: Element): Element[] {
  return a2RowCells(row).filter((cell) => !a2CellHasRatings(cell));
}

const VIEW_LONGER_DESCRIPTION = /view longer description/i;

export function isLongDescriptionControl(el: Element): boolean {
  const label = normalizeText(el.textContent).toLowerCase();
  return VIEW_LONGER_DESCRIPTION.test(label);
}

export function extractCriterionNameFromCell(cell: Element): string {
  const clone = cell.cloneNode(true) as Element;
  for (const interactive of Array.from(clone.querySelectorAll("button, a"))) {
    if (isLongDescriptionControl(interactive)) {
      interactive.remove();
    }
  }

  return normalizeText(clone.textContent);
}

export function extractCriterionNameFromRow(row: Element): string {
  const cells = a2NonRatingCells(row);
  if (cells.length === 0) {
    return "";
  }

  return extractCriterionNameFromCell(cells[0]!);
}

export function hasRatingBody(rating: RubricRating): boolean {
  return rating.title !== "" || rating.description !== "";
}

export function parsePointsFromText(text: string): { description: string; points: string } {
  const pointsMatch = text.match(/(\d+(?:\.\d+)?)\s*pts?\s*$/i);
  if (!pointsMatch || pointsMatch.index == null) {
    return { description: text, points: "" };
  }

  return {
    description: text.slice(0, pointsMatch.index).trim(),
    points: pointsMatch[0].trim(),
  };
}

export function extractA2PointsFromCell(cell: Element): string {
  const suffix = cell.querySelector(rubricCellSelectors.criterionMaxPointsLabel);
  return suffix ? normalizeText(suffix.textContent) : "";
}

export function mergeRating(into: RubricRating, from: RubricRating): RubricRating {
  return {
    title: into.title || from.title,
    description: into.description || from.description,
    points: into.points || from.points,
  };
}
