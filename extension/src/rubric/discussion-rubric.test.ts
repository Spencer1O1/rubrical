import { describe, expect, it } from "vitest";
import { installFixture, loadFixtureHtml } from "../canvas/fixture-harness";
import { extractA2TraditionalRubric } from "./a2";
import { countCriterionLongDescriptionButtons } from "./long-descriptions";
import {
  findDiscussionPostMenuTrigger,
  findDiscussionShowRubricMenuItem,
  hasVisibleRubricCriteria,
} from "./discussion-rubric";

describe("discussion rubric", () => {
  it("finds the post menu trigger on the base discussion fixture", () => {
    installFixture(loadFixtureHtml("discussion-prompt"));
    expect(findDiscussionPostMenuTrigger()?.getAttribute("data-testid")).toBe(
      "discussion-post-menu-trigger",
    );
  });

  it("finds Show Rubric in the open post menu fixture", () => {
    installFixture(loadFixtureHtml("discussion-menu-open"));
    expect(findDiscussionShowRubricMenuItem()?.textContent).toContain("Show Rubric");
  });

  it("finds long-description controls in the rubric tray fixture", () => {
    installFixture(loadFixtureHtml("discussion-rubric-tray"));
    expect(countCriterionLongDescriptionButtons()).toBe(1);
  });

  it("parses the tray from the rubric-open fixture", () => {
    installFixture(loadFixtureHtml("discussion-rubric-tray"));
    expect(hasVisibleRubricCriteria()).toBe(true);

    const table = extractA2TraditionalRubric([]);
    expect(table?.rows.length).toBe(1);
    expect(table?.rows[0]?.criterion).toContain("Content Quality");
    expect(table?.rows[0]?.points).toBe("/1 pts");
  });
});
