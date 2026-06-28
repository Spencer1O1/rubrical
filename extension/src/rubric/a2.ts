import { rubric, rubricCellSelectors } from "../canvas/anchors";
import { firstMatch, queryAnchorAll } from "../canvas/query";
import { isStrictExtraction } from "../strict";
import {
  extractA2PointsFromCell,
  extractCriterionNameFromRow,
  findA2Row,
  a2NonRatingCells,
  hasRatingBody,
  mergeRating,
  normalizeText,
  parsePointsFromText,
} from "./shared";
import type { RubricRating, RubricTable, RubricTableRow } from "./types";
import { DEFAULT_HEADER } from "./types";

function extractA2PointsFromRatingsContainer(
  container: Element,
  pointsByIndex: string[],
  index: number,
): string {
  const row = findA2Row(container);
  if (row) {
    const cells = a2NonRatingCells(row);
    if (cells.length >= 2) {
      const cell = cells[cells.length - 1];
      if (cell) {
        const text = extractA2PointsFromCell(cell);
        if (text) {
          return text;
        }
      }
    }
  }

  return pointsByIndex[index] ?? "";
}

function extractA2PointsFromRow(row: Element): string {
  const cells = a2NonRatingCells(row);
  if (cells.length <= 1) {
    return "";
  }

  const cell = cells[cells.length - 1];
  if (!cell) {
    return "";
  }

  return extractA2PointsFromCell(cell);
}

function findA2RubricRows(): Element[] {
  const seen = new Set<Element>();
  const rows: Element[] = [];

  const addRow = (row: Element | null) => {
    if (row && !seen.has(row)) {
      seen.add(row);
      rows.push(row);
    }
  };

  for (const container of queryAnchorAll(rubric.criterionRatings)) {
    addRow(findA2Row(container));
  }

  for (const rating of queryAnchorAll<HTMLElement>({
    a2: [rubricCellSelectors.ratingCell],
    classic: [],
  })) {
    addRow(findA2Row(rating));
  }

  return rows;
}

function extractA2Header(anchor: Element): string[] {
  const table = anchor.closest("table") ?? anchor.closest('[role="table"]');
  if (!table) {
    return isStrictExtraction() ? [] : DEFAULT_HEADER;
  }

  const headerRow =
    table.querySelector("thead tr") ??
    table.querySelector('[role="rowgroup"] [role="row"]') ??
    table.querySelector("tr");
  if (!headerRow) {
    return isStrictExtraction() ? [] : DEFAULT_HEADER;
  }

  const headerCells = headerRow.querySelectorAll(
    ':scope > th, :scope > [role="columnheader"]',
  );
  if (headerCells.length > 0) {
    const labels = Array.from(headerCells).map((cell) => normalizeText(cell.textContent));
    if (labels.some((label) => label !== "")) {
      return labels;
    }
  }

  const bodyCells = headerRow.querySelectorAll(':scope > td, :scope > [role="cell"]');
  if (bodyCells.length >= 2) {
    const labels = Array.from(bodyCells).map((cell) => normalizeText(cell.textContent));
    if (labels.some((label) => label !== "")) {
      return labels;
    }
  }

  return isStrictExtraction() ? [] : DEFAULT_HEADER;
}

function a2RatingIndexFromTestId(testId: string): number | null {
  const match = testId.match(/-ratings-(\d+)$/);
  return match ? Number(match[1]) : null;
}

function isPointsText(text: string): boolean {
  return /^\d+(?:\.\d+)?\s*pts?\.?$/i.test(text);
}

function parseA2RatingParts(cell: Element): RubricRating | null {
  const paragraphs = cell.querySelectorAll("p");
  if (paragraphs.length >= 2) {
    return {
      title: normalizeText(paragraphs[0]?.textContent),
      description: normalizeText(paragraphs[1]?.textContent),
      points: paragraphs.length >= 3 ? normalizeText(paragraphs[2]?.textContent) : "",
    };
  }

  const flexItems = cell.querySelectorAll(rubricCellSelectors.flexItem);
  if (flexItems.length >= 2) {
    const items = Array.from(flexItems);
    const title = normalizeText(items[0]?.textContent);
    const lastText = normalizeText(items[items.length - 1]?.textContent);
    if (items.length >= 3 && isPointsText(lastText)) {
      return {
        title,
        description: normalizeText(items[1]?.textContent),
        points: lastText,
      };
    }
    return {
      title,
      description: normalizeText(items[1]?.textContent),
      points: "",
    };
  }

  const flexColumn = cell.querySelector(rubricCellSelectors.flexColumn);
  if (flexColumn && flexColumn.children.length >= 2) {
    const items = Array.from(flexColumn.children);
    const title = normalizeText(items[0]?.textContent);
    const lastText = normalizeText(items[items.length - 1]?.textContent);
    if (items.length >= 3 && isPointsText(lastText)) {
      return {
        title,
        description: normalizeText(items[1]?.textContent),
        points: lastText,
      };
    }
    return {
      title,
      description: normalizeText(items[1]?.textContent),
      points: "",
    };
  }

  return null;
}

function parseA2RatingCell(cell: Element): RubricRating {
  const points = normalizeText(
    cell.querySelector(rubricCellSelectors.ratingPoints)?.textContent,
  );

  const structured = parseA2RatingParts(cell);
  if (structured && (hasRatingBody(structured) || structured.points !== "" || points !== "")) {
    return { ...structured, points: structured.points || points };
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

function extractA2Ratings(container: Element): RubricRating[] {
  const cells = container.querySelectorAll<HTMLElement>(rubricCellSelectors.ratingCell);
  const byIndex = new Map<number, RubricRating>();

  for (const cell of Array.from(cells)) {
    const index = a2RatingIndexFromTestId(cell.getAttribute("data-testid") ?? "");
    if (index === null) {
      continue;
    }

    const parsed = parseA2RatingCell(cell);
    if (!hasRatingBody(parsed) && parsed.points === "") {
      continue;
    }

    const existing = byIndex.get(index);
    byIndex.set(index, existing ? mergeRating(existing, parsed) : parsed);
  }

  return Array.from(byIndex.entries())
    .sort(([left], [right]) => left - right)
    .map(([, rating]) => rating)
    .filter(hasRatingBody);
}

export function extractA2TraditionalRubric(longDescriptions: string[]): RubricTable | null {
  const ratingsContainers = queryAnchorAll(rubric.criterionRatings);
  if (ratingsContainers.length > 0) {
    const rows: RubricTableRow[] = [];
    for (let index = 0; index < ratingsContainers.length; index++) {
      const container = ratingsContainers[index]!;
      const row = findA2Row(container);
      const criterion = row ? extractCriterionNameFromRow(row) : "";
      const points = extractA2PointsFromRatingsContainer(container, [], index);
      const ratings = extractA2Ratings(container);

      if (!criterion && ratings.length === 0) {
        continue;
      }

      rows.push({
        criterion,
        criterionLongDescription: longDescriptions[index] ?? "",
        ratings,
        points,
      });
    }

    if (rows.length > 0) {
      const header = extractA2Header(ratingsContainers[0]!);
      return { header, rows };
    }
  }

  const rubricRows = findA2RubricRows();
  if (rubricRows.length === 0) {
    return null;
  }

  const rows: RubricTableRow[] = [];
  let longDescriptionIndex = 0;
  for (const row of rubricRows) {
    const ratingsContainer =
      firstMatch(rubric.criterionRatings.a2, row) ?? row;
    const criterion = extractCriterionNameFromRow(row);
    const points = extractA2PointsFromRow(row);
    const ratings = extractA2Ratings(ratingsContainer);

    if (!criterion && ratings.length === 0) {
      continue;
    }

    rows.push({
      criterion,
      criterionLongDescription: longDescriptions[longDescriptionIndex] ?? "",
      ratings,
      points,
    });
    longDescriptionIndex++;
  }

  if (rows.length === 0) {
    return null;
  }

  const header = extractA2Header(rubricRows[0]!);
  return { header, rows };
}
