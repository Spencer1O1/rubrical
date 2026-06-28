import { describe, expect, it } from "vitest";
import { installFixture, loadFixtureHtml } from "../../canvas/fixture-harness";
import { readDiscussionComposerAttachment } from "./composer";

describe("discussion composer attachment", () => {
  it("reads attachment metadata from the composer fixture", () => {
    installFixture(loadFixtureHtml("discussion-attachment"));
    expect(readDiscussionComposerAttachment()).toEqual({
      fileName: "resume-1.pdf",
      downloadUrl: "https://canvas.instructure.com/files/99543507/download?download_frd=1",
      canvasFileId: "99543507",
    });
  });

  it("returns null when no attachment is present", () => {
    installFixture(loadFixtureHtml("discussion-reply-open"));
    expect(readDiscussionComposerAttachment()).toBeNull();
  });
});
