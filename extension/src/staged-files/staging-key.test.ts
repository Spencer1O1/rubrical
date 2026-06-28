import { describe, expect, it } from "vitest";
import { installFixture, loadFixtureHtml } from "../canvas/fixture-harness";
import {
  stagingKeyFromPage,
  stagingKeyFromSourceUrl,
  stagingKeyKind,
} from "./staging-key";

describe("stagingKeyFromPage", () => {
  it("uses course and assignment ids from the assignment URL", () => {
    installFixture(loadFixtureHtml("1-file-uploaded.html"));
    expect(stagingKeyFromPage()).toBe("807136:assignment:5218393");
    expect(stagingKeyKind(stagingKeyFromPage()!)).toBe("assignment");
  });

  it("uses course and discussion topic ids from the discussion URL", () => {
    installFixture(loadFixtureHtml("3-discussion-reply-open.html"));
    expect(stagingKeyFromPage()).toBe("807136:discussion:3397799");
    expect(stagingKeyKind(stagingKeyFromPage()!)).toBe("discussion");
  });

  it("parses assignment and discussion source URLs", () => {
    expect(
      stagingKeyFromSourceUrl(
        "https://school.instructure.com/courses/807136/assignments/5218393",
      ),
    ).toBe("807136:assignment:5218393");
    expect(
      stagingKeyFromSourceUrl(
        "https://school.instructure.com/courses/807136/discussion_topics/3397799?view=reply",
      ),
    ).toBe("807136:discussion:3397799");
  });
});
