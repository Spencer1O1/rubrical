import { upload } from "../../canvas/anchors";
import { queryAnchor } from "../../canvas/query";
import type { RowAccessibility } from "../types";
import indicatorCss from "./indicators.css";

const STYLE_ID = "rubrical-staged-file-indicator-styles";
const REUPLOAD_LABEL = "Re-upload for Rubrical";

const fileNameSelectors = [
  ...upload.fileRow.fileName.a2,
  ...upload.fileRow.fileName.classic,
] as const;

function ensureIndicatorStyles(): void {
  if (document.getElementById(STYLE_ID)) {
    return;
  }
  const style = document.createElement("style");
  style.id = STYLE_ID;
  style.textContent = indicatorCss;
  document.head.append(style);
}

function findRowElement(fileName: string, fileId: string | null): HTMLTableRowElement | null {
  const table = queryAnchor(upload.table);
  if (!table) {
    return null;
  }

  for (const row of Array.from(table.querySelectorAll("tbody tr"))) {
    if (fileId) {
      const trash = row.querySelector(`button[id="${CSS.escape(fileId)}"]`);
      if (trash) {
        return row as HTMLTableRowElement;
      }
    }

    for (const selector of fileNameSelectors) {
      const title = row.querySelector(selector)?.getAttribute("title")?.trim();
      if (title === fileName) {
        return row as HTMLTableRowElement;
      }
    }
  }

  return null;
}

function isEmptyCanvasCell(cell: HTMLTableCellElement): boolean {
  return cell.childElementCount === 0 && cell.textContent?.trim() === "";
}

function ensureIndicatorCell(row: HTMLTableRowElement): HTMLTableCellElement {
  const existing = row.querySelector<HTMLTableCellElement>("[data-rubrical-file-indicator-cell]");
  if (existing) {
    return existing;
  }

  const cells = Array.from(row.querySelectorAll("td"));
  const emptySlot = cells.find(
    (cell, index) => index >= 2 && isEmptyCanvasCell(cell) && !cell.querySelector("button[id]"),
  );
  if (emptySlot) {
    emptySlot.dataset.rubricalFileIndicatorCell = "slot";
    return emptySlot;
  }

  const cell = document.createElement("td");
  cell.dataset.rubricalFileIndicatorCell = "inserted";
  cell.className = cells.at(-1)?.className ?? cells[0]?.className ?? "";
  cell.setAttribute("dir", "ltr");

  const trashCell = cells.find((candidate) => candidate.querySelector("button[id]"));
  if (trashCell) {
    row.insertBefore(cell, trashCell);
  } else {
    row.append(cell);
  }

  return cell;
}

function removeIndicatorFromRow(row: HTMLTableRowElement): void {
  delete row.dataset.rubricalFileState;
  row.querySelector("[data-rubrical-file-indicator]")?.remove();

  const cell = row.querySelector<HTMLTableCellElement>("[data-rubrical-file-indicator-cell]");
  if (!cell) {
    return;
  }

  if (cell.dataset.rubricalFileIndicatorCell === "inserted") {
    cell.remove();
    return;
  }

  delete cell.dataset.rubricalFileIndicatorCell;
}

function decorateReuploadRow(row: HTMLTableRowElement): void {
  const cell = ensureIndicatorCell(row);
  const existing = cell.querySelector<HTMLSpanElement>("[data-rubrical-file-indicator]");

  for (const stray of Array.from(row.querySelectorAll("[data-rubrical-file-indicator]"))) {
    if (!cell.contains(stray)) {
      stray.remove();
    }
  }

  if (row.dataset.rubricalFileState === "reupload" && existing?.textContent === REUPLOAD_LABEL) {
    return;
  }

  row.dataset.rubricalFileState = "reupload";

  let badge = existing;
  if (!badge) {
    badge = document.createElement("span");
    badge.dataset.rubricalFileIndicator = "true";
    cell.append(badge);
  }

  badge.className = "rubrical-file-indicator";
  badge.setAttribute("role", "status");
  badge.title = REUPLOAD_LABEL;
  badge.setAttribute("aria-label", REUPLOAD_LABEL);
  badge.textContent = REUPLOAD_LABEL;
}

function clearStaleIndicators(activeRows: Set<HTMLTableRowElement>): void {
  for (const row of Array.from(
    document.querySelectorAll<HTMLTableRowElement>("[data-rubrical-file-state]"),
  )) {
    if (!activeRows.has(row)) {
      removeIndicatorFromRow(row);
    }
  }
}

/** Show a re-upload warning only for rows Rubrical cannot read (no staged bytes, no server ref). */
export function decorateUploadedFileIndicators(rows: RowAccessibility[]): void {
  ensureIndicatorStyles();

  const activeRows = new Set<HTMLTableRowElement>();
  for (const rowState of rows) {
    if (rowState.state !== "inaccessible") {
      continue;
    }

    const row = findRowElement(rowState.fileName, rowState.fileId);
    if (!row) {
      continue;
    }
    activeRows.add(row);
    decorateReuploadRow(row);
  }

  clearStaleIndicators(activeRows);
}

export function clearUploadedFileIndicators(): void {
  for (const row of Array.from(
    document.querySelectorAll<HTMLTableRowElement>("[data-rubrical-file-state]"),
  )) {
    removeIndicatorFromRow(row);
  }
}
