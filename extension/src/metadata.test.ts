import { describe, expect, it } from "vitest";
import type { CanvasPageEnv } from "./canvas/assignment-env";
import { installFixture, loadFixtureHtml } from "./canvas/fixture-harness";
import {
  extractAllowedSubmissionTypes,
  extractSubmissionTypeText,
} from "./metadata";

describe("discussion submission metadata", () => {
  it("uses text + upload when the attach button is visible", () => {
    installFixture(loadFixtureHtml("discussion-reply-open"));
    expect(extractAllowedSubmissionTypes()).toEqual(["online_text_entry", "online_upload"]);
    expect(extractSubmissionTypeText()).toBe("Online Text Entry, Online Upload");
  });

  it("prefers the attach button over ENV when the composer is open", () => {
    installFixture(loadFixtureHtml("discussion-reply-open"));
    (window as Window & { ENV?: CanvasPageEnv }).ENV = { can_attach_entries: false };
    expect(extractAllowedSubmissionTypes()).toEqual(["online_text_entry", "online_upload"]);
  });

  it("uses text + upload for a staged attachment without attach-btn", () => {
    installFixture(loadFixtureHtml("discussion-attachment"));
    expect(extractAllowedSubmissionTypes()).toEqual(["online_text_entry", "online_upload"]);
  });

  it("falls back to ENV when the composer is closed", () => {
    installFixture(loadFixtureHtml("discussion-prompt"));
    expect(extractAllowedSubmissionTypes()).toEqual(["online_text_entry", "online_upload"]);
  });

  it("uses text only when attachments are disabled and the composer is closed", () => {
    installFixture(loadFixtureHtml("discussion-prompt"));
    (window as Window & { ENV?: CanvasPageEnv }).ENV = { can_attach_entries: false };
    expect(extractAllowedSubmissionTypes()).toEqual(["online_text_entry"]);
    expect(extractSubmissionTypeText()).toBe("Online Text Entry");
  });
});
