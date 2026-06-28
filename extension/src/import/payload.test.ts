import { describe, expect, it } from "vitest";
import { installFixture, loadFixtureHtml } from "../canvas/fixture-harness";
import { buildImportPayload } from "./payload";
import type { CachedAssignmentContext, LiveImportCapture } from "./types";

const emptyLive: LiveImportCapture = {
  visibleText: "",
  draftText: "",
  draftUrl: "",
  draftKind: "text",
  draftFiles: [],
  draftFileRefs: [],
  stagedUploads: [],
  fileImportWarnings: [],
  capturedAt: "2026-01-01T00:00:00.000Z",
};

function cachedAssignment(metadata: CachedAssignmentContext["metadata"]): CachedAssignmentContext {
  return {
    sourceUrl: "https://canvas.instructure.com/courses/1/assignments/2",
    pageType: "assignment",
    title: "Test",
    instructionsText: "Do the thing.",
    rubric: null,
    metadata,
    cachedAt: "2026-01-01T00:00:00.000Z",
    longDescriptionsFetched: false,
  };
}

describe("buildImportPayload", () => {
  it("reads submission types from the live DOM at click, not stale cache", () => {
    installFixture(loadFixtureHtml("1-file-uploaded.html"));

    const payload = buildImportPayload(
      cachedAssignment({
        dueDateText: "",
        dueAt: "",
        pointsPossibleText: "",
        submissionTypeText: "",
        allowedSubmissionTypes: [],
        courseName: "",
      }),
      emptyLive,
    );

    expect(payload.metadata.allowedSubmissionTypes).toEqual([
      "online_text_entry",
      "online_upload",
    ]);
    expect(payload.metadata.submissionTypeText).toBe("Online Text Entry, Online Upload");
  });

  it("includes file import warnings from live capture", () => {
    installFixture(loadFixtureHtml("1-file-uploaded.html"));

    const payload = buildImportPayload(
      cachedAssignment({
        dueDateText: "",
        dueAt: "",
        pointsPossibleText: "",
        submissionTypeText: "",
        allowedSubmissionTypes: [],
        courseName: "",
      }),
      {
        ...emptyLive,
        fileImportWarnings: ["Could not import 1 uploaded file: resume.pdf"],
      },
    );

    expect(payload.fileImportWarnings).toEqual([
      "Could not import 1 uploaded file: resume.pdf",
    ]);
  });
});
