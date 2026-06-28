import { beforeEach, describe, expect, it } from "vitest";
import { submit, submitAnchorOrder } from "./canvas/anchors";
import { installFixture, loadFixtureHtml } from "./canvas/fixture-harness";
import { queryAnchor } from "./canvas/query";
import { extractInstructions } from "./extractor";
import { extractDraftText } from "./draft";
import { detectActiveSubmissionKind } from "./submission-kind";
import { setStrictExtraction } from "./strict";

describe("detectActiveSubmissionKind", () => {
  beforeEach(() => {
    setStrictExtraction(false);
  });

  it("detects file tab on upload-selected fixture", () => {
    installFixture(loadFixtureHtml("2-text-submission.html"));
    expect(detectActiveSubmissionKind()).toBe("file");
  });

  it("detects text tab on text-selected fixture", () => {
    installFixture(loadFixtureHtml("1-text-submission-tab.html"));
    expect(detectActiveSubmissionKind()).toBe("text");
  });

  it("detects open discussion composer as text draft", () => {
    installFixture(loadFixtureHtml("3-discussion-reply-open.html"));
    expect(detectActiveSubmissionKind()).toBe("text");
  });

  it("does not treat upload-pane presence alone as file mode when text is selected", () => {
    installFixture(`
      <div data-testid="assignment-2-student-content-tabs">
        <div data-testid="submission-type-selector">
          <div data-testid="online_text_entry">
            <span>Submission type Text, currently selected</span>
          </div>
          <div data-testid="online_upload"></div>
        </div>
        <span data-testid="upload-pane"></span>
      </div>
    `);
    expect(detectActiveSubmissionKind()).toBe("text");
  });
});

describe("discussion extraction", () => {
  beforeEach(() => {
    setStrictExtraction(true);
  });

  it("extracts the discussion prompt from the closed discussion fixture", () => {
    installFixture(loadFixtureHtml("3-discussion.html"));
    const prompt = extractInstructions();
    expect(prompt).toContain("Arts Discussion Forum #9");
    expect(prompt).toContain("Vulnerability");
  });

  it("uses the open composer submit anchor, not the closed Reply opener", () => {
    installFixture(loadFixtureHtml("3-discussion-reply-open.html"));
    expect(queryAnchor(submit.discussionEditSubmit)).not.toBeNull();
    expect(queryAnchor(submit.discussionReply)).toBeNull();
    expect(submitAnchorOrder).not.toContain(submit.discussionReply);
  });
});

describe("extractDraftText", () => {
  beforeEach(() => {
    setStrictExtraction(true);
  });

  it("reads TinyMCE content from text-editor on the text-tab fixture", () => {
    installFixture(loadFixtureHtml("1-text-submission-tab.html"));

    (window as Window & { tinymce?: unknown }).tinymce = {
      triggerSave: () => {},
      get: (id: string) =>
        id === "textentry_text"
          ? { save: () => {}, getContent: () => "<p>hello there</p>" }
          : null,
      editors: [{ id: "textentry_text", save: () => {}, getContent: () => "<p>hello there</p>" }],
    };

    expect(extractDraftText()).toBe("<p>hello there</p>");
  });

  it("reads TinyMCE content from DiscussionEdit-container on the reply-open fixture", () => {
    installFixture(loadFixtureHtml("3-discussion-reply-open.html"));

    (window as Window & { tinymce?: unknown }).tinymce = {
      triggerSave: () => {},
      get: (id: string) =>
        id === "message-body-root"
          ? { save: () => {}, getContent: () => "<p>This is the reply box.</p>" }
          : null,
      editors: [
        { id: "message-body-root", save: () => {}, getContent: () => "<p>This is the reply box.</p>" },
      ],
    };

    expect(extractDraftText()).toBe("<p>This is the reply box.</p>");
  });
});
