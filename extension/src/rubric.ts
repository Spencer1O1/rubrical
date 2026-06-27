import { extractNullableWithFallbacks } from "./extract-with-fallbacks";
import { isStrictExtraction } from "./strict";

export type RubricRating = {
  title: string;
  description: string;
  points: string;
};

export type RubricTableRow = {
  criterion: string;
  ratings: RubricRating[];
  points: string;
};

export type RubricTable = {
  header: string[];
  rows: RubricTableRow[];
};

const DEFAULT_HEADER = ["Criteria", "Ratings", "Points"];

/** Canvas A2 traditional rubric — one rating cell per column (see canvas-lms rubric tests). */
const A2_RATING_CELL_SELECTOR = '[data-testid^="traditional-criterion-"][data-testid*="-ratings-"]';

function normalizeText(value: string | null | undefined): string {
  return (value ?? "").replace(/\s+/g, " ").trim();
}

function findA2Row(el: Element): Element | null {
  return el.closest("tr") ?? el.closest('[role="row"]');
}

function a2RowCells(row: Element): Element[] {
  return Array.from(
    row.querySelectorAll(
      ':scope > td, :scope > th, :scope > [role="cell"], :scope > [role="columnheader"]',
    ),
  );
}

function a2CellHasRatings(cell: Element): boolean {
  return cell.matches(A2_RATING_CELL_SELECTOR) || cell.querySelector(A2_RATING_CELL_SELECTOR) !== null;
}

function a2NonRatingCells(row: Element): Element[] {
  return a2RowCells(row).filter((cell) => !a2CellHasRatings(cell));
}

function extractA2CriterionFromRow(row: Element): string {
  const byTestId = row.querySelector('[data-testid="traditional-view-criterion-description"]');
  if (byTestId) {
    const text = normalizeText(byTestId.textContent);
    if (text) {
      return text;
    }
  }

  const cells = a2NonRatingCells(row);
  if (cells.length === 0) {
    return "";
  }

  return normalizeText(cells[0]?.textContent);
}

function extractA2PointsFromRow(row: Element): string {
  const byTestId = row.querySelector('[data-testid="traditional-view-criterion-points"]');
  if (byTestId) {
    const text = normalizeText(byTestId.textContent);
    if (text) {
      return text;
    }
  }

  const cells = a2NonRatingCells(row);
  if (cells.length <= 1) {
    return "";
  }

  return normalizeText(cells[cells.length - 1]?.textContent);
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

  for (const container of Array.from(
    document.querySelectorAll('[data-testid="traditional-view-criterion-ratings"]'),
  )) {
    addRow(findA2Row(container));
  }

  for (const rating of Array.from(
    document.querySelectorAll<HTMLElement>(A2_RATING_CELL_SELECTOR),
  )) {
    addRow(findA2Row(rating));
  }

  return rows;
}

function extractA2Header(anchor: Element): string[] {
  const headerRoot = document.querySelector('[data-testid="traditional-view-rubric-header"]');
  if (headerRoot) {
    const headerCells = headerRoot.querySelectorAll('th, [role="columnheader"]');
    if (headerCells.length > 0) {
      const labels = Array.from(headerCells).map((cell) => normalizeText(cell.textContent));
      if (labels.some((label) => label !== "")) {
        return labels;
      }
    }
  }

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

function hasRatingBody(rating: RubricRating): boolean {
  return rating.title !== "" || rating.description !== "";
}

function parsePointsFromText(text: string): { description: string; points: string } {
  const pointsMatch = text.match(/(\d+(?:\.\d+)?)\s*pts?\s*$/i);
  if (!pointsMatch || pointsMatch.index == null) {
    return { description: text, points: "" };
  }

  return {
    description: text.slice(0, pointsMatch.index).trim(),
    points: pointsMatch[0].trim(),
  };
}

function mergeRating(into: RubricRating, from: RubricRating): RubricRating {
  return {
    title: into.title || from.title,
    description: into.description || from.description,
    points: into.points || from.points,
  };
}

// --- A2 traditional view (default) ---

function a2RatingIndexFromTestId(testId: string): number | null {
  const match = testId.match(/-ratings-(\d+)$/);
  return match ? Number(match[1]) : null;
}

function isPointsText(text: string): boolean {
  return /^\d+(?:\.\d+)?\s*pts?\.?$/i.test(text);
}

/** A2 rating cells expose label, description, and points as separate blocks (p or InstUI flex items). */
function parseA2RatingParts(cell: Element): RubricRating | null {
  const paragraphs = cell.querySelectorAll("p");
  if (paragraphs.length >= 2) {
    return {
      title: normalizeText(paragraphs[0]?.textContent),
      description: normalizeText(paragraphs[1]?.textContent),
      points: paragraphs.length >= 3 ? normalizeText(paragraphs[2]?.textContent) : "",
    };
  }

  const flexItems = cell.querySelectorAll('[class*="flexItem"]');
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

  const flexColumn = cell.querySelector('[class*="flex-flex"]');
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
  const title = normalizeText(
    cell.querySelector('[data-testid*="rating-title"]')?.textContent,
  );
  const description = normalizeText(
    cell.querySelector('[data-testid*="rating-description"]')?.textContent,
  );
  const points = normalizeText(
    cell.querySelector('[data-testid*="rating-points"]')?.textContent,
  );

  if (title || description) {
    return { title, description, points };
  }

  const structured = parseA2RatingParts(cell);
  if (structured && (hasRatingBody(structured) || structured.points !== "")) {
    return structured;
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
  const cells = container.querySelectorAll<HTMLElement>(A2_RATING_CELL_SELECTOR);
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

function extractA2TraditionalRubric(): RubricTable | null {
  const rubricRows = findA2RubricRows();
  if (rubricRows.length === 0) {
    return null;
  }

  const rows: RubricTableRow[] = [];
  for (const row of rubricRows) {
    const ratingsContainer =
      row.querySelector('[data-testid="traditional-view-criterion-ratings"]') ?? row;
    const criterion = extractA2CriterionFromRow(row);
    const points = extractA2PointsFromRow(row);
    const ratings = extractA2Ratings(ratingsContainer);

    if (!criterion && ratings.length === 0) {
      continue;
    }

    rows.push({ criterion, ratings, points });
  }

  if (rows.length === 0) {
    return null;
  }

  const header = extractA2Header(rubricRows[0]!);
  return { header, rows };
}

// --- Classic Canvas rubric table (fallback) ---

function findClassicRubricTable(): HTMLTableElement | null {
  return document.querySelector<HTMLTableElement>(
    "#assignment_show table.rubrics, #assignment_show table.rubric, .rubric_container table.rubrics, .rubric_container table.rubric",
  );
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
  const title = normalizeText(cell.querySelector(".rating-description-title, .header, strong")?.textContent);
  const description = normalizeText(cell.querySelector(".description, .details, p")?.textContent);
  const points = normalizeText(cell.querySelector(".points")?.textContent);

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
      nestedRow.querySelectorAll(":scope > td.rating, :scope > td"),
    );
    const ratings = ratingCells.map((el) => parseClassicRatingCell(el)).filter(hasRatingBody);
    if (ratings.length > 0) {
      return ratings;
    }
  }

  const ratingBoxes = cell.querySelectorAll(":scope > .rating");
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

function extractClassicRubric(): RubricTable | null {
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
  for (const tr of bodyRows) {
    if (tr.querySelector('[data-testid="traditional-view-criterion-ratings"]')) {
      continue;
    }

    const cells = classicRowCells(tr);
    if (cells.length === 0) {
      continue;
    }

    const criterion = normalizeText(cells[0]?.textContent);
    if (!criterion) {
      continue;
    }

    rows.push({
      criterion,
      ratings: extractClassicRatingsFromRow(cells),
      points: extractClassicPointsFromRow(cells),
    });
  }

  if (rows.length === 0) {
    return null;
  }

  return { header, rows };
}

/** A2 traditional view → classic table (strict: A2 only). */
export function extractRubricTable(): RubricTable | null {
  return extractNullableWithFallbacks(extractA2TraditionalRubric, extractClassicRubric);
}
