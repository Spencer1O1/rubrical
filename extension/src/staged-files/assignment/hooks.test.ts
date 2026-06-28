import { beforeEach, describe, expect, it, vi } from "vitest";

const flushPendingUploads = vi.fn();
const countPendingUploads = vi.fn();

vi.mock("./pending-staging", () => ({
  flushPendingUploads: (...args: unknown[]) => flushPendingUploads(...args),
  countPendingUploads: (...args: unknown[]) => countPendingUploads(...args),
  forgetPendingUpload: vi.fn(),
  rememberPendingUpload: vi.fn(),
}));

vi.mock("../store", () => ({
  putStagedFile: vi.fn(),
}));

vi.mock("../../canvas/anchors", () => ({
  upload: {
    attemptRoot: { a2: [], classic: [] },
    table: { a2: [], classic: [] },
    fileRow: { fileName: { a2: [], classic: [] }, trashButton: { a2: [], classic: [] } },
  },
  uploadFileInputSelector: () => "input[type=file]",
}));

vi.mock("../../canvas/query", () => ({
  queryAnchor: () => null,
}));

describe("awaitStagingIdle", () => {
  beforeEach(() => {
    flushPendingUploads.mockReset();
    countPendingUploads.mockReset();
    flushPendingUploads.mockResolvedValue(0);
  });

  it("throws when pending uploads remain after flush", async () => {
    countPendingUploads.mockReturnValue(2);
    const { awaitStagingIdle } = await import("./hooks");

    await expect(awaitStagingIdle("1:assignment:2")).rejects.toThrow(
      "2 uploaded files could not be staged",
    );
    expect(flushPendingUploads).toHaveBeenCalledWith("1:assignment:2");
  });

  it("resolves when no pending uploads remain", async () => {
    countPendingUploads.mockReturnValue(0);
    const { awaitStagingIdle } = await import("./hooks");

    await expect(awaitStagingIdle("1:assignment:2")).resolves.toBeUndefined();
  });
});
