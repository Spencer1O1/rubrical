import { beforeEach, describe, expect, it, vi } from "vitest";
import { installFixture, loadFixtureHtml } from "../../canvas/fixture-harness";

const listStagedFiles = vi.fn();
const fetchDraftManifestOnce = vi.fn();
const getDraftManifest = vi.fn();

vi.mock("../store", () => ({
  listStagedFiles: (...args: unknown[]) => listStagedFiles(...args),
}));

vi.mock("./manifest-client", () => ({
  fetchDraftManifestOnce: (...args: unknown[]) => fetchDraftManifestOnce(...args),
  getDraftManifest: (...args: unknown[]) => getDraftManifest(...args),
}));

describe("resolveAssignmentFilesForImport", () => {
  beforeEach(() => {
    listStagedFiles.mockReset();
    fetchDraftManifestOnce.mockReset();
    getDraftManifest.mockReset();
    fetchDraftManifestOnce.mockResolvedValue(undefined);
    getDraftManifest.mockReturnValue({ files: [] });
    listStagedFiles.mockResolvedValue([]);
  });

  it("returns skipped file names when canvas rows cannot be resolved", async () => {
    installFixture(loadFixtureHtml("1-file-uploaded.html"));
    const { resolveAssignmentFilesForImport } = await import("./import-resolve");

    const result = await resolveAssignmentFilesForImport();

    expect(result.draftFiles).toEqual([]);
    expect(result.draftFileRefs).toEqual([]);
    expect(result.stagedUploads).toEqual([]);
    expect(result.skipped).toEqual(["resume.pdf"]);
  });

  it("returns staged uploads for rows with extension-stored bytes", async () => {
    installFixture(loadFixtureHtml("1-file-uploaded.html"));
    listStagedFiles.mockResolvedValue([
      {
        assignmentKey: "807136:assignment:123",
        fileName: "resume.pdf",
        normalizedFileName: "resume.pdf",
        stagedAt: "2026-01-01T00:00:00.000Z",
        mimeType: "application/pdf",
      },
    ]);

    const { resolveAssignmentFilesForImport } = await import("./import-resolve");
    const result = await resolveAssignmentFilesForImport();

    expect(result.stagedUploads).toHaveLength(1);
    expect(result.stagedUploads[0]?.fileName).toBe("resume.pdf");
    expect(result.skipped).toEqual([]);
  });
});
