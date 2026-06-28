import { describe, expect, it } from "vitest";
import { installFixture, loadFixtureHtml } from "../canvas/fixture-harness";
import { readDiscussionComposerAttachment } from "./discussion/composer";

describe("discussion composer session", () => {
  it("detects composer attachments even when the hidden file input is absent", () => {
    installFixture(loadFixtureHtml("discussion-attachment"));
    expect(readDiscussionComposerAttachment()).not.toBeNull();
  });
});
