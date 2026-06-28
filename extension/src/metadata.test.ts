import { describe, expect, it } from "vitest";
import type { CanvasPageEnv } from "./canvas/assignment-env";
import { installFixture, loadFixtureHtml } from "./canvas/fixture-harness";
import {
  extractAllowedSubmissionTypes,
  extractSubmissionTypeText,
} from "./metadata";

describe("discussion submission metadata", () => {
  it("uses text + upload when the attach button is visible", () => {
    installFixture(loadFixtureHtml("3-discussion-reply-open.html"));
    expect(extractAllowedSubmissionTypes()).toEqual(["online_text_entry", "online_upload"]);
    expect(extractSubmissionTypeText()).toBe("Online Text Entry, Online Upload");
  });

  it("prefers the attach button over ENV when the composer is open", () => {
    installFixture(loadFixtureHtml("3-discussion-reply-open.html"));
    (window as Window & { ENV?: CanvasPageEnv }).ENV = { can_attach_entries: false };
    expect(extractAllowedSubmissionTypes()).toEqual(["online_text_entry", "online_upload"]);
  });

  it("uses text + upload for a staged attachment without attach-btn", () => {
    installFixture(loadFixtureHtml("3-discussion-attachment.html"));
    expect(extractAllowedSubmissionTypes()).toEqual(["online_text_entry", "online_upload"]);
  });

  it("falls back to ENV when the composer is closed", () => {
    installFixture(loadFixtureHtml("3-discussion.html"));
    expect(extractAllowedSubmissionTypes()).toEqual(["online_text_entry", "online_upload"]);
  });

  it("uses text only when attachments are disabled and the composer is closed", () => {
    installFixture(loadFixtureHtml("3-discussion.html"));
    (window as Window & { ENV?: CanvasPageEnv }).ENV = { can_attach_entries: false };
    expect(extractAllowedSubmissionTypes()).toEqual(["online_text_entry"]);
    expect(extractSubmissionTypeText()).toBe("Online Text Entry");
  });
});
