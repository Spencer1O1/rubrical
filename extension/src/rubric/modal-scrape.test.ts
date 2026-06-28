import { describe, expect, it } from "vitest";
import { installFixture, loadFixtureHtml } from "../canvas/fixture-harness";
import { queryAnchor } from "../canvas/query";
import { discussion } from "../canvas/anchors";
import { closeButtonIn } from "./modal-scrape";

describe("closeButtonIn", () => {
  it("finds InstUI close buttons in the discussion rubric tray fixture", () => {
    installFixture(loadFixtureHtml("3-discussion-rubric-open.html"));
    const tray = queryAnchor(discussion.rubricAssessmentTray);
    expect(tray).not.toBeNull();
    expect(closeButtonIn(tray!)).not.toBeNull();
  });

  it("finds InstUI close buttons in the assignment rubric modal fixture", () => {
    installFixture(loadFixtureHtml("3-discussion-modal-open.html"));
    const modal = queryAnchor(discussion.assignmentRubricModal);
    expect(modal).not.toBeNull();
    expect(closeButtonIn(modal!)).not.toBeNull();
  });
});
