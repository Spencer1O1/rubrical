import { describe, expect, it } from "vitest";
import { installFixture, loadFixtureHtml } from "../canvas/fixture-harness";
import { extractAssignmentContext } from "./assignment-context";

describe("extractAssignmentContext", () => {
  it("does not prefetch submission types", async () => {
    installFixture(loadFixtureHtml("1-file-uploaded.html"));

    const context = await extractAssignmentContext("assignment");

    expect(context.metadata.allowedSubmissionTypes).toEqual([]);
    expect(context.metadata.submissionTypeText).toBe("");
    expect(context.metadata.pointsPossibleText).toContain("Points");
  });
});
