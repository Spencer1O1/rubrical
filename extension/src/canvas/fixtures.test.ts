import { beforeEach, describe, expect, it } from "vitest";
import {
  anchorPresent,
  DECLARED_TEST_IDS,
  FIXTURES_ROOT,
  fixtureContainsTestId,
  installFixture,
  listFixtureCases,
  loadFixtureHtml,
} from "./fixture-harness";
import {
  extractAllowedSubmissionTypes,
  extractDueAtISO,
  extractDueDateText,
  extractPointsPossibleText,
  extractSubmissionTypeText,
} from "../metadata";
import { readDueFromTimeElement } from "./anchors";
import { setStrictExtraction } from "../strict";

describe("canvas fixture contracts", () => {
  beforeEach(() => {
    setStrictExtraction(false);
  });

  describe("declared test ids", () => {
    it("each declared id appears in at least one HTML fixture", () => {
      const htmlFixtures = listFixtureCases().map(({ html }) => loadFixtureHtml(html));
      const missing = DECLARED_TEST_IDS.filter(
        ({ id }) => !htmlFixtures.some((html) => fixtureContainsTestId(html, id)),
      );

      expect(missing, `Add a fixture or remove unused ids: ${JSON.stringify(missing)}`).toEqual([]);
    });
  });

  describe.each(listFixtureCases())("$html", ({ html, expectations }) => {
    beforeEach(() => {
      installFixture(loadFixtureHtml(html));
    });

    it("matches anchor presence expectations", () => {
      for (const [key, expected] of Object.entries(expectations.anchors)) {
        expect(anchorPresent(key), key).toBe(expected);
      }
    });

    it("matches metadata extraction expectations", () => {
      const { extraction } = expectations;
      if (Object.keys(extraction).length === 0) {
        return;
      }

      if (extraction.dueAt !== undefined) {
        expect(readDueFromTimeElement().dueAt).toBe(extraction.dueAt);
        expect(extractDueAtISO()).toBe(extraction.dueAt);
      }

      if (extraction.dueDateText !== undefined) {
        expect(extractDueDateText()).toBe(extraction.dueDateText);
      }

      if (extraction.pointsPossibleText !== undefined) {
        expect(extractPointsPossibleText()).toBe(extraction.pointsPossibleText);
      }

      if (extraction.submissionTypeText !== undefined) {
        expect(extractSubmissionTypeText()).toBe(extraction.submissionTypeText);
      }

      if (extraction.allowedSubmissionTypes !== undefined) {
        expect(extractAllowedSubmissionTypes()).toEqual(extraction.allowedSubmissionTypes);
      }
    });
  });
});

describe("fixture inventory", () => {
  it("has matching html + expectations for every assignment fixture", () => {
    const cases = listFixtureCases();
    expect(cases.length).toBeGreaterThanOrEqual(11);
    for (const { html, expectations } of cases) {
      expect(expectations.fixture).toBe(html);
    }
    expect(FIXTURES_ROOT.endsWith("/fixtures")).toBe(true);
  });
});
