import { describe, expect, it } from "vitest";
import { installFixture, loadFixtureHtml } from "../../canvas/fixture-harness";
import { isAssignmentFileUploadEditable } from "./hooks";

describe("isAssignmentFileUploadEditable", () => {
  it("is false on submitted read-only assignments", () => {
    installFixture(loadFixtureHtml("4-submitted.html"));
    expect(isAssignmentFileUploadEditable()).toBe(false);
  });

  it("is true while the student is composing a file upload attempt", () => {
    installFixture(loadFixtureHtml("1-file-uploaded.html"));
    expect(isAssignmentFileUploadEditable()).toBe(true);
  });
});
