import { beforeEach, describe, expect, it, vi } from "vitest";
import { installFixture, loadFixtureHtml } from "../canvas/fixture-harness";

const resolveAssignmentFilesForImport = vi.fn();
const resolveDiscussionAttachmentForImport = vi.fn();

vi.mock("../staged-files", () => ({
  resolveAssignmentFilesForImport: (...args: unknown[]) => resolveAssignmentFilesForImport(...args),
  resolveDiscussionAttachmentForImport: (...args: unknown[]) =>
    resolveDiscussionAttachmentForImport(...args),
}));

vi.mock("../submission-kind", () => ({
  detectActiveSubmissionKind: () => "text",
}));

describe("extractLiveImportCapture", () => {
  beforeEach(() => {
    resolveAssignmentFilesForImport.mockReset();
    resolveDiscussionAttachmentForImport.mockReset();
  });

  it("warns when text submission is active but uploaded files remain on canvas", async () => {
    installFixture(loadFixtureHtml("assignment-file-uploaded"));
    const { extractLiveImportCapture } = await import("./live-capture");

    const capture = await extractLiveImportCapture();

    expect(capture.fileImportWarnings).toEqual([
      "Canvas shows uploaded file (resume.pdf) but the active submission type is text — those files will not be imported",
    ]);
  });
});
