import { describe, expect, it } from "vitest";
import { normalizeSourceUrl } from "./normalize-source-url";

describe("normalizeSourceUrl", () => {
  it("strips query and fragment", () => {
    expect(
      normalizeSourceUrl(
        "https://usu.instructure.com/courses/807136/assignments/5218393?submitting=1#rubric",
      ),
    ).toBe("https://usu.instructure.com/courses/807136/assignments/5218393");
  });

  it("trims trailing slash", () => {
    expect(normalizeSourceUrl("https://school.instructure.com/courses/1/assignments/2/")).toBe(
      "https://school.instructure.com/courses/1/assignments/2",
    );
  });
});
