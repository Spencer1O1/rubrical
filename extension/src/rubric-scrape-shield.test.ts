import { readFileSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, it } from "vitest";

const FIXTURES_ROOT = join(import.meta.dirname, "../../fixtures");

function fixtureHtmlTag(stem: string): string {
  const html = readFileSync(join(FIXTURES_ROOT, `${stem.replace(/\.html$/, "")}.html`), "utf8");
  const match = html.match(/<html[^>]*>/i);
  if (!match) {
    throw new Error(`missing <html> tag in ${stem}`);
  }
  return match[0];
}

describe("InstUI html scroll lock fixture diff", () => {
  it("documents what beginLongDescriptionScrape neutralizes on html", () => {
    const closed = fixtureHtmlTag("assignment-rubric");
    const open = fixtureHtmlTag("assignment-rubric-modal-open");

    expect(closed).not.toContain("position: fixed");
    expect(open).toContain("position: fixed");
    expect(open).toContain("overflow: hidden");
    expect(open).toContain("width: calc(100% - 19px)");
    expect(open).toMatch(/top: -[\d.]+px/);
  });
});
