import { describe, expect, it } from "vitest";
import { newCanvasIdAssignments, snapshotFileIds } from "./id-map-diff";

describe("id-map-diff", () => {
  it("detects newly assigned canvas file ids", () => {
    const previous = snapshotFileIds([
      { normalizedFileName: "resume.pdf", fileId: null },
    ]);

    const assignments = newCanvasIdAssignments(previous, [
      { normalizedFileName: "resume.pdf", fileId: "99543121" },
    ]);

    expect(assignments).toEqual([
      { rowIndex: 0, normalizedFileName: "resume.pdf", fileId: "99543121" },
    ]);
  });

  it("ignores unchanged ids", () => {
    const previous = snapshotFileIds([
      { normalizedFileName: "resume.pdf", fileId: "99543121" },
    ]);

    const assignments = newCanvasIdAssignments(previous, [
      { normalizedFileName: "resume.pdf", fileId: "99543121" },
    ]);

    expect(assignments).toEqual([]);
  });

  it("detects id changes for the same filename", () => {
    const previous = snapshotFileIds([
      { normalizedFileName: "resume.pdf", fileId: "1" },
    ]);

    const assignments = newCanvasIdAssignments(previous, [
      { normalizedFileName: "resume.pdf", fileId: "2" },
    ]);

    expect(assignments).toEqual([
      { rowIndex: 0, normalizedFileName: "resume.pdf", fileId: "2" },
    ]);
  });

  it("tracks duplicate filenames independently by row index", () => {
    const previous = snapshotFileIds([
      { normalizedFileName: "essay.pdf", fileId: "1" },
      { normalizedFileName: "essay.pdf", fileId: null },
    ]);

    const assignments = newCanvasIdAssignments(previous, [
      { normalizedFileName: "essay.pdf", fileId: "1" },
      { normalizedFileName: "essay.pdf", fileId: "2" },
    ]);

    expect(assignments).toEqual([
      { rowIndex: 1, normalizedFileName: "essay.pdf", fileId: "2" },
    ]);
  });
});
