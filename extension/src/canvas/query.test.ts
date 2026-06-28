import { Window } from "happy-dom";
import { beforeEach, describe, expect, it } from "vitest";
import type { CanvasAnchor } from "./anchors/types";
import {
  anyAnchorPresent,
  combinedSelector,
  extractAnchor,
  firstMatch,
  queryAnchor,
  runTiers,
  testId,
} from "./query";
import { setStrictExtraction } from "../strict";

const dom = new Window();

function setHtml(html: string): void {
  dom.document.body.innerHTML = html;
}

describe("canvas query helpers", () => {
  beforeEach(() => {
    setStrictExtraction(false);
    setHtml("");
    (globalThis as typeof globalThis & { document: Document }).document =
      dom.document as unknown as Document;
  });

  it("builds data-testid selectors", () => {
    expect(testId("uploaded_files_table")).toBe('[data-testid="uploaded_files_table"]');
  });

  it("firstMatch returns the first matching selector in order", () => {
    setHtml('<div id="a"></div><div id="b"></div>');
    const match = firstMatch(["#missing", "#a", "#b"]);
    expect(match?.id).toBe("a");
  });

  it("queryAnchor prefers a2 over classic", () => {
    const anchor: CanvasAnchor = {
      a2: [testId("attempt-tab")],
      classic: ["#assignment_show"],
    };
    setHtml('<div id="assignment_show"></div><div data-testid="attempt-tab"></div>');
    expect(queryAnchor(anchor)?.getAttribute("data-testid")).toBe("attempt-tab");
  });

  it("queryAnchor falls back to classic when a2 is missing", () => {
    const anchor: CanvasAnchor = {
      a2: [testId("attempt-tab")],
      classic: ["#assignment_show"],
    };
    setHtml('<div id="assignment_show"></div>');
    expect(queryAnchor(anchor)?.id).toBe("assignment_show");
  });

  it("queryAnchor skips classic and extra in strict mode", () => {
    setStrictExtraction(true);
    const anchor: CanvasAnchor = {
      a2: [testId("attempt-tab")],
      classic: ["#assignment_show"],
      extra: ["main"],
    };
    setHtml('<div id="assignment_show"></div><main></main>');
    expect(queryAnchor(anchor)).toBeNull();
  });

  it("anyAnchorPresent checks all tiers", () => {
    const anchor: CanvasAnchor = {
      a2: [testId("missing")],
      classic: ["#assignment_show"],
    };
    setHtml('<div id="assignment_show"></div>');
    expect(anyAnchorPresent(anchor)).toBe(true);
  });

  it("combinedSelector joins all tiers", () => {
    const anchor: CanvasAnchor = {
      a2: [testId("a")],
      classic: ["#b"],
      extra: ["main"],
    };
    expect(combinedSelector(anchor)).toBe('[data-testid="a"], #b, main');
  });

  it("runTiers prefers earlier tiers", () => {
    expect(
      runTiers([
        () => "",
        () => "classic",
        () => "env",
      ]),
    ).toBe("classic");
  });

  it("extractAnchor falls back to env after DOM tiers", () => {
    const anchor: CanvasAnchor = {
      a2: [testId("missing")],
      classic: ["#missing-classic"],
      env: () => "from-env",
    };
    expect(extractAnchor(anchor)).toBe("from-env");
  });

  it("extractAnchor skips env in strict mode", () => {
    setStrictExtraction(true);
    const anchor: CanvasAnchor = {
      a2: [testId("missing")],
      classic: ["#assignment_show"],
      env: () => "from-env",
    };
    expect(extractAnchor(anchor)).toBe("");
  });
});
