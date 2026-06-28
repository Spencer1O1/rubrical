import { describe, expect, it } from "vitest";
import { installFixture, loadFixtureHtml } from "../../canvas/fixture-harness";
import { normalizeFileName } from "../staging-key";
import { mergeRowAccessibility } from "./merge";
import { scanAssignmentUploadedRows } from "./canvas-rows";

describe("scanAssignmentUploadedRows", () => {
  it("returns uploaded table rows from the file-upload fixture", () => {
    installFixture(loadFixtureHtml("1-file-uploaded.html"));
    expect(scanAssignmentUploadedRows()).toEqual([
      {
        fileName: "resume.pdf",
        fileId: "99543121",
      },
    ]);
  });

  it("returns no rows when no files are uploaded", () => {
    installFixture(loadFixtureHtml("1-text-submission-tab.html"));
    expect(scanAssignmentUploadedRows()).toEqual([]);
  });

  it("includes filenames without a file extension", () => {
    installFixture("<div></div>");
    const root = document.createElement("div");
    root.innerHTML = `
      <table data-testid="uploaded_files_table">
        <tbody>
          <tr>
            <td><span title="README">README</span></td>
            <td><button id="123" type="button">Delete</button></td>
          </tr>
        </tbody>
      </table>
    `;
    document.body.append(root);
    expect(scanAssignmentUploadedRows(root)).toEqual([
      { fileName: "README", fileId: "123" },
    ]);
  });
});

describe("assignment upload indicators", () => {
  it("marks uploaded table rows inaccessible after Rubrical drops the server copy", () => {
    installFixture(loadFixtureHtml("1-file-uploaded.html"));
    const rows = mergeRowAccessibility(
      scanAssignmentUploadedRows().map((row) => ({
        ...row,
        normalizedFileName: normalizeFileName(row.fileName),
      })),
      [],
      [],
    );
    expect(rows[0]?.state).toBe("inaccessible");
  });

  it("does not warn when the server manifest still has the uploaded file", () => {
    installFixture(loadFixtureHtml("1-file-uploaded.html"));
    const rows = mergeRowAccessibility(
      scanAssignmentUploadedRows().map((row) => ({
        ...row,
        normalizedFileName: normalizeFileName(row.fileName),
      })),
      [],
      [
        {
          serverFileId: 42,
          fileName: "resume.pdf",
          canvasFileId: "99543121",
          byteSize: 100,
          uploadedAt: "2026-06-26T12:00:00.000Z",
        },
      ],
    );
    expect(rows[0]?.state).toBe("saved");
  });

  it("matches saved manifest rows by canvas file id when Canvas renamed the file", () => {
    installFixture(loadFixtureHtml("1-file-uploaded.html"));
    const rows = mergeRowAccessibility(
      scanAssignmentUploadedRows().map((row) => ({
        ...row,
        normalizedFileName: normalizeFileName(row.fileName),
      })),
      [],
      [
        {
          serverFileId: 42,
          fileName: "resume-original.pdf",
          canvasFileId: "99543121",
          byteSize: 100,
          uploadedAt: "2026-06-26T12:00:00.000Z",
        },
      ],
    );
    expect(rows[0]?.state).toBe("saved");
    expect(rows[0]?.serverFileId).toBe(42);
  });
});
