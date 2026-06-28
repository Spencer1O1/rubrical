import { describe, expect, it } from "vitest";
import { composeFixtureHtml } from "./fixture-compose";
import { FIXTURES_ROOT, loadExpectations } from "./fixture-harness";

describe("composeFixtureHtml", () => {
  it("builds assignment-file-uploaded from base + upload pane fragment", () => {
    const { compose } = loadExpectations("assignment-file-uploaded");
    if (!compose) {
      throw new Error("missing compose spec");
    }

    const html = composeFixtureHtml(FIXTURES_ROOT, compose);
    expect(html).toContain('data-testid="uploaded_files_table"');
    expect(html).toContain("resume.pdf");
    expect(html).toContain('data-testid="input-file-drop"');
  });
});
