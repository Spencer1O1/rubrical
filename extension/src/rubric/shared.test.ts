import { describe, expect, it } from "vitest";
import { installFixture, loadFixtureHtml } from "../canvas/fixture-harness";
import { queryAnchorAll } from "../canvas/query";
import { rubric } from "../canvas/anchors";
import { extractA2PointsFromCell, findA2Row, a2NonRatingCells } from "./shared";

function pointsCellFromFixture(name: string): Element {
  installFixture(loadFixtureHtml(name));
  const container = queryAnchorAll(rubric.criterionRatings)[0]!;
  const row = findA2Row(container)!;
  return a2NonRatingCells(row).at(-1)!;
}

describe("extractA2PointsFromCell", () => {
  it("reads the max-points suffix span on assignment rubrics", () => {
    expect(extractA2PointsFromCell(pointsCellFromFixture("assignment-rubric"))).toBe("/10 pts");
  });

  it("reads the max-points suffix span on discussion assessment trays", () => {
    expect(extractA2PointsFromCell(pointsCellFromFixture("discussion-rubric-tray"))).toBe(
      "/1 pts",
    );
  });
});
