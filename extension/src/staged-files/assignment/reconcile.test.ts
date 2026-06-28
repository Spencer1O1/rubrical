import { describe, expect, it } from "vitest";
import { findReconcilePromotions } from "./reconcile";

describe("findReconcilePromotions", () => {
  it("promotes provisional staged entry when canvas id appears", () => {
    const promotions = findReconcilePromotions(
      [{ rowIndex: 0, normalizedFileName: "notes.pdf", fileId: "123" }],
      [
        {
          assignmentKey: "1:assignment:2",
          fileName: "notes.pdf",
          normalizedFileName: "notes.pdf",
          stagedAt: "2026-06-26T12:00:00.000Z",
          mimeType: "application/pdf",
        },
      ],
    );

    expect(promotions).toEqual([
      {
        normalizedFileName: "notes.pdf",
        stagedAt: "2026-06-26T12:00:00.000Z",
        canvasFileId: "123",
      },
    ]);
  });

  it("matches duplicate filenames in upload order", () => {
    const promotions = findReconcilePromotions(
      [
        { rowIndex: 0, normalizedFileName: "essay.pdf", fileId: "10" },
        { rowIndex: 1, normalizedFileName: "essay.pdf", fileId: "11" },
      ],
      [
        {
          assignmentKey: "1:assignment:2",
          fileName: "essay.pdf",
          normalizedFileName: "essay.pdf",
          stagedAt: "2026-06-26T12:00:00.000Z",
          mimeType: "application/pdf",
        },
        {
          assignmentKey: "1:assignment:2",
          fileName: "essay.pdf",
          normalizedFileName: "essay.pdf",
          stagedAt: "2026-06-26T12:01:00.000Z",
          mimeType: "application/pdf",
        },
      ],
    );

    expect(promotions).toEqual([
      {
        normalizedFileName: "essay.pdf",
        stagedAt: "2026-06-26T12:00:00.000Z",
        canvasFileId: "10",
      },
      {
        normalizedFileName: "essay.pdf",
        stagedAt: "2026-06-26T12:01:00.000Z",
        canvasFileId: "11",
      },
    ]);
  });

  it("promotes provisional staged entry by row index when Canvas renamed the file", () => {
    const promotions = findReconcilePromotions(
      [{ rowIndex: 1, normalizedFileName: "essay-1.pdf", fileId: "11" }],
      [
        {
          assignmentKey: "1:assignment:2",
          fileName: "essay.pdf",
          normalizedFileName: "essay.pdf",
          stagedAt: "2026-06-26T12:00:00.000Z",
          mimeType: "application/pdf",
        },
        {
          assignmentKey: "1:assignment:2",
          fileName: "essay.pdf",
          normalizedFileName: "essay.pdf",
          stagedAt: "2026-06-26T12:01:00.000Z",
          mimeType: "application/pdf",
        },
      ],
    );

    expect(promotions).toEqual([
      {
        normalizedFileName: "essay.pdf",
        stagedAt: "2026-06-26T12:01:00.000Z",
        canvasFileId: "11",
      },
    ]);
  });
});
