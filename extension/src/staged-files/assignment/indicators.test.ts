import { beforeEach, describe, expect, it } from "vitest";
import { decorateUploadedFileIndicators } from "./indicators";
import { normalizeFileName } from "../staging-key";
import { mergeRowAccessibility } from "./merge";
import { scanAssignmentUploadedRows } from "./canvas-rows";
import { installFixture, loadFixtureHtml } from "../../canvas/fixture-harness";

beforeEach(() => {
  if (typeof CSS === "undefined") {
    (globalThis as { CSS: typeof CSS }).CSS = {
      escape: (value: string) => value.replace(/"/g, '\\"'),
    } as typeof CSS;
  }
});

function mergedRows(fixture: string) {
  installFixture(loadFixtureHtml(fixture));
  return mergeRowAccessibility(
    scanAssignmentUploadedRows().map((row) => ({
      ...row,
      normalizedFileName: normalizeFileName(row.fileName),
    })),
    [],
    [],
  );
}

describe("decorateUploadedFileIndicators", () => {
  it("does not warn on submitted read-only file rows before New Attempt", () => {
    const rows = mergedRows("4-submitted.html");
    expect(rows.some((row) => row.state === "inaccessible")).toBe(true);

    decorateUploadedFileIndicators(rows, { fileUploadEditable: false });

    expect(document.querySelector("[data-rubrical-file-indicator]")).toBeNull();
  });

  it("warns on inaccessible rows when file upload is editable", () => {
    const rows = mergedRows("1-file-uploaded.html");
    expect(rows[0]?.state).toBe("inaccessible");

    decorateUploadedFileIndicators(rows, { fileUploadEditable: true });

    const badge = document.querySelector<HTMLElement>("[data-rubrical-file-indicator]");
    expect(badge?.textContent).toBe("Re-upload for Rubrical");
  });
});
