import { describe, expect, it } from "vitest";
import { installFixture, loadFixtureHtml } from "../../canvas/fixture-harness";
import { isAssignmentFileUploadEditable } from "./hooks";

describe("isAssignmentFileUploadEditable", () => {
  it("is false on submitted read-only assignments", () => {
    installFixture(loadFixtureHtml("assignment-submitted"));
    expect(isAssignmentFileUploadEditable()).toBe(false);
  });

  it("is true while the student is composing a file upload attempt", () => {
    installFixture(loadFixtureHtml("assignment-file-uploaded"));
    expect(isAssignmentFileUploadEditable()).toBe(true);
  });
});
