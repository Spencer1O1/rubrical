import { describe, expect, it } from "vitest";
import { mergeRowAccessibility } from "./merge";

describe("mergeRowAccessibility", () => {
  it("prefers staged over saved when both match", () => {
    const rows = mergeRowAccessibility(
      [{ fileName: "a.pdf", normalizedFileName: "a.pdf", fileId: "1" }],
      [
        {
          assignmentKey: "1:assignment:2",
          canvasFileId: "1",
          fileName: "a.pdf",
          normalizedFileName: "a.pdf",
          stagedAt: "2026-06-26T12:00:00.000Z",
          mimeType: "application/pdf",
        },
      ],
      [{ serverFileId: 9, fileName: "a.pdf", byteSize: 10, uploadedAt: "2026-06-26T11:00:00.000Z" }],
    );

    expect(rows[0]?.state).toBe("staged");
  });

  it("marks saved when manifest matches", () => {
    const rows = mergeRowAccessibility(
      [{ fileName: "a.pdf", normalizedFileName: "a.pdf", fileId: "1" }],
      [],
      [{ serverFileId: 9, fileName: "a.pdf", byteSize: 10, uploadedAt: "2026-06-26T11:00:00.000Z" }],
    );

    expect(rows[0]?.state).toBe("saved");
    expect(rows[0]?.serverFileId).toBe(9);
  });

  it("marks inaccessible when nothing matches", () => {
    const rows = mergeRowAccessibility(
      [{ fileName: "a.pdf", normalizedFileName: "a.pdf", fileId: "1" }],
      [],
      [],
    );

    expect(rows[0]?.state).toBe("inaccessible");
  });

  it("matches saved manifest rows by canvas file id when Canvas renamed the file", () => {
    const rows = mergeRowAccessibility(
      [
        {
          fileName: "resume-630db563-2934.pdf",
          normalizedFileName: "resume-630db563-2934.pdf",
          fileId: "99543121",
        },
      ],
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
    expect(rows[0]?.serverFileId).toBe(42);
  });

  it("prefers staged bytes over a stale server file when canvas has a new attachment", () => {
    const rows = mergeRowAccessibility(
      [{ fileName: "new.pdf", normalizedFileName: "new.pdf", fileId: "99543508" }],
      [
        {
          assignmentKey: "807136:discussion:3397799",
          canvasFileId: "99543508",
          fileName: "new.pdf",
          normalizedFileName: "new.pdf",
          stagedAt: "2026-06-26T12:00:00.000Z",
          mimeType: "application/pdf",
        },
      ],
      [
        {
          serverFileId: 42,
          fileName: "old.pdf",
          canvasFileId: "99543507",
          byteSize: 100,
          uploadedAt: "2026-06-26T11:00:00.000Z",
        },
      ],
    );

    expect(rows[0]?.state).toBe("staged");
  });

  it("matches duplicate saved files when Canvas renames the second copy", () => {
    const rows = mergeRowAccessibility(
      [
        { fileName: "essay.pdf", normalizedFileName: "essay.pdf", fileId: "10" },
        { fileName: "essay-1.pdf", normalizedFileName: "essay-1.pdf", fileId: "11" },
      ],
      [],
      [
        {
          serverFileId: 1,
          fileName: "essay.pdf",
          canvasFileId: "10",
          byteSize: 100,
          uploadedAt: "2026-06-26T12:00:00.000Z",
        },
        {
          serverFileId: 2,
          fileName: "essay.pdf",
          canvasFileId: "11",
          byteSize: 100,
          uploadedAt: "2026-06-26T12:01:00.000Z",
        },
      ],
    );

    expect(rows[0]?.state).toBe("saved");
    expect(rows[0]?.serverFileId).toBe(1);
    expect(rows[1]?.state).toBe("saved");
    expect(rows[1]?.serverFileId).toBe(2);
  });

  it("disambiguates duplicate manifest filenames by row order when canvas ids are missing", () => {
    const rows = mergeRowAccessibility(
      [
        { fileName: "essay.pdf", normalizedFileName: "essay.pdf", fileId: null },
        { fileName: "essay-1.pdf", normalizedFileName: "essay-1.pdf", fileId: null },
      ],
      [],
      [
        {
          serverFileId: 1,
          fileName: "essay.pdf",
          byteSize: 100,
          uploadedAt: "2026-06-26T12:00:00.000Z",
        },
        {
          serverFileId: 2,
          fileName: "essay.pdf",
          byteSize: 100,
          uploadedAt: "2026-06-26T12:01:00.000Z",
        },
      ],
    );

    expect(rows[0]?.state).toBe("saved");
    expect(rows[0]?.serverFileId).toBe(1);
    expect(rows[1]?.state).toBe("saved");
    expect(rows[1]?.serverFileId).toBe(2);
  });
});
