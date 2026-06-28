import { rubric, rubricCellSelectors } from "../canvas/anchors";
import { firstMatch, queryAnchor } from "../canvas/query";
import {
  extractCriterionNameFromCell,
  hasRatingBody,
  normalizeText,
  parsePointsFromText,
} from "./shared";
import type { RubricRating, RubricTable, RubricTableRow } from "./types";
import { DEFAULT_HEADER } from "./types";

function findClassicRubricTable(): HTMLTableElement | null {
  return queryAnchor<HTMLTableElement>(rubric.classicTable);
}

function classicRowCells(tr: HTMLTableRowElement): HTMLTableCellElement[] {
  return Array.from(tr.querySelectorAll(":scope > th, :scope > td"));
}

function isClassicHeaderRow(cells: HTMLTableCellElement[]): boolean {
  const joined = cells.map((cell) => normalizeText(cell.textContent).toLowerCase()).join(" ");
  return joined.includes("criteria") && joined.includes("rating");
}

function isClassicPointsCell(text: string): boolean {
  return /^\/\s*\d+(?:\.\d+)?\s*pts?$/i.test(text) || /^\d+(?:\.\d+)?\s*pts?$/i.test(text);
}

function parseClassicRatingCell(cell: Element): RubricRating {
  const title = normalizeText(cell.querySelector(rubricCellSelectors.classicRatingTitle)?.textContent);
  const description = normalizeText(cell.querySelector(rubricCellSelectors.classicRatingDescription)?.textContent);
  const points = normalizeText(cell.querySelector(rubricCellSelectors.classicRatingPoints)?.textContent);

  if (title || description) {
    return { title, description, points };
  }

  const text = normalizeText(cell.textContent);
  if (!text) {
    return { title: "", description: "", points: "" };
  }

  const parsed = parsePointsFromText(text);
  return {
    title: "",
    description: parsed.description,
    points: points || parsed.points,
  };
}

function extractClassicRatingsFromCell(cell: HTMLTableCellElement): RubricRating[] {
  const nestedRow = cell.querySelector(":scope > table tr, :scope > div > table tr");
  if (nestedRow) {
    const ratingCells = Array.from(
      nestedRow.querySelectorAll(rubricCellSelectors.classicNestedRating),
    );
    const ratings = ratingCells.map((el) => parseClassicRatingCell(el)).filter(hasRatingBody);
    if (ratings.length > 0) {
      return ratings;
    }
  }

  const ratingBoxes = cell.querySelectorAll(rubricCellSelectors.classicRatingBox);
  if (ratingBoxes.length > 0) {
    return Array.from(ratingBoxes)
      .map((el) => parseClassicRatingCell(el))
      .filter(hasRatingBody);
  }

  const single = parseClassicRatingCell(cell);
  return hasRatingBody(single) ? [single] : [];
}

function extractClassicRatingsFromRow(cells: HTMLTableCellElement[]): RubricRating[] {
  if (cells.length <= 1) {
    return [];
  }

  if (cells.length >= 4) {
    const ratingCells = cells.slice(1, -1);
    const direct = ratingCells.map((cell) => parseClassicRatingCell(cell)).filter(hasRatingBody);
    if (direct.length > 0) {
      return direct;
    }
    return ratingCells.flatMap((cell) => extractClassicRatingsFromCell(cell));
  }

  const middleCells = cells.slice(1);
  const lastText = normalizeText(middleCells[middleCells.length - 1]?.textContent);
  if (middleCells.length > 1 && isClassicPointsCell(lastText)) {
    const ratingCells = middleCells.slice(0, -1);
    const direct = ratingCells.map((cell) => parseClassicRatingCell(cell)).filter(hasRatingBody);
    if (direct.length > 0) {
      return direct;
    }
    return ratingCells.flatMap((cell) => extractClassicRatingsFromCell(cell));
  }

  return extractClassicRatingsFromCell(cells[1]!);
}

function extractClassicPointsFromRow(cells: HTMLTableCellElement[]): string {
  const lastText = normalizeText(cells[cells.length - 1]?.textContent);
  if (isClassicPointsCell(lastText)) {
    return lastText;
  }
  if (cells.length >= 3) {
    return lastText;
  }
  return "";
}

export function extractClassicRubric(longDescriptions: string[]): RubricTable | null {
  const table = findClassicRubricTable();
  if (!table) {
    return null;
  }

  const allRows = Array.from(
    table.querySelectorAll(":scope > tbody > tr, :scope > thead > tr, :scope > tr"),
  ) as HTMLTableRowElement[];
  if (allRows.length === 0) {
    return null;
  }

  let header = DEFAULT_HEADER;
  let bodyRows = allRows;

  const firstCells = classicRowCells(allRows[0]!);
  if (isClassicHeaderRow(firstCells)) {
    header = firstCells.map((cell) => normalizeText(cell.textContent));
    bodyRows = allRows.slice(1);
  }

  const rows: RubricTableRow[] = [];
  let longDescriptionIndex = 0;
  for (const tr of bodyRows) {
    if (firstMatch(rubric.criterionRatings.a2, tr)) {
      continue;
    }

    const cells = classicRowCells(tr);
    if (cells.length === 0) {
      continue;
    }

    const criterion = extractCriterionNameFromCell(cells[0]!);
    if (!criterion) {
      continue;
    }

    rows.push({
      criterion,
      criterionLongDescription: longDescriptions[longDescriptionIndex] ?? "",
      ratings: extractClassicRatingsFromRow(cells),
      points: extractClassicPointsFromRow(cells),
    });
    longDescriptionIndex++;
  }

  if (rows.length === 0) {
    return null;
  }

  return { header, rows };
}
