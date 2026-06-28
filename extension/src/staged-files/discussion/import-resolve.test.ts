import { beforeEach, describe, expect, it, vi } from "vitest";
import { installFixture, loadFixtureHtml } from "../../canvas/fixture-harness";
import { resolveDiscussionAttachmentForImport } from "./import-resolve";

const listStagedFiles = vi.fn();
const getStagedFilePayload = vi.fn();
const putStagedFileBytes = vi.fn();

vi.mock("../store", () => ({
  listStagedFiles: (...args: unknown[]) => listStagedFiles(...args),
  getStagedFilePayload: (...args: unknown[]) => getStagedFilePayload(...args),
  putStagedFileBytes: (...args: unknown[]) => putStagedFileBytes(...args),
}));

describe("resolveDiscussionAttachmentForImport", () => {
  beforeEach(() => {
    listStagedFiles.mockReset();
    getStagedFilePayload.mockReset();
    putStagedFileBytes.mockReset();
    listStagedFiles.mockResolvedValue([]);
    putStagedFileBytes.mockResolvedValue(undefined);
  });

  it("returns null when the composer has no attachment", async () => {
    installFixture(loadFixtureHtml("3-discussion-reply-open.html"));
    await expect(resolveDiscussionAttachmentForImport()).resolves.toBeNull();
  });

  it("prefers staged session bytes", async () => {
    installFixture(loadFixtureHtml("3-discussion-attachment.html"));
    listStagedFiles.mockResolvedValue([
      {
        assignmentKey: "807136:discussion:3397799",
        canvasFileId: "99543507",
        fileName: "resume-1.pdf",
        normalizedFileName: "resume-1.pdf",
        stagedAt: "2026-06-26T12:00:00.000Z",
        mimeType: "application/pdf",
      },
    ]);
    getStagedFilePayload.mockResolvedValue({
      fileName: "resume-1.pdf",
      mimeType: "application/pdf",
      contentBase64: "c3RhZ2Vk",
      canvasFileId: "99543507",
    });

    await expect(resolveDiscussionAttachmentForImport()).resolves.toEqual({
      fileName: "resume-1.pdf",
      mimeType: "application/pdf",
      contentBase64: "c3RhZ2Vk",
      canvasFileId: "99543507",
      sortOrder: 0,
    });
  });

  it("downloads from Canvas and stages bytes when IDB missed the pick", async () => {
    installFixture(loadFixtureHtml("3-discussion-attachment.html"));
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(new Uint8Array([1, 2, 3]), {
        status: 200,
        headers: { "content-type": "application/pdf" },
      }),
    );

    const file = await resolveDiscussionAttachmentForImport();

    expect(file?.fileName).toBe("resume-1.pdf");
    expect(file?.canvasFileId).toBe("99543507");
    expect(putStagedFileBytes).toHaveBeenCalledOnce();
    fetchMock.mockRestore();
  });
});
